package addresscodec

import (
	"errors"
	"testing"

	"github.com/Peersyst/xrpl-go/address-codec/interfaces"
	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	testcases := []struct {
		name           string
		input          []byte
		inputPrefix    []byte
		inputLength    int
		expectedOutput string
		expectedErr    error
	}{
		{
			name:           "fail - invalid length",
			input:          []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			inputPrefix:    []byte{AccountAddressPrefix},
			inputLength:    17,
			expectedOutput: "",
			expectedErr:    &EncodeLengthError{Instance: "Encode", Expected: 17, Input: 16},
		},
		{
			name:           "pass - successful encode",
			input:          []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			inputPrefix:    []byte{AccountAddressPrefix},
			inputLength:    16,
			expectedOutput: "rrrrrrrrrrrrrrrrrp9U13b",
			expectedErr:    nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Encode(tc.input, tc.inputPrefix, tc.inputLength)
			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.Equal(t, tc.expectedOutput, got)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	testcases := []struct {
		name           string
		input          string
		inputPrefix    []byte
		expectedOutput []byte
		expectedErr    error
	}{
		{
			name:           "pass - successful decode - 1",
			input:          "rrrrrrrrrrrrrrrrrp9U13b",
			inputPrefix:    []byte{AccountAddressPrefix},
			expectedOutput: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectedErr:    nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			res, _ := Decode(tc.input, tc.inputPrefix)
			require.Equal(t, tc.expectedOutput, res)
		})
	}
}

func TestEncodeClassicAddressFromPublicKeyHex(t *testing.T) {
	tt := []struct {
		name           string
		input          string
		expectedOutput string
		expectedErr    error
	}{
		{
			name:           "pass - successfully generate address from a 32-byte ED25519 public key hex string WITH prefix",
			input:          "ED9434799226374926EDA3B54B1B461B4ABF7237962EAE18528FEA67595397FA32",
			expectedOutput: "rDTXLQ7ZKZVKz33zJbHjgVShjsBnqMBhmN",
			expectedErr:    nil,
		},
		{
			name:           "pass - derive correct address from public key",
			input:          "ED731C39781B964904E1FEEFFC9F99442196BCB5F499105A79533E2D678CA7D3D2",
			expectedOutput: "rhTCnDC7v1Jp7NAupzisv6ynWHD161Q9nV",
			expectedErr:    nil,
		},
		{
			name:           "fail - invalid hex string",
			input:          "yurtyurtyurtyurt",
			expectedOutput: "",
			expectedErr:    errors.New("encoding/hex: invalid byte: U+0079 'y'"),
		},
		{
			name:           "fail - invalid public key - too short",
			input:          "ED9434799226374926EDA3B54B1B461B",
			expectedOutput: "",
			expectedErr:    &EncodeLengthError{Instance: "PublicKey", Input: 16, Expected: 33},
		},
		{
			name:           "fail - invalid public key - too long",
			input:          "ED9434799226374926EDA3B54B1B461B4ABF7237962EAE18528FEA67595397FA32ED9434799226374926EDA3B54B1B461B4ABF7237962EAE18528FEA67595397FA32",
			expectedOutput: "",
			expectedErr:    &EncodeLengthError{Instance: "PublicKey", Input: 66, Expected: 33},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got, err := EncodeClassicAddressFromPublicKeyHex(tc.input)

			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, got)
			}
		})
	}
}

