package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestPayChannel(t *testing.T) {
	var s Object = &PayChannel{
		Account:           "rBqb89MRQJnMPq8wTwEbtz4kvxrEDfcYvt",
		Destination:       "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		Amount:            types.XRPCurrencyAmount(4325800),
		Balance:           types.XRPCurrencyAmount(2323423),
		PublicKey:         "32D2471DB72B27E3310F355BB33E339BF26F8392D5A93D3BC0FC3B566612DA0F0A",
		SettleDelay:       3600,
		Expiration:        536027313,
		CancelAfter:       536891313,
		SourceTag:         0,
		DestinationTag:    1002341,
		DestinationNode:   "0000000000000000",
		Flags:             0,
		LedgerEntryType:   PayChannelEntry,
		OwnerNode:         "0000000000000000",
		PreviousTxnID:     "F0AB71E777B2DA54B86231E19B82554EF1F8211F92ECA473121C655BFC5329BF",
		PreviousTxnLgrSeq: 14524914,
	}

	j := `{
	"Account": "rBqb89MRQJnMPq8wTwEbtz4kvxrEDfcYvt",
	"Amount": "4325800",
	"Balance": "2323423",
	"CancelAfter": 536891313,
	"Destination": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"DestinationTag": 1002341,
	"DestinationNode": "0000000000000000",
	"Expiration": 536027313,
	"Flags": 0,
	"LedgerEntryType": "PayChannel",
	"OwnerNode": "0000000000000000",
	"PreviousTxnID": "F0AB71E777B2DA54B86231E19B82554EF1F8211F92ECA473121C655BFC5329BF",
	"PreviousTxnLgrSeq": 14524914,
	"PublicKey": "32D2471DB72B27E3310F355BB33E339BF26F8392D5A93D3BC0FC3B566612DA0F0A",
	"SettleDelay": 3600
}`
	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestPayChannel_EntryType(t *testing.T) {
	s := &PayChannel{}
	require.Equal(t, s.EntryType(), PayChannelEntry)
}
