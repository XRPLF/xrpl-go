package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/require"
)

func TestDepositPreauth(t *testing.T) {
	var s Object = &DepositPreauthObj{
		LedgerEntryType:   DepositPreauthObjEntry,
		Account:           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		Authorize:         "rEhxGqkqPPSxQ3P25J66ft5TwpzV14k2de",
		Flags:             0,
		OwnerNode:         "0000000000000000",
		PreviousTxnID:     "3E8964D5A86B3CD6B9ECB33310D4E073D64C865A5B866200AD2B7E29F8326702",
		PreviousTxnLgrSeq: 7,
	}

	j := `{
	"Flags": 0,
	"LedgerEntryType": "DepositPreauth",
	"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"Authorize": "rEhxGqkqPPSxQ3P25J66ft5TwpzV14k2de",
	"OwnerNode": "0000000000000000",
	"PreviousTxnID": "3E8964D5A86B3CD6B9ECB33310D4E073D64C865A5B866200AD2B7E29F8326702",
	"PreviousTxnLgrSeq": 7
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestDepositPreauth_EntryType(t *testing.T) {
	dp := &DepositPreauthObj{}
	require.Equal(t, dp.EntryType(), DepositPreauthObjEntry)
}