func TestEncodeSeed(t *testing.T) {
	tt := []struct {
		name              string
		input             []byte
		inputEncodingType interfaces.CryptoImplementation
		expectedOutput    string
		expectedErr       error
	}{
		{
			name:              "pass - successful encode - ED25519",
			input:             []byte("yurtyurtyurtyurt"),
			inputEncodingType: crypto.ED25519(),
			expectedOutput:    "sEdTzRkEgPoxDG1mJ6WkSucHWnMkm1H",
			expectedErr:       nil,
		},
		{
			name:              "pass - successful encode - SECP256K1",
			input:             []byte("yurtyurtyurtyurt"),
			inputEncodingType: crypto.SECP256K1(),
			expectedOutput:    "shPSkLzQNWfyXjZ7bbwgCky6twagA",
			expectedErr:       nil,
		},
		{
			name:              "pass - successful encode - ED25519 additional",
			input:             []byte("testingsomething"),
			inputEncodingType: crypto.ED25519(),
			expectedOutput:    "sEdTvLVDRVJsrUyBiCPTHDs46GUKQAr",
			expectedErr:       nil,
		},
		{
			name:              "pass - successful encode - SECP256K1 additional",
			input:             []byte("testingsomething"),
			inputEncodingType: crypto.SECP256K1(),
			expectedOutput:    "shKMVJjV52uudwfS7HzzaiwmZqVeP",
			expectedErr:       nil,
		},
		{
			name:              "fail - unsuccessful encode - invalid entropy length",
			input:             []byte{0x00},
			inputEncodingType: crypto.ED25519(),
			expectedOutput:    "",
			expectedErr:       &EncodeLengthError{Instance: "Entropy", Input: len([]byte{0x00}), Expected: FamilySeedLength},
		},
		{
			name:              "fail - unsuccessful encode - invalid encoding type",
			input:             []byte("testingsomething"),
			inputEncodingType: nil,
			expectedOutput:    "",
			expectedErr:       errors.New("encoding type must be `ed25519` or `secp256k1`"),
		},
		{
			name:              "fail - invalid CryptoAlgorithm Uint type returns err",
			input:             []byte("testingsomething"),
			inputEncodingType: nil,
			expectedOutput:    "",
			expectedErr:       errors.New("encoding type must be `ed25519` or `secp256k1`"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got, err := EncodeSeed(tc.input, tc.inputEncodingType)

			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.Equal(t, tc.expectedOutput, got)
			}
		})
	}
}

func TestDecodeSeed(t *testing.T) {
	tt := []struct {
		name              string
		input             string
		expectedOutput    []byte
		expectedAlgorithm interfaces.CryptoImplementation
		expectedErr       error
	}{
		{
			name:              "pass - successful decode - ED25519",
			input:             "sEdTzRkEgPoxDG1mJ6WkSucHWnMkm1H",
			expectedOutput:    []byte("yurtyurtyurtyurt"),
			expectedAlgorithm: crypto.ED25519(),
			expectedErr:       nil,
		},
		{
			name:              "pass - successful decode - SECP256K1",
			input:             "shPSkLzQNWfyXjZ7bbwgCky6twagA",
			expectedOutput:    []byte("yurtyurtyurtyurt"),
			expectedAlgorithm: crypto.SECP256K1(),
			expectedErr:       nil,
		},
		{
			name:              "pass - successful decode - ED25519 additional",
			input:             "sEdTvLVDRVJsrUyBiCPTHDs46GUKQAr",
			expectedOutput:    []byte("testingsomething"),
			expectedAlgorithm: crypto.ED25519(),
			expectedErr:       nil,
		},
		{
			name:              "pass - successful decode - SECP256K1 additional",
			input:             "shKMVJjV52uudwfS7HzzaiwmZqVeP",
			expectedOutput:    []byte("testingsomething"),
			expectedAlgorithm: crypto.SECP256K1(),
			expectedErr:       nil,
		},
		{
			name:              "fail - unsuccessful decode - invalid seed",
			input:             "yurt",
			expectedOutput:    nil,
			expectedAlgorithm: crypto.ED25519(),
			expectedErr:       errors.New("invalid seed; could not determine encoding algorithm"),
		},
		{
			name:              "fail - unsuccessful decode - invalid seed",
			input:             "sEdTzRkEgPoxDG1mJ6WkSucHWnMkm1D",
			expectedOutput:    nil,
			expectedAlgorithm: nil,
			expectedErr:       ErrInvalidSeed,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got, algorithm, err := DecodeSeed(tc.input)

			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
				require.Nil(t, tc.expectedOutput)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, got)
				require.Equal(t, tc.expectedAlgorithm, algorithm)
			}
		})
	}
}

