package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/require"
)

func TestSignerList(t *testing.T) {
	var s Object = &SignerList{
		LedgerEntryType:   SignerListEntry,
		OwnerNode:         "0000000000000000",
		PreviousTxnID:     "5904C0DC72C58A83AEFED2FFC5386356AA83FCA6A88C89D00646E51E687CDBE4",
		PreviousTxnLgrSeq: 16061435,
		SignerEntries: []SignerEntryWrapper{
			{
				SignerEntry: SignerEntry{
					Account:      "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
					SignerWeight: 2,
				},
			},
			{
				SignerEntry: SignerEntry{
					Account:      "raKEEVSGnKSD9Zyvxu4z6Pqpm4ABH8FS6n",
					SignerWeight: 1,
				},
			},
			{
				SignerEntry: SignerEntry{
					Account:      "rUpy3eEg8rqjqfUoLeBnZkscbKbFsKXC3v",
					SignerWeight: 1,
				},
			},
		},
		SignerListID: 0,
		SignerQuorum: 3,
	}

	j := `{
	"LedgerEntryType": "SignerList",
	"PreviousTxnID": "5904C0DC72C58A83AEFED2FFC5386356AA83FCA6A88C89D00646E51E687CDBE4",
	"PreviousTxnLgrSeq": 16061435,
	"OwnerNode": "0000000000000000",
	"SignerEntries": [
		{
			"SignerEntry": {
				"Account": "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				"SignerWeight": 2
			}
		},
		{
			"SignerEntry": {
				"Account": "raKEEVSGnKSD9Zyvxu4z6Pqpm4ABH8FS6n",
				"SignerWeight": 1
			}
		},
		{
			"SignerEntry": {
				"Account": "rUpy3eEg8rqjqfUoLeBnZkscbKbFsKXC3v",
				"SignerWeight": 1
			}
		}
	],
	"SignerListID": 0,
	"SignerQuorum": 3,
	"Flags": 0
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestSignerList_SetLsfOneOwnerCount(t *testing.T) {
	s := &SignerList{}
	s.SetLsfOneOwnerCount()
	require.Equal(t, s.Flags&LsfOneOwnerCount, LsfOneOwnerCount)
}

func TestSignerList_EntryType(t *testing.T) {
	s := &SignerList{}
	require.Equal(t, s.EntryType(), SignerListEntry)
}

func TestSignerEntryWrapper_Flatten(t *testing.T) {
	s := &SignerEntryWrapper{
		SignerEntry: SignerEntry{
			Account:      "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
			SignerWeight: 2,
		},
	}

	flattened := s.Flatten()
	require.Equal(t, flattened["SignerEntry"], s.SignerEntry.Flatten())
}

func TestSignerEntry_Flatten(t *testing.T) {
	s := &SignerEntry{
		Account:       "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
		SignerWeight:  2,
		WalletLocator: "0000000000000000000000000000000000000000000000000000000000000000",
	}

	flattened := s.Flatten()
	require.Equal(t, flattened["Account"], s.Account.String())
	require.Equal(t, flattened["SignerWeight"], int(s.SignerWeight))
	require.Equal(t, flattened["WalletLocator"], s.WalletLocator.String())
}
