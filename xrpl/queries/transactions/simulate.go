package transactions

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
)

// SimulateRequest is the request type for the XLS-69 simulate command.
// Exactly one transaction must be supplied as TxJSON or TxBlob. Blob inputs stay
// opaque so the server can validate them with its authoritative definitions.
// Binary selects whether the server returns transaction and metadata objects or
// hexadecimal binary blobs.
type SimulateRequest struct {
	common.BaseRequest
	TxJSON transaction.FlatTransaction `json:"tx_json,omitempty"`
	TxBlob string                      `json:"tx_blob,omitempty"`
	Binary bool                        `json:"binary,omitempty"`
}

// Method returns the JSON-RPC method name for SimulateRequest.
func (*SimulateRequest) Method() string {
	return "simulate"
}

// APIVersion returns the rippled API version for SimulateRequest.
func (*SimulateRequest) APIVersion() int {
	return version.RippledAPIV2
}

// MarshalJSON routes the WebSocket transport through the standard library
// encoder: without a json.Marshaler implementation, formatRequest falls back to
// mapstructure, which nests the embedded BaseRequest under its own key instead
// of flattening it.
func (r SimulateRequest) MarshalJSON() ([]byte, error) {
	type simulateRequestAlias SimulateRequest
	return json.Marshal(simulateRequestAlias(r))
}

// Validate verifies the exclusive input variants, hexadecimal blob syntax,
// unsigned JSON transaction shape, and any explicit JSON NetworkID value.
func (r *SimulateRequest) Validate() error {
	_, err := r.validatedTransaction()
	return err
}

// ValidateNetworkID validates JSON input against the client's current target
// network identity. Blob input stays opaque and is validated by the server.
// NetworkID can be omitted because the server autofills it. When supplied in
// JSON, it must match a known target network. A zero expected ID performs type
// validation only.
func (r *SimulateRequest) ValidateNetworkID(expected uint32) error {
	tx, err := r.validatedTransaction()
	if err != nil {
		return err
	}
	if r.TxBlob != "" {
		return nil
	}

	actual, present, err := simulateNetworkID(tx)
	if err != nil {
		return err
	}
	if present && expected != 0 && actual != expected {
		return ErrMismatchedSimulateNetworkID
	}
	return nil
}

func (r *SimulateRequest) validatedTransaction() (transaction.FlatTransaction, error) {
	if r == nil {
		return nil, ErrInvalidSimulateRequest
	}

	hasTxJSON := r.TxJSON != nil
	hasTxBlob := r.TxBlob != ""
	if hasTxJSON == hasTxBlob {
		return nil, ErrInvalidSimulateRequest
	}
	if hasTxBlob {
		if !typecheck.IsHexBlob(r.TxBlob) {
			return nil, ErrInvalidSimulateTxBlob
		}
		return nil, nil
	}

	tx := r.TxJSON
	if err := validateUnsignedSimulateTx(tx); err != nil {
		return nil, err
	}
	if !hasNonEmptyStringField(tx, "TransactionType") || !hasNonEmptyStringField(tx, "Account") {
		return nil, fmt.Errorf("%w: TransactionType and Account must be non-empty strings", ErrInvalidSimulateTxJSON)
	}
	if _, _, err := simulateNetworkID(tx); err != nil {
		return nil, err
	}
	return tx, nil
}

// SimulateResponse is the response returned by simulate. Results reflect the
// server's current open-ledger state and do not guarantee the outcome of a later
// submission because ledger state can change.
//
// JSON responses contain TxJSON and optional Meta. Binary responses contain
// TxBlob and optional MetaBlob. Metadata is absent for engine results that would
// not be included in a ledger, such as non-tec failures.
type SimulateResponse struct {
	Applied             bool                           `json:"applied"`
	EngineResult        string                         `json:"engine_result"`
	EngineResultCode    int                            `json:"engine_result_code"`
	EngineResultMessage string                         `json:"engine_result_message"`
	LedgerIndex         common.LedgerIndex             `json:"ledger_index"`
	TxJSON              transaction.FlatTransaction    `json:"tx_json,omitempty"`
	TxBlob              string                         `json:"tx_blob,omitempty"`
	Meta                *transaction.TxMetadataBuilder `json:"meta,omitempty"`
	MetaBlob            string                         `json:"meta_blob,omitempty"`
}

