package ledger_test

import (
	"encoding/json"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	ledgerquery "github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	ledgertypes "github.com/Peersyst/xrpl-go/xrpl/queries/ledger/types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/require"
)

func TestLedgerDataRequest(t *testing.T) {
	s := ledgerquery.DataRequest{
		LedgerIndex: common.Closed,
		Binary:      true,
		Limit:       5,
	}
	j := `{
	"ledger_index": "closed",
	"binary": true,
	"limit": 5
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func ledgerDataResponseFixture() (ledgerquery.DataResponse, string) {
	payload := `{
  "ledger_hash": "842B57C1CC0613299A686D3E9F310EC0422C84D3911E5056389AA7E5808A93C8",
  "ledger_index": "6",
  "marker": null,
  "state": [
    {
      "data": "0000000000000000",
      "index": "1B8590C01B0006EDFA9ED60296DD052DC5E90F99147FE0D93"
    }
  ]
}`
	expected := ledgerquery.DataResponse{
		LedgerHash:  "842B57C1CC0613299A686D3E9F310EC0422C84D3911E5056389AA7E5808A93C8",
		LedgerIndex: "6",
		State: []ledgertypes.State{
			{
				Data:  "0000000000000000",
				Index: "1B8590C01B0006EDFA9ED60296DD052DC5E90F99147FE0D93",
			},
		},
	}
	return expected, payload
}

func TestLedgerDataResponseSerialize(t *testing.T) {
	value, payload := ledgerDataResponseFixture()
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	require.JSONEq(t, payload, string(encoded))
}

func TestLedgerDataResponseJSONDecode(t *testing.T) {
	want, payload := ledgerDataResponseFixture()
	var got ledgerquery.DataResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestLedgerDataResponseClientDecode(t *testing.T) {
	want, payload := ledgerDataResponseFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got ledgerquery.DataResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}
