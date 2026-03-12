package wallet

import (
	"maps"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/xrpl"
	"github.com/Peersyst/xrpl-go/xrpl/hash"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
)

// SignLoanSetByCounterpartyOptions configures how the counterparty signs a LoanSet transaction.
type SignLoanSetByCounterpartyOptions struct {
	// Multisign indicates the wallet should sign as a multisig account.
	Multisign bool
	// MultisignAccount is the address to sign as (overrides wallet.ClassicAddress when set).
	MultisignAccount string
}

// SignLoanSetByCounterparty signs a LoanSet transaction as the counterparty/borrower.
// The LoanBroker must have already signed the transaction (TxnSignature and SigningPubKey must be set).
// The result is stored in CounterpartySignature on the transaction map.
func SignLoanSetByCounterparty(
	w Wallet,
	tx *transaction.FlatTransaction,
	opts *SignLoanSetByCounterpartyOptions,
) (txBlob string, txHash string, err error) {
	if (*tx)["TransactionType"] != "LoanSet" {
		return "", "", ErrTxMustBeLoanSet
	}

	if cs, ok := (*tx)["CounterpartySignature"]; ok && cs != nil {
		return "", "", ErrCounterpartyAlreadySigned
	}

	txnSig, _ := (*tx)["TxnSignature"].(string)
	signingPubKey, _ := (*tx)["SigningPubKey"].(string)
	if txnSig == "" || signingPubKey == "" {
		return "", "", ErrBrokerMustSignFirst
	}

	multisign, counterpartyAddr := resolveMultisignOpts(w.ClassicAddress.String(), opts)

	sig, err := encodeAndSign(&w, *tx, multisign, counterpartyAddr)
	if err != nil {
		return "", "", err
	}

	var counterpartySignatureMap map[string]any
	if multisign {
		counterpartySignatureMap = map[string]any{
			"Signers": []any{
				map[string]any{
					"Signer": map[string]any{
						"Account":       counterpartyAddr,
						"SigningPubKey": w.PublicKey,
						"TxnSignature":  sig,
					},
				},
			},
		}
	} else {
		counterpartySignatureMap = map[string]any{
			"SigningPubKey": w.PublicKey,
			"TxnSignature":  sig,
		}
	}
	(*tx)["CounterpartySignature"] = counterpartySignatureMap

	txBlob, err = binarycodec.Encode(*tx)
	if err != nil {
		return "", "", err
	}

	txHash, err = hash.SignTxBlob(txBlob)
	if err != nil {
		return "", "", err
	}

	return txBlob, txHash, nil
}

// SignLoanSetByCounterpartyBlob decodes a hex-encoded transaction blob and signs it as the counterparty.
// This is a convenience wrapper around SignLoanSetByCounterparty for callers that have a serialized blob
// rather than a FlatTransaction.
func SignLoanSetByCounterpartyBlob(
	w Wallet,
	blob string,
	opts *SignLoanSetByCounterpartyOptions,
) (tx transaction.FlatTransaction, txBlob string, txHash string, err error) {
	decoded, err := binarycodec.Decode(blob)
	if err != nil {
		return nil, "", "", err
	}
	tx = transaction.FlatTransaction(decoded)
	txBlob, txHash, err = SignLoanSetByCounterparty(w, &tx, opts)
	if err != nil {
		return nil, "", "", err
	}
	return tx, txBlob, txHash, nil
}

