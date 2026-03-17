package crypto

import (
	"errors"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/stretchr/testify/require"
)

func TestSecp256k1_Prefix(t *testing.T) {
	require.Equal(t, secp256K1Prefix, SECP256K1().Prefix())
}

func TestSecp256k1_FamilySeedPrefix(t *testing.T) {
	require.Equal(t, secp256K1FamilySeedPrefix, SECP256K1().FamilySeedPrefix())
}

func TestSecp256k1_deriveKeypair(t *testing.T) {
	testCases := []struct {
		name            string
		seedBytes       []byte
		validator       bool
		expectedPrivKey string
		expectedPubKey  string
		expectedErr     error
	}{
		{
			name:            "pass - valid seed 1",
			seedBytes:       []byte{229, 81, 182, 134, 131, 220, 192, 126, 133, 114, 150, 132, 140, 237, 222, 196},
			validator:       false,
			expectedPubKey:  "02950F4710101A25073BF37086D73FBBD00C7A6B0F91097D8F0BC6D268C400D56E",
			expectedPrivKey: "00B167A9F3B9E60A4F93695713682C102438620AA1785C3AE635F53E5B6261071A",
			expectedErr:     nil,
		},
		{
			name:            "pass - valid seed 2",
			seedBytes:       []byte{124, 228, 51, 247, 54, 54, 81, 51, 239, 86, 226, 187, 232, 20, 111, 163},
			validator:       false,
			expectedPubKey:  "031FBCFDD2EC6C2EDFBBA3866BDBAC28E5253C6A01FE9EFF8CAAE01871F009E837",
			expectedPrivKey: "00A3D1513DBE784107428B363A1F8EAF1377AB63D4D137AB9E28E0BC614C71D8C0",
			expectedErr:     nil,
		},
		{
			name:            "pass - valid seed 3",
			seedBytes:       []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			validator:       false,
			expectedPrivKey: "00D78B9735C3F26501C7337B8A5727FD53A6EFDBC6AA55984F098488561F985E23",
			expectedPubKey:  "030D58EB48B4420B1F7B9DF55087E0E29FEF0E8468F9A6825B01CA2C361042D435",
			expectedErr:     nil,
		},
		{
			name:            "fail - validator set to true",
			seedBytes:       []byte{124, 228, 51, 247, 54, 54, 81, 51, 239, 86, 226, 187, 232, 20, 111, 163},
			validator:       true,
			expectedPubKey:  "",
			expectedPrivKey: "",
			expectedErr:     ErrValidatorKeypairDerivation,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			privKey, pubKey, err := SECP256K1().DeriveKeypair(tc.seedBytes, tc.validator)
			if tc.expectedErr != nil {
				require.Error(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedPrivKey, privKey)
				require.Equal(t, tc.expectedPubKey, pubKey)
			}
		})
	}
}

