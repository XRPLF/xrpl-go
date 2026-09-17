package account_test

import (
	"encoding/json"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	accounttypes "github.com/Peersyst/xrpl-go/xrpl/queries/account/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/require"
)

func TestAccountLinesRequest(t *testing.T) {
	s := account.LinesRequest{
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
	const address = "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu"

	tests := []struct {
		name     string
		request  account.LinesRequest
		expected string
	}{
		{
			name:    "present",
			request: account.LinesRequest{Account: address, IgnoreDefault: true},
			expected: `{
	"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
	"ignore_default": true
}`,
		},
		{
			name:    "omitted",
			request: account.LinesRequest{Account: address},
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

type accountLinesResponseFixture struct {
	name string
	want account.LinesResponse
	json string
}

func accountLinesResponseFixtures() []accountLinesResponseFixture {
	return []accountLinesResponseFixture{
		{
			name: "present",
			want: account.LinesResponse{
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
				Marker:             "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh,5",
				Limit:              10,
			},
			json: `{
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
	"marker": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh,5",
	"limit": 10
}`,
		},
		{
			name: "omitted",
			want: account.LinesResponse{
				Account: "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				Lines:   []accounttypes.TrustLine{},
			},
			json: `{
	"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
	"lines": []
}`,
		},
	}
}

func TestAccountLinesResponseSerialize(t *testing.T) {
	for _, tt := range accountLinesResponseFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testutil.Serialize(t, tt.want, tt.json))
		})
	}
}

func TestAccountLinesResponseJSONDecode(t *testing.T) {
	for _, tt := range accountLinesResponseFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			var got account.LinesResponse
			decoder := json.NewDecoder(strings.NewReader(tt.json))
			decoder.UseNumber()
			require.NoError(t, decoder.Decode(&got))
			require.Equal(t, tt.want, got)
		})
	}
}

func TestAccountLinesResponseClientDecode(t *testing.T) {
	for _, tt := range accountLinesResponseFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			var data map[string]any
			decoder := json.NewDecoder(strings.NewReader(tt.json))
			decoder.UseNumber()
			require.NoError(t, decoder.Decode(&data))
			var got account.LinesResponse
			require.NoError(t, clientinternal.DecodeResultInto(data, &got))
			require.Equal(t, tt.want, got)
		})
	}
}
