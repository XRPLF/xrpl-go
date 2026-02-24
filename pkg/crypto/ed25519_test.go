package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestED25519_Prefix(t *testing.T) {
	require.Equal(t, ed25519Prefix, ED25519().Prefix())
}

func TestED25519_FamilySeedPrefix(t *testing.T) {
	require.Equal(t, ed25519FamilySeedPrefix, ED25519().FamilySeedPrefix())
}

func TestED25519DeriveKeypair(t *testing.T) {
	tt := []struct {
		name       string
		seedBytes  []byte
		validator  bool
		expPubKey  string
		expPrivKey string
		expErr     error
	}{
		{
			name:       "pass- successfully derive keypair",
			seedBytes:  []byte{102, 97, 107, 101, 82, 97, 110, 100, 111, 109, 83, 116, 114, 105, 110, 103},
			validator:  false,
			expPubKey:  "ED4924A9045FE5ED8B22BAA7B6229A72A287CCF3EA287AADD3A032A24C0F008FA6",
			expPrivKey: "EDBB3ECA8985E1484FA6A28C4B30FB0042A2CC5DF3EC8DC37B5F3D126DDFD3CA14",
			expErr:     nil,
		},
		{
			name:       "fail - error if validator is set to true",
			seedBytes:  []byte{102, 97, 107, 101, 82, 97, 110, 100, 111, 109, 83, 116, 114, 105, 110, 103},
			validator:  true,
			expPubKey:  "",
			expPrivKey: "",
			expErr:     ErrValidatorNotSupported,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			priv, pub, err := ED25519().DeriveKeypair(tc.seedBytes, tc.validator)
			if tc.expErr != nil {
				require.Zero(t, pub)
				require.Zero(t, priv)
				require.Error(t, err, tc.expErr.Error())
			} else {
				require.Equal(t, tc.expPrivKey, priv)
				require.Equal(t, tc.expPubKey, pub)
			}
		})
	}
}

func TestED25519Sign(t *testing.T) {
	tt := []struct {
		name         string
		inputMsg     string
		inputPrivKey string
		expected     string
		expectedErr  error
	}{
		{
			name:         "fail - invalid private key hex",
			inputMsg:     "hello world",
			inputPrivKey: "invalid_key",
			expected:     "",
			expectedErr:  ErrInvalidPrivateKey,
		},
		{
			name:         "pass - sign a valid message",
			inputMsg:     "hello world",
			inputPrivKey: "EDBB3ECA8985E1484FA6A28C4B30FB0042A2CC5DF3EC8DC37B5F3D126DDFD3CA14",
			expected:     "E83CAFEAF100793F0C6570D60C7447FF3A87E0DC0CAE9AD90EF0102860EC3BD1D20F432494021F3E19DAFF257A420CA64A49C283AB5AD00B6B0CEA1756151C01",
			expectedErr:  nil,
		},
		{
			name:         "pass - sign a message with a different private key",
			inputMsg:     "hello world",
			inputPrivKey: "ED6BF4E585BA0C4055F6E63D0D6D06E7D8B9F00AA02337BCF864385275892A1EB5",
			expected:     "84F05438BFFC29F49E8DC8865251DA1CEF9A5A9CAA7DC2629985986C35271CC1AC389846F955C548A322F433F387CE928329F091E8FA7E2A8E7DFDAB8E88310B",
			expectedErr:  nil,
		},
		{
			name:         "pass - sign a longer message with a different private key",
			inputMsg:     "hello madsjdadjdas,adajofahffa !$@~",
			inputPrivKey: "ED28E9ADC9383EB494476DCF7D95DD4B16F6A2C325365F9E17007294F4AE487CE0",
			expected:     "77F07A34D408DD8C3C6BCED0E31C2909D8E13ECB15AF15345CA1ECE53118519754971BE1DD7A0A52E5D737D4DBFAD01018727EF1F0BAD06B31CD8D6F3D9E7E05",
			expectedErr:  nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := ED25519().Sign(tc.inputMsg, tc.inputPrivKey)

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

func TestED25519Validate(t *testing.T) {
	tt := []struct {
		name        string
		inputMsg    string
		inputPubKey string
		inputSig    string
		expected    bool
	}{
		{
			name:        "fail - invalid signature hex",
			inputMsg:    "test message",
			inputPubKey: "ED4924A9045FE5ED8B22BAA7B6229A72A287CCF3EA287AADD3A032A24C0F008FA6",
			inputSig:    "invalid_sig",
			expected:    false,
		},
		{
			name:        "fail - invalid signature for message",
			inputMsg:    "test message",
			inputPubKey: "ED4924A9045FE5ED8B22BAA7B6229A72A287CCF3EA287AADD3A032A24C0F008FB6",
			inputSig:    "C001CB8A9883497518917DD16391930F4FEE39CEA76C846CFF4330BA44ED19DC4730056C2C6D7452873DE8120A5023C6807135C6329A89A13BA1D476FE8E7100",
			expected:    false,
		},
		{
			name:        "fail - invalid public key hex",
			inputMsg:    "test message",
			inputPubKey: "invalid_key",
			inputSig:    "C001CB8A9883497518917DD16391930F4FEE39CEA76C846CFF4330BA44ED19DC4730056C2C6D7452873DE8120A5023C6807135C6329A89A13BA1D476FE8E7100",
			expected:    false,
		},
		{
			name:        "pass - valid signature for message",
			inputMsg:    "test message",
			inputPubKey: "ED4924A9045FE5ED8B22BAA7B6229A72A287CCF3EA287AADD3A032A24C0F008FA6",
			inputSig:    "C001CB8A9883497518917DD16391930F4FEE39CEA76C846CFF4330BA44ED19DC4730056C2C6D7452873DE8120A5023C6807135C6329A89A13BA1D476FE8E7100",
			expected:    true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual := ED25519().Validate(tc.inputMsg, tc.inputPubKey, tc.inputSig)
			require.Equal(t, tc.expected, actual)
		})
	}
}
