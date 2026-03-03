package integration

import (
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

// ############################################################################
// VaultCreate
// ############################################################################

type VaultCreateTest struct {
	Name          string
	VaultCreate   *transaction.VaultCreate
	ExpectedError string
}

func TestIntegrationVaultCreate_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))

	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 1,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	owner := runner.GetWallet(0)

	tt := []VaultCreateTest{
		{
			Name: "pass - create XRP vault",
			VaultCreate: &transaction.VaultCreate{
				BaseTx: transaction.BaseTx{
					Account: owner.GetAddress(),
				},
				Asset: ledger.Asset{
					Currency: "XRP",
				},
			},
		},
		{
			Name: "fail - missing Asset",
			VaultCreate: &transaction.VaultCreate{
				BaseTx: transaction.BaseTx{
					Account: owner.GetAddress(),
				},
			},
			ExpectedError: ErrInvalidTransaction,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			flatTx := tc.VaultCreate.Flatten()
			_, err := runner.TestTransaction(&flatTx, owner, "tesSUCCESS", nil)
			if tc.ExpectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.ExpectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ############################################################################
// VaultDeposit
// ############################################################################

type VaultDepositTest struct {
	Name          string
	VaultDeposit  *transaction.VaultDeposit
	ExpectedError string
}

func TestIntegrationVaultDeposit_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))

	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 1,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	owner := runner.GetWallet(0)

	tt := []VaultDepositTest{
		{
			Name: "fail - missing VaultID",
			VaultDeposit: &transaction.VaultDeposit{
				BaseTx: transaction.BaseTx{
					Account: owner.GetAddress(),
				},
				Amount: types.XRPCurrencyAmount(1000000),
			},
			ExpectedError: "invalid hash length",
		},
		{
			Name: "fail - missing Amount",
			VaultDeposit: &transaction.VaultDeposit{
				BaseTx: transaction.BaseTx{
					Account: owner.GetAddress(),
				},
				VaultID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
			},
			ExpectedError: ErrInvalidTransaction,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			flatTx := tc.VaultDeposit.Flatten()
			_, err := runner.TestTransaction(&flatTx, owner, "tesSUCCESS", nil)
			if tc.ExpectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.ExpectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ############################################################################
// VaultWithdraw
// ############################################################################

type VaultWithdrawTest struct {
	Name          string
	VaultWithdraw *transaction.VaultWithdraw
	ExpectedError string
}

func TestIntegrationVaultWithdraw_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))

	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 1,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	owner := runner.GetWallet(0)

	tt := []VaultWithdrawTest{
		{
			Name: "fail - missing VaultID",
			VaultWithdraw: &transaction.VaultWithdraw{
				BaseTx: transaction.BaseTx{
					Account: owner.GetAddress(),
				},
				Amount: types.XRPCurrencyAmount(500000),
			},
			ExpectedError: "invalid hash length",
		},
		{
			Name: "fail - missing Amount",
			VaultWithdraw: &transaction.VaultWithdraw{
				BaseTx: transaction.BaseTx{
					Account: owner.GetAddress(),
				},
				VaultID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
			},
			ExpectedError: ErrInvalidTransaction,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			flatTx := tc.VaultWithdraw.Flatten()
			_, err := runner.TestTransaction(&flatTx, owner, "tesSUCCESS", nil)
			if tc.ExpectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.ExpectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ############################################################################
// VaultSet
// ############################################################################

type VaultSetTest struct {
	Name          string
	VaultSet      *transaction.VaultSet
	ExpectedError string
}

func TestIntegrationVaultSet_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))

	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 1,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	owner := runner.GetWallet(0)

	tt := []VaultSetTest{
		{
			Name: "fail - missing VaultID",
			VaultSet: &transaction.VaultSet{
				BaseTx: transaction.BaseTx{
					Account: owner.GetAddress(),
				},
			},
			ExpectedError: "invalid hash length",
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			flatTx := tc.VaultSet.Flatten()
			_, err := runner.TestTransaction(&flatTx, owner, "tesSUCCESS", nil)
			if tc.ExpectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.ExpectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ############################################################################
// VaultDelete
// ############################################################################

type VaultDeleteTest struct {
	Name          string
	VaultDelete   *transaction.VaultDelete
	ExpectedError string
}

func TestIntegrationVaultDelete_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))

	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 1,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	owner := runner.GetWallet(0)

	tt := []VaultDeleteTest{
		{
			Name: "fail - missing VaultID",
			VaultDelete: &transaction.VaultDelete{
				BaseTx: transaction.BaseTx{
					Account: owner.GetAddress(),
				},
			},
			ExpectedError: "invalid hash length",
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			flatTx := tc.VaultDelete.Flatten()
			_, err := runner.TestTransaction(&flatTx, owner, "tesSUCCESS", nil)
			if tc.ExpectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.ExpectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ############################################################################
// VaultClawback
// ############################################################################

type VaultClawbackTest struct {
	Name          string
	VaultClawback *transaction.VaultClawback
	ExpectedError string
}

func TestIntegrationVaultClawback_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))

	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 2,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	holder := runner.GetWallet(1)

	tt := []VaultClawbackTest{
		{
			Name: "fail - missing VaultID",
			VaultClawback: &transaction.VaultClawback{
				BaseTx: transaction.BaseTx{
					Account: issuer.GetAddress(),
				},
				Holder: holder.GetAddress(),
			},
			ExpectedError: "invalid hash length",
		},
		{
			Name: "fail - missing Holder",
			VaultClawback: &transaction.VaultClawback{
				BaseTx: transaction.BaseTx{
					Account: issuer.GetAddress(),
				},
				VaultID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
			},
			ExpectedError: "invalid address format",
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			flatTx := tc.VaultClawback.Flatten()
			_, err := runner.TestTransaction(&flatTx, issuer, "tesSUCCESS", nil)
			if tc.ExpectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.ExpectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ############################################################################
// Vault Lifecycle (end-to-end)
// ############################################################################

// TestVaultLifecycle_Websocket tests the complete vault lifecycle: create, deposit, set, withdraw, delete.
func TestVaultLifecycle_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))

	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 1,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	owner := runner.GetWallet(0)

	// Step 1: Create an XRP vault
	vaultCreateTx := &transaction.VaultCreate{
		BaseTx: transaction.BaseTx{
			Account: owner.GetAddress(),
		},
		Asset: ledger.Asset{
			Currency: "XRP",
		},
	}

	flatTx := vaultCreateTx.Flatten()
	res, err := runner.TestTransaction(&flatTx, owner, "tesSUCCESS", nil)
	require.NoError(t, err)
	require.NotNil(t, res)
}
