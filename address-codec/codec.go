package addresscodec

import (
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/Peersyst/xrpl-go/address-codec/interfaces"
	"github.com/Peersyst/xrpl-go/pkg/crypto"
)

const (
	// AccountAddressLength is the length in bytes of a classic account address.
	AccountAddressLength = 20

	// AccountPublicKeyLength is the length in bytes of an account public key.
	AccountPublicKeyLength = 33

	// FamilySeedLength is the length in bytes of a family seed.
	FamilySeedLength = 16

	// NodePublicKeyLength is the length in bytes of a node/validation public key.
	NodePublicKeyLength = 33

	// AccountAddressPrefix is the classic address prefix (0x00).
	AccountAddressPrefix = 0x00

	// AccountPublicKeyPrefix is the prefix for account public keys (0x23).
	AccountPublicKeyPrefix = 0x23

	// FamilySeedPrefix is the prefix for family seeds (0x21).
	FamilySeedPrefix = 0x21

	// NodePublicKeyPrefix is the prefix for node/validation public keys (0x1C).
	NodePublicKeyPrefix = 0x1C
)

// Encode returns the Base58Check encoding of a byte slice with the given type prefix,
// ensuring the byte slice has the expected length.
func Encode(b []byte, typePrefix []byte, expectedLength int) (string, error) {

	if len(b) != expectedLength {
		return "", &EncodeLengthError{Instance: "Encode", Expected: expectedLength, Input: len(b)}
	}

	return Base58CheckEncode(b, typePrefix...), nil
}

// Decode returns the decoded byte slice of the base58-encoded string for the given prefix.
func Decode(b58string string, typePrefix []byte) ([]byte, error) {

	prefixLength := len(typePrefix)

	if !bytes.Equal(DecodeBase58(b58string)[:prefixLength], typePrefix) {
		return nil, errors.New("b58string prefix and typeprefix not equal")
	}

	result, err := Base58CheckDecode(b58string)
	result = result[prefixLength:]

	return result, err
}

// EncodeClassicAddressFromPublicKeyHex returns the classic address from a public key hex string.
func EncodeClassicAddressFromPublicKeyHex(pubkeyhex string) (string, error) {

	pubkey, err := hex.DecodeString(pubkeyhex)

	if err != nil {
		return "", err
	} else if len(pubkey) != AccountPublicKeyLength {
		return "", &EncodeLengthError{Instance: "PublicKey", Expected: AccountPublicKeyLength, Input: len(pubkey)}
	}

	accountID := Sha256RipeMD160(pubkey)

	address, err := Encode(accountID, []byte{AccountAddressPrefix}, AccountAddressLength)
	if err != nil {
		return "", err
	}

	if !IsValidClassicAddress(address) {
		return "", ErrInvalidClassicAddress
	}

	return address, nil
}

// DecodeClassicAddressToAccountID returns the prefix and accountID byte slice from a classic address.
func DecodeClassicAddressToAccountID(cAddress string) (typePrefix, accountID []byte, err error) {
	// Use Base58CheckDecode to validate checksum
	decoded, err := Base58CheckDecode(cAddress)
	if err != nil {
		return nil, nil, ErrInvalidClassicAddress
	}

	// Expected length is 21 bytes (1 prefix + 20 accountID) after removing 4-byte checksum
	if len(decoded) != 21 {
		return nil, nil, ErrInvalidClassicAddress
	}

	return decoded[:1], decoded[1:21], nil
}

// EncodeAccountIDToClassicAddress returns the classic address encoding of the accountId.
func EncodeAccountIDToClassicAddress(accountID []byte) (string, error) {
	if len(accountID) != AccountAddressLength {
		return "", ErrInvalidAccountID
	}

	return Base58CheckEncode(accountID, AccountAddressPrefix), nil
}

// EncodeSeed returns a base58 encoding of a seed using the specified encoding type.
func EncodeSeed(entropy []byte, encodingType interfaces.CryptoImplementation) (string, error) {

	if len(entropy) != FamilySeedLength {
		return "", &EncodeLengthError{Instance: "Entropy", Input: len(entropy), Expected: FamilySeedLength}
	}

	if ed25519 := crypto.ED25519(); encodingType == ed25519 {
		return Encode(entropy, ed25519.FamilySeedPrefix(), FamilySeedLength)
	} else if secp256k1 := crypto.SECP256K1(); encodingType == secp256k1 {
		return Encode(entropy, secp256k1.FamilySeedPrefix(), FamilySeedLength)
	}
	return "", errors.New("encoding type must be `ed25519` or `secp256k1`")

}

// DecodeSeed returns the decoded seed and its corresponding algorithm.
func DecodeSeed(seed string) ([]byte, interfaces.CryptoImplementation, error) {

	// decoded := DecodeBase58(seed)
	decoded, err := Base58CheckDecode(seed)

	if err != nil {
		return nil, nil, ErrInvalidSeed
	}

	if bytes.Equal(decoded[:3], []byte{0x01, 0xe1, 0x4b}) {
		return decoded[3:], crypto.ED25519(), nil
	}

	return decoded[1:], crypto.SECP256K1(), nil

}

// EncodeNodePublicKey returns the base58 encoding of a node public key byte slice.
func EncodeNodePublicKey(b []byte) (string, error) {

	if len(b) != NodePublicKeyLength {
		return "", &EncodeLengthError{Instance: "NodePublicKey", Expected: NodePublicKeyLength, Input: len(b)}
	}

	npk := Base58CheckEncode(b, NodePublicKeyPrefix)

	return npk, nil
}

// DecodeNodePublicKey returns the decoded node public key byte slice from a base58 string.
func DecodeNodePublicKey(key string) ([]byte, error) {

	decodedNodeKey, err := Decode(key, []byte{NodePublicKeyPrefix})
	if err != nil {
		return nil, err
	}

	return decodedNodeKey, nil
}

// EncodeAccountPublicKey returns the base58 encoding of an account public key byte slice.
func EncodeAccountPublicKey(b []byte) (string, error) {

	if len(b) != AccountPublicKeyLength {
		return "", &EncodeLengthError{Instance: "AccountPublicKey", Expected: AccountPublicKeyLength, Input: len(b)}
	}

	apk := Base58CheckEncode(b, AccountPublicKeyPrefix)

	return apk, nil
}

// DecodeAccountPublicKey returns the decoded account public key byte slice from a base58 string.
func DecodeAccountPublicKey(key string) ([]byte, error) {

	decodedAccountKey, err := Decode(key, []byte{AccountPublicKeyPrefix})
	if err != nil {
		return nil, err
	}

	return decodedAccountKey, nil
}

// IsValidAddress returns true if the address is valid. Otherwise, it returns false.
// Address can only be a classic address or an x-address.
func IsValidAddress(address string) bool {
	return IsValidClassicAddress(address) || IsValidXAddress(address)
}

// IsValidClassicAddress returns true if the classic address is valid. Otherwise, it returns false.
func IsValidClassicAddress(cAddress string) bool {
	_, _, err := DecodeClassicAddressToAccountID(cAddress)
	return err == nil
}
