package transaction

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ConfidentialMPTClawback allows the issuer to reclaim a holder's entire confidential
// MPT balance. Unlike regular clawback, the issuer must provide a ZK equality proof
// demonstrating knowledge of the encrypted balance since balances are not visible.
// Can only be submitted by the issuer, and only if TfMPTCanClawback was enabled at issuance.
//
// ```json
//
//	{
//	    "TransactionType": "ConfidentialMPTClawback",
//	    "Fee": "10",
//	    "Holder": "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
//	    "MPTokenIssuanceID": "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
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
	// ZKProof is a zero-knowledge proof proving the holder has sufficient confidential
	// balance for the clawback and that the operation is valid.
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

	if tx.MPTokenIssuanceID == "" {
		return false, ErrConfidentialMPTInvalidIssuanceID
	}

	if !addresscodec.IsValidAddress(tx.Holder.String()) {
		return false, ErrConfidentialClawbackInvalidHolder
	}

	if tx.Holder.String() == tx.Account.String() {
		return false, ErrConfidentialClawbackSelfClawback
	}

	if tx.MPTAmount == 0 {
		return false, ErrConfidentialClawbackInvalidAmount
	}

	if !IsValidFixedHexBlob(tx.ZKProof, ClawbackProofLen) {
		return false, ErrConfidentialClawbackBadProof
	}

	return true, nil
}
