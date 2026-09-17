package transaction

import (
	"fmt"
	"math/bits"

	"github.com/Peersyst/xrpl-go/xrpl/flag"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// TfSponsorshipEnd returns reserve responsibility to the sponsee.
	TfSponsorshipEnd uint32 = 0x00010000
	// TfSponsorshipCreate moves an unsponsored target's reserve to a sponsor.
	TfSponsorshipCreate uint32 = 0x00020000
	// TfSponsorshipReassign moves an existing reserve sponsorship to a new sponsor.
	TfSponsorshipReassign uint32 = 0x00040000
)

// SponsorshipTransfer creates, reassigns, or ends reserve sponsorship for an
// account or ledger object. Requires the Sponsor amendment on the network.
type SponsorshipTransfer struct {
	BaseTx
	// ObjectID selects an object's reserve. Nil selects account reserve sponsorship.
	ObjectID *types.Hash256 `json:",omitempty"`
	// Sponsee identifies the target owner when the sponsor ends sponsorship.
	// Empty defaults to Account. Create and reassign must omit this field.
	Sponsee types.Address `json:",omitempty"`
}

// TxType returns the SponsorshipTransfer transaction type.
func (*SponsorshipTransfer) TxType() TxType {
	return SponsorshipTransferTx
}

// Flatten returns the protocol representation of SponsorshipTransfer.
func (tx *SponsorshipTransfer) Flatten() FlatTransaction {
	flattened := tx.BaseTx.Flatten()
	flattened["TransactionType"] = tx.TxType().String()
	if tx.ObjectID != nil {
		flattened["ObjectID"] = tx.ObjectID.String()
	}
	if tx.Sponsee != "" {
		flattened["Sponsee"] = tx.Sponsee.String()
	}
	return flattened
}

// Validate checks the selected operation and its local field requirements.
// Target ownership, existing sponsorship, and reserve sufficiency require ledger state.
func (tx *SponsorshipTransfer) Validate() (bool, error) {
	if tx.TransactionType != tx.TxType() {
		return false, ErrInvalidTransactionType
	}
	if ok, err := tx.BaseTx.Validate(); !ok {
		return false, err
	}
	const scenarios = TfSponsorshipEnd | TfSponsorshipCreate | TfSponsorshipReassign
	if bits.OnesCount32(tx.Flags&scenarios) != 1 || !flag.ContainsOnly(tx.Flags, scenarios|types.TfUniversal) {
		return false, ErrInvalidFlags
	}
	if tx.ObjectID != nil && !IsLedgerEntryID(tx.ObjectID.String()) {
		return false, ErrSponsorshipTransferObjectID
	}
	if flag.Contains(tx.Flags, TfSponsorshipEnd) {
		if tx.Sponsor != "" {
			return false, ErrSponsorshipTransferSponsorNotAllowed
		}
		if tx.Sponsee != "" {
			if err := validateSponsorshipCounterparty(tx.Account, tx.Sponsee); err != nil {
				return false, fmt.Errorf("%w: %w", ErrInvalidSponsee, err)
			}
		}
		return true, nil
	}
	if tx.Sponsor == "" {
		return false, ErrSponsorshipTransferSponsorRequired
	}
	if !flag.Contains(tx.SponsorFlags, types.SpfSponsorReserve) {
		return false, ErrSponsorshipTransferReserveRequired
	}
	if tx.Sponsee != "" {
		return false, ErrSponsorshipTransferSponseeNotAllowed
	}
	// Account-level sponsorship requires explicit authorization. Object-level
	// sponsorship can instead draw from a prefunded reserve budget. BaseTx owns
	// signature shape, including unsigned placeholders in inner Batch transactions.
	if tx.ObjectID == nil && tx.SponsorSignature == nil {
		return false, ErrSponsorshipTransferSignatureRequired
	}
	return true, nil
}

// SetSponsorshipEndFlag enables TfSponsorshipEnd.
func (tx *SponsorshipTransfer) SetSponsorshipEndFlag() {
	tx.Flags |= TfSponsorshipEnd
}

// SetSponsorshipCreateFlag enables TfSponsorshipCreate.
func (tx *SponsorshipTransfer) SetSponsorshipCreateFlag() {
	tx.Flags |= TfSponsorshipCreate
}

// SetSponsorshipReassignFlag enables TfSponsorshipReassign.
func (tx *SponsorshipTransfer) SetSponsorshipReassignFlag() {
	tx.Flags |= TfSponsorshipReassign
}
