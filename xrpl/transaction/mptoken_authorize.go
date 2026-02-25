package transaction

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// TfMPTUnauthorize if set and transaction is submitted by a holder, it indicates that the holder no
	// longer wants to hold the MPToken, which will be deleted as a result. If the the holder's
	// MPToken has non-zero balance while trying to set this flag, the transaction will fail. On
	// the other hand, if set and transaction is submitted by an issuer, it would mean that the
	// issuer wants to unauthorize the holder (only applicable for allow-listing),
	// which would unset the LsfMPTAuthorized flag on the MPToken.
	TfMPTUnauthorize uint32 = 1
)

// MPTokenAuthorize transaction is used to globally lock/unlock a MPTokenIssuance,
// or lock/unlock an individual's MPToken.
type MPTokenAuthorize struct {
	BaseTx
	// Indicates the ID of the MPT involved.
	MPTokenIssuanceID string
	// (Optional) Specifies the holder's address that the issuer wants to authorize.
	// Only used for authorization/allow-listing; must be empty if submitted by the holder.
	Holder *types.Address `json:",omitempty"`
}

// SetMPTUnauthorizeFlag toggles the unauthorize flag on the transaction to indicate that the holder no longer wants to hold the token or that the issuer wants to unauthorize a holder.
func (m *MPTokenAuthorize) SetMPTUnauthorizeFlag() {
	m.Flags |= TfMPTUnauthorize
}

// TxType returns the type of the transaction (MPTokenAuthorize).
func (*MPTokenAuthorize) TxType() TxType {
	return MPTokenAuthorizeTx
}

// Flatten returns the flattened map of the MPTokenAuthorize transaction.
func (m *MPTokenAuthorize) Flatten() FlatTransaction {
	// Add BaseTx fields
	flattened := m.BaseTx.Flatten()

	flattened["TransactionType"] = "MPTokenAuthorize"

	flattened["Account"] = m.Account.String()

	flattened["MPTokenIssuanceID"] = m.MPTokenIssuanceID

	if m.Holder != nil {
		flattened["Holder"] = m.Holder.String()
	}

	return flattened
}

// Validate validates the MPTokenAuthorize transaction ensuring all fields are correct.
func (m *MPTokenAuthorize) Validate() (bool, error) {
	ok, err := m.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	// check owner is a valid xrpl address
	if m.Account != "" && !addresscodec.IsValidAddress(m.Account.String()) {
		return false, ErrInvalidAccount
	}

	// If a Holder is specified, validate it as a proper XRPL address.
	if m.Holder != nil && !addresscodec.IsValidAddress(m.Holder.String()) {
		return false, ErrInvalidAccount
	}

	// check account is not the same as the holder
	if m.Holder != nil && m.Account.String() == m.Holder.String() {
		return false, ErrHolderAccountConflict
	}

	return true, nil
}
