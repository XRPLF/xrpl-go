package addresscodec

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/stretchr/testify/require"
)

// Fixtures represents the structure of address-fixtures.json
type Fixtures struct {
	EncodeDecodeAccountID     []EncodeDecodeTest `json:"encodeDecodeAccountID"`
	EncodeDecodeNodePublic    []EncodeDecodeTest `json:"encodeDecodeNodePublic"`
	EncodeDecodeAccountPublic []EncodeDecodeTest `json:"encodeDecodeAccountPublic"`
	Seeds                     []SeedTest         `json:"seeds"`
	ValidClassicAddresses     []string           `json:"validClassicAddresses"`
	InvalidClassicAddresses   []string           `json:"invalidClassicAddresses"`
	XAddresses                []XAddressTest     `json:"xAddresses"`
	InvalidXAddresses         []InvalidXAddress  `json:"invalidXAddresses"`
	CodecTests                []CodecTest        `json:"codecTests"`
}

type EncodeDecodeTest struct {
	Hex    string `json:"hex"`
	Base58 string `json:"base58"`
}

type SeedTest struct {
	Hex    string `json:"hex"`
	Base58 string `json:"base58"`
	Type   string `json:"type"`
}

type XAddressTest struct {
	ClassicAddress string `json:"classicAddress"`
	Tag            *int64 `json:"tag"` // pointer to handle null
	MainnetAddress string `json:"mainnetAddress"`
	TestnetAddress string `json:"testnetAddress"`
}

type InvalidXAddress struct {
	Address string `json:"address"`
	Error   string `json:"error"`
}

type CodecTest struct {
	Input          string `json:"input"`
	Version        int    `json:"version"`
	ExpectedLength int    `json:"expectedLength"`
	Encoded        string `json:"encoded"`
}

func loadFixtures(t *testing.T) *Fixtures {
	data, err := os.ReadFile("testdata/fixtures/address-fixtures.json")
	require.NoError(t, err, "Failed to read fixtures file")

	var fixtures Fixtures
	err = json.Unmarshal(data, &fixtures)
	require.NoError(t, err, "Failed to parse fixtures JSON")

	return &fixtures
}

// TestCompat_EncodeDecodeAccountID tests encoding and decoding of AccountIDs
// Reference: xrpl.js/packages/ripple-address-codec/test/xrp-codec.test.ts
func TestCompat_EncodeDecodeAccountID(t *testing.T) {
	fixtures := loadFixtures(t)

	for _, tc := range fixtures.EncodeDecodeAccountID {
		t.Run(tc.Base58, func(t *testing.T) {
			// Test encoding
			hexBytes, err := hex.DecodeString(tc.Hex)
			require.NoError(t, err)

			encoded, err := EncodeAccountIDToClassicAddress(hexBytes)
			require.NoError(t, err)
			require.Equal(t, tc.Base58, encoded, "Encoding mismatch for hex: %s", tc.Hex)

			// Test decoding
			_, decoded, err := DecodeClassicAddressToAccountID(tc.Base58)
			require.NoError(t, err)
			require.Equal(t, strings.ToUpper(tc.Hex), strings.ToUpper(hex.EncodeToString(decoded)), "Decoding mismatch for base58: %s", tc.Base58)
		})
	}
}

// TestCompat_EncodeDecodeNodePublic tests encoding and decoding of NodePublic keys
// Reference: xrpl.js/packages/ripple-address-codec/test/xrp-codec.test.ts
func TestCompat_EncodeDecodeNodePublic(t *testing.T) {
	fixtures := loadFixtures(t)

	for _, tc := range fixtures.EncodeDecodeNodePublic {
		t.Run(tc.Base58, func(t *testing.T) {
			// Test encoding
			hexBytes, err := hex.DecodeString(tc.Hex)
			require.NoError(t, err)

			encoded, err := EncodeNodePublicKey(hexBytes)
			require.NoError(t, err)
			require.Equal(t, tc.Base58, encoded, "Encoding mismatch for hex: %s", tc.Hex)

			// Test decoding
			decoded, err := DecodeNodePublicKey(tc.Base58)
			require.NoError(t, err)
			require.Equal(t, strings.ToUpper(tc.Hex), strings.ToUpper(hex.EncodeToString(decoded)), "Decoding mismatch for base58: %s", tc.Base58)
		})
	}
}

