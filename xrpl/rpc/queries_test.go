package rpc

import (
	"encoding/json"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	account "github.com/Peersyst/xrpl-go/xrpl/queries/account"
	accounttypes "github.com/Peersyst/xrpl-go/xrpl/queries/account/types"
	channel "github.com/Peersyst/xrpl-go/xrpl/queries/channel"
	common "github.com/Peersyst/xrpl-go/xrpl/queries/common"
	ledgerqueries "github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	ledgertypes "github.com/Peersyst/xrpl-go/xrpl/queries/ledger/types"
	nft "github.com/Peersyst/xrpl-go/xrpl/queries/nft"
	nfttypes "github.com/Peersyst/xrpl-go/xrpl/queries/nft/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/oracle"
	path "github.com/Peersyst/xrpl-go/xrpl/queries/path"
	pathtypes "github.com/Peersyst/xrpl-go/xrpl/queries/path/types"
	server "github.com/Peersyst/xrpl-go/xrpl/queries/server"
	servertypes "github.com/Peersyst/xrpl-go/xrpl/queries/server/types"
	utility "github.com/Peersyst/xrpl-go/xrpl/queries/utility"
	"github.com/Peersyst/xrpl-go/xrpl/rpc/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestClient_GetAccountInfo(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *account.InfoRequest
		expected      account.InfoResponse
		expectedError string
	}{
		{
			name: "successful account info request",
			mockResponse: `{
				"result": {
					"account_data": {
						"Account": "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
						"Balance": "999999999960",
						"Flags": 0,
						"LedgerEntryType": "AccountRoot",
						"OwnerCount": 0
					}
				}
			}`,
			mockStatus: 200,
			request: &account.InfoRequest{
				Account: "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
			},
			expected: account.InfoResponse{
				AccountData: ledger.AccountRoot{
					Account:         "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
					Balance:         types.XRPCurrencyAmount(999999999960),
					Flags:           0,
					LedgerEntryType: "AccountRoot",
					OwnerCount:      0,
				},
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "actNotFound",
					"status": "error"
				}
			}`,
			mockStatus: 200,
			request: &account.InfoRequest{
				Account: "rInvalidAccount",
			},
			expectedError: "actNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var accountInfoResp account.InfoResponse
			err = resp.GetResult(&accountInfoResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, accountInfoResp)
		})
	}
}

func TestClient_GetAccountChannels(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *account.ChannelsRequest
		expected      account.ChannelsResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"account": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
					"channels": [
						{
							"account": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
							"amount": "1000",
							"balance": "0",
							"channel_id": "C7F634794B79DB40E87179A9D1BF05D05797AE7E92DF8E93FD6656E8C4BE3AE7",
							"destination_account": "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX",
							"public_key": "aBR7mdD75Ycs8DRhMgQ4EMUEmBArF8SEh1hfjrT2V9DQTLNbJVqw",
							"public_key_hex": "03CFD18E689434F032A4E84C63E2A3A6472D684EAF4FD52CA67742F3E24BAE81B2",
							"settle_delay": 60
						}
					],
					"ledger_hash": "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
					"ledger_index": 71766343,
					"validated": true
				}
			}`,
			mockStatus: 200,
			request: &account.ChannelsRequest{
				Account: "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
			},
			expected: account.ChannelsResponse{
				Account: "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
				Channels: []accounttypes.ChannelResult{{
					Account:            "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
					Amount:             "1000",
					Balance:            "0",
					ChannelID:          "C7F634794B79DB40E87179A9D1BF05D05797AE7E92DF8E93FD6656E8C4BE3AE7",
					DestinationAccount: "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX",
					PublicKey:          "aBR7mdD75Ycs8DRhMgQ4EMUEmBArF8SEh1hfjrT2V9DQTLNbJVqw",
					PublicKeyHex:       "03CFD18E689434F032A4E84C63E2A3A6472D684EAF4FD52CA67742F3E24BAE81B2",
					SettleDelay:        60,
				}},
				LedgerHash:  "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
				LedgerIndex: 71766343,
				Validated:   true,
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "actNotFound",
					"status": "error"
				}
			}`,
			mockStatus: 200,
			request: &account.ChannelsRequest{
				Account: "rInvalidAccount",
			},
			expectedError: "actNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var accountChannelsResp account.ChannelsResponse
			err = resp.GetResult(&accountChannelsResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, accountChannelsResp)
		})
	}
}

func TestClient_GetAccountObjects(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *account.ObjectsRequest
		expected      account.ObjectsResponse
		expectedError string
	}{
		{
			name: "successful account objects request",
			mockResponse: `{
				"result": {
					"account": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
					"account_objects": [
						{
							"Balance": {
								"currency": "USD",
								"issuer": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
								"value": "100"
							},
							"Flags": 65536,
							"HighLimit": {
								"currency": "USD",
								"issuer": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
								"value": "0"
							},
							"LedgerEntryType": "RippleState",
							"LowLimit": {
								"currency": "USD",
								"issuer": "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
								"value": "500"
							}
						}
					],
					"ledger_hash": "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
					"ledger_index": 71766343,
					"validated": true
				}
			}`,
			mockStatus: 200,
			request: &account.ObjectsRequest{
				Account: "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
			},
			expected: account.ObjectsResponse{
				Account: "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
				AccountObjects: []ledger.FlatLedgerObject{
					{
						"Balance": map[string]any{
							"currency": "USD",
							"issuer":   "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
							"value":    "100",
						},
						"Flags": json.Number("65536"),
						"HighLimit": map[string]any{
							"currency": "USD",
							"issuer":   "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
							"value":    "0",
						},
						"LedgerEntryType": "RippleState",
						"LowLimit": map[string]any{
							"currency": "USD",
							"issuer":   "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
							"value":    "500",
						},
					},
				},
				LedgerHash:  "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
				LedgerIndex: 71766343,
				Validated:   true,
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "actNotFound",
					"status": "error"
				}
			}`,
			mockStatus: 200,
			request: &account.ObjectsRequest{
				Account: "rInvalidAccount",
			},
			expectedError: "actNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var accountObjectsResp account.ObjectsResponse
			err = resp.GetResult(&accountObjectsResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, accountObjectsResp)
		})
	}
}

