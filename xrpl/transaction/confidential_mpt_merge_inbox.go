package transaction

// ConfidentialMPTMergeInbox requires the ConfidentialTransfer amendment.
// It merges the holder's confidential inbox balance (CB_IN)
// into their main confidential spending balance (CB_S).
//
// When confidential MPT is sent to a holder, it accumulates in their
// "inbox" balance. This transaction allows the holder to merge those
// incoming funds into their main "spending" balance so they can use them.
//
// This transaction requires holder authorization but no cryptographic proof because
// the holder is consolidating their own balances.
//
// ```json
//
//	{
//	    "TransactionType": "ConfidentialMPTMergeInbox",
//	    "Account": "r...",
//	    "MPTokenIssuanceID": "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8"
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
	accountID, err := validateConfidentialMPTBase(&tx.BaseTx)
	if err != nil {
		return false, err
	}
	if _, err := validateConfidentialMPTHolder(tx.MPTokenIssuanceID, accountID); err != nil {
		return false, err
	}

	return true, nil
}
