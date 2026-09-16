package builder

import (
	"fmt"
	"strings"

	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// batchInnerBounds are the inner counts rippled accepts in a Batch.
const (
	minBatchOperations = 2
	maxBatchOperations = 8
)

// BatchOperation is one ordered inner of a confidential Batch.
type BatchOperation interface {
	// batchOperation seals the interface to this package.
	batchOperation()
}

// ConvertOp moves public MPT into a confidential balance, as BuildConvert does.
type ConvertOp struct {
	BuildConvertParams
}

// ConvertBackOp reveals part of a confidential balance as public MPT, as BuildConvertBack does.
type ConvertBackOp struct {
	BuildConvertBackParams
}

// SendOp transfers a confidential amount to another holder, as BuildSend does.
type SendOp struct {
	BuildSendParams
}

// MergeInboxOp folds a holder's inbox into its spending balance, as BuildMergeInbox does.
type MergeInboxOp struct {
	BuildMergeInboxParams
}

// ClawbackOp burns a holder's entire confidential balance, as BuildClawback does.
type ClawbackOp struct {
	BuildClawbackParams
}

// TransactionOp carries a ready-made ordinary transaction as an inner.
type TransactionOp struct {
	Tx BatchInnerTransaction
}

// BatchInnerTransaction is the contract a ready-made inner must meet.
type BatchInnerTransaction interface {
	TxType() transaction.TxType
	Flatten() transaction.FlatTransaction
	Validate() (bool, error)
}

func (ConvertOp) batchOperation()     {}
func (ConvertBackOp) batchOperation() {}
func (SendOp) batchOperation()        {}
func (MergeInboxOp) batchOperation()  {}
func (ClawbackOp) batchOperation()    {}
func (TransactionOp) batchOperation() {}

// BuildBatchParams holds the inputs for BuildBatch.
type BuildBatchParams struct {
	// TxOptions is the outer Batch's nonce. Delegate is rejected: Batch is not delegatable.
	TxOptions
	// Account owns, pays for, and signs the outer Batch.
	Account string
	// Operations are the inners in ledger apply order, between two and eight.
	Operations []BatchOperation
	// Flags is the outer mode. Zero means transaction.TfAllOrNothing, the only supported mode.
	Flags uint32
}

// BuildBatch assembles an ordered Batch of confidential MPT operations. Every MPToken and
// issuance is read from one validated ledger, and the state each inner leaves behind is
// predicted and fed to the next, so later proofs bind balances and versions the ledger will
// actually hold when the Batch applies. Inners are shaped for XLS-56 and left unsigned; Fee
// and LastLedgerSequence are left to the client's autofill.
//
// Only tfAllOrNothing is supported. An inner that reads a balance an earlier merge, clawback,
// or first-time convert reset to the canonical encrypted zero reports
// ErrBatchUnpredictableState, since that ciphertext cannot be reproduced client-side.
func BuildBatch(q LedgerQuerier, p BuildBatchParams) (*transaction.Batch, error) {
	if err := validateBatchParams(p); err != nil {
		return nil, err
	}

	resolved, snapshot, err := resolveTxOptions(q, p.Account, p.TxOptions, transaction.BatchTx)
	if err != nil {
		return nil, err
	}
	p.TxOptions = resolved

	nonces, err := loadBatchSequences(q, p)
	if err != nil {
		return nil, err
	}
	state, err := loadBatchState(snapshot, p.Operations)
	if err != nil {
		return nil, err
	}

	rawTransactions := make([]types.RawTransaction, 0, len(p.Operations))
	for index, operation := range p.Operations {
		inner, err := buildBatchInner(state, nonces, operation)
		if err != nil {
			return nil, fmt.Errorf("operation %d: %w", index, err)
		}
		rawTransactions = append(rawTransactions, types.RawTransaction{RawTransaction: inner})
	}

	batch := &transaction.Batch{
		BaseTx:          baseTx(p.Account, transaction.BatchTx, p.TxOptions),
		RawTransactions: rawTransactions,
	}
	batch.Flags = batchFlags(p.Flags)

	if err := validatePreparedTransaction(batch); err != nil {
		return nil, err
	}
	return batch, nil
}

// batchFlags resolves the outer mode, defaulting an unset value to all-or-nothing.
func batchFlags(flags uint32) uint32 {
	if flags == 0 {
		return transaction.TfAllOrNothing
	}
	return flags
}

// validateBatchParams rejects what the assembler can decide before any ledger access.
func validateBatchParams(p BuildBatchParams) error {
	if p.Account == "" {
		return ErrMissingAccount
	}
	if _, err := decodeBuilderAddress(p.Account); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidAccount, err)
	}
	if count := len(p.Operations); count < minBatchOperations || count > maxBatchOperations {
		return fmt.Errorf("%w: a Batch holds between %d and %d inner transactions, got %d",
			ErrBatchOperationCount, minBatchOperations, maxBatchOperations, count)
	}
	if flags := batchFlags(p.Flags); flags != transaction.TfAllOrNothing {
		return fmt.Errorf("%w: only tfAllOrNothing (%#x) is supported, got %#x",
			ErrBatchModeNotSupported, transaction.TfAllOrNothing, flags)
	}

	for index, operation := range p.Operations {
		if err := validateBatchOperation(operation); err != nil {
			return fmt.Errorf("operation %d: %w", index, err)
		}
	}
	return nil
}

