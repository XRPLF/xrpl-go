package types

import (
	"bytes"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/Peersyst/xrpl-go/binary-codec/types/testutil"
	"github.com/golang/mock/gomock"
)

func TestBlob_FromJson(t *testing.T) {
	tt := []struct {
		name        string
		input       string
		expected    []byte
		expectedErr error
	}{
		{
			name:        "Valid Blob",
			input:       "00",
			expected:    []byte{0},
			expectedErr: nil,
		},
		{
			name:        "Valid Blob",
			input:       "000102030405060708090A0B0C0D0E0F",
			expected:    []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
			expectedErr: nil,
		},
		{
			name:        "Invalid hex string",
			input:       "000102030405060708090A0B0C0D0E0G",
			expected:    nil,
			expectedErr: hex.InvalidByteError('G'),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			blob := &Blob{}
			actual, err := blob.FromJSON(tc.input)
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}
			if !bytes.Equal(actual, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, actual)
			}
		})
	}
}

func TestBlob_ToJson(t *testing.T) {
	tt := []struct {
		name        string
		input       []byte
		expected    any
		opts        []int
		expectedErr error
		setup       func(t *testing.T) (*Blob, *testutil.MockBinaryParser)
	}{
		{
			name:        "Valid Blob",
			input:       []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
			expected:    "000102030405060708090A0B0C0D0E0F",
			opts:        []int{16},
			expectedErr: nil,
			setup: func(t *testing.T) (*Blob, *testutil.MockBinaryParser) {
				ctrl := gomock.NewController(t)
				mock := testutil.NewMockBinaryParser(ctrl)
				mock.EXPECT().ReadBytes(16).Return([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}, nil)
				return &Blob{}, mock
			},
		},
		{
			name:        "ReadBytes error",
			input:       []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
			expected:    nil,
			opts:        []int{16},
			expectedErr: errors.New("errReadBytes"),
			setup: func(t *testing.T) (*Blob, *testutil.MockBinaryParser) {
				ctrl := gomock.NewController(t)
				mock := testutil.NewMockBinaryParser(ctrl)
				mock.EXPECT().ReadBytes(16).Return([]byte{}, errors.New("errReadBytes"))
				return &Blob{}, mock
			},
		},
		{
			name:        "No length prefix",
			input:       []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
			expected:    nil,
			opts:        nil,
			expectedErr: ErrNoLengthPrefix,
			setup: func(t *testing.T) (*Blob, *testutil.MockBinaryParser) {
				ctrl := gomock.NewController(t)
				mock := testutil.NewMockBinaryParser(ctrl)
				return &Blob{}, mock
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			blob, parser := tc.setup(t)
			actual, err := blob.ToJSON(parser, tc.opts...)
			if err != nil && err.Error() != tc.expectedErr.Error() {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}
			if actual != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, actual)
			}
		})
	}
}
