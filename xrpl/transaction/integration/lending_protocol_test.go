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

	// Wallet 0: Vault Owner / Loan Broker
	vaultOwner := runner.GetWallet(0)
	loanBroker := vaultOwner
	// Wallet 1: Depositor
	depositor := runner.GetWallet(1)
	// Wallet 2: Borrower
	borrower := runner.GetWallet(2)

	// Create an XRP vault
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

	// Depositor deposits 10 XRP into the vault
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

	// Create the LoanBroker (Vault Owner acts as Loan Broker)
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

	// Broker initiates the LoanSet with counterparty single-signing
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

// TestIntegrationLoanSetWithMultisignCounterparty_Websocket tests the lending protocol lifecycle
// where the counterparty (borrower) is a multisig account. Two signers each sign the LoanSet
// as counterparty, and their signatures are combined using CombineLoanSetCounterpartySigners.
func TestIntegrationLoanSetWithMultisignCounterparty_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))

	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 6,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	// Wallet 0: Vault Owner / Loan Broker
	vaultOwner := runner.GetWallet(0)
	loanBroker := vaultOwner
	// Wallet 1: MPT issuer
	mptIssuer := runner.GetWallet(1)
	// Wallet 2: Depositor
	depositor := runner.GetWallet(2)
	// Wallet 3: Borrower (multisig account)
	borrower := runner.GetWallet(3)
	// Wallet 4 & 5: Signers for the borrower's multisig
	signer1 := runner.GetWallet(4)
	signer2 := runner.GetWallet(5)

	// Set up SignerList on borrower account (quorum = 2, each signer weight = 1)
	signerListSetTx := &transaction.SignerListSet{
		BaseTx: transaction.BaseTx{
			Account: borrower.GetAddress(),
		},
		SignerQuorum: uint32(2),
		SignerEntries: []ledger.SignerEntryWrapper{
			{
				SignerEntry: ledger.SignerEntry{
					Account:      types.Address(signer1.GetAddress()),
					SignerWeight: 1,
				},
			},
			{
				SignerEntry: ledger.SignerEntry{
					Account:      types.Address(signer2.GetAddress()),
					SignerWeight: 1,
				},
			},
		},
	}

	signerListFlat := signerListSetTx.Flatten()
	_, err = runner.TestTransaction(&signerListFlat, borrower, "tesSUCCESS", nil)
	require.NoError(t, err)

	// Create the MPT issuance
	mpttoken := &transaction.MPTokenIssuanceCreate{
		BaseTx: transaction.BaseTx{
			Account: mptIssuer.GetAddress(),
		},
	}
	mpttoken.SetMPTCanTransferFlag()
	mpttoken.SetMPTCanClawbackFlag()

	mptTokenFlat := transaction.FlatTransaction(mpttoken.Flatten())
	err = client.Autofill(&mptTokenFlat)
	require.NoError(t, err)

	mptTokenBlob, _, err := mptIssuer.Sign(mptTokenFlat)
	require.NoError(t, err)

	mptTokenTxResp, err := client.SubmitTxBlobAndWait(mptTokenBlob, true)
	require.NoError(t, err)

	mptTokenIssuanceId := mptTokenTxResp.Meta.MPTIssuanceID.String()
	require.NotEqual(t, nil, mptTokenIssuanceId)

	// Create an MPT-collateralized vault
	vaultCreateTx := &transaction.VaultCreate{
		BaseTx: transaction.BaseTx{
			Account: vaultOwner.GetAddress(),
		},
		Asset: ledger.Asset{
			MPTIssuanceID: mptTokenIssuanceId,
		},
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

	// MPTokenAuthorize for depositor
	mptAuthorizeTx := &transaction.MPTokenAuthorize{
		BaseTx: transaction.BaseTx{
			Account: depositor.GetAddress(),
		},
		MPTokenIssuanceID: mptTokenIssuanceId,
	}
	mptAuthorizeTxFlat := mptAuthorizeTx.Flatten()
	_, err = runner.TestTransaction(&mptAuthorizeTxFlat, depositor, "tesSUCCESS", nil)
	require.NoError(t, err)

	// Fund the depositor with MPT tokens so they can deposit into the vault
	paymentTx := &transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account: mptIssuer.GetAddress(),
		},
		Destination: depositor.GetAddress(),
		Amount: types.MPTCurrencyAmount{
			MPTIssuanceID: mptTokenIssuanceId,
			Value:         "500000",
		},
	}
	paymentTxFlat := paymentTx.Flatten()
	_, err = runner.TestTransaction(&paymentTxFlat, mptIssuer, "tesSUCCESS", nil)
	require.NoError(t, err)

	// MPTokenAuthorize for loanBroker
	loanBrokerMptAuthorizeTx := &transaction.MPTokenAuthorize{
		BaseTx: transaction.BaseTx{
			Account: loanBroker.GetAddress(),
		},
		MPTokenIssuanceID: mptTokenIssuanceId,
	}
	loanBrokerMptAuthorizeTxFlat := loanBrokerMptAuthorizeTx.Flatten()
	_, err = runner.TestTransaction(&loanBrokerMptAuthorizeTxFlat, loanBroker, "tesSUCCESS", nil)
	require.NoError(t, err)

	// Fund the loan broker with MPT tokens to cover the loan principal
	loanBrokerPaymentTx := &transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account: mptIssuer.GetAddress(),
		},
		Destination: loanBroker.GetAddress(),
		Amount: types.MPTCurrencyAmount{
			MPTIssuanceID: mptTokenIssuanceId,
			Value:         "500000",
		},
	}
	loanBrokerPaymentTxFlat := loanBrokerPaymentTx.Flatten()
	_, err = runner.TestTransaction(&loanBrokerPaymentTxFlat, mptIssuer, "tesSUCCESS", nil)
	require.NoError(t, err)

	// Deposit MPT into vault
	depositAmount := "200000"
	vaultDepositTx := &transaction.VaultDeposit{
		BaseTx: transaction.BaseTx{
			Account: depositor.GetAddress(),
		},
		VaultID: types.Hash256(vaultObjectID),
		Amount: types.MPTCurrencyAmount{
			MPTIssuanceID: mptTokenIssuanceId,
			Value:         depositAmount,
		},
	}
	vaultDepositFlat := vaultDepositTx.Flatten()
	_, err = runner.TestTransaction(&vaultDepositFlat, depositor, "tesSUCCESS", nil)
	require.NoError(t, err)

	// Create the LoanBroker (Vault Owner acts as Loan Broker)
	debtMaximum := types.XRPLNumber("100000")
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

	loanBrokerTxResp, err := client.SubmitTxBlobAndWait(loanBrokerBlob, true)
	require.NoError(t, err)

	// Compute the LoanBroker hash from the response
	loanBrokerAccount, ok := loanBrokerTxResp.TxJSON["Account"].(string)
	require.True(t, ok)
	loanBrokerSequence, err := sequenceFromTx(loanBrokerTxResp.TxJSON)
	require.NoError(t, err)

	loanBrokerObjectID, err := xrplhash.LoanBroker(loanBrokerAccount, loanBrokerSequence)
	require.NoError(t, err)

	// Broker initiates the LoanSet with multisig counterparty signing
	counterparty := borrower.GetAddress()
	paymentTotal := types.PaymentTotal(1)
	interestRate := types.InterestRate(0)
	loanSetTx := &transaction.LoanSet{
		BaseTx: transaction.BaseTx{
			Account: loanBroker.GetAddress(),
		},
		LoanBrokerID:       loanBrokerObjectID,
		PrincipalRequested: types.XRPLNumber("100000"),
		InterestRate:       &interestRate,
		Counterparty:       &counterparty,
		PaymentTotal:       &paymentTotal,
	}

	// Negative test: submit LoanSet without counterparty signature, expect temBAD_SIGNER
	negTestFlat := transaction.FlatTransaction(loanSetTx.Flatten())
	_, err = runner.TestTransaction(&negTestFlat, loanBroker, "temBAD_SIGNER", nil)
	require.NoError(t, err)

	// Autofill + Sign by broker (fresh flat to avoid stale fields from negative test)
	loanSetFlat := transaction.FlatTransaction(loanSetTx.Flatten())
	err = client.Autofill(&loanSetFlat)
	require.NoError(t, err)

	brokerBlob, _, err := loanBroker.Sign(loanSetFlat)
	require.NoError(t, err)

	// Each signer signs the LoanSet as counterparty multisig
	multisignOpts := &wallet.SignLoanSetByCounterpartyOptions{
		Multisign: true,
	}

	signer1Tx, signer1Blob, _, err := wallet.SignLoanSetByCounterpartyBlob(*signer1, brokerBlob, multisignOpts)
	require.NoError(t, err)
	require.NotEmpty(t, signer1Blob)

	signer2Tx, signer2Blob, _, err := wallet.SignLoanSetByCounterpartyBlob(*signer2, brokerBlob, multisignOpts)
	require.NoError(t, err)
	require.NotEmpty(t, signer2Blob)

	// Combine the two counterparty multisig signatures
	combinedTx, combinedBlob, err := wallet.CombineLoanSetCounterpartySigners([]transaction.FlatTransaction{signer1Tx, signer2Tx})
	require.NoError(t, err)
	require.NotNil(t, combinedTx)
	require.NotEmpty(t, combinedBlob)

	// Submit the combined transaction blob
	submitResp, err := client.SubmitTxBlob(combinedBlob, true)
	require.NoError(t, err)
	require.Equal(t, "tesSUCCESS", submitResp.EngineResult)
}
