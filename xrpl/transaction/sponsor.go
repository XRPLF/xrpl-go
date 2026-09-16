package transaction

import (
	"bytes"
	"fmt"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/flag"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// sponsorFields holds the values used by the shared sponsorship rules.
// present distinguishes absent sponsorship from explicitly supplied zero values.
type sponsorFields struct {
	account   types.Address
	sponsor   types.Address
	flags     uint32
	signature *types.SponsorSignature
	inner     bool
	delegate  types.Address
	txType    TxType
	present   bool
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
	switch txType { //nolint:exhaustive // All unlisted transaction types are deliberately rejected.
	case DelegateSetTx, DepositPreauthTx, PaymentTx, SignerListSetTx,
		CheckCancelTx, CheckCashTx, CheckCreateTx, EscrowCancelTx, EscrowCreateTx, EscrowFinishTx,
		PaymentChannelClaimTx, PaymentChannelCreateTx, PaymentChannelFundTx, ClawbackTx,
		MPTokenAuthorizeTx, MPTokenIssuanceCreateTx, MPTokenIssuanceDestroyTx, MPTokenIssuanceSetTx,
		TrustSetTx, CredentialAcceptTx, CredentialCreateTx, CredentialDeleteTx,
		AccountSetTx, SetRegularKeyTx, SponsorshipTransferTx:
		return true
	default:
		return false
	}
}

func validateSponsorSignature(signature *types.SponsorSignature, inner bool) error {
	if signature == nil {
		return nil
	}
	var signingPubKey string
	if signature.SigningPubKey != nil {
		signingPubKey = *signature.SigningPubKey
	}
	if inner {
		if signingPubKey != "" || signature.TxnSignature != nil || signature.Signers != nil {
			return ErrInnerBatchSponsorSignature
		}
		return nil
	}
	if signature.Signers != nil {
		if len(signature.Signers) == 0 || signingPubKey != "" || signature.TxnSignature != nil {
			return ErrInvalidSponsorSignature
		}
		return validateSigners(signature.Signers)
	}
	if signingPubKey == "" || signature.TxnSignature == nil || *signature.TxnSignature == "" {
		return ErrInvalidSponsorSignature
	}
	return nil
}

func sponsorFieldsFromBaseTx(tx *BaseTx) sponsorFields {
	return sponsorFields{
		account:   tx.Account,
		sponsor:   tx.Sponsor,
		flags:     tx.SponsorFlags,
		signature: tx.SponsorSignature,
		inner:     flag.Contains(tx.Flags, types.TfInnerBatchTxn),
		delegate:  tx.Delegate,
		txType:    tx.TransactionType,
		present:   tx.Sponsor != "" || tx.SponsorFlags != 0 || tx.SponsorSignature != nil,
	}
}

// InspectSponsorFields validates raw sponsorship fields and returns an independent
// typed sponsor signature, or nil when absent. It accepts flattened and binary-decoded
// maps without changing them, including pre-funded transactions before autofill.
// It preserves absent versus empty SigningPubKey values and checks only sponsorship
// structure and context, not full transaction validity, cryptographic signatures,
// ledger authorization, or sponsorship balances. On error it returns no signature.
func InspectSponsorFields(tx FlatTransaction) (*types.SponsorSignature, error) {
	var inner bool
	if value, present := tx["Flags"]; present {
		flags, ok := typecheck.ToUint32(value)
		if !ok {
			return nil, ErrInvalidFlagsValue
		}
		inner = flag.Contains(flags, types.TfInnerBatchTxn)
	}
	fields, err := sponsorFieldsFromRaw(tx, inner)
	if err != nil {
		return nil, err
	}
	if err := validateSponsorFields(fields); err != nil {
		return nil, err
	}
	return fields.signature, nil
}

// sponsorFieldsFromInnerRaw converts a Batch inner map after RawTransaction.Validate.
func sponsorFieldsFromInnerRaw(raw map[string]any) (sponsorFields, error) {
	return sponsorFieldsFromRaw(raw, true)
}

func sponsorFieldsFromRaw(raw map[string]any, inner bool) (sponsorFields, error) {
	_, hasSponsor := raw["Sponsor"]
	_, hasFlags := raw["SponsorFlags"]
	_, hasSignature := raw["SponsorSignature"]
	fields := sponsorFields{inner: inner, present: hasSponsor || hasFlags || hasSignature}
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
		var err error
		if inner {
			fields.signature, err = sponsorSignatureFromInnerRaw(raw["SponsorSignature"])
		} else {
			fields.signature, err = sponsorSignatureFromRaw(raw["SponsorSignature"])
		}
		if err != nil {
			return sponsorFields{}, err
		}
	}
	return fields, nil
}

// sponsorSignatureFromRaw preserves presence before applying the shared typed rules.
func sponsorSignatureFromRaw(value any) (*types.SponsorSignature, error) {
	signature, ok := value.(map[string]any)
	if !ok || signature == nil {
		return nil, ErrInvalidSponsorSignature
	}
	fields := &types.SponsorSignature{}
	for name, value := range signature {
		switch name {
		case "SigningPubKey":
			key, ok := value.(string)
			if !ok {
				return nil, ErrInvalidSponsorSignature
			}
			fields.SigningPubKey = &key
		case "TxnSignature":
			sig, ok := value.(string)
			if !ok {
				return nil, ErrInvalidSponsorSignature
			}
			fields.TxnSignature = &sig
		case "Signers":
			signers, err := sponsorSignersFromRaw(value)
			if err != nil {
				return nil, err
			}
			fields.Signers = signers
		default:
			return nil, ErrInvalidSponsorSignature
		}
	}
	return fields, nil
}

func sponsorSignatureFromInnerRaw(value any) (*types.SponsorSignature, error) {
	signature, ok := value.(map[string]any)
	if !ok {
		return nil, ErrInvalidSponsorSignature
	}
	if signature == nil {
		return nil, ErrInnerBatchSponsorSignature
	}
	// Inner signatures allow only an absent or empty SigningPubKey.
	fields := &types.SponsorSignature{}
	for name, value := range signature {
		if name != "SigningPubKey" {
			return nil, ErrInnerBatchSponsorSignature
		}
		key, ok := value.(string)
		if !ok || key != "" {
			return nil, ErrInnerBatchSponsorSignature
		}
		fields.SigningPubKey = &key
	}
	return fields, nil
}

func sponsorSignersFromRaw(value any) ([]types.Signer, error) {
	var entries []any
	switch signers := value.(type) {
	case []any:
		entries = signers
	case []map[string]any:
		entries = make([]any, len(signers))
		for i, signer := range signers {
			entries[i] = signer
		}
	default:
		return nil, ErrInvalidSponsorSignature
	}
	signers := make([]types.Signer, len(entries))
	for i, entry := range entries {
		wrapper, ok := entry.(map[string]any)
		if !ok || len(wrapper) != 1 {
			return nil, ErrInvalidSponsorSignature
		}
		signer, ok := wrapper["Signer"].(map[string]any)
		if !ok || len(signer) != 3 {
			return nil, ErrInvalidSponsorSignature
		}
		account, accountOK := signer["Account"].(string)
		key, keyOK := signer["SigningPubKey"].(string)
		sig, sigOK := signer["TxnSignature"].(string)
		if !accountOK || !keyOK || !sigOK {
			return nil, ErrInvalidSponsorSignature
		}
		signers[i].SignerData = types.SignerData{
			Account: types.Address(account), SigningPubKey: key, TxnSignature: sig,
		}
	}
	return signers, nil
}
