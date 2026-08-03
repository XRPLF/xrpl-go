package nft

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	nfttypes "github.com/Peersyst/xrpl-go/xrpl/queries/nft/types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
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
	const owner = "rsuHaTvJh1bDmDoxX9QcKP7HEBSBt4XsHx"

	firstPage := NFTokenBuyOffersResponse{
		NFTokenID: nftID,
		Offers:    makePaginationOffers(50, "1500", 0, owner),
		Limit:     50,
		Marker:    marker,
	}
	firstPageJSON, err := json.Marshal(firstPage)
	require.NoError(t, err)

	var decodedFirstPage NFTokenBuyOffersResponse
	require.NoError(t, json.Unmarshal(firstPageJSON, &decodedFirstPage))
	require.Equal(t, firstPage, decodedFirstPage)

	firstPageFields := make(map[string]json.RawMessage)
	require.NoError(t, json.Unmarshal(firstPageJSON, &firstPageFields))
	require.Contains(t, firstPageFields, "limit")
	require.Contains(t, firstPageFields, "marker")

	continuation := NFTokenBuyOffersRequest{
		NFTokenID: nftID,
		Limit:     decodedFirstPage.Limit,
		Marker:    decodedFirstPage.Marker,
	}
	continuationJSON := `{
	"nft_id": "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
	"limit": 50,
	"marker": "3212D26DB00031889D4EF7D9129BB0FA673B5B40B1759564486C0F0946BA203F"
}`
	require.NoError(t, testutil.SerializeAndDeserialize(t, continuation, continuationJSON))

	lastPage := NFTokenBuyOffersResponse{
		NFTokenID: nftID,
		Offers: []nfttypes.NFTokenOffer{
			{
				Amount:            "1500",
				Flags:             0,
				NFTokenOfferIndex: marker,
				Owner:             owner,
			},
		},
	}
	lastPageJSON := `{
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
	require.NoError(t, testutil.SerializeAndDeserialize(t, lastPage, lastPageJSON))
}

func makePaginationOffers(count int, amount string, flags uint, owner types.Address) []nfttypes.NFTokenOffer {
	offers := make([]nfttypes.NFTokenOffer, count)
	for i := range offers {
		offers[i] = nfttypes.NFTokenOffer{
			Amount:            amount,
			Flags:             flags,
			NFTokenOfferIndex: fmt.Sprintf("%064X", i+1),
			Owner:             owner,
		}
	}
	return offers
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
