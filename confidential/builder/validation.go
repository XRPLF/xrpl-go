package builder

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// validateHolderRole validates an issuance ID submitted by a holder. XLS-96 forbids the
// issuer from converting, sending, merging, or converting back its own issuance, so the
// builder rejects it before doing any ledger or proof work.
func validateHolderRole(issuanceID, account string) error {
	if !transaction.IsMPTIssuanceID(issuanceID) {
		return ErrInvalidIssuanceID
	}
	if transaction.IsMPTokenIssuer(issuanceID, types.Address(account)) {
		return ErrIssuerNotAllowed
	}
	return nil
}

// validateIssuerRole validates an issuance ID submitted by the issuer. XLS-96 requires
// the clawback submitter to be the issuance issuer.
func validateIssuerRole(issuanceID, account string) error {
	if !transaction.IsMPTIssuanceID(issuanceID) {
		return ErrInvalidIssuanceID
	}
	if !transaction.IsMPTokenIssuer(issuanceID, types.Address(account)) {
		return ErrNotIssuer
	}
	return nil
}

// validateDestinationNotIssuer rejects a send whose destination is the issuance issuer.
// XLS-96: the issuer can never hold a confidential balance, so such a send always fails
// with tecNO_PERMISSION. Rejecting it here avoids paying the fee for a doomed transaction.
//
// Callers must validate issuanceID first, because IsMPTokenIssuer reports false for a
// malformed ID and that would read as "not the issuer" here.
func validateDestinationNotIssuer(issuanceID, destination string) error {
	if transaction.IsMPTokenIssuer(issuanceID, types.Address(destination)) {
		return ErrDestinationIsIssuer
	}
	return nil
}

// validateAmount rejects a zero amount and any amount above the protocol cap.
// XLS-96 caps confidential amounts at MaxMPTAmount even though the range
// proof itself covers the full 64-bit range.
func validateAmount(amount uint64) error {
	if amount == 0 {
		return ErrZeroAmount
	}
	return validateAmountUpperBound(amount)
}

// validateAmountUpperBound enforces only the upper bound. ConfidentialMPTConvert uses it
// because XLS-96 makes a zero-amount convert the opt-in path that registers a
// holder encryption key.
func validateAmountUpperBound(amount uint64) error {
	if !types.MPTPlainAmount(amount).IsValid() {
		return ErrAmountTooLarge
	}
	return nil
}