func TestClient_GetAccountLines(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *account.LinesRequest
		expected      account.LinesResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"account": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
					"lines": [
						{
							"account": "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX",
							"balance": "10",
							"currency": "USD",
							"limit": "100",
							"limit_peer": "0",
							"quality_in": 0,
							"quality_out": 0
						}
					],
					"ledger_hash": "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
					"ledger_index": 71766343,
					"validated": true
				}
			}`,
			mockStatus: 200,
			request: &account.LinesRequest{
				Account: "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
			},
			expected: account.LinesResponse{
				Account: "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
				Lines: []accounttypes.TrustLine{
					{
						Account:    "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX",
						Balance:    "10",
						Currency:   "USD",
						Limit:      "100",
						LimitPeer:  "0",
						QualityIn:  0,
						QualityOut: 0,
					},
				},
				LedgerHash:  "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
				LedgerIndex: 71766343,
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "actNotFound",
					"status": "error"
				}
			}`,
			mockStatus: 200,
			request: &account.LinesRequest{
				Account: "rInvalidAccount",
			},
			expectedError: "actNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var accountLinesResp account.LinesResponse
			err = resp.GetResult(&accountLinesResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, accountLinesResp)
		})
	}
}

func TestClient_GetXrpBalance(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		address       string
		expected      string
		expectedError string
	}{
		{
			name: "successful balance request",
			mockResponse: `{
				"result": {
					"account_data": {
						"Account": "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
						"Balance": "999999999960",
						"Flags": 0,
						"LedgerEntryType": "AccountRoot",
						"OwnerCount": 0
					}
				}
			}`,
			mockStatus: 200,
			address:    "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
			expected:   "999999.99996",
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "actNotFound",
					"status": "error"
				}
			}`,
			mockStatus:    200,
			address:       "rInvalidAccount",
			expectedError: "actNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			balance, err := client.GetXrpBalance(types.Address(tt.address))

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, balance)
		})
	}
}

func TestClient_GetAccountNFTs(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *account.NFTsRequest
		expected      account.NFTsResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"account": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
					"account_nfts": [
						{
							"Flags": 0,
							"Issuer": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
							"NFTokenID": "00080000B4F4AFC5FBCBD76873F18006173D2193467D3EE70000099B00000000",
							"NFTokenTaxon": 0,
							"URI": "697066733A2F2F516D516A447644686F686B6B6454716D78313959724D44697350"
						}
					],
					"ledger_hash": "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
					"ledger_index": 71766343,
					"validated": true
				}
			}`,
			mockStatus: 200,
			request: &account.NFTsRequest{
				Account: "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
			},
			expected: account.NFTsResponse{
				Account: "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
				AccountNFTs: []accounttypes.NFT{
					{
						Flags:        0,
						Issuer:       "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
						NFTokenID:    "00080000B4F4AFC5FBCBD76873F18006173D2193467D3EE70000099B00000000",
						NFTokenTaxon: 0,
						URI:          "697066733A2F2F516D516A447644686F686B6B6454716D78313959724D44697350",
					},
				},
				LedgerHash:  "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
				LedgerIndex: 71766343,
				Validated:   true,
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "actNotFound",
					"status": "error"
				}
			}`,
			mockStatus: 200,
			request: &account.NFTsRequest{
				Account: "rInvalidAccount",
			},
			expectedError: "actNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var accountNFTsResp account.NFTsResponse
			err = resp.GetResult(&accountNFTsResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, accountNFTsResp)
		})
	}
}

func TestClient_GetAccountCurrencies(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *account.CurrenciesRequest
		expected      account.CurrenciesResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"ledger_hash": "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
					"ledger_index": 71766343,
					"receive_currencies": ["USD", "EUR"],
					"send_currencies": ["USD", "EUR"],
					"validated": true
				}
			}`,
			mockStatus: 200,
			request: &account.CurrenciesRequest{
				Account: "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
			},
			expected: account.CurrenciesResponse{
				LedgerHash:        "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
				LedgerIndex:       71766343,
				ReceiveCurrencies: []string{"USD", "EUR"},
				SendCurrencies:    []string{"USD", "EUR"},
				Validated:         true,
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "actNotFound",
					"status": "error"
				}
			}`,
			mockStatus: 200,
			request: &account.CurrenciesRequest{
				Account: "rInvalidAccount",
			},
			expectedError: "actNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var accountCurrenciesResp account.CurrenciesResponse
			err = resp.GetResult(&accountCurrenciesResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, accountCurrenciesResp)
		})
	}
}

func TestClient_GetAccountOffers(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *account.OffersRequest
		expected      account.OffersResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"account": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
					"offers": [
						{
							"flags": 0,
							"quality": "1",
							"seq": 1234,
							"taker_gets": {
								"currency": "USD",
								"issuer": "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX",
								"value": "100"
							},
							"taker_pays": "100000000"
						}
					],
					"ledger_hash": "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
					"ledger_index": 71766343,
					"validated": true
				}
			}`,
			mockStatus: 200,
			request: &account.OffersRequest{
				Account: "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
			},
			expected: account.OffersResponse{
				Account: "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
				Offers: []accounttypes.OfferResult{
					{
						Flags:    0,
						Quality:  "1",
						Sequence: 1234,
						TakerGets: map[string]any{
							"currency": "USD",
							"issuer":   "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX",
							"value":    "100",
						},
						TakerPays: "100000000",
					},
				},
				LedgerHash:  "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
				LedgerIndex: 71766343,
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "actNotFound",
					"status": "error"
				}
			}`,
			mockStatus: 200,
			request: &account.OffersRequest{
				Account: "rInvalidAccount",
			},
			expectedError: "actNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var accountOffersResp account.OffersResponse
			err = resp.GetResult(&accountOffersResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, accountOffersResp)
		})
	}
}

func TestClient_GetAccountTransactions(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *account.TransactionsRequest
		expected      account.TransactionsResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"account": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
					"transactions": [
						{
							"tx_json": {
								"Account": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
								"Amount": "100000000",
								"Destination": "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX",
								"Fee": "12",
								"Flags": 2147483648,
								"Sequence": 1,
								"SigningPubKey": "0330E7FC9D56BB25D6893BA3F317AE5BCF33B3291BD63DB32654A313222F7FD020",
								"TransactionType": "Payment",
								"TxnSignature": "3045022100A7CCD1B5F67D76C7F2C0B6C199C9D2F3721A14A3C69F3CB134E9BF9D9DD6F8B002206A5F974C4F4D07B4F5A99391BF3E93D9B0A7346861E12E924B6451B6D8ED4F09",
								"hash": "2E2DDBF5B8F29AEED7494CC0A863A93A4BD3C066BF880A577C5C466EE2C637DF"
							},
							"meta": {
								"TransactionIndex": 1,
								"TransactionResult": "tesSUCCESS"
							},
							"validated": true
						}
					],
					"ledger_index_min": 71766300,
					"ledger_index_max": 71766343,
					"validated": true
				}
			}`,
			mockStatus: 200,
			request: &account.TransactionsRequest{
				Account: "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
			},
			expected: account.TransactionsResponse{
				Account: "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
				Transactions: []account.Transaction{
					{
						Tx: map[string]any{
							"Account":         "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
							"Amount":          "100000000",
							"Destination":     "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX",
							"Fee":             "12",
							"Flags":           json.Number("2147483648"),
							"Sequence":        json.Number("1"),
							"SigningPubKey":   "0330E7FC9D56BB25D6893BA3F317AE5BCF33B3291BD63DB32654A313222F7FD020",
							"TransactionType": "Payment",
							"TxnSignature":    "3045022100A7CCD1B5F67D76C7F2C0B6C199C9D2F3721A14A3C69F3CB134E9BF9D9DD6F8B002206A5F974C4F4D07B4F5A99391BF3E93D9B0A7346861E12E924B6451B6D8ED4F09",
							"hash":            "2E2DDBF5B8F29AEED7494CC0A863A93A4BD3C066BF880A577C5C466EE2C637DF",
						},
						Meta: transaction.TxObjMeta{
							TransactionIndex:  1,
							TransactionResult: "tesSUCCESS",
						},
						Validated: true,
					},
				},
				LedgerIndexMin: 71766300,
				LedgerIndexMax: 71766343,
				Validated:      true,
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "actNotFound",
					"status": "error"
				}
			}`,
			mockStatus: 200,
			request: &account.TransactionsRequest{
				Account: "rInvalidAccount",
			},
			expectedError: "actNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var accountTransactionsResp account.TransactionsResponse
			err = resp.GetResult(&accountTransactionsResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, accountTransactionsResp)
		})
	}
}

func TestClient_GetGatewayBalances(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *account.GatewayBalancesRequest
		expected      account.GatewayBalancesResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"account": "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
					"assets": {
						"r9F6wk8HkXrgYWoJ7fsv4VrUBVoqDVtzkH": [
							{
								"currency": "BTC",
								"value": "5444166510000000e-26"
							}
						]
					},
					"balances": {
						"rKm4uWpg9tfwbVSeATv4KxDe6mpE9yPkgJ": [
							{
								"currency": "EUR",
								"value": "29826.1965999999"
							}
						],
						"ra7JkEzrgeKHdzKgo4EUUVBnxggY4z37kt": [
							{
								"currency": "USD",
								"value": "13857.70416"
							}
						]
					},
					"ledger_hash": "61DDBF304AF6E8101576BF161D447CA8E4F0170DDFBEAFFD993DC9383D443388",
					"ledger_index": 14483212,
					"obligations": {
						"EUR": "5599.716599999999",
						"USD": "12345.9"
					},
					"status": "success"
				}
			}`,
			mockStatus: 200,
			request: &account.GatewayBalancesRequest{
				Account: "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
			},
			expected: account.GatewayBalancesResponse{
				Account: "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
				Assets: map[string][]account.GatewayBalance{
					"r9F6wk8HkXrgYWoJ7fsv4VrUBVoqDVtzkH": {
						{
							Currency: "BTC",
							Value:    "5444166510000000e-26",
						},
					},
				},
				Balances: map[string][]account.GatewayBalance{
					"rKm4uWpg9tfwbVSeATv4KxDe6mpE9yPkgJ": {
						{
							Currency: "EUR",
							Value:    "29826.1965999999",
						},
					},
					"ra7JkEzrgeKHdzKgo4EUUVBnxggY4z37kt": {
						{
							Currency: "USD",
							Value:    "13857.70416",
						},
					},
				},
				LedgerHash:  "61DDBF304AF6E8101576BF161D447CA8E4F0170DDFBEAFFD993DC9383D443388",
				LedgerIndex: 14483212,
				Obligations: map[string]string{
					"EUR": "5599.716599999999",
					"USD": "12345.9",
				},
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "actNotFound",
					"status": "error"
				}
			}`,
			mockStatus: 200,
			request: &account.GatewayBalancesRequest{
				Account: "rInvalidAccount",
			},
			expectedError: "actNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var gatewayBalancesResp account.GatewayBalancesResponse
			err = resp.GetResult(&gatewayBalancesResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, gatewayBalancesResp)
		})
	}
}

func TestClient_GetChannelVerify(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *channel.VerifyRequest
		expected      channel.VerifyResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"signature_verified": true
				}
			}`,
			mockStatus: 200,
			request: &channel.VerifyRequest{
				Amount:    types.XRPCurrencyAmount(1000000),
				ChannelID: "5DB01B7FFED6B67E6B0414DED11E051D2EE2B7619CE0EAA6286D67A3A4D5BDB3",
				PublicKey: "023693F15967AE357D0327974AD46FE3C127113B1110D6044FD41E723689F81CC6",
				Signature: "304402204EF0AFB78AC23ED1C472E74F4299C0C21F1B21D07EFC0A3838A420F76D783A400220154FB11B6F54320666E4C36CA7F686C16A3A0456800BBC43746F34AF50290064",
			},
			expected: channel.VerifyResponse{
				SignatureVerified: true,
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "channelNotFound",
					"status": "error"
				}
			}`,
			mockStatus: 200,
			request: &channel.VerifyRequest{
				Amount:    types.XRPCurrencyAmount(1000000),
				ChannelID: "invalidChannel",
			},
			expectedError: "channelNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var channelVerifyResp channel.VerifyResponse
			err = resp.GetResult(&channelVerifyResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, channelVerifyResp)
		})
	}
}

func TestClient_GetLedgerIndex(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		expected      common.LedgerIndex
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"ledger_hash": "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
					"ledger_index": 71766343,
					"validated": true
				}
			}`,
			mockStatus: 200,
			expected:   71766343,
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "ledgerNotFound",
					"status": "error"
				}
			}`,
			mockStatus:    200,
			expectedError: "ledgerNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			ledgerIndex, err := client.GetLedgerIndex()

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, ledgerIndex)
		})
	}
}

func TestClient_GetLedgerEntry(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *ledgerqueries.EntryRequest
		expected      *ledgerqueries.EntryResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"index": "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
					"ledger_current_index": 61809073,
					"node_binary": "test",
					"validated": true,
					"deleted_ledger_index": 0
				}
			}`,
			mockStatus: 200,
			request: &ledgerqueries.EntryRequest{
				Index: "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
			},
			expected: &ledgerqueries.EntryResponse{
				Index:              "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
				LedgerCurrentIndex: 61809073,
				NodeBinary:         "test",
				Validated:          true,
				DeletedLedgerIndex: 0,
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "entryNotFound",
					"status": "error"
				}
			}`,
			mockStatus:    200,
			request:       &ledgerqueries.EntryRequest{},
			expectedError: "entryNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.GetLedgerEntry(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, resp)
		})
	}
}


func TestClient_GetClosedLedger(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		expected      *ledgerqueries.ClosedResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"ledger_hash": "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
					"ledger_index": 71766343
				}
			}`,
			mockStatus: 200,
			expected: &ledgerqueries.ClosedResponse{
				LedgerHash:  "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
				LedgerIndex: 71766343,
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "ledgerNotFound",
					"status": "error"
				}
			}`,
			mockStatus:    200,
			expectedError: "ledgerNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.GetClosedLedger()

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, resp)
		})
	}
}

func TestClient_GetCurrentLedger(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		expected      *ledgerqueries.CurrentResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"ledger_current_index": 71766343
				}
			}`,
			mockStatus: 200,
			expected: &ledgerqueries.CurrentResponse{
				LedgerCurrentIndex: 71766343,
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "ledgerNotFound",
					"status": "error"
				}
			}`,
			mockStatus:    200,
			expectedError: "ledgerNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.GetCurrentLedger()

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, resp)
		})
	}
}

func TestClient_GetLedgerData(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *ledgerqueries.DataRequest
		expected      ledgerqueries.DataResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"ledger_hash": "842B57C1CC0613299A686D3E9F310EC0422C84D3911E5056389AA7E5808A93C8",
					"ledger_index": "6",
					"state": [
						{
							"data": "0000000000000000",
							"index": "1B8590C01B0006EDFA9ED60296DD052DC5E90F99147FE0D93"
						}
					]
				}
			}`,
			mockStatus: 200,
			request: &ledgerqueries.DataRequest{
				Binary: true,
			},
			expected: ledgerqueries.DataResponse{
				LedgerHash:  "842B57C1CC0613299A686D3E9F310EC0422C84D3911E5056389AA7E5808A93C8",
				LedgerIndex: "6",
				State: []ledgertypes.State{
					{
						Data:  "0000000000000000",
						Index: "1B8590C01B0006EDFA9ED60296DD052DC5E90F99147FE0D93",
					},
				},
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "ledgerNotFound",
					"status": "error"
				}
			}`,
			mockStatus: 200,
			request: &ledgerqueries.DataRequest{
				Binary: true,
			},
			expectedError: "ledgerNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var ledgerDataResp ledgerqueries.DataResponse
			err = resp.GetResult(&ledgerDataResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, ledgerDataResp)
		})
	}
}

func TestClient_GetLedger(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *ledgerqueries.Request
		expected      ledgerqueries.Response
		expectedError string
	}{
		{
			name: "successful ledger request",
			mockResponse: `{
				"result": {
					"ledger": {
						"account_hash": "A6B48E079B37B7803C21B6F2B9B9169A8498B58DEF2D0783E8BC5493A80832C9",
						"close_flags": 0,
						"close_time": 714214410,
						"close_time_human": "2022-Aug-19 22:20:10.000000000 UTC",
						"close_time_resolution": 10,
						"closed": true,
						"ledger_hash": "E6DB7365949BF9814D76BCC730B01818EB9136A89DB224F3F9F5C3E53B52551C",
						"ledger_index": 71768313,
						"parent_close_time": 714214402,
						"parent_hash": "B508A40BB4E88A778EFDD6B8DB1872C531D4E58B5EE5A4A9E7D7F5C5F3715D6F",
						"transaction_hash": "FC6FFCB71B2527DDD630EE5409D38913B4D4C026AA6C3B14A3E9D4ED45CFE30D"
					},
					"ledger_hash": "E6DB7365949BF9814D76BCC730B01818EB9136A89DB224F3F9F5C3E53B52551C",
					"ledger_index": 71768313,
					"validated": true
				}
			}`,
			mockStatus: 200,
			request:    &ledgerqueries.Request{},
			expected: ledgerqueries.Response{
				Ledger: ledgertypes.BaseLedger{
					AccountHash:         "A6B48E079B37B7803C21B6F2B9B9169A8498B58DEF2D0783E8BC5493A80832C9",
					CloseFlags:          0,
					CloseTime:           714214410,
					CloseTimeHuman:      "2022-Aug-19 22:20:10.000000000 UTC",
					CloseTimeResolution: 10,
					Closed:              true,
					LedgerHash:          "E6DB7365949BF9814D76BCC730B01818EB9136A89DB224F3F9F5C3E53B52551C",
					LedgerIndex:         71768313,
					ParentCloseTime:     714214402,
					ParentHash:          "B508A40BB4E88A778EFDD6B8DB1872C531D4E58B5EE5A4A9E7D7F5C5F3715D6F",
					TransactionHash:     "FC6FFCB71B2527DDD630EE5409D38913B4D4C026AA6C3B14A3E9D4ED45CFE30D",
				},
				LedgerHash:  "E6DB7365949BF9814D76BCC730B01818EB9136A89DB224F3F9F5C3E53B52551C",
				LedgerIndex: 71768313,
				Validated:   true,
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "ledgerNotFound",
					"status": "error"
				}
			}`,
			mockStatus:    200,
			request:       &ledgerqueries.Request{},
			expectedError: "ledgerNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var ledgerResp ledgerqueries.Response
			err = resp.GetResult(&ledgerResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, ledgerResp)
		})
	}
}

