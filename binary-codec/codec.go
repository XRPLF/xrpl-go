// Package binarycodec provides encoding and decoding functionality for XRPL binary format.
package binarycodec

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"strings"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"

	"github.com/Peersyst/xrpl-go/binary-codec/serdes"
	"github.com/Peersyst/xrpl-go/binary-codec/types"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
)

const (
	txMultiSigPrefix          = "534D5400"
	paymentChannelClaimPrefix = "434C4D00"
	txSigPrefix               = "53545800"
	batchPrefix               = "42434800"
	unlModifyTransactionType  = "UNLModify"
	xrplZeroAccount           = "rrrrrrrrrrrrrrrrrrrrrhoLvTp"
)

// Encode converts a JSON transaction object to a hex string in the canonical binary format.
// The binary format is defined in XRPL's core codebase.
func Encode(json map[string]any) (string, error) {
	st := types.NewSTObject(serdes.NewBinarySerializer(serdes.NewFieldIDCodec(definitions.Get())))

	filteredJSON := make(map[string]any, len(json))
	for k, v := range json {
		if definitions.Get().Fields[k] != nil {
			filteredJSON[k] = v
		}
	}

	overrides, err := transactionRawFieldValueOverrides(filteredJSON)
	if err != nil {
		return "", err
	}
	if overrides != nil {
		// UInt16 expects a built-in string. Normalize only the private copy used for encoding.
		filteredJSON["TransactionType"] = unlModifyTransactionType
	}

	b, err := st.FromJSONWithRawFieldValueOverrides(filteredJSON, overrides)
	if err != nil {
		return "", err
	}

	return hexutil.EncodeToUpperHex(b), nil
}

func transactionRawFieldValueOverrides(json map[string]any) (types.RawFieldValueOverrides, error) {
	transactionType, ok := typecheck.ToString(json["TransactionType"])
	if !ok || transactionType != unlModifyTransactionType {
		return nil, nil
	}

	if accountValue, present := json["Account"]; present {
		account, isString := typecheck.ToString(accountValue)
		if !isString {
			return nil, fmt.Errorf("%w; got %T", ErrInvalidUNLModifyAccount, accountValue)
		}
		if account != "" && account != xrplZeroAccount {
			return nil, fmt.Errorf("%w; got %q", ErrInvalidUNLModifyAccount, account)
		}
	}

	// rippled represents the Account of a consensus-generated UNLModify transaction as a
	// default STAccount. STAccount::add serializes that default as a zero-length value:
	// https://github.com/XRPLF/rippled/blob/d4c1359921f34a4e96c5c8483119e59f0e30e4df/src/libxrpl/protocol/STAccount.cpp#L71-L81
	return types.RawFieldValueOverrides{"Account": []byte{}}, nil
}

// EncodeForMultisigning encodes a transaction into binary format in preparation for providing one
// signature towards a multi-signed transaction.
// Only encodes fields that are intended to be signed.
// NOTE: The caller is responsible for setting SigningPubKey to "" for regular multisigning.
// For counterparty signing (e.g. LoanSet), SigningPubKey must remain set to the first signer's
// public key, so this function must not overwrite it.
func EncodeForMultisigning(json map[string]any, xrpAccountID string) (string, error) {
	st := &types.AccountID{}

	suffix, err := st.FromJSON(xrpAccountID)
	if err != nil {
		return "", err
	}

	encoded, err := Encode(signingFieldsOnly(json))
	if err != nil {
		return "", err
	}

	return strings.ToUpper(txMultiSigPrefix + encoded + hex.EncodeToString(suffix)), nil
}

// EncodeForSigning encodes a transaction into binary format in preparation for signing.
func EncodeForSigning(json map[string]any) (string, error) {
	encoded, err := Encode(signingFieldsOnly(json))
	if err != nil {
		return "", err
	}

	return strings.ToUpper(txSigPrefix + encoded), nil
}

// signingFieldsOnly returns a new map containing only the fields from the JSON transaction that are signing fields.
func signingFieldsOnly(json map[string]any) map[string]any {
	signingFields := make(map[string]any, len(json))
	for k, v := range json {
		fi, _ := definitions.Get().GetFieldInstanceByFieldName(k)
		if fi != nil && fi.IsSigningField {
			signingFields[k] = v
		}
	}

	return signingFields
}

