package types

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	transactiontypes "github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestBatchSignable_Flatten(t *testing.T) {
	tc := []struct {
		name string
		bs   BatchSignable
		want map[string]any
	}{
		{
			name: "pass - common BatchV1_1 fields",
			bs: BatchSignable{
				Account:  "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
				Sequence: 5,
				Flags:    1,
				TxIDs:    []string{"tx1", "tx2"},
			},
			want: map[string]any{
				"account":  "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
				"sequence": uint32(5),
				"flags":    uint32(1),
				"txIDs":    []string{"tx1", "tx2"},
			},
		},
		{
			name: "pass - Batch signer and nested signer accounts",
			bs: BatchSignable{
				Account:       "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
				Sequence:      5,
				Flags:         1,
				TxIDs:         []string{"tx1", "tx2"},
				BatchAccount:  "rJCxK2hX9tDMzbnn3cg1GU2g19Kfmhzxkp",
				SignerAccount: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			},
			want: map[string]any{
				"account":       "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
				"sequence":      uint32(5),
				"flags":         uint32(1),
				"txIDs":         []string{"tx1", "tx2"},
				"batchAccount":  "rJCxK2hX9tDMzbnn3cg1GU2g19Kfmhzxkp",
				"signerAccount": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			},
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.bs.Flatten()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestFlatBatchSequenceValue(t *testing.T) {
	tests := []struct {
		name        string
		tx          transaction.FlatTransaction
		want        uint32
		expectedErr error
	}{
		{
			name: "Sequence",
			tx:   transaction.FlatTransaction{"Sequence": uint32(5)},
			want: 5,
		},
		{
			name: "TicketSequence",
			tx: transaction.FlatTransaction{
				"Sequence":       uint32(0),
				"TicketSequence": uint32(7),
			},
			want: 7,
		},
		{
			name: "TicketSequence without Sequence",
			tx:   transaction.FlatTransaction{"TicketSequence": uint32(7)},
			want: 7,
		},
		{
			name: "Sequence with zero TicketSequence",
			tx: transaction.FlatTransaction{
				"Sequence":       uint32(5),
				"TicketSequence": uint32(0),
			},
			want: 5,
		},
		{
			name: "both non-zero",
			tx: transaction.FlatTransaction{
				"Sequence":       uint32(5),
				"TicketSequence": uint32(7),
			},
			expectedErr: ErrBatchSequenceAndTicket,
		},
		{
			name:        "missing Sequence and TicketSequence",
			tx:          transaction.FlatTransaction{},
			expectedErr: ErrBatchSequenceNotSet,
		},
		{
			name:        "malformed Sequence",
			tx:          transaction.FlatTransaction{"Sequence": "invalid"},
			expectedErr: ErrSequenceFieldIsNotAnUint32,
		},
		{
			name: "malformed TicketSequence",
			tx: transaction.FlatTransaction{
				"Sequence":       uint32(5),
				"TicketSequence": "invalid",
			},
			expectedErr: ErrTicketSequenceFieldIsNotAnUint32,
		},
		{
			name: "zero TicketSequence only",
			tx: transaction.FlatTransaction{
				"Sequence":       uint32(0),
				"TicketSequence": uint32(0),
			},
			expectedErr: ErrBatchSequenceNotSet,
		},
		{
			name:        "neither effective value",
			tx:          transaction.FlatTransaction{"Sequence": uint32(0)},
			expectedErr: ErrBatchSequenceNotSet,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := flatBatchSequenceValue(tc.tx)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestBatchSignable_SequenceValue(t *testing.T) {
	const account = "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2"

	tests := []struct {
		name        string
		sequence    uint32
		ticket      uint32
		want        uint32
		expectedErr error
	}{
		{name: "Sequence", sequence: 5, want: 5},
		{name: "TicketSequence", ticket: 7, want: 7},
		{name: "both", sequence: 5, ticket: 7, expectedErr: ErrBatchSequenceAndTicket},
		{name: "neither", expectedErr: ErrBatchSequenceNotSet},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tx := transaction.Batch{
				BaseTx: transaction.BaseTx{
					Account:         transactiontypes.Address(account),
					Sequence:        tc.sequence,
					TicketSequence:  tc.ticket,
					TransactionType: transaction.BatchTx,
				},
			}
			got, err := FromBatchTransaction(&tx)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got.Sequence)
		})
	}
}
