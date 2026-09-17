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

func ledgerClosedResponseFixture() (ledgerquery.ClosedResponse, string) {
	s := ledgerquery.ClosedResponse{
		LedgerHash:  "abc",
		LedgerIndex: 123,
	}
	j := `{
	"ledger_hash": "abc",
	"ledger_index": 123
}`
	return s, j
}

func TestLedgerClosedResponseSerialize(t *testing.T) {
	value, payload := ledgerClosedResponseFixture()
	require.NoError(t, testutil.Serialize(t, value, payload))
}

func TestLedgerClosedResponseJSONDecode(t *testing.T) {
	want, payload := ledgerClosedResponseFixture()
	var got ledgerquery.ClosedResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestLedgerClosedResponseClientDecode(t *testing.T) {
	want, payload := ledgerClosedResponseFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got ledgerquery.ClosedResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}
