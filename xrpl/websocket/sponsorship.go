package websocket

import (
	"context"
	"errors"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
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
	return clientinternal.ValidateSponsorship(ctx, c.requestResultFunc(), isEntryNotFoundError, tx, estimatedFee)
}

// isEntryNotFoundError reports whether a ledger_entry request failed because the ledger has no
// matching entry.
func isEntryNotFoundError(err error) bool {
	var responseErr *ErrorWebsocketClientXrplResponse
	return errors.As(err, &responseErr) && responseErr.Type == entryNotFound
}
