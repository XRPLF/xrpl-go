package account_test

import (
	"encoding/json"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/require"
)

func TestAccountObjectsRequest(t *testing.T) {
	s := account.ObjectsRequest{
		Account:     "rsuHaTvJh1bDmDoxX9QcKP7HEBSBt4XsHx",
		Type:        account.SignerListObject,
		LedgerIndex: common.LedgerIndex(123),
	}

	j := `{
	"account": "rsuHaTvJh1bDmDoxX9QcKP7HEBSBt4XsHx",
	"type": "signer_list",
	"ledger_index": 123
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func accountObjectsResponseFixture() (account.ObjectsResponse, string) {
	payload := `{
  "account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
  "account_objects": [
    {
      "Balance": {
        "currency": "USD",
        "issuer": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
        "value": "100"
      },
      "Flags": 65536,
      "HighLimit": {
        "currency": "USD",
        "issuer": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
        "value": "0"
      },
      "LedgerEntryType": "RippleState",
      "LowLimit": {
        "currency": "USD",
        "issuer": "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
        "value": "500"
      }
    }
  ],
  "ledger_hash": "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
  "ledger_index": 71766343,
  "validated": true
}`
	expected := account.ObjectsResponse{
		Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		AccountObjects: []ledger.FlatLedgerObject{
			{
				"Balance": map[string]any{
					"currency": "USD",
					"issuer":   "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					"value":    "100",
				},
				"Flags": json.Number("65536"),
				"HighLimit": map[string]any{
					"currency": "USD",
					"issuer":   "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					"value":    "0",
				},
				"LedgerEntryType": "RippleState",
				"LowLimit": map[string]any{
					"currency": "USD",
					"issuer":   "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
					"value":    "500",
				},
			},
		},
		LedgerHash:  "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
		LedgerIndex: 71766343,
		Validated:   true,
	}
	return expected, payload
}

func TestAccountObjectsResponseSerialize(t *testing.T) {
	value, payload := accountObjectsResponseFixture()
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	require.JSONEq(t, payload, string(encoded))
}

func TestAccountObjectsResponseJSONDecode(t *testing.T) {
	want, payload := accountObjectsResponseFixture()
	var got account.ObjectsResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestAccountObjectsResponseClientDecode(t *testing.T) {
	want, payload := accountObjectsResponseFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got account.ObjectsResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}
