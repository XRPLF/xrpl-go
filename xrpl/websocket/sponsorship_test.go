package websocket

import (
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

func TestClient_ValidateSponsorship(t *testing.T) {
	coSigned := func(tx transaction.FlatTransaction) transaction.FlatTransaction {
		tx["SponsorSignature"] = map[string]any{
			"SigningPubKey": "ED9434799226374926EDA3B54B1B461B4ABF7237962EAE18528FEA67595397FA32",
			"TxnSignature":  "C3646313B08EED6AF4392261A31B961F10C66CB733DB7F6CD9EAB079857834C8",
		}
		return tx
	}

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
			name:        "pre-funded fee sponsorship within budget",
			tx:          sponsoredPayment(types.SpfSponsorFee),
			message:     sponsorshipMessage(sponsorshipNode(map[string]any{"FeeAmount": "1000000", "MaxFee": "1000"})),
			expectValid: true,
			expectEntry: true,
		},
		{
			name:         "fee budget below the transaction fee is rejected",
			tx:           sponsoredPayment(types.SpfSponsorFee),
			message:      sponsorshipMessage(sponsorshipNode(map[string]any{"FeeAmount": "10"})),
			expectValid:  false,
			expectReason: ErrSponsorshipFeeBudgetExhausted,
			expectEntry:  true,
		},
		{
			name:         "the estimated fee is checked against the MaxFee cap",
			tx:           sponsoredPayment(types.SpfSponsorFee),
			estimatedFee: "5000",
			message:      sponsorshipMessage(sponsorshipNode(map[string]any{"FeeAmount": "1000000", "MaxFee": "1000"})),
			expectValid:  false,
			expectReason: ErrSponsorshipMaxFeeExceeded,
			expectEntry:  true,
		},
		{
			name:        "pre-funded reserve sponsorship with a remaining unit",
			tx:          sponsoredPayment(types.SpfSponsorReserve),
			message:     sponsorshipMessage(sponsorshipNode(map[string]any{"RemainingOwnerCount": uint32(1)})),
			expectValid: true,
			expectEntry: true,
		},
		{
			name:         "pre-funded use needs a signature when the entry requires one",
			tx:           sponsoredPayment(types.SpfSponsorFee),
			message:      sponsorshipMessage(sponsorshipNode(map[string]any{"Flags": uint32(0x00010000), "FeeAmount": "1000000"})),
			expectValid:  false,
			expectReason: ErrSponsorshipFeeSignatureRequired,
			expectEntry:  true,
		},
		{
			name:         "a missing entry without a co-signature is rejected",
			tx:           sponsoredPayment(types.SpfSponsorFee),
			message:      entryNotFoundMessage,
			expectValid:  false,
			expectReason: ErrSponsorshipEntryNotFound,
		},
		{
			name:        "a missing entry with a co-signature is authorized",
			tx:          coSigned(sponsoredPayment(types.SpfSponsorFee)),
			message:     entryNotFoundMessage,
			expectValid: true,
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

// A delegated transaction is sponsored through the delegate, so the lookup must resolve the entry
// against Delegate rather than Account.
func TestClient_ValidateSponsorshipUsesDelegateAsSponsee(t *testing.T) {
	tx := sponsoredPayment(types.SpfSponsorFee)
	tx["Delegate"] = delegateAddress

	client, cleanup := setupTestClient(t, []map[string]any{
		sponsorshipMessage(map[string]any{
			"LedgerEntryType": "Sponsorship",
			"Owner":           sponsorAddress,
			"Sponsee":         delegateAddress,
			"FeeAmount":       "1000000",
		}),
	})
	defer cleanup()

	result, err := client.ValidateSponsorship(tx, "")
	require.NoError(t, err)
	require.True(t, result.Valid)
	require.Equal(t, types.Address(delegateAddress), result.Sponsorship.Sponsee)
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
