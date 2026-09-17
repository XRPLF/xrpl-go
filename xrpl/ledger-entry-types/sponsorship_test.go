package ledger

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestSponsorship(t *testing.T) {
	const base = `"LedgerEntryType":"Sponsorship","Flags":196608,
		"Owner":"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh","Sponsee":"r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		"OwnerNode":"ABCDEF1234567890","SponseeNode":"1",
		"PreviousTxnID":"4294BEBE5B569A18C0A2702387C9B1E7146DC3A5850C1E87204951C6FDAA4C42",
		"PreviousTxnLgrSeq":3,"index":"13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8"`
	tests := []struct {
		name     string
		optional string
		present  bool
		fee      types.XRPCurrencyAmount
		maxFee   types.XRPCurrencyAmount
		count    uint32
	}{
		{name: "absent"},
		{name: "explicit zero", optional: `,"FeeAmount":"0","MaxFee":"0","RemainingOwnerCount":0`, present: true},
		{name: "large values", optional: `,"FeeAmount":"100000000000000000","MaxFee":"9007199254740993","RemainingOwnerCount":4294967295`, present: true, fee: 100000000000000000, maxFee: 9007199254740993, count: 4294967295},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			object, err := EmptyLedgerObject("Sponsorship")
			require.NoError(t, err)
			s, ok := object.(*Sponsorship)
			require.True(t, ok)
			require.Equal(t, SponsorshipEntry, s.EntryType())
			fixture := "{" + base + tt.optional + "}"
			require.NoError(t, json.Unmarshal([]byte(fixture), s))
			require.Equal(t, LsfSponsorshipRequireSignForFee|LsfSponsorshipRequireSignForReserve, s.Flags)
			if tt.present {
				require.NotNil(t, s.FeeAmount)
				require.NotNil(t, s.MaxFee)
				require.NotNil(t, s.RemainingOwnerCount)
				require.Equal(t, tt.fee, *s.FeeAmount)
				require.Equal(t, tt.maxFee, *s.MaxFee)
				require.Equal(t, tt.count, *s.RemainingOwnerCount)
			} else {
				require.Nil(t, s.FeeAmount)
				require.Nil(t, s.MaxFee)
				require.Nil(t, s.RemainingOwnerCount)
			}
			encoded, err := json.Marshal(s)
			require.NoError(t, err)
			require.JSONEq(t, fixture, string(encoded))
			var decoded Sponsorship
			require.NoError(t, json.Unmarshal(encoded, &decoded))
			require.Equal(t, *s, decoded)
		})
	}
}

func TestSponsorshipFlags(t *testing.T) {
	// Pinned rippled LedgerFormats.h (21890d9d).
	require.Equal(t, uint32(0x00010000), LsfSponsorshipRequireSignForFee)
	require.Equal(t, uint32(0x00020000), LsfSponsorshipRequireSignForReserve)
}

func TestLedgerSponsorFields(t *testing.T) {
	// Cover Sponsor presence and omission, including Check with absent SendMax.
	// Escrow has separate coverage.
	entries := []EntryType{
		AccountRootEntry, CheckEntry, CredentialEntry, DelegateEntry,
		DepositPreauthObjEntry, MPTokenEntry, MPTokenIssuanceEntry, PayChannelEntry, SignerListEntry,
	}
	for _, entry := range entries {
		for _, sponsor := range []string{"", "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"} {
			t.Run(fmt.Sprintf("%s/sponsor=%s", entry, sponsor), func(t *testing.T) {
				fixture := fmt.Sprintf(`{"LedgerEntryType":%q}`, entry)
				if sponsor != "" {
					fixture = fmt.Sprintf(`{"LedgerEntryType":%q,"Sponsor":%q}`, entry, sponsor)
				}
				object, err := EmptyLedgerObject(string(entry))
				require.NoError(t, err)
				require.NoError(t, json.Unmarshal([]byte(fixture), object))
				encoded, err := json.Marshal(object)
				require.NoError(t, err)
				var fields map[string]json.RawMessage
				require.NoError(t, json.Unmarshal(encoded, &fields))
				if sponsor == "" {
					require.NotContains(t, fields, "Sponsor")
				} else {
					require.JSONEq(t, fmt.Sprintf("%q", sponsor), string(fields["Sponsor"]))
				}
			})
		}
	}
}