func TestClient_GetNFTBuyOffers(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *nft.NFTokenBuyOffersRequest
		expected      nft.NFTokenBuyOffersResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"nft_id": "00080000B4F4AFC5FBCBD76873F18006173D2193467D3EE70000099B00000000",
					"offers": [
						{
							"amount": "1000000",
							"flags": 0,
							"nft_offer_index": "9B142E2F0F1D31AAE7883C3F65D59D2E4A50BF7A18BE0D6987D0F34B4E4A37FF",
							"owner": "rLpSRZ1MyZkCkCuXhyXmFKXdvuZDZP3XKt"
						}
					]
				}
			}`,
			mockStatus: 200,
			request: &nft.NFTokenBuyOffersRequest{
				NFTokenID: "00080000B4F4AFC5FBCBD76873F18006173D2193467D3EE70000099B00000000",
			},
			expected: nft.NFTokenBuyOffersResponse{
				NFTokenID: "00080000B4F4AFC5FBCBD76873F18006173D2193467D3EE70000099B00000000",
				Offers: []nfttypes.NFTokenOffer{
					{
						Amount:            "1000000",
						Flags:             0,
						NFTokenOfferIndex: "9B142E2F0F1D31AAE7883C3F65D59D2E4A50BF7A18BE0D6987D0F34B4E4A37FF",
						Owner:             "rLpSRZ1MyZkCkCuXhyXmFKXdvuZDZP3XKt",
					},
				},
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "nftNotFound",
					"status": "error"
				}
			}`,
			mockStatus: 200,
			request: &nft.NFTokenBuyOffersRequest{
				NFTokenID: "invalidNFTID",
			},
			expectedError: "nftNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var nftResp nft.NFTokenBuyOffersResponse
			err = resp.GetResult(&nftResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, nftResp)
		})
	}
}

