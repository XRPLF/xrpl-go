package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestAmmendments(t *testing.T) {
	var s Object = &Amendments{
		Amendments: []types.Hash256{
			"42426C4D4F1009EE67080A9B7965B44656D7714D104A72F9B4369F97ABF044EE",
			"4C97EBA926031A7CF7D7B36FDE3ED66DDA5421192D63DE53FFB46E43B9DC8373",
			"6781F8368C4771B83E8B821D88F580202BCB4228075297B19E4FDC5233F1EFDC",
			"740352F2412A9909880C23A559FCECEDA3BE2126FED62FC7660D628A06927F11",
		},
		Flags:           0,
		LedgerEntryType: AmendmentsEntry,
		Majorities: []MajorityEntry{
			{
				Majority: Majority{
					Amendment: "1562511F573A19AE9BD103B5D6B9E01B3B46805AEC5D3C4805C902B514399146",
					CloseTime: 535589001,
				},
			},
		},
		PreviousTxnID:     "1562511F573A19AE9BD103B5D6B9E01B3B46805AEC5D3C4805C902B514399146",
		PreviousTxnLgrSeq: 535589001,
	}

	j := `{
	"Flags": 0,
	"LedgerEntryType": "Amendments",
	"Amendments": [
		"42426C4D4F1009EE67080A9B7965B44656D7714D104A72F9B4369F97ABF044EE",
		"4C97EBA926031A7CF7D7B36FDE3ED66DDA5421192D63DE53FFB46E43B9DC8373",
		"6781F8368C4771B83E8B821D88F580202BCB4228075297B19E4FDC5233F1EFDC",
		"740352F2412A9909880C23A559FCECEDA3BE2126FED62FC7660D628A06927F11"
	],
	"Majorities": [
		{
			"Majority": {
				"Amendment": "1562511F573A19AE9BD103B5D6B9E01B3B46805AEC5D3C4805C902B514399146",
				"CloseTime": 535589001
			}
		}
	],
	"PreviousTxnID": "1562511F573A19AE9BD103B5D6B9E01B3B46805AEC5D3C4805C902B514399146",
	"PreviousTxnLgrSeq": 535589001
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestAmendments_EntryType(t *testing.T) {
	am := &Amendments{}
	require.Equal(t, AmendmentsEntry, am.EntryType())
}
