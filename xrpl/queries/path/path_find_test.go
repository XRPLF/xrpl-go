package path

import (
	"encoding/json"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"

	pathtypes "github.com/Peersyst/xrpl-go/xrpl/queries/path/types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestPathFindCloseRequest(t *testing.T) {
	s := FindCloseRequest{
		Subcommand: Close,
	}

	j := `{
	"subcommand": "close"
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestPathFindStatusRequest(t *testing.T) {
	s := FindStatusRequest{
		Subcommand: Status,
	}

	j := `{
	"subcommand": "status"
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestFindCreateRequestWithDomain(t *testing.T) {
	domain := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	s := FindCreateRequest{
		Subcommand:         Create,
		SourceAccount:      "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		DestinationAccount: "rDCNEYQDfUYEgHjfLZ6CVHXNUCg6SdQgFN",
		DestinationAmount:  types.XRPCurrencyAmount(1000000),
		Domain:             &domain,
	}

	j := `{
	"subcommand": "create",
	"source_account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
	"destination_account": "rDCNEYQDfUYEgHjfLZ6CVHXNUCg6SdQgFN",
	"destination_amount": "1000000",
	"domain": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
}`

	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

type pathFindResponseFixture struct {
	name string
	want FindResponse
	json string
}

func pathFindResponseFixtures() []pathFindResponseFixture {
	return []pathFindResponseFixture{
		{
			name: "XRP source amount",
			want: FindResponse{
				Alternatives: []pathtypes.Alternative{
					{
						PathsComputed: [][]transaction.PathStep{
							{
								{
									Currency: "USD",
									Issuer:   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
									Account:  "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
								},
							},
						},
						SourceAmount: "207669",
						DestinationAmount: map[string]any{
							"currency": "USD",
							"issuer":   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
							"value":    "100",
						},
					},
				},
				DestinationAccount: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				DestinationAmount: map[string]any{
					"currency": "USD",
					"issuer":   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
					"value":    "100",
				},
				SourceAccount: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				FullReply:     true,
				Status:        true,
			},
			json: `{
  "alternatives": [
    {
      "paths_computed": [
        [
          {
            "currency": "USD",
            "issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
            "account": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B"
          }
        ]
      ],
      "source_amount": "207669",
      "destination_amount": {
        "currency": "USD",
        "issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
        "value": "100"
      }
    }
  ],
  "destination_account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
  "destination_amount": {
    "currency": "USD",
    "issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
    "value": "100"
  },
  "source_account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
  "full_reply": true,
  "status": true
}`,
		},
		{
			name: "issued source amount",
			want: FindResponse{
				Alternatives: []pathtypes.Alternative{
					{
						PathsComputed: [][]transaction.PathStep{
							{
								{
									Currency: "USD",
									Issuer:   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
									Account:  "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
								},
							},
						},
						SourceAmount: map[string]any{
							"currency": "USD",
							"issuer":   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
							"value":    "100",
						},
					},
				},
				SourceAccount: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				FullReply:     true,
				Status:        true,
			},
			json: `{
  "alternatives": [
    {
      "paths_computed": [
        [
          {
            "currency": "USD",
            "issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
            "account": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B"
          }
        ]
      ],
      "source_amount": {
        "currency": "USD",
        "issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
        "value": "100"
      }
    }
  ],
  "destination_account": "",
  "destination_amount": null,
  "source_account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
  "full_reply": true,
  "status": true
}`,
		},
		{
			name: "issued source amount",
			want: FindResponse{
				Alternatives:       []pathtypes.Alternative{},
				DestinationAccount: "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				DestinationAmount:  "100",
				SourceAccount:      "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				FullReply:          true,
				Closed:             true,
				Status:             true,
			},
			json: `{
  "alternatives": [],
  "closed": true,
  "destination_account": "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
  "destination_amount": "100",
  "full_reply": true,
  "source_account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
  "status": true
}`,
		},
	}
}

func TestPathFindResponseSerialize(t *testing.T) {
	for _, tt := range pathFindResponseFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.want)
			require.NoError(t, err)
			require.JSONEq(t, tt.json, string(encoded))
		})
	}
}

func TestPathFindResponseJSONDecode(t *testing.T) {
	for _, tt := range pathFindResponseFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			var got FindResponse
			decoder := json.NewDecoder(strings.NewReader(tt.json))
			decoder.UseNumber()
			require.NoError(t, decoder.Decode(&got))
			require.Equal(t, tt.want, got)
		})
	}
}

func TestPathFindResponseClientDecode(t *testing.T) {
	for _, tt := range pathFindResponseFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			var data map[string]any
			decoder := json.NewDecoder(strings.NewReader(tt.json))
			decoder.UseNumber()
			require.NoError(t, decoder.Decode(&data))
			var got FindResponse
			require.NoError(t, clientinternal.DecodeResultInto(data, &got))
			require.Equal(t, tt.want, got)
		})
	}
}