func TestClient_GetNFTSellOffers(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *nft.NFTokenSellOffersRequest
		expected      nft.NFTokenSellOffersResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"offers": [
						{
							"amount": "1000000",
							"flags": 1,
							"nft_offer_index": "049E35762ABDF8E9F810E5CE4C39AA232F8A778B2AA6F5F74987BA4D1BE95F1C",
							"owner": "rLpSRZ1MyZkCkCuXhyXmFKXdvuZDZP3XKt"
						}
					]
				}
			}`,
			mockStatus: 200,
			request: &nft.NFTokenSellOffersRequest{
				NFTokenID: "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
			},
			expected: nft.NFTokenSellOffersResponse{
				Offers: []nfttypes.NFTokenOffer{
					{
						Amount:            "1000000",
						Flags:             1,
						NFTokenOfferIndex: "049E35762ABDF8E9F810E5CE4C39AA232F8A778B2AA6F5F74987BA4D1BE95F1C",
						Owner:             "rLpSRZ1MyZkCkCuXhyXmFKXdvuZDZP3XKt",
					},
				},
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "nftNotFound",
					"status": "error"
				}
			}`,
			mockStatus: 200,
			request: &nft.NFTokenSellOffersRequest{
				NFTokenID: "invalidNFTID",
			},
			expectedError: "nftNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var nftResp nft.NFTokenSellOffersResponse
			err = resp.GetResult(&nftResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, nftResp)
		})
	}
}