func TestSecp256k1_Sign(t *testing.T) {
	testCases := []struct {
		name              string
		message           string
		privKey           string
		expectedSignature string
		wantErr           bool
	}{
		{
			name:              "pass - valid message",
			message:           "test message",
			privKey:           "00D78B9735C3F26501C7337B8A5727FD53A6EFDBC6AA55984F098488561F985E23",
			expectedSignature: "30440220583A91C95E54E6A651C47BEC22744E0B101E2C4060E7B08F6341657DAD9BC3EE02207D1489C7395DB0188D3A56A977ECBA54B36FA9371B40319655B1B4429E33EF2D",
			wantErr:           false,
		},
		{
			name:              "pass - valid signature",
			message:           "Hello World",
			privKey:           "00B167A9F3B9E60A4F93695713682C102438620AA1785C3AE635F53E5B6261071A",
			expectedSignature: "3045022100E1617F1A3C85B5BC8FA6224F893FE9068BEA8F8D075EE144F6F9D255C829761802206FD9B361CDE83A0C3D5654232F1D7CFB1A614E9A8F9B1A861564029065516E64",
			wantErr:           false,
		},
		{
			name:              "pass - valid signature 2",
			message:           "test",
			privKey:           "00B167A9F3B9E60A4F93695713682C102438620AA1785C3AE635F53E5B6261071A",
			expectedSignature: "30450221008F0B50BEEA0C9787E85EEF9624E9385CCBE565B221BAEC2F2DA5F1D9D6D976F7022022C1B1829AE0E758FB690110F245F15433A0579C44910785FE75F93B9D0FB41F",
			wantErr:           false,
		},
		{
			name:              "pass - valid signature 3",
			message:           "message",
			privKey:           "00B167A9F3B9E60A4F93695713682C102438620AA1785C3AE635F53E5B6261071A",
			expectedSignature: "3045022100F07CD8D749AAD8F972475A34591336162A959FCC7F8E692D56410CB70B9634F702201B96AF63E166371D8A2C4C3D4CDA69F6064212D1C28D01F598653BE05C323E8C",
			wantErr:           false,
		},
		{
			name:              "pass - valid signature 4",
			message:           "message2",
			privKey:           "00B167A9F3B9E60A4F93695713682C102438620AA1785C3AE635F53E5B6261071A",
			expectedSignature: "3045022100A2847849BC186B227DB941B1D0A4C39FABBE04A10BF364FC4E394E8B53FD308D02202D47CA9DC35B7FE3E04B578A935CCBE1827B610911709AC13343344F311BD799",
			wantErr:           false,
		},
		{
			name:              "pass - valid signature 5",
			message:           "message3",
			privKey:           "00B167A9F3B9E60A4F93695713682C102438620AA1785C3AE635F53E5B6261071A",
			expectedSignature: "304402202D5CDBCF251868046CB07FC2CB49200FED9FF216D4B38455A1D222ED29E6123B022057E9962B336D180F0B8DCD99B72C30BB09A5451D2059556E3C1E45C1F5D018B6",
			wantErr:           false,
		},
		{
			name:              "pass - valid signature 6",
			message:           "message4",
			privKey:           "00B167A9F3B9E60A4F93695713682C102438620AA1785C3AE635F53E5B6261071A",
			expectedSignature: "3045022100A07B597B3D61C3E97B3CCC2DB65F40B26BAEEF7A3EAF8969C0F4E879DDAD1314022058296AC8B4A6E2D5F33891B5BB2211D2AEF1853DF42452649865AB2FE2C83922",
			wantErr:           false,
		},
		{
			name:              "pass - valid signature 7",
			message:           "message5",
			privKey:           "00B167A9F3B9E60A4F93695713682C102438620AA1785C3AE635F53E5B6261071A",
			expectedSignature: "3044022033950382A62160DBD731D3108C34B07AFD5CD816943931B64E3A25440E8C911902200ABEF5FB3E8B0C4CBD304421B8D3BD6F135D54831FE5426BE74D340ECDFE1F8F",
			wantErr:           false,
		},
		{
			name:              "fail - empty private key",
			message:           "Hello World",
			privKey:           "",
			expectedSignature: "",
			wantErr:           true,
		},
		{
			name:              "fail - invalid private key",
			message:           "Hello World",
			privKey:           "invalid_key",
			expectedSignature: "",
			wantErr:           true,
		},
		{
			name:              "fail - invalid message length",
			message:           "",
			privKey:           "00B167A9F3B9E60A4F93695713682C102438620AA1785C3AE635F53E5B6261071A",
			expectedSignature: "",
			wantErr:           true,
		},
		{
			name:              "fail - invalid private key hex",
			message:           "Hello World",
			privKey:           "00B167A9F3B9E60A4F93695713682C102438620AA1785C3AE635F53E5B6261071X",
			expectedSignature: "",
			wantErr:           true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			signature, err := SECP256K1().Sign(tc.message, tc.privKey)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedSignature, signature)
			}
		})
	}
}

func TestSecp256k1_Validate(t *testing.T) {
	testCases := []struct {
		name      string
		message   string
		signature string
		pubKey    string
		wantValid bool
	}{
		{
			name:      "pass - valid signature",
			message:   "Hello World",
			signature: "3045022100E1617F1A3C85B5BC8FA6224F893FE9068BEA8F8D075EE144F6F9D255C829761802206FD9B361CDE83A0C3D5654232F1D7CFB1A614E9A8F9B1A861564029065516E64",
			pubKey:    "02950F4710101A25073BF37086D73FBBD00C7A6B0F91097D8F0BC6D268C400D56E",
			wantValid: true,
		},
		{
			name:      "pass - valid signature",
			message:   "test",
			signature: "30450221008F0B50BEEA0C9787E85EEF9624E9385CCBE565B221BAEC2F2DA5F1D9D6D976F7022022C1B1829AE0E758FB690110F245F15433A0579C44910785FE75F93B9D0FB41F",
			pubKey:    "02950F4710101A25073BF37086D73FBBD00C7A6B0F91097D8F0BC6D268C400D56E",
			wantValid: true,
		},
		{
			name:      "pass - valid signature with seed 3 keypair",
			message:   "test message",
			signature: "30440220583A91C95E54E6A651C47BEC22744E0B101E2C4060E7B08F6341657DAD9BC3EE02207D1489C7395DB0188D3A56A977ECBA54B36FA9371B40319655B1B4429E33EF2D",
			pubKey:    "030D58EB48B4420B1F7B9DF55087E0E29FEF0E8468F9A6825B01CA2C361042D435",
			wantValid: true,
		},
		{
			name:      "fail - invalid signature",
			message:   "Hello, World!",
			signature: "3045022100B1629F44BB12A86AE5A3D79A4E2BE6A473DBBD3F4FB4E3898A2E9A9BE1A54EF502204C3B0C33C46F5ABDE7C2C1A3F2B79B8A9F3A69D8C7C248B2B5C16A39A9C3B5F6",
			pubKey:    "02950F4710101A25073BF37086D73FBBD00C7A6B0F91097D8F0BC6D268C400D56E",
			wantValid: false,
		},
		{
			name:      "fail - invalid public key",
			message:   "Hello, World!",
			signature: "3045022100B1629F44BB12A86AE5A3D79A4E2BE6A473DBBD3F4FB4E3898A2E9A9BE1A54EF502204C3B0C33C46F5ABDE7C2C1A3F2B79B8A9F3A69D8C7C248B2B5C16A39A9C3B5F5",
			pubKey:    "invalid_key",
			wantValid: false,
		},
		{
			name:      "fail - empty signature",
			message:   "Hello, World!",
			signature: "",
			pubKey:    "02950F4710101A25073BF37086D73FBBD00C7A6B0F91097D8F0BC6D268C400D56E",
			wantValid: false,
		},
		{
			name:      "fail - invalid public key hex",
			message:   "Hello, World!",
			signature: "3045022100B1629F44BB12A86AE5A3D79A4E2BE6A473DBBD3F4FB4E3898A2E9A9BE1A54EF502204C3B0C33C46F5ABDE7C2C1A3F2B79B8A9F3A69D8C7C248B2B5C16A39A9C3B5F5",
			pubKey:    "02950F4710101A25073BF37086D73FBBD00C7A6B0F91097D8F0BC6D268C400D5",
			wantValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid := SECP256K1().Validate(tc.message, tc.pubKey, tc.signature)
			require.Equal(t, tc.wantValid, isValid)
		})
	}
}

