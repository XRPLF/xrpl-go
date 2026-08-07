package client

import (
	"testing"

	transactiontypes "github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestSetValidAddresses(t *testing.T) {
	const (
		classic         = "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59"
		mainnetNoTag    = "X7AcgcsBL6XDcUb289X4mJ8djcdyKaB5hJDWMArnXr61cqZ"
		mainnetTagOne   = "X7AcgcsBL6XDcUb289X4mJ8djcdyKaGZMhc9YTE92ehJ2Fu"
		testnetTag14    = "T719a5UwUCnEs54UsxG9CJYYDhwmFCvqXVCALUGJGSbNV3x"
		pepperClassic   = "rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY"
		mainnetTagZero  = "XV5sbjUmgPpvXv4ixFWZ5ptAYZ6PD2m4Er6SnvjVLpMWPjR"
		invalidXAddress = "XVLhHMPHU98es4dbozjVtdWzVrDjtV5fdx1mHp98tDMoQXa"
	)

	tests := []struct {
		name        string
		tx          map[string]any
		expected    map[string]any
		expectedErr error
	}{
		{
			name:     "classic address is unchanged",
			tx:       map[string]any{"Account": classic},
			expected: map[string]any{"Account": classic},
		},
		{
			name: "named classic addresses normalize to plain strings",
			tx: map[string]any{
				"Account":         transactiontypes.Address(classic),
				"Destination":     transactiontypes.Address(classic),
				"Authorize":       transactiontypes.Address(classic),
				"Unauthorize":     transactiontypes.Address(classic),
				"Owner":           transactiontypes.Address(classic),
				"RegularKey":      transactiontypes.Address(classic),
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{"RawTransaction": map[string]any{"Account": transactiontypes.Address(classic)}},
				},
			},
			expected: map[string]any{
				"Account":         classic,
				"Destination":     classic,
				"Authorize":       classic,
				"Unauthorize":     classic,
				"Owner":           classic,
				"RegularKey":      classic,
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{"RawTransaction": map[string]any{"Account": classic}},
				},
			},
		},
		{
			name:        "invalid X-address is rejected",
			tx:          map[string]any{"Account": invalidXAddress},
			expected:    map[string]any{"Account": invalidXAddress},
			expectedErr: ErrInvalidAddress,
		},
		{
			name:     "mainnet no-tag Account",
			tx:       map[string]any{"Account": mainnetNoTag},
			expected: map[string]any{"Account": classic},
		},
		{
			name: "testnet nonzero Destination tag",
			tx: map[string]any{
				"Destination": testnetTag14,
			},
			expected: map[string]any{
				"Destination":    classic,
				"DestinationTag": uint32(14),
			},
		},
		{
			name: "embedded zero Source tag preserves presence",
			tx: map[string]any{
				"Account": mainnetTagZero,
			},
			expected: map[string]any{
				"Account":   pepperClassic,
				"SourceTag": uint32(0),
			},
		},
		{
			name: "matching explicit tag",
			tx: map[string]any{
				"Destination":    mainnetTagOne,
				"DestinationTag": uint32(1),
			},
			expected: map[string]any{
				"Destination":    classic,
				"DestinationTag": uint32(1),
			},
		},
		{
			name: "embedded nonzero conflicts with explicit zero",
			tx: map[string]any{
				"Destination":    mainnetTagOne,
				"DestinationTag": uint32(0),
			},
			expected: map[string]any{
				"Destination":    mainnetTagOne,
				"DestinationTag": uint32(0),
			},
			expectedErr: ErrMismatchedTag,
		},
		{
			name: "embedded zero conflicts with explicit nonzero",
			tx: map[string]any{
				"Account":   mainnetTagZero,
				"SourceTag": uint32(1),
			},
			expected: map[string]any{
				"Account":   mainnetTagZero,
				"SourceTag": uint32(1),
			},
			expectedErr: ErrMismatchedTag,
		},
		{
			name: "invalid explicit tag type",
			tx: map[string]any{
				"Destination":    mainnetTagOne,
				"DestinationTag": float64(1),
			},
			expected: map[string]any{
				"Destination":    mainnetTagOne,
				"DestinationTag": float64(1),
			},
			expectedErr: ErrTagFieldIsNotAUint32,
		},
		{
			name: "additional untagged address fields convert",
			tx: map[string]any{
				"Authorize":   mainnetNoTag,
				"Unauthorize": mainnetNoTag,
				"Owner":       mainnetNoTag,
				"RegularKey":  mainnetNoTag,
			},
			expected: map[string]any{
				"Authorize":   classic,
				"Unauthorize": classic,
				"Owner":       classic,
				"RegularKey":  classic,
			},
		},
		{
			name:        "tagged Authorize is rejected",
			tx:          map[string]any{"Authorize": mainnetTagOne},
			expected:    map[string]any{"Authorize": mainnetTagOne},
			expectedErr: ErrAccountIDTagNotAllowed,
		},
		{
			name:        "tagged Unauthorize is rejected",
			tx:          map[string]any{"Unauthorize": testnetTag14},
			expected:    map[string]any{"Unauthorize": testnetTag14},
			expectedErr: ErrAccountIDTagNotAllowed,
		},
		{
			name:        "tagged Owner is rejected",
			tx:          map[string]any{"Owner": mainnetTagOne},
			expected:    map[string]any{"Owner": mainnetTagOne},
			expectedErr: ErrAccountIDTagNotAllowed,
		},
		{
			name:        "tagged RegularKey is rejected",
			tx:          map[string]any{"RegularKey": mainnetTagZero},
			expected:    map[string]any{"RegularKey": mainnetTagZero},
			expectedErr: ErrAccountIDTagNotAllowed,
		},
		{
			name: "unsupported tag leaves earlier address unchanged",
			tx: map[string]any{
				"Account": mainnetNoTag,
				"Owner":   mainnetTagOne,
			},
			expected: map[string]any{
				"Account": mainnetNoTag,
				"Owner":   mainnetTagOne,
			},
			expectedErr: ErrAccountIDTagNotAllowed,
		},
		{
			name: "later error leaves named classic address unchanged",
			tx: map[string]any{
				"Account": transactiontypes.Address(classic),
				"Owner":   mainnetTagOne,
			},
			expected: map[string]any{
				"Account": transactiontypes.Address(classic),
				"Owner":   mainnetTagOne,
			},
			expectedErr: ErrAccountIDTagNotAllowed,
		},
		{
			name: "tagged Batch inner field is rejected",
			tx: map[string]any{
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{"RawTransaction": map[string]any{"Owner": testnetTag14}},
				},
			},
			expected: map[string]any{
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{"RawTransaction": map[string]any{"Owner": testnetTag14}},
				},
			},
			expectedErr: ErrAccountIDTagNotAllowed,
		},
		{
			name: "Batch inner addresses and tags",
			tx: map[string]any{
				"Account":         mainnetNoTag,
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{"RawTransaction": map[string]any{
						"Account":     mainnetTagZero,
						"Destination": testnetTag14,
					}},
				},
			},
			expected: map[string]any{
				"Account":         classic,
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{"RawTransaction": map[string]any{
						"Account":        pepperClassic,
						"SourceTag":      uint32(0),
						"Destination":    classic,
						"DestinationTag": uint32(14),
					}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetValidAddresses(tt.tx)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expected, tt.tx)
		})
	}
}