// TestCompat_EncodeDecodeAccountPublic tests encoding and decoding of AccountPublic keys
// Reference: xrpl.js/packages/ripple-address-codec/test/xrp-codec.test.ts
func TestCompat_EncodeDecodeAccountPublic(t *testing.T) {
	fixtures := loadFixtures(t)

	for _, tc := range fixtures.EncodeDecodeAccountPublic {
		t.Run(tc.Base58, func(t *testing.T) {
			// Test encoding
			hexBytes, err := hex.DecodeString(tc.Hex)
			require.NoError(t, err)

			encoded, err := EncodeAccountPublicKey(hexBytes)
			require.NoError(t, err)
			require.Equal(t, tc.Base58, encoded, "Encoding mismatch for hex: %s", tc.Hex)

			// Test decoding
			decoded, err := DecodeAccountPublicKey(tc.Base58)
			require.NoError(t, err)
			require.Equal(t, strings.ToUpper(tc.Hex), strings.ToUpper(hex.EncodeToString(decoded)), "Decoding mismatch for base58: %s", tc.Base58)
		})
	}
}

// TestCompat_EncodeSeed tests seed encoding
// Reference: xrpl.js/packages/ripple-address-codec/test/xrp-codec.test.ts
func TestCompat_EncodeSeed(t *testing.T) {
	fixtures := loadFixtures(t)

	for _, tc := range fixtures.Seeds {
		t.Run(tc.Base58, func(t *testing.T) {
			hexBytes, err := hex.DecodeString(tc.Hex)
			require.NoError(t, err)

			var encoded string
			if tc.Type == "ed25519" {
				encoded, err = EncodeSeed(hexBytes, crypto.ED25519())
			} else {
				encoded, err = EncodeSeed(hexBytes, crypto.SECP256K1())
			}
			require.NoError(t, err)
			require.Equal(t, tc.Base58, encoded, "Seed encoding mismatch for hex: %s, type: %s", tc.Hex, tc.Type)
		})
	}
}

// TestCompat_DecodeSeed tests seed decoding
// Reference: xrpl.js/packages/ripple-address-codec/test/xrp-codec.test.ts
func TestCompat_DecodeSeed(t *testing.T) {
	fixtures := loadFixtures(t)

	for _, tc := range fixtures.Seeds {
		t.Run(tc.Base58, func(t *testing.T) {
			decoded, cryptoType, err := DecodeSeed(tc.Base58)
			require.NoError(t, err)
			require.Equal(t, strings.ToUpper(tc.Hex), strings.ToUpper(hex.EncodeToString(decoded)), "Seed decoding mismatch for base58: %s", tc.Base58)

			// Check type by comparing with known implementations
			if tc.Type == "ed25519" {
				_, ok := cryptoType.(crypto.ED25519CryptoAlgorithm)
				require.True(t, ok, "Expected ed25519 type for base58: %s", tc.Base58)
			} else {
				_, ok := cryptoType.(crypto.SECP256K1CryptoAlgorithm)
				require.True(t, ok, "Expected secp256k1 type for base58: %s", tc.Base58)
			}
		})
	}
}

// TestCompat_IsValidClassicAddress tests classic address validation
// Reference: xrpl.js/packages/ripple-address-codec/test/xrp-codec.test.ts
func TestCompat_IsValidClassicAddress(t *testing.T) {
	fixtures := loadFixtures(t)

	for _, addr := range fixtures.ValidClassicAddresses {
		t.Run("valid_"+addr, func(t *testing.T) {
			require.True(t, IsValidClassicAddress(addr), "Expected %s to be valid", addr)
		})
	}

	for _, addr := range fixtures.InvalidClassicAddresses {
		name := "invalid_" + addr
		if addr == "" {
			name = "invalid_empty"
		}
		t.Run(name, func(t *testing.T) {
			require.False(t, IsValidClassicAddress(addr), "Expected %s to be invalid", addr)
		})
	}
}

