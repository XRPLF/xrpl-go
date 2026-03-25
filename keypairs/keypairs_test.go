package keypairs

import (
	"errors"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/keypairs/interfaces"
	"github.com/Peersyst/xrpl-go/keypairs/testutil"
	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGenerateEncodeSeed(t *testing.T) {
	defaultEntropy := "fakeRandomString"

	tt := []struct {
		name        string
		entropy     string
		malleate    func() interfaces.Randomizer
		algorithm   interfaces.KeypairCryptoAlg
		expected    string
		expectedErr error
	}{
		{
			name: "fail - generate bytes error",
			malleate: func() interfaces.Randomizer {
				rand := testutil.NewMockRandomizer(gomock.NewController(t))
				rand.EXPECT().GenerateBytes(gomock.Any()).AnyTimes().Return(nil, errors.New("error"))
				return rand
			},
			expectedErr: errors.New("error"),
			algorithm:   crypto.ED25519(),
		},
		{
			name:    "pass - empty entropy should generate random seed (ED25519)",
			entropy: "",
			malleate: func() interfaces.Randomizer {
				rand := testutil.NewMockRandomizer(gomock.NewController(t))
				rand.EXPECT().GenerateBytes(gomock.Any()).AnyTimes().Return([]byte(defaultEntropy), nil)
				return rand
			},
			algorithm:   crypto.ED25519(),
			expected:    "sEdTjrdnJaPE2NNjmavQqXQdrf71NiH",
			expectedErr: nil,
		},
		{
			name:    "pass - entropy defined and above family seed length (ED25519)",
			entropy: "setPasswordOverLen16",
			malleate: func() interfaces.Randomizer {
				rand := testutil.NewMockRandomizer(gomock.NewController(t))
				rand.EXPECT().GenerateBytes(gomock.Any()).AnyTimes().Return([]byte("setPasswordOverLen16"), nil)
				return rand
			},
			algorithm:   crypto.ED25519(),
			expected:    "sEdTuXdrgQobjDidph2oMDN36jGZX2U",
			expectedErr: nil,
		},
		{
			name:    "pass - empty entropy should generate random seed (SECP256K1)",
			entropy: "",
			malleate: func() interfaces.Randomizer {
				rand := testutil.NewMockRandomizer(gomock.NewController(t))
				rand.EXPECT().GenerateBytes(gomock.Any()).AnyTimes().Return([]byte(defaultEntropy), nil)
				return rand
			},
			algorithm:   crypto.SECP256K1(),
			expected:    "sh3pdwcaoo7vt5rtrEZJ7a75LnDo3",
			expectedErr: nil,
		},
		{
			name:    "pass - entropy defined and above family seed length (SECP256K1)",
			entropy: "setPasswordOverLen16",
			malleate: func() interfaces.Randomizer {
				rand := testutil.NewMockRandomizer(gomock.NewController(t))
				rand.EXPECT().GenerateBytes(gomock.Any()).AnyTimes().Return([]byte("setPasswordOverLen16"), nil)
				return rand
			},
			algorithm:   crypto.SECP256K1(),
			expected:    "shJYdazRN9dvWbGqCehzHcBKWBaFR",
			expectedErr: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			randomizer := tc.malleate()
			a, err := GenerateSeed(tc.entropy, tc.algorithm, randomizer)

			if tc.expectedErr != nil {
				require.Empty(t, a)
				require.Error(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, a)
			}
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
			inputSeed:      "invalid",
			inputValidator: false,
			expectedErr:    addresscodec.ErrInvalidSeed,
		},
		{
			name:           "fail - invalid ED25519 key",
			inputSeed:      "ED4924A9045FE5ED8B22BAA7B6229A72A287CCF3EA287AADD3A032A24C0F008F",
			inputValidator: false,
			expectedErr:    ErrInvalidCryptoImplementation,
		},
		{
			name:           "pass - derive an ED25519 keypair",
			inputSeed:      "sEdTjrdnJaPE2NNjmavQqXQdrf71NiH",
			inputValidator: false,
			pubKey:         "ED4924A9045FE5ED8B22BAA7B6229A72A287CCF3EA287AADD3A032A24C0F008FA6",
			privKey:        "EDBB3ECA8985E1484FA6A28C4B30FB0042A2CC5DF3EC8DC37B5F3D126DDFD3CA14",
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
				require.Error(t, err, tc.expectedErr.Error())
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
			name:        "pass - derive correct address from public key",
			input:       "ED731C39781B964904E1FEEFFC9F99442196BCB5F499105A79533E2D678CA7D3D2",
			expected:    "rhTCnDC7v1Jp7NAupzisv6ynWHD161Q9nV",
			expectedErr: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := DeriveClassicAddress(tc.input)
			if tc.expectedErr != nil {
				require.Empty(t, actual)
				require.Error(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestSign(t *testing.T) {
	tt := []struct {
		name         string
		inputMsg     string
		inputPrivKey string
		expected     string
		expectedErr  error
	}{
		{
			name:         "fail - invalid private key",
			inputMsg:     "hello world",
			inputPrivKey: "invalid",
			expectedErr:  ErrInvalidCryptoImplementation,
		},
		{
			name:         "pass - sign a message with a ED25519 key",
			inputMsg:     "hello world",
			inputPrivKey: "EDBB3ECA8985E1484FA6A28C4B30FB0042A2CC5DF3EC8DC37B5F3D126DDFD3CA14",
			expected:     "E83CAFEAF100793F0C6570D60C7447FF3A87E0DC0CAE9AD90EF0102860EC3BD1D20F432494021F3E19DAFF257A420CA64A49C283AB5AD00B6B0CEA1756151C01",
			expectedErr:  nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := Sign(tc.inputMsg, tc.inputPrivKey)
			if tc.expectedErr != nil {
				require.Empty(t, actual)
				require.Error(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tt := []struct {
		name        string
		inputMsg    string
		inputPubKey string
		inputSig    string
		expected    bool
		expectedErr error
	}{
		{
			name:        "fail - invalid public key",
			inputMsg:    "test message",
			inputPubKey: "invalid",
			inputSig:    "invalid",
			expectedErr: ErrInvalidCryptoImplementation,
		},
		{
			name:        "pass - valid message with ED25519 key",
			inputMsg:    "test message",
			inputPubKey: "ED4924A9045FE5ED8B22BAA7B6229A72A287CCF3EA287AADD3A032A24C0F008FA6",
			inputSig:    "C001CB8A9883497518917DD16391930F4FEE39CEA76C846CFF4330BA44ED19DC4730056C2C6D7452873DE8120A5023C6807135C6329A89A13BA1D476FE8E7100",
			expected:    true,
			expectedErr: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := Validate(tc.inputMsg, tc.inputPubKey, tc.inputSig)
			if tc.expectedErr != nil {
				require.Zero(t, actual)
				require.Error(t, err, tc.expectedErr.Error())
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
			name:        "fail - derive node address - node prefix not found",
			inputPubKey: "x9KHn8NfbBsZV5q8bLfS72XyGqwFt5mgoPbcTV4c6qKiuPTAtXYk",
			expectedErr: &addresscodec.EncodeLengthError{Instance: "NodePublicKey", Expected: addresscodec.NodePublicKeyLength, Input: 3},
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
				require.Error(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}
