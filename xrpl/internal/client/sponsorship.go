package client

import (
	"encoding/json"
	"fmt"

	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/currency"
	"github.com/Peersyst/xrpl-go/xrpl/flag"
	ledgerentry "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	ledgerquery "github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// SponsorshipValidation reports the outcome of an online sponsorship preflight.
type SponsorshipValidation struct {
	// Valid reports whether the preflight found no blocking sponsorship problem.
	Valid bool
	// Reason explains a rejection.
	Reason error
	// Sponsorship is the Sponsorship ledger entry the preflight checked against.
	Sponsorship *ledgerentry.Sponsorship
	// Fee is the transaction fee, in drops, that the preflight validated.
	Fee currency.Drops
}

// FetchSponsorshipEntry performs the client-specific ledger_entry lookup for the Sponsorship entry
// between a sponsor and a sponsee.
type FetchSponsorshipEntry func(sponsor, sponsee types.Address) (*ledgerentry.Sponsorship, error)

// SponsorshipEntryRequest builds the ledger_entry request for the Sponsorship entry between a
// sponsor and a sponsee.
func SponsorshipEntryRequest(sponsor, sponsee types.Address) *ledgerquery.EntryRequest {
	return &ledgerquery.EntryRequest{
		Sponsorship: ledgerquery.SponsorshipSelector{
			Object: &ledgerquery.SponsorshipSelectorFields{
				Sponsor: sponsor,
				Sponsee: sponsee,
			},
		},
		LedgerIndex: common.Current,
	}
}

// DecodeSponsorshipEntry converts a ledger_entry node into a typed Sponsorship entry.
func DecodeSponsorshipEntry(node ledgerentry.FlatLedgerObject) (*ledgerentry.Sponsorship, error) {
	entryType, _ := typecheck.ToString(node["LedgerEntryType"])
	if ledgerentry.EntryType(entryType) != ledgerentry.SponsorshipEntry {
		return nil, fmt.Errorf("%w: got %q", ErrSponsorshipEntryUnexpectedType, entryType)
	}

	encoded, err := json.Marshal(node)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSponsorshipEntryMalformed, err)
	}
	var entry ledgerentry.Sponsorship
	if err := json.Unmarshal(encoded, &entry); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSponsorshipEntryMalformed, err)
	}
	return &entry, nil
}

// ValidateSponsorship runs an online preflight of a sponsored transaction against the Sponsorship
// ledger entry between its sponsor and sponsee.
func ValidateSponsorship(
	tx map[string]any,
	estimatedFee string,
	fetchSponsorship FetchSponsorshipEntry,
) (SponsorshipValidation, error) {
	sponsor, sponsorFlags, err := sponsorshipFields(tx)
	if err != nil {
		return SponsorshipValidation{}, err
	}

	fee, err := sponsorshipFee(tx, estimatedFee)
	if err != nil {
		return SponsorshipValidation{}, err
	}

	sponsee, delegated, err := sponsorshipSponsee(tx)
	if err != nil {
		return SponsorshipValidation{}, err
	}

	if account, _ := typecheck.ToString(tx["Account"]); account == string(sponsor) {
		return SponsorshipValidation{Reason: ErrSponsorIsAccount, Fee: fee}, nil
	}
	// rippled rejects reserve sponsorship on a delegated transaction outright.
	if delegated && flag.Contains(sponsorFlags, types.SpfSponsorReserve) {
		return SponsorshipValidation{Reason: ErrDelegatedReserveSponsorship, Fee: fee}, nil
	}

	entry, err := fetchSponsorship(sponsor, sponsee)
	if err != nil {
		return SponsorshipValidation{}, err
	}

	coSigned := hasSponsorSignature(tx)
	if entry == nil {
		// rippled requires a Sponsorship entry only for pre-funded sponsorship.
		if !coSigned {
			return SponsorshipValidation{
				Reason: fmt.Errorf("%w: sponsor %s, sponsee %s", ErrSponsorshipEntryNotFound, sponsor, sponsee),
				Fee:    fee,
			}, nil
		}
		return SponsorshipValidation{Valid: true, Fee: fee}, nil
	}

	result := SponsorshipValidation{Sponsorship: entry, Fee: fee}
	result.Reason = rejectSponsorship(entry, sponsorFlags, coSigned, fee)
	result.Valid = result.Reason == nil
	return result, nil
}

