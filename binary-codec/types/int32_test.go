package types

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/serdes"
	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
	"github.com/Peersyst/xrpl-go/binary-codec/types/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestInt32_FromJSON(t *testing.T) {
	tt := []struct {
		name        string
		input       any
		expected    []byte
		expectedErr error
	}{
		{
			name:        "Zero value",
			input:       int32(0),
			expected:    []byte{0, 0, 0, 0},
			expectedErr: nil,
		},
		{
			name:        "Positive int32 - small",
			input:       int32(1),
			expected:    []byte{0, 0, 0, 1},
			expectedErr: nil,
		},
		{
			name:        "Positive int32 - medium",
			input:       int32(100),
			expected:    []byte{0, 0, 0, 100},
			expectedErr: nil,
		},
		{
			name:        "Positive int32 - large",
			input:       int32(255),
			expected:    []byte{0, 0, 0, 255},
			expectedErr: nil,
		},
		{
			name:        "Positive int32 - 1000",
			input:       int32(1000),
			expected:    []byte{0, 0, 3, 232},
			expectedErr: nil,
		},
		{
			name:        "Negative int32 - small",
			input:       int32(-1),
			expected:    []byte{0xff, 0xff, 0xff, 0xff}, // Two's complement of -1
			expectedErr: nil,
		},
		{
			name:        "Negative int32 - medium",
			input:       int32(-100),
			expected:    []byte{0xff, 0xff, 0xff, 0x9c}, // Two's complement of -100
			expectedErr: nil,
		},
		{
			name:        "Negative int32 - large",
			input:       int32(-1000),
			expected:    []byte{0xff, 0xff, 0xfc, 0x18}, // Two's complement of -1000
			expectedErr: nil,
		},
		{
			name:        "Maximum positive int32",
			input:       int32(math.MaxInt32),
			expected:    []byte{0x7f, 0xff, 0xff, 0xff}, // 2147483647
			expectedErr: nil,
		},
		{
			name:        "Minimum negative int32",
			input:       int32(math.MinInt32),
			expected:    []byte{0x80, 0x00, 0x00, 0x00}, // -2147483648
			expectedErr: nil,
		},
		{
			name:        "From int type - positive",
			input:       int(42),
			expected:    []byte{0, 0, 0, 42},
			expectedErr: nil,
		},
		{
			name:        "From int type - negative",
			input:       int(-42),
			expected:    []byte{0xff, 0xff, 0xff, 0xd6}, // Two's complement of -42
			expectedErr: nil,
		},
		{
			name:        "From int64 type - positive",
			input:       int64(1234),
			expected:    []byte{0, 0, 4, 210},
			expectedErr: nil,
		},
		{
			name:        "From int64 type - negative",
			input:       int64(-1234),
			expected:    []byte{0xff, 0xff, 0xfb, 0x2e}, // Two's complement of -1234
			expectedErr: nil,
		},
		{
			name:        "From uint32 type - within range",
			input:       uint32(1000),
			expected:    []byte{0, 0, 3, 232},
			expectedErr: nil,
		},
		{
			name:        "From float64 type - positive whole number",
			input:       float64(500),
			expected:    []byte{0, 0, 1, 244},
			expectedErr: nil,
		},
		{
			name:        "From float64 type - negative whole number",
			input:       float64(-500),
			expected:    []byte{0xff, 0xff, 0xfe, 0x0c}, // Two's complement of -500
			expectedErr: nil,
		},
		{
			name:        "Error - int64 overflow (too large)",
			input:       int64(math.MaxInt32) + 1,
			expected:    nil,
			expectedErr: ErrInt32OutOfRange,
		},
		{
			name:        "Error - int64 underflow (too small)",
			input:       int64(math.MinInt32) - 1,
			expected:    nil,
			expectedErr: ErrInt32OutOfRange,
		},
		{
			name:        "Error - uint32 overflow",
			input:       uint32(math.MaxInt32) + 1,
			expected:    nil,
			expectedErr: ErrInt32OutOfRange,
		},
		{
			name:        "Error - uint64 overflow",
			input:       uint64(math.MaxInt32) + 1,
			expected:    nil,
			expectedErr: ErrInt32OutOfRange,
		},
		{
			name:        "Error - float64 not a whole number",
			input:       float64(123.45),
			expected:    nil,
			expectedErr: ErrInt32OutOfRange,
		},
		{
			name:        "Error - float64 overflow",
			input:       float64(math.MaxInt32) + 1000,
			expected:    nil,
			expectedErr: ErrInt32OutOfRange,
		},
		{
			name:        "Error - float64 underflow",
			input:       float64(math.MinInt32) - 1000,
			expected:    nil,
			expectedErr: ErrInt32OutOfRange,
		},
		{
			name:        "Error - float64 far overflow (1e30)",
			input:       float64(1e30),
			expected:    nil,
			expectedErr: ErrInt32OutOfRange,
		},
		{
			name:        "Error - float64 far underflow (-1e30)",
			input:       float64(-1e30),
			expected:    nil,
			expectedErr: ErrInt32OutOfRange,
		},
		{
			name:        "Error - float64 positive infinity",
			input:       math.Inf(1),
			expected:    nil,
			expectedErr: ErrInt32OutOfRange,
		},
		{
			name:        "Error - float64 negative infinity",
			input:       math.Inf(-1),
			expected:    nil,
			expectedErr: ErrInt32OutOfRange,
		},
		{
			name:        "Error - float64 NaN",
			input:       math.NaN(),
			expected:    nil,
			expectedErr: ErrInt32OutOfRange,
		},
		{
			name:        "Error - invalid type (string)",
			input:       "not a number",
			expected:    nil,
			expectedErr: ErrInt32OutOfRange,
		},
		{
			name:        "Error - invalid type (bool)",
			input:       true,
			expected:    nil,
			expectedErr: ErrInt32OutOfRange,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			int32Type := &Int32{}
			actual, err := int32Type.FromJSON(tc.input)

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

func TestInt32_ToJSON(t *testing.T) {
	defs := definitions.Get()

	tt := []struct {
		name        string
		input       []byte
		malleate    func(t *testing.T) interfaces.BinaryParser
		expected    int32
		expectedErr error
	}{
		{
			name:  "Zero value",
			input: []byte{0, 0, 0, 0},
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{0, 0, 0, 0}, defs)
			},
			expected:    0,
			expectedErr: nil,
		},
		{
			name:  "Positive int32 - small",
			input: []byte{0, 0, 0, 1},
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{0, 0, 0, 1}, defs)
			},
			expected:    1,
			expectedErr: nil,
		},
		{
			name:  "Positive int32 - medium",
			input: []byte{0, 0, 0, 100},
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{0, 0, 0, 100}, defs)
			},
			expected:    100,
			expectedErr: nil,
		},
		{
			name:  "Positive int32 - large",
			input: []byte{0, 0, 0, 255},
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{0, 0, 0, 255}, defs)
			},
			expected:    255,
			expectedErr: nil,
		},
		{
			name:  "Positive int32 - 1000",
			input: []byte{0, 0, 3, 232},
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{0, 0, 3, 232}, defs)
			},
			expected:    1000,
			expectedErr: nil,
		},
		{
			name:  "Negative int32 - small",
			input: []byte{0xff, 0xff, 0xff, 0xff},
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{0xff, 0xff, 0xff, 0xff}, defs)
			},
			expected:    -1,
			expectedErr: nil,
		},
		{
			name:  "Negative int32 - medium",
			input: []byte{0xff, 0xff, 0xff, 0x9c},
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{0xff, 0xff, 0xff, 0x9c}, defs)
			},
			expected:    -100,
			expectedErr: nil,
		},
		{
			name:  "Negative int32 - large",
			input: []byte{0xff, 0xff, 0xfc, 0x18},
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{0xff, 0xff, 0xfc, 0x18}, defs)
			},
			expected:    -1000,
			expectedErr: nil,
		},
		{
			name:  "Maximum positive int32",
			input: []byte{0x7f, 0xff, 0xff, 0xff},
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{0x7f, 0xff, 0xff, 0xff}, defs)
			},
			expected:    math.MaxInt32,
			expectedErr: nil,
		},
		{
			name:  "Minimum negative int32",
			input: []byte{0x80, 0x00, 0x00, 0x00},
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{0x80, 0x00, 0x00, 0x00}, defs)
			},
			expected:    math.MinInt32,
			expectedErr: nil,
		},
		{
			name:  "Error - parser has no data",
			input: []byte{0, 0, 0, 1},
			malleate: func(t *testing.T) interfaces.BinaryParser {
				parserMock := testutil.NewMockBinaryParser(gomock.NewController(t))
				parserMock.EXPECT().ReadBytes(gomock.Any()).Return([]byte{}, errors.New("binary parser has no data"))
				return parserMock
			},
			expected:    0,
			expectedErr: fmt.Errorf("binary parser has no data"),
		},
		{
			name:  "Error - parser has insufficient data",
			input: []byte{0, 0, 0, 1},
			malleate: func(t *testing.T) interfaces.BinaryParser {
				parserMock := testutil.NewMockBinaryParser(gomock.NewController(t))
				parserMock.EXPECT().ReadBytes(gomock.Any()).Return([]byte{}, errors.New("insufficient data"))
				return parserMock
			},
			expected:    0,
			expectedErr: fmt.Errorf("insufficient data"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			int32Type := &Int32{}
			parser := tc.malleate(t)
			actual, err := int32Type.ToJSON(parser)

			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestInt32_RoundTrip(t *testing.T) {
	// Test that FromJSON -> ToJSON produces the same value
	testCases := []int32{
		0,
		1,
		-1,
		100,
		-100,
		1000,
		-1000,
		math.MaxInt32,
		math.MinInt32,
		42,
		-42,
		12345,
		-12345,
	}

	defs := definitions.Get()

	for _, original := range testCases {
		t.Run(fmt.Sprintf("RoundTrip_%d", original), func(t *testing.T) {
			int32Type := &Int32{}

			// Convert to bytes
			bytes, err := int32Type.FromJSON(original)
			require.NoError(t, err)
			require.NotNil(t, bytes)
			require.Len(t, bytes, 4)

			// Convert back to value
			parser := serdes.NewBinaryParser(bytes, defs)
			result, err := int32Type.ToJSON(parser)
			require.NoError(t, err)

			// Verify the value matches
			require.Equal(t, original, result)
		})
	}
}
