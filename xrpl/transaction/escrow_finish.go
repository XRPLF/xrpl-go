package transaction

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// EscrowFinish delivers XRP from a held payment to the recipient.
//
// Example:
//
// ```json
//
//	{
//	    "Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
//	    "TransactionType": "EscrowFinish",
//	    "Owner": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
//	    "OfferSequence": 7,
//	    "Condition": "A0258020E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855810100",
//	    "Fulfillment": "A0028000"
//	}
//
// ````
type EscrowFinish struct {
	BaseTx
	// Set of Credentials to authorize a deposit made by this transaction.
	// Each member of the array must be the ledger entry ID of a Credential entry in the ledger.
	// For details see https://xrpl.org/docs/references/protocol/transactions/types/payment#credential-ids
	CredentialIDs types.CredentialIDs `json:",omitempty"`
	// Address of the source account that funded the held payment.
	Owner types.Address
	// Transaction sequence of EscrowCreate transaction that created the held payment to finish.
	OfferSequence uint32
	// (Optional) Hex value matching the previously-supplied PREIMAGE-SHA-256 crypto-condition of the held payment.
	Condition string `json:",omitempty"`
	// Optional) Hex value of the PREIMAGE-SHA-256 crypto-condition fulfillment matching the held payment's Condition.
	Fulfillment string `json:",omitempty"`
}

// TxType returns the transaction type for this transaction (EscrowFinish).
func (*EscrowFinish) TxType() TxType {
	return EscrowFinishTx
}

// Flatten returns the flattened map of the EscrowFinish transaction.
func (e *EscrowFinish) Flatten() FlatTransaction {
	flattened := e.BaseTx.Flatten()

	flattened["TransactionType"] = "EscrowFinish"

	if e.Owner != "" {
		flattened["Owner"] = e.Owner.String()
	}

	if e.OfferSequence != 0 {
		flattened["OfferSequence"] = e.OfferSequence
	}

	if e.Condition != "" {
		flattened["Condition"] = e.Condition
	}

	if e.Fulfillment != "" {
		flattened["Fulfillment"] = e.Fulfillment
	}

	if len(e.CredentialIDs) > 0 {
		flattened["CredentialIDs"] = e.CredentialIDs.Flatten()
	}

	return flattened
}

// Validate checks if the EscrowFinish struct is valid.
func (e *EscrowFinish) Validate() (bool, error) {
	ok, err := e.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	if !addresscodec.IsValidAddress(e.Owner.String()) {
		return false, ErrEscrowFinishMissingOwner
	}

	if e.OfferSequence == 0 {
		return false, ErrEscrowFinishMissingOfferSequence
	}

	if e.CredentialIDs != nil && !e.CredentialIDs.IsValid() {
		return false, ErrInvalidCredentialIDs
	}

	return true, nil
}
