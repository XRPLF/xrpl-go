package addresscodec

import (
	"bytes"
)

var (
	// MainnetXAddressPrefix is the prefix for mainnet X-address encoding.
	MainnetXAddressPrefix = []byte{0x05, 0x44}
	// TestnetXAddressPrefix is the prefix for testnet X-address encoding.
	TestnetXAddressPrefix = []byte{0x04, 0x93}
	// XAddressLength is the length of an X-address (35 bytes).
	XAddressLength = 35
)

// IsValidXAddress returns true if the x-address is valid. Otherwise, it returns false.
func IsValidXAddress(xAddress string) bool {
	_, _, _, err := DecodeXAddress(xAddress)
	return err == nil
}

// EncodeXAddress returns the x-address encoding of the accountId, tag, and testnet boolean.
// If the accountId is not 20 bytes long, it returns an error.
func EncodeXAddress(accountID []byte, tag uint32, tagFlag, testnetFlag bool) (string, error) {
	if len(accountID) != AccountAddressLength {
		return "", ErrInvalidAccountID
	}

	xAddressBytes := make([]byte, 0, XAddressLength)

	if testnetFlag {
		xAddressBytes = append(xAddressBytes, TestnetXAddressPrefix...)
	} else {
		xAddressBytes = append(xAddressBytes, MainnetXAddressPrefix...)
	}

	xAddressBytes = append(xAddressBytes, accountID...)

	if tagFlag {
		xAddressBytes = append(xAddressBytes, byte(1))
	} else {
		xAddressBytes = append(xAddressBytes, byte(0))
	}

	xAddressBytes = append(
		xAddressBytes,
		byte(tag&0xff),
		byte((tag>>8)&0xff),
		byte((tag>>16)&0xff),
		byte((tag>>24)&0xff),
		0,
		0,
		0,
		0,
	)

	cksum := checksum(xAddressBytes)
	xAddressBytes = append(xAddressBytes, cksum[:]...)

	return EncodeBase58(xAddressBytes), nil
}

// DecodeXAddress returns the accountId, tag, and testnet boolean decoding of the x-address.
// If the x-address is invalid, it returns an error.
func DecodeXAddress(xAddress string) (accountID []byte, tag uint32, testnet bool, err error) {
	// Use Base58CheckDecode to validate checksum
	xAddressBytes, err := Base58CheckDecode(xAddress)
	if err != nil {
		return nil, 0, false, err
	}

	// Verify length (2 prefix + 20 accountID + 1 flag + 8 tag bytes = 31)
	if len(xAddressBytes) != 31 {
		return nil, 0, false, ErrInvalidXAddress
	}

	switch {
	case bytes.HasPrefix(xAddressBytes, MainnetXAddressPrefix):
		testnet = false
	case bytes.HasPrefix(xAddressBytes, TestnetXAddressPrefix):
		testnet = true
	default:
		return nil, 0, false, ErrInvalidXAddress
	}

	tag, _, err = decodeTag(xAddressBytes)
	if err != nil {
		return nil, 0, false, err
	}

	return xAddressBytes[2:22], tag, testnet, nil
}

// XAddressToClassicAddress converts the x-address to a classic address.
// It returns the classic address, tag and testnet boolean.
// If the x-address is invalid, it returns an error.
func XAddressToClassicAddress(xAddress string) (classicAddress string, tag uint32, testnet bool, err error) {
	accountID, tag, testnet, err := DecodeXAddress(xAddress)
	if err != nil {
		return "", 0, false, err
	}

	classicAddress, err = EncodeAccountIDToClassicAddress(accountID)
	if err != nil {
		return "", 0, false, err
	}

	return classicAddress, tag, testnet, nil
}

// ClassicAddressToXAddress converts the classic address to an x-address.
// It returns the x-address.
// If the classic address is invalid, it returns an error.
func ClassicAddressToXAddress(address string, tag uint32, tagFlag, testnetFlag bool) (string, error) {
	_, accountID, err := DecodeClassicAddressToAccountID(address)
	if err != nil {
		return "", err
	}

	return EncodeXAddress(accountID, tag, tagFlag, testnetFlag)
}

// decodeTag returns the tag from the x-address.
// If the tag is invalid, it returns an error.
func decodeTag(xAddressBytes []byte) (uint32, bool, error) {
	flag := xAddressBytes[22]
	if flag >= 2 {
		// No support for 64-bit tags at this time
		return 0, false, ErrUnsupportedXAddress
	}
	if flag == 1 {
		// Little-endian to big-endian (4 bytes for full 32-bit tag support)
		tag := uint32(xAddressBytes[23]) +
			uint32(xAddressBytes[24])*0x100 +
			uint32(xAddressBytes[25])*0x10000 +
			uint32(xAddressBytes[26])*0x1000000
		return tag, true, nil
	}
	// flag == 0 means no tag
	// Verify remaining bytes are zero (reserved for 64-bit tags)
	for i := 23; i < 31; i++ {
		if xAddressBytes[i] != 0 {
			return 0, false, ErrInvalidTag
		}
	}
	return 0, false, nil
}
