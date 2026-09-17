package rpc

import (
	"encoding/json"
	"net/http"
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/amm"
	"github.com/Peersyst/xrpl-go/xrpl/queries/channel"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	ledgerqueries "github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/queries/nft"
	"github.com/Peersyst/xrpl-go/xrpl/queries/oracle"
	"github.com/Peersyst/xrpl-go/xrpl/queries/path"
	pathtypes "github.com/Peersyst/xrpl-go/xrpl/queries/path/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
	servertypes "github.com/Peersyst/xrpl-go/xrpl/queries/server/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/utility"
	"github.com/Peersyst/xrpl-go/xrpl/queries/vault"
	"github.com/Peersyst/xrpl-go/xrpl/rpc/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func setupQueryTestClient(t *testing.T, result any) (*Client, func() map[string]any) {
	t.Helper()
	body, err := json.Marshal(map[string]any{"result": result})
	require.NoError(t, err)
	mock := &testutil.JSONRPCMockClient{}
	mock.DoFunc = testutil.MockResponse(string(body), http.StatusOK, mock)
	config, err := NewClientConfig("http://testnode/", WithHTTPClient(mock))
	require.NoError(t, err)

	return NewClient(config), func() map[string]any {
		t.Helper()
		require.NotNil(t, mock.Spy, "missing request")
		var envelope struct {
			Method string           `json:"method"`
			Params []map[string]any `json:"params"`
		}
		require.NoError(t, json.NewDecoder(mock.Spy.Body).Decode(&envelope))
		require.Len(t, envelope.Params, 1)
		envelope.Params[0]["command"] = envelope.Method
		return envelope.Params[0]
	}
}

// These tests check wrapper wiring. Model tests own detailed response contracts,
// and transport tests own errors and timeouts.
func TestClient_GetAccountInfo(t *testing.T) {
	expected := &account.InfoResponse{AccountData: ledger.AccountRoot{
		Account: "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
	}}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetAccountInfo(&account.InfoRequest{Account: "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn"})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "account_info", request()["command"])
}

