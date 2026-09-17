package client

import (
	"errors"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/currency"
	ledgerentry "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	ledgerquery "github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

const (
	sponsorAddress  = "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH"
	sponseeAddress  = "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf"
	delegateAddress = "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn"
)

// errSponsorshipQuery stands in for a transport or permission failure raised by the client-
// specific ledger_entry lookup.
var errSponsorshipQuery = errors.New("ledger_entry request failed")

func sponsoredTx(sponsorFlags uint32) map[string]any {
	return map[string]any{
		"Account":         sponseeAddress,
		"TransactionType": "Payment",
		"Fee":             "100",
		"Sponsor":         sponsorAddress,
		"SponsorFlags":    sponsorFlags,
	}
}

func sponsorshipEntry(flags uint32, feeAmount, maxFee *types.XRPCurrencyAmount, remaining *uint32) *ledgerentry.Sponsorship {
	return &ledgerentry.Sponsorship{
		LedgerEntryType:     ledgerentry.SponsorshipEntry,
		Flags:               flags,
		Owner:               sponsorAddress,
		Sponsee:             sponseeAddress,
		FeeAmount:           feeAmount,
		MaxFee:              maxFee,
		RemainingOwnerCount: remaining,
	}
}

func xrpAmount(drops uint64) *types.XRPCurrencyAmount {
	amount := types.XRPCurrencyAmount(drops)
	return &amount
}

func ownerCount(count uint32) *uint32 {
	return &count
}

func staticSponsorship(entry *ledgerentry.Sponsorship) FetchSponsorshipEntry {
	return func(_, _ types.Address) (*ledgerentry.Sponsorship, error) {
		return entry, nil
	}
}

func TestValidateSponsorship(t *testing.T) {
	coSignature := map[string]any{
		"SigningPubKey": "ED9434799226374926EDA3B54B1B461B4ABF7237962EAE18528FEA67595397FA32",
		"TxnSignature":  "C3646313B08EED6AF4392261A31B961F10C66CB733DB7F6CD9EAB079857834C8B0334270A2C037E63CDCCC1932E0832882B7B7066ECD2FAEDEB4A83DF8AE6303",
	}

	tests := []struct {
		name         string
		tx           map[string]any
		estimatedFee string
		entry        *ledgerentry.Sponsorship
		expectValid  bool
		expectReason error
	}{
		{
			name:         "no entry and no co-signature is rejected",
			tx:           sponsoredTx(types.SpfSponsorFee),
			expectValid:  false,
			expectReason: ErrSponsorshipEntryNotFound,
		},
		{
			name: "no entry with a co-signature is authorized",
			tx: func() map[string]any {
				tx := sponsoredTx(types.SpfSponsorFee)
				tx["SponsorSignature"] = coSignature
				return tx
			}(),
			expectValid: true,
		},
		{
			name:        "pre-funded fee sponsorship within budget",
			tx:          sponsoredTx(types.SpfSponsorFee),
			entry:       sponsorshipEntry(0, xrpAmount(1000000), xrpAmount(1000), nil),
			expectValid: true,
		},
		{
			name:        "pre-funded fee sponsorship spending the whole budget",
			tx:          sponsoredTx(types.SpfSponsorFee),
			entry:       sponsorshipEntry(0, xrpAmount(100), xrpAmount(100), nil),
			expectValid: true,
		},
		{
			name:         "fee budget below the transaction fee is rejected",
			tx:           sponsoredTx(types.SpfSponsorFee),
			entry:        sponsorshipEntry(0, xrpAmount(99), nil, nil),
			expectValid:  false,
			expectReason: ErrSponsorshipFeeBudgetExhausted,
		},
		{
			name:         "absent fee budget is rejected for fee sponsorship",
			tx:           sponsoredTx(types.SpfSponsorFee),
			entry:        sponsorshipEntry(0, nil, nil, ownerCount(5)),
			expectValid:  false,
			expectReason: ErrSponsorshipFeeAmountMissing,
		},
		{
			name:         "fee above the MaxFee cap is rejected",
			tx:           sponsoredTx(types.SpfSponsorFee),
			entry:        sponsorshipEntry(0, xrpAmount(1000000), xrpAmount(99), nil),
			expectValid:  false,
			expectReason: ErrSponsorshipMaxFeeExceeded,
		},
		{
			name:         "a zero MaxFee cap sponsors no fee",
			tx:           sponsoredTx(types.SpfSponsorFee),
			entry:        sponsorshipEntry(0, xrpAmount(1000000), xrpAmount(0), nil),
			expectValid:  false,
			expectReason: ErrSponsorshipMaxFeeExceeded,
		},
		{
			name:         "the estimated fee overrides the transaction fee",
			tx:           sponsoredTx(types.SpfSponsorFee),
			estimatedFee: "5000",
			entry:        sponsorshipEntry(0, xrpAmount(1000000), xrpAmount(1000), nil),
			expectValid:  false,
			expectReason: ErrSponsorshipMaxFeeExceeded,
		},
		{
			name:         "pre-funded fee sponsorship needs a signature when the entry requires one",
			tx:           sponsoredTx(types.SpfSponsorFee),
			entry:        sponsorshipEntry(ledgerentry.LsfSponsorshipRequireSignForFee, xrpAmount(1000000), nil, nil),
			expectValid:  false,
			expectReason: ErrSponsorshipFeeSignatureRequired,
		},
		{
			name: "a co-signature satisfies the fee signature requirement",
			tx: func() map[string]any {
				tx := sponsoredTx(types.SpfSponsorFee)
				tx["SponsorSignature"] = coSignature
				return tx
			}(),
			entry:       sponsorshipEntry(ledgerentry.LsfSponsorshipRequireSignForFee, xrpAmount(1000000), nil, nil),
			expectValid: true,
		},
		{
			name: "an entry budget still binds a co-signed transaction",
			tx: func() map[string]any {
				tx := sponsoredTx(types.SpfSponsorFee)
				tx["SponsorSignature"] = coSignature
				return tx
			}(),
			entry:        sponsorshipEntry(0, xrpAmount(10), nil, nil),
			expectValid:  false,
			expectReason: ErrSponsorshipFeeBudgetExhausted,
		},
		{
			name:        "pre-funded reserve sponsorship with a remaining unit",
			tx:          sponsoredTx(types.SpfSponsorReserve),
			entry:       sponsorshipEntry(0, nil, nil, ownerCount(1)),
			expectValid: true,
		},
		{
			name:        "reserve sponsorship does not require a remaining owner count",
			tx:          sponsoredTx(types.SpfSponsorReserve),
			entry:       sponsorshipEntry(0, nil, nil, nil),
			expectValid: true,
		},
		{
			name:         "pre-funded reserve sponsorship needs a signature when the entry requires one",
			tx:           sponsoredTx(types.SpfSponsorReserve),
			entry:        sponsorshipEntry(ledgerentry.LsfSponsorshipRequireSignForReserve, nil, nil, ownerCount(5)),
			expectValid:  false,
			expectReason: ErrSponsorshipReserveSignatureRequired,
		},
		{
			name:        "a fee-only signature requirement does not block reserve sponsorship",
			tx:          sponsoredTx(types.SpfSponsorReserve),
			entry:       sponsorshipEntry(ledgerentry.LsfSponsorshipRequireSignForFee, nil, nil, ownerCount(5)),
			expectValid: true,
		},
		{
			name:        "fee and reserve sponsorship within both budgets",
			tx:          sponsoredTx(types.SpfSponsorFee | types.SpfSponsorReserve),
			entry:       sponsorshipEntry(0, xrpAmount(1000000), xrpAmount(1000), ownerCount(2)),
			expectValid: true,
		},
		{
			name: "a zero fee draws nothing from an exhausted fee budget",
			tx: func() map[string]any {
				tx := sponsoredTx(types.SpfSponsorFee)
				tx["Fee"] = "0"
				return tx
			}(),
			entry:       sponsorshipEntry(0, xrpAmount(0), xrpAmount(0), nil),
			expectValid: true,
		},
		{
			name: "a zero fee draws nothing from an absent fee budget",
			tx: func() map[string]any {
				tx := sponsoredTx(types.SpfSponsorFee)
				tx["Fee"] = "0"
				return tx
			}(),
			entry:       sponsorshipEntry(0, nil, nil, nil),
			expectValid: true,
		},
		{
			name: "a nil SponsorSignature is not a co-signature",
			tx: func() map[string]any {
				tx := sponsoredTx(types.SpfSponsorFee)
				tx["SponsorSignature"] = map[string]any(nil)
				return tx
			}(),
			entry:        sponsorshipEntry(ledgerentry.LsfSponsorshipRequireSignForFee, xrpAmount(1000000), nil, nil),
			expectValid:  false,
			expectReason: ErrSponsorshipFeeSignatureRequired,
		},
		{
			name:         "fee and reserve sponsorship still checks the fee budget",
			tx:           sponsoredTx(types.SpfSponsorFee | types.SpfSponsorReserve),
			entry:        sponsorshipEntry(0, xrpAmount(1), nil, ownerCount(0)),
			expectValid:  false,
			expectReason: ErrSponsorshipFeeBudgetExhausted,
		},
		{
			name: "an empty SponsorSignature object is not a co-signature",
			tx: func() map[string]any {
				tx := sponsoredTx(types.SpfSponsorFee)
				tx["SponsorSignature"] = map[string]any{}
				return tx
			}(),
			expectValid:  false,
			expectReason: ErrSponsorshipEntryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateSponsorship(tt.tx, tt.estimatedFee, staticSponsorship(tt.entry))
			require.NoError(t, err)
			require.Equal(t, tt.expectValid, result.Valid)
			if tt.expectReason == nil {
				require.NoError(t, result.Reason)
			} else {
				require.ErrorIs(t, result.Reason, tt.expectReason)
			}
			require.Same(t, tt.entry, result.Sponsorship)

			expectedFee := tt.estimatedFee
			if expectedFee == "" {
				expectedFee, _ = tt.tx["Fee"].(string)
			}
			fee, feeErr := result.Fee.WholeString()
			require.NoError(t, feeErr)
			require.Equal(t, expectedFee, fee)
		})
	}
}

