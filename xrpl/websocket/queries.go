package websocket

import (
	"github.com/Peersyst/xrpl-go/xrpl/currency"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/channel"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/queries/nft"
	"github.com/Peersyst/xrpl-go/xrpl/queries/oracle"
	"github.com/Peersyst/xrpl-go/xrpl/queries/path"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
	"github.com/Peersyst/xrpl-go/xrpl/queries/utility"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// Account queries

// GetAccountInfo retrieves information about an account on the XRP Ledger.
// It takes an AccountInfoRequest as input and returns an AccountInfoResponse,
// along with the raw XRPL response and any error encountered.
func (c *Client) GetAccountInfo(req *account.InfoRequest) (*account.InfoResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var air account.InfoResponse
	err = res.GetResult(&air)
	if err != nil {
		return nil, err
	}
	return &air, nil
}

// GetAccountChannels retrieves a list of payment channels associated with an account.
// It takes an AccountChannelsRequest as input and returns an AccountChannelsResponse,
// along with any error encountered.
func (c *Client) GetAccountChannels(req *account.ChannelsRequest) (*account.ChannelsResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var acr account.ChannelsResponse
	err = res.GetResult(&acr)
	if err != nil {
		return nil, err
	}
	return &acr, nil
}

// GetAccountObjects retrieves a list of objects owned by an account on the XRP Ledger.
// It takes an AccountObjectsRequest as input and returns an AccountObjectsResponse,
// along with any error encountered.
func (c *Client) GetAccountObjects(req *account.ObjectsRequest) (*account.ObjectsResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var acr account.ObjectsResponse
	err = res.GetResult(&acr)
	if err != nil {
		return nil, err
	}
	return &acr, nil
}

// GetXrpBalance retrieves the XRP balance of a given account address.
// It returns the balance as a string in XRP (not drops) and any error encountered.
func (c *Client) GetXrpBalance(address types.Address) (string, error) {
	res, err := c.GetAccountInfo(&account.InfoRequest{
		Account: address,
	})
	if err != nil {
		return "", err
	}

	xrpBalance, err := currency.DropsToXrp(res.AccountData.Balance.String())
	if err != nil {
		return "", err
	}

	return xrpBalance, nil
}

// GetAccountLines retrieves the lines associated with an account on the XRP Ledger.
// It takes an AccountLinesRequest as input and returns an AccountLinesResponse,
// along with any error encountered.
func (c *Client) GetAccountLines(req *account.LinesRequest) (*account.LinesResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var acr account.LinesResponse
	err = res.GetResult(&acr)
	if err != nil {
		return nil, err
	}
	return &acr, nil
}

// GetAccountNFTs retrieves a list of NFTs owned by an account on the XRP Ledger.
// It takes an AccountNFTsRequest as input and returns an AccountNFTsResponse,
// along with any error encountered.
func (c *Client) GetAccountNFTs(req *account.NFTsRequest) (*account.NFTsResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var acr account.NFTsResponse
	err = res.GetResult(&acr)
	if err != nil {
		return nil, err
	}
	return &acr, nil
}

// GetAccountCurrencies retrieves a list of currencies that an account can send or receive.
// It takes an AccountCurrenciesRequest as input and returns an AccountCurrenciesResponse,
// along with any error encountered.
func (c *Client) GetAccountCurrencies(req *account.CurrenciesRequest) (*account.CurrenciesResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var acr account.CurrenciesResponse
	err = res.GetResult(&acr)
	if err != nil {
		return nil, err
	}
	return &acr, nil
}

// GetAccountOffers retrieves a list of offers made by an account that are currently active
// in the XRP Ledger's decentralized exchange.
// It takes an AccountOffersRequest as input and returns an AccountOffersResponse,
// along with any error encountered.
func (c *Client) GetAccountOffers(req *account.OffersRequest) (*account.OffersResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var acr account.OffersResponse
	err = res.GetResult(&acr)
	if err != nil {
		return nil, err
	}
	return &acr, nil
}

// GetAccountTransactions retrieves a list of transactions that involved a specific account.
// It takes an AccountTransactionsRequest as input and returns an AccountTransactionsResponse,
// along with any error encountered.
func (c *Client) GetAccountTransactions(req *account.TransactionsRequest) (*account.TransactionsResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var acr account.TransactionsResponse
	err = res.GetResult(&acr)
	if err != nil {
		return nil, err
	}
	return &acr, nil
}

// GetGatewayBalances retrieves the gateway balances for an account.
// It takes a GatewayBalancesRequest as input and returns a GatewayBalancesResponse,
// along with any error encountered.
func (c *Client) GetGatewayBalances(req *account.GatewayBalancesRequest) (*account.GatewayBalancesResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var acr account.GatewayBalancesResponse
	err = res.GetResult(&acr)
	if err != nil {
		return nil, err
	}
	return &acr, nil
}

// Channel queries

// GetChannelVerify verifies the signature of a payment channel claim.
// It takes a ChannelVerifyRequest as input and returns a ChannelVerifyResponse,
// along with any error encountered.
func (c *Client) GetChannelVerify(req *channel.VerifyRequest) (*channel.VerifyResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var acr channel.VerifyResponse
	err = res.GetResult(&acr)
	if err != nil {
		return nil, err
	}
	return &acr, nil
}

// Ledger queries

// GetLedgerIndex returns the index of the most recently validated ledger.
// It returns the ledger index as a LedgerIndex type and any error encountered.
func (c *Client) GetLedgerIndex() (common.LedgerIndex, error) {
	res, err := c.Request(&ledger.Request{
		LedgerIndex: common.LedgerTitle("validated"),
	})
	if err != nil {
		return 0, err
	}

	var lr ledger.Response
	err = res.GetResult(&lr)
	if err != nil {
		return 0, err
	}
	return lr.LedgerIndex, err
}

// GetLedgerEntry retrieves information about a specific ledger entry.
// It takes a LedgerEntryRequest as input and returns a LedgerEntryResponse,
// along with any error encountered.
func (c *Client) GetLedgerEntry(req *ledger.EntryRequest) (*ledger.EntryResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr ledger.EntryResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// GetClosedLedger retrieves information about the last closed ledger.
// It returns a ClosedResponse containing the ledger information and any error encountered.
func (c *Client) GetClosedLedger() (*ledger.ClosedResponse, error) {
	res, err := c.Request(&ledger.ClosedRequest{})
	if err != nil {
		return nil, err
	}
	var lr ledger.ClosedResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// GetCurrentLedger retrieves information about the current working ledger.
// It returns a CurrentResponse containing the ledger information and any error encountered.
func (c *Client) GetCurrentLedger() (*ledger.CurrentResponse, error) {
	res, err := c.Request(&ledger.CurrentRequest{})
	if err != nil {
		return nil, err
	}
	var lr ledger.CurrentResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// GetLedgerData retrieves contents of a ledger.
// It takes a DataRequest as input and returns a DataResponse containing the ledger data,
// along with any error encountered.
func (c *Client) GetLedgerData(req *ledger.DataRequest) (*ledger.DataResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr ledger.DataResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// GetLedger retrieves information about a specific ledger version.
// It takes a Request as input and returns a Response containing the ledger information,
// along with any error encountered.
func (c *Client) GetLedger(req *ledger.Request) (*ledger.Response, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr ledger.Response
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// NFT queries

// GetNFTBuyOffers retrieves all buy offers for a specific NFT.
// It takes an NFTokenBuyOffersRequest as input and returns an NFTokenBuyOffersResponse,
// along with any error encountered.
func (c *Client) GetNFTBuyOffers(req *nft.NFTokenBuyOffersRequest) (*nft.NFTokenBuyOffersResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr nft.NFTokenBuyOffersResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// GetNFTSellOffers retrieves all sell offers for a specific NFT.
// It takes an NFTokenSellOffersRequest as input and returns an NFTokenSellOffersResponse,
// along with any error encountered.
func (c *Client) GetNFTSellOffers(req *nft.NFTokenSellOffersRequest) (*nft.NFTokenSellOffersResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr nft.NFTokenSellOffersResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// Path queries

// GetBookOffers retrieves a list of offers between two currencies.
// It takes a BookOffersRequest as input and returns a BookOffersResponse,
// along with any error encountered.
func (c *Client) GetBookOffers(req *path.BookOffersRequest) (*path.BookOffersResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr path.BookOffersResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// GetDepositAuthorized checks whether one account is authorized to send payments directly to another.
// It takes a DepositAuthorizedRequest as input and returns a DepositAuthorizedResponse,
// along with any error encountered.
func (c *Client) GetDepositAuthorized(req *path.DepositAuthorizedRequest) (*path.DepositAuthorizedResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr path.DepositAuthorizedResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// FindPathCreate creates a path finding request that will be monitored until it expires or is closed.
// It takes a FindCreateRequest as input and returns a FindResponse,
// along with any error encountered.
func (c *Client) FindPathCreate(req *path.FindCreateRequest) (*path.FindResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr path.FindResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// FindPathClose closes an existing path finding request.
// It takes a FindCloseRequest as input and returns a FindResponse,
// along with any error encountered.
func (c *Client) FindPathClose(req *path.FindCloseRequest) (*path.FindResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr path.FindResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// FindPathStatus checks the status of an existing path finding request.
// It takes a FindStatusRequest as input and returns a FindResponse,
// along with any error encountered.
func (c *Client) FindPathStatus(req *path.FindStatusRequest) (*path.FindResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr path.FindResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// GetRipplePathFind finds paths for a payment between two accounts.
// It takes a RipplePathFindRequest as input and returns a RipplePathFindResponse,
// along with any error encountered.
func (c *Client) GetRipplePathFind(req *path.RipplePathFindRequest) (*path.RipplePathFindResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr path.RipplePathFindResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// Server queries

// GetServerInfo retrieves information about the server.
// It takes a ServerInfoRequest as input and returns a ServerInfoResponse,
// along with any error encountered.
func (c *Client) GetServerInfo(req *server.InfoRequest) (*server.InfoResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var sir server.InfoResponse
	err = res.GetResult(&sir)
	if err != nil {
		return nil, err
	}
	return &sir, err
}

// GetAllFeatures retrieves information about all features supported by the server.
// It takes a FeatureAllRequest as input and returns a FeatureAllResponse,
// along with any error encountered.
func (c *Client) GetAllFeatures(req *server.FeatureAllRequest) (*server.FeatureAllResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr server.FeatureAllResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// GetFeature retrieves information about a specific feature supported by the server.
// It takes a FeatureOneRequest as input and returns a FeatureResponse,
// along with any error encountered.
func (c *Client) GetFeature(req *server.FeatureOneRequest) (*server.FeatureResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr server.FeatureResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// GetFee retrieves the current transaction fee settings from the server.
// It takes a FeeRequest as input and returns a FeeResponse,
// along with any error encountered.
func (c *Client) GetFee(req *server.FeeRequest) (*server.FeeResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr server.FeeResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// GetManifest retrieves public information about a known validator.
// It takes a ManifestRequest as input and returns a ManifestResponse,
// along with any error encountered.
func (c *Client) GetManifest(req *server.ManifestRequest) (*server.ManifestResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr server.ManifestResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// GetServerState retrieves information about the current state of the server.
// It takes a StateRequest as input and returns a StateResponse,
// along with any error encountered.
func (c *Client) GetServerState(req *server.StateRequest) (*server.StateResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr server.StateResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// Oracle queries

// GetAggregatePrice retrieves the aggregate price of an asset.
// It takes a GetAggregatePriceRequest as input and returns a GetAggregatePriceResponse,
// along with any error encountered.
func (c *Client) GetAggregatePrice(req *oracle.GetAggregatePriceRequest) (*oracle.GetAggregatePriceResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr oracle.GetAggregatePriceResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// Utility queries

// Ping tests the connection to the server.
// It takes a PingRequest as input and returns a PingResponse,
// along with any error encountered.
func (c *Client) Ping(req *utility.PingRequest) (*utility.PingResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr utility.PingResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// GetRandom provides a random number from the server.
// It takes a RandomRequest as input and returns a RandomResponse,
// along with any error encountered.
func (c *Client) GetRandom(req *utility.RandomRequest) (*utility.RandomResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr utility.RandomResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}
