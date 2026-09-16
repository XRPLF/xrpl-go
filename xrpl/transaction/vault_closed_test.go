package transaction

import (
	"encoding/json"
	"math"
	"testing"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestVaultCreateClosedValidation(t *testing.T) {
	open, closed, invalid, maxKind := types.VaultKindOpen, types.VaultKindClosed, types.VaultKind(2), types.VaultKind(math.MaxUint8)
	zero, sub, before, tooShort, minDate := uint32(0), uint32(1000), uint32(999), uint32(1179), uint32(1180)
	maxDate, tooLong, overflow := sub+946708559, sub+946708560, uint32(math.MaxUint32)
	lastSubscription := overflow - 180
	tests := []struct {
		name                     string
		kind                     *types.VaultKind
		subscription, redemption *uint32
		err                      error
	}{
		{"default open", nil, nil, nil, nil},
		{"explicit open", &open, nil, nil, nil},
		{"default with subscription", nil, &sub, nil, ErrVaultCreateDatesRequireClosedKind},
		{"open with redemption", &open, nil, &minDate, ErrVaultCreateDatesRequireClosedKind},
		{"open with both", &open, &sub, &minDate, ErrVaultCreateDatesRequireClosedKind},
		{"invalid kind", &invalid, nil, nil, ErrVaultCreateKindInvalid},
		{"maximum unsupported kind", &maxKind, nil, nil, ErrVaultCreateKindInvalid},
		{"both dates missing", &closed, nil, nil, ErrVaultCreateDatesRequired},
		{"subscription missing", &closed, nil, &minDate, ErrVaultCreateDatesRequired},
		{"redemption missing", &closed, &sub, nil, ErrVaultCreateDatesRequired},
		{"equal dates", &closed, &sub, &sub, ErrVaultCreateInvestmentPeriodInvalid},
		{"reversed dates", &closed, &sub, &before, ErrVaultCreateInvestmentPeriodInvalid},
		{"unsigned wraparound", &closed, &overflow, &minDate, ErrVaultCreateInvestmentPeriodInvalid},
		{"below minimum", &closed, &sub, &tooShort, ErrVaultCreateInvestmentPeriodInvalid},
		{"minimum", &closed, &sub, &minDate, nil},
		{"minimum at maximum timestamp", &closed, &lastSubscription, &overflow, nil},
		{"maximum minus one", &closed, &sub, &maxDate, nil},
		{"maximum excluded", &closed, &sub, &tooLong, ErrVaultCreateInvestmentPeriodInvalid},
		{"zero subscription present", &closed, &zero, &minDate, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tx := VaultCreate{
				BaseTx: BaseTx{
					TransactionType: VaultCreateTx,
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
				},
				Asset:            ledger.Asset{Currency: "XRP"},
				VaultKind:        tc.kind,
				SubscriptionDate: tc.subscription,
				RedemptionDate:   tc.redemption,
			}
			ok, err := tx.Validate()
			require.Equal(t, tc.err == nil, ok)
			require.ErrorIs(t, err, tc.err)
		})
	}
}

func TestVaultCreateClosedSerialization(t *testing.T) {
	open, closed := types.VaultKindOpen, types.VaultKindClosed
	zero, minimum := uint32(0), uint32(180)
	lastSubscription, maxDate := uint32(math.MaxUint32-180), uint32(math.MaxUint32)
	tests := []struct {
		name     string
		tx       VaultCreate
		wantJSON string
		wantFlat FlatTransaction
	}{
		{
			name:     "legacy fields omitted",
			tx:       VaultCreate{Asset: ledger.Asset{Currency: "XRP"}},
			wantJSON: `{"Account":"","TransactionType":"","Asset":{"currency":"XRP"}}`,
			wantFlat: FlatTransaction{
				"TransactionType": "VaultCreate",
				"Asset":           map[string]any{"currency": "XRP"},
			},
		},
		{
			name:     "explicit open kind preserved",
			tx:       VaultCreate{Asset: ledger.Asset{Currency: "XRP"}, VaultKind: &open},
			wantJSON: `{"Account":"","TransactionType":"","Asset":{"currency":"XRP"},"VaultKind":0}`,
			wantFlat: FlatTransaction{
				"TransactionType": "VaultCreate",
				"Asset":           map[string]any{"currency": "XRP"},
				"VaultKind":       uint8(0),
			},
		},
		{
			name: "closed with zero subscription",
			tx: VaultCreate{
				Asset:     ledger.Asset{Currency: "XRP"},
				VaultKind: &closed, SubscriptionDate: &zero, RedemptionDate: &minimum,
			},
			wantJSON: `{"Account":"","TransactionType":"","Asset":{"currency":"XRP"},"VaultKind":1,"SubscriptionDate":0,"RedemptionDate":180}`,
			wantFlat: FlatTransaction{
				"TransactionType":  "VaultCreate",
				"Asset":            map[string]any{"currency": "XRP"},
				"VaultKind":        uint8(1),
				"SubscriptionDate": uint32(0),
				"RedemptionDate":   uint32(180),
			},
		},
		{
			name: "maximum date",
			tx: VaultCreate{
				Asset:     ledger.Asset{Currency: "XRP"},
				VaultKind: &closed, SubscriptionDate: &lastSubscription, RedemptionDate: &maxDate,
			},
			wantJSON: `{"Account":"","TransactionType":"","Asset":{"currency":"XRP"},"VaultKind":1,"SubscriptionDate":4294967115,"RedemptionDate":4294967295}`,
			wantFlat: FlatTransaction{
				"TransactionType":  "VaultCreate",
				"Asset":            map[string]any{"currency": "XRP"},
				"VaultKind":        uint8(1),
				"SubscriptionDate": uint32(4294967115),
				"RedemptionDate":   uint32(4294967295),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("JSON", func(t *testing.T) {
				var decoded VaultCreate
				require.NoError(t, json.Unmarshal([]byte(tc.wantJSON), &decoded))
				require.Equal(t, tc.tx, decoded)

				data, err := json.Marshal(tc.tx)
				require.NoError(t, err)
				require.JSONEq(t, tc.wantJSON, string(data))
			})
			t.Run("flatten and binary", func(t *testing.T) {
				flat := tc.tx.Flatten()
				require.Equal(t, tc.wantFlat, flat)
				blob, err := binarycodec.Encode(flat)
				require.NoError(t, err)
				decoded, err := binarycodec.Decode(blob)
				require.NoError(t, err)
				require.Len(t, decoded, len(tc.wantFlat))
				for field, want := range tc.wantFlat {
					require.Contains(t, decoded, field)
					// UInt8 decodes as int, while Flatten uses uint8.
					require.EqualValues(t, want, decoded[field], field)
				}
			})
		})
	}
}
