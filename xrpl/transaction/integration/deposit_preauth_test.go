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
	ExpectedError  string
}

func depositPreauthBaseTest(t *testing.T, client integration.Client) {
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
			flat := tc.DepositPreauth.Flatten()
			_, err := runner.TestTransaction(&flat, account0, "tesSUCCESS", nil)
			if tc.ExpectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.ExpectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func depositPreauthAuthorizeCredentialTest(t *testing.T, client integration.Client) {
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
		flat := credCreateTx.Flatten()
		_, err = runner.TestTransaction(&flat, issuer, "tesSUCCESS", nil)
		require.NoError(t, err)

		credAcceptTx := &transaction.CredentialAccept{
			BaseTx:         transaction.BaseTx{Account: subject.GetAddress()},
			Issuer:         issuer.GetAddress(),
			CredentialType: credType,
		}
		flat = credAcceptTx.Flatten()
		_, err = runner.TestTransaction(&flat, subject, "tesSUCCESS", nil)
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
		flat = depositPreauthTx.Flatten()
		_, err = runner.TestTransaction(&flat, issuer, "tesSUCCESS", nil)
		require.NoError(t, err)
	})
}

func depositPreauthUnauthorizeCredentialTest(t *testing.T, client integration.Client) {
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
		flat := credCreateTx.Flatten()
		_, err = runner.TestTransaction(&flat, issuer, "tesSUCCESS", nil)
		require.NoError(t, err)

		credAcceptTx := &transaction.CredentialAccept{
			BaseTx:         transaction.BaseTx{Account: subject.GetAddress()},
			Issuer:         issuer.GetAddress(),
			CredentialType: credType,
		}
		flat = credAcceptTx.Flatten()
		_, err = runner.TestTransaction(&flat, subject, "tesSUCCESS", nil)
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
		flat = authTx.Flatten()
		_, err = runner.TestTransaction(&flat, issuer, "tesSUCCESS", nil)
		require.NoError(t, err)

		unauthTx := &transaction.DepositPreauth{
			BaseTx:                 transaction.BaseTx{Account: issuer.GetAddress()},
			UnauthorizeCredentials: []types.AuthorizeCredentialsWrapper{authorizeCredential},
		}
		flat = unauthTx.Flatten()
		_, err = runner.TestTransaction(&flat, issuer, "tesSUCCESS", nil)
		require.NoError(t, err)
	})
}

func TestIntegrationDepositPreauth_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	depositPreauthBaseTest(t, client)
	depositPreauthAuthorizeCredentialTest(t, client)
	depositPreauthUnauthorizeCredentialTest(t, client)
}

func TestIntegrationDepositPreauth_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	depositPreauthBaseTest(t, client)
	depositPreauthAuthorizeCredentialTest(t, client)
	depositPreauthUnauthorizeCredentialTest(t, client)
}