// validateBatchOperation rejects an operation the assembler cannot own the nonce of, and a ready-
// made inner whose effect it cannot predict.
func validateBatchOperation(operation BatchOperation) error {
	if operation == nil {
		return ErrBatchMissingOperation
	}
	if plain, ok := operation.(TransactionOp); ok {
		return validatePlainInner(plain)
	}
	options, err := operationOptions(operation)
	if err != nil {
		return err
	}
	if options.Sequence != 0 {
		return ErrBatchInnerSequenceSet
	}
	return nil
}

// operationAccount reports the account whose nonce an operation spends.
func operationAccount(operation BatchOperation) (string, error) {
	switch op := operation.(type) {
	case ConvertOp:
		return op.Account, nil
	case ConvertBackOp:
		return op.Account, nil
	case SendOp:
		return op.Account, nil
	case MergeInboxOp:
		return op.Account, nil
	case ClawbackOp:
		return op.Account, nil
	case TransactionOp:
		return plainInnerAccount(op)
	default:
		return "", fmt.Errorf("%w: %T", ErrBatchInnerNotSupported, operation)
	}
}

// operationOptions reports the nonce options a confidential operation carries.
func operationOptions(operation BatchOperation) (TxOptions, error) {
	switch op := operation.(type) {
	case ConvertOp:
		return op.TxOptions, nil
	case ConvertBackOp:
		return op.TxOptions, nil
	case SendOp:
		return op.TxOptions, nil
	case MergeInboxOp:
		return op.TxOptions, nil
	case ClawbackOp:
		return op.TxOptions, nil
	default:
		return TxOptions{}, fmt.Errorf("%w: %T", ErrBatchInnerNotSupported, operation)
	}
}

// operationCarriesNonce reports whether an inner already has a nonce of its own, and so takes no
// sequence from its account's counter.
func operationCarriesNonce(operation BatchOperation) (bool, error) {
	if plain, ok := operation.(TransactionOp); ok {
		if isNilTx(plain.Tx) {
			return false, ErrBatchMissingOperation
		}
		return plainInnerHasNonce(plain.Tx.Flatten()), nil
	}
	options, err := operationOptions(operation)
	if err != nil {
		return false, err
	}
	return options.TicketSequence != 0, nil
}

// batchNonces allocates each inner its sequence.
type batchNonces struct {
	next map[string]uint32
}

