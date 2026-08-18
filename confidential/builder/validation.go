package builder

import (
	"github.com/Peersyst/xrpl-go/pkg/mptsizes"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// privKeyLen is the hex-encoded length of a private key scalar. Derived from the byte
// length the proof and elgamal decoders enforce, so the two cannot drift.
const privKeyLen = 2 * mptsizes.PrivKeySize

// isValidPrivKey checks if the given hex string is a valid 32-byte private key scalar (64 hex chars).
// Private keys are builder inputs only. They are never carried by a transaction.
func isValidPrivKey(key string) bool {
	return len(key) == privKeyLen && typecheck.IsHex(key)
}

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
