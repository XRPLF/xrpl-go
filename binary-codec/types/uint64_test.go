package types

import (
	"errors"
	"testing"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/serdes"
	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
	"github.com/Peersyst/xrpl-go/binary-codec/types/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestUint64_FromJson(t *testing.T) {
	tt := []struct {
		name        string
		input       any
		expected    []byte
		expectedErr error
	}{
		{
			name:        "fail - value is not a string (int)",
			input:       1,
			expected:    nil,
			expectedErr: ErrInvalidUInt64String,
		},
		{
			name:        "fail - value is not a string (nil)",
			input:       nil,
			expected:    nil,
			expectedErr: ErrInvalidUInt64String,
		},
		{
			name:        "fail - value is not a string (bool)",
			input:       true,
			expected:    nil,
			expectedErr: ErrInvalidUInt64String,
		},
		{
			name:        "fail - invalid hex string",
			input:       "invalid",
			expected:    nil,
			expectedErr: ErrInvalidUInt64String,
		},
		{
			name:        "fail - empty hex string",
			input:       "",
			expected:    nil,
			expectedErr: ErrInvalidUInt64String,
		},
		{
			name:        "fail - 17-character hex string",
			input:       "00000000000000000",
			expected:    nil,
			expectedErr: ErrInvalidUInt64String,
		},
		{
			name:        "fail - smallest value above max uint64 (1 followed by 16 zeros)",
			input:       "10000000000000000",
			expected:    nil,
			expectedErr: ErrInvalidUInt64String,
		},
		{
			name:        "fail - 1 followed by 16 Fs",
			input:       "1FFFFFFFFFFFFFFFF",
			expected:    nil,
			expectedErr: ErrInvalidUInt64String,
		},
		{
			name:        "fail - decimal max uint64 exceeds hex string length",
			input:       "18446744073709551615",
			expected:    nil,
			expectedErr: ErrInvalidUInt64String,
		},
		{
			name:        "pass - valid uint64 hex string",
			input:       "1",
			expected:    []byte{0, 0, 0, 0, 0, 0, 0, 1},
			expectedErr: nil,
		},
		{
			name:        "pass - valid uint64 hex string (2)",
			input:       "100",
			expected:    []byte{0, 0, 0, 0, 0, 0, 1, 0},
			expectedErr: nil,
		},
		{
			name:        "pass - numeric-looking string is treated as hex",
			input:       "1000",
			expected:    []byte{0, 0, 0, 0, 0, 0, 16, 0},
			expectedErr: nil,
		},
		{
			name:        "pass - valid lowercase uint64 hex string",
			input:       "abcdef",
			expected:    []byte{0, 0, 0, 0, 0, 171, 205, 239},
			expectedErr: nil,
		},
		{
			name:        "pass - valid uint64 hex string (large number)",
			input:       "FFFFFFFFFFFFFFFF",
			expected:    []byte{255, 255, 255, 255, 255, 255, 255, 255},
			expectedErr: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			class := &UInt64{}
			actual, err := class.FromJSON(tc.input)
			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestUint64_ToJson(t *testing.T) {
	defs := definitions.Get()

	tt := []struct {
		name        string
		input       []byte
		malleate    func(t *testing.T) interfaces.BinaryParser
		expected    string
		expectedErr error
	}{
		{
			name:  "fail - binary parser has no data",
			input: []byte{},
			malleate: func(t *testing.T) interfaces.BinaryParser {
				parserMock := testutil.NewMockBinaryParser(gomock.NewController(t))
				parserMock.EXPECT().ReadBytes(gomock.Any()).Return([]byte{}, errors.New("binary parser has no data"))
				return parserMock
			},
			expected:    "",
			expectedErr: errors.New("binary parser has no data"),
		},
		{
			name:     "pass - valid uint64",
			input:    []byte{0, 0, 0, 0, 0, 0, 0, 1},
			expected: "0000000000000001",
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{0, 0, 0, 0, 0, 0, 0, 1}, defs)
			},
			expectedErr: nil,
		},
		{
			name:        "pass - valid uint64 (2)",
			input:       []byte{0, 0, 0, 0, 0, 0, 0, 100},
			expected:    "0000000000000064",
			expectedErr: nil,
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{0, 0, 0, 0, 0, 0, 0, 100}, defs)
			},
		},
		{
			name:        "pass - valid uint64 (3)",
			input:       []byte{0, 0, 0, 0, 0, 0, 0, 255},
			expected:    "00000000000000FF",
			expectedErr: nil,
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{0, 0, 0, 0, 0, 0, 0, 255}, defs)
			},
		},
		{
			name:        "pass - valid uint64 (large number)",
			input:       []byte{255, 255, 255, 255, 255, 255, 255, 255},
			expected:    "FFFFFFFFFFFFFFFF", // Max uint64 value
			expectedErr: nil,
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{255, 255, 255, 255, 255, 255, 255, 255}, defs)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			class := &UInt64{}
			parser := tc.malleate(t)
			actual, err := class.ToJSON(parser)
			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestUint64_RoundTrip(t *testing.T) {
	defs := definitions.Get()

	tt := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "lowercase short hex is zero-padded and upper-cased",
			input: "abcdef",
			want:  "0000000000ABCDEF",
		},
		{
			name:  "single digit is zero-padded",
			input: "1",
			want:  "0000000000000001",
		},
		{
			name:  "max value is preserved",
			input: "FFFFFFFFFFFFFFFF",
			want:  "FFFFFFFFFFFFFFFF",
		},
		{
			name:  "mixed case is normalized to upper",
			input: "AbCdEf",
			want:  "0000000000ABCDEF",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			class := &UInt64{}
			encoded, err := class.FromJSON(tc.input)
			require.NoError(t, err)

			decoded, err := class.ToJSON(serdes.NewBinaryParser(encoded, defs))
			require.NoError(t, err)
			require.Equal(t, tc.want, decoded)
		})
	}
}
