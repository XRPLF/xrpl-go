//go:build cgo

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
