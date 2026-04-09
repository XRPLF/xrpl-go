package amm

import (
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestAMMInfoRequest(t *testing.T) {
	s := InfoRequest{
		Asset: ledger.Asset{
			Currency: "USD",
			Issuer:   "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		},
		Asset2: ledger.Asset{
			Currency: "XRP",
		},
	}

	j := `{
	"asset": {
		"currency": "USD",
		"issuer": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	},
	"asset2": {
		"currency": "XRP"
	}
}`
	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestAMMInfoRequest_WithAMMAccount(t *testing.T) {
	s := InfoRequest{
		Asset: ledger.Asset{
			Currency: "USD",
			Issuer:   "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		},
		Asset2: ledger.Asset{
			Currency: "BTC",
			Issuer:   "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
		},
		AMMAccount: "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
	}

	j := `{
	"asset": {
		"currency": "USD",
		"issuer": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	},
	"asset2": {
		"currency": "BTC",
		"issuer": "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe"
	},
	"amm_account": "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S"
}`
	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestAMMInfoResponse(t *testing.T) {
	s := InfoResponse{
		AMM: Info{
			Account: "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
			Amount: types.IssuedCurrencyAmount{
				Currency: "USD",
				Issuer:   "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				Value:    "1000",
			},
			Amount2: types.IssuedCurrencyAmount{
				Currency: "BTC",
				Issuer:   "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
				Value:    "0.5",
			},
			AuctionSlot: &AuctionSlotInfo{
				Account: "rJVUeRqDFNs2xqA7ncVE6ZoAhPUoaJJSQm",
				AuthAccounts: []AuthAccountInfo{
					{Account: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"},
				},
				DiscountedFee: 0,
				Price: types.IssuedCurrencyAmount{
					Currency: "039C99CD9AB0B70B32ECDA51EAAE471625608EA2",
					Issuer:   "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
					Value:    "100",
				},
				Expiration: "2024-01-01T00:00:00Z",
			},
			LPToken: types.IssuedCurrencyAmount{
				Currency: "039C99CD9AB0B70B32ECDA51EAAE471625608EA2",
				Issuer:   "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
				Value:    "22360679.77",
			},
			TradingFee: 500,
			VoteSlots: []VoteSlotInfo{
				{
					Account:    "rJVUeRqDFNs2xqA7ncVE6ZoAhPUoaJJSQm",
					TradingFee: 500,
					VoteWeight: 100000,
				},
			},
		},
		LedgerHash:  "4C99E5F63C0D0B1C2283B4F5DCE2239F80CE92E8B1A6AED1E110C198FC96E659",
		LedgerIndex: 1234,
		Validated:   true,
	}

	j := `{
	"amm": {
		"account": "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
		"amount": {
			"issuer": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			"currency": "USD",
			"value": "1000"
		},
		"amount2": {
			"issuer": "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
			"currency": "BTC",
			"value": "0.5"
		},
		"auction_slot": {
			"account": "rJVUeRqDFNs2xqA7ncVE6ZoAhPUoaJJSQm",
			"auth_accounts": [
				{
					"account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
				}
			],
			"discounted_fee": 0,
			"price": {
				"issuer": "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
				"currency": "039C99CD9AB0B70B32ECDA51EAAE471625608EA2",
				"value": "100"
			},
			"expiration": "2024-01-01T00:00:00Z"
		},
		"lp_token": {
			"issuer": "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
			"currency": "039C99CD9AB0B70B32ECDA51EAAE471625608EA2",
			"value": "22360679.77"
		},
		"trading_fee": 500,
		"vote_slots": [
			{
				"account": "rJVUeRqDFNs2xqA7ncVE6ZoAhPUoaJJSQm",
				"trading_fee": 500,
				"vote_weight": 100000
			}
		]
	},
	"ledger_hash": "4C99E5F63C0D0B1C2283B4F5DCE2239F80CE92E8B1A6AED1E110C198FC96E659",
	"ledger_index": 1234,
	"validated": true
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestAMMInfoResponse_XRPAssets(t *testing.T) {
	s := InfoResponse{
		AMM: Info{
			Account: "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
			Amount:  types.XRPCurrencyAmount(1000000),
			Amount2: types.IssuedCurrencyAmount{
				Currency: "USD",
				Issuer:   "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				Value:    "500",
			},
			LPToken: types.IssuedCurrencyAmount{
				Currency: "039C99CD9AB0B70B32ECDA51EAAE471625608EA2",
				Issuer:   "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
				Value:    "22360679.77",
			},
			TradingFee: 600,
		},
		Validated: false,
	}

	j := `{
	"amm": {
		"account": "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
		"amount": "1000000",
		"amount2": {
			"issuer": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			"currency": "USD",
			"value": "500"
		},
		"lp_token": {
			"issuer": "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
			"currency": "039C99CD9AB0B70B32ECDA51EAAE471625608EA2",
			"value": "22360679.77"
		},
		"trading_fee": 600
	}
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestAMMInfoRequest_Validate(t *testing.T) {
	t.Run("pass - with asset and asset2", func(t *testing.T) {
		req := InfoRequest{
			Asset:  ledger.Asset{Currency: "XRP"},
			Asset2: ledger.Asset{Currency: "USD", Issuer: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"},
		}
		require.NoError(t, req.Validate())
	})

	t.Run("pass - with amm_account", func(t *testing.T) {
		req := InfoRequest{
			AMMAccount: "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
		}
		require.NoError(t, req.Validate())
	})

	t.Run("pass - with amm_account and assets", func(t *testing.T) {
		req := InfoRequest{
			Asset:      ledger.Asset{Currency: "XRP"},
			Asset2:     ledger.Asset{Currency: "USD", Issuer: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"},
			AMMAccount: "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
		}
		require.NoError(t, req.Validate())
	})

	t.Run("fail - no amm_account and no assets", func(t *testing.T) {
		req := InfoRequest{}
		require.ErrorIs(t, req.Validate(), ErrInvalidInfoRequest)
	})

	t.Run("fail - only asset specified", func(t *testing.T) {
		req := InfoRequest{
			Asset: ledger.Asset{Currency: "XRP"},
		}
		require.ErrorIs(t, req.Validate(), ErrInvalidInfoRequest)
	})

	t.Run("fail - only asset2 specified", func(t *testing.T) {
		req := InfoRequest{
			Asset2: ledger.Asset{Currency: "USD", Issuer: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"},
		}
		require.ErrorIs(t, req.Validate(), ErrInvalidInfoRequest)
	})
}
