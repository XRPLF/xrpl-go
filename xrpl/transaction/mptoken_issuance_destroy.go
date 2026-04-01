package transaction

import "github.com/Peersyst/xrpl-go/pkg/typecheck"

// MPTokenIssuanceDestroy transaction is used to remove an MPTokenIssuance object from the directory node
// in which it is being held, effectively removing the token from the ledger ("destroying" it).
//
// If this operation succeeds, the corresponding MPTokenIssuance is removed and the owner’s reserve requirement is reduced by one.
// This operation must fail if there are any holders of the MPT in question.
//
// ```json
//
//	 {
//	     "TransactionType": "MPTokenIssuanceDestroy",
//	     "Fee": "10",
//	     "MPTokenIssuanceID": "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000"
//	}
//
// ```
type MPTokenIssuanceDestroy struct {
	BaseTx
	// Identifies the MPTokenIssuance object to be removed by the transaction.
	MPTokenIssuanceID string
}

// TxType returns the type of the transaction (MPTokenIssuanceDestroy).
func (*MPTokenIssuanceDestroy) TxType() TxType {
	return MPTokenIssuanceDestroyTx
}

// Flatten returns the flattened map of the MPTokenIssuanceDestroy transaction.
func (m *MPTokenIssuanceDestroy) Flatten() FlatTransaction {
	flattened := m.BaseTx.Flatten()

	flattened["TransactionType"] = "MPTokenIssuanceDestroy"

	flattened["MPTokenIssuanceID"] = m.MPTokenIssuanceID

	return flattened
}

// Validate validates the MPTokenIssuanceDestroy transaction ensuring all fields are correct.
func (m *MPTokenIssuanceDestroy) Validate() (bool, error) {
	ok, err := m.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	if m.MPTokenIssuanceID == "" || !typecheck.IsHex(m.MPTokenIssuanceID) {
		return false, ErrInvalidMPTokenIssuanceIDDestroy
	}

	return true, nil
}