// TestCompat_XAddressMainnet tests X-address encoding/decoding for mainnet
// Reference: xrpl.js/packages/ripple-address-codec/test/index.test.ts
func TestCompat_XAddressMainnet(t *testing.T) {
	fixtures := loadFixtures(t)

	for _, tc := range fixtures.XAddresses {
		testName := tc.ClassicAddress
		if tc.Tag != nil {
			testName += "_tag_" + string(rune(*tc.Tag))
		} else {
			testName += "_no_tag"
		}
		t.Run("mainnet_"+testName, func(t *testing.T) {
			var tag uint32
			tagFlag := false
			if tc.Tag != nil {
				tag = uint32(*tc.Tag)
				tagFlag = true
			}

			// Test classic -> X-address conversion
			xAddr, err := ClassicAddressToXAddress(tc.ClassicAddress, tag, tagFlag, false)
			require.NoError(t, err)
			require.Equal(t, tc.MainnetAddress, xAddr, "Classic to X-address conversion failed for %s", tc.ClassicAddress)

			// Test X-address -> classic conversion
			classicAddr, decodedTag, isTestnet, err := XAddressToClassicAddress(tc.MainnetAddress)
			require.NoError(t, err)
			require.Equal(t, tc.ClassicAddress, classicAddr, "X-address to classic conversion failed for %s", tc.MainnetAddress)
			require.False(t, isTestnet, "Expected mainnet address")

			if tc.Tag != nil {
				require.Equal(t, uint32(*tc.Tag), decodedTag, "Tag mismatch for %s", tc.MainnetAddress)
			}

			// Test IsValidXAddress
			require.True(t, IsValidXAddress(tc.MainnetAddress), "Expected %s to be a valid X-address", tc.MainnetAddress)
		})
	}
}

// TestCompat_XAddressTestnet tests X-address encoding/decoding for testnet
// Reference: xrpl.js/packages/ripple-address-codec/test/index.test.ts
func TestCompat_XAddressTestnet(t *testing.T) {
	fixtures := loadFixtures(t)

	for _, tc := range fixtures.XAddresses {
		testName := tc.ClassicAddress
		if tc.Tag != nil {
			testName += "_tag"
		} else {
			testName += "_no_tag"
		}
		t.Run("testnet_"+testName, func(t *testing.T) {
			var tag uint32
			tagFlag := false
			if tc.Tag != nil {
				tag = uint32(*tc.Tag)
				tagFlag = true
			}

			// Test classic -> X-address conversion
			xAddr, err := ClassicAddressToXAddress(tc.ClassicAddress, tag, tagFlag, true)
			require.NoError(t, err)
			require.Equal(t, tc.TestnetAddress, xAddr, "Classic to X-address conversion failed for %s", tc.ClassicAddress)

			// Test X-address -> classic conversion
			classicAddr, decodedTag, isTestnet, err := XAddressToClassicAddress(tc.TestnetAddress)
			require.NoError(t, err)
			require.Equal(t, tc.ClassicAddress, classicAddr, "X-address to classic conversion failed for %s", tc.TestnetAddress)
			require.True(t, isTestnet, "Expected testnet address")

			if tc.Tag != nil {
				require.Equal(t, uint32(*tc.Tag), decodedTag, "Tag mismatch for %s", tc.TestnetAddress)
			}

			// Test IsValidXAddress
			require.True(t, IsValidXAddress(tc.TestnetAddress), "Expected %s to be a valid X-address", tc.TestnetAddress)
		})
	}
}

// TestCompat_InvalidXAddresses tests that invalid X-addresses are properly rejected
// Reference: xrpl.js/packages/ripple-address-codec/test/index.test.ts
func TestCompat_InvalidXAddresses(t *testing.T) {
	fixtures := loadFixtures(t)

	for _, tc := range fixtures.InvalidXAddresses {
		t.Run(tc.Address[:20], func(t *testing.T) {
			require.False(t, IsValidXAddress(tc.Address), "Expected %s to be invalid", tc.Address)

			_, _, _, err := XAddressToClassicAddress(tc.Address)
			require.Error(t, err, "Expected error for invalid X-address: %s", tc.Address)
		})
	}
}
