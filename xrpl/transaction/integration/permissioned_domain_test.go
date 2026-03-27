package integration

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

type PermissionedDomainTest struct {
	Name                  string
	PermissionedDomainSet *transaction.PermissionedDomainSet
	ExpectedError         string
}

func permissionedDomainTest(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 1})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	owner := runner.GetWallet(0)

	tt := []PermissionedDomainTest{
		{
			Name: "pass - lifecycle",
			PermissionedDomainSet: &transaction.PermissionedDomainSet{
				BaseTx: transaction.BaseTx{Account: owner.GetAddress()},
				AcceptedCredentials: types.AuthorizeCredentialList{
					{
						Credential: types.Credential{
							Issuer:         owner.GetAddress(),
							CredentialType: types.CredentialType(hex.EncodeToString([]byte("Passport"))),
						},
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			flat := tc.PermissionedDomainSet.Flatten()
			_, err = runner.TestTransaction(&flat, owner, "tesSUCCESS", nil)
			require.NoError(t, err)

			objects, err := client.GetAccountObjects(&account.ObjectsRequest{
				Account: owner.GetAddress(),
				Type:    account.PermissionedDomainObject,
			})
			require.NoError(t, err)
			require.Len(t, objects.AccountObjects, 1, "there should be exactly one permissioned domain on the ledger")
			domainID := objects.AccountObjects[0]["index"].(string)

			pdDeleteTx := &transaction.PermissionedDomainDelete{
				BaseTx:   transaction.BaseTx{Account: owner.GetAddress()},
				DomainID: domainID,
			}
			flat = pdDeleteTx.Flatten()
			_, err = runner.TestTransaction(&flat, owner, "tesSUCCESS", nil)
			require.NoError(t, err)

			objects, err = client.GetAccountObjects(&account.ObjectsRequest{
				Account: owner.GetAddress(),
			})
			require.NoError(t, err)
			require.Empty(t, objects.AccountObjects, "permissioned domain should have been deleted from account objects")
		})
	}
}

func TestIntegrationPermissionedDomain_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	permissionedDomainTest(t, client)
}

func TestIntegrationPermissionedDomain_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	permissionedDomainTest(t, client)
}
