package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestRippleState(t *testing.T) {
	var s Object = &RippleState{
		Balance: types.IssuedCurrencyAmount{
			Currency: "USD",
			Issuer:   "rrrrrrrrrrrrrrrrrrrrBZbvji",
			Value:    "-10",
		},
		Flags: 393216,
		HighLimit: types.IssuedCurrencyAmount{
			Currency: "USD",
			Issuer:   "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
			Value:    "110",
		},
		HighNode:        "0000000000000000",
		LedgerEntryType: RippleStateEntry,
		LowLimit: types.IssuedCurrencyAmount{
			Currency: "USD",
			Issuer:   "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
			Value:    "0",
		},
		LowNode:           "0000000000000000",
		PreviousTxnID:     "E3FE6EA3D48F0C2B639448020EA4F03D4F4F8FFDB243A852A0F59177921B4879",
		PreviousTxnLgrSeq: 14090896,
	}

	j := `{
	"Balance": {
		"issuer": "rrrrrrrrrrrrrrrrrrrrBZbvji",
		"currency": "USD",
		"value": "-10"
	},
	"Flags": 393216,
	"HighLimit": {
		"issuer": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		"currency": "USD",
		"value": "110"
	},
	"HighNode": "0000000000000000",
	"LedgerEntryType": "RippleState",
	"LowLimit": {
		"issuer": "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
		"currency": "USD",
		"value": "0"
	},
	"LowNode": "0000000000000000",
	"PreviousTxnID": "E3FE6EA3D48F0C2B639448020EA4F03D4F4F8FFDB243A852A0F59177921B4879",
	"PreviousTxnLgrSeq": 14090896
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestRippleState_EntryType(t *testing.T) {
	s := &RippleState{}
	require.Equal(t, s.EntryType(), RippleStateEntry)
}

func TestRippleState_SetLsfAMMNode(t *testing.T) {
	s := &RippleState{}
	s.SetLsfAMMNode()
	require.Equal(t, s.Flags&lsfAMMNode, lsfAMMNode)
}

func TestRippleState_SetLsfLowReserve(t *testing.T) {
	s := &RippleState{}
	s.SetLsfLowReserve()
	require.Equal(t, s.Flags&lsfLowReserve, lsfLowReserve)
}

func TestRippleState_SetLsfHighReserve(t *testing.T) {
	s := &RippleState{}
	s.SetLsfHighReserve()
	require.Equal(t, s.Flags&lsfHighReserve, lsfHighReserve)
}

func TestRippleState_SetLsfLowAuth(t *testing.T) {
	s := &RippleState{}
	s.SetLsfLowAuth()
	require.Equal(t, s.Flags&lsfLowAuth, lsfLowAuth)
}

func TestRippleState_SetLsfHighAuth(t *testing.T) {
	s := &RippleState{}
	s.SetLsfHighAuth()
	require.Equal(t, s.Flags&lsfHighAuth, lsfHighAuth)
}

func TestRippleState_SetLsfLowNoRipple(t *testing.T) {
	s := &RippleState{}
	s.SetLsfLowNoRipple()
	require.Equal(t, s.Flags&lsfLowNoRipple, lsfLowNoRipple)
}

func TestRippleState_SetLsfHighNoRipple(t *testing.T) {
	s := &RippleState{}
	s.SetLsfHighNoRipple()
	require.Equal(t, s.Flags&lsfHighNoRipple, lsfHighNoRipple)
}

func TestRippleState_SetLsfLowFreeze(t *testing.T) {
	s := &RippleState{}
	s.SetLsfLowFreeze()
	require.Equal(t, s.Flags&lsfLowFreeze, lsfLowFreeze)
}

func TestRippleState_SetLsfHighFreeze(t *testing.T) {
	s := &RippleState{}
	s.SetLsfHighFreeze()
	require.Equal(t, s.Flags&lsfHighFreeze, lsfHighFreeze)
}

func TestRippleState_SetLsfLowDeepFreeze(t *testing.T) {
	s := &RippleState{}
	s.SetLsfLowDeepFreeze()
	require.Equal(t, s.Flags&lsfLowDeepFreeze, lsfLowDeepFreeze)
}

func TestRippleState_SetLsfHighDeepFreeze(t *testing.T) {
	s := &RippleState{}
	s.SetLsfHighDeepFreeze()
	require.Equal(t, s.Flags&lsfHighDeepFreeze, lsfHighDeepFreeze)
}
