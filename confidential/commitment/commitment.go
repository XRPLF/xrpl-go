// Package commitment provides a hex-string API for Pedersen commitment creation.
// It wraps the byte-array function in mptcrypto with hex encoding/decoding.
package commitment

import (
	"encoding/hex"
	"fmt"

	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
)

// Create computes a Pedersen commitment for the given amount and blinding factor.
// bfHex must be 64 hex chars (32 bytes). Returns 66 hex chars (33-byte compressed point).
func Create(amount uint64, bfHex string) (string, error) {
	bfBytes, err := hexutil.DecodeFixedHex(bfHex, mptcrypto.BlindingFactorSize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidBlindingFactor, err)
	}

	var bf [mptcrypto.BlindingFactorSize]byte
	copy(bf[:], bfBytes)

	c, err := mptcrypto.PedersenCommitment(amount, bf)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrCommitmentFailed, err)
	}
	return hex.EncodeToString(c[:]), nil
}
