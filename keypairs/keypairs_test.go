package keypairs

import (
	"errors"
	"strings"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/keypairs/interfaces"
	"github.com/Peersyst/xrpl-go/keypairs/testutil"
	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGenerateEncodeSeed(t *testing.T) {
	generatedEntropy := []byte("fakeRandomString")
	legacyOverlongEntropy := []byte("setPasswordOverLen16")
	legacyTruncatedEntropy := legacyOverlongEntropy[:addresscodec.FamilySeedLength]
	rawEntropy := []byte{
		0x10, 0x11, 0x12, 0x13,
		0x14, 0x15, 0x16, 0x17,
		0x18, 0x19, 0x1A, 0x1B,
		0x1C, 0x1D, 0x1E, 0x1F,
	}
	shortEntropy := rawEntropy[:addresscodec.FamilySeedLength-1]
	overlongEntropy := append(append([]byte{}, rawEntropy...), 0x20)

	require.Len(t, generatedEntropy, addresscodec.FamilySeedLength)
	require.Greater(t, len(legacyOverlongEntropy), addresscodec.FamilySeedLength)
	require.Len(t, legacyTruncatedEntropy, addresscodec.FamilySeedLength)
	require.Len(t, rawEntropy, addresscodec.FamilySeedLength)
	require.Len(t, shortEntropy, addresscodec.FamilySeedLength-1)
	require.Len(t, overlongEntropy, addresscodec.FamilySeedLength+1)

	randomizerErr := errors.New("error")

	tt := []struct {
		name        string
		entropy     []byte
		malleate    func() interfaces.Randomizer
		algorithm   interfaces.KeypairCryptoAlg
		expected    string
		expectedErr error
	}{
		{
			name: "fail - generate bytes error",
			malleate: func() interfaces.Randomizer {
				rand := testutil.NewMockRandomizer(gomock.NewController(t))
				rand.EXPECT().GenerateBytes(addresscodec.FamilySeedLength).Times(1).Return(nil, randomizerErr)
				return rand
			},
			expectedErr: randomizerErr,
			algorithm:   crypto.ED25519(),
		},
		{
			name:    "pass - nil entropy should generate random seed (ED25519)",
			entropy: nil,
			malleate: func() interfaces.Randomizer {
				rand := testutil.NewMockRandomizer(gomock.NewController(t))
				rand.EXPECT().GenerateBytes(addresscodec.FamilySeedLength).Times(1).Return(generatedEntropy, nil)
				return rand
			},
			algorithm:   crypto.ED25519(),
			expected:    "sEdTjrdnJaPE2NNjmavQqXQdrf71NiH",
			expectedErr: nil,
		},
		{
			name:        "pass - raw 16-byte entropy (ED25519)",
			entropy:     rawEntropy,
			algorithm:   crypto.ED25519(),
			expected:    "sEdSXGRS5wtAcH33J9e4H7E78vue4iK",
			expectedErr: nil,
		},
		{
			name:        "pass - raw 16-byte entropy (SECP256K1)",
			entropy:     rawEntropy,
			algorithm:   crypto.SECP256K1(),
			expected:    "spvHRYpBKVWy8aYvjDrEJxG79mwN3",
			expectedErr: nil,
		},
		{
			name:        "pass - manually truncated legacy entropy keeps previous seed (ED25519)",
			entropy:     legacyTruncatedEntropy,
			algorithm:   crypto.ED25519(),
			expected:    "sEdTuXdrgQobjDidph2oMDN36jGZX2U",
			expectedErr: nil,
		},
		{
			name:        "fail - raw entropy below family seed length",
			entropy:     shortEntropy,
			algorithm:   crypto.ED25519(),
			expectedErr: ErrInvalidEntropyLength,
		},
		{
			name:    "pass - empty entropy should generate random seed (SECP256K1)",
			entropy: []byte{},
			malleate: func() interfaces.Randomizer {
				rand := testutil.NewMockRandomizer(gomock.NewController(t))
				rand.EXPECT().GenerateBytes(addresscodec.FamilySeedLength).Times(1).Return(generatedEntropy, nil)
				return rand
			},
			algorithm:   crypto.SECP256K1(),
			expected:    "sh3pdwcaoo7vt5rtrEZJ7a75LnDo3",
			expectedErr: nil,
		},
		{
			name:        "pass - manually truncated legacy entropy keeps previous seed (SECP256K1)",
			entropy:     legacyTruncatedEntropy,
			algorithm:   crypto.SECP256K1(),
			expected:    "shJYdazRN9dvWbGqCehzHcBKWBaFR",
			expectedErr: nil,
		},
		{
			name:        "fail - raw entropy above family seed length",
			entropy:     overlongEntropy,
			algorithm:   crypto.SECP256K1(),
			expectedErr: ErrInvalidEntropyLength,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var randomizer interfaces.Randomizer
			if tc.malleate != nil {
				randomizer = tc.malleate()
			}
			a, err := GenerateSeed(tc.entropy, tc.algorithm, randomizer)

			if tc.expectedErr != nil {
				require.Empty(t, a)
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, a)
			}
		})
	}
}

