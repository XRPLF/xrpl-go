package serdes

import (
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/serdes/interfaces"
	"github.com/Peersyst/xrpl-go/binary-codec/serdes/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestBinarySerializer_EncodeVariableLength(t *testing.T) {
	tt := []struct {
		name        string
		len         int
		expected    []byte
		expectedErr error
	}{
		{
			name:        "length less than 193",
			len:         100,
			expected:    []byte{0x64},
			expectedErr: nil,
		},
		{
			name:        "length more than 193 and less than 12481",
			len:         1000,
			expected:    []byte{0xC4, 0x27},
			expectedErr: nil,
		},
		{
			name:        "length more than 12841 ad less than 918744",
			len:         20000,
			expected:    []byte{0xF1, 0x1D, 0x5F},
			expectedErr: nil,
		},
		{
			name:        "length more than 918744",
			len:         1000000,
			expected:    nil,
			expectedErr: ErrLengthPrefixTooLong,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			s := strings.Repeat("A2", tc.len)
			b, _ := hex.DecodeString(s)
			require.Len(t, b, tc.len)
			actual, err := encodeVariableLength(len(b))
			if tc.expectedErr != nil {
				require.Error(t, err, tc.expectedErr.Error())
				require.Nil(t, actual)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestBinarySerializer_Put(t *testing.T) {
	testcases := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "pass",
			input:    []byte{1, 2, 3, 4, 5},
			expected: []byte{1, 2, 3, 4, 5},
		},
		{
			name:     "pass - empty input",
			input:    []byte{},
			expected: []byte(nil),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewBinarySerializer(NewFieldIDCodec(definitions.Get()))
			s.put(tc.input)
			require.Equal(t, tc.expected, s.GetSink())
		})
	}
}

func TestBinarySerializer_GetSink(t *testing.T) {
	testcases := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "pass",
			input:    []byte{1, 2, 3, 4, 5},
			expected: []byte{1, 2, 3, 4, 5},
		},
		{
			name:     "pass - empty input",
			input:    []byte{},
			expected: []byte(nil),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewBinarySerializer(NewFieldIDCodec(definitions.Get()))
			s.put(tc.input)
			require.Equal(t, tc.expected, s.GetSink())
		})
	}
}

func TestBinarySerializer_WriteFieldAndValue(t *testing.T) {
	testcases := []struct {
		name          string
		fieldInstance definitions.FieldInstance
		value         []byte
		malleate      func() interfaces.FieldIDCodec
		expected      []byte
		expectedErr   error
	}{
		{
			name: "fail - field not found",
			malleate: func() interfaces.FieldIDCodec {
				codec := testutil.NewMockFieldIDCodec(gomock.NewController(t))
				codec.EXPECT().Encode("LedgerEntry").Return(nil, errors.New("field not found"))
				return codec
			},
			fieldInstance: definitions.FieldInstance{
				FieldName: "LedgerEntry",
			},
			expectedErr: &definitions.NotFoundError{
				Instance: "FieldName",
				Input:    "LedgerEntry",
			},
		},
		{
			name:  "fail - vle encoded variable length too long",
			value: []byte(strings.Repeat("A", 1000000)),
			malleate: func() interfaces.FieldIDCodec {
				codec := testutil.NewMockFieldIDCodec(gomock.NewController(t))
				codec.EXPECT().Encode("LedgerEntry").Return([]byte{1, 2}, nil)
				return codec
			},
			fieldInstance: definitions.FieldInstance{
				FieldName: "LedgerEntry",
				FieldInfo: &definitions.FieldInfo{
					IsVLEncoded: true,
					Type:        "Blob",
				},
			},
			expectedErr: ErrLengthPrefixTooLong,
		},
		{
			name: "pass - vle encoded",
			malleate: func() interfaces.FieldIDCodec {
				codec := testutil.NewMockFieldIDCodec(gomock.NewController(t))
				codec.EXPECT().Encode("PublicKey").Return([]byte{1, 2}, nil)
				return codec
			},
			fieldInstance: definitions.FieldInstance{
				FieldName: "PublicKey",
				FieldInfo: &definitions.FieldInfo{
					IsVLEncoded: true,
					Type:        "Blob",
				},
			},
			value:       []byte{3, 4, 5},
			expected:    []byte{1, 2, 3, 3, 4, 5},
			expectedErr: nil,
		},
		{
			name: "pass - non-vle encoded",
			malleate: func() interfaces.FieldIDCodec {
				codec := testutil.NewMockFieldIDCodec(gomock.NewController(t))
				codec.EXPECT().Encode("LedgerEntry").Return([]byte{1, 2}, nil)
				return codec
			},
			fieldInstance: definitions.FieldInstance{
				FieldName: "LedgerEntry",
				FieldInfo: &definitions.FieldInfo{
					IsVLEncoded: false,
					Type:        "Uint16",
				},
			},
			value:       []byte{3, 4, 5},
			expected:    []byte{1, 2, 3, 4, 5},
			expectedErr: nil,
		},
		{
			name: "pass - type STObject",
			malleate: func() interfaces.FieldIDCodec {
				codec := testutil.NewMockFieldIDCodec(gomock.NewController(t))
				codec.EXPECT().Encode("LedgerEntry").Return([]byte{1, 2}, nil)
				return codec
			},
			fieldInstance: definitions.FieldInstance{
				FieldName: "LedgerEntry",
				FieldInfo: &definitions.FieldInfo{
					Type: "STObject",
				},
			},
			value:       []byte{3, 4, 5},
			expected:    []byte{1, 2, 3, 4, 5, 0xE1},
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			fieldIDCodec := tc.malleate()
			s := NewBinarySerializer(fieldIDCodec)
			err := s.WriteFieldAndValue(tc.fieldInstance, tc.value)
			if tc.expectedErr != nil {
				require.Error(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, s.GetSink())
			}
		})
	}
}