func TestValidateSponsorshipInputErrors(t *testing.T) {
	tests := []struct {
		name        string
		tx          map[string]any
		expectedErr error
	}{
		{
			name:        "unknown SponsorFlags bit",
			tx:          sponsoredTx(types.SpfSponsorFee | 0x4),
			expectedErr: ErrInvalidSponsorFlags,
		},
		{
			name:        "transaction without sponsorship fields",
			tx:          map[string]any{"Account": sponseeAddress, "Fee": "100"},
			expectedErr: ErrTransactionNotSponsored,
		},
		{
			name:        "transaction with a sponsor but no flags",
			tx:          map[string]any{"Account": sponseeAddress, "Fee": "100", "Sponsor": sponsorAddress},
			expectedErr: ErrInvalidSponsorFlags,
		},
		{
			name: "transaction with zero sponsor flags",
			tx: map[string]any{
				"Account": sponseeAddress, "Fee": "100",
				"Sponsor": sponsorAddress, "SponsorFlags": uint32(0),
			},
			expectedErr: ErrInvalidSponsorFlags,
		},
		{
			name: "transaction with an empty sponsor",
			tx: map[string]any{
				"Account": sponseeAddress, "Fee": "100",
				"Sponsor": "", "SponsorFlags": types.SpfSponsorFee,
			},
			expectedErr: ErrTransactionNotSponsored,
		},
		{
			name: "transaction with a non-string sponsor",
			tx: map[string]any{
				"Account": sponseeAddress, "Fee": "100",
				"Sponsor": 7, "SponsorFlags": types.SpfSponsorFee,
			},
			expectedErr: ErrSponsorFieldIsNotAString,
		},
		{
			name: "transaction with non-numeric sponsor flags",
			tx: map[string]any{
				"Account": sponseeAddress, "Fee": "100",
				"Sponsor": sponsorAddress, "SponsorFlags": "fee",
			},
			expectedErr: ErrSponsorFlagsFieldIsNotAUint32,
		},
		{
			name: "transaction without a fee or an estimate",
			tx: map[string]any{
				"Account": sponseeAddress,
				"Sponsor": sponsorAddress, "SponsorFlags": types.SpfSponsorFee,
			},
			expectedErr: ErrSponsorshipFeeUnavailable,
		},
		{
			name: "transaction with a non-string fee",
			tx: map[string]any{
				"Account": sponseeAddress, "Fee": 100,
				"Sponsor": sponsorAddress, "SponsorFlags": types.SpfSponsorFee,
			},
			expectedErr: ErrSponsorshipFeeIsNotAString,
		},
		{
			name: "transaction with a fractional fee",
			tx: map[string]any{
				"Account": sponseeAddress, "Fee": "10.5",
				"Sponsor": sponsorAddress, "SponsorFlags": types.SpfSponsorFee,
			},
			expectedErr: ErrInvalidSponsorshipFee,
		},
		{
			name: "transaction without an account or a delegate",
			tx: map[string]any{
				"Fee":     "100",
				"Sponsor": sponsorAddress, "SponsorFlags": types.SpfSponsorFee,
			},
			expectedErr: ErrSponsorshipSponseeUnavailable,
		},
		{
			name: "transaction with a non-string delegate",
			tx: map[string]any{
				"Account": sponseeAddress, "Delegate": 7, "Fee": "100",
				"Sponsor": sponsorAddress, "SponsorFlags": types.SpfSponsorFee,
			},
			expectedErr: ErrAddressFieldIsNotAString,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateSponsorship(tt.tx, "", func(_, _ types.Address) (*ledgerentry.Sponsorship, error) {
				t.Fatal("sponsorship lookup must not run for unusable inputs")
				return nil, nil
			})
			require.ErrorIs(t, err, tt.expectedErr)
			require.False(t, result.Valid)
		})
	}
}