// rejectSponsorship applies the require-signature flags and the budget of an existing Sponsorship
// entry, returning nil when the entry covers the transaction.
func rejectSponsorship(
	entry *ledgerentry.Sponsorship,
	sponsorFlags uint32,
	coSigned bool,
	fee currency.Drops,
) error {
	sponsorsFee := flag.Contains(sponsorFlags, types.SpfSponsorFee)
	sponsorsReserve := flag.Contains(sponsorFlags, types.SpfSponsorReserve)

	if !coSigned {
		if sponsorsFee && flag.Contains(entry.Flags, ledgerentry.LsfSponsorshipRequireSignForFee) {
			return ErrSponsorshipFeeSignatureRequired
		}
		if sponsorsReserve && flag.Contains(entry.Flags, ledgerentry.LsfSponsorshipRequireSignForReserve) {
			return ErrSponsorshipReserveSignatureRequired
		}
	}

	// rippled returns early before it resolves a fee payer when the transaction pays no fee, so a
	// zero fee draws nothing from the entry.
	if !sponsorsFee || fee.Cmp(currency.DropsFromUint64(0)) == 0 {
		return nil
	}

	if entry.FeeAmount == nil {
		return ErrSponsorshipFeeAmountMissing
	}
	feeAmount := currency.DropsFromUint64(entry.FeeAmount.Uint64())
	if feeAmount.Cmp(fee) < 0 {
		return fmt.Errorf("%w: FeeAmount %s drops, fee %s drops", ErrSponsorshipFeeBudgetExhausted, entry.FeeAmount, dropsText(fee))
	}

	if entry.MaxFee != nil {
		maxFee := currency.DropsFromUint64(entry.MaxFee.Uint64())
		if fee.Cmp(maxFee) > 0 {
			return fmt.Errorf("%w: MaxFee %s drops, fee %s drops", ErrSponsorshipMaxFeeExceeded, entry.MaxFee, dropsText(fee))
		}
	}

	return nil
}

// sponsorshipFields reads the sponsored-transaction common fields that identify the sponsor and
// the requested sponsorship types.
func sponsorshipFields(tx map[string]any) (types.Address, uint32, error) {
	sponsorValue, hasSponsor := tx["Sponsor"]
	flagsValue, hasFlags := tx["SponsorFlags"]
	if !hasSponsor && !hasFlags {
		return "", 0, ErrTransactionNotSponsored
	}
	if hasSponsor != hasFlags {
		return "", 0, ErrInvalidSponsorFlags
	}

	sponsor, ok := typecheck.ToString(sponsorValue)
	if !ok {
		return "", 0, fmt.Errorf("%w: got %T", ErrSponsorFieldIsNotAString, sponsorValue)
	}
	sponsorFlags, ok := typecheck.ToUint32(flagsValue)
	if !ok {
		return "", 0, fmt.Errorf("%w: got %T", ErrSponsorFlagsFieldIsNotAUint32, flagsValue)
	}
	if sponsor == "" {
		return "", 0, ErrTransactionNotSponsored
	}
	if sponsorFlags == 0 || sponsorFlags&^types.SpfSponsorUniversal != 0 {
		return "", 0, fmt.Errorf("%w: got %#x", ErrInvalidSponsorFlags, sponsorFlags)
	}

	return types.Address(sponsor), sponsorFlags, nil
}

// sponsorshipSponsee resolves the account whose sponsorship is being used, and reports whether a
// Delegate supplied it.
func sponsorshipSponsee(tx map[string]any) (types.Address, bool, error) {
	for _, field := range []string{"Delegate", "Account"} {
		value, exists := tx[field]
		if !exists || value == nil {
			continue
		}
		account, ok := typecheck.ToString(value)
		if !ok {
			return "", false, fmt.Errorf("%w: field %s is a %T", ErrAddressFieldIsNotAString, field, value)
		}
		if account != "" {
			return types.Address(account), field == "Delegate", nil
		}
	}
	return "", false, ErrSponsorshipSponseeUnavailable
}

// hasSponsorSignature reports whether the transaction carries a sponsor co-signature.
func hasSponsorSignature(tx map[string]any) bool {
	signature, ok := tx["SponsorSignature"].(map[string]any)
	return ok && len(signature) > 0
}

// sponsorshipFee prefers an explicitly supplied estimate over the transaction's own Fee, so a
// preflight can run before autofill has set one.
func sponsorshipFee(tx map[string]any, estimatedFee string) (currency.Drops, error) {
	feeText := estimatedFee
	if feeText == "" {
		value, exists := tx["Fee"]
		if !exists || value == nil {
			return currency.Drops{}, ErrSponsorshipFeeUnavailable
		}
		text, ok := typecheck.ToString(value)
		if !ok {
			return currency.Drops{}, fmt.Errorf("%w: got %T", ErrSponsorshipFeeIsNotAString, value)
		}
		if text == "" {
			return currency.Drops{}, ErrSponsorshipFeeUnavailable
		}
		feeText = text
	}

	fee, err := currency.DropsFromString(feeText)
	if err != nil {
		return currency.Drops{}, fmt.Errorf("%w: %q: %w", ErrInvalidSponsorshipFee, feeText, err)
	}
	return fee, nil
}

func dropsText(drops currency.Drops) string {
	text, err := drops.WholeString()
	if err != nil {
		return "unknown"
	}
	return text
}
