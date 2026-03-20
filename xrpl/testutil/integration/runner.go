package integration

import (
	"testing"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/stretchr/testify/require"
)

// Runner manages the setup, execution, and teardown of integration test transactions against an XRPL client.
type Runner struct {
	t      *testing.T
	config *RunnerConfig

	client  Client
	wallets []*wallet.Wallet
}

// TestTransactionOptions defines options for testing transactions, such as skipping autofill.
type TestTransactionOptions struct {
	SkipAutofill bool
}

// NewRunner creates a new runner. It doesn't connect to the websocket or generate wallets until Setup is called.
// A testing.T is required to use the require package.
func NewRunner(t *testing.T, client Client, config *RunnerConfig) *Runner {
	if config == nil {
		config = NewRunnerConfig()
	}
	return &Runner{
		t:      t,
		config: config,
		client: client,
	}
}

// Setup creates a new websocket client and generates the required number of wallets.
// It also connects to the websocket and starts the client.
// For every wallet, it will create a new account and fund it with the faucet.
func (r *Runner) Setup() error {
	if connectable, ok := r.client.(Connectable); ok {
		err := connectable.Connect()
		if err != nil {
			return err
		}
	}

	for i := 0; i < r.config.WalletCount; i++ {
		w, err := wallet.New(crypto.ED25519())
		if err != nil {
			return err
		}
		err = r.FundWallet(&w)
		if err != nil {
			return err
		}
		r.wallets = append(r.wallets, &w)
	}
	return nil
}

// Teardown closes the websocket client.
func (r *Runner) Teardown() error {
	if connectable, ok := r.client.(Connectable); ok {
		err := connectable.Disconnect()
		if err != nil {
			return err
		}
	}

	return nil
}

// TestTransaction submits a signed transaction and validates the result.
// If validate is nil, the transaction is not validated.
func (r *Runner) TestTransaction(flatTx *transaction.FlatTransaction, signer *wallet.Wallet, expectedEngineResult string, opts *TestTransactionOptions) (*transactions.SubmitResponse, error) {
	tx, hash, err := r.processTransaction(flatTx, signer, opts)
	if err != nil {
		return nil, err
	}

	require.NoError(r.t, err)
	require.Equal(r.t, expectedEngineResult, tx.EngineResult)
	require.Equal(r.t, hash, tx.Tx["hash"].(string))

	return tx, nil
}

// TestMultisigTransaction submits a multisigned transaction and validates the result.
// If validate is nil, the transaction is not validated.
func (r *Runner) TestMultisigTransaction(blob string, expectedEngineResult string) (*transactions.SubmitMultisignedResponse, error) {
	tx, err := r.client.SubmitMultisigned(blob, true)
	if err != nil {
		return nil, err
	}

	require.NoError(r.t, err)
	require.Equal(r.t, expectedEngineResult, tx.EngineResult)

	return tx, nil
}

// GetWallet returns a wallet by index.
func (r *Runner) GetWallet(index int) *wallet.Wallet {
	if index < 0 || index >= len(r.wallets) {
		return nil
	}
	return r.wallets[index]
}

// GetWallets returns all wallets.
func (r *Runner) GetWallets() []*wallet.Wallet {
	return r.wallets
}

// GetClient returns the websocket client.
func (r *Runner) GetClient() Client {
	return r.client
}

func (r *Runner) processTransaction(flatTx *transaction.FlatTransaction, signer *wallet.Wallet, opts *TestTransactionOptions) (*transactions.SubmitResponse, string, error) {
	attempts := 0

	for {
		if opts == nil || !opts.SkipAutofill {
			err := r.client.Autofill(flatTx)
			if err != nil {
				return nil, "", err
			}
		}

		blob, hash, err := signer.Sign(*flatTx)
		if err != nil {
			return nil, hash, err
		}

		tx, err := r.client.SubmitTxBlob(blob, true)
		if err != nil {
			return nil, hash, err
		}

		if tx.EngineResult != transaction.TefPAST_SEQ.String() || attempts >= r.config.MaxRetries {
			return tx, hash, nil
		}

		delete(*flatTx, "Sequence")
		delete(*flatTx, "LastLedgerSequence")
		attempts++
	}
}