func TestValidateSponsorshipUsesDelegateAsSponsee(t *testing.T) {
	tx := sponsoredTx(types.SpfSponsorFee)
	tx["Delegate"] = delegateAddress

	var gotSponsor, gotSponsee types.Address
	entry := sponsorshipEntry(0, xrpAmount(1000000), nil, nil)
	result, err := ValidateSponsorship(tx, "", func(sponsor, sponsee types.Address) (*ledgerentry.Sponsorship, error) {
		gotSponsor, gotSponsee = sponsor, sponsee
		return entry, nil
	})

	require.NoError(t, err)
	require.True(t, result.Valid)
	require.Equal(t, types.Address(sponsorAddress), gotSponsor)
	require.Equal(t, types.Address(delegateAddress), gotSponsee)
}

// xrpld rejects reserve sponsorship on a delegated transaction outright, and resolves its reserve
// budget against Account rather than the delegate, so no entry lookup can make the combination
// valid.
func TestValidateSponsorshipRejectsSponsorEqualToAccount(t *testing.T) {
	tx := sponsoredTx(types.SpfSponsorFee)
	tx["Sponsor"] = sponseeAddress

	result, err := ValidateSponsorship(tx, "", func(_, _ types.Address) (*ledgerentry.Sponsorship, error) {
		t.Fatal("ledger_entry must not be queried")
		return nil, nil
	})
	require.NoError(t, err)
	require.False(t, result.Valid)
	require.ErrorIs(t, result.Reason, ErrSponsorIsAccount)
}

