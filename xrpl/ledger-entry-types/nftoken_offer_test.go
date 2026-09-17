package ledger

import (
	"encoding/json"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestNFTokenOffer(t *testing.T) {
	var s Object = &NFTokenOffer{
		Index:             "96F76F27D8A327FC48753167EC04A46AA0E382E6F57F32FD12274144D00F1797",
		Amount:            types.XRPCurrencyAmount(1000000),
		Flags:             1,
		LedgerEntryType:   NFTokenOfferEntry,
		NFTokenID:         "00081B5825A08C22787716FA031B432EBBC1B101BB54875F0002D2A400000000",
		NFTokenOfferNode:  "0",
		Owner:             "rhRxL3MNvuKEjWjL7TBbZSDacb8PmzAd7m",
		OwnerNode:         "17",
		PreviousTxnID:     "BFA9BE27383FA315651E26FDE1FA30815C5A5D0544EE10EC33D3E92532993769",
		PreviousTxnLgrSeq: 75443565,
	}

	j := `{
	"index": "96F76F27D8A327FC48753167EC04A46AA0E382E6F57F32FD12274144D00F1797",
	"Flags": 1,
	"LedgerEntryType": "NFTokenOffer",
	"Amount": "1000000",
	"NFTokenID": "00081B5825A08C22787716FA031B432EBBC1B101BB54875F0002D2A400000000",
	"NFTokenOfferNode": "0",
	"Owner": "rhRxL3MNvuKEjWjL7TBbZSDacb8PmzAd7m",
	"OwnerNode": "17",
	"PreviousTxnID": "BFA9BE27383FA315651E26FDE1FA30815C5A5D0544EE10EC33D3E92532993769",
	"PreviousTxnLgrSeq": 75443565
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestNFTokenOfferUnmarshalErrors(t *testing.T) {
	tests := []struct {
		name, fixture string
	}{
		{"invalid Amount", `{"Flags":99,"Amount":"bad"}`},
		{"ordinary field error after valid Owner", `{"Owner":"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh","Flags":"bad","Amount":"20"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offer := NFTokenOffer{
				Owner:  "rhRxL3MNvuKEjWjL7TBbZSDacb8PmzAd7m",
				Flags:  1,
				Amount: types.XRPCurrencyAmount(10),
			}
			before, err := json.Marshal(offer)
			require.NoError(t, err)
			require.Error(t, json.Unmarshal([]byte(tt.fixture), &offer))
			after, err := json.Marshal(offer)
			require.NoError(t, err)
			require.Equal(t, string(before), string(after))
		})
	}
}

func TestNFTokenOffer_EntryType(t *testing.T) {
	s := &NFTokenOffer{}
	require.Equal(t, NFTokenOfferEntry, s.EntryType())
}
