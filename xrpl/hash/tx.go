package hash

import (
	"encoding/binary"
	"encoding/hex"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
)

// SignTxBlob hashes a signed or pseudo-transaction blob.
// It returns an error if the transaction blob is invalid.
func SignTxBlob(txBlob string) (string, error) {
	tx, err := clientinternal.DecodeTransactionBlob(txBlob)
	if err != nil {
		return "", err
	}

	if err := validateHashableTransactionForm(tx); err != nil {
		return "", err
	}

	return encodeSignedTxBlob(txBlob)
}

// SignTx hashes a signed or pseudo-transaction.
// It returns an error if the transaction is invalid.
func SignTx(tx map[string]any) (string, error) {
	if err := validateHashableTransactionForm(tx); err != nil {
		return "", err
	}

	txBlob, err := binarycodec.Encode(tx)
	if err != nil {
		return "", err
	}

	return encodeSignedTxBlob(txBlob)
}

func encodeSignedTxBlob(txBlob string) (string, error) {
	// Create a byte slice with the correct capacity
	payload := make([]byte, 4+len(txBlob)/2)

	// Convert TRANSACTION_PREFIX to big-endian bytes
	binary.BigEndian.PutUint32(payload[:4], TransactionPrefix)

	// Decode the txBlob into the rest of the payload
	_, err := hex.Decode(payload[4:], []byte(txBlob))
	if err != nil {
		return "", err
	}

	return EncodeToHashString(payload), nil
}

func validateHashableTransactionForm(tx map[string]any) error {
	txType, _ := tx["TransactionType"].(string)
	if transaction.IsPseudoTransactionType(transaction.TxType(txType)) {
		return nil
	}

	// Allow the canonical unsigned form used by inner Batch transactions.
	signingType, err := clientinternal.InspectSignedTransaction(tx, true)
	if err != nil {
		return err
	}
	if signingType == clientinternal.UnsignedTransaction {
		return ErrNonSignedTransaction
	}
	return nil
}
