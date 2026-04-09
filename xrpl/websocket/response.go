package websocket

import (
	"encoding/json"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

var jsonUnmarshalerType = reflect.TypeFor[json.Unmarshaler]()

// jsonUnmarshalerHookFunc returns a mapstructure decode hook that delegates to
// json.Unmarshaler for any target type that implements it. The source value is
// re-encoded to JSON and then passed to the target's UnmarshalJSON method, which
// lets individual types define their own flexible decoding logic (e.g. accepting
// both strings and numbers for a field that is modelled as a string in Go).
func jsonUnmarshalerHookFunc() mapstructure.DecodeHookFuncValue {
	return func(from reflect.Value, to reflect.Value) (any, error) {
		toType := to.Type()
		if toType.Kind() != reflect.Ptr {
			toType = reflect.PointerTo(toType)
		}
		if !toType.Implements(jsonUnmarshalerType) {
			return from.Interface(), nil
		}

		// Re-encode the source (a map[string]any or similar) to JSON bytes.
		jsonBytes, err := json.Marshal(from.Interface())
		if err != nil {
			return nil, err
		}

		// Allocate a new value of the concrete (non-pointer) target type and
		// unmarshal into it via the custom UnmarshalJSON.
		target := reflect.New(to.Type())
		if err := json.Unmarshal(jsonBytes, target.Interface()); err != nil {
			return nil, err
		}
		return target.Elem().Interface(), nil
	}
}

// ResponseWarning represents a warning returned in a WebSocket response.
type ResponseWarning struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// ErrorWebsocketClientXrplResponse represents an error returned by the XRPL WebSocket client.
type ErrorWebsocketClientXrplResponse struct {
	Type    string
	Request map[string]any
}

// Error returns the error type string for the WebSocket client XRPL response.
func (e *ErrorWebsocketClientXrplResponse) Error() string {
	return e.Type
}

// ClientResponse represents a generic XRPL WebSocket client response, including status, result, and warnings.
type ClientResponse struct {
	ID        int               `json:"id"`
	Status    string            `json:"status"`
	Type      string            `json:"type"`
	Error     string            `json:"error,omitempty"`
	Result    map[string]any    `json:"result,omitempty"`
	Value     map[string]any    `json:"value,omitempty"`
	Warning   string            `json:"warning,omitempty"`
	Warnings  []ResponseWarning `json:"warnings,omitempty"`
	Forwarded bool              `json:"forwarded,omitempty"`
}

// GetResult decodes the Result field into the provided variable v using mapstructure.
func (r *ClientResponse) GetResult(v any) error {
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &v,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			jsonUnmarshalerHookFunc(),
			mapstructure.TextUnmarshallerHookFunc(),
		),
	})
	if err != nil {
		return err
	}
	return dec.Decode(r.Result)
}

// CheckError checks if the response contains an error and returns an ErrorWebsocketClientXrplResponse if found.
func (r *ClientResponse) CheckError() error {
	if r.Error != "" {
		return &ErrorWebsocketClientXrplResponse{
			Type:    r.Error,
			Request: r.Value,
		}
	}
	return nil
}
