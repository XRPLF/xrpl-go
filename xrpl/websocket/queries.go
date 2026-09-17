package websocket

import (
	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/amm"
	"github.com/Peersyst/xrpl-go/xrpl/queries/channel"
	"github.com/Peersyst/xrpl-go/xrpl/queries/clio"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/queries/nft"
	"github.com/Peersyst/xrpl-go/xrpl/queries/oracle"
	"github.com/Peersyst/xrpl-go/xrpl/queries/path"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
	"github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/queries/utility"
	"github.com/Peersyst/xrpl-go/xrpl/queries/vault"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// Account queries

// GetAccountInfo retrieves information about an account on the XRP Ledger.
// It takes an AccountInfoRequest as input and returns an AccountInfoResponse,
// along with the raw XRPL response and any error encountered.
func (c *Client) GetAccountInfo(req *account.InfoRequest) (*account.InfoResponse, error) {
	return clientinternal.DecodeResult[account.InfoResponse](c.Request(req))
}

// GetAccountChannels retrieves a list of payment channels associated with an account.
// It takes an AccountChannelsRequest as input and returns an AccountChannelsResponse,
// along with any error encountered.
func (c *Client) GetAccountChannels(req *account.ChannelsRequest) (*account.ChannelsResponse, error) {
	return clientinternal.DecodeResult[account.ChannelsResponse](c.Request(req))
}

// GetAccountObjects retrieves a list of objects owned by an account on the XRP Ledger.
// It takes an AccountObjectsRequest as input and returns an AccountObjectsResponse,
// along with any error encountered.
func (c *Client) GetAccountObjects(req *account.ObjectsRequest) (*account.ObjectsResponse, error) {
	return clientinternal.DecodeResult[account.ObjectsResponse](c.Request(req))
}

// GetXrpBalance retrieves the XRP balance of a given account address.
// It returns the balance as a string in XRP (not drops) and any error encountered.
func (c *Client) GetXrpBalance(address types.Address) (string, error) {
	return clientinternal.GetXRPBalance(c.GetAccountInfo, address, nil)
}

// GetXrpBalanceValidated retrieves the XRP balance of a given account address
// from the most recently validated ledger. It returns the balance as a string
// in XRP (not drops) and any error encountered.
func (c *Client) GetXrpBalanceValidated(address types.Address) (string, error) {
	return clientinternal.GetXRPBalance(c.GetAccountInfo, address, common.Validated)
}

// GetXrpDropsBalanceValidated retrieves the XRP balance of a given account
// address from the most recently validated ledger in drops. Prefer this over
// GetXrpBalanceValidated when callers need integer drops (avoids a round-trip
// through a decimal XRP string).
func (c *Client) GetXrpDropsBalanceValidated(address types.Address) (types.XRPCurrencyAmount, error) {
	return clientinternal.GetXRPDropsBalance(c.GetAccountInfo, address, common.Validated)
}

// GetAccountSponsoring retrieves raw sponsored objects. The server must support
// account_sponsoring, identified as Clio-only by XRPL.js.
func (c *Client) GetAccountSponsoring(req *clio.AccountSponsoringRequest) (*clio.AccountSponsoringResponse, error) {
	return clientinternal.DecodeResult[clio.AccountSponsoringResponse](c.Request(req))
}

// GetAccountLines retrieves the lines associated with an account on the XRP Ledger.
// It takes an AccountLinesRequest as input and returns an AccountLinesResponse,
// along with any error encountered.
func (c *Client) GetAccountLines(req *account.LinesRequest) (*account.LinesResponse, error) {
	return clientinternal.DecodeResult[account.LinesResponse](c.Request(req))
}

// GetAccountNFTs retrieves a list of NFTs owned by an account on the XRP Ledger.
// It takes an AccountNFTsRequest as input and returns an AccountNFTsResponse,
// along with any error encountered.
func (c *Client) GetAccountNFTs(req *account.NFTsRequest) (*account.NFTsResponse, error) {
	return clientinternal.DecodeResult[account.NFTsResponse](c.Request(req))
}

// GetAccountCurrencies retrieves a list of currencies that an account can send or receive.
// It takes an AccountCurrenciesRequest as input and returns an AccountCurrenciesResponse,
// along with any error encountered.
func (c *Client) GetAccountCurrencies(req *account.CurrenciesRequest) (*account.CurrenciesResponse, error) {
	return clientinternal.DecodeResult[account.CurrenciesResponse](c.Request(req))
}

// GetAccountOffers retrieves a list of offers made by an account that are currently active
// in the XRP Ledger's decentralized exchange.
// It takes an AccountOffersRequest as input and returns an AccountOffersResponse,
// along with any error encountered.
func (c *Client) GetAccountOffers(req *account.OffersRequest) (*account.OffersResponse, error) {
	return clientinternal.DecodeResult[account.OffersResponse](c.Request(req))
}

// GetAccountTransactions retrieves a list of transactions that involved a specific account.
// It takes an AccountTransactionsRequest as input and returns an AccountTransactionsResponse,
// along with any error encountered.
func (c *Client) GetAccountTransactions(req *account.TransactionsRequest) (*account.TransactionsResponse, error) {
	return clientinternal.DecodeResult[account.TransactionsResponse](c.Request(req))
}

// GetGatewayBalances retrieves the gateway balances for an account.
// It takes a GatewayBalancesRequest as input and returns a GatewayBalancesResponse,
// along with any error encountered.
func (c *Client) GetGatewayBalances(req *account.GatewayBalancesRequest) (*account.GatewayBalancesResponse, error) {
	return clientinternal.DecodeResult[account.GatewayBalancesResponse](c.Request(req))
}

// Channel queries

// GetChannelVerify verifies the signature of a payment channel claim.
// It takes a ChannelVerifyRequest as input and returns a ChannelVerifyResponse,
// along with any error encountered.
func (c *Client) GetChannelVerify(req *channel.VerifyRequest) (*channel.VerifyResponse, error) {
	return clientinternal.DecodeResult[channel.VerifyResponse](c.Request(req))
}

// Transaction queries

// Simulate executes an unsigned transaction as a dry run without submitting it
// to the network. Results reflect current ledger state and do not guarantee the
// outcome of a later submission.
func (c *Client) Simulate(req *transactions.SimulateRequest) (*transactions.SimulateResponse, error) {
	networkID, _ := c.NetworkIdentity()
	if err := req.ValidateNetworkID(networkID); err != nil {
		return nil, err
	}
	response, err := c.Request(req)
	return clientinternal.DecodeSimulate(req, response, err)
}

// Ledger queries

// GetLedgerIndex returns the index of the most recently validated ledger.
// It returns the ledger index as a LedgerIndex type and any error encountered.
func (c *Client) GetLedgerIndex() (common.LedgerIndex, error) {
	return clientinternal.GetLedgerIndex(c.GetLedger)
}

// GetClosedLedger retrieves information about the last closed ledger.
// It returns a ClosedResponse containing the ledger information and any error encountered.
func (c *Client) GetClosedLedger() (*ledger.ClosedResponse, error) {
	return clientinternal.DecodeResult[ledger.ClosedResponse](c.Request(&ledger.ClosedRequest{}))
}

// GetCurrentLedger retrieves information about the current working ledger.
// It returns a CurrentResponse containing the ledger information and any error encountered.
func (c *Client) GetCurrentLedger() (*ledger.CurrentResponse, error) {
	return clientinternal.DecodeResult[ledger.CurrentResponse](c.Request(&ledger.CurrentRequest{}))
}

// GetLedgerData retrieves contents of a ledger.
// It takes a DataRequest as input and returns a DataResponse containing the ledger data,
// along with any error encountered.
func (c *Client) GetLedgerData(req *ledger.DataRequest) (*ledger.DataResponse, error) {
	return clientinternal.DecodeResult[ledger.DataResponse](c.Request(req))
}

// GetLedger retrieves information about a specific ledger version.
// It takes a Request as input and returns a Response containing the ledger information,
// along with any error encountered.
func (c *Client) GetLedger(req *ledger.Request) (*ledger.Response, error) {
	return clientinternal.DecodeResult[ledger.Response](c.Request(req))
}

// GetLedgerEntry retrieves a specific ledger entry by its index.
// It takes an EntryRequest as input and returns an EntryResponse containing the ledger entry,
// along with any error encountered.
func (c *Client) GetLedgerEntry(req *ledger.EntryRequest) (*ledger.EntryResponse, error) {
	return clientinternal.DecodeResult[ledger.EntryResponse](c.Request(req))
}

// NFT queries

// GetNFTBuyOffers retrieves all buy offers for a specific NFT.
// It takes an NFTokenBuyOffersRequest as input and returns an NFTokenBuyOffersResponse,
// along with any error encountered.
func (c *Client) GetNFTBuyOffers(req *nft.NFTokenBuyOffersRequest) (*nft.NFTokenBuyOffersResponse, error) {
	return clientinternal.DecodeResult[nft.NFTokenBuyOffersResponse](c.Request(req))
}

// GetNFTSellOffers retrieves all sell offers for a specific NFT.
// It takes an NFTokenSellOffersRequest as input and returns an NFTokenSellOffersResponse,
// along with any error encountered.
func (c *Client) GetNFTSellOffers(req *nft.NFTokenSellOffersRequest) (*nft.NFTokenSellOffersResponse, error) {
	return clientinternal.DecodeResult[nft.NFTokenSellOffersResponse](c.Request(req))
}

// Path queries

// GetBookOffers retrieves a list of offers between two currencies.
// It takes a BookOffersRequest as input and returns a BookOffersResponse,
// along with any error encountered.
func (c *Client) GetBookOffers(req *path.BookOffersRequest) (*path.BookOffersResponse, error) {
	return clientinternal.DecodeResult[path.BookOffersResponse](c.Request(req))
}

// GetDepositAuthorized checks whether one account is authorized to send payments directly to another.
// It takes a DepositAuthorizedRequest as input and returns a DepositAuthorizedResponse,
// along with any error encountered.
func (c *Client) GetDepositAuthorized(req *path.DepositAuthorizedRequest) (*path.DepositAuthorizedResponse, error) {
	return clientinternal.DecodeResult[path.DepositAuthorizedResponse](c.Request(req))
}

// FindPathCreate creates a path finding request that will be monitored until it expires or is closed.
// It takes a FindCreateRequest as input and returns a FindResponse,
// along with any error encountered.
func (c *Client) FindPathCreate(req *path.FindCreateRequest) (*path.FindResponse, error) {
	return clientinternal.DecodeResult[path.FindResponse](c.Request(req))
}

// FindPathClose closes an existing path finding request.
// It takes a FindCloseRequest as input and returns a FindResponse,
// along with any error encountered.
func (c *Client) FindPathClose(req *path.FindCloseRequest) (*path.FindResponse, error) {
	return clientinternal.DecodeResult[path.FindResponse](c.Request(req))
}

// FindPathStatus checks the status of an existing path finding request.
// It takes a FindStatusRequest as input and returns a FindResponse,
// along with any error encountered.
func (c *Client) FindPathStatus(req *path.FindStatusRequest) (*path.FindResponse, error) {
	return clientinternal.DecodeResult[path.FindResponse](c.Request(req))
}

// GetRipplePathFind finds paths for a payment between two accounts.
// It takes a RipplePathFindRequest as input and returns a RipplePathFindResponse,
// along with any error encountered.
func (c *Client) GetRipplePathFind(req *path.RipplePathFindRequest) (*path.RipplePathFindResponse, error) {
	return clientinternal.DecodeResult[path.RipplePathFindResponse](c.Request(req))
}

// Server queries

// GetServerInfo retrieves information about the server.
// It takes a ServerInfoRequest as input and returns a ServerInfoResponse,
// along with any error encountered.
func (c *Client) GetServerInfo(req *server.InfoRequest) (*server.InfoResponse, error) {
	return clientinternal.DecodeResult[server.InfoResponse](c.Request(req))
}

// GetServerDefinitions retrieves the serialization definitions supported by the server.
// A request hash that matches the server's definitions produces a hash-only response.
func (c *Client) GetServerDefinitions(req *server.DefinitionsRequest) (*server.DefinitionsResponse, error) {
	response, err := c.Request(req)
	return clientinternal.DecodeServerDefinitions(req, response, err)
}

// GetAllFeatures retrieves information about all features supported by the server.
// It takes a FeatureAllRequest as input and returns a FeatureAllResponse,
// along with any error encountered.
func (c *Client) GetAllFeatures(req *server.FeatureAllRequest) (*server.FeatureAllResponse, error) {
	return clientinternal.DecodeResult[server.FeatureAllResponse](c.Request(req))
}

// GetFeature retrieves information about a specific feature supported by the server.
// It takes a FeatureOneRequest as input and returns a FeatureResponse,
// along with any error encountered.
func (c *Client) GetFeature(req *server.FeatureOneRequest) (*server.FeatureResponse, error) {
	return clientinternal.DecodeResult[server.FeatureResponse](c.Request(req))
}

// GetFee retrieves the current transaction fee settings from the server.
// It takes a FeeRequest as input and returns a FeeResponse,
// along with any error encountered.
func (c *Client) GetFee(req *server.FeeRequest) (*server.FeeResponse, error) {
	return clientinternal.DecodeResult[server.FeeResponse](c.Request(req))
}

// GetManifest retrieves public information about a known validator.
// It takes a ManifestRequest as input and returns a ManifestResponse,
// along with any error encountered.
func (c *Client) GetManifest(req *server.ManifestRequest) (*server.ManifestResponse, error) {
	return clientinternal.DecodeResult[server.ManifestResponse](c.Request(req))
}

// GetServerState retrieves information about the current state of the server.
// It takes a StateRequest as input and returns a StateResponse,
// along with any error encountered.
func (c *Client) GetServerState(req *server.StateRequest) (*server.StateResponse, error) {
	return clientinternal.DecodeResult[server.StateResponse](c.Request(req))
}

// Oracle queries

// GetAggregatePrice retrieves the aggregate price of an asset.
// It takes a GetAggregatePriceRequest as input and returns a GetAggregatePriceResponse,
// along with any error encountered.
func (c *Client) GetAggregatePrice(req *oracle.GetAggregatePriceRequest) (*oracle.GetAggregatePriceResponse, error) {
	return clientinternal.DecodeResult[oracle.GetAggregatePriceResponse](c.Request(req))
}

// AMM queries

// GetAMMInfo retrieves information about an AMM instance.
// It takes an InfoRequest as input and returns an InfoResponse,
// along with any error encountered.
func (c *Client) GetAMMInfo(req *amm.InfoRequest) (*amm.InfoResponse, error) {
	return clientinternal.DecodeResult[amm.InfoResponse](c.Request(req))
}

// Vault queries

// GetVaultInfo retrieves information about a Vault instance.
// It takes a InfoRequest as input and returns a Response,
// along with any error encountered.
func (c *Client) GetVaultInfo(req *vault.InfoRequest) (*vault.Response, error) {
	return clientinternal.DecodeResult[vault.Response](c.Request(req))
}

// Utility queries

// Ping tests the connection to the server.
// It takes a PingRequest as input and returns a PingResponse,
// along with any error encountered.
func (c *Client) Ping(req *utility.PingRequest) (*utility.PingResponse, error) {
	return clientinternal.DecodeResult[utility.PingResponse](c.Request(req))
}

// GetRandom provides a random number from the server.
// It takes a RandomRequest as input and returns a RandomResponse,
// along with any error encountered.
func (c *Client) GetRandom(req *utility.RandomRequest) (*utility.RandomResponse, error) {
	return clientinternal.DecodeResult[utility.RandomResponse](c.Request(req))
}
