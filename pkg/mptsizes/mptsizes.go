// Package mptsizes holds the byte sizes of the XLS-96 Confidential MPT wire primitives,
// matching the defines of the vendored XRPLF/mpt-crypto C library. The primitive, ledger
// and bulletproof sizes come from mpt_protocol.h, the compact proof sizes from
// secp256k1_mpt.h.
//
// The sizes live here, free of CGo, so that both the CGo bindings in
// confidential/mptcrypto and the transaction models in xrpl/transaction can derive from
// one definition. Importing confidential/mptcrypto from xrpl/transaction would link the C
// library into every consumer of the core package.
//
// Because the values are transcribed rather than read from the headers,
// confidential/mptcrypto pins every constant below to its define with a compile-time
// assertion. A vendored header that changes a size breaks that build until this package
// follows.
package mptsizes

// Crypto primitive sizes in bytes, from mpt_protocol.h.
const (
	PrivKeySize        = 32
	PubKeySize         = 33
	BlindingFactorSize = 32
	CiphertextSize     = 66 // two compressed EC points (C1 || C2)
)

// Ledger and hash sizes in bytes, from mpt_protocol.h.
const (
	AccountIDSize  = 20
	IssuanceIDSize = 24
	HashOutputSize = 32 // kMPT_HALF_SHA_SIZE -- output size of context hash functions
	CommitmentSize = 33 // compressed Pedersen commitment point
)

// Proof sizes in bytes. The Schnorr and bulletproof sizes come from mpt_protocol.h,
// the compact sigma proof sizes from secp256k1_mpt.h.
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
