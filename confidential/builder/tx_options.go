package builder

import (
	"errors"
	"fmt"

	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// TxOptions holds what every confidential builder accepts beyond the operation's own
// inputs: the nonce that authorizes the transaction, and the account submitting it.
type TxOptions struct {
	// Sequence is the account sequence the transaction spends. Left zero, a Build* helper
	// reads it from the ledger, and a Prepare* helper that binds no proof leaves it for a
	// later autofill.
	Sequence uint32
	// TicketSequence spends a Ticket in place of the account sequence. XRPL requires
	// Sequence to be zero whenever it is set, so the two are mutually exclusive. Every
	// builder accepts one, and a proof commits to it rather than to an account sequence,
	// because xrpld hashes the sequence proxy. See the package documentation for the one
	// case a Ticket cannot rescue: two transactions of the same holder on one issuance whose
	// proofs both bind that holder's ConfidentialBalanceVersion.
	TicketSequence uint32
	// Delegate is the account submitting on the transaction account's behalf. A builder
	// whose transaction type transaction.NonDelegatableTransactionsMap lists, as it lists
	// ConfidentialMPTConvert, rejects a non-empty value.
	Delegate string
}

// validate rejects an option combination a transaction of txType cannot carry. The delegate
// address conditions and their sentinels are the ones BaseTx.Validate reports, so a builder
// and a preflight fail as one condition, except that these reach the caller before any
// ledger query or proof work. The nonce rules and the delegability of txType have no
// preflight counterpart, so the builder is the only place they are enforced.
func (o TxOptions) validate(account string, txType transaction.TxType) error {
	if o.Sequence != 0 && o.TicketSequence != 0 {
		return ErrConflictingNonce
	}
	if o.Delegate == "" {
		return nil
	}
	if _, nonDelegatable := transaction.NonDelegatableTransactionsMap[txType.String()]; nonDelegatable {
		return ErrDelegateNotAllowed
	}

	delegate, err := decodeBuilderAddress(o.Delegate)
	if err != nil {
		if errors.Is(err, transaction.ErrZeroAccountID) {
			return transaction.ErrDelegateZero
		}
		return fmt.Errorf("%w: %w", transaction.ErrInvalidDelegate, err)
	}
	// Delegate has no companion tag field, so a tagged X-address cannot be used.
	if delegate.HasTag {
		return transaction.ErrDelegateTagNotAllowed
	}
	// The base validation every caller runs first already decoded the account, so this
	// repeat cannot fail. It is reported rather than discarded because a zero AccountID
	// matches no delegate, so discarding it would turn the guard below into a silent no-op
	// if that ordering ever changed.
	submitter, err := decodeBuilderAddress(account)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidAccount, err)
	}
	// Delegate and Account cannot be the same account, in any address form.
	if delegate.AccountID == submitter.AccountID {
		return transaction.ErrDelegateAccountConflict
	}
	return nil
}

// proofSequence resolves the nonce a proof context binds. xrpld hashes the sequence proxy,
// which is the ticket sequence whenever the transaction spends a Ticket, so a proof built
// against a ticket must commit to that value rather than to the account sequence.
func (o TxOptions) proofSequence() (uint32, error) {
	if o.Sequence != 0 {
		return o.Sequence, nil
	}
	if o.TicketSequence != 0 {
		return o.TicketSequence, nil
	}
	return 0, ErrMissingSequence
}

// validateForProof validates the options and resolves the nonce every proof-bearing helper
// commits to. A later autofill cannot supply that nonce without invalidating the proof, so it
// is resolved here rather than left to one. PrepareConvert carries a proof only on the
// first-time form, so it resolves its nonce through proofSequence directly.
func (o TxOptions) validateForProof(account string, txType transaction.TxType) (uint32, error) {
	if err := o.validate(account, txType); err != nil {
		return 0, err
	}
	return o.proofSequence()
}

// resolveTxOptions validates the options and fills in the account sequence when the caller
// supplied neither nonce. The snapshot it returns is unbound in the other case, because no
// account query selected a validated ledger and the first entry read selects one instead.
//
// The resolved options are returned rather than written through a pointer, so a caller that
// forgets to keep them cannot reach the proof with the sequence still zero. Every builder must
// assign the result back onto the params it later passes to its Prepare* helper, because that
// helper reads the nonce the proof commits to out of the same struct.
func resolveTxOptions(q LedgerQuerier, account string, options TxOptions, txType transaction.TxType) (TxOptions, *ledgerSnapshot, error) {
	if err := options.validate(account, txType); err != nil {
		return options, nil, err
	}
	if options.Sequence != 0 || options.TicketSequence != 0 {
		return options, &ledgerSnapshot{q: q}, nil
	}

	sequence, snapshot, err := beginBuild(q, account)
	if err != nil {
		return options, nil, err
	}
	options.Sequence = sequence
	return options, snapshot, nil
}

// baseTx assembles the BaseTx fields every confidential builder shares. The account keeps
// the spelling the caller supplied, as the other address fields do.
func baseTx(account string, txType transaction.TxType, options TxOptions) transaction.BaseTx {
	return transaction.BaseTx{
		Account:         types.Address(account),
		TransactionType: txType,
		Sequence:        options.Sequence,
		TicketSequence:  options.TicketSequence,
		Delegate:        types.Address(options.Delegate),
	}
}
