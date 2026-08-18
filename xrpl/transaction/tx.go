package transaction

import (
	"bytes"
	"errors"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// Tx defines the interface for all transaction types requiring a TxType method.
// TODO: Refactor to use a single interface for all transaction types
type Tx interface {
	TxType() TxType
}

// TxHash represents a transaction hash reference.
type TxHash string

// TxType returns the TxType for a hashed transaction.
func (*TxHash) TxType() TxType {
	return HashedTx
}

// Binary represents a raw transaction blob for submission.
type Binary struct {
	TxBlob string `json:"tx_blob"`
}

// TxType returns the TxType for a binary transaction.
func (tx *Binary) TxType() TxType {
	return BinaryTx
}

// BaseTx contains the common fields for all transactions.
type BaseTx struct {
	// The unique address of the transaction sender.
	Account types.Address
	//
	// The type of transaction. Valid types include: `Payment`, `OfferCreate`,
	// `TrustSet`, and many others.
	//
	TransactionType TxType
	//
	// Integer amount of XRP, in drops, to be destroyed as a cost for
	// distributing this transaction to the network. Some transaction types have
	// different minimum requirements.
	//
	Fee types.XRPCurrencyAmount `json:",omitempty"`
	//
	// The sequence number of the account sending the transaction. A transaction
	// is only valid if the Sequence number is exactly 1 greater than the previous
	// transaction from the same account. The special case 0 means the transaction
	// is using a Ticket instead.
	//
	Sequence uint32 `json:",omitempty"`
	//
	// Hash value identifying another transaction. If provided, this transaction
	// is only valid if the sending account's previously-sent transaction matches
	// the provided hash.
	//
	AccountTxnID types.Hash256 `json:",omitempty"`
	//
	// The delegate account that is sending the transaction.
	//
	Delegate types.Address `json:",omitempty"`
	// Set of bit-flags for this transaction.
	Flags uint32 `json:",omitempty"`
	//
	// Highest ledger index this transaction can appear in. Specifying this field
	// places a strict upper limit on how long the transaction can wait to be
	// validated or rejected.
	//
	LastLedgerSequence uint32 `json:",omitempty"`
	//
	// Additional arbitrary information used to identify this transaction.
	//
	Memos []types.MemoWrapper `json:",omitempty"`
	// The network id of the transaction.
	NetworkID uint32 `json:",omitempty"`
	//
	// Array of objects that represent a multi-signature which authorizes this
	// transaction.
	//
	Signers []types.Signer `json:",omitempty"`
	//
	// Arbitrary integer used to identify the reason for this payment, or a sender
	// on whose behalf this transaction is made. Conventionally, a refund should
	// specify the initial payment's SourceTag as the refund payment's
	// DestinationTag.
	//
	SourceTag uint32 `json:",omitempty"`
	//
	// Hex representation of the public key that corresponds to the private key
	// used to sign this transaction. If an empty string, indicates a
	// multi-signature is present in the Signers field instead.
	//
	SigningPubKey string `json:",omitempty"`
	//
	// The sequence number of the ticket to use in place of a Sequence number. If
	// this is provided, Sequence must be 0. Cannot be used with AccountTxnID.
	//
	TicketSequence uint32 `json:",omitempty"`
	//
	// The signature that verifies this transaction as originating from the
	// account it says it is from.
	//
	TxnSignature string `json:",omitempty"`
}

// TxType returns the transaction type stored in BaseTx.
func (tx *BaseTx) TxType() TxType {
	return tx.TransactionType
}

// Flatten converts BaseTx into a FlatTransaction map for JSON-RPC submission.
func (tx *BaseTx) Flatten() FlatTransaction {
	flattened := make(FlatTransaction)

	if tx.Account != "" {
		flattened["Account"] = tx.Account.String()
	}
	if tx.TransactionType != "" {
		flattened["TransactionType"] = tx.TransactionType.String()
	}
	if tx.Fee != 0 {
		flattened["Fee"] = tx.Fee.String()
	}
	// The XRPL protocol requires Sequence to be exactly 0 when using a ticket.
	if tx.Sequence != 0 || tx.TicketSequence != 0 {
		flattened["Sequence"] = tx.Sequence
	}
	if tx.AccountTxnID != "" {
		flattened["AccountTxnID"] = tx.AccountTxnID.String()
	}
	if tx.Flags != 0 {
		flattened["Flags"] = tx.Flags
	}
	if tx.LastLedgerSequence != 0 {
		flattened["LastLedgerSequence"] = tx.LastLedgerSequence
	}
	if len(tx.Memos) > 0 {
		flattenedMemos := make([]any, 0)
		for _, memo := range tx.Memos {
			flattenedMemo := memo.Flatten()
			if flattenedMemo != nil {
				flattenedMemos = append(flattenedMemos, flattenedMemo)
			}
		}
		flattened["Memos"] = flattenedMemos
	}
	if tx.NetworkID != 0 {
		flattened["NetworkID"] = tx.NetworkID
	}
	if len(tx.Signers) > 0 {
		flattenedSigners := make([]any, len(tx.Signers))
		for i, signer := range tx.Signers {
			flattenedSigners[i] = signer.Flatten()
		}
		flattened["Signers"] = flattenedSigners
	}
	if tx.SourceTag != 0 {
		flattened["SourceTag"] = tx.SourceTag
	}
	// Inner Batch transactions are unsigned but must carry an explicit empty
	// SigningPubKey on the wire.
	if tx.SigningPubKey != "" || tx.Flags&types.TfInnerBatchTxn != 0 {
		flattened["SigningPubKey"] = tx.SigningPubKey
	}
	if tx.TicketSequence != 0 {
		flattened["TicketSequence"] = tx.TicketSequence
	}
	if tx.TxnSignature != "" {
		flattened["TxnSignature"] = tx.TxnSignature
	}
	if tx.Delegate != "" {
		flattened["Delegate"] = tx.Delegate.String()
	}

	return flattened
}

// Validate checks BaseTx fields for correctness, returning false and an error if invalid.
func (tx *BaseTx) Validate() (bool, error) {
	flattenTx := tx.Flatten()

	// Account, Delegate, and each Signers entry are the fields checked against ACCOUNT_ZERO,
	// because they are the ones BaseTx carries that must name a signer. The value is
	// legitimate elsewhere, such as an issuer standing for XRP, so the check is kept off any
	// field that merely identifies an asset or a consensus-generated account. The Signers
	// entries are checked in IsSigner, reached through validateSigners below.
	//
	// An Account pairing a tagged X-address with an explicit SourceTag is not checked here.
	// Client autofill rewrites the address to its classic form and carries the tag across,
	// so the pairing reaches the ledger intact, and rejecting it here would refuse input
	// that works today. The transaction types that resolve the tag themselves check it.
	accountID, _, err := decodeAddressAccountID(tx.Account)
	if err != nil {
		return false, ErrInvalidAccount
	}
	// A pseudo-transaction is generated by consensus rather than signed, and the binary
	// codec requires its Account to be ACCOUNT_ZERO, so the rule below would reject the
	// only value the encoder accepts.
	if !IsPseudoTransactionType(tx.TransactionType) && addresscodec.IsZeroAccountID(accountID) {
		return false, ErrAccountZero
	}

	if tx.TransactionType == "" {
		return false, ErrInvalidTransactionType
	}

	if !typecheck.IsStringNumericUint(tx.Fee.String(), 10, 64) {
		return false, errors.New("invalid fee amount, not a uint")
	}

	err = ValidateOptionalField(flattenTx, "Sequence", typecheck.IsUint32)
	if err != nil {
		return false, err
	}

	err = ValidateOptionalField(flattenTx, "AccountTxnID", typecheck.IsString)
	if err != nil {
		return false, err
	}

	err = ValidateOptionalField(flattenTx, "LastLedgerSequence", typecheck.IsUint32)
	if err != nil {
		return false, err
	}

	err = ValidateOptionalField(flattenTx, "SourceTag", typecheck.IsUint32)
	if err != nil {
		return false, err
	}

	err = ValidateOptionalField(flattenTx, "SigningPubKey", typecheck.IsString)
	if err != nil {
		return false, err
	}

	err = ValidateOptionalField(flattenTx, "TicketSequence", typecheck.IsUint32)
	if err != nil {
		return false, err
	}

	err = ValidateOptionalField(flattenTx, "TxnSignature", typecheck.IsString)
	if err != nil {
		return false, err
	}

	err = ValidateOptionalField(flattenTx, "NetworkID", typecheck.IsUint32)
	if err != nil {
		return false, err
	}

	// Validate Delegate field.
	if tx.Delegate != "" {
		delegateID, hasTag, delegateErr := decodeAddressAccountID(tx.Delegate)
		if delegateErr != nil {
			return false, ErrInvalidDelegate
		}
		if addresscodec.IsZeroAccountID(delegateID) {
			return false, ErrDelegateZero
		}
		// Delegate has no companion tag field, so a tagged X-address cannot be used.
		if hasTag {
			return false, ErrDelegateTagNotAllowed
		}
		// Delegate and Account cannot be the same account, in any address form.
		if bytes.Equal(delegateID, accountID) {
			return false, ErrDelegateAccountConflict
		}
	}

	// memos
	err = validateMemos(tx.Memos)
	if err != nil {
		return false, err
	}

	// signers
	err = validateSigners(tx.Signers)
	if err != nil {
		return false, err
	}

	return true, nil
}
