package transaction

import (
	"bytes"
	"fmt"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ConfidentialMPTSend sends confidential MPT from one account to another
// without revealing the transfer amount publicly. The transferred amount is
// credited to the receiver's inbox balance (CB_IN).
// It requires the ConfidentialTransfer amendment.
//
// The amount is encrypted for the sender, destination, issuer, and optionally
// an auditor. A zero-knowledge proof verifies that the sender has sufficient
// balance and that all encrypted amounts are consistent.
//
// ```json
//
//	{
//	    "TransactionType": "ConfidentialMPTSend",
//	    "Account": "r...",
//	    "MPTokenIssuanceID": "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
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
	// DestinationTag identifies a hosted-account destination. (Optional)
	// XLS-96 does not spell this field out. It comes from the generic transaction fields
	// rippled declares for ConfidentialMPTSend in transactions.macro, and it carries the
	// protocol-wide destination-tag semantics. A destination with lsfRequireDestTag set
	// rejects a transaction that omits it with tecDST_TAG_NEEDED.
	DestinationTag *uint32 `json:",omitempty"`
	// SenderEncryptedAmount is the encrypted transfer amount for the sender.
	// 66 bytes (two 33-byte compressed EC points), hex-encoded.
	SenderEncryptedAmount string
	// DestinationEncryptedAmount is the encrypted transfer amount for the destination.
	// 66 bytes (two 33-byte compressed EC points), hex-encoded.
	DestinationEncryptedAmount string
	// IssuerEncryptedAmount is the encrypted transfer amount for the issuer's tracking purposes.
	// 66 bytes (two 33-byte compressed EC points), hex-encoded.
	IssuerEncryptedAmount string
	// AuditorEncryptedAmount is the encrypted amount for the issuance's auditor. (Optional)
	// It is required if and only if the issuance has an AuditorEncryptionKey set. Omitting it
	// for an issuance that has one, supplying it for an issuance that has none, or encrypting
	// it under any key other than the issuance's auditor key all fail with tecNO_PERMISSION.
	// 66 bytes (two 33-byte compressed EC points), hex-encoded.
	AuditorEncryptedAmount *string `json:",omitempty"`
	// ZKProof is a zero-knowledge proof proving the sender has sufficient balance
	// and that all encrypted amounts are consistent.
	ZKProof string
	// AmountCommitment is the Pedersen commitment to the transfer amount.
	// Required for proof verification.
	AmountCommitment string
	// BalanceCommitment is the Pedersen commitment to the sender's current confidential spending balance.
	BalanceCommitment string
	// CredentialIDs is a set of Credential IDs that may be required for authorized transfers. (Optional)
	// This field also requires the Credentials amendment.
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

	if tx.DestinationTag != nil {
		flattened["DestinationTag"] = *tx.DestinationTag
	}

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
	accountID, err := validateConfidentialMPTBase(&tx.BaseTx)
	if err != nil {
		return false, err
	}
	issuerID, err := validateConfidentialMPTHolder(tx.MPTokenIssuanceID, accountID)
	if err != nil {
		return false, err
	}

	destinationID, sameAccount, destHasTag, err := decodeCounterparty(accountID, tx.Destination)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrConfidentialSendInvalidDestination, err)
	}
	if sameAccount {
		return false, ErrConfidentialSendSelfSend
	}
	if destHasTag && tx.DestinationTag != nil {
		return false, ErrDuplicateXAddressTag
	}
	if bytes.Equal(issuerID, destinationID) {
		return false, ErrConfidentialSendDestinationIsIssuer
	}

	if !IsValidCiphertext(tx.SenderEncryptedAmount) || !IsValidCiphertext(tx.DestinationEncryptedAmount) ||
		!IsValidCiphertext(tx.IssuerEncryptedAmount) {
		return false, ErrConfidentialSendInvalidCiphertext
	}

	if tx.AuditorEncryptedAmount != nil && !IsValidCiphertext(*tx.AuditorEncryptedAmount) {
		return false, ErrConfidentialSendInvalidCiphertext
	}

	if !IsValidCommitment(tx.BalanceCommitment) || !IsValidCommitment(tx.AmountCommitment) {
		return false, ErrConfidentialSendInvalidCommitment
	}

	if !IsValidSendProof(tx.ZKProof) {
		return false, ErrConfidentialSendInvalidProof
	}

	if tx.CredentialIDs != nil && !tx.CredentialIDs.IsValid() {
		return false, ErrInvalidCredentialIDs
	}

	return true, nil
}
