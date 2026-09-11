package addresscodec

import "fmt"

// DecodedAddress is the canonical identity of a classic address or X-address.
type DecodedAddress struct {
	// AccountID is the 20-byte account identifier both address forms encode.
	AccountID [AccountAddressLength]byte
	// Classic is the classic r-address spelling of AccountID.
	Classic string
	// HasTag reports whether the address was an X-address carrying a destination tag.
	HasTag bool
	// Testnet reports whether the address was an X-address encoded for a test network.
	// Both networks encode the same AccountID, so this never changes Classic, and XRPL
	// does not bind an account to a network. It is reported so a caller that knows which
	// network it targets can act on it. Nothing in this module does.
	Testnet bool
}

// DecodeAddress resolves a classic address or an X-address to the account it names.
// It is the decoding counterpart to IsValidAddress, which accepts either form.
//
// Callers that compare accounts should use AccountID rather than the address string,
// because a classic address and its X-address form name the same account while
// spelling it differently. Callers passing an address on to code that understands
// only classic addresses should use Classic.
//
// The returned error wraps ErrInvalidAddressFormat along with the failure from each
// form, so a caller can match the general case or either specific one.
func DecodeAddress(address string) (DecodedAddress, error) {
	var decoded DecodedAddress

	_, accountID, classicErr := DecodeClassicAddressToAccountID(address)
	if classicErr == nil {
		copy(decoded.AccountID[:], accountID)
		decoded.Classic = address
		return decoded, nil
	}

	accountID, _, hasTag, testnet, xAddressErr := DecodeXAddress(address)
	if xAddressErr != nil {
		return DecodedAddress{}, fmt.Errorf("%w %q: classic address: %w, X-address: %w", ErrInvalidAddressFormat, address, classicErr, xAddressErr)
	}

	copy(decoded.AccountID[:], accountID)
	decoded.Classic = Base58CheckEncode(decoded.AccountID[:], AccountAddressPrefix)
	decoded.HasTag = hasTag
	decoded.Testnet = testnet
	return decoded, nil
}

// IsZeroAccountID reports whether accountID is the all-zero ACCOUNT_ZERO identifier.
//
// It expects a decoded AccountID, which is what DecodeAddress and the classic and
// X-address decoders all return, so the input is always AccountAddressLength bytes in
// practice. The length check is a guard for a caller that reaches here with something
// else, not a validation step. Such input is not ACCOUNT_ZERO and reports false, so
// callers must validate the address before relying on this to reject one.
//
// ACCOUNT_ZERO decodes cleanly in either address form, but no keypair can produce it,
// so an account named by it can never sign. It stays a legitimate value elsewhere, so
// apply this only to fields naming an account expected to sign or hold a balance.
func IsZeroAccountID(accountID []byte) bool {
	if len(accountID) != AccountAddressLength {
		return false
	}
	for _, b := range accountID {
		if b != 0 {
			return false
		}
	}
	return true
}
