package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestDelegateSet_TxType(t *testing.T) {
	tx := &DelegateSet{}
	require.Equal(t, DelegateSetTx, tx.TxType())
}

func TestDelegateSet_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		input    *DelegateSet
		expected FlatTransaction
	}{
		{
			name: "pass - valid DelegateSet",
			input: &DelegateSet{
				BaseTx: BaseTx{
					Account:         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
					TransactionType: DelegateSetTx,
					Fee:             types.XRPCurrencyAmount(12),
					Sequence:        1,
				},
				Authorize: "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
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
			},
			expected: FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "DelegateSet",
				"Fee":             "12",
				"Sequence":        uint32(1),
				"Authorize":       "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
				"Permissions": []any{
					map[string]any{
						"Permission": map[string]any{
							"PermissionValue": "Payment",
						},
					},
					map[string]any{
						"Permission": map[string]any{
							"PermissionValue": "TrustlineAuthorize",
						},
					},
				},
			},
		},
		{
			name: "pass - minimal DelegateSet",
			input: &DelegateSet{
				BaseTx: BaseTx{
					Account:         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
					TransactionType: DelegateSetTx,
				},
				Authorize: "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
				Permissions: []types.Permission{
					{
						Permission: types.PermissionValue{
							PermissionValue: "Payment",
						},
					},
				},
			},
			expected: FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "DelegateSet",
				"Authorize":       "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
				"Permissions": []any{
					map[string]any{
						"Permission": map[string]any{
							"PermissionValue": "Payment",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flattened := tt.input.Flatten()
			require.Equal(t, tt.expected, flattened)
		})
	}
}

func TestDelegateSet_Validate(t *testing.T) {
	tests := []struct {
		name     string
		input    *DelegateSet
		expected bool
	}{
		{
			name: "pass - valid DelegateSet",
			input: &DelegateSet{
				BaseTx: BaseTx{
					Account:         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
					TransactionType: DelegateSetTx,
					Fee:             types.XRPCurrencyAmount(12),
					Sequence:        1,
				},
				Authorize: "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
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
			},
			expected: true,
		},
		{
			name: "pass - valid DelegateSet with single permission",
			input: &DelegateSet{
				BaseTx: BaseTx{
					Account:         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
					TransactionType: DelegateSetTx,
					Fee:             types.XRPCurrencyAmount(12),
					Sequence:        1,
				},
				Authorize: "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
				Permissions: []types.Permission{
					{
						Permission: types.PermissionValue{
							PermissionValue: "Payment",
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "fail - missing Authorize",
			input: &DelegateSet{
				BaseTx: BaseTx{
					Account:         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
					TransactionType: DelegateSetTx,
					Fee:             types.XRPCurrencyAmount(12),
					Sequence:        1,
				},
				Permissions: []types.Permission{
					{
						Permission: types.PermissionValue{
							PermissionValue: "Payment",
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "fail - invalid Authorize address",
			input: &DelegateSet{
				BaseTx: BaseTx{
					Account:         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
					TransactionType: DelegateSetTx,
					Fee:             types.XRPCurrencyAmount(12),
					Sequence:        1,
				},
				Authorize: "invalid_address",
				Permissions: []types.Permission{
					{
						Permission: types.PermissionValue{
							PermissionValue: "Payment",
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "fail - Authorize same as Account",
			input: &DelegateSet{
				BaseTx: BaseTx{
					Account:         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
					TransactionType: DelegateSetTx,
					Fee:             types.XRPCurrencyAmount(12),
					Sequence:        1,
				},
				Authorize: "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				Permissions: []types.Permission{
					{
						Permission: types.PermissionValue{
							PermissionValue: "Payment",
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "fail - empty Permissions array",
			input: &DelegateSet{
				BaseTx: BaseTx{
					Account:         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
					TransactionType: DelegateSetTx,
					Fee:             types.XRPCurrencyAmount(12),
					Sequence:        1,
				},
				Authorize:   "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
				Permissions: []types.Permission{},
			},
			expected: false,
		},
		{
			name: "fail - too many permissions",
			input: &DelegateSet{
				BaseTx: BaseTx{
					Account:         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
					TransactionType: DelegateSetTx,
					Fee:             types.XRPCurrencyAmount(12),
					Sequence:        1,
				},
				Authorize: "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
				Permissions: []types.Permission{
					{Permission: types.PermissionValue{PermissionValue: "Payment"}},
					{Permission: types.PermissionValue{PermissionValue: "TrustSet"}},
					{Permission: types.PermissionValue{PermissionValue: "OfferCreate"}},
					{Permission: types.PermissionValue{PermissionValue: "OfferCancel"}},
					{Permission: types.PermissionValue{PermissionValue: "EscrowCreate"}},
					{Permission: types.PermissionValue{PermissionValue: "EscrowFinish"}},
					{Permission: types.PermissionValue{PermissionValue: "EscrowCancel"}},
					{Permission: types.PermissionValue{PermissionValue: "PaymentChannelCreate"}},
					{Permission: types.PermissionValue{PermissionValue: "PaymentChannelFund"}},
					{Permission: types.PermissionValue{PermissionValue: "PaymentChannelClaim"}},
					{Permission: types.PermissionValue{PermissionValue: "CheckCreate"}}, // 11th permission, exceeds max
				},
			},
			expected: false,
		},
		{
			name: "fail - empty permission value",
			input: &DelegateSet{
				BaseTx: BaseTx{
					Account:         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
					TransactionType: DelegateSetTx,
					Fee:             types.XRPCurrencyAmount(12),
					Sequence:        1,
				},
				Authorize: "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
				Permissions: []types.Permission{
					{
						Permission: types.PermissionValue{
							PermissionValue: "",
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "fail - non-delegatable transaction type",
			input: &DelegateSet{
				BaseTx: BaseTx{
					Account:         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
					TransactionType: DelegateSetTx,
					Fee:             types.XRPCurrencyAmount(12),
					Sequence:        1,
				},
				Authorize: "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
				Permissions: []types.Permission{
					{
						Permission: types.PermissionValue{
							PermissionValue: "AccountSet",
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "fail - duplicate permissions",
			input: &DelegateSet{
				BaseTx: BaseTx{
					Account:         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
					TransactionType: DelegateSetTx,
					Fee:             types.XRPCurrencyAmount(12),
					Sequence:        1,
				},
				Authorize: "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
				Permissions: []types.Permission{
					{
						Permission: types.PermissionValue{
							PermissionValue: "Payment",
						},
					},
					{
						Permission: types.PermissionValue{
							PermissionValue: "Payment",
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "fail - invalid BaseTx",
			input: &DelegateSet{
				BaseTx: BaseTx{
					Account:         "invalid_account",
					TransactionType: DelegateSetTx,
					Fee:             types.XRPCurrencyAmount(12),
					Sequence:        1,
				},
				Authorize: "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
				Permissions: []types.Permission{
					{
						Permission: types.PermissionValue{
							PermissionValue: "Payment",
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.input.Validate()
			require.Equal(t, tt.expected, valid)
			if tt.expected {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
