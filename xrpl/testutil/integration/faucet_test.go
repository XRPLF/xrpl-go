package integration

import (
	"testing"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/stretchr/testify/require"
)

type localFundingClient struct {
	Client

	autofillCalled            bool
	submitTxBlobCalled        bool
	submitTxBlobAndWaitCalled bool
	validatedResult           string
}

func (*localFundingClient) FaucetProvider() common.FaucetProvider {
	return nil
}

func (c *localFundingClient) Autofill(tx *transaction.FlatTransaction) error {
	c.autofillCalled = true
	(*tx)["Sequence"] = uint32(1)
	(*tx)["Fee"] = "10"
	(*tx)["LastLedgerSequence"] = uint32(20)
	return nil
}

func (c *localFundingClient) SubmitTxBlob(string, bool) (*transactions.SubmitResponse, error) {
	c.submitTxBlobCalled = true
	return nil, ErrFailedToFundWallet
}

func (c *localFundingClient) SubmitTxBlobAndWait(blob string, _ bool) (*transactions.TxResponse, error) {
	c.submitTxBlobAndWaitCalled = true
	if blob == "" {
		return nil, ErrFailedToFundWallet
	}
	return &transactions.TxResponse{
		Meta: transaction.TxMetadataBuilder{
			TransactionResult: c.validatedResult,
		},
		Validated: true,
	}, nil
}

func TestFundWalletWithGenesisWaitsForValidatedSuccess(t *testing.T) {
	client := &localFundingClient{validatedResult: transaction.TesSUCCESS.String()}
	runner := NewRunner(t, client, nil)
	fundedWallet, err := wallet.New(crypto.ED25519())
	require.NoError(t, err)

	err = runner.FundWallet(&fundedWallet)

	require.NoError(t, err)
	require.True(t, client.autofillCalled)
	require.True(t, client.submitTxBlobAndWaitCalled)
	require.False(t, client.submitTxBlobCalled)
}

func TestFundWalletWithGenesisRejectsValidatedFailure(t *testing.T) {
	client := &localFundingClient{validatedResult: "tecUNFUNDED_PAYMENT"}
	runner := NewRunner(t, client, nil)
	fundedWallet, err := wallet.New(crypto.ED25519())
	require.NoError(t, err)

	err = runner.FundWallet(&fundedWallet)

	require.ErrorIs(t, err, ErrFailedToFundWallet)
	require.ErrorContains(t, err, "tecUNFUNDED_PAYMENT")
	require.True(t, client.submitTxBlobAndWaitCalled)
	require.False(t, client.submitTxBlobCalled)
}
