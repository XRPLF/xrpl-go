package client

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"

	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/currency"
	"github.com/Peersyst/xrpl-go/xrpl/flag"
	ledgerentry "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	ledgerquery "github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
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

// fetchSponsorshipEntry looks up and decodes the Sponsorship entry for a sponsor and sponsee.
// Only an error recognized by the transport's isNotFound classifier means the entry is absent.
func fetchSponsorshipEntry(
	ctx context.Context,
	request RequestResultFunc,
	isNotFound func(error) bool,
	sponsor, sponsee types.Address,
) (*ledgerentry.Sponsorship, error) {
	req := &ledgerquery.EntryRequest{
		Sponsorship: ledgerquery.SponsorshipSelector{
			Object: &ledgerquery.SponsorshipSelectorFields{
				Sponsor: sponsor,
				Sponsee: sponsee,
			},
		},
		// The transaction applies against the open ledger, so the current ledger is the one whose
		// budget and flags decide it, including changes still pending validation. It is also the
		// ledger rippled's ledger_entry defaults to.
		LedgerIndex: common.Current,
	}
	var response ledgerquery.EntryResponse
	if err := request(ctx, req, &response); err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return decodeSponsorshipEntry(response.Node)
}

// decodeSponsorshipEntry converts a ledger_entry node into a typed Sponsorship entry.
func decodeSponsorshipEntry(node ledgerentry.FlatLedgerObject) (*ledgerentry.Sponsorship, error) {
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
// ledger entry between its sponsor and sponsee. The sponsorship fields are first checked by the
// same rules BaseTx.Validate applies, so this helper only adds the checks that need the ledger.
// request and isNotFound adapt the client's transport and entryNotFound error to the shared lookup.
func ValidateSponsorship(
	ctx context.Context,
	request RequestResultFunc,
	isNotFound func(error) bool,
	tx map[string]any,
	estimatedFee string,
) (SponsorshipValidation, error) {
	_, hasSponsor := tx["Sponsor"]
	_, hasFlags := tx["SponsorFlags"]
	_, coSigned := tx["SponsorSignature"]
	if !hasSponsor && !hasFlags && !coSigned {
		return SponsorshipValidation{}, ErrTransactionNotSponsored
	}
	if _, err := transaction.InspectSponsorFields(tx); err != nil {
		return SponsorshipValidation{}, err
	}

	fee, err := sponsorshipFee(tx, estimatedFee)
	if err != nil {
		return SponsorshipValidation{}, err
	}

	sponsor, sponsee, err := sponsorshipParties(tx)
	if err != nil {
		return SponsorshipValidation{}, err
	}
	// The shared rules above already accepted SponsorFlags as a uint32.
	sponsorFlags, _ := typecheck.ToUint32(tx["SponsorFlags"])

	entry, err := fetchSponsorshipEntry(ctx, request, isNotFound, sponsor, sponsee)
	if err != nil {
		return SponsorshipValidation{}, err
	}

	if entry == nil {
		// rippled requires a Sponsorship entry only for pre-funded sponsorship. A present
		// SponsorSignature is a co-signature; the ledger verifies it cryptographically.
		if !coSigned {
			return SponsorshipValidation{
				Reason: fmt.Errorf("%w: sponsor %s, sponsee %s", ErrSponsorshipEntryNotFound, sponsor, sponsee),
				Fee:    fee,
			}, nil
		}
		return SponsorshipValidation{Valid: true, Fee: fee}, nil
	}

	if entry.Owner != sponsor || entry.Sponsee != sponsee {
		return SponsorshipValidation{}, fmt.Errorf(
			"%w: requested sponsor %s and sponsee %s, got Owner %s and Sponsee %s",
			ErrSponsorshipEntryMismatch, sponsor, sponsee, entry.Owner, entry.Sponsee,
		)
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

	// The fee comes from DropsFromString, which rejects fractional drops, so it always has a whole
	// representation.
	feeText, _ := fee.WholeString()

	if entry.FeeAmount == nil {
		return ErrSponsorshipFeeAmountMissing
	}
	feeAmount := currency.DropsFromUint64(entry.FeeAmount.Uint64())
	if feeAmount.Cmp(fee) < 0 {
		return fmt.Errorf("%w: FeeAmount %s drops, fee %s drops", ErrSponsorshipFeeBudgetExhausted, entry.FeeAmount, feeText)
	}

	if entry.MaxFee != nil {
		maxFee := currency.DropsFromUint64(entry.MaxFee.Uint64())
		if fee.Cmp(maxFee) > 0 {
			return fmt.Errorf("%w: MaxFee %s drops, fee %s drops", ErrSponsorshipMaxFeeExceeded, entry.MaxFee, feeText)
		}
	}

	return nil
}

// sponsorshipParties resolves the sponsor and the sponsee the lookup uses, as classic addresses.
// xrpld accepts only classic addresses in the sponsorship selector, so X-addresses are converted
// with the autofill helper on a copy, leaving the caller's transaction unchanged. The sponsee is
// the Delegate when present and the Account otherwise, as rippled's getInitiator resolves it.
func sponsorshipParties(tx map[string]any) (types.Address, types.Address, error) {
	normalized := maps.Clone(tx)
	// Batch inner transactions are shared with the caller's map and play no part in this lookup,
	// so they are left out rather than rewritten in place.
	delete(normalized, "RawTransactions")
	if err := SetValidAddresses(normalized); err != nil {
		return "", "", err
	}

	sponsor, _ := typecheck.ToString(normalized["Sponsor"])
	sponsee, _ := typecheck.ToString(normalized["Delegate"])
	if sponsee == "" {
		sponsee, _ = typecheck.ToString(normalized["Account"])
	}
	return types.Address(sponsor), types.Address(sponsee), nil
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
