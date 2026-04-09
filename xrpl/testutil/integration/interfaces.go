package integration

import (
	"github.com/Peersyst/xrpl-go/xrpl/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/amm"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

// FaucetProvider provides faucet funding for wallets in integration tests.
type FaucetProvider interface {
	common.FaucetProvider
}

// Client defines the interface for submitting transactions and funding wallets in integration tests.
type Client interface {
	FaucetProvider() common.FaucetProvider

	FundWallet(wallet *wallet.Wallet) error
	Autofill(tx *transaction.FlatTransaction) error
	SubmitTxBlob(txBlob string, failHard bool) (*transactions.SubmitResponse, error)
	SubmitTxBlobAndWait(txBlob string, failHard bool) (*transactions.TxResponse, error)
	SubmitMultisigned(blob string, validate bool) (*transactions.SubmitMultisignedResponse, error)
	GetAccountObjects(req *account.ObjectsRequest) (*account.ObjectsResponse, error)
	GetAccountLines(req *account.LinesRequest) (*account.LinesResponse, error)
	GetAccountOffers(req *account.OffersRequest) (*account.OffersResponse, error)
	GetAccountNFTs(req *account.NFTsRequest) (*account.NFTsResponse, error)
	GetAMMInfo(req *amm.InfoRequest) (*amm.InfoResponse, error)
	GetLedger(req *ledger.Request) (*ledger.Response, error)
}

// Connectable defines methods to connect and disconnect the integration client.
type Connectable interface {
	Connect() error
	Disconnect() error
}
