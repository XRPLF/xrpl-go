package rpc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/rpc/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

const (
	sponsorAddress  = "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH"
	sponseeAddress  = "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf"
	delegateAddress = "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW"
)

func sponsoredPayment(sponsorFlags uint32) transaction.FlatTransaction {
	return transaction.FlatTransaction{
		"Account":         sponseeAddress,
		"TransactionType": "Payment",
		"Fee":             "100",
		"Sponsor":         sponsorAddress,
		"SponsorFlags":    sponsorFlags,
	}
}

func sponsorshipResponse(node string) string {
	return `{"result":{"index":"13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8","ledger_current_index":61966146,"node":` + node + `,"validated":false}}`
}

func newSponsorshipTestClient(t *testing.T, mockResponse string) (*Client, *testutil.JSONRPCMockClient) {
	t.Helper()

	mockClient := &testutil.JSONRPCMockClient{}
	mockClient.DoFunc = testutil.MockResponse(mockResponse, 200, mockClient)

	config, err := NewClientConfig("http://testnode/", WithHTTPClient(mockClient))
	require.NoError(t, err)

	return NewClient(config), mockClient
}

// TestClient_ValidateSponsorship covers what the RPC wrapper adds over the shared preflight, whose
// sponsorship rules are tested in xrpl/internal/client: decoding a found entry, forwarding the
// estimated fee, and reading entryNotFound as an absent entry.
func TestClient_ValidateSponsorship(t *testing.T) {
	tests := []struct {
		name         string
		tx           transaction.FlatTransaction
		estimatedFee string
		mockResponse string
		expectValid  bool
		expectReason error
		expectEntry  bool
	}{
		{
			name:         "decodes a found entry into the result",
			tx:           sponsoredPayment(types.SpfSponsorFee),
			mockResponse: sponsorshipResponse(`{"LedgerEntryType":"Sponsorship","Owner":"` + sponsorAddress + `","Sponsee":"` + sponseeAddress + `","FeeAmount":"1000000","MaxFee":"1000"}`),
			expectValid:  true,
			expectEntry:  true,
		},
		{
			name:         "forwards the estimated fee",
			tx:           sponsoredPayment(types.SpfSponsorFee),
			estimatedFee: "5000",
			mockResponse: sponsorshipResponse(`{"LedgerEntryType":"Sponsorship","Owner":"` + sponsorAddress + `","Sponsee":"` + sponseeAddress + `","FeeAmount":"1000000","MaxFee":"1000"}`),
			expectValid:  false,
			expectReason: ErrSponsorshipMaxFeeExceeded,
			expectEntry:  true,
		},
		{
			name:         "reads entryNotFound as an absent entry",
			tx:           sponsoredPayment(types.SpfSponsorFee),
			mockResponse: `{"result":{"error":"entryNotFound","status":"error"}}`,
			expectValid:  false,
			expectReason: ErrSponsorshipEntryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := newSponsorshipTestClient(t, tt.mockResponse)

			result, err := client.ValidateSponsorship(tt.tx, tt.estimatedFee)
			require.NoError(t, err)
			require.Equal(t, tt.expectValid, result.Valid)
			if tt.expectReason == nil {
				require.NoError(t, result.Reason)
			} else {
				require.ErrorIs(t, result.Reason, tt.expectReason)
			}
			if tt.expectEntry {
				require.NotNil(t, result.Sponsorship)
			} else {
				require.Nil(t, result.Sponsorship)
			}
		})
	}
}

// A ledger_entry failure other than entryNotFound must never be read as an absent sponsorship.
func TestClient_ValidateSponsorshipPropagatesQueryErrors(t *testing.T) {
	client, _ := newSponsorshipTestClient(t, `{"result":{"error":"noPermission","status":"error"}}`)

	result, err := client.ValidateSponsorship(sponsoredPayment(types.SpfSponsorFee), "")

	var clientErr *ClientError
	require.ErrorAs(t, err, &clientErr)
	require.Equal(t, "noPermission", clientErr.ErrorString)
	require.False(t, result.Valid)
	require.Nil(t, result.Sponsorship)
}

