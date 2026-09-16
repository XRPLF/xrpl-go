//revive:disable:var-naming
package types

import "math/big"

const (
	// TfFullyCanonicalSig requires a canonical transaction signature.
	TfFullyCanonicalSig uint32 = 0x80000000
	// TfInnerBatchTxn identifies a transaction as an inner Batch transaction.
	TfInnerBatchTxn uint32 = 0x40000000
	// TfUniversal is the mask of all protocol-wide transaction flags.
	TfUniversal uint32 = TfFullyCanonicalSig | TfInnerBatchTxn
)

// SponsorFlags values for the sponsored-transaction common field SponsorFlags.
const (
	// SpfSponsorFee requests that the sponsor pays the transaction fee.
	SpfSponsorFee uint32 = 0x00000001
	// SpfSponsorReserve requests that the sponsor pays the reserve of any account or object the
	// transaction creates.
	SpfSponsorReserve uint32 = 0x00000002
	// SpfSponsorUniversal is the mask of all defined sponsorship flags.
	SpfSponsorUniversal uint32 = SpfSponsorFee | SpfSponsorReserve
)

// IsFlagEnabled performs bitwise AND (&) to check if a flag is enabled within Flags (as a number).
func IsFlagEnabled(flags, checkFlag uint32) bool {
	flagsBigInt := new(big.Int).SetUint64(uint64(flags))
	checkFlagBigInt := new(big.Int).SetUint64(uint64(checkFlag))
	return new(big.Int).And(flagsBigInt, checkFlagBigInt).Cmp(checkFlagBigInt) == 0
}
