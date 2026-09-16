package builder

import (
	"fmt"
	"reflect"

	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// batchInnerBounds are the inner counts rippled accepts in a Batch.
const (
	minBatchOperations = 2
	maxBatchOperations = 8
)

// BatchOperation is one ordered inner of a confidential Batch. ConvertOp, ConvertBackOp, SendOp,
// MergeInboxOp, ClawbackOp, and TransactionOp implement it, each as a value or as a non-nil
// pointer.
type BatchOperation interface {
	// batchStep validates the operation's own inputs and normalizes it into the step the
	// assembler builds. It seals the interface to this package.
	batchStep() (batchStep, error)
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

// BuildBatch assembles an ordered Batch of confidential MPT operations. Every operation's inputs
// are validated before any ledger access. Every inner's nonce is then resolved, and every
// MPToken and issuance is read from one validated ledger, before any proof is generated. The
// state each inner leaves behind is predicted and fed to the next, including the canonical
// encrypted zero a merge, a clawback, or a first-time convert writes, so later proofs bind
// balances and versions the ledger will actually hold when the Batch applies. Inners are shaped
// for XLS-56 and left unsigned; Fee and LastLedgerSequence are left to the client's autofill.
//
// Only tfAllOrNothing is supported.
func BuildBatch(q LedgerQuerier, p BuildBatchParams) (*transaction.Batch, error) {
	steps, err := planBatch(p)
	if err != nil {
		return nil, err
	}

	resolved, snapshot, err := resolveTxOptions(q, p.Account, p.TxOptions, transaction.BatchTx)
	if err != nil {
		return nil, err
	}
	p.TxOptions = resolved

	nonces, err := planBatchNonces(q, p.Account, p.TxOptions, steps)
	if err != nil {
		return nil, err
	}
	state, err := loadBatchState(snapshot, steps)
	if err != nil {
		return nil, err
	}

	rawTransactions := make([]types.RawTransaction, 0, len(steps))
	for index, step := range steps {
		inner, err := step.build(state, nonces[index])
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

// planBatch rejects everything the assembler can decide from the inputs alone and normalizes
// each operation into its step, so no invalid operation costs a ledger read or a proof.
func planBatch(p BuildBatchParams) ([]batchStep, error) {
	if p.Account == "" {
		return nil, ErrMissingAccount
	}
	if _, err := decodeBuilderAddress(p.Account); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidAccount, err)
	}
	if count := len(p.Operations); count < minBatchOperations || count > maxBatchOperations {
		return nil, fmt.Errorf("%w: a Batch holds between %d and %d inner transactions, got %d",
			ErrBatchOperationCount, minBatchOperations, maxBatchOperations, count)
	}
	if flags := batchFlags(p.Flags); flags != transaction.TfAllOrNothing {
		return nil, fmt.Errorf("%w: only tfAllOrNothing (%#x) is supported, got %#x",
			ErrBatchModeNotSupported, transaction.TfAllOrNothing, flags)
	}

	steps := make([]batchStep, 0, len(p.Operations))
	for index, operation := range p.Operations {
		step, err := planOperation(operation)
		if err != nil {
			return nil, fmt.Errorf("operation %d: %w", index, err)
		}
		steps = append(steps, step)
	}
	return steps, nil
}

// planOperation normalizes one operation, rejecting a nil operation in either form before its
// value-receiver method could dereference a nil pointer.
func planOperation(operation BatchOperation) (batchStep, error) {
	if isNilValue(operation) {
		return batchStep{}, ErrBatchMissingOperation
	}
	return operation.batchStep()
}

// batchStep is one validated operation, reduced to what nonce planning, state loading, and
// inner assembly need, so none of them switches on the operation type again.
type batchStep struct {
	// account is the account whose nonce the inner spends, as the caller spelled it.
	account string
	// options carries the nonce the caller supplied, if any, and the inner's Delegate.
	options TxOptions
	// ticketCount is the number of Tickets a TicketCreate inner creates, each of which moves
	// its account's sequence past the one the inner itself spends.
	ticketCount uint32
	// issuanceID is the issuance a confidential operation acts on, empty for a TransactionOp.
	issuanceID string
	// tokens are the MPTokens a confidential operation reads or writes, with what its
	// transactor requires of each.
	tokens []tokenReference
	// needsIssuerKey is set where the inner encrypts to the issuer key, which is every
	// confidential operation but the inbox merge.
	needsIssuerKey bool
	// build assembles the inner against the predicted state and advances that state.
	build func(state *batchState, nonce TxOptions) (transaction.FlatTransaction, error)
}

// tokenReference names one MPToken an operation touches.
type tokenReference struct {
	holder string
	access mptokenAccess
}

// confidentialStep completes a confidential operation's step once its own validator has run:
// it applies the TxOptions contract the standalone builder applies, and rejects a caller-set
// sequence, which the assembler derives from the operation's position instead.
func confidentialStep(
	account string,
	options TxOptions,
	txType transaction.TxType,
	issuanceID string,
	tokens []tokenReference,
	prepare func(state *batchState, nonce TxOptions) (BatchInnerTransaction, error),
) (batchStep, error) {
	if err := options.validate(account, txType); err != nil {
		return batchStep{}, err
	}
	if options.Sequence != 0 {
		return batchStep{}, ErrBatchInnerSequenceSet
	}
	return batchStep{
		account:        account,
		options:        options,
		issuanceID:     issuanceID,
		tokens:         tokens,
		needsIssuerKey: txType != transaction.ConfidentialMPTMergeInboxTx,
		build: func(state *batchState, nonce TxOptions) (transaction.FlatTransaction, error) {
			tx, err := prepare(state, nonce)
			if err != nil {
				return nil, err
			}
			return shapeBatchInner(tx.Flatten()), nil
		},
	}, nil
}

func (op ConvertOp) batchStep() (batchStep, error) {
	if err := validateConvertBase(op.BuildConvertParams); err != nil {
		return batchStep{}, err
	}
	return confidentialStep(op.Account, op.TxOptions, transaction.ConfidentialMPTConvertTx, op.IssuanceID,
		[]tokenReference{{holder: op.Account, access: holderAccess}},
		func(state *batchState, nonce TxOptions) (BatchInnerTransaction, error) {
			return prepareBatchConvert(state, op, nonce)
		})
}

func (op ConvertBackOp) batchStep() (batchStep, error) {
	if err := validateConvertBackBase(op.BuildConvertBackParams); err != nil {
		return batchStep{}, err
	}
	if err := op.BalanceRange.Validate(); err != nil {
		return batchStep{}, err
	}
	return confidentialStep(op.Account, op.TxOptions, transaction.ConfidentialMPTConvertBackTx, op.IssuanceID,
		[]tokenReference{{holder: op.Account, access: spenderAccess}},
		func(state *batchState, nonce TxOptions) (BatchInnerTransaction, error) {
			return prepareBatchConvertBack(state, op, nonce)
		})
}

func (op SendOp) batchStep() (batchStep, error) {
	if err := validateSendBase(op.BuildSendParams); err != nil {
		return batchStep{}, err
	}
	if err := op.BalanceRange.Validate(); err != nil {
		return batchStep{}, err
	}
	return confidentialStep(op.Account, op.TxOptions, transaction.ConfidentialMPTSendTx, op.IssuanceID,
		[]tokenReference{{holder: op.Account, access: spenderAccess}, {holder: op.Destination, access: holderAccess}},
		func(state *batchState, nonce TxOptions) (BatchInnerTransaction, error) {
			return prepareBatchSend(state, op, nonce)
		})
}

func (op MergeInboxOp) batchStep() (batchStep, error) {
	if err := validateMergeInboxBase(op.BuildMergeInboxParams); err != nil {
		return batchStep{}, err
	}
	return confidentialStep(op.Account, op.TxOptions, transaction.ConfidentialMPTMergeInboxTx, op.IssuanceID,
		[]tokenReference{{holder: op.Account, access: holderAccess}},
		func(state *batchState, nonce TxOptions) (BatchInnerTransaction, error) {
			return prepareBatchMergeInbox(state, op, nonce)
		})
}

func (op ClawbackOp) batchStep() (batchStep, error) {
	if err := validateClawbackBase(op.BuildClawbackParams); err != nil {
		return batchStep{}, err
	}
	if err := op.BalanceRange.Validate(); err != nil {
		return batchStep{}, err
	}
	return confidentialStep(op.Account, op.TxOptions, transaction.ConfidentialMPTClawbackTx, op.IssuanceID,
		[]tokenReference{{holder: op.Holder, access: clawbackAccess}},
		func(state *batchState, nonce TxOptions) (BatchInnerTransaction, error) {
			return prepareBatchClawback(state, op, nonce)
		})
}

// isNilValue reports a nil interface or a typed nil pointer.
func isNilValue(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	return v.Kind() == reflect.Pointer && v.IsNil()
}

// nonceKey is one sequence or Ticket number of one account.
type nonceKey struct {
	account string
	value   uint32
}

// batchNoncePlan tracks the nonces of one Batch while they are resolved in apply order.
type batchNoncePlan struct {
	q LedgerQuerier
	// next is the sequence each account spends next, loaded from the open ledger on first use.
	next map[string]uint32
	// spent is every sequence and Ticket already taken. rippled rejects an all-or-nothing Batch
	// whose inners repeat a number for one account, whether as a sequence or as a Ticket.
	spent map[nonceKey]struct{}
}

// planBatchNonces resolves the nonce of every inner before any proof binds one. Each account's
// inners spend consecutive sequences in apply order: the outer account's start one past the
// sequence the Batch spends, and every other account's at its current sequence. A Ticket spends
// no sequence, but each number, including the outer Batch's own Ticket, is spent once, and a
// TicketCreate moves its account's sequence past every Ticket it creates.
func planBatchNonces(q LedgerQuerier, account string, outer TxOptions, steps []batchStep) ([]TxOptions, error) {
	batchAccount, err := decodeBuilderAddress(account)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidAccount, err)
	}

	plan := &batchNoncePlan{
		q:     q,
		next:  make(map[string]uint32, len(steps)+1),
		spent: make(map[nonceKey]struct{}, len(steps)+1),
	}
	if outer.Sequence != 0 {
		plan.next[batchAccount.Classic] = outer.Sequence + 1
	}
	if outer.TicketSequence != 0 {
		plan.spent[nonceKey{account: batchAccount.Classic, value: outer.TicketSequence}] = struct{}{}
	}

	nonces := make([]TxOptions, 0, len(steps))
	for index, step := range steps {
		nonce, err := plan.resolve(step)
		if err != nil {
			return nil, fmt.Errorf("operation %d: %w", index, err)
		}
		nonces = append(nonces, nonce)
	}
	return nonces, nil
}

