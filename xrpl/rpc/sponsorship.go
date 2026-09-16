package rpc

import (
	"context"
	"errors"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	ledgerentry "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	ledgerquery "github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// SponsorshipValidation reports the outcome of an online sponsorship preflight.
type SponsorshipValidation = clientinternal.SponsorshipValidation

// ValidateSponsorship checks a sponsored transaction against the Sponsorship ledger entry between
// its sponsor and sponsee, using the current ledger.
func (c *Client) ValidateSponsorship(
	tx transaction.FlatTransaction,
	estimatedFee string,
) (SponsorshipValidation, error) {
	return c.ValidateSponsorshipContext(context.Background(), tx, estimatedFee)
}

// ValidateSponsorshipContext is ValidateSponsorship with a caller-supplied context governing the
// ledger_entry lookup.
func (c *Client) ValidateSponsorshipContext(
	ctx context.Context,
	tx transaction.FlatTransaction,
	estimatedFee string,
) (SponsorshipValidation, error) {
	return clientinternal.ValidateSponsorship(tx, estimatedFee,
		func(sponsor, sponsee types.Address) (*ledgerentry.Sponsorship, error) {
			return c.fetchSponsorshipEntry(ctx, sponsor, sponsee)
		},
	)
}

// fetchSponsorshipEntry looks up the Sponsorship entry for a sponsor and sponsee.
func (c *Client) fetchSponsorshipEntry(
	ctx context.Context,
	sponsor, sponsee types.Address,
) (*ledgerentry.Sponsorship, error) {
	var response ledgerquery.EntryResponse
	if err := c.requestResult(ctx, clientinternal.SponsorshipEntryRequest(sponsor, sponsee), &response); err != nil {
		if isEntryNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return clientinternal.DecodeSponsorshipEntry(response.Node)
}

// isEntryNotFoundError reports whether a ledger_entry request failed because the ledger has no
// matching entry.
func isEntryNotFoundError(err error) bool {
	var clientErr *ClientError
	return errors.As(err, &clientErr) && clientErr.ErrorString == entryNotFound
}
