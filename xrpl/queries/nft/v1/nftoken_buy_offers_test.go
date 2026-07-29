package v1

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	nfttypes "github.com/Peersyst/xrpl-go/xrpl/queries/nft/types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

func TestNFTokenBuyOffersRequest(t *testing.T) {
	s := NFTokenBuyOffersRequest{
		NFTokenID:   "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
		LedgerIndex: common.Validated,
	}

	j := `{
	"nft_id": "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
	"ledger_index": "validated"
}`

	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestNFTokenBuyOffers_Pagination(t *testing.T) {
	const nftID = "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007"
	const marker = "3212D26DB00031889D4EF7D9129BB0FA673B5B40B1759564486C0F0946BA203F"

	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{
			name: "first page response",
			value: NFTokenBuyOffersResponse{
				NFTokenID: nftID,
				Offers:    []nfttypes.NFTokenOffer{},
				Limit:     50,
				Marker:    marker,
			},
			expected: `{
	"nft_id": "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
	"offers": [],
	"limit": 50,
	"marker": "3212D26DB00031889D4EF7D9129BB0FA673B5B40B1759564486C0F0946BA203F"
}`,
		},
		{
			name: "continuation request",
			value: NFTokenBuyOffersRequest{
				NFTokenID: nftID,
				Limit:     50,
				Marker:    marker,
			},
			expected: `{
	"nft_id": "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
	"limit": 50,
	"marker": "3212D26DB00031889D4EF7D9129BB0FA673B5B40B1759564486C0F0946BA203F"
}`,
		},
		{
			name: "last page response without marker",
			value: NFTokenBuyOffersResponse{
				NFTokenID: nftID,
				Offers:    []nfttypes.NFTokenOffer{},
				Limit:     50,
			},
			expected: `{
	"nft_id": "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
	"offers": [],
	"limit": 50
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := testutil.SerializeAndDeserialize(t, tt.value, tt.expected); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestNFTokenBuyOffersResponse(t *testing.T) {
	s := NFTokenBuyOffersResponse{
		NFTokenID: "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
		Offers: []nfttypes.NFTokenOffer{
			{
				Amount:            types.XRPCurrencyAmount(1500),
				Flags:             0,
				NFTokenOfferIndex: "3212D26DB00031889D4EF7D9129BB0FA673B5B40B1759564486C0F0946BA203F",
				Owner:             "rsuHaTvJh1bDmDoxX9QcKP7HEBSBt4XsHx",
			},
		},
	}

	j := `{
	"nft_id": "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
	"offers": [
		{
			"amount": "1500",
			"flags": 0,
			"nft_offer_index": "3212D26DB00031889D4EF7D9129BB0FA673B5B40B1759564486C0F0946BA203F",
			"owner": "rsuHaTvJh1bDmDoxX9QcKP7HEBSBt4XsHx"
		}
	]
}`

	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}
