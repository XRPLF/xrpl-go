package utility

import (
	"encoding/json"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"

	"github.com/stretchr/testify/require"
)

func pingResponseFixture() (PingResponse, string) {
	payload := `{
  "role": "admin",
  "unlimited": true
}`
	expected := PingResponse{
		Role:      "admin",
		Unlimited: true,
	}
	return expected, payload
}

func TestPingResponseSerialize(t *testing.T) {
	value, payload := pingResponseFixture()
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	require.JSONEq(t, payload, string(encoded))
}

func TestPingResponseJSONDecode(t *testing.T) {
	want, payload := pingResponseFixture()
	var got PingResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestPingResponseClientDecode(t *testing.T) {
	want, payload := pingResponseFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got PingResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}