// allocate consumes an account's next sequence, or reports that the inner spends a Ticket and
// needs none.
func (n *batchNonces) allocate(account string, options TxOptions) (uint32, error) {
	if options.TicketSequence != 0 {
		return 0, nil
	}
	decoded, err := decodeBuilderAddress(account)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrInvalidAccount, err)
	}
	sequence, ok := n.next[decoded.Classic]
	if !ok {
		return 0, fmt.Errorf("%w: no sequence resolved for %s", ErrInvalidLedgerState, decoded.Classic)
	}
	n.next[decoded.Classic] = sequence + 1
	return sequence, nil
}

// claim accepts a caller-set sequence only if it is the one the account would be allocated at this
// position, so a ready-made inner cannot collide with or skip past an allocated one.
func (n *batchNonces) claim(account string, sequence uint32) error {
	decoded, err := decodeBuilderAddress(account)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidAccount, err)
	}
	next, tracked := n.next[decoded.Classic]
	if !tracked {
		return nil
	}
	if sequence != next {
		return fmt.Errorf("%w: got %d, position requires %d", ErrBatchInnerSequenceMismatch, sequence, next)
	}
	n.next[decoded.Classic] = sequence + 1
	return nil
}

// loadBatchSequences resolves the first sequence each inner account will spend.
func loadBatchSequences(q LedgerQuerier, p BuildBatchParams) (*batchNonces, error) {
	batchAccount, err := decodeBuilderAddress(p.Account)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidAccount, err)
	}

	nonces := &batchNonces{next: make(map[string]uint32, len(p.Operations)+1)}
	if p.Sequence != 0 {
		nonces.next[batchAccount.Classic] = p.Sequence + 1
	}
	for _, operation := range p.Operations {
		ownNonce, err := operationCarriesNonce(operation)
		if err != nil {
			return nil, err
		}
		if ownNonce {
			continue
		}
		account, err := operationAccount(operation)
		if err != nil {
			return nil, err
		}
		decoded, err := decodeBuilderAddress(account)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidAccount, err)
		}
		if _, loaded := nonces.next[decoded.Classic]; loaded {
			continue
		}

		sequence, err := currentAccountSequence(q, decoded.Classic)
		if err != nil {
			return nil, err
		}
		nonces.next[decoded.Classic] = sequence
	}
	return nonces, nil
}

// currentAccountSequence reads the sequence an account will next spend from the open ledger.
func currentAccountSequence(q LedgerQuerier, classic string) (uint32, error) {
	info, err := getAccountInfo(q, types.Address(classic), common.Current)
	if err != nil {
		return 0, err
	}
	if info.AccountData.Sequence == 0 {
		return 0, fmt.Errorf("%w: account_info reported sequence 0 for %s", ErrInvalidLedgerState, classic)
	}
	return info.AccountData.Sequence, nil
}

// loadBatchState reads the initial state of every issuance and MPToken the operations reference,
// all from one validated ledger.
func loadBatchState(snapshot *ledgerSnapshot, operations []BatchOperation) (*batchState, error) {
	state := &batchState{
		tokens:    make(map[tokenKey]*tokenState, len(operations)),
		issuances: make(map[string]*batchIssuance, len(operations)),
	}

	usable := make(map[tokenKey]bool)
	provable := make(map[string]bool)

	type reference struct {
		key    tokenKey
		holder string
	}
	issuanceIDs := make([]string, 0, len(operations))
	references := make([]reference, 0, len(operations)+1)
	for _, operation := range operations {
		issuanceID, holders, err := operationReferences(operation)
		if err != nil {
			return nil, err
		}
		if issuanceID == "" {
			continue
		}
		normalized := strings.ToUpper(issuanceID)
		if _, seen := provable[normalized]; !seen {
			provable[normalized] = false
			issuanceIDs = append(issuanceIDs, normalized)
		}
		if _, isMerge := operation.(MergeInboxOp); !isMerge {
			provable[normalized] = true
		}

		_, isClawback := operation.(ClawbackOp)
		for _, holder := range holders {
			key, err := newTokenKey(holder, issuanceID)
			if err != nil {
				return nil, err
			}
			if _, seen := usable[key]; !seen {
				references = append(references, reference{key: key, holder: holder})
			}
			usable[key] = usable[key] || !isClawback
		}
	}

	for _, issuanceID := range issuanceIDs {
		issuance, err := readBatchIssuance(snapshot, issuanceID, provable[issuanceID])
		if err != nil {
			return nil, fmt.Errorf("issuance %s: %w", issuanceID, err)
		}
		state.issuances[issuanceID] = issuance
	}
	for _, ref := range references {
		issuance, err := state.issuance(ref.key.issuance)
		if err != nil {
			return nil, err
		}
		token, err := readBatchTokenState(snapshot, issuance.issuanceState, ref.key, ref.holder, usable[ref.key])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", ref.key, err)
		}
		state.tokens[ref.key] = token
	}
	return state, nil
}

