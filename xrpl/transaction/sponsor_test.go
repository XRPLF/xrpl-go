package transaction

import (
	"encoding/json"
	"maps"
	"testing"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

// TestSponsorCodecRoundTrip checks the model-to-codec boundary, not signature
// validity or independent correctness of the protocol definitions.
func TestSponsorCodecRoundTrip(t *testing.T) {
	key, signature, empty := "AB", "CD", ""
	signers := []types.Signer{{SignerData: types.SignerData{
		Account:       "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		SigningPubKey: key,
		TxnSignature:  signature,
	}}}
	tests := []struct {
		name      string
		signature *types.SponsorSignature
		flags     uint32
	}{
		{"prefunded", nil, 0},
		{"single", &types.SponsorSignature{SigningPubKey: &key, TxnSignature: &signature}, 0},
		{"multi", &types.SponsorSignature{Signers: signers}, 0},
		{"multi with empty key", &types.SponsorSignature{SigningPubKey: &empty, Signers: signers}, 0},
		{"inner empty object", &types.SponsorSignature{}, types.TfInnerBatchTxn},
		{"inner empty key", &types.SponsorSignature{SigningPubKey: &empty}, types.TfInnerBatchTxn},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := BaseTx{
				Account:          "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				TransactionType:  PaymentTx,
				Flags:            tt.flags,
				Sponsor:          "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				SponsorFlags:     types.SpfSponsorReserve,
				SponsorSignature: tt.signature,
			}
			flattened := tx.Flatten()
			encoded, err := binarycodec.Encode(flattened)
			require.NoError(t, err)
			require.NotEmpty(t, encoded)
			decoded, err := binarycodec.Decode(encoded)
			require.NoError(t, err)
			expectedJSON, err := json.Marshal(flattened)
			require.NoError(t, err)
			decodedJSON, err := json.Marshal(decoded)
			require.NoError(t, err)
			require.JSONEq(t, string(expectedJSON), string(decodedJSON))
		})
	}
}

func TestBaseTxSponsorValidation(t *testing.T) {
	const account = "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	const sponsor = "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59"
	tests := []struct {
		name        string
		sponsor     types.Address
		flags       uint32
		txType      TxType
		delegate    types.Address
		inner       bool
		expectedErr error
	}{
		{"absent", "", 0, PaymentTx, "", false, nil},
		{"fee", sponsor, types.SpfSponsorFee, OfferCreateTx, "", false, nil},
		{"reserve", sponsor, types.SpfSponsorReserve, PaymentTx, "", false, nil},
		{"both", sponsor, types.SpfSponsorFee | types.SpfSponsorReserve, PaymentTx, "", false, nil},
		{"missing flags", sponsor, 0, PaymentTx, "", false, ErrSponsorFieldsMissing},
		{"missing sponsor", "", types.SpfSponsorFee, PaymentTx, "", false, ErrSponsorFieldsMissing},
		{"unknown flags", sponsor, 4, PaymentTx, "", false, ErrInvalidSponsorFlags},
		{"high bit", sponsor, 0x80000000, PaymentTx, "", false, ErrInvalidSponsorFlags},
		{"invalid address", "invalid", types.SpfSponsorFee, PaymentTx, "", false, ErrInvalidSponsor},
		{"zero address", "rrrrrrrrrrrrrrrrrrrrrhoLvTp", types.SpfSponsorFee, PaymentTx, "", false, ErrSponsorZero},
		{"self", account, types.SpfSponsorFee, PaymentTx, "", false, ErrSponsorAccountConflict},
		{"untagged X address", "X7AcgcsBL6XDcUb289X4mJ8djcdyKaB5hJDWMArnXr61cqZ", types.SpfSponsorFee, PaymentTx, "", false, nil},
		{"tagged X address", "X7AcgcsBL6XDcUb289X4mJ8djcdyKaGZMhc9YTE92ehJ2Fu", types.SpfSponsorFee, PaymentTx, "", false, ErrSponsorTagNotAllowed},
		{"reserve not allowed", sponsor, types.SpfSponsorReserve, OfferCreateTx, "", false, ErrReserveSponsorshipNotAllowed},
		{"reserve with delegate", sponsor, types.SpfSponsorReserve, PaymentTx, sponsor, false, ErrSponsorDelegateConflict},
		{"fee with delegate", sponsor, types.SpfSponsorFee, PaymentTx, sponsor, false, nil},
		{"inner reserve", sponsor, types.SpfSponsorReserve, PaymentTx, "", true, nil},
		{"inner fee", sponsor, types.SpfSponsorFee, PaymentTx, "", true, ErrInnerBatchFeeSponsorship},
		{"inner both", sponsor, 3, PaymentTx, "", true, ErrInnerBatchFeeSponsorship},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := BaseTx{
				Account: account, TransactionType: tt.txType, Sponsor: tt.sponsor, SponsorFlags: tt.flags, Delegate: tt.delegate,
			}
			if tt.inner {
				tx.Flags = types.TfInnerBatchTxn
			}

			// The raw entry point shares the typed sponsorship rules, without
			// requiring Fee or Sequence to have been autofilled.
			_, rawErr := InspectSponsorFields(tx.Flatten())
			if tt.expectedErr == nil {
				require.NoError(t, rawErr)
			} else {
				require.ErrorIs(t, rawErr, tt.expectedErr)
			}

			valid, err := tx.Validate()
			require.Equal(t, tt.expectedErr == nil, valid)
			if tt.expectedErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.expectedErr)
			}
		})
	}
}

