// Package xrpl provides utilities for working with the XRP Ledger.
package xrpl

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"maps"
	"sort"
	"slices"

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

	if err := SortSigners(signers); err != nil {
		return "", err
	}
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
		return nil, fmt.Errorf("%w: missing Signers", ErrInvalidSigner)
	}

	signers, ok := raw.([]any)
	if !ok || len(signers) == 0 {
		return nil, fmt.Errorf("%w: Signers must be a non-empty array", ErrInvalidSigner)
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
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMultisignInvalidSignature, err)
	}
	if !valid {
		return ErrMultisignInvalidSignature
	}
	return nil
}

func signerFields(signer any) (account, signingPubKey, txnSignature string, err error) {
	wrapper, ok := signer.(map[string]any)
	if !ok {
		return "", "", "", fmt.Errorf("%w: signer must be an object", ErrInvalidSigner)
	}

	rawSignerData, ok := wrapper["Signer"]
	if !ok {
		return "", "", "", fmt.Errorf("%w: missing Signer", ErrInvalidSigner)
	}

	signerData, ok := rawSignerData.(map[string]any)
	if !ok {
		return "", "", "", fmt.Errorf("%w: Signer must be an object", ErrInvalidSigner)
	}

	account, ok = signerData["Account"].(string)
	if !ok || account == "" {
		return "", "", "", fmt.Errorf("%w: missing Account", ErrInvalidSigner)
	}

	signingPubKey, ok = signerData["SigningPubKey"].(string)
	if !ok || signingPubKey == "" {
		return "", "", "", fmt.Errorf("%w: missing SigningPubKey", ErrInvalidSigner)
	}

	txnSignature, ok = signerData["TxnSignature"].(string)
	if !ok || txnSignature == "" {
		return "", "", "", fmt.Errorf("%w: missing TxnSignature", ErrInvalidSigner)
	}

	return account, signingPubKey, txnSignature, nil
}

// SortSigners sorts signers ascending by their decoded account ID bytes.
func SortSigners(signers []any) error {
	return SortByAccountID(signers, signerAccount)
}

// SortByAccountID sorts items in place by the decoded bytes of each item's classic XRPL account address.
// Use it for canonical signer ordering when different signer representations store the account in
// different fields. The account function extracts the classic address from an item and may return an
// error when the item does not contain one. SortByAccountID validates and decodes every account before
// sorting, so items stay in their original order if extraction or decoding fails.
func SortByAccountID[T any](items []T, account func(T) (string, error)) error {
	type sortableItem struct {
		item      T
		accountID []byte
	}

	sortable := make([]sortableItem, len(items))

	for i, item := range items {
		addr, err := account(item)
		if err != nil {
			return fmt.Errorf("sort by account ID: extract account at index %d: %w", i, err)
		}

		_, accountID, err := addresscodec.DecodeClassicAddressToAccountID(addr)
		if err != nil {
			return fmt.Errorf("sort by account ID: decode account at index %d: %w", i, err)
		}

		sortable[i] = sortableItem{item: item, accountID: accountID}
	}

	slices.SortFunc(sortable, func(a, b sortableItem) int {
		return bytes.Compare(a.accountID, b.accountID)
	})

	for i, item := range sortable {
		items[i] = item.item
	}

	return nil
}

func signerAccount(signer any) (string, error) {
	signerMap, ok := signer.(map[string]any)
	if !ok {
		return "", ErrInvalidSigner
	}

	signerData, ok := signerMap["Signer"].(map[string]any)
	if !ok {
		return "", ErrInvalidSigner
	}

	account, ok := signerData["Account"].(string)
	if !ok || account == "" {
		return "", ErrInvalidSigner
	}

	return account, nil
}