// readBatchIssuance reads one issuance and seeds its running confidential supply.
func readBatchIssuance(snapshot *ledgerSnapshot, issuanceID string, needsIssuerKey bool) (*batchIssuance, error) {
	read := readIssuance
	if needsIssuerKey {
		read = getProvableIssuance
	}
	issuance, err := read(snapshot, issuanceID)
	if err != nil {
		return nil, err
	}
	return &batchIssuance{issuanceState: issuance, outstanding: issuance.confidentialOutstanding}, nil
}

// operationReferences reports the issuance an operation acts on and the holders whose MPToken
// state it reads or writes.
func operationReferences(operation BatchOperation) (string, []string, error) {
	switch op := operation.(type) {
	case ConvertOp:
		return op.IssuanceID, []string{op.Account}, nil
	case ConvertBackOp:
		return op.IssuanceID, []string{op.Account}, nil
	case SendOp:
		return op.IssuanceID, []string{op.Account, op.Destination}, nil
	case MergeInboxOp:
		return op.IssuanceID, []string{op.Account}, nil
	case ClawbackOp:
		return op.IssuanceID, []string{op.Holder}, nil
	case TransactionOp:
		return "", nil, nil
	default:
		return "", nil, fmt.Errorf("%w: %T", ErrBatchInnerNotSupported, operation)
	}
}

// buildBatchInner builds one inner against the current predictions, advances them by what that
// inner does, and returns the inner shaped for a Batch.
func buildBatchInner(state *batchState, nonces *batchNonces, operation BatchOperation) (transaction.FlatTransaction, error) {
	if plain, ok := operation.(TransactionOp); ok {
		return buildPlainInner(nonces, plain)
	}

	account, err := operationAccount(operation)
	if err != nil {
		return nil, err
	}
	options, err := operationOptions(operation)
	if err != nil {
		return nil, err
	}
	sequence, err := nonces.allocate(account, options)
	if err != nil {
		return nil, err
	}
	options.Sequence = sequence

	tx, err := prepareBatchOperation(state, operation, options)
	if err != nil {
		return nil, err
	}
	return shapeBatchInner(tx), nil
}

// prepareBatchOperation dispatches to the Prepare helper of the operation's own builder, feeding
// it the predicted state in place of the ledger reads the Build helper would make, and then
// advances the predictions.
func prepareBatchOperation(state *batchState, operation BatchOperation, options TxOptions) (BatchInnerTransaction, error) {
	switch op := operation.(type) {
	case ConvertOp:
		return prepareBatchConvert(state, op, options)
	case ConvertBackOp:
		return prepareBatchConvertBack(state, op, options)
	case SendOp:
		return prepareBatchSend(state, op, options)
	case MergeInboxOp:
		return prepareBatchMergeInbox(state, op, options)
	case ClawbackOp:
		return prepareBatchClawback(state, op, options)
	default:
		return nil, fmt.Errorf("%w: %T", ErrBatchInnerNotSupported, operation)
	}
}

// shapeBatchInner shapes a built transaction as an XLS-56 inner.
func shapeBatchInner(tx BatchInnerTransaction) transaction.FlatTransaction {
	flat := tx.Flatten()
	flags, _ := flat["Flags"].(uint32)
	flat["Flags"] = flags | types.TfInnerBatchTxn
	flat["Fee"] = "0"
	flat["SigningPubKey"] = ""
	return flat
}
