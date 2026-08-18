package transaction

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

// ConfidentialMPTClawback requires the ConfidentialTransfer amendment.
// It burns a holder's entire confidential MPT balance and reduces
// outstanding supply. The issuer must provide a ZK equality proof demonstrating knowledge
// of the encrypted balance since balances are not visible. Account must be the issuer, but
// an authorized Delegate can submit the transaction. The issuance must have LsfMPTCanClawback enabled.
//
// ```json
//
//	{
//	    "TransactionType": "ConfidentialMPTClawback",
//	    "Account": "r...",
//	    "Holder": "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
//	    "MPTokenIssuanceID": "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
//	    "MPTAmount": "1000",
//	    "ZKProof": "AABB..."
//	}
//
// ```
type ConfidentialMPTClawback struct {
	BaseTx
	// MPTokenIssuanceID identifies the MPTokenIssuance from which to clawback.
	MPTokenIssuanceID string
	// Holder is the holder account from which to clawback confidential balance.
	Holder types.Address
	// MPTAmount is the amount of MPT to clawback from the holder's confidential balance.
	// Must be greater than 0.
	MPTAmount types.MPTPlainAmount
	// ZKProof proves that MPTAmount equals the holder balance encrypted for the issuer.
	ZKProof string
}

// TxType returns the transaction type (ConfidentialMPTClawback).
func (*ConfidentialMPTClawback) TxType() TxType {
	return ConfidentialMPTClawbackTx
}

// Flatten returns the flattened map of the ConfidentialMPTClawback transaction.
func (tx *ConfidentialMPTClawback) Flatten() FlatTransaction {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	flattened["Holder"] = tx.Holder.String()

	flattened["MPTokenIssuanceID"] = tx.MPTokenIssuanceID

	flattened["MPTAmount"] = tx.MPTAmount.Flatten()

	flattened["ZKProof"] = tx.ZKProof

	return flattened
}

// Validate validates the ConfidentialMPTClawback transaction.
func (tx *ConfidentialMPTClawback) Validate() (bool, error) {
	ok, err := tx.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}
	accountID, err := validateConfidentialMPTBase(&tx.BaseTx)
	if err != nil {
		return false, err
	}
	if err := validateConfidentialMPTIssuer(tx.MPTokenIssuanceID, accountID); err != nil {
		return false, err
	}

	_, sameAccount, holderHasTag, err := decodeCounterparty(accountID, tx.Holder)
	if err != nil {
		return false, ErrConfidentialClawbackInvalidHolder
	}
	if sameAccount {
		return false, ErrConfidentialClawbackSelfClawback
	}
	// ConfidentialMPTClawback has no HolderTag field, so an embedded tag has nowhere to go.
	if holderHasTag {
		return false, ErrConfidentialClawbackHolderTagNotAllowed
	}

	if tx.MPTAmount.IsZero() || !tx.MPTAmount.IsValid() {
		return false, ErrConfidentialClawbackInvalidAmount
	}

	if !IsValidClawbackProof(tx.ZKProof) {
		return false, ErrConfidentialClawbackBadProof
	}

	return true, nil
}