type unsupportedCryptoAlgorithm struct{}

func (unsupportedCryptoAlgorithm) DeriveKeypair([]byte, bool) (string, string, error) {
	return "", "", nil
}

func (unsupportedCryptoAlgorithm) Sign(string, string) (string, error) {
	return "", nil
}

func (unsupportedCryptoAlgorithm) Validate(string, string, string) bool {
	return false
}

func TestGenerateSeedDoesNotWrapUnsupportedAlgorithmAsEntropyLength(t *testing.T) {
	rawEntropy := []byte{
		0x10, 0x11, 0x12, 0x13,
		0x14, 0x15, 0x16, 0x17,
		0x18, 0x19, 0x1A, 0x1B,
		0x1C, 0x1D, 0x1E, 0x1F,
	}

	tt := []struct {
		name      string
		algorithm interfaces.KeypairCryptoAlg
	}{
		{
			name: "nil algorithm",
		},
		{
			name:      "unsupported algorithm",
			algorithm: unsupportedCryptoAlgorithm{},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			seed, err := GenerateSeed(rawEntropy, tc.algorithm, nil)

			require.Empty(t, seed)
			require.Error(t, err)
			require.NotErrorIs(t, err, ErrInvalidEntropyLength)
			require.EqualError(t, err, "encoding type must be `ed25519` or `secp256k1`")
		})
	}
}

func TestDeriveKeypair(t *testing.T) {
	tt := []struct {
		name           string
		inputSeed      string
		inputValidator bool
		pubKey         string
		privKey        string
		expectedErr    error
	}{
		{
			name:           "fail - invalid seed",
			inputSeed:      "sSensitiveSeedMaterial",
			inputValidator: false,
			expectedErr:    addresscodec.ErrInvalidSeed,
		},
		{
			name:           "fail - invalid seed length",
			inputSeed:      addresscodec.Base58CheckEncode(nil, addresscodec.FamilySeedPrefix),
			inputValidator: false,
			expectedErr:    addresscodec.ErrInvalidSeedLength,
		},
		{
			name:           "fail - invalid seed prefix",
			inputSeed:      addresscodec.Base58CheckEncode([]byte("random"), 0x22),
			inputValidator: false,
			expectedErr:    addresscodec.ErrInvalidSeedPrefix,
		},
		{
			name:           "fail - invalid ED25519 key",
			inputSeed:      "ED4924A9045FE5ED8B22BAA7B6229A72A287CCF3EA287AADD3A032A24C0F008F",
			inputValidator: false,
			expectedErr:    addresscodec.ErrInvalidSeed,
		},
		{
			name:           "pass - derive an ED25519 keypair",
			inputSeed:      "sEdTjrdnJaPE2NNjmavQqXQdrf71NiH",
			inputValidator: false,
			pubKey:         testEdPublicKey,
			privKey:        testEdPrivateKey,
			expectedErr:    nil,
		},
		{
			name:           "pass - derive an SECP256K1 keypair",
			inputSeed:      "sh3pdwcaoo7vt5rtrEZJ7a75LnDo3",
			inputValidator: false,
			pubKey:         "03A947D71477652C445B20F5226FAA4DF6CD716786E17D016E9A37FBA5379AF02B",
			privKey:        "00204795BCAB502D01C06B2C700936204B26C58D7048D3D4DBFE890BA05BA1D68D",
			expectedErr:    nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			priv, pub, err := DeriveKeypair(tc.inputSeed, tc.inputValidator)

			if tc.expectedErr != nil {
				require.Empty(t, pub)
				require.Empty(t, priv)
				require.ErrorIs(t, err, tc.expectedErr)
				require.NotContains(t, err.Error(), tc.inputSeed)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.pubKey, pub)
				require.Equal(t, tc.privKey, priv)
			}
		})
	}
}

