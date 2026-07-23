package addresscodec

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsValidXAddress(t *testing.T) {
	invalidReservedTagBytes := xAddressWithTagReservedBytes()

	testcases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "pass - valid x-address",
			input:    "X7AcgcsBL6XDcUb289X4mJ8djcdyKaB5hJDWMArnXr61cqZ",
			expected: true,
		},
		{
			name:     "pass - invalid x-address",
			input:    "invalid",
			expected: false,
		},
		{
			name:     "fail - invalid x-address with non-zero reserved tag bytes",
			input:    invalidReservedTagBytes,
			expected: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, IsValidXAddress(tc.input))
		})
	}
}

func TestEncodeXAddress(t *testing.T) {
	testcases := []struct {
		name        string
		input       []byte
		tag         uint32
		tagFlag     bool
		testnetFlag bool
		expected    string
		expectedErr error
	}{
		{
			name:        "fail - invalid accountId length",
			input:       []byte{1, 2, 3},
			tag:         0,
			tagFlag:     false,
			testnetFlag: false,
			expectedErr: ErrInvalidAccountID,
		},
		{
			name: "pass - valid testnet x-address",
			input: []byte{
				94, 123, 17, 37, 35, 246,
				141, 47, 94, 135, 157, 180,
				234, 197, 28, 102, 152, 166,
				147, 4,
			},
			tag:         0,
			tagFlag:     false,
			testnetFlag: true,
			expected:    "T719a5UwUCnEs54UsxG9CJYYDhwmFCqkr7wxCcNcfZ6p5GZ",
			expectedErr: nil,
		},
		{
			name: "pass - valid testnet x-address with tag",
			input: []byte{
				94, 123, 17, 37, 35, 246,
				141, 47, 94, 135, 157, 180,
				234, 197, 28, 102, 152, 166,
				147, 4,
			},
			tag:         22,
			tagFlag:     true,
			testnetFlag: true,
			expected:    "T719a5UwUCnEs54UsxG9CJYYDhwmFCvzHM39KcuJw6gp2gS",
			expectedErr: nil,
		},
		{
			name: "pass - valid mainnet x-address",
			input: []byte{
				94, 123, 17, 37, 35, 246,
				141, 47, 94, 135, 157, 180,
				234, 197, 28, 102, 152, 166,
				147, 4,
			},
			tag:         0,
			tagFlag:     false,
			testnetFlag: false,
			expected:    "X7AcgcsBL6XDcUb289X4mJ8djcdyKaB5hJDWMArnXr61cqZ",
			expectedErr: nil,
		},
		{
			name: "pass - valid mainnet x-address with tag",
			input: []byte{
				94, 123, 17, 37, 35, 246,
				141, 47, 94, 135, 157, 180,
				234, 197, 28, 102, 152, 166,
				147, 4,
			},
			tag:         22,
			tagFlag:     true,
			testnetFlag: false,
			expected:    "X7AcgcsBL6XDcUb289X4mJ8djcdyKaGxLBw6rACm2heBxVn",
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := EncodeXAddress(tc.input, tc.tag, tc.tagFlag, tc.testnetFlag)
			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				return
			} else {
				require.Equal(t, tc.expected, actual)
				require.NoError(t, err)
			}
		})
	}
}

