package builder

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
)

// SpendingBalanceParams holds the inputs for GetSpendingBalance.
type SpendingBalanceParams struct {
	// Holder is the account whose balance is read, as a classic or X-address.
	Holder string
	// IssuanceID is the MPTokenIssuanceID the balance belongs to.
	IssuanceID string
	// HolderPrivKey is the holder's ElGamal private key.
	HolderPrivKey string
	// BalanceRange bounds the decryption search. It is capped at the issuance's
	// ConfidentialOutstandingAmount.
	BalanceRange elgamal.AmountRange
}

// GetSpendingBalance reads a holder's ConfidentialBalanceSpending from one validated ledger
// and decrypts it with the holder's ElGamal private key. The inbox is not included, because
// it cannot be spent until a ConfidentialMPTMergeInbox moves it. An MPToken with no spending
// ciphertext reads as zero without decrypting. Decryption requires CGo.
//
// Errors: the parameter validation errors (ErrMissingHolder, ErrInvalidHolder,
// ErrMissingIssuanceID, ErrInvalidIssuanceID, ErrIssuerNotAllowed, ErrMissingHolderKey,
// ErrInvalidPrivKey), ErrMPTokenNotFound, ErrIssuanceNotFound, ErrConfidentialDisabled,
// ErrLedgerQuery, ErrInvalidLedgerState, elgamal.ErrInvalidAmountRange when BalanceRange is
// invalid or starts above the issuance's confidential supply, and ErrCryptoFailed when
// decryption fails, including when the balance is outside the search range.
func GetSpendingBalance(q LedgerQuerier, p SpendingBalanceParams) (uint64, error) {
	if err := validateSpendingBalanceParams(p); err != nil {
		return 0, err
	}

	// No account query: the first entry read selects the validated ledger.
	snapshot := &ledgerSnapshot{q: q}

	resp, err := getMPTokenEntry(snapshot, p.IssuanceID, p.Holder)
	if err != nil {
		return 0, err
	}
	balanceCt, err := optionalString(resp.Node, "ConfidentialBalanceSpending")
	if err != nil {
		return 0, err
	}
	if balanceCt == "" {
		return 0, nil
	}

	issuance, err := readIssuance(snapshot, p.IssuanceID)
	if err != nil {
		return 0, err
	}
	searchRange, err := boundedBalanceRange(p.BalanceRange, issuance.confidentialOutstanding)
	if err != nil {
		return 0, err
	}

	balance, err := elgamal.Decrypt(balanceCt, p.HolderPrivKey, searchRange)
	if err != nil {
		return 0, fmt.Errorf("%w: failed to decrypt spending balance: %w", ErrCryptoFailed, err)
	}
	return balance, nil
}

// validateSpendingBalanceParams rejects invalid inputs before any ledger query.
func validateSpendingBalanceParams(p SpendingBalanceParams) error {
	if p.Holder == "" {
		return ErrMissingHolder
	}
	if _, err := decodeBuilderAddress(p.Holder); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidHolder, err)
	}
	if p.IssuanceID == "" {
		return ErrMissingIssuanceID
	}
	if err := validateHolderRole(p.IssuanceID, p.Holder); err != nil {
		return err
	}
	if p.HolderPrivKey == "" {
		return ErrMissingHolderKey
	}
	if !isValidPrivateKey(p.HolderPrivKey) {
		return ErrInvalidPrivKey
	}
	return p.BalanceRange.Validate()
}

// boundedBalanceRange caps High at the issuance's confidential supply, which no single
// balance can exceed. A Low above the supply is an invalid range.
func boundedBalanceRange(searchRange elgamal.AmountRange, confidentialOutstanding uint64) (elgamal.AmountRange, error) {
	if searchRange.Low > confidentialOutstanding {
		return elgamal.AmountRange{}, fmt.Errorf("%w: BalanceRange.Low %d exceeds the issuance confidential outstanding amount %d",
			elgamal.ErrInvalidAmountRange, searchRange.Low, confidentialOutstanding)
	}
	if searchRange.High > confidentialOutstanding {
		searchRange.High = confidentialOutstanding
	}
	return searchRange, nil
}
