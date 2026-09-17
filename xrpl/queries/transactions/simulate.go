package transactions

import (
	"encoding/json"
	"fmt"

	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
)

// SimulateRequest is the request type for the XLS-69 simulate command.
// The server requires exactly one unsigned transaction as TxJSON or TxBlob and
// validates all request fields. The client only rejects nil requests.
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

// Validate only rejects a nil request. The server validates all request fields.
func (r *SimulateRequest) Validate() error {
	if r == nil {
		return ErrInvalidSimulateRequest
	}
	return nil
}

// ValidateNetworkID retains the nil-request check without inspecting NetworkID.
//
// Deprecated: the server validates NetworkID. Use Validate for the nil-request check.
func (r *SimulateRequest) ValidateNetworkID(_ *uint32) error {
	return r.Validate()
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