func TestDecodeAddressToAccountID(t *testing.T) {
	tt := []struct {
		name              string
		input             string
		expectedPrefix    []byte
		expectedAccountID []byte
		expectedErr       error
	}{
		{
			name:              "pass - successful decode - 1",
			input:             "rDTXLQ7ZKZVKz33zJbHjgVShjsBnqMBhmN",
			expectedPrefix:    []byte{AccountAddressPrefix},
			expectedAccountID: []byte{0x88, 0xa5, 0xa5, 0x7c, 0x82, 0x9f, 0x40, 0xf2, 0x5e, 0xa8, 0x33, 0x85, 0xbb, 0xde, 0x6c, 0x3d, 0x8b, 0x4c, 0xa0, 0x82},
			expectedErr:       nil,
		},
		{
			name:              "pass - successful decode - 2",
			input:             "rJKhsipKHooQbtS3v5Jro6N5Q7TMNPkoAs",
			expectedPrefix:    []byte{AccountAddressPrefix},
			expectedAccountID: []byte{0xbd, 0xe4, 0x2b, 0xbd, 0x77, 0x5b, 0x46, 0x7e, 0x34, 0xfe, 0x48, 0x52, 0xe7, 0xce, 0x3d, 0xd2, 0x61, 0x3, 0xf7, 0x6c},
			expectedErr:       nil,
		},
		{
			name:              "fail - unsuccessful decode - 1",
			input:             "yurt",
			expectedPrefix:    nil,
			expectedAccountID: nil,
			expectedErr:       ErrInvalidClassicAddress,
		},
		{
			name:              "fail - unsuccessful decode - 2",
			input:             "davidschwartz",
			expectedPrefix:    nil,
			expectedAccountID: nil,
			expectedErr:       ErrInvalidClassicAddress,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			typePrefix, accountID, err := DecodeClassicAddressToAccountID(tc.input)

			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
				require.Nil(t, tc.expectedPrefix, typePrefix)
				require.Nil(t, tc.expectedAccountID, accountID)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedPrefix, typePrefix)
				require.Equal(t, tc.expectedAccountID, accountID)
			}
		})
	}
}

func TestEncodeAccountIDToClassicAddress(t *testing.T) {
	testcases := []struct {
		name        string
		input       []byte
		expected    string
		expectedErr error
	}{
		{
			name:        "fail - invalid accountId length",
			input:       []byte{1, 2, 3},
			expectedErr: ErrInvalidAccountID,
		},
		{
			name: "pass - valid accountId",
			input: []byte{
				94, 123, 17, 37, 35, 246,
				141, 47, 94, 135, 157, 180,
				234, 197, 28, 102, 152, 166,
				147, 4,
			},
			expected:    "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := EncodeAccountIDToClassicAddress(tc.input)

			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestIsValidClassicAddress(t *testing.T) {
	tt := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "pass - valid classic address",
			input:    "rDTXLQ7ZKZVKz33zJbHjgVShjsBnqMBhmN",
			expected: true,
		},
		{
			name:     "fail - invalid classic address",
			input:    "yurt",
			expected: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expected != true {
				require.False(t, IsValidClassicAddress(tc.input))
			} else {
				require.True(t, IsValidClassicAddress(tc.input))
			}
		})
	}
}

func TestEncodeNodePublicKey(t *testing.T) {
	tt := []struct {
		name           string
		input          []byte
		expectedOutput string
		expectedErr    error
	}{
		{
			name:           "pass - successful encode",
			input:          []byte{0x3, 0x5f, 0x6d, 0xdb, 0xd6, 0xaf, 0xc5, 0xf2, 0xcb, 0x3d, 0x7d, 0x8, 0x0, 0x55, 0x77, 0x58, 0xdc, 0xc9, 0x2a, 0xc5, 0x29, 0x2d, 0x5d, 0x4f, 0x36, 0x68, 0x31, 0x52, 0x69, 0x19, 0x3e, 0x59, 0xea},
			expectedOutput: "n9MDGCfimuyCmKXUAMcR12rv39PE6PY5YfFpNs75ZjtY3UWt31td",
			expectedErr:    nil,
		},
		{
			name:           "fail - length error",
			input:          []byte{0x00},
			expectedOutput: "",
			expectedErr:    &EncodeLengthError{Instance: "NodePublicKey", Expected: NodePublicKeyLength, Input: 1},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res, err := EncodeNodePublicKey(tc.input)

			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, res)
			}
		})
	}
}

