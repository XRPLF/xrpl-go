package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/require"
)

func TestNegativeUNL(t *testing.T) {
	var s Object = &NegativeUNL{
		DisabledValidators: []DisabledValidatorEntry{
			{
				DisabledValidator: DisabledValidator{
					FirstLedgerSequence: 1609728,
					PublicKey:           "ED6629D456285AE3613B285F65BBFF168D695BA3921F309949AFCD2CA7AFEC16FE",
				},
			},
		},
		Flags:           0,
		LedgerEntryType: NegativeUNLEntry,
	}

	j := `{
	"Flags": 0,
	"LedgerEntryType": "NegativeUNL",
	"DisabledValidators": [
		{
			"DisabledValidator": {
				"FirstLedgerSequence": 1609728,
				"PublicKey": "ED6629D456285AE3613B285F65BBFF168D695BA3921F309949AFCD2CA7AFEC16FE"
			}
		}
	]
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestNegativeUNL_EntryType(t *testing.T) {
	s := &NegativeUNL{}
	require.Equal(t, NegativeUNLEntry, s.EntryType())
}
