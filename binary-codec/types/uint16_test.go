package types

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/types/testutil"
	"github.com/golang/mock/gomock"
)

func TestUint16_FromJson(t *testing.T) {
	tt := []struct {
		name        string
		input       any
		expected    []byte
		expectedErr error
	}{
		{
			name:     "fail - invalid ledger entry type",
			input:    "invalid",
			expected: nil,
			expectedErr: &definitions.NotFoundError{
				Instance: "LedgerEntryTypeName",
				Input:    "invalid",
			},
		},
		{
			name:        "pass - valid uint16 from uint16",
			input:       1,
			expected:    []byte{0, 1},
			expectedErr: nil,
		},
		{
			name:        "pass - valid uint16 from uint16 (2)",
			input:       100,
			expected:    []byte{0, 100},
			expectedErr: nil,
		},
		{
			name:        "pass - valid uint16 from uint16 (3)",
			input:       255,
			expected:    []byte{0, 255},
			expectedErr: nil,
		},
		{
			name:        "pass - valid uint16 from TransactionType",
			input:       "Payment",
			expected:    []byte{0, 0},
			expectedErr: nil,
		},
		{
			name:        "pass - valid uint16 from TransactionType (2)",
			input:       "EscrowCreate",
			expected:    []byte{0, 1},
			expectedErr: nil,
		},
		// TODO: Add test for overflow case
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			class := &UInt16{}
			actual, err := class.FromJSON(tc.input)
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}
			if !bytes.Equal(actual, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, actual)
			}
		})
	}
}

func TestUint16_ToJson(t *testing.T) {
	tt := []struct {
		name        string
		input       []byte
		expected    any
		expectedErr error
		setup       func(t *testing.T) (*UInt16, *testutil.MockBinaryParser)
	}{
		{
			name:        "Valid uint16",
			input:       []byte{0, 1},
			expected:    1,
			expectedErr: nil,
			setup: func(t *testing.T) (*UInt16, *testutil.MockBinaryParser) {
				ctrl := gomock.NewController(t)
				mockParser := testutil.NewMockBinaryParser(ctrl)
				mockParser.EXPECT().ReadBytes(2).Return([]byte{0, 1}, nil)
				return &UInt16{}, mockParser
			},
		},
		{
			name:        "Valid uint16 (2)",
			input:       []byte{0, 100},
			expected:    100,
			expectedErr: nil,
			setup: func(t *testing.T) (*UInt16, *testutil.MockBinaryParser) {
				ctrl := gomock.NewController(t)
				mockParser := testutil.NewMockBinaryParser(ctrl)
				mockParser.EXPECT().ReadBytes(2).Return([]byte{0, 100}, nil)
				return &UInt16{}, mockParser
			},
		},
		{
			name:        "Valid uint16 (3)",
			input:       []byte{0, 255},
			expected:    255,
			expectedErr: nil,
			setup: func(t *testing.T) (*UInt16, *testutil.MockBinaryParser) {
				ctrl := gomock.NewController(t)
				mockParser := testutil.NewMockBinaryParser(ctrl)
				mockParser.EXPECT().ReadBytes(2).Return([]byte{0, 255}, nil)
				return &UInt16{}, mockParser
			},
		},
		{
			name:        "Invalid ReadBytes",
			input:       []byte{0, 1},
			expected:    nil,
			expectedErr: errors.New("readBytes: error"),
			setup: func(t *testing.T) (*UInt16, *testutil.MockBinaryParser) {
				ctrl := gomock.NewController(t)
				mockParser := testutil.NewMockBinaryParser(ctrl)
				mockParser.EXPECT().ReadBytes(2).Return([]byte{}, errors.New("readBytes: error"))
				return &UInt16{}, mockParser
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			class, mockParser := tc.setup(t)
			actual, err := class.ToJSON(mockParser)
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}
			if actual != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, actual)
			}
		})
	}
}
