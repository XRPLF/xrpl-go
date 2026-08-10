package escrow

import (
	"encoding/json"
	"strconv"
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

func testIntegrationEscrowCreate(t *testing.T, client integration.Client) {
	t.Run("pass - base escrow create", func(t *testing.T) {
		runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 2})
		err := runner.Setup()
		require.NoError(t, err)
		defer runner.Teardown()

		sender := runner.GetWallet(0)
		receiver := runner.GetWallet(1)
		closeTime := getLedgerCloseTime(t, client)

		escrowCreateTx := &transaction.EscrowCreate{
			BaseTx:      transaction.BaseTx{Account: sender.GetAddress()},
			Amount:      types.XRPCurrencyAmount(10000),
			Destination: receiver.GetAddress(),
			FinishAfter: uint32(closeTime + 2),
		}
		flatEscrowCreateTx := escrowCreateTx.Flatten()
		_, err = runner.TestTransaction(&flatEscrowCreateTx, sender, "tesSUCCESS", nil)
		require.NoError(t, err)

		objects, err := client.GetAccountObjects(&account.ObjectsRequest{
			Account: sender.GetAddress(),
			Type:    account.EscrowObject,
		})
		require.NoError(t, err)
		require.Len(t, objects.AccountObjects, 1)

		escrowEntry := integration.DecodeLedgerObject[ledger.Escrow](t, objects.AccountObjects[0])
		require.Equal(t, ledger.EscrowEntry, escrowEntry.LedgerEntryType)
		require.Equal(t, sender.GetAddress(), escrowEntry.Account)
		require.Equal(t, receiver.GetAddress(), escrowEntry.Destination)
		require.Equal(t, types.XRPCurrencyAmount(10000), escrowEntry.Amount)
		require.NotEmpty(t, escrowEntry.OwnerNode)
		require.Empty(t, escrowEntry.IssuerNode, "XRP escrows carry no IssuerNode")
	})

	t.Run("pass - token escrow issuer node uses nonzero server string", func(t *testing.T) {
		runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 3})
		err := runner.Setup()
		require.NoError(t, err)
		defer runner.Teardown()

		issuer := runner.GetWallet(0)
		sender := runner.GetWallet(1)
		receiver := runner.GetWallet(2)

		accountSetTx := &transaction.AccountSet{
			BaseTx: transaction.BaseTx{Account: issuer.GetAddress()},
		}
		accountSetTx.SetAsfAllowTrustLineLocking()
		flatAccountSetTx := accountSetTx.Flatten()
		_, err = runner.TestTransaction(&flatAccountSetTx, issuer, "tesSUCCESS", nil)
		require.NoError(t, err)

		// An owner-directory page holds 32 entries. The extra ticket forces
		// subsequent issuer-owned objects onto the second page, with hint 1.
		const ticketCount = 33
		ticketCreateTx := (&transaction.TicketCreate{
			BaseTx:      transaction.BaseTx{Account: issuer.GetAddress()},
			TicketCount: ticketCount,
		}).Flatten()
		_, err = runner.TestTransaction(&ticketCreateTx, issuer, "tesSUCCESS", nil)
		require.NoError(t, err)

		tickets, err := client.GetAccountObjects(&account.ObjectsRequest{
			Account: issuer.GetAddress(),
			Type:    account.TicketObject,
		})
		require.NoError(t, err)
		require.Len(t, tickets.AccountObjects, ticketCount)

		trustSetTx := &transaction.TrustSet{
			BaseTx: transaction.BaseTx{Account: sender.GetAddress()},
			LimitAmount: types.IssuedCurrencyAmount{
				Currency: "USD",
				Issuer:   issuer.GetAddress(),
				Value:    "1000",
			},
		}
		flatTrustSetTx := trustSetTx.Flatten()
		_, err = runner.TestTransaction(&flatTrustSetTx, sender, "tesSUCCESS", nil)
		require.NoError(t, err)

		paymentTx := &transaction.Payment{
			BaseTx:      transaction.BaseTx{Account: issuer.GetAddress()},
			Destination: sender.GetAddress(),
			Amount: types.IssuedCurrencyAmount{
				Currency: "USD",
				Issuer:   issuer.GetAddress(),
				Value:    "100",
			},
		}
		flatPaymentTx := paymentTx.Flatten()
		_, err = runner.TestTransaction(&flatPaymentTx, issuer, "tesSUCCESS", nil)
		require.NoError(t, err)

		closeTime := getLedgerCloseTime(t, client)
		escrowCreateTx := &transaction.EscrowCreate{
			BaseTx: transaction.BaseTx{Account: sender.GetAddress()},
			Amount: types.IssuedCurrencyAmount{
				Currency: "USD",
				Issuer:   issuer.GetAddress(),
				Value:    "10",
			},
			Destination: receiver.GetAddress(),
			FinishAfter: uint32(closeTime + 20),
			CancelAfter: uint32(closeTime + 300),
		}
		flatEscrowCreateTx := escrowCreateTx.Flatten()
		_, err = runner.TestTransaction(&flatEscrowCreateTx, sender, "tesSUCCESS", nil)
		require.NoError(t, err)

		objects, err := client.GetAccountObjects(&account.ObjectsRequest{
			Account: sender.GetAddress(),
			Type:    account.EscrowObject,
		})
		require.NoError(t, err)
		require.Len(t, objects.AccountObjects, 1)

		rawEscrow := objects.AccountObjects[0]
		issuerNode, ok := rawEscrow["IssuerNode"].(string)
		require.True(t, ok, "rippled must return IssuerNode as a JSON string, got %T", rawEscrow["IssuerNode"])
		require.NotEmpty(t, issuerNode)
		parsedIssuerNode, err := strconv.ParseUint(issuerNode, 16, 64)
		require.NoError(t, err, "IssuerNode must contain a hexadecimal UInt64")
		require.Equal(t, uint64(1), parsedIssuerNode)
		require.Equal(t, "1", issuerNode, "rippled must encode the nonzero UInt64 page hint as a string")

		encodedEscrow, err := json.Marshal(rawEscrow)
		require.NoError(t, err)

		var previousModel struct {
			IssuerNode uint64
		}
		err = json.Unmarshal(encodedEscrow, &previousModel)
		require.Error(t, err, "the previous uint64 field must reject rippled's quoted IssuerNode")
		var typeError *json.UnmarshalTypeError
		require.ErrorAs(t, err, &typeError)
		require.Equal(t, "IssuerNode", typeError.Field)
		require.Equal(t, "string", typeError.Value)
		require.Equal(t, "uint64", typeError.Type.String())

		var escrowEntry ledger.Escrow
		require.NoError(t, json.Unmarshal(encodedEscrow, &escrowEntry))
		require.Equal(t, issuerNode, escrowEntry.IssuerNode)
		require.Equal(t, sender.GetAddress(), escrowEntry.Account)
		require.Equal(t, receiver.GetAddress(), escrowEntry.Destination)
	})
}

func TestIntegrationEscrowCreate_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationEscrowCreate(t, client)
}

func TestIntegrationEscrowCreate_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationEscrowCreate(t, client)
}
