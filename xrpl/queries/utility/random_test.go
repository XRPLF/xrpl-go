package utility

import (
	"encoding/json"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/require"
)

func randomResponseFixture() (RandomResponse, string) {
	s := RandomResponse{
		Random: "8ED765AEBBD6767603C2C9375B2679AEC76E6A8133EF59F04F9FC1AAA70E41AF",
	}
	j := `{
	"random": "8ED765AEBBD6767603C2C9375B2679AEC76E6A8133EF59F04F9FC1AAA70E41AF"
}`
	return s, j
}

func TestRandomResponseSerialize(t *testing.T) {
	value, payload := randomResponseFixture()
	require.NoError(t, testutil.Serialize(t, value, payload))
}

func TestRandomResponseJSONDecode(t *testing.T) {
	want, payload := randomResponseFixture()
	var got RandomResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestRandomResponseClientDecode(t *testing.T) {
	want, payload := randomResponseFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got RandomResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}