func TestValidateSponsorshipRejectsDelegatedReserveSponsorship(t *testing.T) {
	tests := []struct {
		name         string
		sponsorFlags uint32
	}{
		{name: "reserve sponsorship", sponsorFlags: types.SpfSponsorReserve},
		{name: "fee and reserve sponsorship", sponsorFlags: types.SpfSponsorFee | types.SpfSponsorReserve},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := sponsoredTx(tt.sponsorFlags)
			tx["Delegate"] = delegateAddress

			result, err := ValidateSponsorship(tx, "", func(_, _ types.Address) (*ledgerentry.Sponsorship, error) {
				t.Fatal("delegated reserve sponsorship must not reach the ledger")
				return nil, nil
			})

			require.NoError(t, err)
			require.False(t, result.Valid)
			require.ErrorIs(t, result.Reason, ErrDelegatedReserveSponsorship)
			require.Nil(t, result.Sponsorship)
		})
	}
}

func TestValidateSponsorshipPropagatesQueryErrors(t *testing.T) {
	result, err := ValidateSponsorship(sponsoredTx(types.SpfSponsorFee), "", func(_, _ types.Address) (*ledgerentry.Sponsorship, error) {
		return nil, errSponsorshipQuery
	})

	require.ErrorIs(t, err, errSponsorshipQuery)
	require.False(t, result.Valid)
	require.Nil(t, result.Sponsorship)
}