func TestDecodeXAddress(t *testing.T) {
	invalidReservedTagBytes := xAddressWithTagReservedBytes()
	explicitZeroTag, err := EncodeXAddress([]byte{
		94, 123, 17, 37, 35, 246,
		141, 47, 94, 135, 157, 180,
		234, 197, 28, 102, 152, 166,
		147, 4,
	}, 0, true, false)
	require.NoError(t, err)

	testcases := []struct {
		name              string
		input             string
		expectedAccountId []byte
		expectedTag       uint32
		expectedHasTag    bool
		expectedTestnet   bool
		expectedErr       error
	}{
		{
			name:        "fail - invalid x-address",
			input:       "invalid",
			expectedErr: ErrInvalidFormat,
		},
		{
			name:  "pass - valid testnet x-address",
			input: "T719a5UwUCnEs54UsxG9CJYYDhwmFCqkr7wxCcNcfZ6p5GZ",
			expectedAccountId: []byte{
				94, 123, 17, 37, 35, 246,
				141, 47, 94, 135, 157, 180,
				234, 197, 28, 102, 152, 166,
				147, 4,
			},
			expectedTag:     0,
			expectedHasTag:  false,
			expectedTestnet: true,
			expectedErr:     nil,
		},
		{
			name:  "pass - valid testnet x-address with tag",
			input: "T719a5UwUCnEs54UsxG9CJYYDhwmFCvzHM39KcuJw6gp2gS",
			expectedAccountId: []byte{
				94, 123, 17, 37, 35, 246,
				141, 47, 94, 135, 157, 180,
				234, 197, 28, 102, 152, 166,
				147, 4,
			},
			expectedTag:     22,
			expectedHasTag:  true,
			expectedTestnet: true,
			expectedErr:     nil,
		},
		{
			name:  "pass - valid mainnet x-address",
			input: "X7AcgcsBL6XDcUb289X4mJ8djcdyKaB5hJDWMArnXr61cqZ",
			expectedAccountId: []byte{
				94, 123, 17, 37, 35, 246,
				141, 47, 94, 135, 157, 180,
				234, 197, 28, 102, 152, 166,
				147, 4,
			},
			expectedTag:     0,
			expectedHasTag:  false,
			expectedTestnet: false,
			expectedErr:     nil,
		},
		{
			name:  "pass - valid mainnet x-address with tag",
			input: "X7AcgcsBL6XDcUb289X4mJ8djcdyKaGxLBw6rACm2heBxVn",
			expectedAccountId: []byte{
				94, 123, 17, 37, 35, 246,
				141, 47, 94, 135, 157, 180,
				234, 197, 28, 102, 152, 166,
				147, 4,
			},
			expectedTag:     22,
			expectedHasTag:  true,
			expectedTestnet: false,
			expectedErr:     nil,
		},
		{
			name:  "pass - valid mainnet x-address with explicit zero tag",
			input: explicitZeroTag,
			expectedAccountId: []byte{
				94, 123, 17, 37, 35, 246,
				141, 47, 94, 135, 157, 180,
				234, 197, 28, 102, 152, 166,
				147, 4,
			},
			expectedTag:     0,
			expectedHasTag:  true,
			expectedTestnet: false,
			expectedErr:     nil,
		},
		{
			name:        "fail - x-address with non-zero reserved tag bytes",
			input:       invalidReservedTagBytes,
			expectedErr: ErrInvalidTag,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualAccountId, actualTag, actualHasTag, actualTestnet, err := DecodeXAddress(tc.input)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedAccountId, actualAccountId)
				require.Equal(t, tc.expectedTag, actualTag)
				require.Equal(t, tc.expectedHasTag, actualHasTag)
				require.Equal(t, tc.expectedTestnet, actualTestnet)
			}
		})
	}
}

