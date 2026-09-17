package transaction

import (
	"bytes"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/flag"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// validateConfidentialMPTBase validates the fields every confidential MPT transaction
// shares and returns the decoded submitting AccountID for the caller to reuse. XLS-96
// defines no transaction-specific flags for these transaction types, so only the
// protocol-wide flags are accepted.
func validateConfidentialMPTBase(tx *BaseTx) ([]byte, error) {
	if !flag.ContainsOnly(tx.Flags, types.TfUniversal) {
		return nil, ErrConfidentialMPTInvalidFlags
	}

	accountID, accountHasTag, err := decodeAddressAccountID(tx.Account)
	if err != nil {
		return nil, ErrInvalidAccount
	}

	// An Account given as a tagged X-address already carries a source tag, so pairing it
	// with an explicit SourceTag names the tag twice. The binary codec rejects that on
	// encode, and Clawback.Validate makes the same check for XLS-39.
	if accountHasTag && tx.SourceTag != 0 {
		return nil, ErrDuplicateXAddressTag
	}

	return accountID, nil
}

// validateConfidentialMPTHolder validates an issuance ID submitted by the holder in
// accountID and returns the issuer AccountID for the caller to reuse. XLS-96 forbids the
// issuer from holding a confidential balance of its own issuance.
func validateConfidentialMPTHolder(issuanceID string, accountID []byte) ([]byte, error) {
	issuerID, ok := mptIssuerAccountID(issuanceID)
	if !ok {
		return nil, ErrConfidentialMPTInvalidIssuanceID
	}
	if bytes.Equal(issuerID, accountID) {
		return nil, ErrConfidentialMPTIssuerNotAllowed
	}
	return issuerID, nil
}

// validateConfidentialMPTIssuer validates an issuance ID submitted by the issuer in
// accountID.
func validateConfidentialMPTIssuer(issuanceID string, accountID []byte) error {
	issuerID, ok := mptIssuerAccountID(issuanceID)
	if !ok {
		return ErrConfidentialMPTInvalidIssuanceID
	}
	if !bytes.Equal(issuerID, accountID) {
		return ErrConfidentialMPTIssuerRequired
	}
	return nil
}

// IsValidBlindingFactor reports whether bf is a 32-byte blinding factor encoded as
// 64 hexadecimal characters.
func IsValidBlindingFactor(bf string) bool {
	return isValidFixedHexBlob(bf, BlindingFactorLen)
}

// IsValidSchnorrProof reports whether proof is a 64-byte Schnorr proof of knowledge
// encoded as 128 hexadecimal characters.
func IsValidSchnorrProof(proof string) bool {
	return isValidFixedHexBlob(proof, SchnorrProofLen)
}

// IsValidSendProof reports whether proof is a 946-byte confidential send proof bundle
// encoded as 1892 hexadecimal characters.
func IsValidSendProof(proof string) bool {
	return isValidFixedHexBlob(proof, SendProofLen)
}

// IsValidConvertBackProof reports whether proof is an 816-byte confidential convert back
// proof bundle encoded as 1632 hexadecimal characters.
func IsValidConvertBackProof(proof string) bool {
	return isValidFixedHexBlob(proof, ConvertBackProofLen)
}

// IsValidClawbackProof reports whether proof is a 64-byte compact Clawback sigma proof
// encoded as 128 hexadecimal characters.
func IsValidClawbackProof(proof string) bool {
	return isValidFixedHexBlob(proof, ClawbackProofLen)
}

// IsValidCompressedEncryptionKey reports whether key is a 33-byte compressed secp256k1
// point encoded as 66 hexadecimal characters and lying on the secp256k1 curve. A key that
// is well formed but off the curve is rejected.
// Used for HolderEncryptionKey, IssuerEncryptionKey, and AuditorEncryptionKey per XLS-96.
func IsValidCompressedEncryptionKey(key string) bool {
	return crypto.IsCompressedSECP256K1Point(key)
}

// IsValidCommitment reports whether s is a 33-byte Pedersen commitment encoded as
// 66 hexadecimal characters and lying on the secp256k1 curve.
func IsValidCommitment(s string) bool {
	return crypto.IsCompressedSECP256K1Point(s)
}

// IsValidCiphertext reports whether s is a 66-byte EC-ElGamal ciphertext: two
// concatenated compressed secp256k1 points encoded as 132 hexadecimal characters.
// XLS-96 rejects ciphertexts that are well formed but do not lie on the curve.
func IsValidCiphertext(s string) bool {
	return len(s) == CiphertextLen &&
		crypto.IsCompressedSECP256K1Point(s[:CompressedPointLen]) &&
		crypto.IsCompressedSECP256K1Point(s[CompressedPointLen:])
}
