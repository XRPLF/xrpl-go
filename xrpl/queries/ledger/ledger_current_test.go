package ledger_test

import (
	"encoding/json"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	ledgerquery "github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/require"
)

func ledgerCurrentResponseFixture() (ledgerquery.CurrentResponse, string) {
	s := ledgerquery.CurrentResponse{
		LedgerCurrentIndex: 123,
	}
	j := `{
	"ledger_current_index": 123
}`
	return s, j
}

func TestLedgerCurrentResponseSerialize(t *testing.T) {
	value, payload := ledgerCurrentResponseFixture()
	require.NoError(t, testutil.Serialize(t, value, payload))
}

func TestLedgerCurrentResponseJSONDecode(t *testing.T) {
	want, payload := ledgerCurrentResponseFixture()
	var got ledgerquery.CurrentResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestLedgerCurrentResponseClientDecode(t *testing.T) {
	want, payload := ledgerCurrentResponseFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got ledgerquery.CurrentResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}
