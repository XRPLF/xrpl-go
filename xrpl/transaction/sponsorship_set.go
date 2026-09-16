package transaction

import (
	"fmt"
	"strconv"

	"github.com/Peersyst/xrpl-go/xrpl/currency"
	"github.com/Peersyst/xrpl-go/xrpl/flag"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// TfSponsorshipSetRequireSignForFee requires the sponsor to co-sign fee sponsorship.
	TfSponsorshipSetRequireSignForFee uint32 = 0x00010000
	// TfSponsorshipClearRequireSignForFee permits prefunded fees without a sponsor signature.
	TfSponsorshipClearRequireSignForFee uint32 = 0x00020000
	// TfSponsorshipSetRequireSignForReserve requires the sponsor to co-sign reserve sponsorship.
	TfSponsorshipSetRequireSignForReserve uint32 = 0x00040000
	// TfSponsorshipClearRequireSignForReserve permits prefunded reserves without a sponsor signature.
	TfSponsorshipClearRequireSignForReserve uint32 = 0x00080000
	// TfDeleteObject deletes the Sponsorship object without ending existing reserve sponsorships.
	TfDeleteObject uint32 = 0x00100000

	sponsorshipSetModifyFlags = TfSponsorshipSetRequireSignForFee | TfSponsorshipClearRequireSignForFee |
		TfSponsorshipSetRequireSignForReserve | TfSponsorshipClearRequireSignForReserve
)

// SponsorshipSet creates, updates, or deletes a prefunded Sponsorship object.
// Only the sponsor can create or update it. Either party can delete it.
// Requires the Sponsor amendment on the network.
type SponsorshipSet struct {
	BaseTx
	// Sponsee identifies the other account when Account is the sponsorship owner.
	Sponsee types.Address `json:",omitempty"`
	// CounterpartySponsor identifies the owner when the sponsee deletes the object.
	// Exactly one of Sponsee and CounterpartySponsor must be supplied.
	CounterpartySponsor types.Address `json:",omitempty"`
	// FeeAmountDelta adds or removes fee budget, in canonical signed integer drops.
	// It must be nonzero when present. Nil leaves the existing budget unchanged.
	FeeAmountDelta *string `json:",omitempty"`
	// MaxFee sets the maximum sponsored fee per transaction when positive.
	// Zero removes the per-transaction limit. Nil leaves the existing limit unchanged.
	MaxFee *types.XRPCurrencyAmount `json:",omitempty"`
	// RemainingOwnerCountDelta adds or removes reserve units. It must be nonzero
	// when present. Nil leaves the existing reserve budget unchanged.
	RemainingOwnerCountDelta *int32 `json:",omitempty"`
}

// TxType returns the SponsorshipSet transaction type.
func (*SponsorshipSet) TxType() TxType {
	return SponsorshipSetTx
}

// Flatten returns the protocol representation of SponsorshipSet.
func (tx *SponsorshipSet) Flatten() FlatTransaction {
	flattened := tx.BaseTx.Flatten()
	flattened["TransactionType"] = tx.TxType().String()
	if tx.Sponsee != "" {
		flattened["Sponsee"] = tx.Sponsee.String()
	}
	if tx.CounterpartySponsor != "" {
		flattened["CounterpartySponsor"] = tx.CounterpartySponsor.String()
	}
	if tx.FeeAmountDelta != nil {
		flattened["FeeAmountDelta"] = *tx.FeeAmountDelta
	}
	if tx.MaxFee != nil {
		flattened["MaxFee"] = tx.MaxFee.String()
	}
	if tx.RemainingOwnerCountDelta != nil {
		flattened["RemainingOwnerCountDelta"] = *tx.RemainingOwnerCountDelta
	}
	return flattened
}

