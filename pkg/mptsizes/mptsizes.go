// Package mptsizes holds the byte sizes of the XLS-96 Confidential MPT wire primitives,
// matching the mpt_protocol.h defines of the vendored XRPLF/mpt-crypto C library.
//
// The sizes live here, free of CGo, so that both the CGo bindings in
// confidential/mptcrypto and the transaction models in xrpl/transaction can derive from
// one definition. Importing confidential/mptcrypto from xrpl/transaction would link the C
// library into every consumer of the core package.
package mptsizes

// Crypto primitive sizes in bytes.
const (
	PrivKeySize        = 32
	PubKeySize         = 33
	BlindingFactorSize = 32
	CiphertextSize     = 66 // two compressed EC points (C1 || C2)
)

// Ledger and hash sizes in bytes.
const (
	AccountIDSize  = 20
	IssuanceIDSize = 24
	HashOutputSize = 32 // kMPT_HALF_SHA_SIZE -- output size of context hash functions
	CommitmentSize = 33 // compressed Pedersen commitment point
)

// Proof sizes in bytes.
const (
	SchnorrProofSize            = 64
	SingleBulletproofSize       = 688
	DoubleBulletproofSize       = 754
	CompactClawbackProofSize    = 64
	CompactConvertBackProofSize = 128
	CompactSendProofSize        = 192
	ConvertBackProofSize        = CompactConvertBackProofSize + SingleBulletproofSize
	SendProofSize               = CompactSendProofSize + DoubleBulletproofSize
)
