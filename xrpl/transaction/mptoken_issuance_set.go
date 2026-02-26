package transaction

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/flag"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// MPTokenIssuanceSet Flags
const (
	// TfMPTLock if set, indicates that all MPT balances for this asset should be locked.
	TfMPTLock uint32 = 0x00000001
	// TfMPTUnlock if set, indicates that all MPT balances for this asset should be unlocked.
	TfMPTUnlock uint32 = 0x00000002
)

// MPTokenIssuanceSet transaction is used to globally lock/unlock a MPTokenIssuance,
// or lock/unlock an individual's MPToken.
//
// ```json
//
//	{
//	      "TransactionType": "MPTokenIssuanceSet",
//	      "Fee": "10",
//	      "MPTokenIssuanceID": "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
//	      "Flags": 1
//	}
//
// ```
type MPTokenIssuanceSet struct {
	BaseTx
	// The MPTokenIssuance identifier.
	MPTokenIssuanceID string
	// (Optional) XRPL Address of an individual token holder balance to lock/unlock. If omitted, this transaction applies to all any accounts holding MPTs.
	Holder *types.Address
}

// TxType returns the type of the transaction (MPTokenIssuanceSet).
func (*MPTokenIssuanceSet) TxType() TxType {
	return MPTokenIssuanceSetTx
}

// Flatten returns the flattened map of the MPTokenIssuanceSet transaction.
func (m *MPTokenIssuanceSet) Flatten() FlatTransaction {
	flattened := m.BaseTx.Flatten()

	flattened["TransactionType"] = "MPTokenIssuanceSet"

	flattened["MPTokenIssuanceID"] = m.MPTokenIssuanceID

	if m.Holder != nil {
		flattened["Holder"] = m.Holder.String()
	}

	return flattened
}

// SetMPTLockFlag sets the TfMPTLock flag on the transaction.
// Indicates that all MPT balances for this asset should be locked.
func (m *MPTokenIssuanceSet) SetMPTLockFlag() {
	m.Flags |= TfMPTLock
}

// SetMPTUnlockFlag sets the TfMPTUnlock flag on the transaction.
// Indicates that all MPT balances for this asset should be unlocked.
func (m *MPTokenIssuanceSet) SetMPTUnlockFlag() {
	m.Flags |= TfMPTUnlock
}

// Validate validates the MPTokenIssuanceSet transaction ensuring all fields are correct.
func (m *MPTokenIssuanceSet) Validate() (bool, error) {
	ok, err := m.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	// If a Holder is specified, validate it as a proper XRPL address.
	if m.Holder != nil && !addresscodec.IsValidAddress(m.Holder.String()) {
		return false, ErrInvalidAccount
	}

	// Check flag conflict: TfMPTLock and TfMPTUnlock cannot both be enabled
	isLock := flag.Contains(m.Flags, TfMPTLock)
	isUnlock := flag.Contains(m.Flags, TfMPTUnlock)

	if isLock && isUnlock {
		return false, ErrMPTokenIssuanceSetFlags
	}

	return true, nil
}
