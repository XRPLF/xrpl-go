package transaction

import (
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/currency"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAMMDeposit_TxType(t *testing.T) {
	tx := &AMMDeposit{}
	assert.Equal(t, AMMDepositTx, tx.TxType())
}

func TestAMMDeposit_Flatten(t *testing.T) {
	tx := &AMMDeposit{
		BaseTx: BaseTx{
			Account:  "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
			Fee:      types.XRPCurrencyAmount(10),
			Flags:    1048576,
			Sequence: 7,
		},
		Asset: ledger.Asset{
			Currency: "TST",
			Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
		},
		Asset2: ledger.Asset{
			Currency: "XRP",
		},
		Amount: types.IssuedCurrencyAmount{
			Value:    "2.5",
			Currency: "TST",
			Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
		},
		Amount2: types.XRPCurrencyAmount(30000000),
		EPrice: types.IssuedCurrencyAmount{
			Value:    "1.5",
			Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
			Currency: "TST",
		},
		LPTokenOut: types.IssuedCurrencyAmount{
			Value:    "100",
			Currency: "TST",
			Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
		},
		TradingFee: 10,
	}

	flattened := tx.Flatten()

	expected := `{
	"Account":         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"Fee":             "10",
	"Flags":           1048576,
	"Sequence":        7,
	"TransactionType": "AMMDeposit",
	"Asset": {
		"currency": "TST",
		"issuer":   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"
	},
	"Asset2": {
		"currency": "XRP"
	},
	"Amount": {
		"issuer":   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
		"value":    "2.5",
		"currency": "TST"
	},
	"Amount2": "30000000",
	"EPrice": {
		"value": "1.5",
		"issuer": "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
		"currency": "TST"
	},
	"LPTokenOut": {
		"value": "100",
		"currency": "TST",
		"issuer": "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"
	},
	"TradingFee": 10
}`

	err := testutil.CompareFlattenAndExpected(flattened, []byte(expected))
	if err != nil {
		t.Error(err)
	}
}

func TestAMMDeposit_Flags(t *testing.T) {
	tests := []struct {
		name     string
		setter   func(*AMMDeposit)
		expected uint32
	}{
		{
			name: "pass - SetLPTokentFlag",
			setter: func(a *AMMDeposit) {
				a.SetLPTokentFlag()
			},
			expected: TfLPToken,
		},
		{
			name: "pass - SetSingleAssetFlag",
			setter: func(a *AMMDeposit) {
				a.SetSingleAssetFlag()
			},
			expected: TfSingleAsset,
		},
		{
			name: "pass - SetTwoAssetFlag",
			setter: func(a *AMMDeposit) {
				a.SetTwoAssetFlag()
			},
			expected: TfTwoAsset,
		},
		{
			name: "pass - SetOneAssetLPTokenFlag",
			setter: func(a *AMMDeposit) {
				a.SetOneAssetLPTokenFlag()
			},
			expected: TfOneAssetLPToken,
		},
		{
			name: "pass - SetLimitLPTokenFlag",
			setter: func(a *AMMDeposit) {
				a.SetLimitLPTokenFlag()
			},
			expected: TfLimitLPToken,
		},
		{
			name: "pass - SetTwoAssetIfEmptyFlag",
			setter: func(a *AMMDeposit) {
				a.SetTwoAssetIfEmptyFlag()
			},
			expected: TfTwoAssetIfEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AMMDeposit{}
			tt.setter(a)
			if a.Flags != tt.expected {
				t.Errorf("Expected AMMDeposit Flags to be %d, got %d", tt.expected, a.Flags)
			}
		})
	}
}

