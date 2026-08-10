package mpt

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/hash"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

func testIntegrationMPTokenClawback(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 2})
	require.NoError(t, runner.Setup())
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	holder := runner.GetWallet(1)

	issuanceCreate := &transaction.MPTokenIssuanceCreate{
		BaseTx: transaction.BaseTx{Account: issuer.GetAddress()},
	}
	issuanceCreate.SetMPTCanClawbackFlag()
	flatIssuanceCreate := issuanceCreate.Flatten()
	createResponse, err := runner.TestTransaction(&flatIssuanceCreate, issuer, "tesSUCCESS", nil)
	require.NoError(t, err)

	sequence := integration.TxFieldUint32(t, createResponse.Tx, "Sequence")
	mptID, err := hash.MPTID(sequence, issuer.GetAddress().String())
	require.NoError(t, err)

	authorize := transaction.MPTokenAuthorize{
		BaseTx:            transaction.BaseTx{Account: holder.GetAddress()},
		MPTokenIssuanceID: mptID,
	}
	flatAuthorize := authorize.Flatten()
	_, err = runner.TestTransaction(&flatAuthorize, holder, "tesSUCCESS", nil)
	require.NoError(t, err)

	const issuedAmount = "100"
	payment := transaction.Payment{
		BaseTx:      transaction.BaseTx{Account: issuer.GetAddress()},
		Destination: holder.GetAddress(),
		Amount: types.MPTCurrencyAmount{
			MPTIssuanceID: mptID,
			Value:         issuedAmount,
		},
	}
	flatPayment := payment.Flatten()
	_, err = runner.TestTransaction(&flatPayment, issuer, "tesSUCCESS", nil)
	require.NoError(t, err)

	holderObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account: holder.GetAddress(),
		Type:    account.MPTokenObject,
	})
	require.NoError(t, err)
	require.Len(t, holderObjects.AccountObjects, 1)
	heldToken := integration.DecodeLedgerObject[ledger.MPToken](t, holderObjects.AccountObjects[0])
	require.Equal(t, issuedAmount, heldToken.MPTAmount)

	const clawbackAmount = "40"
	clawback := transaction.Clawback{
		BaseTx: transaction.BaseTx{Account: issuer.GetAddress()},
		Amount: types.MPTCurrencyAmount{
			MPTIssuanceID: mptID,
			Value:         clawbackAmount,
		},
		Holder: holder.GetAddress(),
	}
	flatClawback := clawback.Flatten()
	_, err = runner.TestTransaction(&flatClawback, issuer, "tesSUCCESS", nil)
	require.NoError(t, err)

	holderObjects, err = client.GetAccountObjects(&account.ObjectsRequest{
		Account: holder.GetAddress(),
		Type:    account.MPTokenObject,
	})
	require.NoError(t, err)
	require.Len(t, holderObjects.AccountObjects, 1)
	heldToken = integration.DecodeLedgerObject[ledger.MPToken](t, holderObjects.AccountObjects[0])
	require.Equal(t, "60", heldToken.MPTAmount)

	issuanceObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account: issuer.GetAddress(),
		Type:    account.MPTIssuanceObject,
	})
	require.NoError(t, err)
	require.Len(t, issuanceObjects.AccountObjects, 1)
	issuance := integration.DecodeLedgerObject[ledger.MPTokenIssuance](t, issuanceObjects.AccountObjects[0])
	require.Equal(t, "60", issuance.OutstandingAmount)
}

func TestIntegrationMPTokenClawback_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationMPTokenClawback(t, client)
}

func TestIntegrationMPTokenClawback_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationMPTokenClawback(t, client)
}