// resolve assigns one inner its nonce and records what the inner consumes.
func (p *batchNoncePlan) resolve(step batchStep) (TxOptions, error) {
	decoded, err := decodeBuilderAddress(step.account)
	if err != nil {
		return TxOptions{}, fmt.Errorf("%w: %w", ErrInvalidAccount, err)
	}
	account := decoded.Classic

	nonce := step.options
	value := nonce.TicketSequence
	if nonce.TicketSequence == 0 {
		next, err := p.sequence(account)
		if err != nil {
			return TxOptions{}, err
		}
		if nonce.Sequence != 0 && nonce.Sequence != next {
			return TxOptions{}, fmt.Errorf("%w: got %d, position requires %d", ErrBatchInnerSequenceMismatch, nonce.Sequence, next)
		}
		nonce.Sequence = next
		p.next[account] = next + 1
		value = next
	}

	key := nonceKey{account: account, value: value}
	if _, spent := p.spent[key]; spent {
		return TxOptions{}, fmt.Errorf("%w: %d for %s", ErrBatchDuplicateNonce, value, account)
	}
	p.spent[key] = struct{}{}

	if step.ticketCount != 0 {
		next, err := p.sequence(account)
		if err != nil {
			return TxOptions{}, err
		}
		p.next[account] = next + step.ticketCount
	}
	return nonce, nil
}

