// Package xrpl provides utilities for working with the XRP Ledger.
package xrpl

import (
	"bytes"
	"sort"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
)

// Multisign is a utility for signing a transaction offline.
// It takes a list of transaction blobs and returns the multisigned transaction blob.
// These transaction blobs must be signed with the wallet.Multisign method.
// They cannot contain SigningPubKey, otherwise the transaction will fail to submit.
// If an error occurs, it will return an error.
func Multisign(blobs ...string) (string, error) {
	if len(blobs) == 0 {
		return "", ErrNoTxToMultisign
	}

	signers := make([]interface{}, 0)
	for _, blob := range blobs {
		tx, err := binarycodec.Decode(blob)
		if err != nil {
			return "", err
		}
		if pk, ok := tx["SigningPubKey"].(string); ok && pk != "" {
			return "", ErrMultisignNonEmptySigningPubKey
		}

		signers = append(signers, tx["Signers"].([]interface{})...)
	}

	tx, err := binarycodec.Decode(blobs[0])
	if err != nil {
		return "", err
	}

	SortSigners(signers)
	tx["Signers"] = signers

	blob, err := binarycodec.Encode(tx)
	if err != nil {
		return "", err
	}

	return blob, nil
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
