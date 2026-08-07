package client

import (
	"fmt"
	"maps"
	"reflect"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/flag"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// SignedTransactionForm identifies the structurally valid signing form of a
// transaction. These forms do not establish cryptographic validity. rippled is
// authoritative for signature verification.
type SignedTransactionForm uint8

const (
	// UnsignedTransaction has no signing fields.
	UnsignedTransaction SignedTransactionForm = iota
	// SingleSignedTransaction has a nonempty SigningPubKey and TxnSignature.
	SingleSignedTransaction
	// MultiSignedTransaction has a nonempty, structurally complete Signers array.
	MultiSignedTransaction
	// InnerBatchTransaction is intentionally unsigned and has an empty SigningPubKey.
	InnerBatchTransaction
)

// InspectSignedTransaction returns the transaction's structural signing form.
// It checks field presence, types, and nonempty values only. rippled remains
// authoritative for whether signatures are cryptographically valid.
//
// When allowInnerBatch is true, a transaction carrying tfInnerBatchTxn is valid
// only when SigningPubKey is explicitly empty and TxnSignature and Signers are
// absent. Inner Batch transactions must never be submitted independently.
func InspectSignedTransaction(tx map[string]any, allowInnerBatch bool) (SignedTransactionForm, error) {
	pubKeyValue, hasPubKey := tx["SigningPubKey"]
	signatureValue, hasSignature := tx["TxnSignature"]
	signersValue, hasSigners := tx["Signers"]

	if isInnerBatchTransaction(tx) {
		if !allowInnerBatch {
			return UnsignedTransaction, invalidSignedForm("inner Batch transactions cannot be submitted independently")
		}
		pubKey, ok := pubKeyValue.(string)
		if !ok || pubKey != "" || hasSignature || hasSigners {
			return UnsignedTransaction, invalidSignedForm("inner Batch transactions require an empty SigningPubKey and no TxnSignature or Signers")
		}
		return InnerBatchTransaction, nil
	}

	if hasSigners {
		if hasSignature {
			return UnsignedTransaction, invalidSignedForm("TxnSignature and Signers cannot be mixed")
		}
		pubKey, ok := pubKeyValue.(string)
		if !hasPubKey || !ok || pubKey != "" {
			return UnsignedTransaction, invalidSignedForm("a multisigned transaction's top-level SigningPubKey must be present and empty")
		}
		if err := validateSigners(signersValue); err != nil {
			return UnsignedTransaction, err
		}
		return MultiSignedTransaction, nil
	}

	if !hasPubKey && !hasSignature {
		return UnsignedTransaction, nil
	}

	pubKey, pubKeyOK := pubKeyValue.(string)
	signature, signatureOK := signatureValue.(string)
	if !pubKeyOK || pubKey == "" || !signatureOK || signature == "" {
		return UnsignedTransaction, invalidSignedForm("single-signing requires nonempty SigningPubKey and TxnSignature strings")
	}

	return SingleSignedTransaction, nil
}

// InspectSignedBatchInners requires every inner transaction of a signed outer
// Batch to use the canonical inner Batch form. A valid outer signature must not
// let a malformed inner transaction (a non-empty TxnSignature/Signers, or a
// missing tfInnerBatchTxn/SigningPubKey) reach the network. Callers apply this
// to a decoded signed blob, after the inner transactions carry their final
// wire fields.
func InspectSignedBatchInners(tx map[string]any) error {
	if txType, _ := tx["TransactionType"].(string); txType != "Batch" {
		return nil
	}
	inners, err := rawTransactionObjects(tx["RawTransactions"])
	if err != nil {
		return err
	}
	if len(inners) < 2 || len(inners) > 8 {
		return fmt.Errorf("%w: got %d", ErrBatchRawTransactionsCount, len(inners))
	}
	for _, inner := range inners {
		signingType, err := InspectSignedTransaction(inner, true)
		if err != nil {
			return err
		}
		if signingType != InnerBatchTransaction {
			return invalidSignedForm("Batch inner transactions must use the inner Batch form")
		}
	}
	return nil
}

// rawTransactionObjects extracts the inner transaction maps from a Batch
// RawTransactions array, accepting both the flattened ([]map[string]any) and
// binary-decoded ([]any) representations.
func rawTransactionObjects(value any) ([]map[string]any, error) {
	var wrappers []any
	switch typed := value.(type) {
	case nil:
		return nil, nil
	case []map[string]any:
		wrappers = make([]any, len(typed))
		for i := range typed {
			wrappers[i] = typed[i]
		}
	case []any:
		wrappers = typed
	default:
		return nil, ErrRawTransactionsFieldIsNotAnArray
	}

	inners := make([]map[string]any, 0, len(wrappers))
	for _, wrapper := range wrappers {
		wrapperMap, ok := wrapper.(map[string]any)
		if !ok {
			return nil, ErrRawTransactionFieldIsNotAnObject
		}
		inner, ok := wrapperMap["RawTransaction"].(map[string]any)
		if !ok {
			return nil, ErrRawTransactionFieldIsNotAnObject
		}
		inners = append(inners, inner)
	}
	return inners, nil
}

func isInnerBatchTransaction(tx map[string]any) bool {
	flags, ok := typecheck.ToUint32(tx["Flags"])
	return ok && flag.Contains(flags, types.TfInnerBatchTxn)
}

func validateSigners(value any) error {
	signers, ok := signerEntries(value)
	if !ok || len(signers) == 0 {
		return invalidSignedForm("Signers must be a nonempty array")
	}

	for i, entry := range signers {
		wrapper, ok := entry.(map[string]any)
		if !ok {
			return invalidSignedForm(fmt.Sprintf("Signers[%d] must be an object", i))
		}
		signer, ok := wrapper["Signer"].(map[string]any)
		if !ok {
			return invalidSignedForm(fmt.Sprintf("Signers[%d].Signer must be an object", i))
		}
		for _, field := range []string{"Account", "SigningPubKey", "TxnSignature"} {
			value, ok := signer[field].(string)
			if !ok || value == "" {
				return invalidSignedForm(fmt.Sprintf("Signers[%d].Signer.%s must be a nonempty string", i, field))
			}
		}
	}

	return nil
}

func signerEntries(value any) ([]any, bool) {
	switch signers := value.(type) {
	case []any:
		return signers, true
	case []map[string]any:
		entries := make([]any, len(signers))
		for i := range signers {
			entries[i] = signers[i]
		}
		return entries, true
	default:
		return nil, false
	}
}

func invalidSignedForm(detail string) error {
	return fmt.Errorf("%w: %s", ErrInvalidSignedTransaction, detail)
}

// AutofillBatchRawTransactions validates and fills the transport-independent
// fields of flattened Batch inner transactions. fetchAccountSequence supplies
// only the client-specific account_info request.
func AutofillBatchRawTransactions(
	tx map[string]any,
	fetchAccountSequence func(string) (uint32, error),
) error {
	rawTxs, ok := tx["RawTransactions"].([]map[string]any)
	if !ok {
		return ErrRawTransactionsFieldIsNotAnArray
	}

	type validatedInner struct {
		rawTx   map[string]any
		account string
	}
	inners := make([]validatedInner, 0, len(rawTxs))
	for _, wrapper := range rawTxs {
		inner, ok := wrapper["RawTransaction"].(map[string]any)
		if !ok {
			return ErrRawTransactionFieldIsNotAnObject
		}

		account, ok := inner["Account"].(string)
		if !ok {
			return ErrAccountFieldIsNotAString
		}
		if fee, exists := inner["Fee"]; exists {
			fee, ok := fee.(string)
			if !ok || fee != "0" {
				return types.ErrBatchInnerTransactionInvalid
			}
		}
		if signingPubKey, exists := inner["SigningPubKey"]; exists {
			signingPubKey, ok := signingPubKey.(string)
			if !ok || signingPubKey != "" {
				return ErrSigningPubKeyFieldMustBeEmpty
			}
		}
		if _, exists := inner["LastLedgerSequence"]; exists {
			return ErrLastLedgerSequenceFieldMustBeAbsent
		}
		if _, exists := inner["TxnSignature"]; exists {
			return ErrTxnSignatureFieldMustBeEmpty
		}
		if _, exists := inner["Signers"]; exists {
			return ErrSignersFieldMustBeEmpty
		}

		inners = append(inners, validatedInner{rawTx: inner, account: account})
	}

	accountSequences := make(map[string]uint32, len(inners))
	for _, inner := range inners {
		if _, exists := inner.rawTx["Fee"]; !exists {
			inner.rawTx["Fee"] = "0"
		}
		if _, exists := inner.rawTx["SigningPubKey"]; !exists {
			inner.rawTx["SigningPubKey"] = ""
		}
		if inner.rawTx["Sequence"] != nil || inner.rawTx["TicketSequence"] != nil {
			continue
		}

		sequence, cached := accountSequences[inner.account]
		if !cached {
			var err error
			sequence, err = fetchAccountSequence(inner.account)
			if err != nil {
				return err
			}
			if inner.account == tx["Account"] {
				sequence++
			}
		}
		inner.rawTx["Sequence"] = sequence
		accountSequences[inner.account] = sequence + 1
	}

	return nil
}

// DecodeTransactionBlob decodes a transaction blob and converts malformed
// codec panics into ordinary errors at public client/hash boundaries.
func DecodeTransactionBlob(txBlob string) (map[string]any, error) {
	return decodeTransactionBlob(txBlob, binarycodec.Decode)
}

func decodeTransactionBlob(
	txBlob string,
	decode func(string) (map[string]any, error),
) (tx map[string]any, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			tx = nil
			if recoveredErr, ok := recovered.(error); ok {
				err = fmt.Errorf("decode transaction blob: %w", recoveredErr)
				return
			}
			err = fmt.Errorf("decode transaction blob: %v", recovered)
		}
	}()
	return decode(txBlob)
}