// sequence reports the sequence an account spends next, reading it on first use.
func (p *batchNoncePlan) sequence(account string) (uint32, error) {
	if next, loaded := p.next[account]; loaded {
		return next, nil
	}
	next, err := currentAccountSequence(p.q, account)
	if err != nil {
		return 0, err
	}
	p.next[account] = next
	return next, nil
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

// loadBatchState reads the initial state of every issuance and MPToken the steps reference, all
// from one validated ledger, holding each MPToken to the combined requirements of every
// operation that touches it.
func loadBatchState(snapshot *ledgerSnapshot, steps []batchStep) (*batchState, error) {
	state := &batchState{
		tokens:    make(map[tokenKey]*tokenState, len(steps)),
		issuances: make(map[string]*batchIssuance, len(steps)),
	}

	type reference struct {
		key    tokenKey
		holder string
	}
	provable := make(map[string]bool)
	issuanceIDs := make([]string, 0, len(steps))
	access := make(map[tokenKey]mptokenAccess)
	references := make([]reference, 0, len(steps)+1)
	for _, step := range steps {
		if step.issuanceID == "" {
			continue
		}
		issuanceID := normalizeIssuanceID(step.issuanceID)
		if _, seen := provable[issuanceID]; !seen {
			issuanceIDs = append(issuanceIDs, issuanceID)
		}
		provable[issuanceID] = provable[issuanceID] || step.needsIssuerKey

		for _, token := range step.tokens {
			key, err := newTokenKey(token.holder, step.issuanceID)
			if err != nil {
				return nil, err
			}
			if _, seen := access[key]; !seen {
				references = append(references, reference{key: key, holder: token.holder})
			}
			access[key] = access[key].merge(token.access)
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
		token, err := readBatchTokenState(snapshot, issuance.issuanceState, ref.key, ref.holder, access[ref.key])
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

// shapeBatchInner shapes a flattened transaction as an XLS-56 inner.
func shapeBatchInner(flat transaction.FlatTransaction) transaction.FlatTransaction {
	flags, _ := flat["Flags"].(uint32)
	flat["Flags"] = flags | types.TfInnerBatchTxn
	flat["Fee"] = "0"
	flat["SigningPubKey"] = ""
	return flat
}
