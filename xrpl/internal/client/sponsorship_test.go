package client

import (
	"context"
	"errors"
	"maps"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/currency"
	ledgerentry "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	ledgerquery "github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
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

// errSponsorshipEntryAbsent stands in for a transport's entryNotFound error.
var errSponsorshipEntryAbsent = errors.New("entryNotFound")

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

func sponsorSigners() []any {
	return []any{
		map[string]any{"Signer": map[string]any{
			"Account":       delegateAddress,
			"SigningPubKey": "ED9434799226374926EDA3B54B1B461B4ABF7237962EAE18528FEA67595397FA32",
			"TxnSignature":  "C3646313B08EED6AF4392261A31B961F10C66CB733DB7F6CD9EAB079857834C8B0334270A2C037E63CDCCC1932E0832882B7B7066ECD2FAEDEB4A83DF8AE6303",
		}},
	}
}

func withSponsorSignature(signature any) map[string]any {
	tx := sponsoredTx(types.SpfSponsorFee)
	tx["SponsorSignature"] = signature
	return tx
}

// innerBatchWithSponsorSignature is a sponsored inner Batch transaction carrying the given
// SponsorSignature.
func innerBatchWithSponsorSignature(signature any) map[string]any {
	tx := withSponsorSignature(signature)
	tx["SponsorFlags"] = types.SpfSponsorReserve
	tx["Flags"] = types.TfInnerBatchTxn
	return tx
}

// xAddress encodes a classic test address as a mainnet X-address, with a tag when hasTag is set.
func xAddress(t *testing.T, classic string, tag uint32, hasTag bool) string {
	t.Helper()
	address, err := addresscodec.ClassicAddressToXAddress(classic, tag, hasTag, false)
	require.NoError(t, err)
	return address
}

func staticSponsorship(entry *ledgerentry.Sponsorship) RequestResultFunc {
	return func(_ context.Context, _ Request, result any) error {
		if entry == nil {
			return errSponsorshipEntryAbsent
		}
		return DecodeResultInto(map[string]any{"node": entry}, result)
	}
}

func isSponsorshipEntryAbsent(err error) bool {
	return errors.Is(err, errSponsorshipEntryAbsent)
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
			name: "no entry with a multisigned co-signature is authorized",
			tx: func() map[string]any {
				tx := sponsoredTx(types.SpfSponsorFee)
				tx["SponsorSignature"] = map[string]any{"Signers": sponsorSigners()}
				return tx
			}(),
			expectValid: true,
		},
		{
			name: "no entry with the inner Batch placeholder co-signature is authorized",
			tx: func() map[string]any {
				tx := sponsoredTx(types.SpfSponsorReserve)
				tx["Flags"] = types.TfInnerBatchTxn
				tx["SponsorSignature"] = map[string]any{"SigningPubKey": ""}
				return tx
			}(),
			expectValid: true,
		},
		{
			name:        "no entry with an empty inner Batch SponsorSignature is authorized",
			tx:          innerBatchWithSponsorSignature(map[string]any{}),
			expectValid: true,
		},
		{
			name:        "no entry with a multisigned co-signature and an empty SigningPubKey is authorized",
			tx:          withSponsorSignature(map[string]any{"SigningPubKey": "", "Signers": sponsorSigners()}),
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
			name:         "fee and reserve sponsorship still checks the fee budget",
			tx:           sponsoredTx(types.SpfSponsorFee | types.SpfSponsorReserve),
			entry:        sponsorshipEntry(0, xrpAmount(1), nil, ownerCount(0)),
			expectValid:  false,
			expectReason: ErrSponsorshipFeeBudgetExhausted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateSponsorship(t.Context(), staticSponsorship(tt.entry), isSponsorshipEntryAbsent, tt.tx, tt.estimatedFee)
			require.NoError(t, err)
			require.Equal(t, tt.expectValid, result.Valid)
			if tt.expectReason == nil {
				require.NoError(t, result.Reason)
			} else {
				require.ErrorIs(t, result.Reason, tt.expectReason)
			}
			require.Equal(t, tt.entry, result.Sponsorship)

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

// TestValidateSponsorshipInputErrors pins that the sponsorship fields are checked by the rules
// BaseTx.Validate applies, and that no unusable input reaches the ledger.
func TestValidateSponsorshipInputErrors(t *testing.T) {
	with := func(fields map[string]any) map[string]any {
		tx := sponsoredTx(types.SpfSponsorFee)
		for name, value := range fields {
			if value == nil {
				delete(tx, name)
				continue
			}
			tx[name] = value
		}
		return tx
	}

	tests := []struct {
		name        string
		tx          map[string]any
		expectedErr error
	}{
		{
			name:        "transaction without sponsorship fields",
			tx:          map[string]any{"Account": sponseeAddress, "TransactionType": "Payment", "Fee": "100"},
			expectedErr: ErrTransactionNotSponsored,
		},
		{
			name:        "unknown SponsorFlags bit",
			tx:          sponsoredTx(types.SpfSponsorFee | 0x4),
			expectedErr: transaction.ErrInvalidSponsorFlags,
		},
		{
			name:        "sponsor without flags",
			tx:          with(map[string]any{"SponsorFlags": nil}),
			expectedErr: transaction.ErrSponsorFieldsMissing,
		},
		{
			name:        "zero sponsor flags",
			tx:          with(map[string]any{"SponsorFlags": uint32(0)}),
			expectedErr: transaction.ErrSponsorFieldsMissing,
		},
		{
			name:        "empty sponsor",
			tx:          with(map[string]any{"Sponsor": ""}),
			expectedErr: transaction.ErrSponsorFieldsMissing,
		},
		{
			name:        "SponsorSignature without Sponsor and SponsorFlags",
			tx:          with(map[string]any{"Sponsor": nil, "SponsorFlags": nil, "SponsorSignature": map[string]any{"SigningPubKey": "ED94", "TxnSignature": "C364"}}),
			expectedErr: transaction.ErrSponsorFieldsMissing,
		},
		{
			name:        "non-string sponsor",
			tx:          with(map[string]any{"Sponsor": 7}),
			expectedErr: transaction.ErrInvalidSponsor,
		},
		{
			name:        "malformed sponsor address",
			tx:          with(map[string]any{"Sponsor": "not-an-address"}),
			expectedErr: transaction.ErrInvalidSponsor,
		},
		{
			name:        "non-numeric sponsor flags",
			tx:          with(map[string]any{"SponsorFlags": "fee"}),
			expectedErr: transaction.ErrInvalidSponsorFlags,
		},
		{
			name:        "transaction without an account",
			tx:          with(map[string]any{"Account": nil}),
			expectedErr: transaction.ErrInvalidAccount,
		},
		{
			name:        "non-string delegate",
			tx:          with(map[string]any{"Delegate": 7}),
			expectedErr: transaction.ErrInvalidDelegate,
		},
		{
			name:        "malformed delegate address",
			tx:          with(map[string]any{"Delegate": "not-an-address"}),
			expectedErr: ErrInvalidAddress,
		},
		{
			name:        "tagged delegate address",
			tx:          with(map[string]any{"Delegate": xAddress(t, delegateAddress, 7, true)}),
			expectedErr: ErrAccountIDTagNotAllowed,
		},
		{
			name:        "sponsor equal to account",
			tx:          with(map[string]any{"Sponsor": sponseeAddress}),
			expectedErr: transaction.ErrSponsorAccountConflict,
		},
		{
			name:        "sponsor equal to an X-address account",
			tx:          with(map[string]any{"Account": xAddress(t, sponseeAddress, 7, true), "Sponsor": sponseeAddress}),
			expectedErr: transaction.ErrSponsorAccountConflict,
		},
		{
			name:        "X-address sponsor equal to account",
			tx:          with(map[string]any{"Sponsor": xAddress(t, sponseeAddress, 0, false)}),
			expectedErr: transaction.ErrSponsorAccountConflict,
		},
		{
			name:        "reserve sponsorship on a delegated transaction",
			tx:          with(map[string]any{"SponsorFlags": types.SpfSponsorReserve, "Delegate": delegateAddress}),
			expectedErr: transaction.ErrSponsorDelegateConflict,
		},
		{
			name:        "reserve sponsorship on a type outside the allow-list",
			tx:          with(map[string]any{"SponsorFlags": types.SpfSponsorReserve, "TransactionType": "OfferCreate"}),
			expectedErr: transaction.ErrReserveSponsorshipNotAllowed,
		},
		{
			name: "fee sponsorship on an inner Batch transaction",
			tx: func() map[string]any {
				tx := innerBatchWithSponsorSignature(map[string]any{"SigningPubKey": ""})
				tx["SponsorFlags"] = types.SpfSponsorFee
				tx["Fee"] = "0"
				return tx
			}(),
			expectedErr: transaction.ErrInnerBatchFeeSponsorship,
		},
		{
			name:        "transaction without a fee or an estimate",
			tx:          with(map[string]any{"Fee": nil}),
			expectedErr: ErrSponsorshipFeeUnavailable,
		},
		{
			name:        "non-string fee",
			tx:          with(map[string]any{"Fee": 100}),
			expectedErr: ErrSponsorshipFeeIsNotAString,
		},
		{
			name:        "fractional fee",
			tx:          with(map[string]any{"Fee": "10.5"}),
			expectedErr: ErrInvalidSponsorshipFee,
		},
		{
			name:        "SponsorSignature that is not an object",
			tx:          withSponsorSignature("signed"),
			expectedErr: transaction.ErrInvalidSponsorSignature,
		},
		{
			name:        "nil SponsorSignature map",
			tx:          withSponsorSignature(map[string]any(nil)),
			expectedErr: transaction.ErrInvalidSponsorSignature,
		},
		{
			name:        "SponsorSignature with unrelated fields",
			tx:          withSponsorSignature(map[string]any{"x": "y"}),
			expectedErr: transaction.ErrInvalidSponsorSignature,
		},
		{
			name:        "SponsorSignature with SigningPubKey but no TxnSignature",
			tx:          withSponsorSignature(map[string]any{"SigningPubKey": "ED94"}),
			expectedErr: transaction.ErrInvalidSponsorSignature,
		},
		{
			name:        "SponsorSignature with TxnSignature but no SigningPubKey",
			tx:          withSponsorSignature(map[string]any{"TxnSignature": "C364"}),
			expectedErr: transaction.ErrInvalidSponsorSignature,
		},
		{
			name:        "SponsorSignature mixing Signers and TxnSignature",
			tx:          withSponsorSignature(map[string]any{"Signers": sponsorSigners(), "TxnSignature": "C364"}),
			expectedErr: transaction.ErrInvalidSponsorSignature,
		},
		{
			name:        "SponsorSignature mixing Signers and a nonempty SigningPubKey",
			tx:          withSponsorSignature(map[string]any{"SigningPubKey": "ED94", "Signers": sponsorSigners()}),
			expectedErr: transaction.ErrInvalidSponsorSignature,
		},
		{
			name:        "SponsorSignature with empty Signers",
			tx:          withSponsorSignature(map[string]any{"Signers": []any{}}),
			expectedErr: transaction.ErrInvalidSponsorSignature,
		},
		{
			name:        "empty SigningPubKey on an ordinary transaction",
			tx:          withSponsorSignature(map[string]any{"SigningPubKey": ""}),
			expectedErr: transaction.ErrInvalidSponsorSignature,
		},
		{
			name:        "empty SponsorSignature object on an ordinary transaction",
			tx:          withSponsorSignature(map[string]any{}),
			expectedErr: transaction.ErrInvalidSponsorSignature,
		},
		{
			name:        "single-signed SponsorSignature on an inner Batch transaction",
			tx:          innerBatchWithSponsorSignature(map[string]any{"SigningPubKey": "ED94", "TxnSignature": "C364"}),
			expectedErr: transaction.ErrInvalidSponsorSignature,
		},
		{
			name:        "multisigned SponsorSignature on an inner Batch transaction",
			tx:          innerBatchWithSponsorSignature(map[string]any{"SigningPubKey": "", "Signers": sponsorSigners()}),
			expectedErr: transaction.ErrInvalidSponsorSignature,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := func(context.Context, Request, any) error {
				t.Fatal("sponsorship lookup must not run for unusable inputs")
				return nil
			}
			result, err := ValidateSponsorship(t.Context(), request, isSponsorshipEntryAbsent, tt.tx, "")
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
	entry.Sponsee = delegateAddress
	request := func(ctx context.Context, req Request, result any) error {
		entryRequest, ok := req.(*ledgerquery.EntryRequest)
		require.True(t, ok)
		gotSponsor, gotSponsee = entryRequest.Sponsorship.Object.Sponsor, entryRequest.Sponsorship.Object.Sponsee
		return staticSponsorship(entry)(ctx, req, result)
	}
	result, err := ValidateSponsorship(t.Context(), request, isSponsorshipEntryAbsent, tx, "")

	require.NoError(t, err)
	require.True(t, result.Valid)
	require.Equal(t, types.Address(sponsorAddress), gotSponsor)
	require.Equal(t, types.Address(delegateAddress), gotSponsee)
}

// The lookup must use classic sponsor and sponsee addresses without validating unrelated
// fields or modifying any part of the caller's transaction, including Batch inner transactions.
func TestValidateSponsorshipLooksUpClassicAddresses(t *testing.T) {
	account := xAddress(t, sponseeAddress, 7, true)
	sponsor := xAddress(t, sponsorAddress, 0, false)
	tests := []struct {
		name    string
		fields  map[string]any
		sponsee types.Address
	}{
		{
			name:    "tagged account",
			sponsee: sponseeAddress,
		},
		{
			name:    "untagged delegate",
			fields:  map[string]any{"Delegate": xAddress(t, delegateAddress, 0, false)},
			sponsee: delegateAddress,
		},
		{
			name:    "unrelated invalid destination",
			fields:  map[string]any{"Destination": "not-an-address"},
			sponsee: sponseeAddress,
		},
		{
			name:    "JSON-decoded source tag",
			fields:  map[string]any{"SourceTag": float64(7)},
			sponsee: sponseeAddress,
		},
		{
			name:    "conflicting source tag is outside the lookup scope",
			fields:  map[string]any{"SourceTag": uint32(8)},
			sponsee: sponseeAddress,
		},
		{
			name: "Batch inner addresses are untouched",
			fields: map[string]any{
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{"RawTransaction": map[string]any{"Account": account, "Sponsor": sponsor}},
				},
			},
			sponsee: sponseeAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := sponsoredTx(types.SpfSponsorFee)
			tx["Account"] = account
			tx["Sponsor"] = sponsor
			maps.Copy(tx, tt.fields)
			before := CloneTransaction(tx)

			entry := sponsorshipEntry(0, xrpAmount(1000000), nil, nil)
			entry.Sponsee = tt.sponsee
			var gotSponsor, gotSponsee types.Address
			request := func(ctx context.Context, req Request, result any) error {
				entryRequest, ok := req.(*ledgerquery.EntryRequest)
				require.True(t, ok)
				gotSponsor, gotSponsee = entryRequest.Sponsorship.Object.Sponsor, entryRequest.Sponsorship.Object.Sponsee
				return staticSponsorship(entry)(ctx, req, result)
			}
			result, err := ValidateSponsorship(t.Context(), request, isSponsorshipEntryAbsent, tx, "")

			require.NoError(t, err)
			require.True(t, result.Valid)
			require.Equal(t, types.Address(sponsorAddress), gotSponsor)
			require.Equal(t, tt.sponsee, gotSponsee)
			require.Equal(t, before, tx)
		})
	}
}

func TestValidateSponsorshipRejectsMismatchedEntry(t *testing.T) {
	tests := []struct {
		name  string
		entry *ledgerentry.Sponsorship
	}{
		{
			name: "different owner",
			entry: func() *ledgerentry.Sponsorship {
				entry := sponsorshipEntry(0, xrpAmount(1000000), nil, nil)
				entry.Owner = delegateAddress
				return entry
			}(),
		},
		{
			name: "different sponsee",
			entry: func() *ledgerentry.Sponsorship {
				entry := sponsorshipEntry(0, xrpAmount(1000000), nil, nil)
				entry.Sponsee = delegateAddress
				return entry
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateSponsorship(t.Context(), staticSponsorship(tt.entry), isSponsorshipEntryAbsent, sponsoredTx(types.SpfSponsorFee), "")

			require.ErrorIs(t, err, ErrSponsorshipEntryMismatch)
			require.False(t, result.Valid)
			require.Nil(t, result.Sponsorship)
		})
	}
}

func TestValidateSponsorshipPropagatesQueryErrors(t *testing.T) {
	request := func(context.Context, Request, any) error {
		return errSponsorshipQuery
	}
	result, err := ValidateSponsorship(t.Context(), request, isSponsorshipEntryAbsent, sponsoredTx(types.SpfSponsorFee), "")

	require.ErrorIs(t, err, errSponsorshipQuery)
	require.False(t, result.Valid)
	require.Nil(t, result.Sponsorship)
}

// The shared lookup must preserve the caller's context and current-ledger selector.
func TestValidateSponsorshipRequest(t *testing.T) {
	ctx := t.Context()
	var gotContext context.Context
	var gotRequest Request
	entry := sponsorshipEntry(0, xrpAmount(1000000), nil, nil)
	request := func(ctx context.Context, req Request, result any) error {
		gotContext, gotRequest = ctx, req
		return staticSponsorship(entry)(ctx, req, result)
	}

	result, err := ValidateSponsorship(ctx, request, isSponsorshipEntryAbsent, sponsoredTx(types.SpfSponsorFee), "")

	require.NoError(t, err)
	require.True(t, result.Valid)
	require.Equal(t, entry, result.Sponsorship)
	require.Same(t, ctx, gotContext)
	entryRequest, ok := gotRequest.(*ledgerquery.EntryRequest)
	require.True(t, ok)
	require.NoError(t, entryRequest.Validate())
	require.Equal(t, "ledger_entry", entryRequest.Method())
	require.Equal(t, common.Current, entryRequest.LedgerIndex)
	require.Equal(t, &ledgerquery.SponsorshipSelectorFields{
		Sponsor: sponsorAddress,
		Sponsee: sponseeAddress,
	}, entryRequest.Sponsorship.Object)
}

// A successful transport request must still fail when its node cannot be decoded.
func TestFetchSponsorshipEntryDecoding(t *testing.T) {
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
			request := func(_ context.Context, _ Request, result any) error {
				response, ok := result.(*ledgerquery.EntryResponse)
				require.True(t, ok)
				response.Node = tt.node
				return nil
			}
			entry, err := fetchSponsorshipEntry(t.Context(), request, isSponsorshipEntryAbsent, sponsorAddress, sponseeAddress)
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
	result, err := ValidateSponsorship(t.Context(), staticSponsorship(entry), isSponsorshipEntryAbsent, sponsoredTx(types.SpfSponsorFee), "12")

	require.NoError(t, err)
	require.True(t, result.Valid)
	require.Equal(t, entry, result.Sponsorship)
	require.Equal(t, 0, result.Fee.Cmp(currency.DropsFromUint64(12)))
}
