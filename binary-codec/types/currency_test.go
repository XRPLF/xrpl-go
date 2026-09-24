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
		expected any
		opts     []int
		err      error
		setup    func(t *testing.T) (*Currency, *testutil.MockBinaryParser)
	}{
		{
			name:     "pass - XRP currency",
			expected: "XRP",
			opts:     []int{20},
			err:      nil,
			setup: func(t *testing.T) (*Currency, *testutil.MockBinaryParser) {
				ctrl := gomock.NewController(t)
				mock := testutil.NewMockBinaryParser(ctrl)
				mock.EXPECT().ReadBytes(20).Return(XRPBytes, nil)
				return &Currency{}, mock
			},
		},
		{
			name:     "pass - 3 letter currency code",
			expected: "USD",
			opts:     []int{20},
			err:      nil,
			setup: func(t *testing.T) (*Currency, *testutil.MockBinaryParser) {
				ctrl := gomock.NewController(t)
				mock := testutil.NewMockBinaryParser(ctrl)
				mock.EXPECT().ReadBytes(20).Return([]byte{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x55, 0x53, 0x44, 0x00, 0x00, 0x00, 0x00, 0x00,
				}, nil)
				return &Currency{}, mock
			},
		},
		{
			name:     "pass - hex currency code",
			expected: "0102030405060708090a0b0c0d0e0f1011121314",
			opts:     []int{20},
			err:      nil,
			setup: func(t *testing.T) (*Currency, *testutil.MockBinaryParser) {
				ctrl := gomock.NewController(t)
				mock := testutil.NewMockBinaryParser(ctrl)
				mock.EXPECT().ReadBytes(20).Return([]byte{
					0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a,
					0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14,
				}, nil)
				return &Currency{}, mock
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			currency, parser := tc.setup(t)
			actual, err := currency.ToJSON(parser, tc.opts...)

			if tc.err != nil {
				require.Error(t, err)
				require.Equal(t, tc.err, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, actual)
			}
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
			name:     "fail - 3 letter code outside the IOU alphabet",
			input:    "AD/",
			expected: nil,
			err:      errInvalidCurrencyCode,
		},
		{
			name:     "fail - 0x prefix is not a currency code",
			input:    "0x0102030405060708090A0B0C0D0E0F1011121314",
			expected: nil,
			err:      &InvalidCodeError{"0x0102030405060708090A0B0C0D0E0F1011121314"},
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
