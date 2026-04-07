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

type CredentialDeleteTest struct {
	Name             string
	CredentialCreate *transaction.CredentialCreate
	CredentialAccept *transaction.CredentialAccept
	CredentialDelete *transaction.CredentialDelete
}

func testIntegrationCredentialDelete(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 2,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	subject := runner.GetWallet(1)
	tt := []CredentialDeleteTest{
		{
			Name: "pass - credential delete",
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
			CredentialDelete: &transaction.CredentialDelete{
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

			flatCredentialDeleteTx := tc.CredentialDelete.Flatten()
			_, err = runner.TestTransaction(&flatCredentialDeleteTx, subject, "tesSUCCESS", nil)
			require.NoError(t, err)

			subjectObjectsFinal, err := client.GetAccountObjects(&account.ObjectsRequest{
				Account: subject.GetAddress(),
				Type:    account.CredentialObject,
			})
			require.NoError(t, err)
			require.Empty(t, subjectObjectsFinal.AccountObjects)
			issuerObjectsFinal, err := client.GetAccountObjects(&account.ObjectsRequest{
				Account: issuer.GetAddress(),
				Type:    account.CredentialObject,
			})
			require.NoError(t, err)
			require.Empty(t, issuerObjectsFinal.AccountObjects)
		})
	}
}

func TestIntegrationCredentialDelete_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationCredentialDelete(t, client)
}

func TestIntegrationCredentialDelete_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationCredentialDelete(t, client)
}
