package types

import (
	"testing"

	"github.com/Peersyst/xrpl-go/binary-codec/types/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestCurrency_ToJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "pass - XRP currency",
			input:    XRPBytes,
			expected: "XRP",
		},
		{
			name: "pass - 3 letter currency code",
			input: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x55, 0x53, 0x44, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			expected: "USD",
		},
		{
			name: "pass - hex currency code",
			input: []byte{
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a,
				0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14,
			},
			expected: "0102030405060708090A0B0C0D0E0F1011121314",
		},
		{
			name: "pass - non-standard currency with three non-ISO bytes",
			input: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			expected: "000000000000000000000000FFFFFF0000000000",
		},
		{
			name: "pass - non-standard currency with the XRP code bytes",
			input: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 'X', 'R', 'P', 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			expected: "0000000000000000000000005852500000000000",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parser := testutil.NewMockBinaryParser(gomock.NewController(t))
			parser.EXPECT().ReadBytes(20).Return(tc.input, nil)

			actual, err := (&Currency{}).ToJSON(parser)

			require.NoError(t, err)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestCurrency_FromJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected []byte
		err      error
	}{
		{
			name:  "pass - XRP currency",
			input: "XRP",
			expected: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			err: nil,
		},
		{
			name:  "pass - 3 letter currency code",
			input: "USD",
			expected: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x55, 0x53, 0x44, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			err: nil,
		},
		{
			name:  "pass - hex currency code",
			input: "0102030405060708090A0B0C0D0E0F1011121314",
			expected: []byte{
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a,
				0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14,
			},
			err: nil,
		},
		{
			name:     "fail - invalid currency",
			input:    123,
			expected: nil,
			err:      ErrInvalidCurrency,
		},
		{
			name:     "fail - hex currency shorter than 20 bytes",
			input:    "ABCD",
			expected: nil,
			err:      ErrInvalidCurrency,
		},
		{
			name:     "fail - hex currency longer than 20 bytes",
			input:    "0102030405060708090A0B0C0D0E0F101112131415",
			expected: nil,
			err:      ErrInvalidCurrency,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			currency := &Currency{}
			actual, err := currency.FromJSON(tc.input)

			if tc.err != nil {
				require.Error(t, err)
				require.Equal(t, tc.err, err)
				require.Nil(t, actual)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}