// Validate verifies the mutually exclusive JSON and binary response variants,
// including hexadecimal binary payloads. Metadata remains optional in either
// variant because non-tec engine failures do not produce it.
func (r SimulateResponse) Validate() error {
	if r.Applied {
		return fmt.Errorf("%w: applied must be false for a dry run", ErrInvalidSimulateResponse)
	}
	if r.EngineResult == "" || r.EngineResultMessage == "" {
		return fmt.Errorf("%w: engine result fields must not be empty", ErrInvalidSimulateResponse)
	}

	hasTxJSON := len(r.TxJSON) > 0
	hasTxBlob := r.TxBlob != ""
	if hasTxJSON == hasTxBlob {
		return ErrInvalidSimulateResponse
	}

	if hasTxJSON {
		if r.MetaBlob != "" {
			return fmt.Errorf("%w: JSON response cannot contain meta_blob", ErrInvalidSimulateResponse)
		}
		return nil
	}

	if r.Meta != nil {
		return fmt.Errorf("%w: binary response cannot contain meta", ErrInvalidSimulateResponse)
	}
	if !typecheck.IsHexBlob(r.TxBlob) {
		return fmt.Errorf("%w: tx_blob must be hexadecimal", ErrInvalidSimulateResponse)
	}
	if r.MetaBlob != "" && !typecheck.IsHexBlob(r.MetaBlob) {
		return fmt.Errorf("%w: meta_blob must be hexadecimal", ErrInvalidSimulateResponse)
	}
	return nil
}

// ValidateForRequest verifies that the response payload matches the output mode
// selected by the request.
func (r SimulateResponse) ValidateForRequest(req *SimulateRequest) error {
	if req == nil {
		return ErrInvalidSimulateRequest
	}
	if err := r.Validate(); err != nil {
		return err
	}
	if req.Binary && r.TxBlob == "" {
		return fmt.Errorf("%w: binary output was requested but the response contains JSON output", ErrInvalidSimulateResponse)
	}
	if !req.Binary && len(r.TxJSON) == 0 {
		return fmt.Errorf("%w: JSON output was requested but the response contains binary output", ErrInvalidSimulateResponse)
	}
	return nil
}

// MarshalJSON validates and encodes a JSON or binary simulate response.
func (r SimulateResponse) MarshalJSON() ([]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	type simulateResponseAlias SimulateResponse
	return json.Marshal(simulateResponseAlias(r))
}

// UnmarshalJSON decodes and validates a JSON or binary simulate response.
func (r *SimulateResponse) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for _, name := range []string{"applied", "engine_result", "engine_result_code", "engine_result_message", "ledger_index"} {
		value, present := fields[name]
		if !present || string(value) == "null" {
			return fmt.Errorf("%w: missing %s", ErrInvalidSimulateResponse, name)
		}
	}

	_, hasTxJSON := fields["tx_json"]
	_, hasTxBlob := fields["tx_blob"]
	if hasTxJSON == hasTxBlob {
		return ErrInvalidSimulateResponse
	}
	_, hasMeta := fields["meta"]
	_, hasMetaBlob := fields["meta_blob"]

	type simulateResponseAlias SimulateResponse
	var decoded simulateResponseAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidSimulateResponse, err)
	}
	response := SimulateResponse(decoded)
	if hasTxJSON && response.TxJSON == nil {
		return fmt.Errorf("%w: tx_json must be an object", ErrInvalidSimulateResponse)
	}
	if hasMeta && response.Meta == nil {
		return fmt.Errorf("%w: meta must be an object", ErrInvalidSimulateResponse)
	}
	if hasMetaBlob && response.MetaBlob == "" {
		return fmt.Errorf("%w: meta_blob must be a non-empty hexadecimal string", ErrInvalidSimulateResponse)
	}
	if err := response.Validate(); err != nil {
		return err
	}

	*r = response
	return nil
}

func validateUnsignedSimulateTx(tx transaction.FlatTransaction) error {
	if value, present := tx["TxnSignature"]; present {
		signature, ok := underlyingString(value)
		if !ok {
			return fmt.Errorf("%w: TxnSignature must be a string", ErrInvalidSimulateTxJSON)
		}
		if signature != "" {
			return fmt.Errorf("%w: TxnSignature is non-empty", ErrSignedSimulateTransaction)
		}
	}

	// SigningPubKey identifies the signing key but does not sign the transaction.
	if value, present := tx["SigningPubKey"]; present {
		if _, ok := underlyingString(value); !ok {
			return fmt.Errorf("%w: SigningPubKey must be a string", ErrInvalidSimulateTxJSON)
		}
	}

	if signers, present := tx["Signers"]; present {
		if err := validateUnsignedSimulateSigners(signers, "Signers"); err != nil {
			return err
		}
	}

	if batchSigners, present := tx["BatchSigners"]; present {
		if err := validateUnsignedSimulateBatchSigners(batchSigners); err != nil {
			return err
		}
	}
	return nil
}

type simulateSignerFields struct {
	SigningPubKey json.RawMessage `json:"SigningPubKey"`
	TxnSignature  json.RawMessage `json:"TxnSignature"`
}

