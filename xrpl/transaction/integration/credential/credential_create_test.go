package credential

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

type CredentialCreateTest struct {
	Name             string
	CredentialCreate *transaction.CredentialCreate
}

func testIntegrationCredentialCreate(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 2,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	subject := runner.GetWallet(1)

	credentialType := types.CredentialType(hex.EncodeToString([]byte("Test Credential Type")))

	tt := []CredentialCreateTest{
		{
			Name: "pass - credential accept",
			CredentialCreate: &transaction.CredentialCreate{
				BaseTx: transaction.BaseTx{
					Account: issuer.GetAddress(),
				},
				Subject:        subject.GetAddress(),
				CredentialType: credentialType,
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
			receivedCredentialType := accountObjects.AccountObjects[0]["CredentialType"].(string)
			require.Equal(t, strings.ToLower(credentialType.String()), strings.ToLower(receivedCredentialType))
			require.Equal(t, accountObjects.AccountObjects[0]["Subject"].(string), subject.GetAddress().String())
		})
	}
}

func TestIntegrationCredentialCreate_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationCredentialCreate(t, client)
}

func TestIntegrationCredentialCreate_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationCredentialCreate(t, client)
}
