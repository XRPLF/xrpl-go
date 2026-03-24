package integration

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

type TrustSet struct {
	Name          string
	TrustSet      *transaction.TrustSet
	ExpectedError string
}

func TrustSetTest(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 2,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	receiver := runner.GetWallet(1)

	tt := []TrustSet{
		{
			Name: "base trust set",
			TrustSet: &transaction.TrustSet{
				BaseTx: transaction.BaseTx{
					Account: issuer.GetAddress(),
				},

				LimitAmount: types.IssuedCurrencyAmount{
					Currency: "USD",
					Issuer:   receiver.GetAddress(),
					Value:    "100000000000000",
				},
			},
		},
		{
			Name: "trust set - quality < 1",
			TrustSet: &transaction.TrustSet{
				BaseTx: transaction.BaseTx{
					Account: issuer.GetAddress(),
				},
				LimitAmount: types.IssuedCurrencyAmount{
					Currency: "USD",
					Issuer:   receiver.GetAddress(),
					Value:    "100000000000000",
				},
				QualityIn:  990000000,
				QualityOut: 990000000,
			},
		},
		{
			Name: "trust set - quality > 1",
			TrustSet: &transaction.TrustSet{
				BaseTx: transaction.BaseTx{
					Account: issuer.GetAddress(),
				},
				LimitAmount: types.IssuedCurrencyAmount{
					Currency: "USD",
					Issuer:   receiver.GetAddress(),
					Value:    "100000000000000",
				},
				QualityOut: 1010000000,
				QualityIn:  1010000000,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			flatSetupTx := tc.TrustSet.Flatten()
			_, err := runner.TestTransaction(&flatSetupTx, issuer, "tesSUCCESS", nil)
			require.NoError(t, err)
		})
	}
}

func fronzenTrustlineTrustSetTest(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 2,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	receiver := runner.GetWallet(1)

	tt := []TrustSet{
		{
			Name: "frozen trustline test",
			TrustSet: &transaction.TrustSet{
				BaseTx: transaction.BaseTx{
					Account: issuer.GetAddress(),
				},

				LimitAmount: types.IssuedCurrencyAmount{
					Currency: "USD",
					Issuer:   receiver.GetAddress(),
					Value:    "100000000000000",
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			tc.TrustSet.SetSetDeepFreezeFlag()
			tc.TrustSet.SetSetFreezeFlag()
			flatSetupTx := tc.TrustSet.Flatten()
			_, err := runner.TestTransaction(&flatSetupTx, issuer, "tesSUCCESS", nil)
			require.NoError(t, err)

			accountLines, err := client.GetAccountLines(&account.LinesRequest{
				Account: issuer.GetAddress(),
			})
			require.NoError(t, err)
			require.Len(t, accountLines.Lines, 1)

			rippleState := accountLines.Lines[0]

			require.True(t, rippleState.Freeze)
		})
	}
}

func TestIntegrationTrustSet_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	TrustSetTest(t, client)
	fronzenTrustlineTrustSetTest(t, client)
}

func TestIntegrationTrustSet_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	TrustSetTest(t, client)
	fronzenTrustlineTrustSetTest(t, client)
}
