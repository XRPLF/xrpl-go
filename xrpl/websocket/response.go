package websocket

import (
	"github.com/Peersyst/xrpl-go/pkg/decodehook"
	"github.com/go-viper/mapstructure/v2"
)

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
			decodehook.JSON(),
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
