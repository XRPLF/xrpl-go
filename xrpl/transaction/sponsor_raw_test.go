package transaction

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

// Protect raw field presence: conversion must not hide null, mixed, or unknown members.
func TestValidateSponsorFieldsRawSignature(t *testing.T) {
	signer := orderedTransactionSigners(t, 1)[0]
	validSigner := signer.Flatten()
	tests := []struct {
		name  string
		value any
		valid bool
	}{
		{"single", map[string]any{"SigningPubKey": "AB", "TxnSignature": "CD"}, true},
		{"multi decoded", map[string]any{"Signers": []any{validSigner}}, true},
		{"multi flattened", map[string]any{"Signers": []map[string]any{validSigner}}, true},
		{"multi empty key", map[string]any{"SigningPubKey": "", "Signers": []any{validSigner}}, true},
		{"multi nonempty key", map[string]any{"SigningPubKey": "AB", "Signers": []any{validSigner}}, false},
		{"multi null key", map[string]any{"SigningPubKey": nil, "Signers": []any{validSigner}}, false},
		{"multi null signature", map[string]any{"TxnSignature": nil, "Signers": []any{validSigner}}, false},
		{"multi empty signature", map[string]any{"TxnSignature": "", "Signers": []any{validSigner}}, false},
		{"null", nil, false},
		{"typed nil", map[string]any(nil), false},
		{"wrong type", "AB", false},
		{"empty object", map[string]any{}, false},
		{"empty signers", map[string]any{"Signers": []any{}}, false},
		{"null signers", map[string]any{"Signers": nil}, false},
		{"typed nil signers", map[string]any{"Signers": []any(nil)}, false},
		{"null signature", map[string]any{"SigningPubKey": "AB", "TxnSignature": nil}, false},
		{"unknown member", map[string]any{"SigningPubKey": "AB", "TxnSignature": "CD", "Account": "bad"}, false},
		{"wrong wrapper", map[string]any{"Signers": []any{"bad"}}, false},
		{"unknown wrapper member", map[string]any{"Signers": []any{map[string]any{"Signer": validSigner["Signer"], "Extra": true}}}, false},
		{"unknown signer member", map[string]any{"Signers": []any{map[string]any{"Signer": map[string]any{"Account": signer.SignerData.Account.String(), "SigningPubKey": "AB", "TxnSignature": "CD", "Extra": true}}}}, false},
		{"null signer field", map[string]any{"Signers": []any{map[string]any{"Signer": map[string]any{"Account": nil, "SigningPubKey": "AB", "TxnSignature": "CD"}}}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := rawSponsoredPayment()
			tx["SigningPubKey"] = ""
			tx["SponsorSignature"] = tt.value
			before, err := json.Marshal(tx)
			require.NoError(t, err)
			err = ValidateSponsorFields(tx)
			signature, inspectErr := InspectSponsorFields(tx)
			if tt.valid {
				require.NoError(t, err)
				require.NoError(t, inspectErr)
				require.NotNil(t, signature)
			} else {
				require.Error(t, err)
				require.ErrorIs(t, inspectErr, err)
				require.Nil(t, signature, "failed inspection must not publish a partial signature")
			}
			after, err := json.Marshal(tx)
			require.NoError(t, err)
			require.Equal(t, string(before), string(after))
			require.Contains(t, tx, "SigningPubKey")
		})
	}
}

// Inspection retains the typed result without losing field presence or aliasing raw input.
func TestInspectSponsorFields(t *testing.T) {
	key, sig, empty := "AB", "CD", ""
	signers := orderedTransactionSigners(t, 1)
	tests := []struct {
		name      string
		signature *types.SponsorSignature
		inner     bool
	}{
		{"absent", nil, false},
		{"single", &types.SponsorSignature{SigningPubKey: &key, TxnSignature: &sig}, false},
		{"multi absent key", &types.SponsorSignature{Signers: signers}, false},
		{"multi empty key", &types.SponsorSignature{SigningPubKey: &empty, Signers: signers}, false},
		{"inner empty object", &types.SponsorSignature{}, true},
		{"inner empty key", &types.SponsorSignature{SigningPubKey: &empty}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := rawSponsoredPayment()
			if tt.inner {
				tx["Flags"] = types.TfInnerBatchTxn
				tx["SponsorFlags"] = types.SpfSponsorReserve
			}
			if tt.signature != nil {
				tx["SponsorSignature"] = tt.signature.Flatten()
			}
			before, err := json.Marshal(tx)
			require.NoError(t, err)
			got, err := InspectSponsorFields(tx)
			require.NoError(t, err)
			require.Equal(t, tt.signature, got)
			if got != nil {
				if got.SigningPubKey != nil {
					*got.SigningPubKey = "EF"
				}
				if got.TxnSignature != nil {
					*got.TxnSignature = "EF"
				}
				if len(got.Signers) > 0 {
					got.Signers[0].SignerData.TxnSignature = "EF"
				}
			}
			after, err := json.Marshal(tx)
			require.NoError(t, err)
			require.Equal(t, string(before), string(after))
		})
	}
}

func rawSponsoredPayment() FlatTransaction {
	return FlatTransaction{"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "Sponsor": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59", "SponsorFlags": types.SpfSponsorFee}
}

func TestValidateSponsorFieldsNumericFlags(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr error
	}{
		{"int", 1, nil},
		{"uint32", uint32(1), nil},
		{"float64", float64(1), nil},
		{"json number", json.Number("1"), nil},
		{"fraction", 1.5, ErrInvalidSponsorFlags},
		{"negative", -1, ErrInvalidSponsorFlags},
		{"overflow", uint64(math.MaxUint32) + 1, ErrInvalidSponsorFlags},
		{"null", nil, ErrInvalidSponsorFlags},
		{"string", "1", ErrInvalidSponsorFlags},
		{"zero", 0, ErrSponsorFieldsMissing},
		{"unknown bit", 4, ErrInvalidSponsorFlags},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := rawSponsoredPayment()
			tx["SponsorFlags"] = tt.value
			err := ValidateSponsorFields(tx)
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

func TestValidateSponsorFieldsContext(t *testing.T) {
	tests := []struct {
		name, field string
		value       any
		wantErr     error
	}{
		{"bad account", "Account", nil, ErrInvalidAccount},
		{"bad sponsor", "Sponsor", nil, ErrInvalidSponsor},
		{"self", "Sponsor", "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", ErrSponsorAccountConflict},
		{"bad type", "TransactionType", nil, ErrInvalidTransactionType},
		{"bad delegate", "Delegate", nil, ErrInvalidDelegate},
		{"bad flags", "Flags", nil, ErrInvalidFlagsValue},
		{"inner fee", "Flags", types.TfInnerBatchTxn, ErrInnerBatchFeeSponsorship},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := rawSponsoredPayment()
			tx[tt.field] = tt.value
			require.ErrorIs(t, ValidateSponsorFields(tx), tt.wantErr)
		})
	}
	require.NoError(t, ValidateSponsorFields(nil), "this helper validates only sponsorship, not complete transactions")
	tx := rawSponsoredPayment()
	delete(tx, "SponsorFlags")
	require.ErrorIs(t, ValidateSponsorFields(tx), ErrSponsorFieldsMissing)
	tx = rawSponsoredPayment()
	delete(tx, "Sponsor")
	require.ErrorIs(t, ValidateSponsorFields(tx), ErrSponsorFieldsMissing)
}
