package integration

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

const (
	// LocalGenesisAddress is the default XRPL address used by the local genesis faucet.
	LocalGenesisAddress types.Address = "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	// LocalGenesisSeed is the seed corresponding to the LocalGenesisAddress genesis account.
	LocalGenesisSeed string = "snoPBrXtMeMyMHUVTgbuqAfg1SUTb"
)

// FundWallet funds a wallet with the client's faucet provider.
// If the faucet provider is nil, it will fund the wallet with the local genesis wallet.
func (f *Runner) FundWallet(wallet *wallet.Wallet) error {
	attempts := 0

	if f.client.FaucetProvider() == nil {
		return f.fundWalletWithGenesis(wallet)
	}

	for {
		err := f.client.FundWallet(wallet)
		if err == nil {
			return nil
		}
		if attempts >= f.config.MaxRetries {
			break
		}
		attempts++
	}

	return ErrFailedToFundWallet
}

// fundWalletWithGenesis funds a wallet with the local genesis wallet.
func (f *Runner) fundWalletWithGenesis(w *wallet.Wallet) error {
	genesisWallet, err := wallet.FromSeed(LocalGenesisSeed, "")
	if err != nil {
		return err
	}
	genesisAddress := genesisWallet.GetAddress()

	payment := &transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account: genesisAddress,
		},
		Amount:      types.XRPCurrencyAmount(400000000),
		Destination: w.GetAddress(),
	}

	flatTx := payment.Flatten()
	if err := f.client.Autofill(&flatTx); err != nil {
		return err
	}

	blob, _, err := genesisWallet.Sign(flatTx)
	if err != nil {
		return err
	}

	response, err := f.client.SubmitTxBlobAndWait(blob, true)
	if err != nil {
		return err
	}
	if response.Meta.TransactionResult != transaction.TesSUCCESS.String() {
		return fmt.Errorf(
			"%w: validated transaction result %q",
			ErrFailedToFundWallet,
			response.Meta.TransactionResult,
		)
	}

	return nil
}