// EncodeForSigningClaim encodes a payment channel claim into binary format in preparation for signing.
func EncodeForSigningClaim(json map[string]any) (string, error) {
	if json["Channel"] == nil || json["Amount"] == nil {
		return "", ErrSigningClaimFieldNotFound
	}

	channel, err := types.NewHash256().FromJSON(json["Channel"])
	if err != nil {
		return "", err
	}

	t := &types.Amount{}
	amount, err := t.FromJSON(json["Amount"])
	if err != nil {
		return "", err
	}

	if bytes.HasPrefix(amount, []byte{0x40}) {
		amount = bytes.Replace(amount, []byte{0x40}, []byte{0x00}, 1)
	}

	return strings.ToUpper(paymentChannelClaimPrefix + hex.EncodeToString(channel) + hex.EncodeToString(amount)), nil
}

// EncodeForSigningBatch encodes a BatchV1_1 transaction payload in preparation for signing.
func EncodeForSigningBatch(json map[string]any) (string, error) {
	if json["account"] == nil {
		return "", ErrBatchAccountFieldNotFound
	}
	if json["sequence"] == nil {
		return "", ErrBatchSequenceFieldNotFound
	}
	if json["flags"] == nil {
		return "", ErrBatchFlagsFieldNotFound
	}
	if json["txIDs"] == nil {
		return "", ErrBatchTxIDsFieldNotFound
	}

	batchAccount, hasBatchAccount := json["batchAccount"]
	signerAccount, hasSignerAccount := json["signerAccount"]
	if hasSignerAccount && !hasBatchAccount {
		return "", ErrBatchSignerAccountWithoutBatchAccount
	}

	txIDs, ok := json["txIDs"].([]string)
	if !ok {
		return "", ErrBatchTxIDsNotArray
	}
	if _, ok = json["flags"].(uint32); !ok {
		return "", ErrBatchFlagsNotUInt32
	}

	txIDsLength := len(txIDs)
	if txIDsLength > math.MaxUint32 {
		return "", ErrBatchTxIDsLengthTooLong
	}

	var result strings.Builder
	result.WriteString(batchPrefix)

	accountBytes, err := (&types.AccountID{}).FromJSON(json["account"])
	if err != nil {
		return "", fmt.Errorf("encode BatchV1_1 account: %w", err)
	}
	result.WriteString(hex.EncodeToString(accountBytes))

	sequenceBytes, err := (&types.UInt32{}).FromJSON(json["sequence"])
	if err != nil {
		return "", fmt.Errorf("encode BatchV1_1 sequence: %w", err)
	}
	result.WriteString(hex.EncodeToString(sequenceBytes))

	flagsBytes, err := (&types.UInt32{}).FromJSON(json["flags"])
	if err != nil {
		return "", fmt.Errorf("encode BatchV1_1 flags: %w", err)
	}
	result.WriteString(hex.EncodeToString(flagsBytes))

	txIDsLengthBytes, err := (&types.UInt32{}).FromJSON(uint32(txIDsLength))
	if err != nil {
		return "", fmt.Errorf("encode BatchV1_1 txIDs length: %w", err)
	}
	result.WriteString(hex.EncodeToString(txIDsLengthBytes))

	for i, txID := range txIDs {
		txIDBytes, err := types.NewHash256().FromJSON(txID)
		if err != nil {
			return "", fmt.Errorf("encode BatchV1_1 txIDs[%d]: %w", i, err)
		}
		result.WriteString(hex.EncodeToString(txIDBytes))
	}

	if hasBatchAccount {
		batchAccountBytes, err := (&types.AccountID{}).FromJSON(batchAccount)
		if err != nil {
			return "", fmt.Errorf("encode BatchV1_1 batchAccount: %w", err)
		}
		result.WriteString(hex.EncodeToString(batchAccountBytes))
	}

	if hasSignerAccount {
		signerAccountBytes, err := (&types.AccountID{}).FromJSON(signerAccount)
		if err != nil {
			return "", fmt.Errorf("encode BatchV1_1 signerAccount: %w", err)
		}
		result.WriteString(hex.EncodeToString(signerAccountBytes))
	}

	return strings.ToUpper(result.String()), nil
}

// Decode decodes a hex string in the canonical binary format into a JSON transaction object.
func Decode(hexEncoded string) (map[string]any, error) {
	b, err := hex.DecodeString(hexEncoded)
	if err != nil {
		return nil, err
	}
	p := serdes.NewBinaryParser(b, definitions.Get())
	st := types.NewSTObject(serdes.NewBinarySerializer(serdes.NewFieldIDCodec(definitions.Get())))
	m, err := st.ToJSON(p)
	if err != nil {
		return nil, err
	}

	return m.(map[string]any), nil
}
