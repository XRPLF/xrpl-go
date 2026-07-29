package account

import (
	"testing"

	accounttypes "github.com/Peersyst/xrpl-go/xrpl/queries/account/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

func TestAccountLinesRequest(t *testing.T) {
	s := LinesRequest{
		Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
		Peer:        "rnZvsWuLem5Ha46AZs61jLWR9R5esinkG3",
		LedgerHash:  "1EDBBA3C793863366DF5B31C2174B6B5E6DF6DB89A7212B86838489148E2A581",
		LedgerIndex: common.LedgerIndex(256),
		Limit:       10,
		Marker:      map[string]any{"abc": "def"},
	}

	j := `{
	"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
	"ledger_hash": "1EDBBA3C793863366DF5B31C2174B6B5E6DF6DB89A7212B86838489148E2A581",
	"ledger_index": 256,
	"peer": "rnZvsWuLem5Ha46AZs61jLWR9R5esinkG3",
	"limit": 10,
	"marker": {
		"abc": "def"
	}
}`

	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestAccountLinesRequest_IgnoreDefault(t *testing.T) {
	const account = "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu"

	tests := []struct {
		name     string
		request  LinesRequest
		expected string
	}{
		{
			name:    "present",
			request: LinesRequest{Account: account, IgnoreDefault: true},
			expected: `{
	"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
	"ignore_default": true
}`,
		},
		{
			name:    "omitted",
			request: LinesRequest{Account: account},
			expected: `{
	"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu"
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := testutil.SerializeAndDeserialize(t, tt.request, tt.expected); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestAccountLinesResponse(t *testing.T) {
	tests := []struct {
		name     string
		response LinesResponse
		expected string
	}{
		{
			name: "present",
			response: LinesResponse{
				Account: "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				Lines: []accounttypes.TrustLine{
					{
						Account:    "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
						Balance:    "123",
						Currency:   "USD",
						Limit:      "456",
						LimitPeer:  "10",
						QualityIn:  1,
						QualityOut: 2,
					},
				},
				LedgerCurrentIndex: 123,
				LedgerIndex:        345,
				LedgerHash:         "abc",
				Limit:              10,
			},
			expected: `{
	"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
	"lines": [
		{
			"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
			"balance": "123",
			"currency": "USD",
			"limit": "456",
			"limit_peer": "10",
			"quality_in": 1,
			"quality_out": 2
		}
	],
	"ledger_current_index": 123,
	"ledger_index": 345,
	"ledger_hash": "abc",
	"limit": 10
}`,
		},
		{
			name: "omitted",
			response: LinesResponse{
				Account: "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				Lines:   []accounttypes.TrustLine{},
			},
			expected: `{
	"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
	"lines": []
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := testutil.SerializeAndDeserialize(t, tt.response, tt.expected); err != nil {
				t.Error(err)
			}
		})
	}
}
