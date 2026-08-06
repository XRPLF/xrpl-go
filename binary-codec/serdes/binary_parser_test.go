package serdes

import (
	"errors"
	"testing"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/serdes/interfaces"
	"github.com/Peersyst/xrpl-go/binary-codec/serdes/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestBinaryParser_ReadVariableLength(t *testing.T) {
	tt := []struct {
		name        string
		input       []byte
		output      int
		expectedErr error
	}{
		{
			name:        "fail - no more bytes",
			input:       []byte{},
			output:      0,
			expectedErr: ErrParserOutOfBound,
		},
		{
			name:        "fail - invalid length > 192 & length < 241",
			input:       []byte{193},
			output:      0,
			expectedErr: ErrParserOutOfBound,
		},
		{
			name:        "fail - invalid length > 240 & length < 255",
			input:       []byte{241},
			output:      0,
			expectedErr: ErrParserOutOfBound,
		},
		{
			name:        "fail - invalid length > 240 & length < 255",
			input:       []byte{241, 1},
			output:      0,
			expectedErr: ErrParserOutOfBound,
		},
		{
			name:   "pass - length less than 193",
			input:  []byte{190, 230, 131},
			output: 190,
		},
		{
			name:   "pass - length > 192 & length < 241",
			input:  []byte{195, 230, 112, 234, 98},
			output: 935,
		},
		{
			name:   "pass - length > 240 & length < 255",
			input:  []byte{242, 112, 78, 95, 115},
			output: 106767,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			p := NewBinaryParser(tc.input, definitions.Get())
			actual, err := p.ReadVariableLength()
			if tc.expectedErr != nil {
				require.Error(t, err)
				return
			}
			require.Equal(t, tc.output, actual)
		})
	}
}

func TestBinaryParser_ReadByte(t *testing.T) {
	testcases := []struct {
		name        string
		input       []byte
		expected    byte
		expectedErr error
	}{
		{
			name:        "fail - no more bytes",
			input:       []byte{},
			expected:    0,
			expectedErr: ErrParserOutOfBound,
		},
		{
			name:        "pass - returns first byte",
			input:       []byte{1, 2, 3, 4, 5},
			expected:    1,
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewBinaryParser(tc.input, definitions.Get())
			actual, err := p.ReadByte()
			if tc.expectedErr != nil {
				require.Error(t, err)
				return
			}
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestBinaryParser_ReadBytes(t *testing.T) {
	testcases := []struct {
		name            string
		input           []byte
		length          int
		expected        []byte
		expectedErr     error
		expectedHasMore bool
	}{
		{
			name:            "fail - negative length",
			input:           []byte{1, 2, 3, 4, 5},
			length:          -1,
			expectedErr:     ErrParserOutOfBound,
			expectedHasMore: true,
		},
		{
			name:            "pass - zero length",
			input:           []byte{1, 2, 3, 4, 5},
			length:          0,
			expected:        nil,
			expectedHasMore: true,
		},
		{
			name:            "pass - exact length",
			input:           []byte{1, 2, 3, 4, 5},
			length:          5,
			expected:        []byte{1, 2, 3, 4, 5},
			expectedHasMore: false,
		},
		{
			name:            "fail - over-read",
			input:           []byte{1, 2, 3, 4, 5},
			length:          6,
			expectedErr:     ErrParserOutOfBound,
			expectedHasMore: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewBinaryParser(tc.input, definitions.Get())
			actual, err := p.ReadBytes(tc.length)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
				require.Nil(t, actual)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, actual)
			}
			require.Equal(t, tc.expectedHasMore, p.HasMore())
		})
	}
}

func TestBinaryParser_ReadBytes_NegativeLengthDoesNotConsume(t *testing.T) {
	input := []byte{1, 2, 3, 4, 5}
	p := NewBinaryParser(input, definitions.Get())

	_, err := p.ReadBytes(-1)
	require.ErrorIs(t, err, ErrParserOutOfBound)

	remaining, err := p.ReadBytes(len(input))
	require.NoError(t, err)
	require.Equal(t, input, remaining)
}

func TestBinaryParser_HasMore(t *testing.T) {
	testcases := []struct {
		name     string
		input    []byte
		expected bool
	}{
		{
			name:     "pass - has more bytes",
			input:    []byte{1, 2, 3, 4, 5},
			expected: true,
		},
		{
			name:     "pass - no more bytes",
			input:    []byte{},
			expected: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewBinaryParser(tc.input, definitions.Get())
			require.Equal(t, tc.expected, p.HasMore())
		})
	}
}

