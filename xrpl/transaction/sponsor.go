package transaction

import (
	"bytes"
	"fmt"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

func validateSponsorFields(tx *BaseTx) error {
	if tx.Sponsor == "" && tx.SponsorFlags == 0 && tx.SponsorSignature == nil {
		return nil
	}
	if tx.Sponsor == "" || tx.SponsorFlags == 0 {
		return ErrSponsorFieldsMissing
	}
	sponsorID, hasTag, err := decodeAddressAccountID(tx.Sponsor)
	if err != nil {
		return ErrInvalidSponsor
	}
	if addresscodec.IsZeroAccountID(sponsorID) {
		return ErrSponsorZero
	}
	if hasTag {
		return ErrSponsorTagNotAllowed
	}
	accountID, _, err := decodeAddressAccountID(tx.Account)
	if err != nil {
		return ErrInvalidAccount
	}
	if bytes.Equal(sponsorID, accountID) {
		return ErrSponsorAccountConflict
	}
	if tx.SponsorFlags & ^(types.SpfSponsorFee|types.SpfSponsorReserve) != 0 {
		return ErrInvalidSponsorFlags
	}
	if tx.SponsorFlags&types.SpfSponsorReserve != 0 {
		if !isReserveSponsorable(tx.TransactionType) {
			return fmt.Errorf("%w: %s", ErrReserveSponsorshipNotAllowed, tx.TransactionType)
		}
		if tx.Delegate != "" {
			return ErrSponsorDelegateConflict
		}
	}
	inner := tx.Flags&types.TfInnerBatchTxn != 0
	if inner && tx.SponsorFlags&types.SpfSponsorFee != 0 {
		return ErrInnerBatchFeeSponsorship
	}
	return validateSponsorSignature(tx.SponsorSignature, inner)
}

// isReserveSponsorable follows isReserveSponsorAllowed in rippled
// 21890d9dafac235d8c631486c6026471b21c970e. Object creation alone is not sufficient.
func isReserveSponsorable(txType TxType) bool {
	switch txType.String() {
	case "DelegateSet", "DepositPreauth", "Payment", "SignerListSet",
		"CheckCancel", "CheckCash", "CheckCreate", "EscrowCancel", "EscrowCreate", "EscrowFinish",
		"PaymentChannelClaim", "PaymentChannelCreate", "PaymentChannelFund", "Clawback",
		"MPTokenAuthorize", "MPTokenIssuanceCreate", "MPTokenIssuanceDestroy", "MPTokenIssuanceSet",
		"TrustSet", "CredentialAccept", "CredentialCreate", "CredentialDelete",
		"AccountSet", "SetRegularKey", "SponsorshipTransfer":
		return true
	default:
		return false
	}
}

func validateSponsorSignature(signature *types.SponsorSignature, inner bool) error {
	if signature == nil {
		return nil
	}
	if inner {
		if signature.SigningPubKey == nil || *signature.SigningPubKey != "" ||
			signature.TxnSignature != nil || signature.Signers != nil {
			return ErrInnerBatchSponsorSignature
		}
		return nil
	}
	if signature.Signers != nil {
		if len(signature.Signers) == 0 ||
			(signature.SigningPubKey != nil && *signature.SigningPubKey != "") || signature.TxnSignature != nil {
			return ErrInvalidSponsorSignature
		}
		return validateSigners(signature.Signers)
	}
	if signature.SigningPubKey == nil || *signature.SigningPubKey == "" ||
		signature.TxnSignature == nil || *signature.TxnSignature == "" {
		return ErrInvalidSponsorSignature
	}
	return nil
}

// validateRawSponsorFields adapts Batch raw maps to the common sponsor validator.
// Remove this private adapter once Batch validates its inner transactions directly.
func validateRawSponsorFields(raw map[string]any) error {
	_, hasSponsor := raw["Sponsor"]
	_, hasFlags := raw["SponsorFlags"]
	_, hasSignature := raw["SponsorSignature"]
	if !hasSponsor && !hasFlags && !hasSignature {
		return nil
	}
	// Report omitted required keys before converting any supplied values.
	// When both keys exist, malformed values retain their field-specific errors.
	if !hasSponsor || !hasFlags {
		return ErrSponsorFieldsMissing
	}
	sponsor, ok := typecheck.ToString(raw["Sponsor"])
	if !ok || sponsor == "" {
		return ErrInvalidSponsor
	}
	flags, ok := typecheck.ToUint32(raw["SponsorFlags"])
	if !ok || flags == 0 {
		return ErrInvalidSponsorFlags
	}
	account, ok := typecheck.ToString(raw["Account"])
	if !ok {
		return ErrInvalidAccount
	}
	txType, ok := typecheck.ToString(raw["TransactionType"])
	if !ok {
		return ErrInvalidTransactionType
	}
	// Batch.Validate is the only caller and runs RawTransaction.Validate first,
	// which requires Flags to be uint32 and contain TfInnerBatchTxn.
	txFlags, _ := typecheck.ToUint32(raw["Flags"])
	tx := BaseTx{
		Account:         types.Address(account),
		TransactionType: TxType(txType),
		Sponsor:         types.Address(sponsor),
		SponsorFlags:    flags,
		Flags:           txFlags,
	}
	if value, present := raw["Delegate"]; present {
		delegate, ok := typecheck.ToString(value)
		if !ok || delegate == "" {
			return ErrInvalidDelegate
		}
		tx.Delegate = types.Address(delegate)
	}
	if hasSignature {
		signature, ok := raw["SponsorSignature"].(map[string]any)
		if !ok {
			return ErrInvalidSponsorSignature
		}
		// Only the unsigned placeholder is allowed. Check before conversion so
		// extra fields cannot disappear from validation while remaining on the wire.
		if len(signature) != 1 {
			return ErrInnerBatchSponsorSignature
		}
		key, ok := signature["SigningPubKey"].(string)
		if !ok {
			return ErrInnerBatchSponsorSignature
		}
		tx.SponsorSignature = &types.SponsorSignature{SigningPubKey: &key}
	}
	return validateSponsorFields(&tx)
}
