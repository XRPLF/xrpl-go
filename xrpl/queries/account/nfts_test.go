package account

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
	s := NFTsResponse{
		Account: "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
		AccountNFTs: []accounttypes.NFT{
			{
				Flags:        accounttypes.Burnable | accounttypes.OnlyXRP,
				Issuer:       "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				NFTokenID:    "abc",
				NFTokenTaxon: 123,
				URI:          "def",
				NFTSerial:    456,
			},
		},
		LedgerIndex:        123,
		LedgerHash:         "abc",
		LedgerCurrentIndex: 1234,
		Validated:          true,
		Marker:             "abc",
		Limit:              123,
	}

	j := `{
	"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
	"account_nfts": [
		{
			"Flags": 3,
			"Issuer": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
			"NFTokenID": "abc",
			"NFTokenTaxon": 123,
			"URI": "def",
			"nft_serial": 456
		}
	],
	"ledger_index": 123,
	"ledger_hash": "abc",
	"ledger_current_index": 1234,
	"validated": true,
	"marker": "abc",
	"limit": 123
}`
	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}
