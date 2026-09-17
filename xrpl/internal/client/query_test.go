package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestDecodeResult(t *testing.T) {
	decodeErr := errors.New("decode failed")
	for _, tt := range []struct {
		name      string
		value     int
		decodeErr error
	}{
		{name: "decoded value", value: 42},
		{name: "zero value still returns a result"},
		{name: "partial result discarded on decode failure", value: 42, decodeErr: fmt.Errorf("field: %w", decodeErr)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			response := responseDecoderFunc(func(v any) error {
				calls++
				*v.(*int) = tt.value
				return tt.decodeErr
			})
			result, err := DecodeResult[int](response, nil)
			require.Equal(t, 1, calls)
			if tt.decodeErr != nil {
				require.Same(t, tt.decodeErr, err)
				require.ErrorIs(t, err, decodeErr)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tt.value, *result)
			}
		})
	}
}

func TestDecodeResultInto(t *testing.T) {
	// JSON tags and both hooks must survive moving decoding out of the transports.
	type response struct {
		Index uint32                  `json:"ledger_index"`
		XRP   types.XRPCurrencyAmount `json:"xrp_amount"`
		MPT   types.MPTPlainAmount    `json:"mpt_amount"`
	}
	for _, tt := range []struct {
		name   string
		number any
	}{
		{name: "RPC number", number: json.Number("123")},
		{name: "WebSocket number", number: float64(123)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var result response
			err := DecodeResultInto(map[string]any{
				"ledger_index": tt.number,
				"xrp_amount":   "9007199254740993",
				"mpt_amount":   "42",
			}, &result)
			require.NoError(t, err)
			require.Equal(t, response{Index: 123, XRP: 9007199254740993, MPT: 42}, result)
		})
	}
}

func TestDecodeResultIntoRejectsInvalidFields(t *testing.T) {
	for _, tt := range []struct {
		name  string
		value any
	}{
		{name: "wrong type", value: true},
		{name: "invalid amount text", value: "not an amount"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var result struct {
				Amount types.XRPCurrencyAmount `json:"amount"`
			}
			err := DecodeResultInto(map[string]any{"amount": tt.value}, &result)
			require.Error(t, err)
		})
	}
}

func TestDecodeResultRequestError(t *testing.T) {
	requestErr := errors.New("request failed")
	// The request error takes precedence even if a transport supplies a response.
	decodeCalls := 0
	unexpected := responseDecoderFunc(func(any) error {
		decodeCalls++
		return errors.New("unexpected decode")
	})
	for _, tt := range []struct {
		name     string
		response ResponseDecoder
	}{
		{name: "no response"},
		{name: "response ignored", response: unexpected},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodeResult[int](tt.response, requestErr)
			require.Same(t, requestErr, err)
			require.Nil(t, result)
			require.Zero(t, decodeCalls)
		})
	}
}
