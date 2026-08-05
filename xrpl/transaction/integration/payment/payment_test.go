package payment

import (
	"os"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/hash"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

func testIntegrationPayment(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 2})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	sender := runner.GetWallet(0)
	receiver := runner.GetWallet(1)

	t.Run("pass - base payment", func(t *testing.T) {
		paymentTx := &transaction.Payment{
			BaseTx:      transaction.BaseTx{Account: sender.GetAddress()},
			Amount:      types.XRPCurrencyAmount(1000),
			Destination: receiver.GetAddress(),
		}
		flatPaymentTx := paymentTx.Flatten()
		_, err = runner.TestTransaction(&flatPaymentTx, sender, "tesSUCCESS", nil)
		require.NoError(t, err)
	})

	t.Run("pass - X-address autofill and network identity", func(t *testing.T) {
		testnetAddress := os.Getenv(integration.IntegrationEnvVar) != string(integration.LocalnetEnv)
		sourceXAddress, err := addresscodec.ClassicAddressToXAddress(
			sender.GetAddress().String(),
			0,
			true,
			testnetAddress,
		)
		require.NoError(t, err)
		destinationXAddress, err := addresscodec.ClassicAddressToXAddress(
			receiver.GetAddress().String(),
			14,
			true,
			testnetAddress,
		)
		require.NoError(t, err)

		paymentTx := &transaction.Payment{
			BaseTx:      transaction.BaseTx{Account: types.Address(sourceXAddress)},
			Amount:      types.XRPCurrencyAmount(1000),
			Destination: types.Address(destinationXAddress),
		}
		flatPaymentTx := paymentTx.Flatten()
		_, err = runner.TestTransaction(&flatPaymentTx, sender, "tesSUCCESS", nil)
		require.NoError(t, err)
		networkID := discoveredNetworkID(t, client)
		require.Equal(t, sender.GetAddress().String(), flatPaymentTx["Account"])
		require.Equal(t, uint32(0), flatPaymentTx["SourceTag"])
		require.Equal(t, receiver.GetAddress().String(), flatPaymentTx["Destination"])
		require.Equal(t, uint32(14), flatPaymentTx["DestinationTag"])
		if networkID > rpc.RestrictedNetworks {
			require.Equal(t, networkID, flatPaymentTx["NetworkID"])
		} else {
			require.NotContains(t, flatPaymentTx, "NetworkID")
		}
	})

	t.Run("pass - payment specifying amount field", func(t *testing.T) {
		paymentTx := &transaction.Payment{
			BaseTx:      transaction.BaseTx{Account: sender.GetAddress()},
			Amount:      types.XRPCurrencyAmount(1000),
			Destination: receiver.GetAddress(),
		}
		flatPaymentTx := paymentTx.Flatten()
		res, err := runner.TestTransaction(&flatPaymentTx, sender, "tesSUCCESS", nil)
		require.NoError(t, err)
		require.Equal(t, 0, res.EngineResultCode)
		require.Equal(t, "1000", res.Tx["Amount"])
	})

	t.Run("pass - payment specifying delivery max field", func(t *testing.T) {
		paymentTx := &transaction.Payment{
			BaseTx:      transaction.BaseTx{Account: sender.GetAddress()},
			DeliverMax:  types.XRPCurrencyAmount(1000),
			Destination: receiver.GetAddress(),
		}
		flatPaymentTx := paymentTx.Flatten()
		res, err := runner.TestTransaction(&flatPaymentTx, sender, "tesSUCCESS", nil)
		require.NoError(t, err)
		require.Equal(t, 0, res.EngineResultCode)
		require.Equal(t, "1000", res.Tx["Amount"])
	})

	t.Run("pass - payment with identical DeliverMax and Amount fields", func(t *testing.T) {
		paymentTx := &transaction.Payment{
			BaseTx:      transaction.BaseTx{Account: sender.GetAddress()},
			Amount:      types.XRPCurrencyAmount(1000),
			DeliverMax:  types.XRPCurrencyAmount(1000),
			Destination: receiver.GetAddress(),
		}
		flatPaymentTx := paymentTx.Flatten()
		res, err := runner.TestTransaction(&flatPaymentTx, sender, "tesSUCCESS", nil)
		require.NoError(t, err)
		require.Equal(t, 0, res.EngineResultCode)
		require.Equal(t, "1000", res.Tx["Amount"])
	})

	t.Run("pass - MPT payment", func(t *testing.T) {
		issuer := sender
		holder := receiver

		mptIssuanceCreateTx := &transaction.MPTokenIssuanceCreate{
			BaseTx: transaction.BaseTx{Account: issuer.GetAddress()},
		}
		flatMptIssuanceCreateTx := mptIssuanceCreateTx.Flatten()
		res, err := runner.TestTransaction(&flatMptIssuanceCreateTx, issuer, "tesSUCCESS", nil)
		require.NoError(t, err)

		sequence := integration.TxFieldUint32(t, res.Tx, "Sequence")

		mptID, err := hash.MPTID(sequence, issuer.GetAddress().String())

		objects, err := client.GetAccountObjects(&account.ObjectsRequest{
			Account: issuer.GetAddress(),
			Type:    account.MPTIssuanceObject,
		})
		require.NoError(t, err)
		require.Len(t, objects.AccountObjects, 1, "should be exactly one issuance on the ledger")

		mptTokenAuthTx := &transaction.MPTokenAuthorize{
			BaseTx:            transaction.BaseTx{Account: holder.GetAddress()},
			MPTokenIssuanceID: mptID,
		}
		flatMptTokenAuthTx := mptTokenAuthTx.Flatten()
		_, err = runner.TestTransaction(&flatMptTokenAuthTx, holder, "tesSUCCESS", nil)
		require.NoError(t, err)

		objects, err = client.GetAccountObjects(&account.ObjectsRequest{
			Account: holder.GetAddress(),
			Type:    account.MPTokenObject,
		})
		require.NoError(t, err)
		require.Len(t, objects.AccountObjects, 1, "holder owns 1 MPToken on the ledger")

		paymentTx := &transaction.Payment{
			BaseTx: transaction.BaseTx{Account: issuer.GetAddress()},
			Amount: types.MPTCurrencyAmount{
				MPTIssuanceID: mptID,
				Value:         "100",
			},
			Destination: holder.GetAddress(),
		}
		flatPaymentTx := paymentTx.Flatten()
		_, err = runner.TestTransaction(&flatPaymentTx, issuer, "tesSUCCESS", nil)
		require.NoError(t, err)

		objects, err = client.GetAccountObjects(&account.ObjectsRequest{
			Account: issuer.GetAddress(),
			Type:    account.MPTIssuanceObject,
		})
		require.NoError(t, err)
		require.Equal(t, "100", objects.AccountObjects[0]["OutstandingAmount"])
	})
}

func discoveredNetworkID(t *testing.T, client integration.Client) uint32 {
	t.Helper()

	var networkID *uint32
	switch cl := client.(type) {
	case *rpc.Client:
		networkID = cl.NetworkID
	case *websocket.Client:
		networkID = cl.NetworkID
	default:
		t.Fatalf("unsupported integration client type %T", client)
	}
	require.NotNil(t, networkID)
	return *networkID
}

func TestIntegrationPayment_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationPayment(t, client)
}

func TestIntegrationPayment_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationPayment(t, client)
}
