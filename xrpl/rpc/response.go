package rpc

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/mitchellh/mapstructure"
)

// Response represents a JSON-RPC response from an XRPL server.
type Response struct {
	Result    AnyJSON               `json:"result"`
	Warning   string                `json:"warning,omitempty"`
	Warnings  []XRPLResponseWarning `json:"warnings,omitempty"`
	Forwarded bool                  `json:"forwarded,omitempty"`
}

// XRPLResponseWarning represents a warning returned by the XRPL server.
type XRPLResponseWarning struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// AnyJSON is an alias for transaction.FlatTransaction used for generic JSON result data.
type AnyJSON transaction.FlatTransaction

// APIWarning represents a warning from the API with optional details.
type APIWarning struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// GetResult decodes the RPC response result into the provided value using mapstructure.
func (r Response) GetResult(v any) error {
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &v, DecodeHook: mapstructure.TextUnmarshallerHookFunc(),
	})
	if err != nil {
		return err
	}
	err = dec.Decode(r.Result)
	if err != nil {
		return err
	}
	return nil
}

// XRPLResponse defines the interface for types that can extract a result from an RPC response.
type XRPLResponse interface {
	GetResult(v any) error
}