func TestAMMDeposit_Validate(t *testing.T) {
	const issuer = types.Address("rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd")
	amount := types.IssuedCurrencyAmount{Value: "2.5", Currency: "TST", Issuer: issuer}
	amount2 := types.XRPCurrencyAmount(30000000)
	ePrice := types.IssuedCurrencyAmount{Value: "1.5", Currency: "TST", Issuer: issuer}
	lpTokenOut := types.IssuedCurrencyAmount{Value: "100", Currency: "039C99CD9AB0B70B32ECDA51EAAE471625608EA2", Issuer: "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S"}
	taglessIssuer, err := addresscodec.ClassicAddressToXAddress(issuer.String(), 0, false, false)
	require.NoError(t, err)
	canonicalAmount := types.IssuedCurrencyAmount{
		Value: "1", Currency: "0000000000000000000000005453540000000000", Issuer: types.Address(taglessIssuer),
	}
	zeroToken := lpTokenOut
	zeroToken.Value = "0"
	negativeToken := lpTokenOut
	negativeToken.Value = "-1"
	mptAmount := types.MPTCurrencyAmount{MPTIssuanceID: "00000001abcdef0123456789abcdef0123456789abcdef01", Value: "1"}
	mptAsset := ledger.Asset{MPTIssuanceID: "00000001ABCDEF0123456789ABCDEF0123456789ABCDEF01"}

	// deposit builds a TST/XRP deposit with the given mode flags and optional amount fields.
	deposit := func(flags uint32, amount, amount2, ePrice, lpTokenOut types.CurrencyAmount) *AMMDeposit {
		return &AMMDeposit{
			BaseTx: BaseTx{
				Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				TransactionType: AMMDepositTx,
				Fee:             types.XRPCurrencyAmount(10),
				Flags:           flags,
				Sequence:        7,
			},
			Asset:      ledger.Asset{Currency: "TST", Issuer: issuer},
			Asset2:     ledger.Asset{Currency: "XRP"},
			Amount:     amount,
			Amount2:    amount2,
			EPrice:     ePrice,
			LPTokenOut: lpTokenOut,
		}
	}
	withTradingFee := func(tx *AMMDeposit, fee uint16) *AMMDeposit {
		tx.TradingFee = fee
		return tx
	}
	withAssets := func(tx *AMMDeposit, asset, asset2 ledger.Asset) *AMMDeposit {
		tx.Asset, tx.Asset2 = asset, asset2
		return tx
	}
	invalidField := func(field string) error {
		return ErrTransactionInvalidField{Type: AMMDepositTx.String(), Field: field}
	}

	tests := []struct {
		name        string
		tx          *AMMDeposit
		expectedErr error
	}{
		{
			name: "pass - EPrice issue differs from Amount",
			tx:   deposit(TfLimitLPToken, amount, nil, amount2, nil),
		},
		{
			name: "pass - zero Amount with EPrice",
			tx:   deposit(TfLimitLPToken, types.XRPCurrencyAmount(0), nil, ePrice, nil),
		},
		{
			name: "pass - EPrice issue outside pool",
			tx:   deposit(TfLimitLPToken, amount, nil, lpTokenOut, nil),
		},
		{
			name: "pass - XRP pointer in Amount",
			tx:   deposit(TfTwoAsset, &amount2, amount, nil, nil),
		},
		{
			name: "pass - XRP pointer in Amount2",
			tx:   deposit(TfTwoAsset, amount, &amount2, nil, nil),
		},
		{
			name: "pass - token pointer in Amount",
			tx:   deposit(TfSingleAsset, &amount, nil, nil, nil),
		},
		{
			name: "pass - MPT pointer in Amount",
			tx:   withAssets(deposit(TfSingleAsset, &mptAmount, nil, nil, nil), mptAsset, ledger.Asset{Currency: "XRP"}),
		},
		{
			name:        "fail - XRP pointers still cannot name the same issue",
			tx:          deposit(TfTwoAsset, &amount2, &amount2, nil, nil),
			expectedErr: ErrAMMInvalidTokens,
		},
		{
			name:        "fail - XRP above maximum",
			tx:          deposit(TfSingleAsset, types.XRPCurrencyAmount(currency.MaxNativeDrops+1), nil, nil, nil),
			expectedErr: currency.ErrInvalidNativeAmount,
		},
		{
			name: "pass - MPT EPrice with nonzero issuer",
			tx:   deposit(TfLimitLPToken, amount, nil, mptAmount, nil),
		},
		{
			name: "fail - MPT EPrice with zero issuer",
			tx: deposit(TfLimitLPToken, amount, nil,
				types.MPTCurrencyAmount{MPTIssuanceID: "000000010000000000000000000000000000000000000000", Value: "1"}, nil),
			expectedErr: invalidField("EPrice"),
		},
		{
			name:        "fail - MPT LPTokenOut",
			tx:          deposit(TfLPToken, nil, nil, nil, mptAmount),
			expectedErr: ErrInvalidTokenFields,
		},
		{
			name:        "fail - XRP LPTokenOut",
			tx:          deposit(TfLPToken, nil, nil, nil, amount2),
			expectedErr: ErrInvalidTokenType,
		},
		{
			name:        "fail - undefined flag with one deposit mode",
			tx:          deposit(TfTwoAsset|0x2, amount, amount2, nil, nil),
			expectedErr: ErrInvalidFlags,
		},
		{
			name: "pass - all universal flags with one deposit mode",
			tx:   deposit(TfTwoAsset|types.TfUniversal, amount, amount2, nil, nil),
		},
		{
			name: "fail - same canonical pool issue",
			tx: withAssets(deposit(TfSingleAsset, amount, nil, nil, nil),
				ledger.Asset{Currency: "TST", Issuer: issuer},
				ledger.Asset{Currency: canonicalAmount.Currency, Issuer: canonicalAmount.Issuer}),
			expectedErr: ErrAMMInvalidTokens,
		},
		{
			name: "fail - same MPT pool issue with different hex case",
			tx: withAssets(deposit(TfLPToken, nil, nil, nil, lpTokenOut),
				ledger.Asset{MPTIssuanceID: "00000001ABCDEF0123456789ABCDEF0123456789ABCDEF01"},
				ledger.Asset{MPTIssuanceID: "00000001abcdef0123456789abcdef0123456789abcdef01"}),
			expectedErr: ErrAMMInvalidTokens,
		},
		{
			name: "fail - both pool assets are XRP",
			tx: withAssets(deposit(TfLPToken, nil, nil, nil, lpTokenOut),
				ledger.Asset{Currency: "XRP"}, ledger.Asset{Currency: "XRP"}),
			expectedErr: ErrAMMInvalidTokens,
		},
		{
			name:        "fail - same canonical amount issues",
			tx:          deposit(TfTwoAsset, amount, canonicalAmount, nil, nil),
			expectedErr: ErrAMMInvalidTokens,
		},
		{
			name:        "fail - Amount outside pool",
			tx:          deposit(TfSingleAsset, lpTokenOut, nil, nil, nil),
			expectedErr: ErrAMMInvalidTokens,
		},
		{
			name:        "fail - Amount2 outside pool",
			tx:          deposit(TfTwoAsset, amount, lpTokenOut, nil, nil),
			expectedErr: ErrAMMInvalidTokens,
		},
		{
			name:        "fail - zero LPTokenOut",
			tx:          deposit(TfLPToken, nil, nil, nil, zeroToken),
			expectedErr: ErrInvalidTokenValue,
		},
		{
			name:        "fail - negative LPTokenOut",
			tx:          deposit(TfLPToken, nil, nil, nil, negativeToken),
			expectedErr: ErrInvalidTokenValue,
		},
		{
			name:        "fail - zero EPrice",
			tx:          deposit(TfLimitLPToken, amount, nil, types.XRPCurrencyAmount(0), nil),
			expectedErr: ErrInvalidTokenValue,
		},
		{
			name:        "fail - negative EPrice",
			tx:          deposit(TfLimitLPToken, amount, nil, negativeToken, nil),
			expectedErr: ErrInvalidTokenValue,
		},
		{
			name:        "fail - zero Amount without EPrice",
			tx:          deposit(TfSingleAsset, types.XRPCurrencyAmount(0), nil, nil, nil),
			expectedErr: ErrInvalidTokenValue,
		},
		{
			name:        "fail - zero Amount2",
			tx:          deposit(TfTwoAsset, amount, types.XRPCurrencyAmount(0), nil, nil),
			expectedErr: ErrInvalidTokenValue,
		},
		{
			name: "pass - canonical amount matches pool",
			tx:   deposit(TfTwoAsset, canonicalAmount, amount2, nil, nil),
		},
		{
			name:        "fail - same XRP amount issues",
			tx:          deposit(TfTwoAsset, amount2, amount2, nil, nil),
			expectedErr: ErrAMMInvalidTokens,
		},
		{
			name: "fail - same MPT amount issues",
			tx: withAssets(deposit(TfTwoAsset, mptAmount,
				types.MPTCurrencyAmount{MPTIssuanceID: mptAsset.MPTIssuanceID, Value: "2"}, nil, nil),
				mptAsset, ledger.Asset{Currency: "XRP"}),
			expectedErr: ErrAMMInvalidTokens,
		},
		{
			name: "pass - MPT amount matches pool with different hex case",
			tx: withAssets(deposit(TfTwoAsset, mptAmount, amount2, nil, nil),
				mptAsset, ledger.Asset{Currency: "XRP"}),
		},
		{
			name: "pass - noCurrency pool asset",
			tx: withAssets(deposit(TfSingleAsset,
				types.IssuedCurrencyAmount{Value: "1", Currency: "0000000000000000000000000000000000000001", Issuer: issuer}, nil, nil, nil),
				ledger.Asset{Currency: "0000000000000000000000000000000000000001", Issuer: issuer}, ledger.Asset{Currency: "XRP"}),
		},
		{
			name: "pass - deposit amounts in reverse pool order",
			tx:   deposit(TfTwoAsset, amount2, amount, nil, nil),
		},
		{
			name: "pass - TfTwoAsset with Amount and Amount2",
			tx:   deposit(TfTwoAsset, amount, amount2, nil, nil),
		},
		{
			name: "pass - TfTwoAsset with LPTokenOut minimum",
			tx:   deposit(TfTwoAsset, amount, amount2, nil, lpTokenOut),
		},
		{
			name: "pass - universal flag with one deposit mode",
			tx:   deposit(TfTwoAsset|types.TfFullyCanonicalSig, amount, amount2, nil, nil),
		},
		{
			name:        "fail - TradingFee exceeds maximum",
			tx:          withTradingFee(deposit(TfTwoAssetIfEmpty, amount, amount2, nil, nil), AmmMaxTradingFee+1),
			expectedErr: ErrAMMTradingFeeTooHigh{Value: AmmMaxTradingFee + 1, Limit: AmmMaxTradingFee},
		},
		{
			name:        "fail - TradingFee set outside two asset if empty mode",
			tx:          withTradingFee(deposit(TfTwoAsset, amount, amount2, nil, nil), 1),
			expectedErr: ErrAMMDepositTradingFeeRequiresEmptyPool,
		},
		{
			name: "fail - invalid AMMDeposit BaseTx without TransactionType",
			tx: func() *AMMDeposit {
				tx := deposit(TfTwoAsset, amount, amount2, nil, nil)
				tx.TransactionType = ""
				return tx
			}(),
			expectedErr: ErrInvalidTransactionType,
		},
		{
			name:        "fail - no deposit mode flag",
			tx:          deposit(0, amount, amount2, nil, nil),
			expectedErr: ErrInvalidFlags,
		},
		{
			name:        "fail - two deposit mode flags",
			tx:          deposit(TfTwoAsset|TfLPToken, amount, amount2, nil, lpTokenOut),
			expectedErr: ErrInvalidFlags,
		},
		{
			name: "fail - flags error precedes identical pool assets",
			tx: withAssets(deposit(TfTwoAsset|TfLPToken, amount, amount2, nil, nil),
				ledger.Asset{Currency: "XRP"}, ledger.Asset{Currency: "XRP"}),
			expectedErr: ErrInvalidFlags,
		},
		{
			name:        "fail - invalid AMMDeposit with Amount2 but no Amount",
			tx:          deposit(TfTwoAsset, nil, amount2, nil, nil),
			expectedErr: ErrAMMMustSetAmountWithAmount2,
		},
		{
			name:        "fail - invalid AMMDeposit with EPrice but no Amount",
			tx:          deposit(TfLimitLPToken, nil, nil, ePrice, nil),
			expectedErr: ErrAMMMustSetAmountWithEPrice,
		},
		{
			name:        "fail - invalid AMMDeposit with no LPTokenOut or Amount",
			tx:          deposit(TfLPToken, nil, nil, nil, nil),
			expectedErr: ErrAMMAtLeastOneAssetMustBeSet,
		},
		{
			name: "pass - TfLPToken with LPTokenOut",
			tx:   deposit(TfLPToken, nil, nil, nil, lpTokenOut),
		},
		{
			name: "pass - TfLPToken with LPTokenOut and both deposit minimums",
			tx:   deposit(TfLPToken, amount, amount2, nil, lpTokenOut),
		},
		{
			name:        "fail - TfLPToken without LPTokenOut",
			tx:          deposit(TfLPToken, amount, amount2, nil, nil),
			expectedErr: ErrMissingField{Field: "LPTokenOut"},
		},
		{
			name:        "fail - TfLPToken with Amount but no Amount2",
			tx:          deposit(TfLPToken, amount, nil, nil, lpTokenOut),
			expectedErr: ErrMissingField{Field: "Amount2"},
		},
		{
			name:        "fail - TfLPToken with EPrice",
			tx:          deposit(TfLPToken, amount, amount2, ePrice, lpTokenOut),
			expectedErr: invalidField("EPrice"),
		},
		{
			name: "pass - TfSingleAsset with Amount",
			tx:   deposit(TfSingleAsset, amount, nil, nil, nil),
		},
		{
			name: "pass - TfSingleAsset with LPTokenOut minimum",
			tx:   deposit(TfSingleAsset, amount, nil, nil, lpTokenOut),
		},
		{
			name:        "fail - TfSingleAsset without Amount",
			tx:          deposit(TfSingleAsset, nil, nil, nil, lpTokenOut),
			expectedErr: ErrMissingField{Field: "Amount"},
		},
		{
			name:        "fail - TfSingleAsset with Amount2",
			tx:          deposit(TfSingleAsset, amount, amount2, nil, nil),
			expectedErr: invalidField("Amount2"),
		},
		{
			name:        "fail - TfSingleAsset with EPrice",
			tx:          deposit(TfSingleAsset, amount, nil, ePrice, nil),
			expectedErr: invalidField("EPrice"),
		},
		{
			name:        "fail - TfTwoAsset without Amount",
			tx:          deposit(TfTwoAsset, nil, nil, nil, lpTokenOut),
			expectedErr: ErrMissingField{Field: "Amount"},
		},
		{
			name:        "fail - TfTwoAsset without Amount2",
			tx:          deposit(TfTwoAsset, amount, nil, nil, nil),
			expectedErr: ErrMissingField{Field: "Amount2"},
		},
		{
			name:        "fail - TfTwoAsset with EPrice",
			tx:          deposit(TfTwoAsset, amount, amount2, ePrice, nil),
			expectedErr: invalidField("EPrice"),
		},
		{
			name: "pass - TfOneAssetLPToken with Amount and LPTokenOut",
			tx:   deposit(TfOneAssetLPToken, amount, nil, nil, lpTokenOut),
		},
		{
			name:        "fail - TfOneAssetLPToken without Amount",
			tx:          deposit(TfOneAssetLPToken, nil, nil, nil, lpTokenOut),
			expectedErr: ErrMissingField{Field: "Amount"},
		},
		{
			name:        "fail - TfOneAssetLPToken without LPTokenOut",
			tx:          deposit(TfOneAssetLPToken, amount, nil, nil, nil),
			expectedErr: ErrMissingField{Field: "LPTokenOut"},
		},
		{
			name:        "fail - TfOneAssetLPToken with Amount2",
			tx:          deposit(TfOneAssetLPToken, amount, amount2, nil, lpTokenOut),
			expectedErr: invalidField("Amount2"),
		},
		{
			name:        "fail - TfOneAssetLPToken with EPrice",
			tx:          deposit(TfOneAssetLPToken, amount, nil, ePrice, lpTokenOut),
			expectedErr: invalidField("EPrice"),
		},
		{
			name: "pass - TfLimitLPToken with Amount and EPrice",
			tx:   deposit(TfLimitLPToken, amount, nil, ePrice, nil),
		},
		{
			name:        "fail - TfLimitLPToken without Amount",
			tx:          deposit(TfLimitLPToken, nil, nil, nil, lpTokenOut),
			expectedErr: ErrMissingField{Field: "Amount"},
		},
		{
			name:        "fail - TfLimitLPToken without EPrice",
			tx:          deposit(TfLimitLPToken, amount, nil, nil, nil),
			expectedErr: ErrMissingField{Field: "EPrice"},
		},
		{
			name:        "fail - TfLimitLPToken with Amount2",
			tx:          deposit(TfLimitLPToken, amount, amount2, ePrice, nil),
			expectedErr: invalidField("Amount2"),
		},
		{
			name:        "fail - TfLimitLPToken with LPTokenOut",
			tx:          deposit(TfLimitLPToken, amount, nil, ePrice, lpTokenOut),
			expectedErr: invalidField("LPTokenOut"),
		},
		{
			name: "pass - TfTwoAssetIfEmpty with Amount, Amount2 and TradingFee",
			tx:   withTradingFee(deposit(TfTwoAssetIfEmpty, amount, amount2, nil, nil), AmmMaxTradingFee),
		},
		{
			name:        "fail - TfTwoAssetIfEmpty without Amount",
			tx:          deposit(TfTwoAssetIfEmpty, nil, nil, nil, lpTokenOut),
			expectedErr: ErrMissingField{Field: "Amount"},
		},
		{
			name:        "fail - TfTwoAssetIfEmpty without Amount2",
			tx:          deposit(TfTwoAssetIfEmpty, amount, nil, nil, nil),
			expectedErr: ErrMissingField{Field: "Amount2"},
		},
		{
			name:        "fail - TfTwoAssetIfEmpty with EPrice",
			tx:          deposit(TfTwoAssetIfEmpty, amount, amount2, ePrice, nil),
			expectedErr: invalidField("EPrice"),
		},
		{
			name:        "fail - TfTwoAssetIfEmpty with LPTokenOut",
			tx:          deposit(TfTwoAssetIfEmpty, amount, amount2, nil, lpTokenOut),
			expectedErr: invalidField("LPTokenOut"),
		},
		{
			name: "fail - invalid AMMDeposit, invalid Asset",
			tx: func() *AMMDeposit {
				tx := deposit(TfLPToken, nil, nil, nil, lpTokenOut)
				tx.Asset.Currency = ""
				return tx
			}(),
			expectedErr: ErrMissingAssetCurrency,
		},
		{
			name: "fail - invalid AMMDeposit, invalid Asset2",
			tx: func() *AMMDeposit {
				tx := deposit(TfLPToken, nil, nil, nil, lpTokenOut)
				tx.Asset2.Issuer = issuer
				return tx
			}(),
			expectedErr: ErrInvalidAssetIssuer,
		},
		{
			name:        "fail - invalid AMMDeposit, invalid Amount",
			tx:          deposit(TfSingleAsset, types.IssuedCurrencyAmount{Value: "1"}, nil, nil, nil),
			expectedErr: ErrInvalidTokenFields,
		},
		{
			name:        "fail - invalid AMMDeposit, invalid Amount2",
			tx:          deposit(TfTwoAsset, amount, types.IssuedCurrencyAmount{Value: "1"}, nil, nil),
			expectedErr: ErrInvalidTokenFields,
		},
		{
			name:        "fail - invalid AMMDeposit, invalid EPrice",
			tx:          deposit(TfLimitLPToken, amount, nil, types.IssuedCurrencyAmount{Value: "1"}, nil),
			expectedErr: ErrInvalidTokenFields,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.tx.Validate()
			if tt.expectedErr == nil {
				require.NoError(t, err)
				require.True(t, valid)
				return
			}
			require.ErrorIs(t, err, tt.expectedErr)
			require.False(t, valid)
		})
	}
}
