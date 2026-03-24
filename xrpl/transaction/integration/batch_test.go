package integration

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

type BatchTest struct {
	Name          string
	Batch         *transaction.Batch
	ExpectedError string
}

var CreatePaymentTx = func(sender, receiver *wallet.Wallet, amount types.CurrencyAmount) *transaction.Payment {
	return &transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account:         sender.GetAddress(),
			TransactionType: transaction.PaymentTx,
			Flags:           0x40000000,
		},
		Amount:      amount,
		Destination: receiver.GetAddress(),
	}
}

func TestIntegrationBatch_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))

	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 3,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	sender := runner.GetWallet(0)
	receiver := runner.GetWallet(1)

	tt := []BatchTest{
		{
			Name: "pass - valid batch transaction",
			Batch: &transaction.Batch{
				BaseTx: transaction.BaseTx{
					Account:         runner.GetWallet(0).GetAddress(),
					TransactionType: transaction.BatchTx,
				},
				RawTransactions: []types.RawTransaction{
					{
						RawTransaction: CreatePaymentTx(sender, receiver, types.XRPCurrencyAmount(1)).Flatten(),
					},
					{
						RawTransaction: CreatePaymentTx(sender, receiver, types.XRPCurrencyAmount(1)).Flatten(),
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Batch.SetAllOrNothingFlag()
			flatTx := tc.Batch.Flatten()
			err := client.Autofill(&flatTx)

			require.NoError(t, err)

			_, err = runner.TestTransaction(&flatTx, sender, "tesSUCCESS", &integration.TestTransactionOptions{
				SkipAutofill: true,
			})
			require.NoError(t, err)
		})
	}
}

func TestIntegrationBatchMultisign_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))

	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 3,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	sender := runner.GetWallet(0)
	sender2 := runner.GetWallet(1)
	receiver := runner.GetWallet(2)

	tt := []BatchTest{
		{
			Name: "pass - valid batch transaction",
			Batch: &transaction.Batch{
				BaseTx: transaction.BaseTx{
					Account:         sender.GetAddress(),
					TransactionType: transaction.BatchTx,
				},
				RawTransactions: []types.RawTransaction{
					{
						RawTransaction: CreatePaymentTx(sender, receiver, types.XRPCurrencyAmount(1)).Flatten(),
					},
					{
						RawTransaction: CreatePaymentTx(sender2, receiver, types.XRPCurrencyAmount(1)).Flatten(),
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Batch.SetAllOrNothingFlag()
			flatTx := tc.Batch.Flatten()
			err := client.AutofillMultisigned(&flatTx, 1)
			require.NoError(t, err)

			err = wallet.SignMultiBatch(*sender2, &flatTx, nil)

			require.NoError(t, err)

			_, err = runner.TestTransaction(&flatTx, sender, "tesSUCCESS", &integration.TestTransactionOptions{
				SkipAutofill: true,
			})

			require.NoError(t, err)
		})
	}
}

func TestIntegrationBatch_RPC(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)

	runner := integration.NewRunner(t, client, integration.NewRunnerConfig(
		integration.WithWallets(3),
	))

	err = runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	sender := runner.GetWallet(0)
	receiver := runner.GetWallet(1)

	tt := []BatchTest{
		{
			Name: "pass - valid batch transaction",
			Batch: &transaction.Batch{
				BaseTx: transaction.BaseTx{
					Account:         runner.GetWallet(0).GetAddress(),
					TransactionType: transaction.BatchTx,
				},
				RawTransactions: []types.RawTransaction{
					{
						RawTransaction: CreatePaymentTx(sender, receiver, types.XRPCurrencyAmount(1)).Flatten(),
					},
					{
						RawTransaction: CreatePaymentTx(sender, receiver, types.XRPCurrencyAmount(1)).Flatten(),
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Batch.SetAllOrNothingFlag()
			flatTx := tc.Batch.Flatten()
			err := client.Autofill(&flatTx)

			require.NoError(t, err)

			_, err = runner.TestTransaction(&flatTx, sender, "tesSUCCESS", &integration.TestTransactionOptions{
				SkipAutofill: true,
			})
			require.NoError(t, err)
		})
	}
}

func TestIntegrationBatchMultisign_RPC(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)

	runner := integration.NewRunner(t, client, integration.NewRunnerConfig(
		integration.WithWallets(3),
	))

	err = runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	sender := runner.GetWallet(0)
	sender2 := runner.GetWallet(1)
	receiver := runner.GetWallet(2)

	tt := []BatchTest{
		{
			Name: "pass - valid batch transaction",
			Batch: &transaction.Batch{
				BaseTx: transaction.BaseTx{
					Account:         sender.GetAddress(),
					TransactionType: transaction.BatchTx,
				},
				RawTransactions: []types.RawTransaction{
					{
						RawTransaction: CreatePaymentTx(sender, receiver, types.XRPCurrencyAmount(1)).Flatten(),
					},
					{
						RawTransaction: CreatePaymentTx(sender2, receiver, types.XRPCurrencyAmount(1)).Flatten(),
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Batch.SetAllOrNothingFlag()
			flatTx := tc.Batch.Flatten()
			err := client.AutofillMultisigned(&flatTx, 1)
			require.NoError(t, err)

			err = wallet.SignMultiBatch(*sender2, &flatTx, nil)

			require.NoError(t, err)

			_, err = runner.TestTransaction(&flatTx, sender, "tesSUCCESS", &integration.TestTransactionOptions{
				SkipAutofill: true,
			})

			require.NoError(t, err)
		})
	}
}