func TestDeriveClassicAddress(t *testing.T) {
	tt := []struct {
		name        string
		input       string
		expected    string
		expectedErr error
	}{
		{
			name:     "pass - derive address from ED25519 public key",
			input:    testEdPublicKey,
			expected: "rhBtDFHj2EiBqXFDobmjxysCHa3ngd6dbX",
		},
		{
			name:     "pass - derive address from compressed secp256k1 public key with even Y",
			input:    testSecpCompressedEvenKey,
			expected: "rnWL65ZQpCTYUZvEKCWprKk1bRugwkf261",
		},
		{
			name:     "pass - derive address from compressed secp256k1 public key with odd Y",
			input:    testSecpCompressedOddKey,
			expected: "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
		},
		{
			name:     "pass - derive address from uncompressed secp256k1 public key",
			input:    testSecpUncompressedKey,
			expected: "rH6TMqaYe6A2A7eXeih1tW228o7P6dNjKz",
		},
		{
			name:        "fail - invalid compressed secp256k1 curve point",
			input:       "02" + strings.Repeat("FF", 32),
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "fail - invalid uncompressed secp256k1 curve point",
			input:       "04" + strings.Repeat("FF", 64),
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "fail - malformed public key",
			input:       testEdPublicKey[:len(testEdPublicKey)-1] + "Z",
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "fail - truncated public key",
			input:       testSecpCompressedEvenKey[:len(testSecpCompressedEvenKey)-2],
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "fail - oversized public key",
			input:       testSecpUncompressedKey + "00",
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "fail - private key type mismatch",
			input:       testSecpPrefixedPrivateKey,
			expectedErr: ErrInvalidPublicKeyFormat,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := DeriveClassicAddress(tc.input)
			if tc.expectedErr != nil {
				require.Empty(t, actual)
				require.ErrorIs(t, err, tc.expectedErr)
				require.NotContains(t, err.Error(), tc.input)
				if len(tc.input) >= 18 {
					require.NotContains(t, err.Error(), tc.input[2:18])
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}

// secpHelloWorldSignature is the deterministic secp256k1 signature of "Hello World" by
// testSecpRawPrivateKey (public key testSecpCompressedEvenKey). TestSign asserts it is
// produced and TestValidate asserts it verifies.
const secpHelloWorldSignature = "3045022100E1617F1A3C85B5BC8FA6224F893FE9068BEA8F8D075EE144F6F9D255C829761802206FD9B361CDE83A0C3D5654232F1D7CFB1A614E9A8F9B1A861564029065516E64"

// secpHelloWorldHighSSignature is the malleable high-S alternative to
// secpHelloWorldSignature. XRPL requires the low-S form.
const secpHelloWorldHighSSignature = "3046022100E1617F1A3C85B5BC8FA6224F893FE9068BEA8F8D075EE144F6F9D255C829761802210090264C9E3217C5F3C2A9ABDCD0E28303A04D8E4C1FAD85B5AA6E5BFC6AE4D2DD"

func TestSign(t *testing.T) {
	tt := []struct {
		name         string
		inputMsg     string
		inputPrivKey string
		expected     string
		expectedErr  error
	}{
		{
			name:        "fail - empty private key",
			inputMsg:    "hello world",
			expectedErr: ErrInvalidPrivateKeyFormat,
		},
		{
			name:         "fail - short private key",
			inputMsg:     "hello world",
			inputPrivKey: "E",
			expectedErr:  ErrInvalidPrivateKeyFormat,
		},
		{
			name:         "fail - malformed private key",
			inputMsg:     "hello world",
			inputPrivKey: testSecpRawPrivateKey[:len(testSecpRawPrivateKey)-1] + "Z",
			expectedErr:  ErrInvalidPrivateKeyFormat,
		},
		{
			name:         "fail - truncated ED25519 private key",
			inputMsg:     "hello world",
			inputPrivKey: testEdPrivateKey[:len(testEdPrivateKey)-4],
			expectedErr:  ErrInvalidPrivateKeyFormat,
		},
		{
			name:         "fail - oversized ED25519 private key",
			inputMsg:     "hello world",
			inputPrivKey: testEdPrivateKey + "00",
			expectedErr:  ErrInvalidPrivateKeyFormat,
		},
		{
			name:         "fail - compressed public key type mismatch",
			inputMsg:     "hello world",
			inputPrivKey: testSecpCompressedEvenKey,
			expectedErr:  ErrInvalidPrivateKeyFormat,
		},
		{
			name:         "fail - uncompressed public key type mismatch",
			inputMsg:     "hello world",
			inputPrivKey: testSecpUncompressedKey,
			expectedErr:  ErrInvalidPrivateKeyFormat,
		},
		{
			name:         "pass - sign with ED25519 private key",
			inputMsg:     "hello world",
			inputPrivKey: testEdPrivateKey,
			expected:     "E83CAFEAF100793F0C6570D60C7447FF3A87E0DC0CAE9AD90EF0102860EC3BD1D20F432494021F3E19DAFF257A420CA64A49C283AB5AD00B6B0CEA1756151C01",
		},
		{
			name:         "pass - sign with raw secp256k1 private key",
			inputMsg:     "Hello World",
			inputPrivKey: testSecpRawPrivateKey,
			expected:     secpHelloWorldSignature,
		},
		{
			name:         "pass - sign with 00-prefixed secp256k1 private key",
			inputMsg:     "Hello World",
			inputPrivKey: testSecpPrefixedPrivateKey,
			expected:     secpHelloWorldSignature,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := Sign(tc.inputMsg, tc.inputPrivKey)
			if tc.expectedErr != nil {
				require.Empty(t, actual)
				require.ErrorIs(t, err, tc.expectedErr)
				require.ErrorIs(t, err, ErrInvalidCryptoImplementation)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestSignErrorsRedactPrivateKeyMaterial(t *testing.T) {
	testcases := []struct {
		name  string
		input string
	}{
		{
			name:  "malformed raw key",
			input: testSecpRawPrivateKey[:len(testSecpRawPrivateKey)-1] + "Z",
		},
		{
			name:  "unsupported prefixed key",
			input: "01" + testSecpRawPrivateKey,
		},
		{
			name:  "oversized ED25519 key",
			input: testEdPrivateKey + "00",
		},
		{
			name:  "compressed public key",
			input: testSecpCompressedEvenKey,
		},
		{
			name:  "uncompressed public key",
			input: testSecpUncompressedKey,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Sign("redaction test", tc.input)

			require.ErrorIs(t, err, ErrInvalidPrivateKeyFormat)
			require.NotContains(t, err.Error(), tc.input)
			require.NotContains(t, err.Error(), tc.input[:2])
			require.NotContains(t, err.Error(), tc.input[2:18])
		})
	}
}

func TestValidate(t *testing.T) {
	const (
		edSignature      = "C001CB8A9883497518917DD16391930F4FEE39CEA76C846CFF4330BA44ED19DC4730056C2C6D7452873DE8120A5023C6807135C6329A89A13BA1D476FE8E7100"
		secpOddSignature = "30440220583A91C95E54E6A651C47BEC22744E0B101E2C4060E7B08F6341657DAD9BC3EE02207D1489C7395DB0188D3A56A977ECBA54B36FA9371B40319655B1B4429E33EF2D"
	)

	tt := []struct {
		name        string
		inputMsg    string
		inputPubKey string
		inputSig    string
		expected    bool
		expectedErr error
	}{
		{
			name:        "fail - empty public key",
			inputMsg:    "test message",
			inputSig:    "invalid",
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "fail - short public key",
			inputMsg:    "test message",
			inputPubKey: "E",
			inputSig:    "invalid",
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "fail - malformed full-length public key",
			inputMsg:    "test message",
			inputPubKey: testEdPublicKey[:len(testEdPublicKey)-1] + "Z",
			inputSig:    "invalid",
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "fail - truncated public key",
			inputMsg:    "test message",
			inputPubKey: testSecpCompressedOddKey[:len(testSecpCompressedOddKey)-2],
			inputSig:    "invalid",
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "fail - oversized public key",
			inputMsg:    "test message",
			inputPubKey: testSecpUncompressedKey + "00",
			inputSig:    "invalid",
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "fail - raw private key type mismatch",
			inputMsg:    "test message",
			inputPubKey: testSecpRawPrivateKey,
			inputSig:    "invalid",
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "fail - prefixed private key type mismatch",
			inputMsg:    "test message",
			inputPubKey: testSecpPrefixedPrivateKey,
			inputSig:    "invalid",
			expectedErr: ErrInvalidPublicKeyFormat,
		},
		{
			name:        "fail - invalid signature with valid public key",
			inputMsg:    "test message",
			inputPubKey: testEdPublicKey,
			inputSig:    "invalid",
			expected:    false,
		},
		{
			name:        "pass - verify with ED25519 public key",
			inputMsg:    "test message",
			inputPubKey: testEdPublicKey,
			inputSig:    edSignature,
			expected:    true,
		},
		{
			name:        "pass - verify with compressed secp256k1 public key with even Y",
			inputMsg:    "Hello World",
			inputPubKey: testSecpCompressedEvenKey,
			inputSig:    secpHelloWorldSignature,
			expected:    true,
		},
		{
			name:        "fail - reject high-S secp256k1 signature",
			inputMsg:    "Hello World",
			inputPubKey: testSecpCompressedEvenKey,
			inputSig:    secpHelloWorldHighSSignature,
			expected:    false,
		},
		{
			name:        "pass - verify with compressed secp256k1 public key with odd Y",
			inputMsg:    "test message",
			inputPubKey: testSecpCompressedOddKey,
			inputSig:    secpOddSignature,
			expected:    true,
		},
		{
			name:        "pass - verify with uncompressed secp256k1 public key",
			inputMsg:    "Hello World",
			inputPubKey: testSecpUncompressedKey,
			inputSig:    secpHelloWorldSignature,
			expected:    true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := Validate(tc.inputMsg, tc.inputPubKey, tc.inputSig)
			if tc.expectedErr != nil {
				require.False(t, actual)
				require.ErrorIs(t, err, tc.expectedErr)
				require.ErrorIs(t, err, ErrInvalidCryptoImplementation)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestDeriveNodeAddress(t *testing.T) {
	testcases := []struct {
		name        string
		inputPubKey string
		expected    string
		expectedErr error
	}{
		{
			name:        "fail - derive node address - input too short to base58check-decode",
			inputPubKey: "x",
			expectedErr: addresscodec.ErrInvalidFormat,
		},
		{
			name:        "fail - derive node address - node prefix mismatch",
			inputPubKey: "rfZG9pC1cKF7q96TNZR264H9ykzKCxMyk44ZK8hFL8cNv1G3c8J",
			expectedErr: addresscodec.ErrB58PrefixMismatch,
		},
		{
			name:        "pass - derive correct node address from public key",
			inputPubKey: "n9KHn8NfbBsZV5q8bLfS72XyGqwFt5mgoPbcTV4c6qKiuPTAtXYk",
			expected:    "rU7bM9ENDkybaxNrefAVjdLTyNLuue1KaJ",
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := DeriveNodeAddress(tc.inputPubKey, crypto.SECP256K1())
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}
