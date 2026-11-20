package types

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/Peersyst/xrpl-go/binary-codec/types/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestHash_getLength(t *testing.T) {
	tt := []struct {
		name   string
		length int
	}{
		{
			name:   "Valid length (1)",
			length: 32,
		},
		{
			name:   "Valid length (2)",
			length: 64,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			hash := newHash(tc.length)
			require.Equal(t, tc.length, hash.getLength())
		})
	}
}

func TestHash_FromJson(t *testing.T) {
	tt := []struct {
		name        string
		json        any
		length      int
		expected    []byte
		expectedErr error
	}{
		{
			name:        "Valid hash of length 32",
			json:        "0316020000000000000000000000000000000000000000000000000000000000",
			length:      32,
			expected:    []byte{0x03, 0x16, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectedErr: nil,
		},
		{
			name:        "Invalid hex string",
			json:        "031G020000000000000000000000000000000000000000000000000000000000",
			length:      32,
			expected:    nil,
			expectedErr: &ErrInvalidHexString{Err: hex.InvalidByteError('G')},
		},
		{
			name:        "Invalid hash type",
			json:        123,
			length:      32,
			expectedErr: &ErrInvalidHashType{},
		},
		{
			name:        "Invalid hash length",
			json:        "031602000000000000000000000000000000000000000000000000000000000000",
			length:      32,
			expectedErr: &ErrInvalidHashLength{Expected: 32},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			hash := newHash(tc.length)
			actual, err := hash.FromJSON(tc.json)
			require.Equal(t, tc.expected, actual)
			if tc.expectedErr != nil {
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			}
		})
	}
}

func TestHash_ToJson(t *testing.T) {
	tt := []struct {
		name        string
		hash        []byte
		length      int
		expected    any
		expectedErr error
		setup       func(t *testing.T) (*hash, *testutil.MockBinaryParser)
	}{
		{
			name:     "Valid hash of length 32",
			hash:     []byte{0x03, 0x16, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			length:   32,
			expected: "0316020000000000000000000000000000000000000000000000000000000000",
			setup: func(t *testing.T) (*hash, *testutil.MockBinaryParser) {
				ctrl := gomock.NewController(t)
				mock := testutil.NewMockBinaryParser(ctrl)
				mock.EXPECT().ReadBytes(32).Return([]byte{0x03, 0x16, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, nil)
				return &hash{Length: 32}, mock
			},
		},
		{
			name:        "ReadBytes error",
			hash:        []byte{0x03, 0x16, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			length:      32,
			expected:    nil,
			expectedErr: errors.New("errReadBytes"),
			setup: func(t *testing.T) (*hash, *testutil.MockBinaryParser) {
				ctrl := gomock.NewController(t)
				mock := testutil.NewMockBinaryParser(ctrl)
				mock.EXPECT().ReadBytes(32).Return([]byte{}, errors.New("errReadBytes"))
				return &hash{Length: 32}, mock
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			hash, parser := tc.setup(t)
			actual, err := hash.ToJSON(parser, DecodeContext{})
			require.Equal(t, tc.expected, actual)
			require.Equal(t, tc.expectedErr, err)
		})
	}
}
