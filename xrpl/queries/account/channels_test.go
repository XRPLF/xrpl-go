package account_test

import (
	"encoding/json"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountChannelRequest(t *testing.T) {
	s := account.ChannelsRequest{
		Account:            "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
		DestinationAccount: "rnZvsWuLem5Ha46AZs61jLWR9R5esinkG3",
		LedgerIndex:        common.Validated,
	}

	j := `{
	"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
	"destination_account": "rnZvsWuLem5Ha46AZs61jLWR9R5esinkG3",
	"ledger_index": "validated"
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func accountChannelsResponseFixture() (account.ChannelsResponse, string) {
	s := account.ChannelsResponse{
		Account: "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
		Channels: []types.ChannelResult{
			{
				Account:            "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				Amount:             "100",
				Balance:            "200",
				ChannelID:          "500",
				DestinationAccount: "rnZvsWuLem5Ha46AZs61jLWR9R5esinkG3",
				PublicKey:          "aBR7mdD75Ycs8DRhMgQ4EMUEmBArF8SEh1hfjrT2V9DQTLNbJVqw",
				PublicKeyHex:       "03CFD18E689434F032A4E84C63E2A3A6472D684EAF4FD52CA67742F3E24BAE81B2",
				SettleDelay:        60,
			},
		},
		LedgerIndex: 123,
		LedgerHash:  "abc",
		Validated:   true,
		Limit:       1,
	}
	j := `{
	"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
	"channels": [
		{
			"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
			"amount": "100",
			"balance": "200",
			"channel_id": "500",
			"destination_account": "rnZvsWuLem5Ha46AZs61jLWR9R5esinkG3",
			"settle_delay": 60,
			"public_key": "aBR7mdD75Ycs8DRhMgQ4EMUEmBArF8SEh1hfjrT2V9DQTLNbJVqw",
			"public_key_hex": "03CFD18E689434F032A4E84C63E2A3A6472D684EAF4FD52CA67742F3E24BAE81B2"
		}
	],
	"ledger_index": 123,
	"ledger_hash": "abc",
	"validated": true,
	"limit": 1
}`
	return s, j
}

func TestAccountChannelsResponseSerialize(t *testing.T) {
	value, payload := accountChannelsResponseFixture()
	require.NoError(t, testutil.Serialize(t, value, payload))
}

func TestAccountChannelsResponseJSONDecode(t *testing.T) {
	want, payload := accountChannelsResponseFixture()
	var got account.ChannelsResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestAccountChannelsResponseClientDecode(t *testing.T) {
	want, payload := accountChannelsResponseFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got account.ChannelsResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}

func TestValidate(t *testing.T) {
	s := account.ChannelsRequest{
		Account: "",
	}

	err := s.Validate()

	assert.EqualError(t, err, "no account ID specified")
}
