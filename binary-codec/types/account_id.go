//revive:disable:var-naming
package types

import (
	"fmt"
	"strings"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
)

// AccountID struct represents an account ID.
type AccountID struct{}

// FromJSON is a method for the AccountID type that takes a value as a parameter,
// serializes it to a byte slice representing an AccountID.
// Decodes the ClassicAddress to an AccountID and returns the byte representation.
// AccountIDs that appear as stand-alone fields (such as Account and Destination)
// are length-prefixed despite being a fixed 160 bits in length. As a result,
// the length indicator for these fields is always the byte 0x14.
// AccountIDs that appear as children of special fields (Amount issuer and PathSet account) are not length-prefixed.
// So in Amount and PathSet fields, don't use the length indicator 0x14.
func (a *AccountID) FromJSON(value any) ([]byte, error) {
	strValue, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("expected a string but got %T", value)
	}

	switch {
	case strings.HasPrefix(strValue, "r"):
		_, accountID, err := addresscodec.DecodeClassicAddressToAccountID(strValue)
		if err != nil {
			return nil, err
		}
		return accountID, nil

	case strings.HasPrefix(strValue, "X"):
		accountID, _, _, err := addresscodec.DecodeXAddress(strValue)
		if err != nil {
			return nil, err
		}
		return accountID, nil

	default:
		return nil, addresscodec.ErrInvalidAddressFormat
	}
}

// ToJSON is a method for the AccountID type that deserializes a byte slice
// representation of an AccountID into a JSON value.
// It takes a binary parser and an optional length prefix size as arguments.
// The method reads the bytes using the binary parser,
// then encodes the result to an AccountID format.
// If no length prefix size is given, it returns an ErrNoLengthPrefix error.
// If the length is 0, it returns an empty string.
func (a *AccountID) ToJSON(p interfaces.BinaryParser, opts ...int) (any, error) {
	if opts == nil {
		return nil, ErrNoLengthPrefix
	}

	if opts[0] == 0 {
		return "", nil
	}

	b, err := p.ReadBytes(opts[0])
	if err != nil {
		return nil, err
	}
	return addresscodec.Encode(b, []byte{addresscodec.AccountAddressPrefix}, addresscodec.AccountAddressLength)
}
