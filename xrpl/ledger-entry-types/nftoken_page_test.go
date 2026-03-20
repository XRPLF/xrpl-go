package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestNFTokenPage(t *testing.T) {
	var s Object = &NFTokenPage{
		LedgerEntryType:   NFTokenPageEntry,
		PreviousTxnID:     "95C8761B22894E328646F7A70035E9DFBECC90EDD83E43B7B973F626D21A0822",
		PreviousTxnLgrSeq: 42891441,
		NFTokens: []types.NFToken{
			{
				NFTokenID:  "000B013A95F14B0044F78A264E41713C64B5F89242540EE208C3098E00000D65",
				NFTokenURI: "697066733A2F2F62616679626569676479727A74357366703775646D37687537367568377932366E6634646675796C71616266336F636C67747179353566627A6469",
			},
		},
	}

	j := `{
	"LedgerEntryType": "NFTokenPage",
	"Flags": 0,
	"PreviousTxnID": "95C8761B22894E328646F7A70035E9DFBECC90EDD83E43B7B973F626D21A0822",
	"PreviousTxnLgrSeq": 42891441,
	"NFTokens": [
		{
			"NFTokenID": "000B013A95F14B0044F78A264E41713C64B5F89242540EE208C3098E00000D65",
			"URI": "697066733A2F2F62616679626569676479727A74357366703775646D37687537367568377932366E6634646675796C71616266336F636C67747179353566627A6469"
		}
	]
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestNFTokenPage_EntryType(t *testing.T) {
	s := &NFTokenPage{}
	require.Equal(t, NFTokenPageEntry, s.EntryType())
}
