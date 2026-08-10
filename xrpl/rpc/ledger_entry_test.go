package rpc

import (
	"io"
	"testing"

	ledgerentry "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	ledgerquery "github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/rpc/testutil"
	"github.com/stretchr/testify/require"
)

func TestClient_GetLedgerEntry(t *testing.T) {
	tests := []struct {
		name            string
		mockResponse    string
		request         *ledgerquery.EntryRequest
		expectedRequest string
		expected        *ledgerquery.EntryResponse
	}{
		{
			name: "json response",
			mockResponse: `{
				"result": {
					"index": "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
					"ledger_hash": "31850E8E48E76D1064651DF39DF4E9542E8C90A9A9B629F4DE339EB3FA74F726",
					"ledger_index": 61966146,
					"node": {
						"Account": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
						"Balance": "424021949",
						"LedgerEntryType": "AccountRoot"
					},
					"deleted_ledger_index": 61966150,
					"validated": true
				}
			}`,
			request: &ledgerquery.EntryRequest{
				AccountRoot:    "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
				IncludeDeleted: true,
			},
			expectedRequest: `{
				"method":"ledger_entry",
				"params":[{
					"api_version":2,
					"account_root":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
					"include_deleted":true
				}]
			}`,
			expected: &ledgerquery.EntryResponse{
				Index:       "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
				LedgerHash:  "31850E8E48E76D1064651DF39DF4E9542E8C90A9A9B629F4DE339EB3FA74F726",
				LedgerIndex: common.LedgerIndex(61966146),
				Node: ledgerentry.FlatLedgerObject{
					"Account":         "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
					"Balance":         "424021949",
					"LedgerEntryType": "AccountRoot",
				},
				DeletedLedgerIndex: common.LedgerIndex(61966150),
				Validated:          true,
			},
		},
		{
			name: "binary response",
			mockResponse: `{
				"result": {
					"index": "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
					"ledger_index": 61966146,
					"node_binary": "1100612200000000",
					"validated": true
				}
			}`,
			request: &ledgerquery.EntryRequest{
				Index:  "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
				Binary: true,
			},
			expectedRequest: `{
				"method":"ledger_entry",
				"params":[{
					"api_version":2,
					"index":"13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
					"binary":true
				}]
			}`,
			expected: &ledgerquery.EntryResponse{
				Index:       "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
				LedgerIndex: common.LedgerIndex(61966146),
				NodeBinary:  "1100612200000000",
				Validated:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := testutil.JSONRPCMockClient{}
			mockClient.DoFunc = testutil.MockResponse(tt.mockResponse, 200, &mockClient)

			config, err := NewClientConfig("http://testnode/", WithHTTPClient(&mockClient))
			require.NoError(t, err)

			response, err := NewClient(config).GetLedgerEntry(tt.request)
			require.NoError(t, err)
			require.Equal(t, tt.expected, response)

			requestBody, readErr := io.ReadAll(mockClient.Spy.Body)
			require.NoError(t, readErr)
			require.JSONEq(t, tt.expectedRequest, string(requestBody))
		})
	}
}