func TestDecodeNodePublicKey(t *testing.T) {
	tt := []struct {
		name           string
		input          string
		expectedOutput []byte
		expectedErr    error
	}{
		{
			name:           "pass - successful decode",
			input:          "n9MDGCfimuyCmKXUAMcR12rv39PE6PY5YfFpNs75ZjtY3UWt31td",
			expectedOutput: []byte{0x3, 0x5f, 0x6d, 0xdb, 0xd6, 0xaf, 0xc5, 0xf2, 0xcb, 0x3d, 0x7d, 0x8, 0x0, 0x55, 0x77, 0x58, 0xdc, 0xc9, 0x2a, 0xc5, 0x29, 0x2d, 0x5d, 0x4f, 0x36, 0x68, 0x31, 0x52, 0x69, 0x19, 0x3e, 0x59, 0xea},
			expectedErr:    nil,
		},
		{
			name:           "fail - length error",
			input:          "rfZG9pC1cKF7q96TNZR264H9ykzKCxMyk44ZK8hFL8cNv1G3c8J",
			expectedOutput: nil,
			expectedErr:    errors.New("b58string prefix and typeprefix not equal"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res, err := DecodeNodePublicKey(tc.input)

			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, res)
			}
		})
	}
}

func TestEncodeAccountPublicKey(t *testing.T) {
	tt := []struct {
		name           string
		input          []byte
		expectedOutput string
		expectedErr    error
	}{
		{
			name:           "pass - successful encode",
			input:          []byte{0xed, 0x94, 0x34, 0x79, 0x92, 0x26, 0x37, 0x49, 0x26, 0xed, 0xa3, 0xb5, 0x4b, 0x1b, 0x46, 0x1b, 0x4a, 0xbf, 0x72, 0x37, 0x96, 0x2e, 0xae, 0x18, 0x52, 0x8f, 0xea, 0x67, 0x59, 0x53, 0x97, 0xfa, 0x32},
			expectedOutput: "aKEt5wr2oXW5H55Z4m94ioKb1Drmj42UWoQDvFJZ5LaxPv126G9d",
			expectedErr:    nil,
		},
		{
			name:           "fail - length error",
			input:          []byte{0x00},
			expectedOutput: "",
			expectedErr:    &EncodeLengthError{Instance: "AccountPublicKey", Expected: AccountPublicKeyLength, Input: 1},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res, err := EncodeAccountPublicKey(tc.input)

			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, res)
			}
		})
	}
}

func TestDecodeAccountPublicKey(t *testing.T) {
	tt := []struct {
		name        string
		input       string
		output      []byte
		expectedErr error
	}{
		{
			name:        "pass - successful decode",
			input:       "aKEt5wr2oXW5H55Z4m94ioKb1Drmj42UWoQDvFJZ5LaxPv126G9d",
			output:      []byte{0xed, 0x94, 0x34, 0x79, 0x92, 0x26, 0x37, 0x49, 0x26, 0xed, 0xa3, 0xb5, 0x4b, 0x1b, 0x46, 0x1b, 0x4a, 0xbf, 0x72, 0x37, 0x96, 0x2e, 0xae, 0x18, 0x52, 0x8f, 0xea, 0x67, 0x59, 0x53, 0x97, 0xfa, 0x32},
			expectedErr: nil,
		},
		{
			name:        "fail - length error",
			input:       "nHU75pVH2Tak7adBWNP3H2CU3wcUtSgf45sKrd1uGyFyRcTozXNm",
			output:      nil,
			expectedErr: errors.New("b58string prefix and typeprefix not equal"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res, err := DecodeAccountPublicKey(tc.input)

			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.output, res)
			}
		})
	}
}

func TestIsValidAddress(t *testing.T) {
	testcases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "fail - invalid classic address",
			input:    "yurt",
			expected: false,
		},
		{
			name:     "pass - valid classic address",
			input:    "rDTXLQ7ZKZVKz33zJbHjgVShjsBnqMBhmN",
			expected: true,
		},

		{
			name:     "pass - valid x address",
			input:    "X7AcgcsBL6XDcUb289X4mJ8djcdyKaB5hJDWMArnXr61cqZ",
			expected: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, IsValidAddress(tc.input))
		})
	}
}
