package keypairs

import (
	"testing"

	"github.com/Peersyst/xrpl-go/keypairs/interfaces"
	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/stretchr/testify/require"
)

const (
	testEdPublicKey  = "ED4924A9045FE5ED8B22BAA7B6229A72A287CCF3EA287AADD3A032A24C0F008FA6"
	testEdPrivateKey = "EDBB3ECA8985E1484FA6A28C4B30FB0042A2CC5DF3EC8DC37B5F3D126DDFD3CA14"

	testSecpRawPrivateKey      = "B167A9F3B9E60A4F93695713682C102438620AA1785C3AE635F53E5B6261071A"
	testSecpPrefixedPrivateKey = "00" + testSecpRawPrivateKey
	testSecpCompressedEvenKey  = "02950F4710101A25073BF37086D73FBBD00C7A6B0F91097D8F0BC6D268C400D56E"
	testSecpCompressedOddKey   = "030D58EB48B4420B1F7B9DF55087E0E29FEF0E8468F9A6825B01CA2C361042D435"
	testSecpUncompressedKey    = "04950F4710101A25073BF37086D73FBBD00C7A6B0F91097D8F0BC6D268C400D56EB97051C3A7F51A6ECC0EFF7F7622437593DB7E20165EEF8100570288AF3D5A3C"
)

func TestGetCryptoImplementationFromKey(t *testing.T) {
	testcases := []struct {
		name     string
		keyType  keyType
		input    string
		expected interfaces.KeypairCryptoAlg
	}{
		{
			name:     "public - ED25519",
			keyType:  publicKeyType,
			input:    testEdPublicKey,
			expected: crypto.ED25519(),
		},
		{
			name:     "public - compressed secp256k1 even Y",
			keyType:  publicKeyType,
			input:    testSecpCompressedEvenKey,
			expected: crypto.SECP256K1(),
		},
		{
			name:     "public - compressed secp256k1 odd Y",
			keyType:  publicKeyType,
			input:    testSecpCompressedOddKey,
			expected: crypto.SECP256K1(),
		},
		{
			name:     "public - uncompressed secp256k1",
			keyType:  publicKeyType,
			input:    testSecpUncompressedKey,
			expected: crypto.SECP256K1(),
		},
		{
			name:     "private - ED25519",
			keyType:  privateKeyType,
			input:    testEdPrivateKey,
			expected: crypto.ED25519(),
		},
		{
			name:     "private - raw secp256k1",
			keyType:  privateKeyType,
			input:    testSecpRawPrivateKey,
			expected: crypto.SECP256K1(),
		},
		{
			name:     "private - raw secp256k1 beginning with ED byte",
			keyType:  privateKeyType,
			input:    "ED" + testSecpRawPrivateKey[2:],
			expected: crypto.SECP256K1(),
		},
		{
			name:     "private - 00-prefixed secp256k1",
			keyType:  privateKeyType,
			input:    testSecpPrefixedPrivateKey,
			expected: crypto.SECP256K1(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual, _, err := getCryptoImplementationFromKey(tc.input, tc.keyType)

			require.NoError(t, err)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestGetCryptoImplementationFromKeyRejectsInvalidFormats(t *testing.T) {
	testcases := []struct {
		name        string
		keyType     keyType
		input       string
		expectedErr error
	}{
		{
			name:        "public - empty",
			keyType:     publicKeyType,
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "public - odd hex length",
			keyType:     publicKeyType,
			input:       testEdPublicKey[:len(testEdPublicKey)-1],
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "public - malformed full-length hex",
			keyType:     publicKeyType,
			input:       testEdPublicKey[:len(testEdPublicKey)-1] + "Z",
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "public - truncated ED25519",
			keyType:     publicKeyType,
			input:       testEdPublicKey[:len(testEdPublicKey)-2],
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "public - oversized ED25519",
			keyType:     publicKeyType,
			input:       testEdPublicKey + "00",
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "public - unsupported prefix",
			keyType:     publicKeyType,
			input:       "05" + testSecpCompressedEvenKey[2:],
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "public - uncompressed key truncated to compressed length",
			keyType:     publicKeyType,
			input:       testSecpUncompressedKey[:len(testSecpCompressedEvenKey)],
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "public - uncompressed key oversized",
			keyType:     publicKeyType,
			input:       testSecpUncompressedKey + "00",
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "public - raw private key type mismatch",
			keyType:     publicKeyType,
			input:       testSecpRawPrivateKey,
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "public - prefixed private key type mismatch",
			keyType:     publicKeyType,
			input:       testSecpPrefixedPrivateKey,
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "private - empty",
			keyType:     privateKeyType,
			expectedErr: ErrInvalidPrivateKeyFormat,
		},
		{
			name:        "private - odd hex length",
			keyType:     privateKeyType,
			input:       testSecpRawPrivateKey + "0",
			expectedErr: ErrInvalidPrivateKeyFormat,
		},
		{
			name:        "private - malformed full-length hex",
			keyType:     privateKeyType,
			input:       testSecpRawPrivateKey[:len(testSecpRawPrivateKey)-1] + "Z",
			expectedErr: ErrInvalidPrivateKeyFormat,
		},
		{
			name:        "private - truncated ED25519",
			keyType:     privateKeyType,
			input:       testEdPrivateKey[:len(testEdPrivateKey)-4],
			expectedErr: ErrInvalidPrivateKeyFormat,
		},
		{
			name:        "private - oversized ED25519",
			keyType:     privateKeyType,
			input:       testEdPrivateKey + "00",
			expectedErr: ErrInvalidPrivateKeyFormat,
		},
		{
			name:        "private - unsupported 33-byte prefix",
			keyType:     privateKeyType,
			input:       "01" + testSecpRawPrivateKey,
			expectedErr: ErrInvalidPrivateKeyFormat,
		},
		{
			name:        "private - compressed public key type mismatch",
			keyType:     privateKeyType,
			input:       testSecpCompressedEvenKey,
			expectedErr: ErrInvalidPrivateKeyFormat,
		},
		{
			name:        "private - uncompressed public key type mismatch",
			keyType:     privateKeyType,
			input:       testSecpUncompressedKey,
			expectedErr: ErrInvalidPrivateKeyFormat,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual, decoded, err := getCryptoImplementationFromKey(tc.input, tc.keyType)

			require.Nil(t, actual)
			require.Nil(t, decoded)
			require.ErrorIs(t, err, tc.expectedErr)
			require.ErrorIs(t, err, ErrInvalidCryptoImplementation)
			if tc.input != "" {
				require.NotContains(t, err.Error(), tc.input)
			}
		})
	}
}
