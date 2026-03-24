package transaction

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ConfidentialMPTSend sends confidential MPT from one account to another
// without revealing the transfer amount publicly. The transferred amount is
// credited to the receiver's inbox balance (CB_IN).
//
// The amount is encrypted for the sender, destination, issuer, and optionally
// an auditor. A zero-knowledge proof verifies that the sender has sufficient
// balance and that all encrypted amounts are consistent.
//
// ```json
//
//	{
//	    "TransactionType": "ConfidentialMPTSend",
//	    "Fee": "10",
//	    "MPTokenIssuanceID": "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
//	    "Destination": "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
//	    "SenderEncryptedAmount": "AABB...",
//	    "DestinationEncryptedAmount": "CCDD...",
//	    "IssuerEncryptedAmount": "EEFF...",
//	    "ZKProof": "1122...",
//	    "AmountCommitment": "3344...",
//	    "BalanceCommitment": "5566..."
//	}
//
// ```
type ConfidentialMPTSend struct {
	BaseTx
	// MPTokenIssuanceID identifies the MPTokenIssuance being transferred.
	MPTokenIssuanceID string
	// Destination is the account receiving the confidential MPT.
	Destination types.Address
	// SenderEncryptedAmount is the encrypted transfer amount for the sender.
	// 66 bytes (two 33-byte compressed EC points), hex-encoded.
	SenderEncryptedAmount string
	// DestinationEncryptedAmount is the encrypted transfer amount for the destination.
	// 66 bytes (two 33-byte compressed EC points), hex-encoded.
	DestinationEncryptedAmount string
	// IssuerEncryptedAmount is the encrypted transfer amount for the issuer's tracking purposes.
	// 66 bytes (two 33-byte compressed EC points), hex-encoded.
	IssuerEncryptedAmount string
	// AuditorEncryptedAmount is the encrypted amount for the auditor (if configured). (Optional)
	// 66 bytes (two 33-byte compressed EC points), hex-encoded.
	AuditorEncryptedAmount *string `json:",omitempty"`
	// ZKProof is a zero-knowledge proof proving the sender has sufficient balance
	// and that all encrypted amounts are consistent.
	ZKProof string
	// AmountCommitment is the Pedersen commitment to the transfer amount.
	// Required for proof verification.
	AmountCommitment string
	// BalanceCommitment is the Pedersen commitment to the sender's remaining balance after transfer.
	BalanceCommitment string
	// CredentialIDs is a set of Credential IDs that may be required for authorized transfers. (Optional)
	CredentialIDs types.CredentialIDs `json:",omitempty"`
}

// TxType returns the transaction type (ConfidentialMPTSend).
func (*ConfidentialMPTSend) TxType() TxType {
	return ConfidentialMPTSendTx
}

// Flatten returns the flattened map of the ConfidentialMPTSend transaction.
func (tx *ConfidentialMPTSend) Flatten() FlatTransaction {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	flattened["MPTokenIssuanceID"] = tx.MPTokenIssuanceID

	flattened["Destination"] = tx.Destination.String()

	flattened["SenderEncryptedAmount"] = tx.SenderEncryptedAmount

	flattened["DestinationEncryptedAmount"] = tx.DestinationEncryptedAmount

	flattened["IssuerEncryptedAmount"] = tx.IssuerEncryptedAmount

	if tx.AuditorEncryptedAmount != nil {
		flattened["AuditorEncryptedAmount"] = *tx.AuditorEncryptedAmount
	}

	flattened["ZKProof"] = tx.ZKProof

	flattened["AmountCommitment"] = tx.AmountCommitment

	flattened["BalanceCommitment"] = tx.BalanceCommitment

	if len(tx.CredentialIDs) > 0 {
		flattened["CredentialIDs"] = tx.CredentialIDs.Flatten()
	}

	return flattened
}

// Validate validates the ConfidentialMPTSend transaction.
func (tx *ConfidentialMPTSend) Validate() (bool, error) {
	ok, err := tx.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	if tx.MPTokenIssuanceID == "" {
		return false, ErrConfidentialMPTInvalidIssuanceID
	}

	if !addresscodec.IsValidAddress(tx.Destination.String()) {
		return false, ErrConfidentialSendInvalidDestination
	}

	if tx.Destination.String() == tx.Account.String() {
		return false, ErrConfidentialSendSelfSend
	}

	if !IsValidHexBlob(tx.SenderEncryptedAmount) || !IsValidHexBlob(tx.DestinationEncryptedAmount) ||
		!IsValidHexBlob(tx.IssuerEncryptedAmount) || !IsValidHexBlob(tx.ZKProof) ||
		!IsValidHexBlob(tx.BalanceCommitment) || !IsValidHexBlob(tx.AmountCommitment) {
		return false, ErrConfidentialSendMissingFields
	}

	if tx.AuditorEncryptedAmount != nil && !IsValidHexBlob(*tx.AuditorEncryptedAmount) {
		return false, ErrConfidentialSendMissingFields
	}

	if tx.CredentialIDs != nil && !tx.CredentialIDs.IsValid() {
		return false, ErrInvalidCredentialIDs
	}

	return true, nil
}
