package hash

import (
	"encoding/binary"
	"encoding/hex"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
)

// SignTxBlob hashes a signed transaction blob
// It takes a transaction blob and returns the hash of the signed transaction.
// It returns an error if the transaction blob is invalid.
func SignTxBlob(txBlob string) (string, error) {
	tx, err := clientinternal.DecodeTransactionBlob(txBlob)
	if err != nil {
		return "", err
	}

	if err := isTxValid(tx); err != nil {
		return "", err
	}

	return encodeSignedTxBlob(txBlob)
}

// SignTx hashes a signed transaction
// It takes a signed transaction and returns the hash of the signed transaction.
// It returns an error if the transaction is invalid.
func SignTx(tx map[string]any) (string, error) {
	if err := isTxValid(tx); err != nil {
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

func isTxValid(tx map[string]any) error {
	form, err := clientinternal.InspectSignedTransaction(tx, true)
	if err != nil {
		return err
	}
	if form == clientinternal.UnsignedTransaction {
		return ErrNonSignedTransaction
	}
	return nil
}