// CloneTransaction recursively copies the map and JSON-like nested maps/slices
// used by flattened transactions so autofill and submission can safely mutate a
// working value without changing caller-owned data.
func CloneTransaction(tx map[string]any) map[string]any {
	if tx == nil {
		return nil
	}
	cloned := make(map[string]any, len(tx))
	for key, value := range tx {
		cloned[key] = cloneTransactionValue(value)
	}
	return cloned
}

// ReplaceTransactionContents publishes a completed working copy only after validation
// succeeds while preserving the destination map identity for existing caller aliases.
func ReplaceTransactionContents(destination, source map[string]any) {
	clear(destination)
	maps.Copy(destination, source)
}

func cloneTransactionValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return CloneTransaction(typed)
	case []any:
		cloned := make([]any, len(typed))
		for i := range typed {
			cloned[i] = cloneTransactionValue(typed[i])
		}
		return cloned
	case []string:
		cloned := make([]string, len(typed))
		copy(cloned, typed)
		return cloned
	case [][]any:
		cloned := make([][]any, len(typed))
		for i := range typed {
			cloned[i] = cloneTransactionValue(typed[i]).([]any)
		}
		return cloned
	case []map[string]any:
		cloned := make([]map[string]any, len(typed))
		for i := range typed {
			cloned[i] = CloneTransaction(typed[i])
		}
		return cloned
	default:
		return value
	}
}

// NormalizeDeliverMax converts the API v2 Payment alias DeliverMax to the
// wire-compatible Amount field and removes DeliverMax. It returns a shared
// error and leaves tx unchanged when both fields have different values.
func NormalizeDeliverMax(tx map[string]any) error {
	deliverMax, present := tx["DeliverMax"]
	if !present {
		return nil
	}
	if amount, hasAmount := tx["Amount"]; hasAmount {
		if !reflect.DeepEqual(amount, deliverMax) {
			return ErrAmountAndDeliverMaxMustBeIdentical
		}
	} else {
		tx["Amount"] = deliverMax
	}
	delete(tx, "DeliverMax")
	return nil
}

// SubmissionFailHard forces fail_hard for AccountDelete, as recommended by
// XRPL guidance to reduce the risk of paying its high transaction cost on failure.
func SubmissionFailHard(tx map[string]any, requested bool) bool {
	return requested || tx["TransactionType"] == "AccountDelete"
}
