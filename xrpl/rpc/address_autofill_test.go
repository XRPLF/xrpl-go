package rpc

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/stretchr/testify/require"
)

// Exhaustive X-address and tag policy cases live in xrpl/internal/client.
// These cases only verify the client wiring: delegation through
// FlatTransaction and the re-exported error identities.
func TestClientSetValidTransactionAddressesXAddress(t *testing.T) {
	const (
		classic         = "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59"
		testnetTag14    = "T719a5UwUCnEs54UsxG9CJYYDhwmFCvqXVCALUGJGSbNV3x"
		invalidXAddress = "XVLhHMPHU98es4dbozjVtdWzVrDjtV5fdx1mHp98tDMoQXa"
	)

	cl := &Client{}
	tests := []struct {
		name        string
		tx          transaction.FlatTransaction
		expected    transaction.FlatTransaction
		expectedErr error
	}{
		{
			name: "converts X-address and applies embedded tag",
			tx:   transaction.FlatTransaction{"Destination": testnetTag14},
			expected: transaction.FlatTransaction{
				"Destination":    classic,
				"DestinationTag": uint32(14),
			},
		},
		{
			name: "explicit tag conflict",
			tx: transaction.FlatTransaction{
				"Destination":    testnetTag14,
				"DestinationTag": uint32(13),
			},
			expected: transaction.FlatTransaction{
				"Destination":    testnetTag14,
				"DestinationTag": uint32(13),
			},
			expectedErr: ErrMismatchedTag,
		},
		{
			name:        "invalid X-address",
			tx:          transaction.FlatTransaction{"Account": invalidXAddress},
			expected:    transaction.FlatTransaction{"Account": invalidXAddress},
			expectedErr: ErrInvalidAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cl.setValidTransactionAddresses(&tt.tx)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expected, tt.tx)
		})
	}
}
