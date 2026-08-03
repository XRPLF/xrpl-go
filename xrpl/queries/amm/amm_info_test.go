package amm

import (
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestAMMInfoRequest(t *testing.T) {
	tests := []struct {
		name     string
		request  InfoRequest
		expected string
	}{
		{
			name: "asset pair",
			request: InfoRequest{
				Asset:  ledger.Asset{Currency: "USD", Issuer: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"},
				Asset2: ledger.Asset{Currency: "XRP"},
			},
			expected: `{
	"asset": {
		"currency": "USD",
		"issuer": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	},
	"asset2": {
		"currency": "XRP"
	}
}`,
		},
		{
			name: "asset pair with liquidity provider",
			request: InfoRequest{
				Asset:   ledger.Asset{Currency: "XRP"},
				Asset2:  ledger.Asset{Currency: "USD", Issuer: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"},
				Account: "rJVUeRqDFNs2xqA7ncVE6ZoAhPUoaJJSQm",
			},
			expected: `{
	"asset": {
		"currency": "XRP"
	},
	"asset2": {
		"currency": "USD",
		"issuer": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	},
	"account": "rJVUeRqDFNs2xqA7ncVE6ZoAhPUoaJJSQm"
}`,
		},
		{
			name: "amm_account",
			request: InfoRequest{
				AMMAccount: "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
			},
			expected: `{
	"amm_account": "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S"
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

func TestAMMInfoResponse(t *testing.T) {
	assetFrozen := true
	asset2Frozen := false

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
			AssetFrozen:  &assetFrozen,
			Asset2Frozen: &asset2Frozen,
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
				Expiration:   "2024-01-01T00:00:00Z",
				TimeInterval: 0,
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
		"asset_frozen": true,
		"asset2_frozen": false,
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
			"expiration": "2024-01-01T00:00:00Z",
			"time_interval": 0
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
	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestAuctionSlotInfo_TimeIntervalExpiredSentinel(t *testing.T) {
	s := AuctionSlotInfo{
		Account:       "rJVUeRqDFNs2xqA7ncVE6ZoAhPUoaJJSQm",
		DiscountedFee: 0,
		Price: types.IssuedCurrencyAmount{
			Currency: "039C99CD9AB0B70B32ECDA51EAAE471625608EA2",
			Issuer:   "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
			Value:    "100",
		},
		TimeInterval: 20,
	}

	j := `{
	"account": "rJVUeRqDFNs2xqA7ncVE6ZoAhPUoaJJSQm",
	"discounted_fee": 0,
	"price": {
		"issuer": "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
		"currency": "039C99CD9AB0B70B32ECDA51EAAE471625608EA2",
		"value": "100"
	},
	"time_interval": 20
}`
	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestAMMInfoResponse_XRPAssets(t *testing.T) {
	asset2Frozen := false

	s := InfoResponse{
		AMM: Info{
			Account: "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
			Amount:  types.XRPCurrencyAmount(1000000),
			Amount2: types.IssuedCurrencyAmount{
				Currency: "USD",
				Issuer:   "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				Value:    "500",
			},
			Asset2Frozen: &asset2Frozen,
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
		"asset2_frozen": false,
		"lp_token": {
			"issuer": "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
			"currency": "039C99CD9AB0B70B32ECDA51EAAE471625608EA2",
			"value": "22360679.77"
		},
		"trading_fee": 600
	}
}`
	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestAMMInfoResponse_FrozenFieldOptionality(t *testing.T) {
	unfrozen := false
	issuedAmount := types.IssuedCurrencyAmount{
		Currency: "USD",
		Issuer:   "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		Value:    "500",
	}
	lpToken := types.IssuedCurrencyAmount{
		Currency: "039C99CD9AB0B70B32ECDA51EAAE471625608EA2",
		Issuer:   "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
		Value:    "22360679.77",
	}

	tests := []struct {
		name     string
		info     Info
		expected string
	}{
		{
			name: "asset is XRP",
			info: Info{
				Account:      "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
				Amount:       types.XRPCurrencyAmount(1000000),
				Amount2:      issuedAmount,
				Asset2Frozen: &unfrozen,
				LPToken:      lpToken,
			},
			expected: `{
	"account": "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
	"amount": "1000000",
	"amount2": {
		"issuer": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		"currency": "USD",
		"value": "500"
	},
	"asset2_frozen": false,
	"lp_token": {
		"issuer": "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
		"currency": "039C99CD9AB0B70B32ECDA51EAAE471625608EA2",
		"value": "22360679.77"
	},
	"trading_fee": 0
}`,
		},
		{
			name: "asset2 is XRP",
			info: Info{
				Account:     "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
				Amount:      issuedAmount,
				Amount2:     types.XRPCurrencyAmount(1000000),
				AssetFrozen: &unfrozen,
				LPToken:     lpToken,
			},
			expected: `{
	"account": "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
	"amount": {
		"issuer": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		"currency": "USD",
		"value": "500"
	},
	"amount2": "1000000",
	"asset_frozen": false,
	"lp_token": {
		"issuer": "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
		"currency": "039C99CD9AB0B70B32ECDA51EAAE471625608EA2",
		"value": "22360679.77"
	},
	"trading_fee": 0
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := testutil.SerializeAndDeserialize(t, tt.info, tt.expected); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestAMMInfoRequest_Validate(t *testing.T) {
	asset := ledger.Asset{Currency: "XRP"}
	asset2 := ledger.Asset{Currency: "USD", Issuer: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"}
	const ammAccount = "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S"
	const liquidityProvider = "rJVUeRqDFNs2xqA7ncVE6ZoAhPUoaJJSQm"

	tests := []struct {
		name     string
		request  InfoRequest
		expected error
	}{
		{name: "no selector", request: InfoRequest{}, expected: ErrInvalidInfoRequest},
		{name: "asset only", request: InfoRequest{Asset: asset}, expected: ErrInvalidInfoRequest},
		{name: "asset2 only", request: InfoRequest{Asset2: asset2}, expected: ErrInvalidInfoRequest},
		{name: "asset pair", request: InfoRequest{Asset: asset, Asset2: asset2}},
		{name: "amm_account only", request: InfoRequest{AMMAccount: ammAccount}},
		{name: "asset and amm_account", request: InfoRequest{Asset: asset, AMMAccount: ammAccount}, expected: ErrInvalidInfoRequest},
		{name: "asset2 and amm_account", request: InfoRequest{Asset2: asset2, AMMAccount: ammAccount}, expected: ErrInvalidInfoRequest},
		{name: "asset pair and amm_account", request: InfoRequest{Asset: asset, Asset2: asset2, AMMAccount: ammAccount}, expected: ErrInvalidInfoRequest},
		{name: "liquidity provider only", request: InfoRequest{Account: liquidityProvider}, expected: ErrInvalidInfoRequest},
		{name: "amm_account with liquidity provider", request: InfoRequest{AMMAccount: ammAccount, Account: liquidityProvider}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.expected == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, tt.expected)
		})
	}
}