func TestXAddressToClassicAddress(t *testing.T) {
	testcases := []struct {
		name                   string
		input                  string
		expectedClassicAddress string
		expectedTag            uint32
		expectedHasTag         bool
		expectedTestnet        bool
		expectedErr            error
	}{
		{
			name:        "fail - invalid x-address",
			input:       "invalid",
			expectedErr: ErrInvalidFormat,
		},
		{
			name:                   "pass - valid testnet x-address",
			input:                  "T719a5UwUCnEs54UsxG9CJYYDhwmFCqkr7wxCcNcfZ6p5GZ",
			expectedClassicAddress: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
			expectedTag:            0,
			expectedHasTag:         false,
			expectedTestnet:        true,
			expectedErr:            nil,
		},
		{
			name:                   "pass - valid mainnet x-address",
			input:                  "X7AcgcsBL6XDcUb289X4mJ8djcdyKaB5hJDWMArnXr61cqZ",
			expectedClassicAddress: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
			expectedTag:            0,
			expectedHasTag:         false,
			expectedTestnet:        false,
			expectedErr:            nil,
		},
		{
			name:                   "pass - valid mainnet x-address with explicit zero tag",
			input:                  "XV5sbjUmgPpvXv4ixFWZ5ptAYZ6PD2m4Er6SnvjVLpMWPjR",
			expectedClassicAddress: "rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY",
			expectedTag:            0,
			expectedHasTag:         true,
			expectedTestnet:        false,
			expectedErr:            nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualClassicAddress, actualTag, actualHasTag, actualTestnet, err := XAddressToClassicAddress(tc.input)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedClassicAddress, actualClassicAddress)
				require.Equal(t, tc.expectedTag, actualTag)
				require.Equal(t, tc.expectedHasTag, actualHasTag)
				require.Equal(t, tc.expectedTestnet, actualTestnet)
			}
		})
	}
}

func TestClassicAddressToXAddress(t *testing.T) {
	testcases := []struct {
		name        string
		input       string
		tag         uint32
		tagFlag     bool
		testnetFlag bool
		expected    string
		expectedErr error
	}{
		{
			name:        "fail - invalid classic address",
			input:       "invalid",
			expectedErr: ErrInvalidClassicAddress,
		},
		{
			name:        "pass - valid testnet classic address",
			input:       "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
			tag:         0,
			tagFlag:     false,
			testnetFlag: true,
			expected:    "T719a5UwUCnEs54UsxG9CJYYDhwmFCqkr7wxCcNcfZ6p5GZ",
			expectedErr: nil,
		},
		{
			name:        "pass - valid mainnet classic address",
			input:       "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
			tag:         0,
			tagFlag:     false,
			testnetFlag: false,
			expected:    "X7AcgcsBL6XDcUb289X4mJ8djcdyKaB5hJDWMArnXr61cqZ",
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := ClassicAddressToXAddress(tc.input, tc.tag, tc.tagFlag, tc.testnetFlag)
			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestDecodeTag(t *testing.T) {
	testcases := []struct {
		name        string
		input       []byte
		expectedTag uint32
		hasTag      bool
		expectedErr error
	}{
		{
			name:        "fail - unsupported 64-bit tag (flag >= 2)",
			input:       []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedErr: ErrUnsupportedXAddress,
		},
		{
			name:        "pass - valid tag - 1",
			input:       []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedTag: 1,
			hasTag:      true,
			expectedErr: nil,
		},
		{
			name:        "pass - valid tag - 0",
			input:       []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedTag: 0,
			hasTag:      true,
			expectedErr: nil,
		},
		{
			name:        "fail - non-zero reserved bytes with 32-bit tag",
			input:       []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0},
			expectedErr: ErrInvalidTag,
		},
		{
			name:        "pass - no tag (flag = 0)",
			input:       []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedTag: 0,
			hasTag:      false,
			expectedErr: nil,
		},
		{
			name:        "pass - large tag (32-bit max)",
			input:       []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xFF, 0xFF, 0xFF, 0xFF, 0, 0, 0, 0, 0},
			expectedTag: 4294967295,
			hasTag:      true,
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual, hasTag, err := decodeTag(tc.input)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedTag, actual)
				require.Equal(t, tc.hasTag, hasTag)
			}
		})
	}
}

func xAddressWithTagReservedBytes() string {
	accountID := []byte{
		94, 123, 17, 37, 35, 246,
		141, 47, 94, 135, 157, 180,
		234, 197, 28, 102, 152, 166,
		147, 4,
	}

	payload := make([]byte, 0, AccountAddressLength+9)
	payload = append(payload, accountID...)
	payload = append(payload, 1, 22, 0, 0, 0, 1, 0, 0, 0)

	return Base58CheckEncode(payload, MainnetXAddressPrefix...)
}
