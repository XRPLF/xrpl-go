package websocket

import (
	"context"
	"maps"
	"testing"

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

func sponsorshipMessage(node map[string]any) map[string]any {
	return map[string]any{
		"id": 1,
		"result": map[string]any{
			"index":                "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
			"ledger_current_index": uint32(61966146),
			"node":                 node,
			"validated":            false,
		},
	}
}

func sponsorshipNode(fields map[string]any) map[string]any {
	node := map[string]any{
		"LedgerEntryType": "Sponsorship",
		"Owner":           sponsorAddress,
		"Sponsee":         sponseeAddress,
	}
	maps.Copy(node, fields)
	return node
}

var entryNotFoundMessage = map[string]any{
	"id":     1,
	"status": "error",
	"type":   "response",
	"error":  "entryNotFound",
}

// TestClient_ValidateSponsorship covers what the WebSocket wrapper adds over the shared preflight,
// whose sponsorship rules are tested in xrpl/internal/client: decoding a found entry, forwarding the
// estimated fee, and reading entryNotFound as an absent entry.
func TestClient_ValidateSponsorship(t *testing.T) {
	tests := []struct {
		name         string
		tx           transaction.FlatTransaction
		estimatedFee string
		message      map[string]any
		expectValid  bool
		expectReason error
		expectEntry  bool
	}{
		{
			name:        "decodes a found entry into the result",
			tx:          sponsoredPayment(types.SpfSponsorFee),
			message:     sponsorshipMessage(sponsorshipNode(map[string]any{"FeeAmount": "1000000", "MaxFee": "1000"})),
			expectValid: true,
			expectEntry: true,
		},
		{
			name:         "forwards the estimated fee",
			tx:           sponsoredPayment(types.SpfSponsorFee),
			estimatedFee: "5000",
			message:      sponsorshipMessage(sponsorshipNode(map[string]any{"FeeAmount": "1000000", "MaxFee": "1000"})),
			expectValid:  false,
			expectReason: ErrSponsorshipMaxFeeExceeded,
			expectEntry:  true,
		},
		{
			name:         "reads entryNotFound as an absent entry",
			tx:           sponsoredPayment(types.SpfSponsorFee),
			message:      entryNotFoundMessage,
			expectValid:  false,
			expectReason: ErrSponsorshipEntryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := setupTestClient(t, []map[string]any{tt.message})
			defer cleanup()

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

// TestClient_ValidateSponsorshipRequest checks the outgoing ledger_entry request: the sponsee is the
// Delegate when present and the Account otherwise, and the current ledger is selected.
func TestClient_ValidateSponsorshipRequest(t *testing.T) {
	tests := []struct {
		name    string
		tx      transaction.FlatTransaction
		sponsee string
	}{
		{
			name:    "looks the entry up against the transaction account",
			tx:      sponsoredPayment(types.SpfSponsorFee),
			sponsee: sponseeAddress,
		},
		{
			name: "looks the entry up against the delegate",
			tx: func() transaction.FlatTransaction {
				tx := sponsoredPayment(types.SpfSponsorFee)
				tx["Delegate"] = delegateAddress
				return tx
			}(),
			sponsee: delegateAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, request := setupQueryTestClient(t, map[string]any{
				"node": sponsorshipNode(map[string]any{"Sponsee": tt.sponsee, "FeeAmount": "1000000"}),
			})

			result, err := client.ValidateSponsorship(tt.tx, "")
			require.NoError(t, err)
			require.True(t, result.Valid)

			got := request()
			require.Equal(t, "ledger_entry", got["command"])
			require.Equal(t, map[string]any{"sponsor": sponsorAddress, "sponsee": tt.sponsee}, got["sponsorship"])
			require.Equal(t, "current", got["ledger_index"])
		})
	}
}

func TestClient_ValidateSponsorshipContextCancellation(t *testing.T) {
	client, cleanup := setupTestClient(t, []map[string]any{})
	defer cleanup()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := client.ValidateSponsorshipContext(ctx, sponsoredPayment(types.SpfSponsorFee), "")

	require.ErrorIs(t, err, context.Canceled)
	require.False(t, result.Valid)
	require.Nil(t, result.Sponsorship)
}

// A ledger_entry failure other than entryNotFound must never be read as an absent sponsorship.
func TestClient_ValidateSponsorshipPropagatesQueryErrors(t *testing.T) {
	client, cleanup := setupTestClient(t, []map[string]any{{
		"id":     1,
		"status": "error",
		"type":   "response",
		"error":  "noPermission",
	}})
	defer cleanup()

	result, err := client.ValidateSponsorship(sponsoredPayment(types.SpfSponsorFee), "")

	var responseErr *ErrorWebsocketClientXrplResponse
	require.ErrorAs(t, err, &responseErr)
	require.Equal(t, "noPermission", responseErr.Type)
	require.False(t, result.Valid)
	require.Nil(t, result.Sponsorship)
}

func TestClient_ValidateSponsorshipRejectsUndecodableEntries(t *testing.T) {
	tests := []struct {
		name        string
		message     map[string]any
		expectedErr error
	}{
		{
			name:        "another entry type",
			message:     sponsorshipMessage(map[string]any{"LedgerEntryType": "AccountRoot", "Account": sponseeAddress}),
			expectedErr: ErrSponsorshipEntryUnexpectedType,
		},
		{
			name:        "a fee amount with the wrong wire type",
			message:     sponsorshipMessage(sponsorshipNode(map[string]any{"FeeAmount": uint32(1000000)})),
			expectedErr: ErrSponsorshipEntryMalformed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := setupTestClient(t, []map[string]any{tt.message})
			defer cleanup()

			result, err := client.ValidateSponsorship(sponsoredPayment(types.SpfSponsorFee), "")
			require.ErrorIs(t, err, tt.expectedErr)
			require.False(t, result.Valid)
			require.Nil(t, result.Sponsorship)
		})
	}
}

// An input the shared preflight rejects must not reach the network.
func TestClient_ValidateSponsorshipRejectsUnsponsoredTransactions(t *testing.T) {
	client, cleanup := setupTestClient(t, []map[string]any{})
	defer cleanup()

	result, err := client.ValidateSponsorship(transaction.FlatTransaction{
		"Account":         sponseeAddress,
		"TransactionType": "Payment",
		"Fee":             "100",
	}, "")

	require.ErrorIs(t, err, ErrTransactionNotSponsored)
	require.False(t, result.Valid)
}