// CombineLoanSetCounterpartySigners merges counterparty multisig transactions into a single transaction blob.
// All transactions must represent the same transaction (excluding CounterpartySignature.Signers).
func CombineLoanSetCounterpartySigners(transactions []transaction.FlatTransaction) (transaction.FlatTransaction, string, error) {
	if len(transactions) == 0 {
		return nil, "", ErrNoTransactionsToSign
	}

	// Validate and extract CounterpartySignature.Signers from each tx.
	allSigners := make([]any, 0, len(transactions))
	for _, tx := range transactions {
		if tx["TransactionType"] != "LoanSet" {
			return nil, "", ErrTxMustBeLoanSet
		}
		cs, ok := tx["CounterpartySignature"].(map[string]any)
		if !ok {
			return nil, "", ErrTxMustIncludeCounterpartySigners
		}
		signers, ok := cs["Signers"].([]any)
		if !ok || len(signers) == 0 {
			return nil, "", ErrTxMustIncludeCounterpartySigners
		}
		allSigners = append(allSigners, signers...)
	}

	if err := assertTransactionsEqual(transactions); err != nil {
		return nil, "", err
	}

	// Sort all signers by Account descending.
	xrpl.SortSigners(allSigners)

	firstTx := transactions[0]
	firstTx["CounterpartySignature"].(map[string]any)["Signers"] = allSigners

	encoded, err := binarycodec.Encode(firstTx)
	if err != nil {
		return nil, "", err
	}

	return firstTx, encoded, nil
}

// CombineLoanSetCounterpartySignersBlob decodes hex-encoded transaction blobs and combines
// their counterparty signers. This is a convenience wrapper around CombineLoanSetCounterpartySigners
// for callers that have serialized blobs rather than FlatTransactions.
func CombineLoanSetCounterpartySignersBlob(blobs []string) (transaction.FlatTransaction, string, error) {
	if len(blobs) == 0 {
		return nil, "", ErrNoTransactionsToSign
	}

	transactions := make([]transaction.FlatTransaction, len(blobs))
	for i, blob := range blobs {
		decoded, err := binarycodec.Decode(blob)
		if err != nil {
			return nil, "", err
		}
		transactions[i] = transaction.FlatTransaction(decoded)
	}

	return CombineLoanSetCounterpartySigners(transactions)
}

// encodeAndSign encodes tx for signing (multisig or single) and returns the hex signature.
// A shallow copy of tx is used because EncodeForSigning/EncodeForMultisigning mutate the map
// by removing non-signing fields, and the caller still needs the original fields (e.g. TxnSignature).
func encodeAndSign(w *Wallet, tx transaction.FlatTransaction, multisign bool, addr string) (string, error) {
	txCopy := make(transaction.FlatTransaction, len(tx))
	maps.Copy(txCopy, tx)

	var encoded string
	var err error
	if multisign {
		encoded, err = binarycodec.EncodeForMultisigning(txCopy, addr)
	} else {
		encoded, err = binarycodec.EncodeForSigning(txCopy)
	}
	if err != nil {
		return "", err
	}
	return w.computeSignature(encoded)
}

// resolveMultisignOpts returns the multisign flag and the counterparty address from opts.
func resolveMultisignOpts(defaultAddr string, opts *SignLoanSetByCounterpartyOptions) (bool, string) {
	if opts != nil {
		if opts.MultisignAccount != "" {
			return true, opts.MultisignAccount
		}
		if opts.Multisign {
			return true, defaultAddr
		}
	}
	return false, defaultAddr
}

// assertTransactionsEqual returns an error if any transaction in the slice differs from the first,
// ignoring CounterpartySignature.Signers.
func assertTransactionsEqual(transactions []transaction.FlatTransaction) error {
	var referenceBlob string
	for i, tx := range transactions {
		encoded, err := binarycodec.Encode(txWithoutCounterpartySigners(tx))
		if err != nil {
			return err
		}
		if i == 0 {
			referenceBlob = encoded
		} else if encoded != referenceBlob {
			return ErrLoanSetTxNotEqual
		}
	}
	return nil
}

// txWithoutCounterpartySigners returns a shallow copy of tx with CounterpartySignature.Signers removed.
// Used for equality comparison across blobs.
func txWithoutCounterpartySigners(tx transaction.FlatTransaction) transaction.FlatTransaction {
	txCopy := make(transaction.FlatTransaction, len(tx))
	maps.Copy(txCopy, tx)

	cs, ok := tx["CounterpartySignature"].(map[string]any)
	if !ok {
		return txCopy
	}

	// Shallow-copy the CounterpartySignature map and delete Signers.
	csCopy := make(map[string]any, len(cs))
	maps.Copy(csCopy, cs)
	delete(csCopy, "Signers")
	txCopy["CounterpartySignature"] = csCopy

	return txCopy
}
