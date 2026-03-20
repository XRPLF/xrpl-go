package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestDelegate_EntryType(t *testing.T) {
	delegate := &Delegate{}
	require.Equal(t, DelegateEntry, delegate.EntryType())
}

func TestDelegate_Serialization(t *testing.T) {
	tests := []struct {
		name     string
		delegate *Delegate
		expected string
	}{
		{
			name: "pass - valid Delegate",
			delegate: &Delegate{
				Index:           types.Hash256("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
				LedgerEntryType: DelegateEntry,
				Flags:           0,
				Account:         types.Address("rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH"),
				Authorize:       types.Address("rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf"),
				Permissions: []types.Permission{
					{
						Permission: types.PermissionValue{
							PermissionValue: "Payment",
						},
					},
					{
						Permission: types.PermissionValue{
							PermissionValue: "TrustlineAuthorize",
						},
					},
				},
				OwnerNode:         "0000000000000000",
				PreviousTxnID:     types.Hash256("F19AD4577212D3BEACA0F75FE1BA1644F2E854D46E8D62E9C95D18E9708CBFB1"),
				PreviousTxnLgrSeq: 4,
			},
			expected: `{
	"index": "A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9",
	"LedgerEntryType": "Delegate",
	"Flags": 0,
	"Account": "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
	"Authorize": "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
	"Permissions": [
		{
			"Permission": {
				"PermissionValue": "Payment"
			}
		},
		{
			"Permission": {
				"PermissionValue": "TrustlineAuthorize"
			}
		}
	],
	"OwnerNode": "0000000000000000",
	"PreviousTxnID": "F19AD4577212D3BEACA0F75FE1BA1644F2E854D46E8D62E9C95D18E9708CBFB1",
	"PreviousTxnLgrSeq": 4
}`,
		},
		{
			name: "pass - minimal Delegate",
			delegate: &Delegate{
				LedgerEntryType: DelegateEntry,
				Flags:           0,
				Account:         types.Address("rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH"),
				Authorize:       types.Address("rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf"),
				Permissions: []types.Permission{
					{
						Permission: types.PermissionValue{
							PermissionValue: "Payment",
						},
					},
				},
				OwnerNode:         "0000000000000000",
				PreviousTxnID:     types.Hash256("F19AD4577212D3BEACA0F75FE1BA1644F2E854D46E8D62E9C95D18E9708CBFB1"),
				PreviousTxnLgrSeq: 4,
			},
			expected: `{
	"LedgerEntryType": "Delegate",
	"Flags": 0,
	"Account": "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
	"Authorize": "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
	"Permissions": [
		{
			"Permission": {
				"PermissionValue": "Payment"
			}
		}
	],
	"OwnerNode": "0000000000000000",
	"PreviousTxnID": "F19AD4577212D3BEACA0F75FE1BA1644F2E854D46E8D62E9C95D18E9708CBFB1",
	"PreviousTxnLgrSeq": 4
}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := testutil.SerializeAndDeserialize(t, test.delegate, test.expected); err != nil {
				t.Error(err)
			}
		})
	}
}
