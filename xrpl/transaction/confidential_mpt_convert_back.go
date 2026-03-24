package transaction

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

// ConfidentialMPTConvertBack converts confidential (encrypted) MPT balance back into
// public MPT balance. This requires a zero-knowledge proof (ZKProof) to verify that
// the holder has sufficient confidential balance without revealing the actual amounts.
//
// ```json
//
//	{
//	    "TransactionType": "ConfidentialMPTConvertBack",
//	    "Fee": "10",
//	    "MPTokenIssuanceID": "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
//	    "MPTAmount": "1000",
//	    "HolderEncryptedAmount": "AABB...",
//	    "IssuerEncryptedAmount": "CCDD...",
//	    "BlindingFactor": "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF",
//	    "BalanceCommitment": "EEFF...",
//	    "ZKProof": "1122..."
//	}
//
// ```
type ConfidentialMPTConvertBack struct {
	BaseTx
	// MPTokenIssuanceID identifies the MPTokenIssuance for which to convert balance.
	MPTokenIssuanceID string
	// MPTAmount is the amount of MPT to convert from confidential to public balance.
	// Must be greater than 0.
	MPTAmount types.MPTPlainAmount
	// HolderEncryptedAmount is the encrypted amount being deducted from the holder's confidential balance.
	// 66 bytes (two 33-byte compressed EC points), hex-encoded.
	HolderEncryptedAmount string
	// IssuerEncryptedAmount is the encrypted amount for the issuer's tracking purposes.
	// 66 bytes (two 33-byte compressed EC points), hex-encoded.
	IssuerEncryptedAmount string
	// BlindingFactor is the blinding factor used in the Pedersen commitment.
	// Required for proof verification.
	BlindingFactor string
	// AuditorEncryptedAmount is the encrypted amount for the auditor (if configured). (Optional)
	// 66 bytes (two 33-byte compressed EC points), hex-encoded.
	AuditorEncryptedAmount *string `json:",omitempty"`
	// BalanceCommitment is the Pedersen commitment to the holder's remaining balance after conversion.
	// Required for balance verification.
	BalanceCommitment string
	// ZKProof is a zero-knowledge proof proving the holder has sufficient confidential
	// balance and that the conversion is valid.
	ZKProof string
}

// TxType returns the transaction type (ConfidentialMPTConvertBack).
func (*ConfidentialMPTConvertBack) TxType() TxType {
	return ConfidentialMPTConvertBackTx
}

// Flatten returns the flattened map of the ConfidentialMPTConvertBack transaction.
func (tx *ConfidentialMPTConvertBack) Flatten() FlatTransaction {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	flattened["MPTokenIssuanceID"] = tx.MPTokenIssuanceID

	flattened["MPTAmount"] = tx.MPTAmount.Flatten()

	flattened["HolderEncryptedAmount"] = tx.HolderEncryptedAmount

	flattened["IssuerEncryptedAmount"] = tx.IssuerEncryptedAmount

	flattened["BlindingFactor"] = tx.BlindingFactor

	if tx.AuditorEncryptedAmount != nil {
		flattened["AuditorEncryptedAmount"] = *tx.AuditorEncryptedAmount
	}

	flattened["BalanceCommitment"] = tx.BalanceCommitment

	flattened["ZKProof"] = tx.ZKProof

	return flattened
}

// Validate validates the ConfidentialMPTConvertBack transaction.
func (tx *ConfidentialMPTConvertBack) Validate() (bool, error) {
	ok, err := tx.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	if tx.MPTokenIssuanceID == "" {
		return false, ErrConfidentialMPTInvalidIssuanceID
	}

	if tx.MPTAmount == 0 {
		return false, ErrConfidentialConvertBackInvalidAmount
	}

	if !IsValidBlindingFactor(tx.BlindingFactor) {
		return false, ErrConfidentialConvertBackInvalidBlindingFactor
	}

	if !IsValidHexBlob(tx.HolderEncryptedAmount) || !IsValidHexBlob(tx.IssuerEncryptedAmount) ||
		!IsValidHexBlob(tx.BalanceCommitment) || !IsValidHexBlob(tx.ZKProof) {
		return false, ErrConfidentialConvertBackMissingFields
	}

	if tx.AuditorEncryptedAmount != nil && !IsValidHexBlob(*tx.AuditorEncryptedAmount) {
		return false, ErrConfidentialConvertBackMissingFields
	}

	return true, nil
}
