package client

import (
	"testing"

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
		name                string
		tx                  map[string]any
		expected            map[string]any
		expectedErr         error
		expectedTagMismatch bool
	}{
		{
			name:     "classic address is unchanged",
			tx:       map[string]any{"Account": classic},
			expected: map[string]any{"Account": classic},
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
			expectedTagMismatch: true,
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
			expectedTagMismatch: true,
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
			name: "additional address fields discard embedded tags",
			tx: map[string]any{
				"Authorize":   mainnetNoTag,
				"Unauthorize": testnetTag14,
				"Owner":       mainnetTagOne,
				"RegularKey":  mainnetTagZero,
			},
			expected: map[string]any{
				"Authorize":   classic,
				"Unauthorize": classic,
				"Owner":       classic,
				"RegularKey":  pepperClassic,
			},
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
			switch {
			case tt.expectedTagMismatch:
				var mismatch ErrMismatchedTag
				require.ErrorAs(t, err, &mismatch)
			case tt.expectedErr != nil:
				require.ErrorIs(t, err, tt.expectedErr)
			default:
				require.NoError(t, err)
			}
			require.Equal(t, tt.expected, tt.tx)
		})
	}
}
