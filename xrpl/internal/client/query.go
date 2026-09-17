package client

import (
	"context"

	"github.com/Peersyst/xrpl-go/pkg/decodehook"
	"github.com/go-viper/mapstructure/v2"
)

// Request is the request contract shared client helpers need. Its method set
// is the union of the RPC and WebSocket client request interfaces, so a value of
// this type is assignable to either client's request parameter.
type Request interface {
	Method() string
	Validate() error
	APIVersion() int
	SetAPIVersion(apiVersion int)
}

// RequestResultFunc issues a request over a client's transport and decodes the
// response result into result.
type RequestResultFunc func(ctx context.Context, req Request, result any) error

// ResponseDecoder decodes a transport response payload into a typed result.
// Both the RPC and WebSocket client responses satisfy it.
type ResponseDecoder interface {
	GetResult(v any) error
}

// DecodeResult handles a query's request error and decodes its response.
// It returns the original error and no result if either step fails, including
// when decoding partially populated the result. A successful request must
// provide a response decoder.
func DecodeResult[T any](response ResponseDecoder, requestErr error) (*T, error) {
	if requestErr != nil {
		return nil, requestErr
	}
	var result T
	if err := response.GetResult(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DecodeResultInto decodes an unpacked query result using JSON field names and
// the JSON and text unmarshaling hooks shared by both client transports.
func DecodeResultInto(data any, result any) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &result,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			decodehook.JSON(),
			mapstructure.TextUnmarshallerHookFunc(),
		),
	})
	if err != nil {
		return err
	}
	return decoder.Decode(data)
}
