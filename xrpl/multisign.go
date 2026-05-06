// Package xrpl provides utilities for working with the XRP Ledger.
package xrpl

import (
	"bytes"
	"maps"
	"sort"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
)

// Multisign is a utility for signing a transaction offline.
// It takes a list of transaction blobs and returns the multisigned transaction blob.
// These transaction blobs must be signed with the wallet.Multisign method.
// They cannot contain SigningPubKey, otherwise the transaction will fail to submit.
// All blobs must represent the same transaction (excluding Signers); otherwise
// ErrMultisignTxNotEqual is returned.
// If an error occurs, it will return an error.
func Multisign(blobs ...string) (string, error) {
	if len(blobs) == 0 {
		return "", ErrNoTxToMultisign
	}

	var firstTx map[string]any
	var referenceBlob string
	signers := make([]any, 0)
	for i, blob := range blobs {
		tx, err := binarycodec.Decode(blob)
		if err != nil {
			return "", err
		}
		if pk, ok := tx["SigningPubKey"].(string); ok && pk != "" {
			return "", ErrMultisignNonEmptySigningPubKey
		}

		encoded, err := encodeWithoutSigners(tx)
		if err != nil {
			return "", err
		}
		if i == 0 {
			referenceBlob = encoded
			firstTx = tx
		} else if encoded != referenceBlob {
			return "", ErrMultisignTxNotEqual
		}

		signers = append(signers, tx["Signers"].([]any)...)
	}

	SortSigners(signers)
	firstTx["Signers"] = signers

	blob, err := binarycodec.Encode(firstTx)
	if err != nil {
		return "", err
	}

	return blob, nil
}

// encodeWithoutSigners returns the binary-encoded form of tx with the Signers
// field removed. Used to compare blobs that represent the same transaction but
// carry different signers. A shallow copy is made so removing Signers does not
// mutate tx.
func encodeWithoutSigners(tx map[string]any) (string, error) {
	stripped := make(map[string]any, len(tx))
	maps.Copy(stripped, tx)
	delete(stripped, "Signers")
	return binarycodec.Encode(stripped)
}

// SortSigners sorts signers ascending by their decoded account ID bytes.
// This matches xrpl.js's compareSigners which sorts by addressToBigNumber ascending.
func SortSigners(signers []any) {
	sort.Slice(signers, func(i, j int) bool {
		iAccount := signers[i].(map[string]any)["Signer"].(map[string]any)["Account"].(string)
		jAccount := signers[j].(map[string]any)["Signer"].(map[string]any)["Account"].(string)

		_, iBytes, _ := addresscodec.DecodeClassicAddressToAccountID(iAccount)
		_, jBytes, _ := addresscodec.DecodeClassicAddressToAccountID(jAccount)

		return bytes.Compare(iBytes, jBytes) < 0
	})
}
