package types

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/serdes"
	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
	"github.com/Peersyst/xrpl-go/binary-codec/types/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestUint32_FromJson(t *testing.T) {
	tt := []struct {
		name        string
		input       any
		expected    []byte
		expectedErr error
	}{
		{
			name:        "Valid uint32",
			input:       uint32(1),
			expected:    []byte{0, 0, 0, 1},
			expectedErr: nil,
		},
		{
			name:        "Valid uint32 (2)",
			input:       uint32(100),
			expected:    []byte{0, 0, 0, 100},
			expectedErr: nil,
		},
		{
			name:        "Valid uint32 (3)",
			input:       uint32(255),
			expected:    []byte{0, 0, 0, 255},
			expectedErr: nil,
		},
		{
			name:        "Valid uint32 - max value",
			input:       uint32(4294967295),
			expected:    []byte{0xFF, 0xFF, 0xFF, 0xFF},
			expectedErr: nil,
		},
		{
			name:        "Valid uint32 - from int",
			input:       int(1000),
			expected:    []byte{0, 0, 3, 232},
			expectedErr: nil,
		},
		{
			name:        "Valid uint32 - from int64",
			input:       int64(5000),
			expected:    []byte{0, 0, 19, 136},
			expectedErr: nil,
		},
		{
			name:        "Valid uint32 - from uint64",
			input:       uint64(10000),
			expected:    []byte{0, 0, 39, 16},
			expectedErr: nil,
		},
		{
			name:        "Valid uint32 - from float64",
			input:       float64(500),
			expected:    []byte{0, 0, 1, 244},
			expectedErr: nil,
		},
		{
			name:        "Error - negative int",
			input:       int(-1),
			expected:    nil,
			expectedErr: ErrUInt32OutOfRange,
		},
		{
			name:        "Error - negative int64",
			input:       int64(-100),
			expected:    nil,
			expectedErr: ErrUInt32OutOfRange,
		},
		{
			name:        "Error - int64 overflow",
			input:       int64(4294967296), // MaxUint32 + 1
			expected:    nil,
			expectedErr: ErrUInt32OutOfRange,
		},
		{
			name:        "Error - uint64 overflow",
			input:       uint64(4294967296), // MaxUint32 + 1
			expected:    nil,
			expectedErr: ErrUInt32OutOfRange,
		},
		{
			name:        "Error - float64 not a whole number",
			input:       float64(123.45),
			expected:    nil,
			expectedErr: ErrUInt32OutOfRange,
		},
		{
			name:        "Error - float64 overflow",
			input:       float64(4294967296), // MaxUint32 + 1
			expected:    nil,
			expectedErr: ErrUInt32OutOfRange,
		},
		{
			name:        "Error - negative float64",
			input:       float64(-50),
			expected:    nil,
			expectedErr: ErrUInt32OutOfRange,
		},
		{
			name:        "Error - invalid type (string)",
			input:       "not a number",
			expected:    nil,
			expectedErr: ErrUInt32OutOfRange,
		},
		{
			name:        "Error - invalid type (bool)",
			input:       true,
			expected:    nil,
			expectedErr: ErrUInt32OutOfRange,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			uint32Type := &UInt32{}
			actual, err := uint32Type.FromJSON(tc.input)

			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
				require.Nil(t, actual)
			} else {
				require.NoError(t, err)
				require.True(t, bytes.Equal(actual, tc.expected),
					"Expected %v, got %v", tc.expected, actual)
			}
		})
	}
}

func TestUint32_ToJson(t *testing.T) {
	defs := definitions.Get()

	tt := []struct {
		name        string
		malleate    func(t *testing.T) interfaces.BinaryParser
		expected    uint32
		expectedErr error
	}{
		{
			name: "fail - invalid uint32",
			malleate: func(t *testing.T) interfaces.BinaryParser {
				parserMock := testutil.NewMockBinaryParser(gomock.NewController(t))
				parserMock.EXPECT().ReadBytes(gomock.Any()).Return([]byte{}, errors.New("binary parser has no data"))
				return parserMock
			},
			expected:    0,
			expectedErr: fmt.Errorf("binary parser has no data"),
		},
		{
			name: "pass - valid uint32",
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{0, 0, 0, 1}, defs)
			},
			expected:    1,
			expectedErr: nil,
		},
		{
			name: "pass - valid uint32 (2)",
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{0, 0, 0, 100}, defs)
			},
			expected:    100,
			expectedErr: nil,
		},
		{
			name: "pass - valid uint32 (3)",
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{0, 0, 0, 255}, defs)
			},
			expected:    255,
			expectedErr: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			class := &UInt32{}
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
