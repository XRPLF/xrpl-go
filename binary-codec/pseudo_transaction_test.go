package binarycodec

import (
	"maps"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeUNLModifyAccountWorkaround(t *testing.T) {
	const expected = "120066240000000026040B52006840000000000000007300701321EDB6FC8E803EE8EDC2793F1EC917B2EE41D35255618DEB91D3F9B1FC89B75D4539810000101101"

	base := map[string]any{
		"Fee":                "0",
		"LedgerSequence":     uint32(67850752),
		"Sequence":           uint32(0),
		"SigningPubKey":      "",
		"TransactionType":    "UNLModify",
		"UNLModifyDisabling": uint8(1),
		"UNLModifyValidator": "EDB6FC8E803EE8EDC2793F1EC917B2EE41D35255618DEB91D3F9B1FC89B75D4539",
	}

	for _, account := range []string{"", "rrrrrrrrrrrrrrrrrrrrrhoLvTp"} {
		t.Run("Account "+account, func(t *testing.T) {
			tx := make(map[string]any, len(base)+1)
			maps.Copy(tx, base)
			tx["Account"] = account

			actual, err := Encode(tx)
			require.NoError(t, err)
			require.Equal(t, expected, actual)
		})
	}
}
