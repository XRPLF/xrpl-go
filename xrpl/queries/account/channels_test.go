package account

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/account/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/assert"
)

func TestAccountChannelRequest(t *testing.T) {
	s := ChannelsRequest{
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

func TestAccountChannelsResponse(t *testing.T) {
	s := ChannelsResponse{
		Account: "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
		Channels: []types.ChannelResult{
			{
				Account:            "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				Amount:             "100",
				Balance:            "200",
				ChannelID:          "500",
				DestinationAccount: "rnZvsWuLem5Ha46AZs61jLWR9R5esinkG3",
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
			"destination_account": "rnZvsWuLem5Ha46AZs61jLWR9R5esinkG3"
		}
	],
	"ledger_index": 123,
	"ledger_hash": "abc",
	"validated": true,
	"limit": 1
}`
	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestValidate(t *testing.T) {
	s := ChannelsRequest{
		Account: "",
	}

	err := s.Validate()

	assert.EqualError(t, err, "no account ID specified")
}
