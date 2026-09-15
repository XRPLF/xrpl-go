package transaction

import (
	"bytes"
	"fmt"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// sponsorFields holds the values used by the shared sponsorship rules.
// present distinguishes absent sponsorship from explicitly supplied zero values.
type sponsorFields struct {
	account   types.Address
	sponsor   types.Address
	flags     uint32
	signature *sponsorSignatureFields
	inner     bool
	delegate  types.Address
	txType    TxType
	present   bool
}

// sponsorSignatureFields preserves field presence independently of field values.
// For inner transactions, only the presence of TxnSignature and Signers matters.
type sponsorSignatureFields struct {
	signingPubKey   string
	txnSignature    string
	signers         []types.Signer
	hasTxnSignature bool
	hasSigners      bool
}

func validateSponsorFields(fields sponsorFields) error {
	if !fields.present {
		return nil
	}
	// Client policy: consensus-generated transactions cannot be sponsored.
	if IsPseudoTransactionType(fields.txType) {
		return ErrPseudoTransactionSponsorship
	}
	if fields.sponsor == "" || fields.flags == 0 {
		return ErrSponsorFieldsMissing
	}
	sponsorID, hasTag, err := decodeAddressAccountID(fields.sponsor)
	if err != nil {
		return ErrInvalidSponsor
	}
	if addresscodec.IsZeroAccountID(sponsorID) {
		return ErrSponsorZero
	}
	if hasTag {
		return ErrSponsorTagNotAllowed
	}
	accountID, _, err := decodeAddressAccountID(fields.account)
	if err != nil {
		return ErrInvalidAccount
	}
	if bytes.Equal(sponsorID, accountID) {
		return ErrSponsorAccountConflict
	}
	if fields.flags & ^(types.SpfSponsorFee|types.SpfSponsorReserve) != 0 {
		return ErrInvalidSponsorFlags
	}
	if fields.flags&types.SpfSponsorReserve != 0 {
		if !isReserveSponsorable(fields.txType) {
			return fmt.Errorf("%w: %s", ErrReserveSponsorshipNotAllowed, fields.txType)
		}
		if fields.delegate != "" {
			return ErrSponsorDelegateConflict
		}
	}
	if fields.inner && fields.flags&types.SpfSponsorFee != 0 {
		return ErrInnerBatchFeeSponsorship
	}
	return validateSponsorSignature(fields.signature, fields.inner)
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

func validateSponsorSignature(signature *sponsorSignatureFields, inner bool) error {
	if signature == nil {
		return nil
	}
	if inner {
		if signature.signingPubKey != "" || signature.hasTxnSignature || signature.hasSigners {
			return ErrInnerBatchSponsorSignature
		}
		return nil
	}
	if signature.hasSigners {
		if len(signature.signers) == 0 || signature.signingPubKey != "" || signature.hasTxnSignature {
			return ErrInvalidSponsorSignature
		}
		return validateSigners(signature.signers)
	}
	if signature.signingPubKey == "" || signature.txnSignature == "" {
		return ErrInvalidSponsorSignature
	}
	return nil
}

func sponsorFieldsFromBaseTx(tx *BaseTx) sponsorFields {
	return sponsorFields{
		account:   tx.Account,
		sponsor:   tx.Sponsor,
		flags:     tx.SponsorFlags,
		signature: sponsorSignatureFromTyped(tx.SponsorSignature),
		inner:     tx.Flags&types.TfInnerBatchTxn != 0,
		delegate:  tx.Delegate,
		txType:    tx.TransactionType,
		present:   tx.Sponsor != "" || tx.SponsorFlags != 0 || tx.SponsorSignature != nil,
	}
}

func sponsorSignatureFromTyped(signature *types.SponsorSignature) *sponsorSignatureFields {
	if signature == nil {
		return nil
	}
	fields := &sponsorSignatureFields{
		signers:         signature.Signers,
		hasTxnSignature: signature.TxnSignature != nil,
		hasSigners:      signature.Signers != nil,
	}
	if signature.SigningPubKey != nil {
		fields.signingPubKey = *signature.SigningPubKey
	}
	if signature.TxnSignature != nil {
		fields.txnSignature = *signature.TxnSignature
	}
	return fields
}

// sponsorFieldsFromInnerRaw converts a Batch inner map after RawTransaction.Validate.
// JSON conversion errors are reported before the shared sponsorship rules run.
func sponsorFieldsFromInnerRaw(raw map[string]any) (sponsorFields, error) {
	_, hasSponsor := raw["Sponsor"]
	_, hasFlags := raw["SponsorFlags"]
	_, hasSignature := raw["SponsorSignature"]
	fields := sponsorFields{inner: true, present: hasSponsor || hasFlags || hasSignature}
	if !fields.present {
		return fields, nil
	}
	if hasSponsor {
		sponsor, ok := typecheck.ToString(raw["Sponsor"])
		if !ok {
			return sponsorFields{}, ErrInvalidSponsor
		}
		fields.sponsor = types.Address(sponsor)
	}
	if hasFlags {
		flags, ok := typecheck.ToUint32(raw["SponsorFlags"])
		if !ok {
			return sponsorFields{}, ErrInvalidSponsorFlags
		}
		fields.flags = flags
	}
	account, ok := typecheck.ToString(raw["Account"])
	if !ok {
		return sponsorFields{}, ErrInvalidAccount
	}
	fields.account = types.Address(account)
	txType, ok := typecheck.ToString(raw["TransactionType"])
	if !ok {
		return sponsorFields{}, ErrInvalidTransactionType
	}
	fields.txType = TxType(txType)
	if value, present := raw["Delegate"]; present {
		delegate, ok := typecheck.ToString(value)
		if !ok || delegate == "" {
			return sponsorFields{}, ErrInvalidDelegate
		}
		fields.delegate = types.Address(delegate)
	}
	if hasSignature {
		signature, err := sponsorSignatureFromInnerRaw(raw["SponsorSignature"])
		if err != nil {
			return sponsorFields{}, err
		}
		fields.signature = signature
	}
	return fields, nil
}

func sponsorSignatureFromInnerRaw(value any) (*sponsorSignatureFields, error) {
	signature, ok := value.(map[string]any)
	if !ok {
		return nil, ErrInvalidSponsorSignature
	}
	if signature == nil {
		return nil, ErrInnerBatchSponsorSignature
	}
	// Unknown fields must not disappear during conversion while remaining on the wire.
	for name := range signature {
		switch name {
		case "SigningPubKey", "TxnSignature", "Signers":
		default:
			return nil, ErrInnerBatchSponsorSignature
		}
	}
	fields := &sponsorSignatureFields{}
	if value, present := signature["SigningPubKey"]; present {
		key, ok := value.(string)
		if !ok {
			return nil, ErrInnerBatchSponsorSignature
		}
		fields.signingPubKey = key
	}
	// Inner signatures forbid these fields regardless of their values, including null.
	// Record their presence for the shared validator without decoding unused payloads.
	_, fields.hasTxnSignature = signature["TxnSignature"]
	_, fields.hasSigners = signature["Signers"]
	return fields, nil
}
