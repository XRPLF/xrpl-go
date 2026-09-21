// Package types provides core transaction types and helpers for the XRPL Go library.
//
//revive:disable:var-naming
package types

import addresscodec "github.com/Peersyst/xrpl-go/address-codec"

// Address represents a classic XRPL account address as a string.
type Address string

func (a Address) String() string {
	return string(a)
}

// Flatten returns the form of the address to place inside a nested object, such as an
// issuer. The codec can route an X-address tag only for top-level account fields, so a
// tagless X-address becomes its classic form here.
//
// Every other input is returned unchanged. A classic address is already in its final
// form. A tagged X-address, including one with an explicit zero tag, keeps its tag so
// that later validation or encoding can reject it instead of silently dropping the tag.
// A malformed address is left for the same later stages to report.
func (a Address) Flatten() string {
	decoded, err := addresscodec.DecodeAddress(a.String())
	if err != nil || decoded.HasTag {
		return a.String()
	}
	return decoded.Classic
}
