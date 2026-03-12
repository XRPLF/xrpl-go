package integration

import (
	"encoding/json"
	"fmt"
	"testing"

	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

// sequenceFromTx extracts the Sequence field from a FlatTransaction as uint32.
func sequenceFromTx(tx transaction.FlatTransaction) (uint32, error) {
	switch v := tx["Sequence"].(type) {
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			return 0, err
		}
		return uint32(n), nil
	case float64:
		return uint32(v), nil
	case uint32:
		return v, nil
	case int:
		return uint32(v), nil
	default:
		return 0, fmt.Errorf("unexpected Sequence type: %T", tx["Sequence"])
	}
}

// TestIntegrationLoanSetWithSingleSigning_Websocket tests the full lending protocol lifecycle
// with single signing: VaultCreate -> VaultDeposit -> LoanBrokerSet -> LoanSet with counterparty signing.
func TestIntegrationLoanSetWithSingleSigning_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))

	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 3,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	// Wallet 0: Vault Owner / Loan Broker (same account as in the JS test)
	vaultOwner := runner.GetWallet(0)
	loanBroker := vaultOwner
	// Wallet 1: Depositor
	depositor := runner.GetWallet(1)
	// Wallet 2: Borrower
	borrower := runner.GetWallet(2)

	// Step 1: Create an XRP vault
	assetsMaximum := types.XRPLNumber("1e17")
	vaultCreateTx := &transaction.VaultCreate{
		BaseTx: transaction.BaseTx{
			Account: vaultOwner.GetAddress(),
		},
		Asset: ledger.Asset{
			Currency: "XRP",
		},
		AssetsMaximum: &assetsMaximum,
	}

	vaultCreateFlat := vaultCreateTx.Flatten()
	vaultCreateResp, err := runner.TestTransaction(&vaultCreateFlat, vaultOwner, "tesSUCCESS", nil)
	require.NoError(t, err)
	require.NotNil(t, vaultCreateResp)

	// Compute the Vault hash from the response
	vaultCreateAccount, ok := vaultCreateResp.Tx["Account"].(string)
	require.True(t, ok)
	vaultCreateSequence, err := sequenceFromTx(vaultCreateResp.Tx)
	require.NoError(t, err)

	vaultObjectID, err := xrplhash.Vault(vaultCreateAccount, vaultCreateSequence)
	require.NoError(t, err)

	// Step 2: Depositor deposits 10 XRP into the vault
	vaultDepositTx := &transaction.VaultDeposit{
		BaseTx: transaction.BaseTx{
			Account: depositor.GetAddress(),
		},
		VaultID: types.Hash256(vaultObjectID),
		Amount:  types.XRPCurrencyAmount(10_000_000),
	}

	vaultDepositFlat := vaultDepositTx.Flatten()
	_, err = runner.TestTransaction(&vaultDepositFlat, depositor, "tesSUCCESS", nil)
	require.NoError(t, err)

	// Step 3: Create LoanBroker (Vault Owner == Loan Broker)
	debtMaximum := types.XRPLNumber("25000000")
	loanBrokerSetTx := &transaction.LoanBrokerSet{
		BaseTx: transaction.BaseTx{
			Account: loanBroker.GetAddress(),
		},
		VaultID:     vaultObjectID,
		DebtMaximum: &debtMaximum,
	}

	loanBrokerFlat := transaction.FlatTransaction(loanBrokerSetTx.Flatten())
	err = client.Autofill(&loanBrokerFlat)
	require.NoError(t, err)

	loanBrokerBlob, _, err := loanBroker.Sign(loanBrokerFlat)
	require.NoError(t, err)

	// SubmitTxBlobAndWait ensures the LoanBrokerSet (and all prior txs) are validated
	// before the LoanSet autofill, which looks up the borrower account in the validated ledger.
	loanBrokerTxResp, err := client.SubmitTxBlobAndWait(loanBrokerBlob, true)
	require.NoError(t, err)

	// Compute the LoanBroker hash from the response
	loanBrokerAccount, ok := loanBrokerTxResp.TxJSON["Account"].(string)
	require.True(t, ok)
	loanBrokerSequence, err := sequenceFromTx(loanBrokerTxResp.TxJSON)
	require.NoError(t, err)

	loanBrokerObjectID, err := xrplhash.LoanBroker(loanBrokerAccount, loanBrokerSequence)
	require.NoError(t, err)

	// Step 4: Broker initiates the Loan with counterparty signing
	counterparty := types.Address(borrower.GetAddress())
	paymentTotal := types.PaymentTotal(1)
	loanSetTx := &transaction.LoanSet{
		BaseTx: transaction.BaseTx{
			Account: loanBroker.GetAddress(),
		},
		LoanBrokerID:       loanBrokerObjectID,
		PrincipalRequested: types.XRPLNumber("5000000"),
		Counterparty:       &counterparty,
		PaymentTotal:       &paymentTotal,
	}

	// Autofill + Sign by broker
	loanSetFlat := transaction.FlatTransaction(loanSetTx.Flatten())
	err = client.Autofill(&loanSetFlat)
	require.NoError(t, err)

	brokerBlob, _, err := loanBroker.Sign(loanSetFlat)
	require.NoError(t, err)

	// Sign by counterparty (borrower) using SignLoanSetByCounterpartyBlob
	_, counterpartyBlob, _, err := wallet.SignLoanSetByCounterpartyBlob(*borrower, brokerBlob, nil)
	require.NoError(t, err)
	require.NotEmpty(t, counterpartyBlob)

	// Submit the counterparty-signed transaction blob
	submitResp, err := client.SubmitTxBlob(counterpartyBlob, true)
	require.NoError(t, err)
	require.Equal(t, "tesSUCCESS", submitResp.EngineResult)
}
