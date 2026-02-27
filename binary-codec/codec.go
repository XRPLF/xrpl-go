// Package binarycodec provides encoding and decoding functionality for XRPL binary format.
package binarycodec

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math"
	"strings"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"

	"github.com/Peersyst/xrpl-go/binary-codec/serdes"
	"github.com/Peersyst/xrpl-go/binary-codec/types"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
)

var (
	// Static errors

	// ErrSigningClaimFieldNotFound is returned when the 'Channel' & 'Amount' fields are both required, but were not found.
	ErrSigningClaimFieldNotFound = errors.New("'Channel' & 'Amount' fields are both required, but were not found")
	// ErrBatchFlagsFieldNotFound is returned when the 'flags' field is missing.
	ErrBatchFlagsFieldNotFound = errors.New("no field `flags`")
	// ErrBatchTxIDsFieldNotFound is returned when the 'txIDs' field is missing.
	ErrBatchTxIDsFieldNotFound = errors.New("no field `txIDs`")
	// ErrBatchTxIDsNotArray is returned when the 'txIDs' field is not an array.
	ErrBatchTxIDsNotArray = errors.New("txIDs field must be an array")
	// ErrBatchTxIDNotString is returned when a txID is not a string.
	ErrBatchTxIDNotString = errors.New("each txID must be a string")
	// ErrBatchFlagsNotUInt32 is returned when the 'flags' field is not a uint32.
	ErrBatchFlagsNotUInt32 = errors.New("flags field must be a uint32")
	// ErrBatchTxIDsLengthTooLong is returned when the 'txIDs' field is too long.
	ErrBatchTxIDsLengthTooLong = errors.New("txIDs length exceeds maximum uint32 value")
)

const (
	txMultiSigPrefix          = "534D5400"
	paymentChannelClaimPrefix = "434C4D00"
	txSigPrefix               = "53545800"
	batchPrefix               = "42434800"
)

// Encode converts a JSON transaction object to a hex string in the canonical binary format.
// The binary format is defined in XRPL's core codebase.
func Encode(json map[string]any) (string, error) {
	st := types.NewSTObject(serdes.NewBinarySerializer(serdes.NewFieldIDCodec(definitions.Get())))

	// Iterate over the keys in the provided JSON
	for k := range json {

		// Get the FieldIdNameMap from the definitions package
		fh := definitions.Get().Fields[k]

		// If the field is not found in the FieldIdNameMap, delete it from the JSON

		if fh == nil {
			delete(json, k)
			continue
		}
	}

	b, err := st.FromJSON(json)
	if err != nil {
		return "", err
	}

	return hexutil.EncodeToUpperHex(b), nil
}

// EncodeForMultisigning encodes a transaction into binary format in preparation for providing one
// signature towards a multi-signed transaction.
// Only encodes fields that are intended to be signed.
func EncodeForMultisigning(json map[string]any, xrpAccountID string) (string, error) {
	st := &types.AccountID{}

	// SigningPubKey is required for multi-signing but should be set to empty string.

	json["SigningPubKey"] = ""

	suffix, err := st.FromJSON(xrpAccountID)
	if err != nil {
		return "", err
	}

	encoded, err := Encode(removeNonSigningFields(json))

	if err != nil {
		return "", err
	}

	return strings.ToUpper(txMultiSigPrefix + encoded + hex.EncodeToString(suffix)), nil
}

// EncodeForSigning encodes a transaction into binary format in preparation for signing.
func EncodeForSigning(json map[string]any) (string, error) {

	encoded, err := Encode(removeNonSigningFields(json))

	if err != nil {
		return "", err
	}

	return strings.ToUpper(txSigPrefix + encoded), nil
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

// EncodeForSigningBatch encodes a batch transaction into binary format in preparation for signing.
func EncodeForSigningBatch(json map[string]any) (string, error) {
	if json["flags"] == nil {
		return "", ErrBatchFlagsFieldNotFound
	}
	if json["txIDs"] == nil {
		return "", ErrBatchTxIDsFieldNotFound
	}

	// Extract and validate txIDs
	txIDsInterface, ok := json["txIDs"].([]string)
	if !ok {
		return "", ErrBatchTxIDsNotArray
	}

	// Validate flags type
	_, ok = json["flags"].(uint32)
	if !ok {
		return "", ErrBatchFlagsNotUInt32
	}

	// Create UInt32 for flags
	flagsType := &types.UInt32{}
	flagsBytes, err := flagsType.FromJSON(json["flags"])
	if err != nil {
		return "", err
	}

	// Create UInt32 for txIDs length
	txIDsLengthType := &types.UInt32{}
	txIDsLength := len(txIDsInterface)
	if txIDsLength > math.MaxUint32 {
		return "", ErrBatchTxIDsLengthTooLong
	}
	txIDsLengthBytes, err := txIDsLengthType.FromJSON(uint32(txIDsLength))
	if err != nil {
		return "", err
	}

	// Build the result string
	result := batchPrefix + hex.EncodeToString(flagsBytes) + hex.EncodeToString(txIDsLengthBytes)

	// Add each transaction ID
	for _, txID := range txIDsInterface {
		hash256 := types.NewHash256()
		txIDBytes, err := hash256.FromJSON(txID)
		if err != nil {
			return "", err
		}
		result += hex.EncodeToString(txIDBytes)
	}

	return strings.ToUpper(result), nil
}

// removeNonSigningFields removes the fields from a JSON transaction object that should not be signed.
func removeNonSigningFields(json map[string]any) map[string]any {
	for k := range json {
		fi, _ := definitions.Get().GetFieldInstanceByFieldName(k)

		if fi != nil && !fi.IsSigningField {
			delete(json, k)
		}
	}

	return json
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
