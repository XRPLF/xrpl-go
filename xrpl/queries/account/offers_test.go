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

func TestAccountOffersRequest(t *testing.T) {
	s := account.OffersRequest{
		Account:     "abc",
		LedgerIndex: common.LedgerIndex(10),
		Marker:      "123",
	}
	j := `{
	"account": "abc",
	"ledger_index": 10,
	"marker": "123"
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

type accountOffersResponseFixture struct {
	name string
	want account.OffersResponse
	json string
}

func accountOffersResponseFixtures() []accountOffersResponseFixture {
	return []accountOffersResponseFixture{
		{
			name: "issued taker gets",
			want: account.OffersResponse{
				Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				Offers: []accounttypes.OfferResult{
					{
						Flags:    0,
						Quality:  "1",
						Sequence: 1234,
						TakerGets: map[string]any{
							"currency": "USD",
							"issuer":   "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
							"value":    "100",
						},
						TakerPays: "100000000",
					},
				},
				LedgerHash:  "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
				LedgerIndex: 71766343,
			},
			json: `{
  "account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
  "offers": [
    {
      "flags": 0,
      "quality": "1",
      "seq": 1234,
      "taker_gets": {
        "currency": "USD",
        "issuer": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
        "value": "100"
      },
      "taker_pays": "100000000"
    }
  ],
  "ledger_hash": "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
  "ledger_index": 71766343
}`,
		},
		{
			name: "XRP taker gets",
			want: account.OffersResponse{
				Account: "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
				Offers: []accounttypes.OfferResult{
					{
						Flags:     0,
						Sequence:  1337,
						TakerGets: "100000000",
						TakerPays: map[string]any{
							"currency": "USD",
							"issuer":   "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
							"value":    "100",
						},
					},
				},
				LedgerCurrentIndex: 14380380,
			},
			json: `{
  "account": "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
  "ledger_current_index": 14380380,
  "offers": [
    {
      "flags": 0,
      "quality": "",
      "seq": 1337,
      "taker_gets": "100000000",
      "taker_pays": {
        "currency": "USD",
        "issuer": "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
        "value": "100"
      }
    }
  ]
}`,
		},
	}
}

func TestAccountOffersResponseSerialize(t *testing.T) {
	for _, tt := range accountOffersResponseFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.want)
			require.NoError(t, err)
			require.JSONEq(t, tt.json, string(encoded))
		})
	}
}

func TestAccountOffersResponseJSONDecode(t *testing.T) {
	for _, tt := range accountOffersResponseFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			var got account.OffersResponse
			decoder := json.NewDecoder(strings.NewReader(tt.json))
			decoder.UseNumber()
			require.NoError(t, decoder.Decode(&got))
			require.Equal(t, tt.want, got)
		})
	}
}

func TestAccountOffersResponseClientDecode(t *testing.T) {
	for _, tt := range accountOffersResponseFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			var data map[string]any
			decoder := json.NewDecoder(strings.NewReader(tt.json))
			decoder.UseNumber()
			require.NoError(t, decoder.Decode(&data))
			var got account.OffersResponse
			require.NoError(t, clientinternal.DecodeResultInto(data, &got))
			require.Equal(t, tt.want, got)
		})
	}
}
