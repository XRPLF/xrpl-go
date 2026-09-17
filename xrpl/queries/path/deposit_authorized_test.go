package path

import (
	"encoding/json"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/require"
)

func TestDepositAuthorizedRequest(t *testing.T) {
	s := DepositAuthorizedRequest{
		SourceAccount:      "rEhxGqkqPPSxQ3P25J66ft5TwpzV14k2de",
		DestinationAccount: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		LedgerIndex:        common.Validated,
	}

	j := `{
	"source_account": "rEhxGqkqPPSxQ3P25J66ft5TwpzV14k2de",
	"destination_account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"ledger_index": "validated"
}`

	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func depositAuthorizedResponseFixture() (DepositAuthorizedResponse, string) {
	s := DepositAuthorizedResponse{
		DepositAuthorized:  true,
		DestinationAccount: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		LedgerHash:         "BD03A10653ED9D77DCA859B7A735BF0580088A8F287FA2C5403E0A19C58EF322",
		LedgerIndex:        8,
		LedgerCurrentIndex: 9,
		SourceAccount:      "rEhxGqkqPPSxQ3P25J66ft5TwpzV14k2de",
		Validated:          true,
	}
	j := `{
	"deposit_authorized": true,
	"destination_account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"ledger_hash": "BD03A10653ED9D77DCA859B7A735BF0580088A8F287FA2C5403E0A19C58EF322",
	"ledger_index": 8,
	"ledger_current_index": 9,
	"source_account": "rEhxGqkqPPSxQ3P25J66ft5TwpzV14k2de",
	"validated": true
}`
	return s, j
}

func TestDepositAuthorizedResponseSerialize(t *testing.T) {
	value, payload := depositAuthorizedResponseFixture()
	require.NoError(t, testutil.Serialize(t, value, payload))
}

func TestDepositAuthorizedResponseJSONDecode(t *testing.T) {
	want, payload := depositAuthorizedResponseFixture()
	var got DepositAuthorizedResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestDepositAuthorizedResponseClientDecode(t *testing.T) {
	want, payload := depositAuthorizedResponseFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got DepositAuthorizedResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}
