//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package mptcrypto

/*
#cgo CFLAGS: -I${SRCDIR}/../deps/include -I${SRCDIR}/../deps/include/utility
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/../deps/libs/linux-amd64 -lmpt-crypto -lsecp256k1 -lcrypto -lstdc++ -lz -lm -ldl -lpthread
#cgo linux,arm64 LDFLAGS: -L${SRCDIR}/../deps/libs/linux-arm64 -lmpt-crypto -lsecp256k1 -lcrypto -lstdc++ -lz -lm -ldl -lpthread
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/../deps/libs/darwin-arm64 -lmpt-crypto -lsecp256k1 -lcrypto -lc++ -lz -lm
#cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/../deps/libs/darwin-amd64 -lmpt-crypto -lsecp256k1 -lcrypto -lc++ -lz -lm

#include "mpt_utility.h"
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// GenerateKeypair creates a new secp256k1 ElGamal keypair.
// Returns a 32-byte private key and a 33-byte compressed public key.
func GenerateKeypair() (privkey [PrivKeySize]byte, pubkey [PubKeySize]byte, err error) {
	ret := C.mpt_generate_keypair(
		(*C.uint8_t)(unsafe.Pointer(&privkey[0])),
		(*C.uint8_t)(unsafe.Pointer(&pubkey[0])),
	)
	if ret != 0 {
		return privkey, pubkey, fmt.Errorf("mpt_generate_keypair failed with code %d", ret)
	}
	return
}

// GenerateBlindingFactor returns a random 32-byte scalar suitable for ElGamal encryption.
func GenerateBlindingFactor() (bf [BlindingFactorSize]byte, err error) {
	ret := C.mpt_generate_blinding_factor(
		(*C.uint8_t)(unsafe.Pointer(&bf[0])),
	)
	if ret != 0 {
		return bf, fmt.Errorf("mpt_generate_blinding_factor failed with code %d", ret)
	}
	return
}

// EncryptAmount encrypts a uint64 amount under a compressed public key using a blinding factor.
// Returns a 66-byte ciphertext (two compressed EC points: C1 || C2).
func EncryptAmount(amount uint64, pubkey [PubKeySize]byte, bf [BlindingFactorSize]byte) (ct [CiphertextSize]byte, err error) {
	ret := C.mpt_encrypt_amount(
		C.uint64_t(amount),
		(*C.uint8_t)(unsafe.Pointer(&pubkey[0])),
		(*C.uint8_t)(unsafe.Pointer(&bf[0])),
		(*C.uint8_t)(unsafe.Pointer(&ct[0])),
	)
	if ret != 0 {
		return ct, fmt.Errorf("mpt_encrypt_amount failed with code %d", ret)
	}
	return
}

// DecryptAmount decrypts a 66-byte ElGamal ciphertext using a private key.
// Returns the plaintext uint64 amount.
func DecryptAmount(ciphertext [CiphertextSize]byte, privkey [PrivKeySize]byte) (uint64, error) {
	var amount C.uint64_t
	ret := C.mpt_decrypt_amount(
		(*C.uint8_t)(unsafe.Pointer(&ciphertext[0])),
		(*C.uint8_t)(unsafe.Pointer(&privkey[0])),
		&amount,
	)
	if ret != 0 {
		return 0, fmt.Errorf("mpt_decrypt_amount failed with code %d", ret)
	}
	return uint64(amount), nil
}