func TestSponsorshipEntryRequest(t *testing.T) {
	request := SponsorshipEntryRequest(sponsorAddress, sponseeAddress)

	require.NoError(t, request.Validate())
	require.Equal(t, "ledger_entry", request.Method())
	require.Equal(t, common.Current, request.LedgerIndex)
	require.Equal(t, &ledgerquery.SponsorshipSelectorFields{
		Sponsor: sponsorAddress,
		Sponsee: sponseeAddress,
	}, request.Sponsorship.Object)
}

func TestDecodeSponsorshipEntry(t *testing.T) {
	tests := []struct {
		name        string
		node        ledgerentry.FlatLedgerObject
		expected    *ledgerentry.Sponsorship
		expectedErr error
	}{
		{
			name: "full entry",
			node: ledgerentry.FlatLedgerObject{
				"LedgerEntryType":     "Sponsorship",
				"Flags":               float64(ledgerentry.LsfSponsorshipRequireSignForFee),
				"Owner":               sponsorAddress,
				"Sponsee":             sponseeAddress,
				"FeeAmount":           "1000000",
				"MaxFee":              "1000",
				"RemainingOwnerCount": float64(5),
				"OwnerNode":           "0000000000000000",
				"SponseeNode":         "0000000000000000",
			},
			expected: &ledgerentry.Sponsorship{
				LedgerEntryType:     ledgerentry.SponsorshipEntry,
				Flags:               ledgerentry.LsfSponsorshipRequireSignForFee,
				Owner:               sponsorAddress,
				Sponsee:             sponseeAddress,
				FeeAmount:           xrpAmount(1000000),
				MaxFee:              xrpAmount(1000),
				RemainingOwnerCount: ownerCount(5),
				OwnerNode:           "0000000000000000",
				SponseeNode:         "0000000000000000",
			},
		},
		{
			name: "reserve-only entry preserves the count without fee budget fields",
			node: ledgerentry.FlatLedgerObject{
				"LedgerEntryType":     "Sponsorship",
				"Owner":               sponsorAddress,
				"Sponsee":             sponseeAddress,
				"RemainingOwnerCount": float64(5),
			},
			expected: &ledgerentry.Sponsorship{
				LedgerEntryType:     ledgerentry.SponsorshipEntry,
				Owner:               sponsorAddress,
				Sponsee:             sponseeAddress,
				RemainingOwnerCount: ownerCount(5),
			},
		},
		{
			name: "entry without optional budget fields",
			node: ledgerentry.FlatLedgerObject{
				"LedgerEntryType": "Sponsorship",
				"Owner":           sponsorAddress,
				"Sponsee":         sponseeAddress,
			},
			expected: &ledgerentry.Sponsorship{
				LedgerEntryType: ledgerentry.SponsorshipEntry,
				Owner:           sponsorAddress,
				Sponsee:         sponseeAddress,
			},
		},
		{
			name:        "another entry type",
			node:        ledgerentry.FlatLedgerObject{"LedgerEntryType": "AccountRoot"},
			expectedErr: ErrSponsorshipEntryUnexpectedType,
		},
		{
			name:        "node without an entry type",
			node:        ledgerentry.FlatLedgerObject{"Owner": sponsorAddress},
			expectedErr: ErrSponsorshipEntryUnexpectedType,
		},
		{
			name: "fee amount with the wrong wire type",
			node: ledgerentry.FlatLedgerObject{
				"LedgerEntryType": "Sponsorship",
				"Owner":           sponsorAddress,
				"Sponsee":         sponseeAddress,
				"FeeAmount":       float64(1000000),
			},
			expectedErr: ErrSponsorshipEntryMalformed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := DecodeSponsorshipEntry(tt.node)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				require.Nil(t, entry)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, entry)
		})
	}
}

func TestValidateSponsorshipReportsTheCheckedFee(t *testing.T) {
	entry := sponsorshipEntry(0, xrpAmount(1000000), nil, nil)
	result, err := ValidateSponsorship(sponsoredTx(types.SpfSponsorFee), "12", staticSponsorship(entry))

	require.NoError(t, err)
	require.True(t, result.Valid)
	require.Same(t, entry, result.Sponsorship)
	require.Equal(t, 0, result.Fee.Cmp(currency.DropsFromUint64(12)))
}
