package credential

import (
	"encoding/hex"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

type CredentialAcceptTest struct {
	Name             string
	CredentialCreate *transaction.CredentialCreate
	CredentialAccept *transaction.CredentialAccept
}

func testIntegrationCredentialAccept(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 2,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	subject := runner.GetWallet(1)
	tt := []CredentialAcceptTest{
		{
			Name: "pass - credential accept",
			CredentialCreate: &transaction.CredentialCreate{
				BaseTx: transaction.BaseTx{
					Account: issuer.GetAddress(),
				},
				Subject:        subject.GetAddress(),
				CredentialType: types.CredentialType(hex.EncodeToString([]byte("Test Credential Type"))),
			},
			CredentialAccept: &transaction.CredentialAccept{
				BaseTx: transaction.BaseTx{
					Account: subject.GetAddress(),
				},
				Issuer:         issuer.GetAddress(),
				CredentialType: types.CredentialType(hex.EncodeToString([]byte("Test Credential Type"))),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			flatCreateTx := tc.CredentialCreate.Flatten()
			_, err := runner.TestTransaction(&flatCreateTx, issuer, "tesSUCCESS", nil)
			require.NoError(t, err)

			accountObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
				Account: subject.GetAddress(),
				Type:    account.CredentialObject,
			})
			require.NoError(t, err)
			require.Len(t, accountObjects.AccountObjects, 1)

			flatCredentialAcceptTx := tc.CredentialAccept.Flatten()
			_, err = runner.TestTransaction(&flatCredentialAcceptTx, subject, "tesSUCCESS", nil)
			require.NoError(t, err)
			accountObjects, err = client.GetAccountObjects(&account.ObjectsRequest{
				Account: subject.GetAddress(),
				Type:    account.CredentialObject,
			})
			require.NoError(t, err)
			require.Len(t, accountObjects.AccountObjects, 1)
		})
	}
}

func TestIntegrationCredentialAccept_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationCredentialAccept(t, client)
}

func TestIntegrationCredentialAccept_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationCredentialAccept(t, client)
}
