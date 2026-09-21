// Package flatten holds helpers shared by the Flatten methods of XRPL models.
package flatten

import addresscodec "github.com/Peersyst/xrpl-go/address-codec"

// ClassicIfTagless returns the classic form of a valid X-address that carries no tag.
//
// Every other input is returned unchanged. A classic address is already in its final
// form. A tagged X-address, including one with an explicit zero tag, keeps its tag so
// that later validation or encoding can reject it instead of silently dropping the tag.
// A malformed address is left for the same later stages to report.
func ClassicIfTagless(address string) string {
	decoded, err := addresscodec.DecodeAddress(address)
	if err != nil || decoded.HasTag {
		return address
	}
	return decoded.Classic
}
