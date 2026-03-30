package integration

import (
	"encoding/hex"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

type DepositPreauthTest struct {
	Name           string
	DepositPreauth *transaction.DepositPreauth
}

func testIntegrationDepositPreauthBase(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 2})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	account0 := runner.GetWallet(0)
	account1 := runner.GetWallet(1)

	tt := []DepositPreauthTest{
		{
			Name: "pass - base",
			DepositPreauth: &transaction.DepositPreauth{
				BaseTx:    transaction.BaseTx{Account: account0.GetAddress()},
				Authorize: account1.GetAddress(),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			flatDepositPreauthTx := tc.DepositPreauth.Flatten()
			_, err := runner.TestTransaction(&flatDepositPreauthTx, account0, "tesSUCCESS", nil)
			require.NoError(t, err)
		})
	}
}

func testIntegrationDepositPreauthAuthCredential(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 2})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	subject := runner.GetWallet(1)

	credType := types.CredentialType(hex.EncodeToString([]byte("credential_type")))
	t.Run("pass - authorizeCredential base case", func(t *testing.T) {
		credCreateTx := &transaction.CredentialCreate{
			BaseTx:         transaction.BaseTx{Account: issuer.GetAddress()},
			Subject:        subject.GetAddress(),
			CredentialType: credType,
		}
		flatCredentialCreateTx := credCreateTx.Flatten()
		_, err = runner.TestTransaction(&flatCredentialCreateTx, issuer, "tesSUCCESS", nil)
		require.NoError(t, err)

		credAcceptTx := &transaction.CredentialAccept{
			BaseTx:         transaction.BaseTx{Account: subject.GetAddress()},
			Issuer:         issuer.GetAddress(),
			CredentialType: credType,
		}
		flatCredentialAcceptTx := credAcceptTx.Flatten()
		_, err = runner.TestTransaction(&flatCredentialAcceptTx, subject, "tesSUCCESS", nil)
		require.NoError(t, err)

		depositPreauthTx := &transaction.DepositPreauth{
			BaseTx: transaction.BaseTx{Account: issuer.GetAddress()},
			AuthorizeCredentials: []types.AuthorizeCredentialsWrapper{
				{
					Credential: types.AuthorizeCredentials{
						Issuer:         issuer.GetAddress(),
						CredentialType: credType,
					},
				},
			},
		}
		flatDepositPreauthTx := depositPreauthTx.Flatten()
		_, err = runner.TestTransaction(&flatDepositPreauthTx, issuer, "tesSUCCESS", nil)
		require.NoError(t, err)
	})
}

func testIntegrationDepositPreauthUnauthCredential(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 2})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	subject := runner.GetWallet(1)

	credType := types.CredentialType(hex.EncodeToString([]byte("credential_type")))

	t.Run("pass - unauthorizeCredential base case", func(t *testing.T) {
		credCreateTx := &transaction.CredentialCreate{
			BaseTx:         transaction.BaseTx{Account: issuer.GetAddress()},
			Subject:        subject.GetAddress(),
			CredentialType: credType,
		}
		flatCredentialCreateTx := credCreateTx.Flatten()
		_, err = runner.TestTransaction(&flatCredentialCreateTx, issuer, "tesSUCCESS", nil)
		require.NoError(t, err)

		credAcceptTx := &transaction.CredentialAccept{
			BaseTx:         transaction.BaseTx{Account: subject.GetAddress()},
			Issuer:         issuer.GetAddress(),
			CredentialType: credType,
		}
		flatCredentialAcceptTx := credAcceptTx.Flatten()
		_, err = runner.TestTransaction(&flatCredentialAcceptTx, subject, "tesSUCCESS", nil)
		require.NoError(t, err)

		authorizeCredential := types.AuthorizeCredentialsWrapper{
			Credential: types.AuthorizeCredentials{
				Issuer:         issuer.GetAddress(),
				CredentialType: credType,
			},
		}

		authTx := &transaction.DepositPreauth{
			BaseTx:               transaction.BaseTx{Account: issuer.GetAddress()},
			AuthorizeCredentials: []types.AuthorizeCredentialsWrapper{authorizeCredential},
		}
		flatAuthTx := authTx.Flatten()
		_, err = runner.TestTransaction(&flatAuthTx, issuer, "tesSUCCESS", nil)
		require.NoError(t, err)

		unauthTx := &transaction.DepositPreauth{
			BaseTx:                 transaction.BaseTx{Account: issuer.GetAddress()},
			UnauthorizeCredentials: []types.AuthorizeCredentialsWrapper{authorizeCredential},
		}
		flatUnauthTx := unauthTx.Flatten()
		_, err = runner.TestTransaction(&flatUnauthTx, issuer, "tesSUCCESS", nil)
		require.NoError(t, err)
	})
}

func TestIntegrationDepositPreauth_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationDepositPreauthBase(t, client)
	testIntegrationDepositPreauthAuthCredential(t, client)
	testIntegrationDepositPreauthUnauthCredential(t, client)
}

func TestIntegrationDepositPreauth_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationDepositPreauthBase(t, client)
	testIntegrationDepositPreauthAuthCredential(t, client)
	testIntegrationDepositPreauthUnauthCredential(t, client)
}
