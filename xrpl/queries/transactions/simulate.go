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

// restrictedNetworkID mirrors rpc.RestrictedNetworks and websocket.RestrictedNetworks,
// which this package cannot import without a cycle: networks with IDs above this
// threshold require an explicit NetworkID on every transaction.
const restrictedNetworkID = 1024

// SimulateRequest is the request type for the XLS-69 simulate command.
// Exactly one transaction must be supplied as TxJSON or TxBlob. Non-empty
// signature fields in TxJSON are rejected locally; the server validates the
// decoded contents of a serialized TxBlob. Binary selects whether the server
// returns transaction and metadata objects or hexadecimal binary blobs.
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

// Validate verifies the exclusive input variants, blob encoding, unsigned JSON
// transaction shape, and any explicit NetworkID value.
func (r *SimulateRequest) Validate() error {
	hasTxJSON := r.TxJSON != nil
	hasTxBlob := r.TxBlob != ""
	if hasTxJSON == hasTxBlob {
		return ErrInvalidSimulateRequest
	}

	if hasTxBlob {
		if !typecheck.IsHexBlob(r.TxBlob) {
			return ErrInvalidSimulateTxBlob
		}
		return nil
	}

	if err := validateUnsignedSimulateTx(r.TxJSON); err != nil {
		return err
	}
	if !hasNonEmptyStringField(r.TxJSON, "TransactionType") || !hasNonEmptyStringField(r.TxJSON, "Account") {
		return fmt.Errorf("%w: TransactionType and Account must be non-empty strings", ErrInvalidSimulateTxJSON)
	}
	if _, _, err := simulateNetworkID(r.TxJSON); err != nil {
		return err
	}
	return nil
}

// ValidateNetworkID validates tx_json against the client's current target
// network identity. Networks with IDs above 1024 require an explicit matching
// NetworkID. On other identified networks, a supplied NetworkID must still match.
// Blob inputs are already serialized and are left for the server to validate.
func (r *SimulateRequest) ValidateNetworkID(expected uint32) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if r.TxJSON == nil {
		return nil
	}

	actual, present, err := simulateNetworkID(r.TxJSON)
	if err != nil {
		return err
	}
	if expected > restrictedNetworkID && !present {
		return ErrMissingSimulateNetworkID
	}
	if present && expected != 0 && actual != expected {
		return ErrMismatchedSimulateNetworkID
	}
	return nil
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
	for _, field := range []string{"TxnSignature", "SigningPubKey"} {
		value, present := tx[field]
		if !present {
			continue
		}
		stringValue, ok := underlyingString(value)
		if !ok {
			return fmt.Errorf("%w: %s must be a string", ErrInvalidSimulateTxJSON, field)
		}
		if stringValue != "" {
			return fmt.Errorf("%w: %s is non-empty", ErrSignedSimulateTransaction, field)
		}
	}

	signers, present := tx["Signers"]
	if !present {
		return nil
	}
	value := reflect.ValueOf(signers)
	if value.Kind() != reflect.Array && value.Kind() != reflect.Slice {
		return fmt.Errorf("%w: Signers must be an array", ErrInvalidSimulateTxJSON)
	}
	if value.Len() > 0 {
		return fmt.Errorf("%w: Signers is non-empty", ErrSignedSimulateTransaction)
	}
	return nil
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
