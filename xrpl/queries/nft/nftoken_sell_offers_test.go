package nft

import (
	"encoding/json"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"

	nfttypes "github.com/Peersyst/xrpl-go/xrpl/queries/nft/types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestNFTokenSellOffersRequest(t *testing.T) {
	s := NFTokenSellOffersRequest{
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

func TestNFTokenSellOffers_Pagination(t *testing.T) {
	const nftID = "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007"
	const marker = "9E28E366573187F8E5B85CE301F229E061A619EE5A589EF740088F8843BF10A1"
	const owner = "rLpSRZ1E8JHyNDZeHYsQs1R5cwDCB3uuZt"

	firstPage := NFTokenSellOffersResponse{
		NFTokenID: nftID,
		Offers:    makePaginationOffers(50, "1000", 1, owner),
		Limit:     50,
		Marker:    marker,
	}
	firstPageJSON, err := json.Marshal(firstPage)
	require.NoError(t, err)

	var decodedFirstPage NFTokenSellOffersResponse
	require.NoError(t, json.Unmarshal(firstPageJSON, &decodedFirstPage))
	require.Equal(t, firstPage, decodedFirstPage)

	firstPageFields := make(map[string]json.RawMessage)
	require.NoError(t, json.Unmarshal(firstPageJSON, &firstPageFields))
	require.Contains(t, firstPageFields, "limit")
	require.Contains(t, firstPageFields, "marker")

	continuation := NFTokenSellOffersRequest{
		NFTokenID: nftID,
		Limit:     decodedFirstPage.Limit,
		Marker:    decodedFirstPage.Marker,
	}
	continuationJSON := `{
	"nft_id": "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
	"limit": 50,
	"marker": "9E28E366573187F8E5B85CE301F229E061A619EE5A589EF740088F8843BF10A1"
}`
	require.NoError(t, testutil.SerializeAndDeserialize(t, continuation, continuationJSON))

	lastPage := NFTokenSellOffersResponse{
		NFTokenID: nftID,
		Offers: []nfttypes.NFTokenOffer{
			{
				Amount:            "1000",
				Flags:             1,
				NFTokenOfferIndex: marker,
				Owner:             owner,
			},
		},
	}
	lastPageJSON := `{
	"nft_id": "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
	"offers": [
		{
			"amount": "1000",
			"flags": 1,
			"nft_offer_index": "9E28E366573187F8E5B85CE301F229E061A619EE5A589EF740088F8843BF10A1",
			"owner": "rLpSRZ1E8JHyNDZeHYsQs1R5cwDCB3uuZt"
		}
	]
}`
	require.NoError(t, testutil.SerializeAndDeserialize(t, lastPage, lastPageJSON))
}

func nftSellOffersResponseFixture() (NFTokenSellOffersResponse, string) {
	s := NFTokenSellOffersResponse{
		NFTokenID: "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
		Offers: []nfttypes.NFTokenOffer{
			{
				Amount:            "1000",
				Flags:             1,
				NFTokenOfferIndex: "9E28E366573187F8E5B85CE301F229E061A619EE5A589EF740088F8843BF10A1",
				Owner:             "rLpSRZ1E8JHyNDZeHYsQs1R5cwDCB3uuZt",
			},
		},
	}
	j := `{
	"nft_id": "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
	"offers": [
		{
			"amount": "1000",
			"flags": 1,
			"nft_offer_index": "9E28E366573187F8E5B85CE301F229E061A619EE5A589EF740088F8843BF10A1",
			"owner": "rLpSRZ1E8JHyNDZeHYsQs1R5cwDCB3uuZt"
		}
	]
}`
	return s, j
}

func TestNFTokenSellOffersResponseSerialize(t *testing.T) {
	value, payload := nftSellOffersResponseFixture()
	value.Offers[0].Amount = types.XRPCurrencyAmount(1000)
	require.NoError(t, testutil.Serialize(t, value, payload))
}

func TestNFTokenSellOffersResponseJSONDecode(t *testing.T) {
	want, payload := nftSellOffersResponseFixture()
	var got NFTokenSellOffersResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestNFTokenSellOffersResponseClientDecode(t *testing.T) {
	want, payload := nftSellOffersResponseFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got NFTokenSellOffersResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}
