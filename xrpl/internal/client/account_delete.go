package client

import (
	"context"
	"fmt"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// CheckAccountDeleteBlockers checks validated ledger objects and sponsorship before account deletion.
func CheckAccountDeleteBlockers(ctx context.Context, request RequestResultFunc, address types.Address, destination string) error {
	var accObjects account.ObjectsResponse
	if err := request(ctx, &account.ObjectsRequest{
		Account:              address,
		LedgerIndex:          common.LedgerTitle("validated"),
		DeletionBlockersOnly: true,
	}, &accObjects); err != nil {
		return err
	}

	if len(accObjects.AccountObjects) > 0 {
		return ErrAccountCannotBeDeleted
	}

	var info account.InfoResponse
	if err := request(ctx, &account.InfoRequest{
		Account:     address,
		LedgerIndex: common.LedgerTitle("validated"),
	}, &info); err != nil {
		return err
	}
	root := info.AccountData
	// Field presence blocks deletion, including explicit zero.
	if root.SponsoringOwnerCount != nil || root.SponsoringAccountCount != nil {
		return ErrAccountHasSponsorshipObligations
	}
	if root.Sponsor != "" && destination != "" {
		sponsor, err := addresscodec.DecodeAddress(root.Sponsor.String())
		if err != nil {
			return fmt.Errorf("decode account Sponsor: %w", err)
		}
		dest, err := addresscodec.DecodeAddress(destination)
		if err != nil {
			return fmt.Errorf("decode AccountDelete Destination: %w", err)
		}
		if sponsor.AccountID != dest.AccountID {
			return ErrAccountDeleteSponsorMismatch
		}
	}
	return nil
}
