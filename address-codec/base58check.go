package addresscodec

import (
	"crypto/sha256"
	"crypto/subtle"
)

// checksum: first four bytes of sha256^2
func checksum(input []byte) (cksum [4]byte) {
	h := sha256.Sum256(input)
	h2 := sha256.Sum256(h[:])
	copy(cksum[:], h2[:4])
	return cksum
}

// Base58CheckEncode prepends a version byte, appends a four-byte checksum, and returns
// the Base58Check encoding of the input byte slice.
func Base58CheckEncode(input []byte, prefix ...byte) string {
	b := make([]byte, 0, 1+len(input)+4)
	b = append(b, prefix...)
	b = append(b, input...)

	cksum := checksum(b)
	b = append(b, cksum[:]...)
	return EncodeBase58(b)
}

// Base58CheckDecode decodes a Base58Check encoded string and verifies the checksum.
func Base58CheckDecode(input string) (result []byte, err error) {
	decoded := DecodeBase58(input)
	if len(decoded) < 5 {
		return nil, ErrInvalidFormat
	}

	result = decoded[:len(decoded)-4]
	actualChecksum := decoded[len(decoded)-4:]
	expectedChecksum := checksum(result)
	if len(actualChecksum) != len(expectedChecksum) || subtle.ConstantTimeCompare(actualChecksum, expectedChecksum[:]) != 1 {
		return nil, ErrChecksum
	}
	return
}