func TestSponsorSignatureValidation(t *testing.T) {
	key, signature, empty := "AB", "CD", ""
	signers := []types.Signer{{SignerData: types.SignerData{
		Account: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", SigningPubKey: key, TxnSignature: signature,
	}}}
	tests := []struct {
		name      string
		signature *types.SponsorSignature
		inner     bool
		valid     bool
	}{
		{"absent", nil, false, true},
		{"inner absent", nil, true, true},
		{"single", &types.SponsorSignature{SigningPubKey: &key, TxnSignature: &signature}, false, true},
		{"multi", &types.SponsorSignature{Signers: signers}, false, true},
		{"empty object", &types.SponsorSignature{}, false, false},
		{"empty array", &types.SponsorSignature{Signers: []types.Signer{}}, false, false},
		{"key only", &types.SponsorSignature{SigningPubKey: &key}, false, false},
		{"signature only", &types.SponsorSignature{TxnSignature: &signature}, false, false},
		{"empty key", &types.SponsorSignature{SigningPubKey: &empty, TxnSignature: &signature}, false, false},
		{"empty signature", &types.SponsorSignature{SigningPubKey: &key, TxnSignature: &empty}, false, false},
		{"both methods", &types.SponsorSignature{SigningPubKey: &key, TxnSignature: &signature, Signers: signers}, false, false},
		{"empty key with multi", &types.SponsorSignature{SigningPubKey: &empty, Signers: signers}, false, true},
		{"nonempty key with multi", &types.SponsorSignature{SigningPubKey: &key, Signers: signers}, false, false},
		{"empty signature with multi", &types.SponsorSignature{TxnSignature: &empty, Signers: signers}, false, false},
		{"partial signature with multi", &types.SponsorSignature{TxnSignature: &signature, Signers: signers}, false, false},
		{"invalid signer", &types.SponsorSignature{Signers: []types.Signer{{}}}, false, false},
		{"inner empty key", &types.SponsorSignature{SigningPubKey: &empty}, true, true},
		{"inner empty object", &types.SponsorSignature{}, true, true},
		{"inner nonempty key", &types.SponsorSignature{SigningPubKey: &key}, true, false},
		{"inner empty signature present", &types.SponsorSignature{SigningPubKey: &empty, TxnSignature: &empty}, true, false},
		{"inner empty signature without key", &types.SponsorSignature{TxnSignature: &empty}, true, false},
		{"inner empty signers present", &types.SponsorSignature{SigningPubKey: &empty, Signers: []types.Signer{}}, true, false},
		{"inner empty signers without key", &types.SponsorSignature{Signers: []types.Signer{}}, true, false},
		{"inner signed", &types.SponsorSignature{SigningPubKey: &key, TxnSignature: &signature}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := BaseTx{
				Account: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", TransactionType: PaymentTx,
				Sponsor: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59", SponsorFlags: types.SpfSponsorReserve, SponsorSignature: tt.signature,
			}
			if tt.inner {
				tx.Flags = types.TfInnerBatchTxn
			}
			valid, err := tx.Validate()
			require.Equal(t, tt.valid, valid)
			if tt.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestSponsorSelfComparisonUsesAccountID(t *testing.T) {
	tx := BaseTx{
		Account: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59", TransactionType: PaymentTx,
		Sponsor: "X7AcgcsBL6XDcUb289X4mJ8djcdyKaB5hJDWMArnXr61cqZ", SponsorFlags: types.SpfSponsorFee,
	}
	valid, err := tx.Validate()
	require.False(t, valid)
	require.ErrorIs(t, err, ErrSponsorAccountConflict)
}

func TestBaseTxSponsorSerialization(t *testing.T) {
	key, signature, empty := "AB", "CD", ""
	tests := []struct {
		name      string
		signature *types.SponsorSignature
		expected  string
	}{
		{"multi with empty key", &types.SponsorSignature{SigningPubKey: &empty, Signers: []types.Signer{{SignerData: types.SignerData{Account: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", SigningPubKey: key, TxnSignature: signature}}}}, `{"SigningPubKey":"","Signers":[{"Signer":{"Account":"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh","SigningPubKey":"AB","TxnSignature":"CD"}}]}`},
		{"unsigned", nil, `null`},
		{"single", &types.SponsorSignature{SigningPubKey: &key, TxnSignature: &signature}, `{"SigningPubKey":"AB","TxnSignature":"CD"}`},
		{"inner empty object", &types.SponsorSignature{}, `{}`},
		{"inner empty key", &types.SponsorSignature{SigningPubKey: &empty}, `{"SigningPubKey":""}`},
		{"multi", &types.SponsorSignature{Signers: []types.Signer{{SignerData: types.SignerData{Account: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", SigningPubKey: key, TxnSignature: signature}}}}, `{"Signers":[{"Signer":{"Account":"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh","SigningPubKey":"AB","TxnSignature":"CD"}}]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := BaseTx{Sponsor: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59", SponsorFlags: 3, SponsorSignature: tt.signature}
			encoded, err := json.Marshal(tx)
			require.NoError(t, err)
			var fields map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(encoded, &fields))
			flat := tx.Flatten()
			require.Equal(t, tx.Sponsor.String(), flat["Sponsor"])
			require.Equal(t, uint32(3), flat["SponsorFlags"])
			if tt.signature == nil {
				require.NotContains(t, fields, "SponsorSignature")
				require.NotContains(t, flat, "SponsorSignature")
			} else {
				require.JSONEq(t, tt.expected, string(fields["SponsorSignature"]))
				flatJSON, err := json.Marshal(flat["SponsorSignature"])
				require.NoError(t, err)
				require.JSONEq(t, tt.expected, string(flatJSON))
			}
			var decoded BaseTx
			require.NoError(t, json.Unmarshal(encoded, &decoded))
			require.Equal(t, tx, decoded)
		})
	}
	var tx BaseTx
	encoded, err := json.Marshal(tx)
	require.NoError(t, err)
	var fields map[string]any
	require.NoError(t, json.Unmarshal(encoded, &fields))
	for _, field := range []string{"Sponsor", "SponsorFlags", "SponsorSignature"} {
		require.NotContains(t, fields, field)
		require.NotContains(t, tx.Flatten(), field)
	}
}

func TestSponsorSignaturePlaceholderEncoding(t *testing.T) {
	empty := ""
	tests := []struct {
		name      string
		signature *types.SponsorSignature
		expected  string
	}{
		{"absent", nil, ""},
		{"empty object", &types.SponsorSignature{}, "E026E1"},
		{"empty key", &types.SponsorSignature{SigningPubKey: &empty}, "E0267300E1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := BaseTx{SponsorSignature: tt.signature}
			encoded, err := binarycodec.Encode(tx.Flatten())
			require.NoError(t, err)
			require.Equal(t, tt.expected, encoded)
		})
	}
}

func TestBatchValidatesInnerSponsorship(t *testing.T) {
	tests := []struct {
		name        string
		fields      map[string]any
		expectedErr error
	}{
		{
			name:        "unsponsored",
			fields:      map[string]any{},
			expectedErr: nil,
		},
		{
			name: "EnableAmendment fee sponsorship",
			fields: map[string]any{
				"TransactionType": EnableAmendmentTx.String(),
				"Account":         "rrrrrrrrrrrrrrrrrrrrrhoLvTp",
				"Sponsor":         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":    types.SpfSponsorFee,
			},
			expectedErr: ErrPseudoTransactionSponsorship,
		},
		{
			name: "EnableAmendment reserve sponsorship",
			fields: map[string]any{
				"TransactionType": EnableAmendmentTx.String(),
				"Account":         "rrrrrrrrrrrrrrrrrrrrrhoLvTp",
				"Sponsor":         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":    types.SpfSponsorReserve,
			},
			expectedErr: ErrPseudoTransactionSponsorship,
		},
		{
			name: "EnableAmendment fee and reserve sponsorship",
			fields: map[string]any{
				"TransactionType": EnableAmendmentTx.String(),
				"Account":         "rrrrrrrrrrrrrrrrrrrrrhoLvTp",
				"Sponsor":         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":    types.SpfSponsorFee | types.SpfSponsorReserve,
			},
			expectedErr: ErrPseudoTransactionSponsorship,
		},
		{
			name: "SetFee fee sponsorship",
			fields: map[string]any{
				"TransactionType": SetFeeTx.String(),
				"Account":         "rrrrrrrrrrrrrrrrrrrrrhoLvTp",
				"Sponsor":         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":    types.SpfSponsorFee,
			},
			expectedErr: ErrPseudoTransactionSponsorship,
		},
		{
			name: "SetFee reserve sponsorship",
			fields: map[string]any{
				"TransactionType": SetFeeTx.String(),
				"Account":         "rrrrrrrrrrrrrrrrrrrrrhoLvTp",
				"Sponsor":         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":    types.SpfSponsorReserve,
			},
			expectedErr: ErrPseudoTransactionSponsorship,
		},
		{
			name: "SetFee fee and reserve sponsorship",
			fields: map[string]any{
				"TransactionType": SetFeeTx.String(),
				"Account":         "rrrrrrrrrrrrrrrrrrrrrhoLvTp",
				"Sponsor":         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":    types.SpfSponsorFee | types.SpfSponsorReserve,
			},
			expectedErr: ErrPseudoTransactionSponsorship,
		},
		{
			name: "UNLModify fee sponsorship",
			fields: map[string]any{
				"TransactionType": UNLModifyTx.String(),
				"Account":         "rrrrrrrrrrrrrrrrrrrrrhoLvTp",
				"Sponsor":         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":    types.SpfSponsorFee,
			},
			expectedErr: ErrPseudoTransactionSponsorship,
		},
		{
			name: "UNLModify reserve sponsorship",
			fields: map[string]any{
				"TransactionType": UNLModifyTx.String(),
				"Account":         "rrrrrrrrrrrrrrrrrrrrrhoLvTp",
				"Sponsor":         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":    types.SpfSponsorReserve,
			},
			expectedErr: ErrPseudoTransactionSponsorship,
		},
		{
			name: "UNLModify fee and reserve sponsorship",
			fields: map[string]any{
				"TransactionType": UNLModifyTx.String(),
				"Account":         "rrrrrrrrrrrrrrrrrrrrrhoLvTp",
				"Sponsor":         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":    types.SpfSponsorFee | types.SpfSponsorReserve,
			},
			expectedErr: ErrPseudoTransactionSponsorship,
		},
		{
			name: "flags omitted",
			fields: map[string]any{
				"Sponsor": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
			},
			expectedErr: ErrSponsorFieldsMissing,
		},
		{
			name: "sponsor omitted",
			fields: map[string]any{
				"SponsorFlags": types.SpfSponsorReserve,
			},
			expectedErr: ErrSponsorFieldsMissing,
		},
		{
			name: "signature only",
			fields: map[string]any{
				"SponsorSignature": map[string]any{"SigningPubKey": ""},
			},
			expectedErr: ErrSponsorFieldsMissing,
		},
		{
			name: "null sponsor with flags omitted",
			fields: map[string]any{
				"Sponsor": nil,
			},
			expectedErr: ErrInvalidSponsor,
		},
		{
			name: "null flags with sponsor omitted",
			fields: map[string]any{
				"SponsorFlags": nil,
			},
			expectedErr: ErrInvalidSponsorFlags,
		},
		{
			name: "zero flags without sponsor",
			fields: map[string]any{
				"SponsorFlags": uint32(0),
			},
			expectedErr: ErrSponsorFieldsMissing,
		},
		{
			name: "empty sponsor without flags",
			fields: map[string]any{
				"Sponsor": "",
			},
			expectedErr: ErrSponsorFieldsMissing,
		},
		{
			name: "empty sponsor and zero flags",
			fields: map[string]any{
				"Sponsor":      "",
				"SponsorFlags": uint32(0),
			},
			expectedErr: ErrSponsorFieldsMissing,
		},
		{
			name: "reserve without signature",
			fields: map[string]any{
				"Sponsor":      "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags": types.SpfSponsorReserve,
			},
			expectedErr: nil,
		},
		{
			name: "reserve placeholder",
			fields: map[string]any{
				"Sponsor":          "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":     types.SpfSponsorReserve,
				"SponsorSignature": map[string]any{"SigningPubKey": ""},
			},
			expectedErr: nil,
		},
		{
			name: "fee",
			fields: map[string]any{
				"Sponsor":      "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags": types.SpfSponsorFee,
			},
			expectedErr: ErrInnerBatchFeeSponsorship,
		},
		{
			name: "both flags",
			fields: map[string]any{
				"Sponsor":      "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags": uint32(3),
			},
			expectedErr: ErrInnerBatchFeeSponsorship,
		},
		{
			name: "zero flags",
			fields: map[string]any{
				"Sponsor":      "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags": uint32(0),
			},
			expectedErr: ErrSponsorFieldsMissing,
		},
		{
			name: "unknown signature member before zero flags",
			fields: map[string]any{
				"Sponsor":          "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":     uint32(0),
				"SponsorSignature": map[string]any{"Account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59"},
			},
			expectedErr: ErrInnerBatchSponsorSignature,
		},
		{
			name: "forbidden signers before invalid flag bits",
			fields: map[string]any{
				"Sponsor":          "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":     uint32(4),
				"SponsorSignature": map[string]any{"Signers": nil},
			},
			expectedErr: ErrInnerBatchSponsorSignature,
		},
		{
			name: "fractional flags",
			fields: map[string]any{
				"Sponsor":      "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags": 2.5,
			},
			expectedErr: ErrInvalidSponsorFlags,
		},
		{
			name: "overflow flags",
			fields: map[string]any{
				"Sponsor":      "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags": uint64(1) << 32,
			},
			expectedErr: ErrInvalidSponsorFlags,
		},
		{
			name: "null flags",
			fields: map[string]any{
				"Sponsor":      "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags": nil,
			},
			expectedErr: ErrInvalidSponsorFlags,
		},
		{
			name: "null sponsor",
			fields: map[string]any{
				"Sponsor":      nil,
				"SponsorFlags": types.SpfSponsorReserve,
			},
			expectedErr: ErrInvalidSponsor,
		},
		{
			name: "empty sponsor",
			fields: map[string]any{
				"Sponsor":      "",
				"SponsorFlags": types.SpfSponsorReserve,
			},
			expectedErr: ErrSponsorFieldsMissing,
		},
		{
			name: "null signature",
			fields: map[string]any{
				"Sponsor":          "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":     types.SpfSponsorReserve,
				"SponsorSignature": nil,
			},
			expectedErr: ErrInvalidSponsorSignature,
		},
		{
			name: "empty signature object",
			fields: map[string]any{
				"Sponsor":          "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":     types.SpfSponsorReserve,
				"SponsorSignature": map[string]any{},
			},
			expectedErr: nil,
		},
		{
			name: "nil signature map",
			fields: map[string]any{
				"Sponsor":          "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":     types.SpfSponsorReserve,
				"SponsorSignature": map[string]any(nil),
			},
			expectedErr: ErrInnerBatchSponsorSignature,
		},
		{
			name: "null signing key",
			fields: map[string]any{
				"Sponsor":          "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":     types.SpfSponsorReserve,
				"SponsorSignature": map[string]any{"SigningPubKey": nil},
			},
			expectedErr: ErrInnerBatchSponsorSignature,
		},
		{
			name: "non-string signing key",
			fields: map[string]any{
				"Sponsor":          "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":     types.SpfSponsorReserve,
				"SponsorSignature": map[string]any{"SigningPubKey": 0},
			},
			expectedErr: ErrInnerBatchSponsorSignature,
		},
		{
			name: "empty signature without key",
			fields: map[string]any{
				"Sponsor":          "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":     types.SpfSponsorReserve,
				"SponsorSignature": map[string]any{"TxnSignature": ""},
			},
			expectedErr: ErrInnerBatchSponsorSignature,
		},
		{
			name: "null signature without key",
			fields: map[string]any{
				"Sponsor":          "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":     types.SpfSponsorReserve,
				"SponsorSignature": map[string]any{"TxnSignature": nil},
			},
			expectedErr: ErrInnerBatchSponsorSignature,
		},
		{
			name: "empty signers without key",
			fields: map[string]any{
				"Sponsor":          "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":     types.SpfSponsorReserve,
				"SponsorSignature": map[string]any{"Signers": []any{}},
			},
			expectedErr: ErrInnerBatchSponsorSignature,
		},
		{
			name: "null signers without key",
			fields: map[string]any{
				"Sponsor":          "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":     types.SpfSponsorReserve,
				"SponsorSignature": map[string]any{"Signers": nil},
			},
			expectedErr: ErrInnerBatchSponsorSignature,
		},
		{
			name: "extra serialized member",
			fields: map[string]any{
				"Sponsor":      "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags": types.SpfSponsorReserve,
				"SponsorSignature": map[string]any{
					"SigningPubKey": "",
					"Account":       "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				},
			},
			expectedErr: ErrInnerBatchSponsorSignature,
		},
		{
			name: "wrong sole member",
			fields: map[string]any{
				"Sponsor":          "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":     types.SpfSponsorReserve,
				"SponsorSignature": map[string]any{"Account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59"},
			},
			expectedErr: ErrInnerBatchSponsorSignature,
		},
		{
			name: "empty signature field",
			fields: map[string]any{
				"Sponsor":          "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":     types.SpfSponsorReserve,
				"SponsorSignature": map[string]any{"SigningPubKey": "", "TxnSignature": ""},
			},
			expectedErr: ErrInnerBatchSponsorSignature,
		},
		{
			name: "null signature field",
			fields: map[string]any{
				"Sponsor":          "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":     types.SpfSponsorReserve,
				"SponsorSignature": map[string]any{"SigningPubKey": "", "TxnSignature": nil},
			},
			expectedErr: ErrInnerBatchSponsorSignature,
		},
		{
			name: "empty signers",
			fields: map[string]any{
				"Sponsor":          "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":     types.SpfSponsorReserve,
				"SponsorSignature": map[string]any{"SigningPubKey": "", "Signers": []any{}},
			},
			expectedErr: ErrInnerBatchSponsorSignature,
		},
		{
			name: "nonempty key",
			fields: map[string]any{
				"Sponsor":          "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"SponsorFlags":     types.SpfSponsorReserve,
				"SponsorSignature": map[string]any{"SigningPubKey": "AB"},
			},
			expectedErr: ErrInnerBatchSponsorSignature,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inner := map[string]any{
				"TransactionType": "Payment",
				"Account":         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				"Flags":           types.TfInnerBatchTxn,
				"SigningPubKey":   "",
			}
			maps.Copy(inner, tt.fields)
			batch := sponsorTestBatch(inner)
			original, err := json.Marshal(inner)
			require.NoError(t, err)
			valid, err := batch.Validate()
			require.Equal(t, tt.expectedErr == nil, valid)
			if tt.expectedErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.expectedErr)
			}
			after, err := json.Marshal(inner)
			require.NoError(t, err)
			require.Equal(t, original, after)
		})
	}
}

func TestInnerSponsorshipValidationTypedAndRaw(t *testing.T) {
	const account = "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	const sponsor = "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59"
	key, empty := "AB", ""
	tests := []struct {
		name        string
		sponsor     types.Address
		flags       uint32
		signature   *types.SponsorSignature
		expectedErr error
		rawErr      error
	}{
		{"reserve", sponsor, types.SpfSponsorReserve, nil, nil, nil},
		{"empty object", sponsor, types.SpfSponsorReserve, &types.SponsorSignature{}, nil, nil},
		{"empty key", sponsor, types.SpfSponsorReserve, &types.SponsorSignature{SigningPubKey: &empty}, nil, nil},
		{"missing sponsor", "", types.SpfSponsorReserve, nil, ErrSponsorFieldsMissing, ErrSponsorFieldsMissing},
		{"zero flags", sponsor, 0, nil, ErrSponsorFieldsMissing, ErrSponsorFieldsMissing},
		{"invalid sponsor", "invalid", types.SpfSponsorReserve, nil, ErrInvalidSponsor, ErrInvalidSponsor},
		{"self sponsor", account, types.SpfSponsorReserve, nil, ErrSponsorAccountConflict, ErrSponsorAccountConflict},
		{"invalid flags", sponsor, 4, nil, ErrInvalidSponsorFlags, ErrInvalidSponsorFlags},
		{"fee", sponsor, types.SpfSponsorFee, nil, ErrInnerBatchFeeSponsorship, ErrInnerBatchFeeSponsorship},
		{"nonempty key", sponsor, types.SpfSponsorReserve, &types.SponsorSignature{SigningPubKey: &key}, ErrInnerBatchSponsorSignature, ErrInnerBatchSponsorSignature},
		{"signature present", sponsor, types.SpfSponsorReserve, &types.SponsorSignature{TxnSignature: &empty}, ErrInnerBatchSponsorSignature, ErrInnerBatchSponsorSignature},

		// Raw parsing rejects forbidden signature fields before validating sponsor flags.
		{"zero flags and signature", sponsor, 0, &types.SponsorSignature{TxnSignature: &empty}, ErrSponsorFieldsMissing, ErrInnerBatchSponsorSignature},
		{"invalid flags and signature", sponsor, 4, &types.SponsorSignature{TxnSignature: &empty}, ErrInvalidSponsorFlags, ErrInnerBatchSponsorSignature},
		{"fee and signature", sponsor, types.SpfSponsorFee, &types.SponsorSignature{TxnSignature: &empty}, ErrInnerBatchFeeSponsorship, ErrInnerBatchSponsorSignature},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := BaseTx{
				Account: account, TransactionType: PaymentTx, Flags: types.TfInnerBatchTxn,
				Sponsor: tt.sponsor, SponsorFlags: tt.flags, SponsorSignature: tt.signature,
			}
			raw := tx.Flatten()
			// Keep explicit zero values so conversion cannot hide them as absent fields.
			raw["Sponsor"] = tt.sponsor.String()
			raw["SponsorFlags"] = tt.flags
			batch := sponsorTestBatch(raw)
			typedValid, typedErr := tx.Validate()
			rawValid, rawErr := batch.Validate()
			require.Equal(t, tt.expectedErr == nil, typedValid)
			require.Equal(t, typedValid, rawValid)
			require.ErrorIs(t, typedErr, tt.expectedErr)
			require.ErrorIs(t, rawErr, tt.rawErr)
		})
	}
}

func sponsorTestBatch(inner map[string]any) Batch {
	return Batch{
		BaseTx: BaseTx{Account: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", TransactionType: BatchTx, Flags: TfAllOrNothing},
		RawTransactions: []types.RawTransaction{
			{RawTransaction: inner},
			{RawTransaction: map[string]any{
				"TransactionType": "AccountSet",
				"Account":         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				"Flags":           types.TfInnerBatchTxn,
				"SigningPubKey":   "",
			}},
		},
	}
}

func TestReserveSponsorshipAllowList(t *testing.T) {
	for _, txType := range []TxType{
		DelegateSetTx, DepositPreauthTx, PaymentTx, SignerListSetTx,
		CheckCancelTx, CheckCashTx, CheckCreateTx, EscrowCancelTx, EscrowCreateTx, EscrowFinishTx,
		PaymentChannelClaimTx, PaymentChannelCreateTx, PaymentChannelFundTx, ClawbackTx,
		MPTokenAuthorizeTx, MPTokenIssuanceCreateTx, MPTokenIssuanceDestroyTx, MPTokenIssuanceSetTx,
		TrustSetTx, CredentialAcceptTx, CredentialCreateTx, CredentialDeleteTx,
		AccountSetTx, SetRegularKeyTx, SponsorshipTransferTx,
	} {
		t.Run(txType.String(), func(t *testing.T) {
			tx := BaseTx{
				Account: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", TransactionType: txType,
				Sponsor: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59", SponsorFlags: types.SpfSponsorReserve,
			}
			valid, err := tx.Validate()
			require.NoError(t, err)
			require.True(t, valid)
		})
	}
}

func TestSponsorSignatureRequiresSponsorAndFlags(t *testing.T) {
	key, signature := "AB", "CD"
	tests := []struct {
		name    string
		sponsor types.Address
		flags   uint32
	}{
		{"neither", "", 0},
		{"sponsor only", "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59", 0},
		{"flags only", "", types.SpfSponsorFee},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := BaseTx{
				Account:          "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				TransactionType:  PaymentTx,
				Sponsor:          tt.sponsor,
				SponsorFlags:     tt.flags,
				SponsorSignature: &types.SponsorSignature{SigningPubKey: &key, TxnSignature: &signature},
			}
			valid, err := tx.Validate()
			require.False(t, valid)
			require.ErrorIs(t, err, ErrSponsorFieldsMissing)
		})
	}
}

func TestSponsorAddressErrorsWrapSharedConditions(t *testing.T) {
	tests := []struct {
		name      string
		sponsor   types.Address
		condition error
	}{
		{"zero", "rrrrrrrrrrrrrrrrrrrrrhoLvTp", ErrZeroAccountID},
		{"tag", "X7AcgcsBL6XDcUb289X4mJ8djcdyKaGZMhc9YTE92ehJ2Fu", ErrAccountIDTagNotAllowed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := BaseTx{
				Account:         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				TransactionType: PaymentTx,
				Sponsor:         tt.sponsor,
				SponsorFlags:    types.SpfSponsorFee,
			}
			valid, err := tx.Validate()
			require.False(t, valid)
			require.ErrorIs(t, err, ErrInvalidSponsor)
			require.ErrorIs(t, err, tt.condition)
		})
	}
}

func TestInnerSponsorSignatureErrorWrapsInvalidSignature(t *testing.T) {
	key := "AB"
	err := validateSponsorSignature(&types.SponsorSignature{SigningPubKey: &key}, true)
	require.ErrorIs(t, err, ErrInnerBatchSponsorSignature)
	require.ErrorIs(t, err, ErrInvalidSponsorSignature)
}

// TestReserveSponsorshipRejectsUnsupportedTypes covers the complement of the
// pinned reserve allow-list, except pseudo-transactions. TestTx_Validate covers
// their separate policy, which rejects all sponsorship.
// Keep this explicit list independent of isReserveSponsorable.
func TestReserveSponsorshipRejectsUnsupportedTypes(t *testing.T) {
	for _, txType := range []TxType{
		AMMBidTx,
		AMMClawbackTx,
		AMMCreateTx,
		AMMDeleteTx,
		AMMDepositTx,
		AMMVoteTx,
		AMMWithdrawTx,
		AccountDeleteTx,
		BinaryTx,
		BatchTx,
		ConfidentialMPTClawbackTx,
		ConfidentialMPTConvertTx,
		ConfidentialMPTConvertBackTx,
		ConfidentialMPTMergeInboxTx,
		ConfidentialMPTSendTx,
		DIDDeleteTx,
		DIDSetTx,
		HashedTx,
		"Invalid",
		LedgerStateFixTx,
		LoanBrokerCoverClawbackTx,
		LoanBrokerCoverDepositTx,
		LoanBrokerCoverWithdrawTx,
		LoanBrokerDeleteTx,
		LoanBrokerSetTx,
		LoanDeleteTx,
		LoanManageTx,
		LoanPayTx,
		LoanSetTx,
		NFTokenAcceptOfferTx,
		NFTokenBurnTx,
		NFTokenCancelOfferTx,
		NFTokenCreateOfferTx,
		NFTokenMintTx,
		NFTokenModifyTx,
		OfferCancelTx,
		OfferCreateTx,
		OracleDeleteTx,
		OracleSetTx,
		PermissionedDomainDeleteTx,
		PermissionedDomainSetTx,
		SponsorshipSetTx,
		TicketCreateTx,
		VaultClawbackTx,
		VaultCreateTx,
		VaultDeleteTx,
		VaultDepositTx,
		VaultSetTx,
		VaultWithdrawTx,
		XChainAccountCreateCommitTx,
		XChainAddAccountCreateAttestationTx,
		XChainAddClaimAttestationTx,
		XChainClaimTx,
		XChainCommitTx,
		XChainCreateBridgeTx,
		XChainCreateClaimIDTx,
		XChainModifyBridgeTx,
		"UnknownTransaction",
	} {
		t.Run(txType.String(), func(t *testing.T) {
			tx := BaseTx{
				Account:         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				TransactionType: txType,
				Sponsor:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				SponsorFlags:    types.SpfSponsorReserve,
			}
			valid, err := tx.Validate()
			require.False(t, valid)
			require.ErrorIs(t, err, ErrReserveSponsorshipNotAllowed)
			tx.SponsorFlags |= types.SpfSponsorFee
			valid, err = tx.Validate()
			require.False(t, valid)
			require.ErrorIs(t, err, ErrReserveSponsorshipNotAllowed)
		})
	}
}