func TestBinaryParser_Peek(t *testing.T) {
	testcases := []struct {
		name        string
		input       []byte
		expected    byte
		expectedErr error
	}{
		{
			name:        "fail - no more bytes",
			input:       []byte{},
			expected:    0,
			expectedErr: ErrParserOutOfBound,
		},
		{
			name:        "pass - returns first byte",
			input:       []byte{1, 2, 3, 4, 5},
			expected:    1,
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewBinaryParser(tc.input, definitions.Get())
			actual, err := p.Peek()
			if tc.expectedErr != nil {
				require.Error(t, err)
				return
			}
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestBinaryParser_ReadFieldHeader(t *testing.T) {
	testcases := []struct {
		name        string
		input       []byte
		expected    *definitions.FieldHeader
		expectedErr error
	}{
		{
			name:        "fail - no more bytes",
			input:       []byte{},
			expected:    nil,
			expectedErr: ErrParserOutOfBound,
		},
		{
			name:        "fail - no more bytes",
			input:       []byte{0},
			expected:    nil,
			expectedErr: ErrParserOutOfBound,
		},
		{
			name:        "fail - invalid fieldcode",
			input:       []byte{16, 0},
			expected:    nil,
			expectedErr: ErrInvalidFieldcode,
		},
		{
			name:        "fail - invalid typecode",
			input:       []byte{0, 0},
			expected:    nil,
			expectedErr: ErrInvalidTypecode,
		},
		{
			name:        "fail - invalid fieldcode",
			input:       []byte{0, 16},
			expected:    nil,
			expectedErr: ErrInvalidFieldcode,
		},
		{
			name:        "pass - returns first byte",
			input:       []byte{155},
			expected:    &definitions.FieldHeader{TypeCode: 9, FieldCode: 11},
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewBinaryParser(tc.input, definitions.Get())
			actual, err := p.readFieldHeader()
			if tc.expectedErr != nil {
				require.Error(t, err)
				return
			}
			require.Equal(t, tc.expectedErr, err)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestBinaryParser_ReadField(t *testing.T) {
	testcases := []struct {
		name        string
		input       []byte
		malleate    func() interfaces.Definitions
		expected    *definitions.FieldInstance
		expectedErr error
	}{
		{
			name:  "fail - no more bytes",
			input: []byte{},
			malleate: func() interfaces.Definitions {
				return definitions.Get()
			},
			expected:    nil,
			expectedErr: ErrParserOutOfBound,
		},
		{
			name:  "fail - invalid typecode",
			input: []byte{0, 0},
			malleate: func() interfaces.Definitions {
				return definitions.Get()
			},
			expected:    nil,
			expectedErr: ErrInvalidTypecode,
		},
		{
			name:  "fail - invalid fieldcode",
			input: []byte{0, 16},
			malleate: func() interfaces.Definitions {
				return definitions.Get()
			},
			expected:    nil,
			expectedErr: ErrInvalidFieldcode,
		},
		{
			name:  "fail - field not found",
			input: []byte{30},
			malleate: func() interfaces.Definitions {
				defs := testutil.NewMockDefinitions(gomock.NewController(t))
				defs.EXPECT().GetFieldNameByFieldHeader(gomock.Any()).AnyTimes().Return("", errors.New("field not found"))
				return defs
			},
			expected:    nil,
			expectedErr: errors.New("field not found"),
		},
		{
			name:  "fail - field instance not found",
			input: []byte{30},
			malleate: func() interfaces.Definitions {
				defs := testutil.NewMockDefinitions(gomock.NewController(t))
				defs.EXPECT().GetFieldNameByFieldHeader(gomock.Any()).AnyTimes().Return("AccountRoot", nil)
				defs.EXPECT().GetFieldInstanceByFieldName(gomock.Any()).AnyTimes().Return(nil, errors.New("field instance not found"))
				return defs
			},
			expected:    nil,
			expectedErr: errors.New("field instance not found"),
		},
		{
			name:  "pass - returns field instance",
			input: []byte{30},
			malleate: func() interfaces.Definitions {
				defs := testutil.NewMockDefinitions(gomock.NewController(t))
				defs.EXPECT().GetFieldNameByFieldHeader(gomock.Any()).AnyTimes().Return("AccountRoot", nil)
				defs.EXPECT().GetFieldInstanceByFieldName(gomock.Any()).AnyTimes().Return(&definitions.FieldInstance{
					FieldHeader: &definitions.FieldHeader{TypeCode: 9, FieldCode: 11},
				}, nil)
				return defs
			},
			expected: &definitions.FieldInstance{
				FieldHeader: &definitions.FieldHeader{TypeCode: 9, FieldCode: 11},
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			definitions := tc.malleate()
			p := NewBinaryParser(tc.input, definitions)
			actual, err := p.ReadField()
			if tc.expectedErr != nil {
				require.Error(t, err)
				return
			}
			require.Equal(t, tc.expected, actual)
		})
	}
}