// Validate checks local field and operation rules. Budget sufficiency, account
// existence, and whether a Sponsorship object already exists require ledger state.
func (tx *SponsorshipSet) Validate() (bool, error) {
	if tx.TransactionType != tx.TxType() {
		return false, ErrInvalidTransactionType
	}
	if ok, err := tx.BaseTx.Validate(); !ok {
		return false, err
	}
	if (tx.Sponsee == "") == (tx.CounterpartySponsor == "") {
		return false, ErrSponsorshipSetCounterpartyConflict
	}
	if tx.Sponsee != "" {
		if err := validateSponsorshipCounterparty(tx.Account, tx.Sponsee); err != nil {
			return false, fmt.Errorf("%w: %w", ErrInvalidSponsee, err)
		}
	} else if err := validateSponsorshipCounterparty(tx.Account, tx.CounterpartySponsor); err != nil {
		return false, fmt.Errorf("%w: %w", ErrInvalidCounterpartySponsor, err)
	}
	if !flag.ContainsOnly(tx.Flags, sponsorshipSetModifyFlags|TfDeleteObject|types.TfUniversal) ||
		flag.Contains(tx.Flags, TfSponsorshipSetRequireSignForFee|TfSponsorshipClearRequireSignForFee) ||
		flag.Contains(tx.Flags, TfSponsorshipSetRequireSignForReserve|TfSponsorshipClearRequireSignForReserve) {
		return false, ErrInvalidFlags
	}

	if flag.Contains(tx.Flags, TfDeleteObject) {
		if flag.ContainsAny(tx.Flags, sponsorshipSetModifyFlags) {
			return false, ErrInvalidFlags
		}
		if tx.FeeAmountDelta != nil || tx.MaxFee != nil || tx.RemainingOwnerCountDelta != nil {
			return false, ErrSponsorshipSetDeleteConflict
		}
		return true, nil
	}
	if tx.CounterpartySponsor != "" {
		return false, ErrSponsorshipSetCounterpartyCannotModify
	}
	if tx.FeeAmountDelta == nil && tx.MaxFee == nil && tx.RemainingOwnerCountDelta == nil &&
		!flag.ContainsAny(tx.Flags, sponsorshipSetModifyFlags) {
		return false, ErrSponsorshipSetEmptyUpdate
	}
	if tx.FeeAmountDelta != nil {
		// Canonical spelling follows xrpl.js. The magnitude bound matches the
		// native amount codec, so a successful validation can be encoded exactly.
		delta, err := strconv.ParseInt(*tx.FeeAmountDelta, 10, 64)
		if err != nil || delta == 0 || strconv.FormatInt(delta, 10) != *tx.FeeAmountDelta ||
			delta > int64(currency.MaxNativeDrops) || delta < -int64(currency.MaxNativeDrops) {
			return false, ErrSponsorshipSetFeeAmountDelta
		}
	}
	if tx.MaxFee != nil && tx.MaxFee.Uint64() > currency.MaxNativeDrops {
		return false, ErrSponsorshipSetMaxFee
	}
	if tx.RemainingOwnerCountDelta != nil && *tx.RemainingOwnerCountDelta == 0 {
		return false, ErrSponsorshipSetRemainingOwnerCountDelta
	}
	return true, nil
}

// validateSponsorshipCounterparty checks the shared counterparty rules for
// SponsorshipSet and SponsorshipTransfer after their BaseTx validation.
func validateSponsorshipCounterparty(account, counterparty types.Address) error {
	accountID, _, err := decodeAddressAccountID(account)
	if err != nil {
		return ErrInvalidAccount
	}
	_, same, hasTag, err := decodeCounterparty(accountID, counterparty)
	if err != nil {
		return err
	}
	if hasTag {
		return ErrAccountIDTagNotAllowed
	}
	if same {
		return ErrSponsorshipAccountConflict
	}
	return nil
}

// SetSponsorshipSetRequireSignForFeeFlag enables TfSponsorshipSetRequireSignForFee.
func (tx *SponsorshipSet) SetSponsorshipSetRequireSignForFeeFlag() {
	tx.Flags |= TfSponsorshipSetRequireSignForFee
}

// SetSponsorshipClearRequireSignForFeeFlag enables TfSponsorshipClearRequireSignForFee.
func (tx *SponsorshipSet) SetSponsorshipClearRequireSignForFeeFlag() {
	tx.Flags |= TfSponsorshipClearRequireSignForFee
}

// SetSponsorshipSetRequireSignForReserveFlag enables TfSponsorshipSetRequireSignForReserve.
func (tx *SponsorshipSet) SetSponsorshipSetRequireSignForReserveFlag() {
	tx.Flags |= TfSponsorshipSetRequireSignForReserve
}

// SetSponsorshipClearRequireSignForReserveFlag enables TfSponsorshipClearRequireSignForReserve.
func (tx *SponsorshipSet) SetSponsorshipClearRequireSignForReserveFlag() {
	tx.Flags |= TfSponsorshipClearRequireSignForReserve
}

// SetDeleteObjectFlag enables TfDeleteObject.
func (tx *SponsorshipSet) SetDeleteObjectFlag() {
	tx.Flags |= TfDeleteObject
}
