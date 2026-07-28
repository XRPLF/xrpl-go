package transaction

// ConfidentialMPTMergeInbox merges the holder's confidential inbox balance (CB_IN)
// into their main confidential spending balance (CB_S).
//
// When confidential MPT is sent to a holder, it accumulates in their
// "inbox" balance. This transaction allows the holder to merge those
// incoming funds into their main "spending" balance so they can use them.
//
// This transaction is permissionless and requires no cryptographic proof because
// the holder is simply consolidating their own balances.
//
// ```json
//
//	{
//	    "TransactionType": "ConfidentialMPTMergeInbox",
//	    "Fee": "10",
//	    "MPTokenIssuanceID": "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000"
//	}
//
// ```
type ConfidentialMPTMergeInbox struct {
	BaseTx
	// MPTokenIssuanceID identifies the MPTokenIssuance for which to merge inbox balance.
	MPTokenIssuanceID string
}

// TxType returns the transaction type (ConfidentialMPTMergeInbox).
func (*ConfidentialMPTMergeInbox) TxType() TxType {
	return ConfidentialMPTMergeInboxTx
}

// Flatten returns the flattened map of the ConfidentialMPTMergeInbox transaction.
func (tx *ConfidentialMPTMergeInbox) Flatten() FlatTransaction {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	flattened["MPTokenIssuanceID"] = tx.MPTokenIssuanceID

	return flattened
}

// Validate validates the ConfidentialMPTMergeInbox transaction.
func (tx *ConfidentialMPTMergeInbox) Validate() (bool, error) {
	ok, err := tx.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	if tx.MPTokenIssuanceID == "" {
		return false, ErrConfidentialMPTInvalidIssuanceID
	}

	return true, nil
}
