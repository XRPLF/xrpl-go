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
			name:        "fail - value is not a string",
			input:       1,
			expected:    nil,
			expectedErr: ErrInvalidUInt64String,
		},
		{
			name:        "fail - invalid hex string",
			input:       "invalid",
			expected:    nil,
			expectedErr: errors.New("encoding/hex: invalid byte: U+0049 'I'"),
		},
		{
			name:        "pass - valid uint64 numeric string",
			input:       "1",
			expected:    []byte{0, 0, 0, 0, 0, 0, 0, 1},
			expectedErr: nil,
		},
		{
			name:        "pass - valid uint64 numeric string (2)",
			input:       "100",
			expected:    []byte{0, 0, 0, 0, 0, 0, 1, 0},
			expectedErr: nil,
		},
		{
			name:        "pass - valid uint64 numeric string (3)",
			input:       "255",
			expected:    []byte{0, 0, 0, 0, 0, 0, 2, 85},
			expectedErr: nil,
		},
		{
			name:        "pass - valid uint64 non-numeric string (large number)",
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
