package ledger

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestCheck(t *testing.T) {
	var s Object = &Check{
		Account:           "rUn84CUYbNjRoTQ6mSW7BVJPSVJNLb1QLo",
		Destination:       "rfkE1aSy9G8Upk4JssnwBxhEv5p4mn2KTy",
		DestinationNode:   "0000000000000000",
		DestinationTag:    1,
		Expiration:        570113521,
		Flags:             0,
		InvoiceID:         "46060241FABCF692D4D934BA2A6C4427CD4279083E38C77CBE642243E43BE291",
		LedgerEntryType:   CheckEntry,
		OwnerNode:         "0000000000000000",
		PreviousTxnID:     "5463C6E08862A1FAE5EDAC12D70ADB16546A1F674930521295BC082494B62924",
		PreviousTxnLgrSeq: 6,
		Sequence:          2,
		SendMax:           types.XRPCurrencyAmount(1000),
	}

	j := `{
	"LedgerEntryType": "Check",
	"Flags": 0,
	"Account": "rUn84CUYbNjRoTQ6mSW7BVJPSVJNLb1QLo",
	"Destination": "rfkE1aSy9G8Upk4JssnwBxhEv5p4mn2KTy",
	"DestinationNode": "0000000000000000",
	"DestinationTag": 1,
	"Expiration": 570113521,
	"InvoiceID": "46060241FABCF692D4D934BA2A6C4427CD4279083E38C77CBE642243E43BE291",
	"OwnerNode": "0000000000000000",
	"PreviousTxnID": "5463C6E08862A1FAE5EDAC12D70ADB16546A1F674930521295BC082494B62924",
	"PreviousTxnLgrSeq": 6,
	"SendMax": "1000",
	"Sequence": 2
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestCheckUnmarshalAmounts(t *testing.T) {
	tests := []struct {
		name   string
		amount string
		want   types.CurrencyAmount
	}{
		{"XRP", `"9007199254740993"`, types.XRPCurrencyAmount(9007199254740993)},
		{"IOU", `{"currency":"USD","issuer":"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh","value":"12.5"}`, types.IssuedCurrencyAmount{Currency: "USD", Issuer: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", Value: "12.5"}},
		{"MPT", `{"mpt_issuance_id":"00000001A407AF5856CE97C4D485191ACFDCB37A7DEB3EDC","value":"10"}`, types.MPTCurrencyAmount{MPTIssuanceID: "00000001A407AF5856CE97C4D485191ACFDCB37A7DEB3EDC", Value: "10"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := fmt.Sprintf(`{
				"index":"13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
				"LedgerEntryType":"Check","Flags":0,
				"Account":"rUn84CUYbNjRoTQ6mSW7BVJPSVJNLb1QLo",
				"Destination":"rfkE1aSy9G8Upk4JssnwBxhEv5p4mn2KTy",
				"DestinationNode":"ABCDEF","DestinationTag":1,"Expiration":570113521,
				"InvoiceID":"46060241FABCF692D4D934BA2A6C4427CD4279083E38C77CBE642243E43BE291",
				"OwnerNode":"123ABC",
				"PreviousTxnID":"5463C6E08862A1FAE5EDAC12D70ADB16546A1F674930521295BC082494B62924",
				"PreviousTxnLgrSeq":6,"SendMax":%s,"Sequence":2,"SourceTag":3,
				"Sponsor":"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
			}`, tt.amount)
			want := Check{
				Index:           "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
				LedgerEntryType: CheckEntry,
				Account:         "rUn84CUYbNjRoTQ6mSW7BVJPSVJNLb1QLo",
				Destination:     "rfkE1aSy9G8Upk4JssnwBxhEv5p4mn2KTy",
				DestinationNode: "ABCDEF", DestinationTag: 1, Expiration: 570113521,
				InvoiceID:         "46060241FABCF692D4D934BA2A6C4427CD4279083E38C77CBE642243E43BE291",
				OwnerNode:         "123ABC",
				PreviousTxnID:     "5463C6E08862A1FAE5EDAC12D70ADB16546A1F674930521295BC082494B62924",
				PreviousTxnLgrSeq: 6, SendMax: tt.want, Sequence: 2, SourceTag: 3,
				Sponsor: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			}
			object, err := EmptyLedgerObject("Check")
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal([]byte(fixture), object))
			require.Equal(t, &want, object)
			encoded, err := json.Marshal(object)
			require.NoError(t, err)
			require.JSONEq(t, fixture, string(encoded))
			var roundtrip Check
			require.NoError(t, json.Unmarshal(encoded, &roundtrip))
			require.Equal(t, want, roundtrip)
		})
	}
}

func TestCheckUnmarshalEmptyAmount(t *testing.T) {
	for _, fixture := range []string{`{}`, `{"SendMax":null}`, "{\"SendMax\": \n null }"} {
		t.Run(fixture, func(t *testing.T) {
			check := Check{Sponsor: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", SendMax: types.XRPCurrencyAmount(10), Sequence: 2}
			require.NoError(t, json.Unmarshal([]byte(fixture), &check))
			require.Equal(t, Check{}, check)
			encoded, err := json.Marshal(check)
			require.NoError(t, err)
			var fields map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(encoded, &fields))
			require.Equal(t, "null", string(fields["SendMax"]))
		})
	}
}

func TestCheckUnmarshalNull(t *testing.T) {
	for _, fixture := range []string{"null", " \n null\t "} {
		t.Run(fixture, func(t *testing.T) {
			check := Check{Sponsor: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", SendMax: types.XRPCurrencyAmount(10), Sequence: 2}
			before, err := json.Marshal(check)
			require.NoError(t, err)
			// Call directly as well as through encoding/json to cover whitespace.
			require.NoError(t, check.UnmarshalJSON([]byte(fixture)))
			require.NoError(t, json.Unmarshal([]byte(fixture), &check))
			after, err := json.Marshal(check)
			require.NoError(t, err)
			require.Equal(t, string(before), string(after))
		})
	}
}

func TestCheckUnmarshalErrors(t *testing.T) {
	tests := []struct {
		name, fixture string
		wantErr       error
		amountError   bool
	}{
		{"invalid XRP", `{"Sequence":99,"SendMax":"bad"}`, strconv.ErrSyntax, true},
		{"array amount", `{"Sequence":99,"SendMax":[]}`, nil, true},
		{"invalid IOU value", `{"Sequence":99,"SendMax":{"currency":"USD","value":1}}`, nil, true},
		{"mixed amount fields", `{"Sequence":99,"SendMax":{"currency":"USD","mpt_issuance_id":"id"}}`, types.ErrMixedCurrencyAmountFields, true},
		{"ordinary field error", `{"SendMax":"10","Sequence":"bad"}`, nil, false},
		{"top-level array", `[]`, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := Check{Sponsor: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", SendMax: types.XRPCurrencyAmount(10), Sequence: 2}
			before, err := json.Marshal(check)
			require.NoError(t, err)
			err = json.Unmarshal([]byte(tt.fixture), &check)
			require.Error(t, err)
			require.Equal(t, tt.amountError, strings.HasPrefix(err.Error(), "Check.SendMax: "))
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			}
			after, err := json.Marshal(check)
			require.NoError(t, err)
			require.Equal(t, string(before), string(after))
		})
	}
}

func TestCheck_EntryType(t *testing.T) {
	s := &Check{}
	require.Equal(t, CheckEntry, s.EntryType())
}