func TestClient_ValidateSponsorshipRejectsUndecodableEntries(t *testing.T) {
	tests := []struct {
		name         string
		mockResponse string
		expectedErr  error
	}{
		{
			name:         "another entry type",
			mockResponse: sponsorshipResponse(`{"LedgerEntryType":"AccountRoot","Account":"` + sponseeAddress + `"}`),
			expectedErr:  ErrSponsorshipEntryUnexpectedType,
		},
		{
			name:         "a fee amount with the wrong wire type",
			mockResponse: sponsorshipResponse(`{"LedgerEntryType":"Sponsorship","Owner":"` + sponsorAddress + `","FeeAmount":1000000}`),
			expectedErr:  ErrSponsorshipEntryMalformed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := newSponsorshipTestClient(t, tt.mockResponse)

			result, err := client.ValidateSponsorship(sponsoredPayment(types.SpfSponsorFee), "")
			require.ErrorIs(t, err, tt.expectedErr)
			require.False(t, result.Valid)
			require.Nil(t, result.Sponsorship)
		})
	}
}

func TestClient_ValidateSponsorshipRequest(t *testing.T) {
	tests := []struct {
		name            string
		tx              transaction.FlatTransaction
		sponsee         string
		expectedRequest string
	}{
		{
			name:    "looks the entry up against the transaction account",
			tx:      sponsoredPayment(types.SpfSponsorFee),
			sponsee: sponseeAddress,
			expectedRequest: `{"method":"ledger_entry","params":[{
				"api_version":2,
				"sponsorship":{"sponsor":"` + sponsorAddress + `","sponsee":"` + sponseeAddress + `"},
				"ledger_index":"current"
			}]}`,
		},
		{
			name: "looks the entry up against the delegate",
			tx: func() transaction.FlatTransaction {
				tx := sponsoredPayment(types.SpfSponsorFee)
				tx["Delegate"] = delegateAddress
				return tx
			}(),
			sponsee: delegateAddress,
			expectedRequest: `{"method":"ledger_entry","params":[{
				"api_version":2,
				"sponsorship":{"sponsor":"` + sponsorAddress + `","sponsee":"` + delegateAddress + `"},
				"ledger_index":"current"
			}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mockClient := newSponsorshipTestClient(t, sponsorshipResponse(
				`{"LedgerEntryType":"Sponsorship","Owner":"`+sponsorAddress+`","Sponsee":"`+tt.sponsee+`","FeeAmount":"1000000"}`,
			))

			result, err := client.ValidateSponsorship(tt.tx, "")
			require.NoError(t, err)
			require.True(t, result.Valid)

			requestBody, readErr := io.ReadAll(mockClient.Spy.Body)
			require.NoError(t, readErr)
			require.JSONEq(t, tt.expectedRequest, string(requestBody))
		})
	}
}

func TestClient_ValidateSponsorshipContextCancellation(t *testing.T) {
	client, _ := newSponsorshipTestClient(t, `{"result":{}}`)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := client.ValidateSponsorshipContext(ctx, sponsoredPayment(types.SpfSponsorFee), "")

	require.ErrorIs(t, err, context.Canceled)
	require.False(t, result.Valid)
	require.Nil(t, result.Sponsorship)
}

// An input the shared preflight rejects must not reach the network.
func TestClient_ValidateSponsorshipRejectsUnsponsoredTransactions(t *testing.T) {
	client, mockClient := newSponsorshipTestClient(t, `{"result":{}}`)

	result, err := client.ValidateSponsorship(transaction.FlatTransaction{
		"Account":         sponseeAddress,
		"TransactionType": "Payment",
		"Fee":             "100",
	}, "")

	require.ErrorIs(t, err, clientinternal.ErrTransactionNotSponsored)
	require.False(t, result.Valid)
	require.Nil(t, mockClient.Spy)
}

// Sponsorship validation is opt-in, so preparing a sponsored transaction must not query the ledger
// for its Sponsorship entry.
func TestClient_AutofillDoesNotQuerySponsorship(t *testing.T) {
	mockClient := &testutil.JSONRPCMockClient{}
	var methods []string
	mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		var decoded struct {
			Method string `json:"method"`
		}
		require.NoError(t, json.Unmarshal(body, &decoded))
		methods = append(methods, decoded.Method)
		return testutil.MockResponse(`{"result":{}}`, 200, mockClient)(req)
	}

	config, err := NewClientConfig("http://testnode/", WithHTTPClient(mockClient), WithNetworkIdentity(0, "1.12.0"))
	require.NoError(t, err)

	// LastLedgerSequence is left out so autofill still reaches the network, which proves the
	// recorder observes the requests autofill does make.
	tx := sponsoredPayment(types.SpfSponsorFee)
	tx["Sequence"] = uint32(42)
	tx["Flags"] = uint32(0)

	require.NoError(t, NewClient(config).Autofill(&tx))
	require.NotEmpty(t, methods)
	require.NotContains(t, methods, "ledger_entry")
}
