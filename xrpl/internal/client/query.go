package client

import (
	"github.com/Peersyst/xrpl-go/pkg/decodehook"
	"github.com/go-viper/mapstructure/v2"
)

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
