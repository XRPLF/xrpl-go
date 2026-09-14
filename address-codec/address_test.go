package addresscodec

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testClassicAddress = "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"
	testAccountIDHex   = "d28b177e48d9a8d057e70f7e464b498367281b98"
)

// xAddressOf encodes testClassicAddress in the X-address form a case exercises, so the
// table names the form instead of carrying an opaque literal.
func xAddressOf(t *testing.T, tag uint32, hasTag, testnet bool) string {
	t.Helper()
	encoded, err := ClassicAddressToXAddress(testClassicAddress, tag, hasTag, testnet)
	require.NoError(t, err)
	return encoded
}

func TestDecodeAddress(t *testing.T) {
	// Anchored to a known value so a truncated copy cannot corrupt every case alike
	// and still compare equal.
	wantAccountID, err := hex.DecodeString(testAccountIDHex)
	require.NoError(t, err)

	testcases := []struct {
		name        string
		address     string
		wantHasTag  bool
		wantTestnet bool
	}{
		{
			name:    "pass - classic address",
			address: testClassicAddress,
		},
		{
			name:    "pass - untagged x-address",
			address: xAddressOf(t, 0, false, false),
		},
		{
			name:       "pass - tagged x-address",
			address:    xAddressOf(t, 42, true, false),
			wantHasTag: true,
		},
		{
			// A zero tag is still a tag, so it must stay distinguishable from a tagless
			// x-address.
			name:       "pass - x-address with a zero tag",
			address:    xAddressOf(t, 0, true, false),
			wantHasTag: true,
		},
		{
			name:        "pass - untagged testnet x-address",
			address:     xAddressOf(t, 0, false, true),
			wantTestnet: true,
		},
		{
			name:        "pass - tagged testnet x-address",
			address:     xAddressOf(t, 42, true, true),
			wantHasTag:  true,
			wantTestnet: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			decoded, err := DecodeAddress(tc.address)
			require.NoError(t, err)
			require.Equal(t, wantAccountID, decoded.AccountID[:])
			// Every form resolves to the same account, so Classic is the same spelling
			// regardless of the tag or the network the x-address was encoded for.
			require.Equal(t, testClassicAddress, decoded.Classic)
			require.Equal(t, tc.wantHasTag, decoded.HasTag)
			// A classic address names no network, so it is never reported as testnet.
			require.Equal(t, tc.wantTestnet, decoded.Testnet)
		})
	}
}

// rawXAddress base58check-encodes an x-address payload directly, so a case can
// exercise a prefix or flag byte the encoder would never produce.
func rawXAddress(t *testing.T, prefix []byte, flag byte) string {
	t.Helper()
	accountID, err := hex.DecodeString(testAccountIDHex)
	require.NoError(t, err)

	payload := make([]byte, 0, 31)
	payload = append(payload, prefix...)
	payload = append(payload, accountID...)
	payload = append(payload, flag)
	payload = append(payload, make([]byte, 8)...)
	return Base58CheckEncode(payload)
}

func TestDecodeAddress_Invalid(t *testing.T) {
	testcases := []struct {
		name     string
		address  string
		wantErrs []error
	}{
		{
			name:     "fail - empty",
			address:  "",
			wantErrs: []error{ErrInvalidClassicAddress, ErrInvalidFormat},
		},
		{
			name:     "fail - not base58",
			address:  "invalid",
			wantErrs: []error{ErrInvalidClassicAddress, ErrInvalidFormat},
		},
		{
			// Decodes as base58 and is the right length, so it reaches the
			// x-address prefix check rather than failing earlier.
			name:     "fail - x-address with an unknown prefix",
			address:  rawXAddress(t, []byte{0x00, 0x00}, 0),
			wantErrs: []error{ErrInvalidClassicAddress, ErrInvalidXAddress},
		},
		{
			name:     "fail - x-address with a 64-bit tag",
			address:  rawXAddress(t, MainnetXAddressPrefix, 2),
			wantErrs: []error{ErrInvalidClassicAddress, ErrUnsupportedXAddress},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := DecodeAddress(tc.address)
			// The wrap keeps the general case and both per-form failures matchable.
			require.ErrorIs(t, err, ErrInvalidAddressFormat)
			for _, wantErr := range tc.wantErrs {
				require.ErrorIs(t, err, wantErr)
			}
		})
	}
}

func TestIsZeroAccountID(t *testing.T) {
	nonZero := make([]byte, AccountAddressLength)
	nonZero[len(nonZero)-1] = 1

	testcases := []struct {
		name      string
		accountID []byte
		want      bool
	}{
		{
			name:      "pass - all-zero account ID",
			accountID: make([]byte, AccountAddressLength),
			want:      true,
		},
		{
			name:      "fail - non-zero account ID",
			accountID: nonZero,
		},
		// A slice that is not an account ID at all is not ACCOUNT_ZERO, so the length
		// cases must not fall through the loop and report true.
		{
			name:      "fail - nil",
			accountID: nil,
		},
		{
			name:      "fail - empty",
			accountID: []byte{},
		},
		{
			name:      "fail - too short",
			accountID: make([]byte, AccountAddressLength-1),
		},
		{
			name:      "fail - too long",
			accountID: make([]byte, AccountAddressLength+1),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, IsZeroAccountID(tc.accountID))
		})
	}
}
