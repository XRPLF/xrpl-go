package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestSponsorship_EntryType(t *testing.T) {
	sponsorship := &Sponsorship{}
	require.Equal(t, SponsorshipEntry, sponsorship.EntryType())
}

func TestSponsorship_EmptyLedgerObject(t *testing.T) {
	object, err := EmptyLedgerObject(string(SponsorshipEntry))
	require.NoError(t, err)
	require.Equal(t, &Sponsorship{}, object)
}

func TestSponsorship_Serialization(t *testing.T) {
	feeAmount := types.XRPCurrencyAmount(1000000)
	maxFee := types.XRPCurrencyAmount(1000)
	remainingOwnerCount := uint32(5)

	tests := []struct {
		name        string
		sponsorship *Sponsorship
		expected    string
	}{
		{
			name: "pass - valid Sponsorship",
			sponsorship: &Sponsorship{
				Index:               types.Hash256("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
				LedgerEntryType:     SponsorshipEntry,
				Flags:               LsfSponsorshipRequireSignForFee | LsfSponsorshipRequireSignForReserve,
				Owner:               types.Address("rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH"),
				Sponsee:             types.Address("rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf"),
				FeeAmount:           &feeAmount,
				MaxFee:              &maxFee,
				RemainingOwnerCount: &remainingOwnerCount,
				OwnerNode:           "0000000000000000",
				SponseeNode:         "0000000000000001",
				PreviousTxnID:       types.Hash256("F19AD4577212D3BEACA0F75FE1BA1644F2E854D46E8D62E9C95D18E9708CBFB1"),
				PreviousTxnLgrSeq:   12345678,
			},
			expected: `{
	"index": "A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9",
	"LedgerEntryType": "Sponsorship",
	"Flags": 196608,
	"Owner": "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
	"Sponsee": "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
	"FeeAmount": "1000000",
	"MaxFee": "1000",
	"RemainingOwnerCount": 5,
	"OwnerNode": "0000000000000000",
	"SponseeNode": "0000000000000001",
	"PreviousTxnID": "F19AD4577212D3BEACA0F75FE1BA1644F2E854D46E8D62E9C95D18E9708CBFB1",
	"PreviousTxnLgrSeq": 12345678
}`,
		},
		{
			name: "pass - reserve-only Sponsorship omits fee budget fields",
			sponsorship: &Sponsorship{
				LedgerEntryType:     SponsorshipEntry,
				Flags:               0,
				Owner:               types.Address("rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH"),
				Sponsee:             types.Address("rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf"),
				RemainingOwnerCount: &remainingOwnerCount,
				OwnerNode:           "0000000000000000",
				SponseeNode:         "0000000000000000",
				PreviousTxnID:       types.Hash256("F19AD4577212D3BEACA0F75FE1BA1644F2E854D46E8D62E9C95D18E9708CBFB1"),
				PreviousTxnLgrSeq:   12345678,
			},
			expected: `{
	"LedgerEntryType": "Sponsorship",
	"Flags": 0,
	"Owner": "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
	"Sponsee": "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
	"RemainingOwnerCount": 5,
	"OwnerNode": "0000000000000000",
	"SponseeNode": "0000000000000000",
	"PreviousTxnID": "F19AD4577212D3BEACA0F75FE1BA1644F2E854D46E8D62E9C95D18E9708CBFB1",
	"PreviousTxnLgrSeq": 12345678
}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := testutil.SerializeAndDeserialize(t, test.sponsorship, test.expected); err != nil {
				t.Error(err)
			}
		})
	}
}
