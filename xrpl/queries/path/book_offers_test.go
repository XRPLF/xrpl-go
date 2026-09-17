package path

import (
	"encoding/json"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"

	pathtypes "github.com/Peersyst/xrpl-go/xrpl/queries/path/types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestBookOffersRequest(t *testing.T) {
	s := BookOffersRequest{
		Taker: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		TakerGets: pathtypes.BookOfferCurrency{
			Currency: "XRP",
		},
		TakerPays: pathtypes.BookOfferCurrency{
			Currency: "USD",
			Issuer:   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
		},
		Limit: 10,
	}
	j := `{
	"taker_gets": {
		"currency": "XRP"
	},
	"taker_pays": {
		"currency": "USD",
		"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B"
	},
	"taker": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
	"limit": 10
}`

	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestBookOffersRequestWithDomain(t *testing.T) {
	domain := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	s := BookOffersRequest{
		Taker: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		TakerGets: pathtypes.BookOfferCurrency{
			Currency: "XRP",
		},
		TakerPays: pathtypes.BookOfferCurrency{
			Currency: "USD",
			Issuer:   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
		},
		Limit:  10,
		Domain: &domain,
	}
	j := `{
	"taker_gets": {
		"currency": "XRP"
	},
	"taker_pays": {
		"currency": "USD",
		"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B"
	},
	"taker": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
	"limit": 10,
	"domain": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
}`

	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func bookOffersResponseFixture() (BookOffersResponse, string) {
	s := BookOffersResponse{
		LedgerCurrentIndex: 7035305,
		LedgerIndex:        7035305,
		LedgerHash:         "123",
		Validated:          true,
		Offers: []pathtypes.BookOffer{
			{
				Account:           "rM3X3QSr8icjTGpaF52dozhbT2BZSXJQYM",
				BookDirectory:     "7E5F614417C2D0A7CEFEB73C4AA773ED5B078DE2B5771F6D55055E4C405218EB",
				BookNode:          "0000000000000000",
				Flags:             0,
				LedgerEntryType:   ledger.OfferEntry,
				OwnerNode:         "0000000000000AE0",
				PreviousTxnID:     "6956221794397C25A53647182E5C78A439766D600724074C99D78982E37599F1",
				PreviousTxnLgrSeq: 7022646,
				Sequence:          264542,
				TakerGets:         map[string]any{"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B", "currency": "EUR", "value": "17.90363633316433"},
				TakerPays:         map[string]any{"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B", "currency": "USD", "value": "27.05340557506234"},
				Quality:           "1.511056473200875",
				OwnerFunds:        "100",
				TakerGetsFunded:   "1000000",
				TakerPaysFunded:   "100",
			},
			{
				Account:           "rhsxKNyN99q6vyYCTHNTC1TqWCeHr7PNgp",
				BookDirectory:     "7E5F614417C2D0A7CEFEB73C4AA773ED5B078DE2B5771F6D5505DCAA8FE12000",
				BookNode:          "0000000000000000",
				Flags:             131072,
				LedgerEntryType:   ledger.OfferEntry,
				OwnerNode:         "0000000000000001",
				PreviousTxnID:     "8AD748CD489F7FF34FCD4FB73F77F1901E27A6EFA52CCBB0CCDAAB934E5E754D",
				PreviousTxnLgrSeq: 7007546,
				Sequence:          265,
				TakerGets:         map[string]any{"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B", "currency": "EUR", "value": "2.542743233917848"},
				TakerPays:         map[string]any{"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B", "currency": "USD", "value": "4.19552633596446"},
				Quality:           "1.65",
			},
		},
	}
	j := `{
	"ledger_current_index": 7035305,
	"ledger_index": 7035305,
	"ledger_hash": "123",
	"offers": [
		{
			"Flags": 0,
			"LedgerEntryType": "Offer",
			"Account": "rM3X3QSr8icjTGpaF52dozhbT2BZSXJQYM",
			"BookDirectory": "7E5F614417C2D0A7CEFEB73C4AA773ED5B078DE2B5771F6D55055E4C405218EB",
			"BookNode": "0000000000000000",
			"OwnerNode": "0000000000000AE0",
			"PreviousTxnID": "6956221794397C25A53647182E5C78A439766D600724074C99D78982E37599F1",
			"PreviousTxnLgrSeq": 7022646,
			"Sequence": 264542,
			"TakerPays": {
				"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
				"currency": "USD",
				"value": "27.05340557506234"
			},
			"TakerGets": {
				"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
				"currency": "EUR",
				"value": "17.90363633316433"
			},
			"owner_funds": "100",
			"taker_gets_funded": "1000000",
			"taker_pays_funded": "100",
			"quality": "1.511056473200875"
		},
		{
			"Flags": 131072,
			"LedgerEntryType": "Offer",
			"Account": "rhsxKNyN99q6vyYCTHNTC1TqWCeHr7PNgp",
			"BookDirectory": "7E5F614417C2D0A7CEFEB73C4AA773ED5B078DE2B5771F6D5505DCAA8FE12000",
			"BookNode": "0000000000000000",
			"OwnerNode": "0000000000000001",
			"PreviousTxnID": "8AD748CD489F7FF34FCD4FB73F77F1901E27A6EFA52CCBB0CCDAAB934E5E754D",
			"PreviousTxnLgrSeq": 7007546,
			"Sequence": 265,
			"TakerPays": {
				"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
				"currency": "USD",
				"value": "4.19552633596446"
			},
			"TakerGets": {
				"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
				"currency": "EUR",
				"value": "2.542743233917848"
			},
			"quality": "1.65"
		}
	],
	"validated": true
}`
	return s, j
}

func TestBookOffersResponseSerialize(t *testing.T) {
	value, payload := bookOffersResponseFixture()
	value.Offers[0].TakerGets = types.IssuedCurrencyAmount{
		Currency: "EUR",
		Issuer:   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
		Value:    "17.90363633316433",
	}
	value.Offers[0].TakerPays = types.IssuedCurrencyAmount{
		Currency: "USD",
		Issuer:   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
		Value:    "27.05340557506234",
	}
	value.Offers[1].TakerGets = types.IssuedCurrencyAmount{
		Currency: "EUR",
		Issuer:   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
		Value:    "2.542743233917848",
	}
	value.Offers[1].TakerPays = types.IssuedCurrencyAmount{
		Currency: "USD",
		Issuer:   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
		Value:    "4.19552633596446",
	}
	require.NoError(t, testutil.Serialize(t, value, payload))
}

func TestBookOffersResponseJSONDecode(t *testing.T) {
	want, payload := bookOffersResponseFixture()
	var got BookOffersResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestBookOffersResponseClientDecode(t *testing.T) {
	want, payload := bookOffersResponseFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got BookOffersResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}
