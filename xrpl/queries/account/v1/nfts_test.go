package v1

import (
	"testing"

	accounttypes "github.com/Peersyst/xrpl-go/xrpl/queries/account/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

func TestAccountNFTsRequest(t *testing.T) {
	s := NFTsRequest{
		Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
		LedgerIndex: common.Validated,
		LedgerHash:  "123",
		Limit:       2,
	}

	j := `{
	"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
	"ledger_index": "validated",
	"ledger_hash": "123",
	"limit": 2
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestAccountNFTsResponse(t *testing.T) {
	tests := []struct {
		name     string
		response NFTsResponse
		expected string
	}{
		{
			name: "validated ledger",
			response: NFTsResponse{
				Account: "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				AccountNFTs: []accounttypes.NFT{
					{
						Flags:        accounttypes.Burnable | accounttypes.OnlyXRP,
						Issuer:       "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
						NFTokenID:    "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
						NFTokenTaxon: 123,
						URI:          "def",
						NFTSerial:    456,
					},
				},
				LedgerIndex: 123,
				LedgerHash:  "4C99E5F63C0D0B1C2283B4F5DCE2239F80CE92E8B1A6AED1E110C198FC96E659",
				Validated:   true,
				Marker:      "abc",
				Limit:       123,
			},
			expected: `{
	"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
	"account_nfts": [
		{
			"Flags": 3,
			"Issuer": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
			"NFTokenID": "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
			"NFTokenTaxon": 123,
			"URI": "def",
			"nft_serial": 456
		}
	],
	"ledger_index": 123,
	"ledger_hash": "4C99E5F63C0D0B1C2283B4F5DCE2239F80CE92E8B1A6AED1E110C198FC96E659",
	"validated": true,
	"marker": "abc",
	"limit": 123
}`,
		},
		{
			name: "open ledger",
			response: NFTsResponse{
				Account:            "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				AccountNFTs:        []accounttypes.NFT{},
				LedgerCurrentIndex: 1234,
				Validated:          false,
			},
			expected: `{
	"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
	"account_nfts": [],
	"ledger_current_index": 1234,
	"validated": false
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
