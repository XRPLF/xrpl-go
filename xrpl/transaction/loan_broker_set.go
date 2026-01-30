package transaction

import (
	"strconv"

	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// LoanBrokerSetMaxDataLength is the maximum length in characters for the Data field.
	LoanBrokerSetMaxDataLength = 512
	// LoanBrokerSetMaxManagementFeeRate is the maximum value for ManagementFeeRate (10000 = 10%).
	LoanBrokerSetMaxManagementFeeRate = 10000
	// LoanBrokerSetMaxCoverRateMinimum is the maximum value for CoverRateMinimum (100000 = 100%).
	LoanBrokerSetMaxCoverRateMinimum = 100000
	// LoanBrokerSetMaxCoverRateLiquidation is the maximum value for CoverRateLiquidation (100000 = 100%).
	LoanBrokerSetMaxCoverRateLiquidation = 100000
)

// LoanBrokerSet creates a new LoanBroker object or updates an existing one.
//
// ```json
//
//	{
//	  "TransactionType": "LoanBrokerSet",
//	  "Account": "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
//	  "VaultID": "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"
//	}
//
// ```
type LoanBrokerSet struct {
	BaseTx
	// The Vault ID that the Lending Protocol will use to access liquidity.
	VaultID string
	// The Loan Broker ID that the transaction is modifying.
	LoanBrokerID *types.LoanBrokerID `json:",omitempty"`
	// Arbitrary metadata in hex format. The field is limited to 512 characters.
	Data *types.Data `json:",omitempty"`
	// The 1/10th basis point fee charged by the Lending Protocol Owner. Valid values are between 0 and 10000 inclusive (1% - 10%).
	ManagementFeeRate *types.InterestRate `json:",omitempty"`
	// The maximum amount the protocol can owe the Vault.
	// The default value of 0 means there is no limit to the debt. Must not be negative.
	DebtMaximum *types.XRPLNumber `json:",omitempty"`
	// The 1/10th basis point DebtTotal that the first loss capital must cover. Valid values are between 0 and 100000 inclusive.
	CoverRateMinimum *types.InterestRate `json:",omitempty"`
	// The 1/10th basis point of minimum required first loss capital liquidated to cover a Loan default.
	// Valid values are between 0 and 100000 inclusive.
	CoverRateLiquidation *types.InterestRate `json:",omitempty"`
}

// TxType returns the TxType for LoanBrokerSet transactions.
func (tx *LoanBrokerSet) TxType() TxType {
	return LoanBrokerSetTx
}

// Flatten returns a map representation of the LoanBrokerSet transaction for JSON-RPC submission.
func (tx *LoanBrokerSet) Flatten() map[string]interface{} {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	if tx.Account != "" {
		flattened["Account"] = tx.Account.String()
	}

	flattened["VaultID"] = tx.VaultID

	if tx.LoanBrokerID != nil && *tx.LoanBrokerID != "" {
		flattened["LoanBrokerID"] = string(*tx.LoanBrokerID)
	}

	if tx.Data != nil && *tx.Data != "" {
		flattened["Data"] = string(*tx.Data)
	}

	if tx.ManagementFeeRate != nil && *tx.ManagementFeeRate != 0 {
		flattened["ManagementFeeRate"] = uint32(*tx.ManagementFeeRate)
	}

	if tx.DebtMaximum != nil && *tx.DebtMaximum != "" {
		flattened["DebtMaximum"] = tx.DebtMaximum.String()
	}

	if tx.CoverRateMinimum != nil && *tx.CoverRateMinimum != 0 {
		flattened["CoverRateMinimum"] = uint32(*tx.CoverRateMinimum)
	}

	if tx.CoverRateLiquidation != nil && *tx.CoverRateLiquidation != 0 {
		flattened["CoverRateLiquidation"] = uint32(*tx.CoverRateLiquidation)
	}

	return flattened
}

// Validate checks LoanBrokerSet transaction fields and returns false with an error if invalid.
func (tx *LoanBrokerSet) Validate() (bool, error) {
	if ok, err := tx.BaseTx.Validate(); !ok {
		return false, err
	}

	if tx.VaultID == "" {
		return false, ErrLoanBrokerSetVaultIDRequired
	}

	if !IsLedgerEntryID(tx.VaultID) {
		return false, ErrLoanBrokerSetVaultIDInvalid
	}

	if tx.LoanBrokerID != nil && *tx.LoanBrokerID != "" {
		if !IsLedgerEntryID(tx.LoanBrokerID.Value()) {
			return false, ErrLoanBrokerSetLoanBrokerIDInvalid
		}
	}

	if tx.Data != nil && *tx.Data != "" {
		if !ValidateHexMetadata(tx.Data.Value(), LoanBrokerSetMaxDataLength) {
			return false, ErrLoanBrokerSetDataInvalid
		}
	}

	if tx.ManagementFeeRate != nil && *tx.ManagementFeeRate > LoanBrokerSetMaxManagementFeeRate {
		return false, ErrLoanBrokerSetManagementFeeRateInvalid
	}

	if tx.DebtMaximum != nil && *tx.DebtMaximum != "" {
		if !typecheck.IsXRPLNumber(tx.DebtMaximum.String()) {
			return false, ErrLoanBrokerSetDebtMaximumInvalid
		}
		// Check that DebtMaximum is non-negative
		val, err := strconv.ParseFloat(tx.DebtMaximum.String(), 64)
		if err != nil || val < 0 {
			return false, ErrLoanBrokerSetDebtMaximumNegative
		}
	}

	if tx.CoverRateMinimum != nil && *tx.CoverRateMinimum > LoanBrokerSetMaxCoverRateMinimum {
		return false, ErrLoanBrokerSetCoverRateMinimumInvalid
	}

	if tx.CoverRateLiquidation != nil && *tx.CoverRateLiquidation > LoanBrokerSetMaxCoverRateLiquidation {
		return false, ErrLoanBrokerSetCoverRateLiquidationInvalid
	}

	// Validate that either both CoverRateMinimum and CoverRateLiquidation are zero,
	// or both are non-zero.
	coverRateMinimumValue := uint32(0)
	if tx.CoverRateMinimum != nil {
		coverRateMinimumValue = tx.CoverRateMinimum.Value()
	}
	coverRateLiquidationValue := uint32(0)
	if tx.CoverRateLiquidation != nil {
		coverRateLiquidationValue = tx.CoverRateLiquidation.Value()
	}

	if (coverRateMinimumValue == 0 && coverRateLiquidationValue != 0) ||
		(coverRateMinimumValue != 0 && coverRateLiquidationValue == 0) {
		return false, ErrLoanBrokerSetCoverRatesMismatch
	}

	return true, nil
}
