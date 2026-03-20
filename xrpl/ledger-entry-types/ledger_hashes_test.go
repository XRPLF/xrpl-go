package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestLedgerHashes(t *testing.T) {
	var s Object = &Hashes{
		FirstLedgerSequence: 2,
		Flags:               0,
		Hashes: []types.Hash256{
			"D638208ADBD04CBB10DE7B645D3AB4BA31489379411A3A347151702B6401AA78",
			"254D690864E418DDD9BCAC93F41B1F53B1AE693FC5FE667CE40205C322D1BE3B",
			"A2B31D28905E2DEF926362822BC412B12ABF6942B73B72A32D46ED2ABB7ACCFA",
			"AB4014846DF818A4B43D6B1686D0DE0644FE711577C5AB6F0B2A21CCEE280140",
			"3383784E82A8BA45F4DD5EF4EE90A1B2D3B4571317DBAC37B859836ADDE644C1",
		},
		LastLedgerSequence: 33872029,
		LedgerEntryType:    LedgerHashesEntry,
	}

	j := `{
	"Flags": 0,
	"LedgerEntryType": "LedgerHashes",
	"FirstLedgerSequence": 2,
	"Hashes": [
		"D638208ADBD04CBB10DE7B645D3AB4BA31489379411A3A347151702B6401AA78",
		"254D690864E418DDD9BCAC93F41B1F53B1AE693FC5FE667CE40205C322D1BE3B",
		"A2B31D28905E2DEF926362822BC412B12ABF6942B73B72A32D46ED2ABB7ACCFA",
		"AB4014846DF818A4B43D6B1686D0DE0644FE711577C5AB6F0B2A21CCEE280140",
		"3383784E82A8BA45F4DD5EF4EE90A1B2D3B4571317DBAC37B859836ADDE644C1"
	],
	"LastLedgerSequence": 33872029
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestLedgerHashes_EntryType(t *testing.T) {
	s := &Hashes{}
	require.Equal(t, LedgerHashesEntry, s.EntryType())
}