func TestClient_GetAccountChannels(t *testing.T) {
	expected := &account.ChannelsResponse{
		Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetAccountChannels(&account.ChannelsRequest{
		Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "account_channels", request()["command"])
}

func TestClient_GetAccountObjects(t *testing.T) {
	expected := &account.ObjectsResponse{
		Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "account_objects", request()["command"])
}

func TestClient_GetAccountLines(t *testing.T) {
	expected := &account.LinesResponse{
		Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetAccountLines(&account.LinesRequest{
		Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "account_lines", request()["command"])
}

func TestClient_GetAccountNFTs(t *testing.T) {
	expected := &account.NFTsResponse{
		Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetAccountNFTs(&account.NFTsRequest{
		Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "account_nfts", request()["command"])
}

func TestClient_GetAccountCurrencies(t *testing.T) {
	expected := &account.CurrenciesResponse{
		LedgerHash: "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetAccountCurrencies(&account.CurrenciesRequest{
		Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "account_currencies", request()["command"])
}

func TestClient_GetAccountOffers(t *testing.T) {
	expected := &account.OffersResponse{
		Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetAccountOffers(&account.OffersRequest{
		Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "account_offers", request()["command"])
}

func TestClient_GetAccountTransactions(t *testing.T) {
	expected := &account.TransactionsResponse{
		Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetAccountTransactions(&account.TransactionsRequest{
		Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "account_tx", request()["command"])
}

func TestClient_GetGatewayBalances(t *testing.T) {
	expected := &account.GatewayBalancesResponse{
		Account: "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetGatewayBalances(&account.GatewayBalancesRequest{
		Account: "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
	})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "gateway_balances", request()["command"])
}

func TestClient_GetChannelVerify(t *testing.T) {
	expected := &channel.VerifyResponse{
		SignatureVerified: true,
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetChannelVerify(&channel.VerifyRequest{
		Amount:    types.XRPCurrencyAmount(1000000),
		ChannelID: "5DB01B7FFED6B67E6B0414DED11E051D2EE2B7619CE0EAA6286D67A3A4D5BDB3",
		PublicKey: "023693F15967AE357D0327974AD46FE3C127113B1110D6044FD41E723689F81CC6",
		Signature: "304402204EF0AFB78AC23ED1C472E74F4299C0C21F1B21D07EFC0A3838A420F76D783A400220154FB11B6F54320666E4C36CA7F686C16A3A0456800BBC43746F34AF50290064",
	})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "channel_verify", request()["command"])
}

func TestClient_GetClosedLedger(t *testing.T) {
	expected := &ledgerqueries.
		ClosedResponse{LedgerIndex: 123}

	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetClosedLedger()
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "ledger_closed", request()["command"])
}

func TestClient_GetCurrentLedger(t *testing.T) {
	expected := &ledgerqueries.CurrentResponse{
		LedgerCurrentIndex: 71766343,
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetCurrentLedger()
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "ledger_current", request()["command"])
}

func TestClient_GetLedgerData(t *testing.T) {
	expected := &ledgerqueries.DataResponse{
		LedgerHash: "842B57C1CC0613299A686D3E9F310EC0422C84D3911E5056389AA7E5808A93C8",
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetLedgerData(&ledgerqueries.DataRequest{
		Binary: true,
	})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "ledger_data", request()["command"])
}

func TestClient_GetLedger(t *testing.T) {
	expected := &ledgerqueries.
		Response{LedgerIndex: 123}

	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetLedger(&ledgerqueries.Request{})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "ledger", request()["command"])
}

func TestClient_GetNFTBuyOffers(t *testing.T) {
	expected := &nft.NFTokenBuyOffersResponse{
		NFTokenID: "00080000B4F4AFC5FBCBD76873F18006173D2193467D3EE70000099B00000000",
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetNFTBuyOffers(&nft.NFTokenBuyOffersRequest{
		NFTokenID: "00080000B4F4AFC5FBCBD76873F18006173D2193467D3EE70000099B00000000",
	})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "nft_buy_offers", request()["command"])
}

func TestClient_GetNFTSellOffers(t *testing.T) {
	expected := &nft.NFTokenSellOffersResponse{NFTokenID: "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007"}

	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetNFTSellOffers(&nft.NFTokenSellOffersRequest{
		NFTokenID: "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
	})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "nft_sell_offers", request()["command"])
}

func TestClient_GetBookOffers(t *testing.T) {
	expected := &path.BookOffersResponse{
		LedgerCurrentIndex: 1234,
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetBookOffers(&path.BookOffersRequest{
		TakerGets: pathtypes.BookOfferCurrency{
			Currency: "USD",
			Issuer:   "rLpSRZ1MyZkCkCuXhyXmFKXdvuZDZP3XKt",
		},
		TakerPays: pathtypes.BookOfferCurrency{
			Currency: "USD",
			Issuer:   "rLpSRZ1MyZkCkCuXhyXmFKXdvuZDZP3XKt",
		},
	})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "book_offers", request()["command"])
}

func TestClient_GetDepositAuthorized(t *testing.T) {
	expected := &path.DepositAuthorizedResponse{
		DepositAuthorized: true,
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetDepositAuthorized(&path.DepositAuthorizedRequest{
		SourceAccount:      "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		DestinationAccount: "rEhxGqkqPPSxQ3P25J66ft5TwpzV14k2de",
	})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "deposit_authorized", request()["command"])
}

func TestClient_FindPathCreate(t *testing.T) {
	expected := &path.FindResponse{SourceAccount: "rSource"}

	client, request := setupQueryTestClient(t, expected)
	response, err := client.FindPathCreate(&path.FindCreateRequest{Subcommand: path.Create})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	got := request()
	require.Equal(t, "path_find", got["command"])
	require.Equal(t, "create", got["subcommand"])
}

func TestClient_FindPathClose(t *testing.T) {
	expected := &path.FindResponse{SourceAccount: "rSource"}

	client, request := setupQueryTestClient(t, expected)
	response, err := client.FindPathClose(&path.FindCloseRequest{Subcommand: path.Close})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	got := request()
	require.Equal(t, "path_find", got["command"])
	require.Equal(t, "close", got["subcommand"])
}

func TestClient_FindPathStatus(t *testing.T) {
	expected := &path.FindResponse{SourceAccount: "rSource"}

	client, request := setupQueryTestClient(t, expected)
	response, err := client.FindPathStatus(&path.FindStatusRequest{Subcommand: path.Status})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	got := request()
	require.Equal(t, "path_find", got["command"])
	require.Equal(t, "status", got["subcommand"])
}

func TestClient_GetRipplePathFind(t *testing.T) {
	expected := &path.RipplePathFindResponse{DestinationAccount: "rDestination"}

	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetRipplePathFind(&path.RipplePathFindRequest{})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "ripple_path_find", request()["command"])
}

func TestClient_GetServerInfo(t *testing.T) {
	expected := &server.InfoResponse{
		Info: servertypes.Info{
			BuildVersion: "1.9.4",
		},
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetServerInfo(&server.InfoRequest{})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "server_info", request()["command"])
}

func TestClient_GetAllFeatures(t *testing.T) {
	expected := &server.FeatureAllResponse{
		Features: map[string]servertypes.FeatureStatus{
			"42": {
				Enabled: true,
			},
		},
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetAllFeatures(&server.FeatureAllRequest{})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	got := request()
	require.Equal(t, "feature", got["command"])
	require.NotContains(t, got, "feature")
}

func TestClient_GetFeature(t *testing.T) {
	expected := &server.FeatureResponse{
		"feature": servertypes.FeatureStatus{
			Enabled: false,
		},
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetFeature(&server.FeatureOneRequest{
		Feature: "TrustSetAuth",
	})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	got := request()
	require.Equal(t, "feature", got["command"])
	require.Equal(t, "TrustSetAuth", got["feature"])
}

func TestClient_GetFee(t *testing.T) {
	expected := &server.FeeResponse{
		CurrentLedgerSize: "14",
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetFee(&server.FeeRequest{})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "fee", request()["command"])
}

func TestClient_GetManifest(t *testing.T) {
	expected := &server.ManifestResponse{
		Details: server.ManifestDetails{
			MasterKey: "nHUon2tpyJEHHYGmxqeGu37cvPYHzrMtUNQyNinmg4rMdN5q58Bt",
		},
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetManifest(&server.ManifestRequest{
		PublicKey: "nHUon2tpyJEHHYGmxqeGu37cvPYHzrMtUNQyNinmg4rMdN5q58Bt",
	})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "manifest", request()["command"])
}

func TestClient_GetServerState(t *testing.T) {
	expected := &server.StateResponse{
		State: servertypes.State{
			BuildVersion: "1.9.4",
		},
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetServerState(&server.StateRequest{})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "server_state", request()["command"])
}

func TestClient_GetAggregatePrice(t *testing.T) {
	expected := &oracle.GetAggregatePriceResponse{
		Median: "123.45",
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetAggregatePrice(&oracle.GetAggregatePriceRequest{})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "get_aggregate_price", request()["command"])
}

func TestClient_Ping(t *testing.T) {
	expected := &utility.PingResponse{
		Role: "admin",
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.Ping(&utility.PingRequest{})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "ping", request()["command"])
}

func TestClient_GetRandom(t *testing.T) {
	expected := &utility.RandomResponse{
		Random: "8ED765AEBBD6767603C2C9375B2679AEF42BC63BE8B160B10982D767EC309E44",
	}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetRandom(&utility.RandomRequest{})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "random", request()["command"])
}

func TestClient_XRPBalance(t *testing.T) {
	client, request := setupQueryTestClient(t, &account.InfoResponse{AccountData: ledger.AccountRoot{Balance: 1234567}})
	balance, err := client.GetXrpBalance("rAccount")
	require.NoError(t, err)
	require.Equal(t, "1.234567", balance)
	got := request()
	require.Equal(t, "account_info", got["command"])
	require.Equal(t, "rAccount", got["account"])
	require.NotContains(t, got, "ledger_index")
}

func TestClient_ValidatedXRPBalance(t *testing.T) {
	client, request := setupQueryTestClient(t, &account.InfoResponse{AccountData: ledger.AccountRoot{Balance: 1234567}})
	balance, err := client.GetXrpBalanceValidated("rAccount")
	require.NoError(t, err)
	require.Equal(t, "1.234567", balance)
	got := request()
	require.Equal(t, "account_info", got["command"])
	require.Equal(t, "rAccount", got["account"])
	require.Equal(t, "validated", got["ledger_index"])
}

func TestClient_ValidatedXRPDropsBalance(t *testing.T) {
	client, request := setupQueryTestClient(t, &account.InfoResponse{AccountData: ledger.AccountRoot{Balance: 1234567}})
	balance, err := client.GetXrpDropsBalanceValidated("rAccount")
	require.NoError(t, err)
	require.Equal(t, types.XRPCurrencyAmount(1234567), balance)
	got := request()
	require.Equal(t, "account_info", got["command"])
	require.Equal(t, "rAccount", got["account"])
	require.Equal(t, "validated", got["ledger_index"])
}

func TestClient_GetLedgerIndex(t *testing.T) {
	client, request := setupQueryTestClient(t, &ledgerqueries.Response{LedgerIndex: 123})
	index, err := client.GetLedgerIndex()
	require.NoError(t, err)
	require.Equal(t, common.LedgerIndex(123), index)
	got := request()
	require.Equal(t, "ledger", got["command"])
	require.Equal(t, "validated", got["ledger_index"])
}

func TestClient_GetVaultInfo(t *testing.T) {
	expected := &vault.Response{Vault: vault.Vault{Account: "rAccount"}}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetVaultInfo(&vault.InfoRequest{VaultID: "20B136D7BF6D2E3D610E28E3E6BE09F5C8F4F0241BBF6E2D072AE1BACB1388F5"})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	require.Equal(t, "vault_info", request()["command"])
}

func TestClient_GetAMMInfo(t *testing.T) {
	expected := &amm.InfoResponse{AMM: amm.Info{
		Account: "rAMM",
		Amount:  types.XRPCurrencyAmount(1000000),
		Amount2: types.IssuedCurrencyAmount{Currency: "USD", Issuer: "rIssuer", Value: "1"},
	}}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetAMMInfo(&amm.InfoRequest{AMMAccount: "rAMM"})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	got := request()
	require.Equal(t, "amm_info", got["command"])
	require.Equal(t, "rAMM", got["amm_account"])
}
