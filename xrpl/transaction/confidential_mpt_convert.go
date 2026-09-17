package transaction

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

// ConfidentialMPTConvert converts public MPT balance into confidential (encrypted) balance.
// It requires the ConfidentialTransfer amendment.
// The amount being converted is specified in cleartext, but the resulting balance is encrypted
// using EC-ElGamal encryption. On first use, the holder registers their ElGamal encryption key
// by providing HolderEncryptionKey and ZKProof (Schnorr proof of knowledge).
// ZKProof must be present if and only if HolderEncryptionKey is present,
// and must be absent when HolderEncryptionKey is absent.
//
// ```json
//
//	{
//	    "TransactionType": "ConfidentialMPTConvert",
//	    "Account": "r...",
//	    "MPTokenIssuanceID": "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
//	    "MPTAmount": "1000",
//	    "HolderEncryptedAmount": "AABB...",
//	    "IssuerEncryptedAmount": "CCDD...",
//	    "BlindingFactor": "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"
//	}
//
// ```
type ConfidentialMPTConvert struct {
	BaseTx
	// MPTokenIssuanceID identifies the MPTokenIssuance for which to convert balance.
	MPTokenIssuanceID string
	// MPTAmount is the amount of MPT to convert from public to confidential balance.
	MPTAmount types.MPTPlainAmount
	// HolderEncryptionKey is the holder's EC-ElGamal public key for encryption. (Optional)
	// Required if the holder doesn't already have a key registered.
	// 33 bytes compressed EC point, hex-encoded (66 hex chars).
	HolderEncryptionKey *string `json:",omitempty"`
	// HolderEncryptedAmount is the encrypted amount for the holder's confidential balance.
	// 66 bytes (two 33-byte compressed EC points), hex-encoded.
	HolderEncryptedAmount string
	// IssuerEncryptedAmount is the encrypted amount for the issuer's tracking purposes.
	// 66 bytes (two 33-byte compressed EC points), hex-encoded.
	IssuerEncryptedAmount string
	// AuditorEncryptedAmount is the encrypted amount for the issuance's auditor. (Optional)
	// It is required if and only if the issuance has an AuditorEncryptionKey set. Omitting it
	// for an issuance that has one, supplying it for an issuance that has none, or encrypting
	// it under any key other than the issuance's auditor key all fail with tecNO_PERMISSION.
	// 66 bytes (two 33-byte compressed EC points), hex-encoded.
	AuditorEncryptedAmount *string `json:",omitempty"`
	// BlindingFactor is the 32-byte scalar value used to encrypt the amount.
	// Used by validators to verify the ciphertexts match the plaintext MPTAmount.
	BlindingFactor string
	// ZKProof is a Schnorr proof of knowledge. Required if and only if
	// HolderEncryptionKey is present, must be absent otherwise. (Optional)
	ZKProof *string `json:",omitempty"`
}

// TxType returns the transaction type (ConfidentialMPTConvert).
func (*ConfidentialMPTConvert) TxType() TxType {
	return ConfidentialMPTConvertTx
}

// Flatten returns the flattened map of the ConfidentialMPTConvert transaction.
func (tx *ConfidentialMPTConvert) Flatten() FlatTransaction {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	flattened["MPTokenIssuanceID"] = tx.MPTokenIssuanceID

	flattened["MPTAmount"] = tx.MPTAmount.Flatten()

	if tx.HolderEncryptionKey != nil {
		flattened["HolderEncryptionKey"] = *tx.HolderEncryptionKey
	}

	flattened["HolderEncryptedAmount"] = tx.HolderEncryptedAmount

	flattened["IssuerEncryptedAmount"] = tx.IssuerEncryptedAmount

	if tx.AuditorEncryptedAmount != nil {
		flattened["AuditorEncryptedAmount"] = *tx.AuditorEncryptedAmount
	}

	flattened["BlindingFactor"] = tx.BlindingFactor

	if tx.ZKProof != nil {
		flattened["ZKProof"] = *tx.ZKProof
	}

	return flattened
}

// Validate validates the ConfidentialMPTConvert transaction.
func (tx *ConfidentialMPTConvert) Validate() (bool, error) {
	ok, err := tx.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}
	accountID, err := validateConfidentialMPTBase(&tx.BaseTx)
	if err != nil {
		return false, err
	}
	if _, err := validateConfidentialMPTHolder(tx.MPTokenIssuanceID, accountID); err != nil {
		return false, err
	}
	if !tx.MPTAmount.IsValid() {
		return false, ErrConfidentialMPTInvalidAmount
	}

	// HolderEncryptionKey and ZKProof must both be present or both absent.
	hasKey := tx.HolderEncryptionKey != nil
	hasProof := tx.ZKProof != nil
	if hasKey != hasProof {
		return false, ErrConfidentialConvertKeyProofMismatch
	}

	if hasKey && !IsValidCompressedEncryptionKey(*tx.HolderEncryptionKey) {
		return false, ErrConfidentialConvertInvalidEncryptionKey
	}

	if hasProof && !IsValidSchnorrProof(*tx.ZKProof) {
		return false, ErrConfidentialConvertInvalidProofLength
	}

	if !IsValidBlindingFactor(tx.BlindingFactor) {
		return false, ErrConfidentialConvertInvalidBlindingFactor
	}

	if !IsValidCiphertext(tx.HolderEncryptedAmount) || !IsValidCiphertext(tx.IssuerEncryptedAmount) {
		return false, ErrConfidentialConvertInvalidCiphertext
	}

	if tx.AuditorEncryptedAmount != nil && !IsValidCiphertext(*tx.AuditorEncryptedAmount) {
		return false, ErrConfidentialConvertInvalidCiphertext
	}

	return true, nil
}