func TestClient_GetBookOffers(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *path.BookOffersRequest
		expected      path.BookOffersResponse
		expectedError string
	}{
		{
			name: "successful book offers request",
			mockResponse: `{
				"result": {
					"ledger_current_index": 1234,
					"offers": [
						{
							"Account": "rLpSRZ1MyZkCkCuXhyXmFKXdvuZDZP3XKt",
							"BookDirectory": "DFA3B6DDAB58C7E8E5D944E736DA4B7046C30E4F460FD9DE4C124AF94ED1781B",
							"BookNode": "0000000000000000",
							"Flags": 0,
							"LedgerEntryType": "Offer",
							"OwnerNode": "0000000000000000",
							"PreviousTxnID": "F0AB71E777B2DA54B86231E19B82554EF1F8211F98007E68D5B7D1BE76D11F3A",
							"PreviousTxnLgrSeq": 1234,
							"Sequence": 1,
							"taker_gets_funded": "1000000",
							"taker_pays_funded": "100",
							"quality": "100"
						}
					]
				}
			}`,
			mockStatus: 200,
			request: &path.BookOffersRequest{
				TakerGets: pathtypes.BookOfferCurrency{
					Currency: "USD",
					Issuer:   "rLpSRZ1MyZkCkCuXhyXmFKXdvuZDZP3XKt",
				},
				TakerPays: pathtypes.BookOfferCurrency{
					Currency: "USD",
					Issuer:   "rLpSRZ1MyZkCkCuXhyXmFKXdvuZDZP3XKt",
				},
			},
			expected: path.BookOffersResponse{
				LedgerCurrentIndex: 1234,
				Offers: []pathtypes.BookOffer{
					{
						Account:           "rLpSRZ1MyZkCkCuXhyXmFKXdvuZDZP3XKt",
						BookDirectory:     "DFA3B6DDAB58C7E8E5D944E736DA4B7046C30E4F460FD9DE4C124AF94ED1781B",
						BookNode:          "0000000000000000",
						Flags:             0,
						LedgerEntryType:   "Offer",
						OwnerNode:         "0000000000000000",
						PreviousTxnID:     "F0AB71E777B2DA54B86231E19B82554EF1F8211F98007E68D5B7D1BE76D11F3A",
						PreviousTxnLgrSeq: 1234,
						Sequence:          1,
						TakerGetsFunded:   "1000000",
						TakerPaysFunded:   "100",
						Quality:           "100",
					},
				},
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "invalidParams",
					"status": "error"
				}
			}`,
			mockStatus:    200,
			request:       &path.BookOffersRequest{},
			expectedError: "invalidParams",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var bookResp path.BookOffersResponse
			err = resp.GetResult(&bookResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, bookResp)
		})
	}
}

func TestClient_GetDepositAuthorized(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *path.DepositAuthorizedRequest
		expected      path.DepositAuthorizedResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"deposit_authorized": true,
					"destination_account": "rEhxGqkqPPSxQ3P25J66ft5TwpzV14k2de",
					"ledger_current_index": 14,
					"source_account": "rsUiUMpnrgxQp24dJYZDhmV4bE3aBtQyt8",
					"validated": false,
					"status": "success"
				}
			}`,
			mockStatus: 200,
			request: &path.DepositAuthorizedRequest{
				SourceAccount:      "rsUiUMpnrgxQp24dJYZDhmV4bE3aBtQyt8",
				DestinationAccount: "rEhxGqkqPPSxQ3P25J66ft5TwpzV14k2de",
			},
			expected: path.DepositAuthorizedResponse{
				DepositAuthorized:  true,
				DestinationAccount: "rEhxGqkqPPSxQ3P25J66ft5TwpzV14k2de",
				LedgerCurrentIndex: 14,
				SourceAccount:      "rsUiUMpnrgxQp24dJYZDhmV4bE3aBtQyt8",
				Validated:          false,
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "invalidParams",
					"status": "error"
				}
			}`,
			mockStatus:    200,
			request:       &path.DepositAuthorizedRequest{},
			expectedError: "invalidParams",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var depositResp path.DepositAuthorizedResponse
			err = resp.GetResult(&depositResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, depositResp)
		})
	}
}

func TestClient_FindPathCreate(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *path.FindCreateRequest
		expected      path.FindResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"alternatives": [
						{
							"paths_computed": [
								[
									{
										"currency": "USD",
										"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
										"account": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B"
									}
								]
							],
							"source_amount": "207669",
							"destination_amount": {
								"currency": "USD",
								"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B", 
								"value": "100"
							}
						}
					],
					"destination_account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					"destination_amount": {
						"currency": "USD",
						"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
						"value": "100"
					},
					"source_account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					"full_reply": true,
					"status": true
				}
			}`,
			mockStatus: 200,
			request:    &path.FindCreateRequest{},
			expected: path.FindResponse{
				Alternatives: []pathtypes.Alternative{
					{
						PathsComputed: [][]transaction.PathStep{
							{
								{
									Currency: "USD",
									Issuer:   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
									Account:  "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
								},
							},
						},
						SourceAmount: "207669",
						DestinationAmount: map[string]any{
							"currency": "USD",
							"issuer":   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
							"value":    "100",
						},
					},
				},
				DestinationAccount: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				DestinationAmount: map[string]any{
					"currency": "USD",
					"issuer":   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
					"value":    "100",
				},
				SourceAccount: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				FullReply:     true,
				Status:        true,
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "invalidParams",
					"status": "error"
				}
			}`,
			mockStatus:    200,
			request:       &path.FindCreateRequest{},
			expectedError: "invalidParams",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var findResp path.FindResponse
			err = resp.GetResult(&findResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, findResp)
		})
	}
}

func TestClient_FindPathClose(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *path.FindCloseRequest
		expected      path.FindResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"alternatives": [
						{
							"paths_computed": [
								[
									{
										"currency": "USD",
										"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
										"account": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B"
									}
								]
							],
							"source_amount": {
								"currency": "USD",
								"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
								"value": "100"
							}
						}
					],
					"source_account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					"full_reply": true,
					"status": true
				}
			}`,
			mockStatus: 200,
			request:    &path.FindCloseRequest{},
			expected: path.FindResponse{
				Alternatives: []pathtypes.Alternative{
					{
						PathsComputed: [][]transaction.PathStep{
							{
								{
									Currency: "USD",
									Issuer:   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
									Account:  "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
								},
							},
						},
						SourceAmount: map[string]any{
							"currency": "USD",
							"issuer":   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
							"value":    "100",
						},
					},
				},
				SourceAccount: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				FullReply:     true,
				Status:        true,
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "invalidParams",
					"status": "error"
				}
			}`,
			mockStatus:    200,
			request:       &path.FindCloseRequest{},
			expectedError: "invalidParams",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var findResp path.FindResponse
			err = resp.GetResult(&findResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, findResp)
		})
	}
}

func TestClient_FindPathStatus(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *path.FindStatusRequest
		expected      path.FindResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"alternatives": [
						{
							"paths_computed": [
								[
									{
										"currency": "USD",
										"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
										"account": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B"
									}
								]
							],
							"source_amount": "207669",
							"destination_amount": {
								"currency": "USD",
								"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
								"value": "100"
							}
						}
					],
					"destination_account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					"destination_amount": {
						"currency": "USD",
						"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
						"value": "100"
					},
					"source_account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					"full_reply": true,
					"status": true
				}
			}`,
			mockStatus: 200,
			request:    &path.FindStatusRequest{},
			expected: path.FindResponse{
				Alternatives: []pathtypes.Alternative{
					{
						PathsComputed: [][]transaction.PathStep{
							{
								{
									Currency: "USD",
									Issuer:   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
									Account:  "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
								},
							},
						},
						SourceAmount: "207669",
						DestinationAmount: map[string]any{
							"value":    "100",
							"currency": "USD",
							"issuer":   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
						},
					},
				},
				DestinationAccount: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				DestinationAmount: map[string]any{
					"value":    "100",
					"currency": "USD",
					"issuer":   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
				},
				SourceAccount: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				FullReply:     true,
				Status:        true,
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "invalidParams",
					"status": false
				}
			}`,
			mockStatus:    200,
			request:       &path.FindStatusRequest{},
			expectedError: "invalidParams",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var findResp path.FindResponse
			err = resp.GetResult(&findResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, findResp)
		})
	}
}

func TestClient_GetRipplePathFind(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *path.RipplePathFindRequest
		expected      path.RipplePathFindResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"alternatives": [
						{
							"paths_computed": [
								[
									{
										"currency": "USD",
										"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
										"account": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B"
									}
								]
							],
							"source_amount": {
								"currency": "USD",
								"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
								"value": "100"
							}
						}
					],
					"destination_account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					"destination_currencies": ["USD"],
					"source_account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					"full_reply": true,
					"status": true
				}
			}`,
			mockStatus: 200,
			request:    &path.RipplePathFindRequest{},
			expected: path.RipplePathFindResponse{
				Alternatives: []pathtypes.RippleAlternative{
					{
						PathsComputed: [][]transaction.PathStep{
							{
								{
									Currency: "USD",
									Issuer:   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
									Account:  "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
								},
							},
						},
						SourceAmount: map[string]any{
							"currency": "USD",
							"issuer":   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
							"value":    "100",
						},
					},
				},
				DestinationAccount:    "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				DestinationCurrencies: []string{"USD"},
				SourceAccount:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				FullReply:             true,
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "invalidParams",
					"status": "error"
				}
			}`,
			mockStatus:    200,
			request:       &path.RipplePathFindRequest{},
			expectedError: "invalidParams",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var pathResp path.RipplePathFindResponse
			err = resp.GetResult(&pathResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, pathResp)
		})
	}
}

func TestClient_GetServerInfo(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *server.InfoRequest
		expected      server.InfoResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"info": {
						"build_version": "1.9.4",
						"complete_ledgers": "32570-6595042",
						"hostid": "ARTS",
						"io_latency_ms": 1,
						"last_close": {
							"converge_time_s": 2,
							"proposers": 4
						},
						"load_factor": 1,
						"peers": 53,
						"pubkey_node": "n94RkpbJYRYYGNqKHXdZxNZnWnUCp1xqmXKRqqAqR9n4SB8UmgK",
						"server_state": "full",
						"validated_ledger": {
							"age": 2,
							"base_fee_xrp": 0.00001,
							"hash": "4482DEE5362332F54A4036ED57EE1767C9F33CF7CE5A6670355C16CECE381D46",
							"reserve_base_xrp": 20,
							"reserve_inc_xrp": 5,
							"seq": 6595042
						},
						"validation_quorum": 3
					}
				}
			}`,
			mockStatus: 200,
			request:    &server.InfoRequest{},
			expected: server.InfoResponse{
				Info: servertypes.Info{
					BuildVersion:    "1.9.4",
					CompleteLedgers: "32570-6595042",
					HostID:          "ARTS",
					IOLatencyMS:     1,
					LastClose: servertypes.ServerClose{
						ConvergeTimeS: 2,
						Proposers:     4,
					},
					LoadFactor:  1,
					Peers:       53,
					PubkeyNode:  "n94RkpbJYRYYGNqKHXdZxNZnWnUCp1xqmXKRqqAqR9n4SB8UmgK",
					ServerState: "full",
					ValidatedLedger: servertypes.ClosedLedger{
						Age:            2,
						BaseFeeXRP:     0.00001,
						Hash:           "4482DEE5362332F54A4036ED57EE1767C9F33CF7CE5A6670355C16CECE381D46",
						ReserveBaseXRP: 20,
						ReserveIncXRP:  5,
						Seq:            6595042,
					},
					ValidationQuorum: 3,
				},
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "invalidParams",
					"status": "error"
				}
			}`,
			mockStatus:    200,
			request:       &server.InfoRequest{},
			expectedError: "invalidParams",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var infoResp server.InfoResponse
			err = resp.GetResult(&infoResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, infoResp)
		})
	}
}

func TestClient_GetAllFeatures(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *server.FeatureAllRequest
		expected      server.FeatureAllResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"features": {
						"42": {
							"enabled": true,
							"name": "MultiSign",
							"supported": true
						},
						"43": {
							"enabled": false,
							"name": "TrustSetAuth",
							"supported": true
						}
					}
				}
			}`,
			mockStatus: 200,
			request:    &server.FeatureAllRequest{},
			expected: server.FeatureAllResponse{
				Features: map[string]servertypes.FeatureStatus{
					"42": {
						Enabled:   true,
						Name:      "MultiSign",
						Supported: true,
					},
					"43": {
						Enabled:   false,
						Name:      "TrustSetAuth",
						Supported: true,
					},
				},
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "invalidParams",
					"status": "error"
				}
			}`,
			mockStatus:    200,
			request:       &server.FeatureAllRequest{},
			expectedError: "invalidParams",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var featuresResp server.FeatureAllResponse
			err = resp.GetResult(&featuresResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, featuresResp)
		})
	}
}

func TestClient_GetFeature(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *server.FeatureOneRequest
		expected      server.FeatureResponse
		expectedError string
	}{
		{
			name: "successful feature request",
			mockResponse: `{
				"result": {
					"feature": {
						"enabled": false,
						"name": "TrustSetAuth",
						"supported": true
					}
				}
			}`,
			mockStatus: 200,
			request: &server.FeatureOneRequest{
				Feature: "TrustSetAuth",
			},
			expected: server.FeatureResponse{
				"feature": servertypes.FeatureStatus{
					Enabled:   false,
					Name:      "TrustSetAuth",
					Supported: true,
				},
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "invalidParams",
					"status": "error"
				}
			}`,
			mockStatus:    200,
			request:       &server.FeatureOneRequest{Feature: "InvalidFeature"},
			expectedError: "invalidParams",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var featureResp server.FeatureResponse
			err = resp.GetResult(&featureResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, featureResp)
		})
	}
}

