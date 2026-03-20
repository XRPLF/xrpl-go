package crypto

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkDERHexToSig(b *testing.B) {
	hexSignature := "3045022100E1617F1A3C85B5BC8FA6224F893FE9068BEA8F8D075EE144F6F9D255C829761802206FD9B361CDE83A0C3D5654232F1D7CFB1A614E9A8F9B1A861564029065516E64"
	for i := 0; i < b.N; i++ {
		DERHexToSig(hexSignature)
	}
}

func TestDERHexToSig(t *testing.T) {
	testCases := []struct {
		name         string
		hexSignature string
		expectedR    string
		expectedS    string
		expectError  error
	}{
		{
			name:         "fail - invalid hex string",
			hexSignature: "invalid",
			expectedR:    "",
			expectedS:    "",
			expectError:  ErrInvalidHexString,
		},
		{
			name:         "fail - invalid signature tag",
			hexSignature: "3145022100E1617F1A3C85B5BC8FA6224F893FE9068BEA8F8D075EE144F6F9D255C829761802206FD9B361CDE83A0C3D5654232F1D7CFB1A614E9A8F9B1A861564029065516E64",
			expectedR:    "",
			expectedS:    "",
			expectError:  ErrInvalidDERSignature,
		},
		{
			name:         "fail - invalid length",
			hexSignature: "3044022100E1617F1A3C85B5BC8FA6224F893FE9068BEA8F8D075EE144F6F9D255C829761802206FD9B361CDE83A0C3D5654232F1D7CFB1A614E9A8F9B1A861564029065516E64",
			expectedR:    "",
			expectedS:    "",
			expectError:  ErrInvalidDERSignature,
		},
		{
			name:         "fail - invalid parseInt",
			hexSignature: "3003020301",
			expectedR:    "",
			expectedS:    "",
			expectError:  ErrInvalidDERSignature,
		},
		{
			name:         "fail - invalid parseInt - 2",
			hexSignature: "3006020101020301",
			expectedR:    "",
			expectedS:    "",
			expectError:  ErrInvalidDERSignature,
		},
		{
			name:         "fail - invalid leftover bytes",
			hexSignature: "300702010102010101",
			expectedR:    "",
			expectedS:    "",
			expectError:  ErrLeftoverBytes,
		},
		{
			name:         "pass - valid DER signature",
			hexSignature: "3045022100E1617F1A3C85B5BC8FA6224F893FE9068BEA8F8D075EE144F6F9D255C829761802206FD9B361CDE83A0C3D5654232F1D7CFB1A614E9A8F9B1A861564029065516E64",
			expectedR:    "e1617f1a3c85b5bc8fa6224f893fe9068bea8f8d075ee144f6f9d255c8297618",
			expectedS:    "6fd9b361cde83a0c3d5654232f1d7cfb1a614e9a8f9b1a861564029065516e64",
			expectError:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, s, err := DERHexToSig(tc.hexSignature)

			if tc.expectError != nil {
				require.Equal(t, tc.expectError, err)
			} else {
				require.Equal(t, tc.expectedR, hex.EncodeToString(r))
				require.Equal(t, tc.expectedS, hex.EncodeToString(s))
				require.NoError(t, err)
			}
		})
	}
}

func BenchmarkDERHexFromSig(b *testing.B) {
	rHex := "E1617F1A3C85B5BC8FA6224F893FE9068BEA8F8D075EE144F6F9D255C8297618"
	sHex := "6FD9B361CDE83A0C3D5654232F1D7CFB1A614E9A8F9B1A861564029065516E64"
	for i := 0; i < b.N; i++ {
		DERHexFromSig(rHex, sHex)
	}
}

func TestDERHexFromSig(t *testing.T) {
	testCases := []struct {
		name        string
		rHex        string
		sHex        string
		expectedDER string
		expectError error
	}{
		{
			name:        "pass - valid r and s values",
			rHex:        "e1617f1a3c85b5bc8fa6224f893fe9068bea8f8d075ee144f6f9d255c8297618",
			sHex:        "6fd9b361cde83a0c3d5654232f1d7cfb1a614e9a8f9b1a861564029065516e64",
			expectedDER: "3045022100e1617f1a3c85b5bc8fa6224f893fe9068bea8f8d075ee144f6f9d255c829761802206fd9b361cde83a0c3d5654232f1d7cfb1a614e9a8f9b1a861564029065516e64",
			expectError: nil,
		},
		{
			name:        "fail - invalid r hex string",
			rHex:        "invalid",
			sHex:        "6fd9b361cde83a0c3d5654232f1d7cfb1a614e9a8f9b1a861564029065516e64",
			expectedDER: "",
			expectError: ErrInvalidHexString,
		},
		{
			name:        "fail - invalid s hex string",
			rHex:        "e1617f1a3c85b5bc8fa6224f893fe9068bea8f8d075ee144f6f9d255c8297618",
			sHex:        "invalid",
			expectedDER: "",
			expectError: ErrInvalidHexString,
		},
		{
			name:        "pass - r value with leading zero",
			rHex:        "00e1617f1a3c85b5bc8fa6224f893fe9068bea8f8d075ee144f6f9d255c8297618",
			sHex:        "6fd9b361cde83a0c3d5654232f1d7cfb1a614e9a8f9b1a861564029065516e64",
			expectedDER: "3045022100e1617f1a3c85b5bc8fa6224f893fe9068bea8f8d075ee144f6f9d255c829761802206fd9b361cde83a0c3d5654232f1d7cfb1a614e9a8f9b1a861564029065516e64",
			expectError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := DERHexFromSig(tc.rHex, tc.sHex)

			if tc.expectError != nil {
				require.Equal(t, tc.expectError, err)
			} else {
				require.Equal(t, tc.expectedDER, result)
				require.NoError(t, err)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	testcases := []struct {
		name              string
		data              []byte
		expectedInt       *big.Int
		expectedRemaining []byte
		expectError       error
	}{
		{
			name:              "fail - not enough data",
			data:              []byte{},
			expectedInt:       nil,
			expectedRemaining: nil,
			expectError:       ErrInvalidDERNotEnoughData,
		},
		{
			name:              "fail - invalid tag",
			data:              []byte{0x01, 0x02, 0x03},
			expectedInt:       nil,
			expectedRemaining: nil,
			expectError:       ErrInvalidDERIntegerTag,
		},
		{
			name:              "fail - invalid length",
			data:              []byte{0x02, 0x02, 0x01},
			expectedInt:       nil,
			expectedRemaining: nil,
			expectError:       ErrInvalidDERNotEnoughData,
		},
		{
			name:              "pass - valid data",
			data:              []byte{0x02, 0x01, 0x01},
			expectedInt:       big.NewInt(1),
			expectedRemaining: []byte{},
			expectError:       nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			n, remaining, err := parseInt(tc.data)
			if tc.expectError != nil {
				require.Equal(t, tc.expectError, err)
			} else {
				require.Equal(t, tc.expectedInt.Bytes(), n.Bytes())
				require.Equal(t, tc.expectedRemaining, remaining)
				require.NoError(t, err)
			}
		})
	}
}