type simulateSignerEntry struct {
	Signer *simulateSignerFields `json:"Signer"`
}

type simulateBatchSignerFields struct {
	SigningPubKey json.RawMessage `json:"SigningPubKey"`
	TxnSignature  json.RawMessage `json:"TxnSignature"`
	Signers       json.RawMessage `json:"Signers"`
}

type simulateBatchSignerEntry struct {
	BatchSigner *simulateBatchSignerFields `json:"BatchSigner"`
}

func validateUnsignedSimulateSigners(signers any, field string) error {
	data, err := json.Marshal(signers)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrInvalidSimulateTxJSON, field, err)
	}
	if string(data) == "null" {
		return fmt.Errorf("%w: %s must be an array", ErrInvalidSimulateTxJSON, field)
	}

	entries := make([]simulateSignerEntry, 0)
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("%w: %s must be an array of Signer objects: %w", ErrInvalidSimulateTxJSON, field, err)
	}

	for i, entry := range entries {
		signerField := fmt.Sprintf("%s[%d].Signer", field, i)
		if entry.Signer == nil {
			return fmt.Errorf("%w: %s must be an object", ErrInvalidSimulateTxJSON, signerField)
		}

		if _, _, err := decodeSimulateString(entry.Signer.SigningPubKey, signerField+".SigningPubKey"); err != nil {
			return err
		}
		signature, signaturePresent, err := decodeSimulateString(entry.Signer.TxnSignature, signerField+".TxnSignature")
		if err != nil {
			return err
		}
		if signaturePresent && signature != "" {
			return fmt.Errorf("%w: %s.TxnSignature is non-empty", ErrSignedSimulateTransaction, signerField)
		}
	}
	return nil
}

func validateUnsignedSimulateBatchSigners(batchSigners any) error {
	data, err := json.Marshal(batchSigners)
	if err != nil {
		return fmt.Errorf("%w: BatchSigners: %w", ErrInvalidSimulateTxJSON, err)
	}
	if string(data) == "null" {
		return fmt.Errorf("%w: BatchSigners must be an array", ErrInvalidSimulateTxJSON)
	}

	entries := make([]simulateBatchSignerEntry, 0)
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("%w: BatchSigners must be an array of BatchSigner objects: %w", ErrInvalidSimulateTxJSON, err)
	}

	for i, entry := range entries {
		batchSignerField := fmt.Sprintf("BatchSigners[%d].BatchSigner", i)
		if entry.BatchSigner == nil {
			return fmt.Errorf("%w: %s must be an object", ErrInvalidSimulateTxJSON, batchSignerField)
		}

		if _, _, err := decodeSimulateString(entry.BatchSigner.SigningPubKey, batchSignerField+".SigningPubKey"); err != nil {
			return err
		}
		signature, signaturePresent, err := decodeSimulateString(
			entry.BatchSigner.TxnSignature,
			batchSignerField+".TxnSignature",
		)
		if err != nil {
			return err
		}
		if signaturePresent && signature != "" {
			return fmt.Errorf("%w: %s.TxnSignature is non-empty", ErrSignedSimulateTransaction, batchSignerField)
		}

		if len(entry.BatchSigner.Signers) > 0 {
			if err := validateUnsignedSimulateSigners(entry.BatchSigner.Signers, batchSignerField+".Signers"); err != nil {
				return err
			}
		}
	}
	return nil
}

func decodeSimulateString(raw json.RawMessage, field string) (string, bool, error) {
	if len(raw) == 0 {
		return "", false, nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", true, fmt.Errorf("%w: %s must be a string", ErrInvalidSimulateTxJSON, field)
	}
	return value, true, nil
}

func hasNonEmptyStringField(tx transaction.FlatTransaction, field string) bool {
	value, present := tx[field]
	if !present {
		return false
	}
	stringValue, ok := underlyingString(value)
	return ok && stringValue != ""
}

// underlyingString accepts named string types (e.g. types.Address) that appear
// in hand-built FlatTransaction maps, so a plain .(string) assertion is not enough.
// A nil value yields reflect.Invalid and is rejected by the kind check.
func underlyingString(value any) (string, bool) {
	reflected := reflect.ValueOf(value)
	if reflected.Kind() != reflect.String {
		return "", false
	}
	return reflected.String(), true
}

func simulateNetworkID(tx transaction.FlatTransaction) (uint32, bool, error) {
	value, present := tx["NetworkID"]
	if !present {
		return 0, false, nil
	}
	networkID, ok := typecheck.ToUint32(value)
	if !ok {
		return 0, true, ErrInvalidSimulateNetworkID
	}
	return networkID, true, nil
}