func TestClient_GetFee(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *server.FeeRequest
		expected      server.FeeResponse
		expectedError string
	}{
		{
			name: "successful fee request",
			mockResponse: `{
				"result": {
					"current_ledger_size": "14",
					"current_queue_size": "0",
					"drops": {
						"base_fee": "10",
						"median_fee": "5000",
						"minimum_fee": "10",
						"open_ledger_fee": "10"
					},
					"expected_ledger_size": "24",
					"ledger_current_index": 26575101,
					"levels": {
						"median_level": "128000",
						"minimum_level": "256",
						"open_ledger_level": "256",
						"reference_level": "256"
					},
					"max_queue_size": "480"
				}
			}`,
			mockStatus: 200,
			request:    &server.FeeRequest{},
			expected: server.FeeResponse{
				CurrentLedgerSize:  "14",
				CurrentQueueSize:   "0",
				ExpectedLedgerSize: "24",
				LedgerCurrentIndex: 26575101,
				MaxQueueSize:       "480",
				Drops: servertypes.FeeDrops{
					BaseFee:       types.XRPCurrencyAmount(10),
					MedianFee:     types.XRPCurrencyAmount(5000),
					MinimumFee:    types.XRPCurrencyAmount(10),
					OpenLedgerFee: types.XRPCurrencyAmount(10),
				},
				Levels: servertypes.FeeLevels{
					MedianLevel:     types.XRPCurrencyAmount(128000),
					MinimumLevel:    types.XRPCurrencyAmount(256),
					OpenLedgerLevel: types.XRPCurrencyAmount(256),
					ReferenceLevel:  types.XRPCurrencyAmount(256),
				},
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "invalidParams",
					"status": "error"
				}
			}`,
			mockStatus:    200,
			request:       &server.FeeRequest{},
			expectedError: "invalidParams",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var feeResp server.FeeResponse
			err = resp.GetResult(&feeResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, feeResp)
		})
	}
}

func TestClient_GetManifest(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *server.ManifestRequest
		expected      server.ManifestResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"details": {
						"master_key": "nHUon2tpyJEHHYGmxqeGu37cvPYHzrMtUNQyNinmg4rMdN5q58Bt",
						"seq": 1
					},
					"manifest": "JAAAAAFxIe3AkJgOyqs3y+UuiAI27Ff3Mvn5G2RpVxGLPD2R8iL+dXMhA5SR0Yj7w8e4ko/kY5tFqQWwL1EWGYqC+vP8+0CXtaZwdkYwRAIgCLEe7i+4egt/fQjGcNZmYXHAZCvR1qNBqGLBJvZe9MQCICG2TvyPd293/6h/0UplCukiHwKcW8MJGdqU6+4utUKKcBJA9D1iAMo7YFCpn+V3OYrhaCRZ0jHT7d5iZwusXnKQu+q3yUx6NnVWyQs3QVohN3PVAWmz0DwYZeBo1RqTGhPQCw==",
					"requested": "nHUon2tpyJEHHYGmxqeGu37cvPYHzrMtUNQyNinmg4rMdN5q58Bt"
				}
			}`,
			mockStatus: 200,
			request: &server.ManifestRequest{
				PublicKey: "nHUon2tpyJEHHYGmxqeGu37cvPYHzrMtUNQyNinmg4rMdN5q58Bt",
			},
			expected: server.ManifestResponse{
				Details: server.ManifestDetails{
					MasterKey: "nHUon2tpyJEHHYGmxqeGu37cvPYHzrMtUNQyNinmg4rMdN5q58Bt",
					Seq:       1,
				},
				Manifest:  "JAAAAAFxIe3AkJgOyqs3y+UuiAI27Ff3Mvn5G2RpVxGLPD2R8iL+dXMhA5SR0Yj7w8e4ko/kY5tFqQWwL1EWGYqC+vP8+0CXtaZwdkYwRAIgCLEe7i+4egt/fQjGcNZmYXHAZCvR1qNBqGLBJvZe9MQCICG2TvyPd293/6h/0UplCukiHwKcW8MJGdqU6+4utUKKcBJA9D1iAMo7YFCpn+V3OYrhaCRZ0jHT7d5iZwusXnKQu+q3yUx6NnVWyQs3QVohN3PVAWmz0DwYZeBo1RqTGhPQCw==",
				Requested: "nHUon2tpyJEHHYGmxqeGu37cvPYHzrMtUNQyNinmg4rMdN5q58Bt",
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "invalidParams",
					"status": "error"
				}
			}`,
			mockStatus:    200,
			request:       &server.ManifestRequest{},
			expectedError: "invalidParams",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var manifestResp server.ManifestResponse
			err = resp.GetResult(&manifestResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, manifestResp)
		})
	}
}

func TestClient_GetServerState(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *server.StateRequest
		expected      server.StateResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"state": {
						"build_version": "1.9.4",
						"complete_ledgers": "32570-6595042",
						"io_latency_ms": 1,
						"last_close": {
							"converge_time": 2,
							"proposers": 4
						},
						"load_factor": 1,
						"peers": 21,
						"pubkey_node": "n9KUjqxCr5FKThSNXdzb7oqN8rYwScB2dUnNqxQxbEA17JkaWy5x",
						"server_state": "full",
						"validated_ledger": {
							"base_fee": 1,
							"close_time": 638329241,
							"hash": "4BC50C9B0D8515D3EAAE1E74B29A95804346C491EE1A95BF25E4AAB854A6A652",
							"reserve_base": 20,
							"reserve_inc": 5,
							"seq": 6595042
						},
						"validation_quorum": 4
					}
				}
			}`,
			mockStatus: 200,
			request:    &server.StateRequest{},
			expected: server.StateResponse{
				State: servertypes.State{
					BuildVersion:    "1.9.4",
					CompleteLedgers: "32570-6595042",
					IOLatencyMS:     1,
					LastClose:       servertypes.CloseState{ConvergeTime: 2, Proposers: 4},
					LoadFactor:      1,
					Peers:           21,
					PubkeyNode:      "n9KUjqxCr5FKThSNXdzb7oqN8rYwScB2dUnNqxQxbEA17JkaWy5x",
					ServerState:     "full",
					ValidatedLedger: servertypes.LedgerState{
						BaseFee:     1,
						CloseTime:   638329241,
						Hash:        "4BC50C9B0D8515D3EAAE1E74B29A95804346C491EE1A95BF25E4AAB854A6A652",
						ReserveBase: 20,
						ReserveInc:  5,
						Seq:         6595042,
					},
					ValidationQuorum: 4,
				},
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "invalidParams",
					"status": "error"
				}
			}`,
			mockStatus:    200,
			request:       &server.StateRequest{},
			expectedError: "invalidParams",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var stateResp server.StateResponse
			err = resp.GetResult(&stateResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, stateResp)
		})
	}
}

func TestClient_GetAggregatePrice(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *oracle.GetAggregatePriceRequest
		expected      oracle.GetAggregatePriceResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"median": "123.45",
					"time": 1234567890
				}
			}`,
			mockStatus: 200,
			request:    &oracle.GetAggregatePriceRequest{},
			expected: oracle.GetAggregatePriceResponse{
				Median: "123.45",
				Time:   1234567890,
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "invalidParams",
					"status": "error"
				}
			}`,
			mockStatus:    200,
			request:       &oracle.GetAggregatePriceRequest{},
			expectedError: "invalidParams",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var priceResp oracle.GetAggregatePriceResponse
			err = resp.GetResult(&priceResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, priceResp)
		})
	}
}

func TestClient_Ping(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *utility.PingRequest
		expected      utility.PingResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"role": "admin",
					"unlimited": true
				}
			}`,
			mockStatus: 200,
			request:    &utility.PingRequest{},
			expected: utility.PingResponse{
				Role:      "admin",
				Unlimited: true,
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "invalidParams",
					"status": "error"
				}
			}`,
			mockStatus:    200,
			request:       &utility.PingRequest{},
			expectedError: "invalidParams",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var pingResp utility.PingResponse
			err = resp.GetResult(&pingResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, pingResp)
		})
	}
}

func TestClient_GetRandom(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockStatus    int
		request       *utility.RandomRequest
		expected      utility.RandomResponse
		expectedError string
	}{
		{
			name: "successful response",
			mockResponse: `{
				"result": {
					"random": "8ED765AEBBD6767603C2C9375B2679AEF42BC63BE8B160B10982D767EC309E44"
				}
			}`,
			mockStatus: 200,
			request:    &utility.RandomRequest{},
			expected: utility.RandomResponse{
				Random: "8ED765AEBBD6767603C2C9375B2679AEF42BC63BE8B160B10982D767EC309E44",
			},
		},
		{
			name: "error response",
			mockResponse: `{
				"result": {
					"error": "invalidParams",
					"status": "error"
				}
			}`,
			mockStatus:    200,
			request:       &utility.RandomRequest{},
			expectedError: "invalidParams",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := testutil.JSONRPCMockClient{}
			mc.DoFunc = testutil.MockResponse(tt.mockResponse, tt.mockStatus, &mc)

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(&mc))
			require.NoError(t, err)

			client := NewClient(cfg)

			resp, err := client.Request(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)

			var randomResp utility.RandomResponse
			err = resp.GetResult(&randomResp)
			require.NoError(t, err)

			require.Equal(t, tt.expected, randomResp)
		})
	}
}
