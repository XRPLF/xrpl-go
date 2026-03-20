package transaction

// DIDSet creates or updates a DID ledger entry, setting Data, DIDDocument, or URI. (Requires the DID amendment)
//
// Example:
// ```json
//
//	{
//	    "TransactionType": "DIDSet",
//	    "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
//	    "Fee": "10",
//	    "Sequence": 391,
//	    "URI": "697...",
//	    "Data": "...",
//	    "SigningPubKey": "0330..."
//	}
//
// ```
type DIDSet struct {
	BaseTx
	// The public attestations of identity credentials associated with the DID.
	Data string `json:",omitempty"`
	// The DID document associated with the DID.
	DIDDocument string `json:",omitempty"`
	// The URI associated with the DID.
	URI string `json:",omitempty"`
}

// TxType returns the type of the transaction.
func (tx *DIDSet) TxType() TxType {
	return DIDSetTx
}

// Flatten returns a flattened version of the transaction.
func (tx *DIDSet) Flatten() FlatTransaction {
	flattened := tx.BaseTx.Flatten()
	flattened["TransactionType"] = tx.TxType().String()

	if tx.Data != "" {
		flattened["Data"] = tx.Data
	}

	if tx.DIDDocument != "" {
		flattened["DIDDocument"] = tx.DIDDocument
	}

	if tx.URI != "" {
		flattened["URI"] = tx.URI
	}

	return flattened
}

// Validate checks DIDSet transaction fields and returns false with an error if invalid.
func (tx *DIDSet) Validate() (bool, error) {
	if ok, err := tx.BaseTx.Validate(); !ok {
		return false, err
	}

	if tx.Data == "" && tx.DIDDocument == "" && tx.URI == "" {
		return false, ErrDIDSetMustSetEitherDataOrDIDDocumentOrURI
	}

	return true, nil
}
