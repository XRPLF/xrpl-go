package transaction

import (
	"strings"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/stretchr/testify/require"
)

const (
	testCompressedPoint1 = "0279BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798"
	testCompressedPoint2 = "02C6047F9441ED7D6D3045406E95C07CD85C778E4B8CEF3CA7ABAC09B95C709EE5"
	testCompressedPoint3 = "02F9308A019258C31049344F85F89D5229B531C845836F99B08601F113BCE036F9"
	testCiphertext       = testCompressedPoint1 + testCompressedPoint2
	testCiphertext2      = testCompressedPoint2 + testCompressedPoint3
	testCiphertext3      = testCompressedPoint3 + testCompressedPoint1
)

var (
	testSchnorrProof   = strings.Repeat("CD", 64)
	testBlindingFactor = strings.Repeat("EF", 32)
)

// X-address encodings of the confidential MPT test accounts. An X-address and a classic
// address can name the same account, so the validation suites use these to prove the
// self-reference checks compare AccountIDs rather than the encoded address strings.
const (
	// testXAddressAccount is the X-address form of the account used by newValidConfidentialMPTSend.
	testXAddressAccount = "XVYaPuwjbmRPA9pdyiXAGXsw8NhgJqESZxvSGuTLKhngUD4"
	// testXAddressDestination is the X-address form of that fixture's Destination, untagged.
	testXAddressDestination = "XVwik2uCUAoRY8hPnsfgWzcd2CAgSgELgpfRiw1rGg2FwzN"
	// testXAddressTaggedDestination is the same destination carrying an embedded tag of 7.
	testXAddressTaggedDestination = "XVwik2uCUAoRY8hPnsfgWzcd2CAgSgLJXgaNpGnFmfQYmtR"
	// testXAddressTaggedAccount is that fixture's Account carrying an embedded tag of 7.
	testXAddressTaggedAccount = "XVYaPuwjbmRPA9pdyiXAGXsw8NhgJqLQRcuNSzg9oRVsamV"
)

// TestConfidentialMPTXAddressFixtures pins the X-address fixtures to the classic
// addresses they are meant to encode, so a change to either surfaces here rather than
// as an unexplained pass in the self-reference cases.
func TestConfidentialMPTXAddressFixtures(t *testing.T) {
	tests := []struct {
		name    string
		xAddr   string
		classic string
		tag     uint32
		hasTag  bool
	}{
		{name: "account", xAddr: testXAddressAccount, classic: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"},
		{name: "destination", xAddr: testXAddressDestination, classic: "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP"},
		{name: "tagged destination", xAddr: testXAddressTaggedDestination, classic: "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP", tag: 7, hasTag: true},
		{name: "tagged account", xAddr: testXAddressTaggedAccount, classic: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD", tag: 7, hasTag: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, wantID, err := addresscodec.DecodeClassicAddressToAccountID(test.classic)
			require.NoError(t, err)

			gotID, tag, hasTag, _, err := addresscodec.DecodeXAddress(test.xAddr)
			require.NoError(t, err)
			require.Equal(t, wantID, gotID)
			require.Equal(t, test.hasTag, hasTag)
			require.Equal(t, test.tag, tag)
		})
	}
}
