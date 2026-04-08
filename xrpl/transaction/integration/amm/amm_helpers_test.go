package amm

import (
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	accounttypes "github.com/Peersyst/xrpl-go/xrpl/queries/account/types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/stretchr/testify/require"
)

// ammPool holds the state for an AMM pool created during integration tests.
type ammPool struct {
	issuerWallet *wallet.Wallet
	lpWallet     *wallet.Wallet
	testWallet   *wallet.Wallet
	asset        ledger.Asset
	asset2       ledger.Asset
}

// createAMMPool creates the base AMM pool (issuer + lp wallets, trust line, payment, AMMCreate).
func createAMMPool(t *testing.T, runner *integration.Runner, client integration.Client, enableClawback bool) *ammPool {
	t.Helper()

	issuerWallet := runner.GetWallet(0)
	lpWallet := runner.GetWallet(1)

	accountSetTx := &transaction.AccountSet{
		BaseTx: transaction.BaseTx{
			Account: issuerWallet.GetAddress(),
		},
		SetFlag: transaction.AsfDefaultRipple,
	}
	flatAccountSetTx := accountSetTx.Flatten()
	_, err := runner.TestTransaction(&flatAccountSetTx, issuerWallet, "tesSUCCESS", nil)
	require.NoError(t, err)

	if enableClawback {
		clawbackTx := &transaction.AccountSet{
			BaseTx: transaction.BaseTx{
				Account: issuerWallet.GetAddress(),
			},
			SetFlag: transaction.AsfAllowTrustLineClawback,
		}
		flatClawbackTx := clawbackTx.Flatten()
		_, err = runner.TestTransaction(&flatClawbackTx, issuerWallet, "tesSUCCESS", nil)
		require.NoError(t, err)
	}

	trustSetTx := &transaction.TrustSet{
		BaseTx: transaction.BaseTx{
			Account: lpWallet.GetAddress(),
		},
		LimitAmount: types.IssuedCurrencyAmount{
			Currency: "USD",
			Issuer:   issuerWallet.GetAddress(),
			Value:    "1000",
		},
	}
	trustSetTx.SetClearNoRippleFlag()
	flatTrustSetTx := trustSetTx.Flatten()
	_, err = runner.TestTransaction(&flatTrustSetTx, lpWallet, "tesSUCCESS", nil)
	require.NoError(t, err)

	paymentTx := &transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account: issuerWallet.GetAddress(),
		},
		Destination: lpWallet.GetAddress(),
		Amount: types.IssuedCurrencyAmount{
			Currency: "USD",
			Issuer:   issuerWallet.GetAddress(),
			Value:    "500",
		},
	}
	flatPaymentTx := paymentTx.Flatten()
	_, err = runner.TestTransaction(&flatPaymentTx, issuerWallet, "tesSUCCESS", nil)
	require.NoError(t, err)

	ammCreateTx := &transaction.AMMCreate{
		BaseTx: transaction.BaseTx{
			Account: lpWallet.GetAddress(),
		},
		Amount:     types.XRPCurrencyAmount(250),
		Amount2:    types.IssuedCurrencyAmount{Currency: "USD", Issuer: issuerWallet.GetAddress(), Value: "250"},
		TradingFee: 12,
	}
	flatAMMCreateTx := ammCreateTx.Flatten()
	_, err = runner.TestTransaction(&flatAMMCreateTx, lpWallet, "tesSUCCESS", nil)
	require.NoError(t, err)

	return &ammPool{
		issuerWallet: issuerWallet,
		lpWallet:     lpWallet,
		asset:        ledger.Asset{Currency: "XRP"},
		asset2:       ledger.Asset{Currency: "USD", Issuer: issuerWallet.GetAddress()},
	}
}

// setupAMMPool creates the AMM pool and deposits from a testWallet so it becomes an LP.
// Equivalent to the JS setupAMMPool utility.
func setupAMMPool(t *testing.T, runner *integration.Runner, client integration.Client) *ammPool {
	t.Helper()

	pool := createAMMPool(t, runner, client, false)

	testWallet := runner.GetWallet(2)
	pool.testWallet = testWallet

	// testWallet deposits 1000 XRP drops (single asset) to become an LP
	depositTx := &transaction.AMMDeposit{
		BaseTx: transaction.BaseTx{
			Account: testWallet.GetAddress(),
		},
		Asset:  pool.asset,
		Asset2: pool.asset2,
		Amount: types.XRPCurrencyAmount(1000),
	}
	depositTx.SetSingleAssetFlag()
	flatDepositTx := depositTx.Flatten()
	_, err := runner.TestTransaction(&flatDepositTx, testWallet, "tesSUCCESS", nil)
	require.NoError(t, err)

	return pool
}

// getLPToken returns the LP token currency code and issuer from the wallet's account lines.
// LP tokens have a 40-character hex currency code.
func getLPToken(t *testing.T, client integration.Client, walletAddr types.Address) accounttypes.TrustLine {
	t.Helper()

	lines, err := client.GetAccountLines(&account.LinesRequest{
		Account: walletAddr,
	})
	require.NoError(t, err)

	for _, line := range lines.Lines {
		if len(line.Currency) == 40 {
			return line
		}
	}
	t.Fatal("LP token trust line not found for wallet " + walletAddr)
	return accounttypes.TrustLine{}
}
