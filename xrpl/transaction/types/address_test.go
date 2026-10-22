package types

import (
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/stretchr/testify/require"
)

func TestAddress_Flatten(t *testing.T) {
	const classic = "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"

	xAddress := func(tag uint32, hasTag, testnet bool) string {
		t.Helper()
		encoded, err := addresscodec.ClassicAddressToXAddress(classic, tag, hasTag, testnet)
		require.NoError(t, err)
		return encoded
	}

	taggedMainnet := xAddress(7, true, false)
	zeroTagMainnet := xAddress(0, true, false)
	taggedTestnet := xAddress(7, true, true)

	tests := []struct {
		name    string
		address string
		want    string
	}{
		{name: "pass - classic address is unchanged", address: classic, want: classic},
		{name: "pass - tagless mainnet X-address becomes classic", address: xAddress(0, false, false), want: classic},
		{name: "pass - tagless testnet X-address becomes classic", address: xAddress(0, false, true), want: classic},
		{name: "pass - tagged mainnet X-address is unchanged", address: taggedMainnet, want: taggedMainnet},
		{name: "pass - explicit zero tag X-address is unchanged", address: zeroTagMainnet, want: zeroTagMainnet},
		{name: "pass - tagged testnet X-address is unchanged", address: taggedTestnet, want: taggedTestnet},
		{name: "pass - malformed address is unchanged", address: "rEXAMPLE", want: "rEXAMPLE"},
		{name: "pass - empty address is unchanged", address: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, Address(tt.address).Flatten())
		})
	}
}
