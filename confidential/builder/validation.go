package builder

import (
	"fmt"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// decodeBuilderAddress decodes an address field and rejects what a builder can never
// use: input that is neither address form, and ACCOUNT_ZERO, which decodes cleanly but
// can never sign a transaction nor hold an MPToken. Callers read the tag state and
// compare accounts from the result.
//
// The returned error names the reason. Callers wrap it in their field-specific sentinel,
// so a user can match the field with errors.Is and still read why the address failed.
// ACCOUNT_ZERO reports transaction.ErrZeroAccountID, the condition BaseTx.Validate also
// reports, so a builder and a preflight stay matchable as one condition.
func decodeBuilderAddress(address string) (addresscodec.DecodedAddress, error) {
	decoded, err := addresscodec.DecodeAddress(address)
	if err != nil {
		return addresscodec.DecodedAddress{}, err
	}
	if addresscodec.IsZeroAccountID(decoded.AccountID[:]) {
		return addresscodec.DecodedAddress{}, transaction.ErrZeroAccountID
	}
	return decoded, nil
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

type txValidator interface {
	Validate() (bool, error)
}

// validatePreparedTransaction runs a prepared transaction through its own Validate so a
// field the builder assembled wrongly is reported here instead of at submission. Both
// failure signals are honored, because txValidator is satisfied by any transaction type
// and a rejection reported through the bool alone would otherwise pass unnoticed.
func validatePreparedTransaction(tx txValidator) error {
	ok, err := tx.Validate()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidTransaction, err)
	}
	if !ok {
		return ErrInvalidTransaction
	}
	return nil
}
