// Package xrpl provides utilities for working with the XRP Ledger.
package xrpl

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"maps"
	"sort"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/keypairs"
)

// Multisign is a utility for signing a transaction offline.
// It takes a list of transaction blobs and returns the multisigned transaction blob.
// These transaction blobs must be signed with the wallet.Multisign method.
// They cannot contain SigningPubKey, otherwise the transaction will fail to submit.
// All blobs must represent the same transaction (excluding Signers); otherwise
// ErrMultisignTxNotEqual is returned.
// Every signer signature must be valid; otherwise ErrMultisignInvalidSignature
// is returned.
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

		txWithoutSigners := shallowCopyWithoutSigners(tx)

		encoded, err := binarycodec.Encode(txWithoutSigners)
		if err != nil {
			return "", err
		}

		if i == 0 {
			referenceBlob = encoded
			firstTx = tx
		} else if encoded != referenceBlob {
			return "", ErrMultisignTxNotEqual
		}

		txSigners, err := signersFromTx(tx)
		if err != nil {
			return "", err
		}
		if err := validateSignerSignatures(txWithoutSigners, txSigners); err != nil {
			return "", err
		}
		signers = append(signers, txSigners...)
	}

	SortSigners(signers)
	firstTx["Signers"] = signers

	blob, err := binarycodec.Encode(firstTx)
	if err != nil {
		return "", err
	}

	return blob, nil
}

// shallowCopyWithoutSigners returns a shallow copy of tx with Signers removed.
// It does not deep-copy nested maps or slices, so callers must not mutate
// shared nested values through the returned map.
func shallowCopyWithoutSigners(tx map[string]any) map[string]any {
	stripped := make(map[string]any, len(tx))
	maps.Copy(stripped, tx)
	delete(stripped, "Signers")
	return stripped
}

func signersFromTx(tx map[string]any) ([]any, error) {
	raw, ok := tx["Signers"]
	if !ok {
		return nil, fmt.Errorf("%w: missing Signers", ErrMultisignInvalidSigner)
	}

	signers, ok := raw.([]any)
	if !ok || len(signers) == 0 {
		return nil, fmt.Errorf("%w: Signers must be a non-empty array", ErrMultisignInvalidSigner)
	}

	return signers, nil
}

func validateSignerSignatures(txWithoutSigners map[string]any, signers []any) error {
	for _, signer := range signers {
		if err := validateSignerSignature(txWithoutSigners, signer); err != nil {
			return err
		}
	}
	return nil
}

func validateSignerSignature(txWithoutSigners map[string]any, signer any) error {
	account, signingPubKey, txnSignature, err := signerFields(signer)
	if err != nil {
		return err
	}

	payloadHex, err := binarycodec.EncodeForMultisigning(txWithoutSigners, account)
	if err != nil {
		return err
	}
	payloadBytes, err := hex.DecodeString(payloadHex)
	if err != nil {
		return err
	}

	valid, err := keypairs.Validate(string(payloadBytes), signingPubKey, txnSignature)
	if err != nil || !valid {
		return ErrMultisignInvalidSignature
	}
	return nil
}

func signerFields(signer any) (account, signingPubKey, txnSignature string, err error) {
	wrapper, ok := signer.(map[string]any)
	if !ok {
		return "", "", "", fmt.Errorf("%w: signer must be an object", ErrMultisignInvalidSigner)
	}

	rawSignerData, ok := wrapper["Signer"]
	if !ok {
		return "", "", "", fmt.Errorf("%w: missing Signer", ErrMultisignInvalidSigner)
	}

	signerData, ok := rawSignerData.(map[string]any)
	if !ok {
		return "", "", "", fmt.Errorf("%w: Signer must be an object", ErrMultisignInvalidSigner)
	}

	account, ok = signerData["Account"].(string)
	if !ok || account == "" {
		return "", "", "", fmt.Errorf("%w: missing Account", ErrMultisignInvalidSigner)
	}

	signingPubKey, ok = signerData["SigningPubKey"].(string)
	if !ok || signingPubKey == "" {
		return "", "", "", fmt.Errorf("%w: missing SigningPubKey", ErrMultisignInvalidSigner)
	}

	txnSignature, ok = signerData["TxnSignature"].(string)
	if !ok || txnSignature == "" {
		return "", "", "", fmt.Errorf("%w: missing TxnSignature", ErrMultisignInvalidSigner)
	}

	return account, signingPubKey, txnSignature, nil
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
