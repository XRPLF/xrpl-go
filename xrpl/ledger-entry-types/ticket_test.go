package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/require"
)

func TestTicket(t *testing.T) {
	var s Object = &Ticket{
		Account:           "rEhxGqkqPPSxQ3P25J66ft5TwpzV14k2de",
		Flags:             0,
		LedgerEntryType:   TicketEntry,
		OwnerNode:         "0000000000000000",
		PreviousTxnID:     "F19AD4577212D3BEACA0F75FE1BA1644F2E854D46E8D62E9C95D18E9708CBFB1",
		PreviousTxnLgrSeq: 4,
		TicketSequence:    3,
	}

	j := `{
	"Account": "rEhxGqkqPPSxQ3P25J66ft5TwpzV14k2de",
	"Flags": 0,
	"LedgerEntryType": "Ticket",
	"OwnerNode": "0000000000000000",
	"PreviousTxnID": "F19AD4577212D3BEACA0F75FE1BA1644F2E854D46E8D62E9C95D18E9708CBFB1",
	"PreviousTxnLgrSeq": 4,
	"TicketSequence": 3
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestTicket_EntryType(t *testing.T) {
	s := &Ticket{}
	require.Equal(t, TicketEntry, s.EntryType())
}