func TestSecp256k1_deriveScalar_doesNotMutateInput(t *testing.T) {
	tests := []struct {
		name    string
		discrim *uint32
	}{
		{
			name:    "nil discrim",
			discrim: nil,
		},
		{
			name:    "discrim=0",
			discrim: func() *uint32 { v := uint32(0); return &v }(),
		},
		{
			name:    "discrim=7",
			discrim: func() *uint32 { v := uint32(7); return &v }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// seed 1
			seed := []byte{229, 81, 182, 134, 131, 220, 192, 126, 133, 114, 150, 132, 140, 237, 222, 196}
			original := make([]byte, len(seed))
			copy(original, seed)

			SECP256K1().deriveScalar(seed, tt.discrim)
			require.Equal(t, original, seed, "deriveScalar must not mutate input bytes")
		})
	}
}

func TestSecp256k1_deriveScalar_discriminatorAffectsOutput(t *testing.T) {
	// seed 1
	seed := []byte{229, 81, 182, 134, 131, 220, 192, 126, 133, 114, 150, 132, 140, 237, 222, 196}

	discrims := []struct {
		name    string
		discrim *uint32
	}{
		{name: "nil", discrim: nil},
		{name: "0", discrim: func() *uint32 { v := uint32(0); return &v }()},
		{name: "1", discrim: func() *uint32 { v := uint32(1); return &v }()},
	}

	scalars := make([]secp256k1.ModNScalar, len(discrims))
	for i, d := range discrims {
		scalars[i] = SECP256K1().deriveScalar(seed, d.discrim)
	}

	for i := 0; i < len(discrims); i++ {
		for j := i + 1; j < len(discrims); j++ {
			t.Run(discrims[i].name+"_vs_"+discrims[j].name, func(t *testing.T) {
				require.NotEqual(t, scalars[i], scalars[j], "different discriminators must produce different scalars")
			})
		}
	}
}

func TestSecp256k1_DerivePublicKeyFromPublicGenerator(t *testing.T) {
	testcases := []struct {
		name        string
		inputPubKey []byte
		expected    []byte
		expectedErr error
	}{
		{
			name:        "fail - derive public key from public generator - invalid input public key",
			inputPubKey: []byte{1, 2, 3},
			expected:    nil,
			expectedErr: errors.New("invalid public key"),
		},
		{
			name:        "pass - derive correct public key from public generator",
			inputPubKey: []byte{2, 96, 177, 143, 143, 27, 242, 159, 10, 244, 101, 28, 252, 88, 117, 180, 216, 33, 99, 169, 245, 4, 160, 213, 193, 34, 255, 255, 181, 74, 233, 165, 154},
			expected:    []byte{3, 142, 217, 120, 94, 231, 252, 104, 116, 69, 224, 217, 64, 101, 167, 79, 246, 206, 198, 80, 106, 3, 199, 56, 0, 117, 216, 26, 43, 158, 126, 134, 129},
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := SECP256K1().DerivePublicKeyFromPublicGenerator(tc.inputPubKey)
			if tc.expectedErr != nil {
				require.Error(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}
