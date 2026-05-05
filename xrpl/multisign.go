// Package xrpl provides utilities for working with the XRP Ledger.
package xrpl

import (
	"bytes"
	"slices"

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

	signers := make([]any, 0)
	for _, blob := range blobs {
		tx, err := binarycodec.Decode(blob)
		if err != nil {
			return "", err
		}
		if pk, ok := tx["SigningPubKey"].(string); ok && pk != "" {
			return "", ErrMultisignNonEmptySigningPubKey
		}

		signers = append(signers, tx["Signers"].([]any)...)
	}

	tx, err := binarycodec.Decode(blobs[0])
	if err != nil {
		return "", err
	}

	if err := SortSigners(signers); err != nil {
		return "", err
	}
	tx["Signers"] = signers

	blob, err := binarycodec.Encode(tx)
	if err != nil {
		return "", err
	}

	return blob, nil
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
			return err
		}

		_, accountID, err := addresscodec.DecodeClassicAddressToAccountID(addr)
		if err != nil {
			return err
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
