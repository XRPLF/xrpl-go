package transaction

import (
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAMMWithdraw_TxType(t *testing.T) {
	tx := &AMMWithdraw{}
	assert.Equal(t, AMMWithdrawTx, tx.TxType())
}

func TestAMMWithdraw_Flags(t *testing.T) {
	tests := []struct {
		name     string
		setter   func(*AMMWithdraw)
		expected uint32
	}{
		{
			name: "pass - SetLPTokentFlag",
			setter: func(a *AMMWithdraw) {
				a.SetLPTokentFlag()
			},
			expected: TfLPToken,
		},
		{
			name: "pass - SetWithdrawAllFlag",
			setter: func(a *AMMWithdraw) {
				a.SetWithdrawAllFlag()
			},
			expected: TfWithdrawAll,
		},
		{
			name: "pass - SetOneAssetWithdrawAllFlag",
			setter: func(a *AMMWithdraw) {
				a.SetOneAssetWithdrawAllFlag()
			},
			expected: TfOneAssetWithdrawAll,
		},
		{
			name: "pass - SetSingleAssetFlag",
			setter: func(a *AMMWithdraw) {
				a.SetSingleAssetFlag()
			},
			expected: TfSingleAsset,
		},
		{
			name: "pass - SetTwoAssetFlag",
			setter: func(a *AMMWithdraw) {
				a.SetTwoAssetFlag()
			},
			expected: TfTwoAsset,
		},
		{
			name: "pass - SetOneAssetLPTokenFlag",
			setter: func(a *AMMWithdraw) {
				a.SetOneAssetLPTokenFlag()
			},
			expected: TfOneAssetLPToken,
		},
		{
			name: "pass - SetLimitLPTokenFlag",
			setter: func(a *AMMWithdraw) {
				a.SetLimitLPTokenFlag()
			},
			expected: TfLimitLPToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AMMWithdraw{}
			tt.setter(a)
			if a.Flags != tt.expected {
				t.Errorf("Expected AMMWithdraw Flags to be %d, got %d", tt.expected, a.Flags)
			}
		})
	}
}

func TestAMMWithdraw_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		input    *AMMWithdraw
		expected string
	}{
		{
			name: "pass - Full AMMWithdraw",
			input: &AMMWithdraw{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Asset: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
				Amount: types.IssuedCurrencyAmount{
					Value:    "5",
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Amount2: types.IssuedCurrencyAmount{
					Value:    "50000000",
					Currency: "ABC",
					Issuer:   "rKswUGcm3wXPSaMfEHdUAQvE8otkZCd1ur",
				},
				EPrice: types.IssuedCurrencyAmount{
					Value:    "1",
					Currency: "TST",
					Issuer:   "rJhPKEN1m6FDGy9FZ85Ek2n3tAyUBR4KBv",
				},
				LPTokenIn: types.IssuedCurrencyAmount{
					Value:    "100",
					Currency: "TST",
					Issuer:   "rQH2Rhja1YRC3spuVukZBu9WzRiD1R9Dcr",
				},
			},
			expected: `{
				"TransactionType": "AMMWithdraw",
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee": "10",
				"Flags": 1048576,
				"Sequence": 10,
				"Asset": {
					"currency": "TST",
					"issuer":   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"
				},
				"Asset2": {
					"currency": "XRP"
				},
				"Amount": {
					"value":    "5",
					"currency": "TST",
					"issuer":   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"
				},
				"Amount2": {
					"value": "50000000",
					"currency": "ABC",
					"issuer": "rKswUGcm3wXPSaMfEHdUAQvE8otkZCd1ur"
				},
				"EPrice": {
					"value": "1",
					"currency": "TST",
					"issuer": "rJhPKEN1m6FDGy9FZ85Ek2n3tAyUBR4KBv"
				},
				"LPTokenIn": {
					"value": "100",
					"currency": "TST",
					"issuer": "rQH2Rhja1YRC3spuVukZBu9WzRiD1R9Dcr"
				}
			}`,
		},
		{
			name: "pass - Minimal AMMWithdraw",
			input: &AMMWithdraw{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Asset: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
			},
			expected: `{
				"TransactionType": "AMMWithdraw",
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee": "10",
				"Flags": 1048576,
				"Sequence": 10,
				"Asset": {
					"currency": "TST",
					"issuer": "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"
				},
				"Asset2": {
					"currency": "XRP"
				}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Flatten()

			err := testutil.CompareFlattenAndExpected(result, []byte(tt.expected))
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestAMMWithdraw_Validate(t *testing.T) {
	const issuer = types.Address("rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd")
	amount := types.IssuedCurrencyAmount{Value: "5", Currency: "TST", Issuer: issuer}
	amount2 := types.XRPCurrencyAmount(50000000)
	ePrice := types.IssuedCurrencyAmount{Value: "1.5", Currency: "TST", Issuer: issuer}
	lpTokenIn := types.IssuedCurrencyAmount{Value: "100", Currency: "039C99CD9AB0B70B32ECDA51EAAE471625608EA2", Issuer: "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S"}
	noLPTokenIn := types.IssuedCurrencyAmount{}
	taglessIssuer, err := addresscodec.ClassicAddressToXAddress(issuer.String(), 0, false, false)
	require.NoError(t, err)
	canonicalAmount := types.IssuedCurrencyAmount{
		Value: "1", Currency: "0000000000000000000000005453540000000000", Issuer: types.Address(taglessIssuer),
	}
	zeroLPTokenIn := lpTokenIn
	zeroLPTokenIn.Value = "0"
	negativeLPTokenIn := lpTokenIn
	negativeLPTokenIn.Value = "-1"
	mptAmount := types.MPTCurrencyAmount{MPTIssuanceID: "00000001abcdef0123456789abcdef0123456789abcdef01", Value: "1"}
	mptAsset := ledger.Asset{MPTIssuanceID: "00000001ABCDEF0123456789ABCDEF0123456789ABCDEF01"}

	// withdraw builds a TST/XRP withdrawal with the given mode flags and optional amount fields.
	withdraw := func(flags uint32, amount, amount2, ePrice types.CurrencyAmount, lpTokenIn types.IssuedCurrencyAmount) *AMMWithdraw {
		return &AMMWithdraw{
			BaseTx: BaseTx{
				Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				TransactionType: AMMWithdrawTx,
				Fee:             types.XRPCurrencyAmount(10),
				Flags:           flags,
				Sequence:        10,
			},
			Asset:     ledger.Asset{Currency: "TST", Issuer: issuer},
			Asset2:    ledger.Asset{Currency: "XRP"},
			Amount:    amount,
			Amount2:   amount2,
			EPrice:    ePrice,
			LPTokenIn: lpTokenIn,
		}
	}
	withAssets := func(tx *AMMWithdraw, asset, asset2 ledger.Asset) *AMMWithdraw {
		tx.Asset, tx.Asset2 = asset, asset2
		return tx
	}
	invalidField := func(field string) error {
		return ErrTransactionInvalidField{Type: AMMWithdrawTx.String(), Field: field}
	}

	tests := []struct {
		name        string
		tx          *AMMWithdraw
		expectedErr error
	}{
		{
			name: "pass - TfLPToken with LPTokenIn",
			tx:   withdraw(TfLPToken, nil, nil, nil, lpTokenIn),
		},
		{
			name: "pass - TfWithdrawAll without amounts",
			tx:   withdraw(TfWithdrawAll, nil, nil, nil, noLPTokenIn),
		},
		{
			name: "pass - TfOneAssetWithdrawAll with Amount",
			tx:   withdraw(TfOneAssetWithdrawAll, amount, nil, nil, noLPTokenIn),
		},
		{
			name: "pass - TfOneAssetWithdrawAll with zero Amount",
			tx:   withdraw(TfOneAssetWithdrawAll, types.XRPCurrencyAmount(0), nil, nil, noLPTokenIn),
		},
		{
			name: "pass - TfSingleAsset with Amount",
			tx:   withdraw(TfSingleAsset, amount, nil, nil, noLPTokenIn),
		},
		{
			name: "pass - TfTwoAsset with Amount and Amount2",
			tx:   withdraw(TfTwoAsset, amount, amount2, nil, noLPTokenIn),
		},
		{
			name: "pass - TfTwoAsset amounts in reverse pool order",
			tx:   withdraw(TfTwoAsset, amount2, amount, nil, noLPTokenIn),
		},
		{
			name: "pass - TfOneAssetLPToken with Amount and LPTokenIn",
			tx:   withdraw(TfOneAssetLPToken, amount, nil, nil, lpTokenIn),
		},
		{
			name: "pass - TfOneAssetLPToken with zero Amount",
			tx:   withdraw(TfOneAssetLPToken, types.XRPCurrencyAmount(0), nil, nil, lpTokenIn),
		},
		{
			name:        "fail - TfLimitLPToken with pool asset EPrice",
			tx:          withdraw(TfLimitLPToken, amount, nil, ePrice, noLPTokenIn),
			expectedErr: ErrAMMInvalidTokens,
		},
		{
			name:        "fail - TfLimitLPToken with zero Amount and pool asset EPrice",
			tx:          withdraw(TfLimitLPToken, types.XRPCurrencyAmount(0), nil, ePrice, noLPTokenIn),
			expectedErr: ErrAMMInvalidTokens,
		},
		{
			name: "pass - EPrice issue outside pool",
			tx:   withdraw(TfLimitLPToken, amount, nil, lpTokenIn, noLPTokenIn),
		},
		{
			name:        "fail - XRP EPrice",
			tx:          withdraw(TfLimitLPToken, amount, nil, amount2, noLPTokenIn),
			expectedErr: ErrAMMInvalidTokens,
		},
		{
			name:        "fail - canonical pool asset EPrice",
			tx:          withdraw(TfLimitLPToken, amount, nil, canonicalAmount, noLPTokenIn),
			expectedErr: ErrAMMInvalidTokens,
		},
		{
			name: "pass - zero Amount with LP token EPrice",
			tx:   withdraw(TfLimitLPToken, types.XRPCurrencyAmount(0), nil, lpTokenIn, noLPTokenIn),
		},
		{
			name: "pass - all universal flags with one withdrawal mode",
			tx:   withdraw(TfTwoAsset|types.TfUniversal, amount, amount2, nil, noLPTokenIn),
		},
		{
			name: "pass - canonical amount matches pool",
			tx:   withdraw(TfTwoAsset, canonicalAmount, amount2, nil, noLPTokenIn),
		},
		{
			name: "pass - MPT amount matches pool with different hex case",
			tx: withAssets(withdraw(TfTwoAsset, mptAmount, amount2, nil, noLPTokenIn),
				mptAsset, ledger.Asset{Currency: "XRP"}),
		},
		{
			name: "fail - invalid AMMWithdraw BaseTx without TransactionType",
			tx: func() *AMMWithdraw {
				tx := withdraw(TfTwoAsset, amount, amount2, nil, noLPTokenIn)
				tx.TransactionType = ""
				return tx
			}(),
			expectedErr: ErrInvalidTransactionType,
		},
		{
			name:        "fail - no withdrawal mode flag",
			tx:          withdraw(0, amount, amount2, nil, noLPTokenIn),
			expectedErr: ErrInvalidFlags,
		},
		{
			name:        "fail - two withdrawal mode flags",
			tx:          withdraw(TfTwoAsset|TfWithdrawAll, amount, amount2, nil, noLPTokenIn),
			expectedErr: ErrInvalidFlags,
		},
		{
			name:        "fail - undefined flag with one withdrawal mode",
			tx:          withdraw(TfTwoAsset|0x2, amount, amount2, nil, noLPTokenIn),
			expectedErr: ErrInvalidFlags,
		},
		{
			name:        "fail - deposit-only TfTwoAssetIfEmpty flag",
			tx:          withdraw(TfTwoAsset|TfTwoAssetIfEmpty, amount, amount2, nil, noLPTokenIn),
			expectedErr: ErrInvalidFlags,
		},
		{
			name:        "fail - Amount2 without Amount",
			tx:          withdraw(TfTwoAsset, nil, amount2, nil, noLPTokenIn),
			expectedErr: ErrAMMMustSetAmountWithAmount2,
		},
		{
			name:        "fail - EPrice without Amount",
			tx:          withdraw(TfLimitLPToken, nil, nil, ePrice, noLPTokenIn),
			expectedErr: ErrAMMMustSetAmountWithEPrice,
		},
		{
			name:        "fail - TfLPToken without LPTokenIn",
			tx:          withdraw(TfLPToken, nil, nil, nil, noLPTokenIn),
			expectedErr: ErrMissingField{Field: "LPTokenIn"},
		},
		{
			name:        "fail - TfLPToken with Amount",
			tx:          withdraw(TfLPToken, amount, nil, nil, lpTokenIn),
			expectedErr: invalidField("Amount"),
		},
		{
			name:        "fail - TfWithdrawAll with LPTokenIn",
			tx:          withdraw(TfWithdrawAll, nil, nil, nil, lpTokenIn),
			expectedErr: invalidField("LPTokenIn"),
		},
		{
			name:        "fail - TfWithdrawAll with Amount",
			tx:          withdraw(TfWithdrawAll, amount, nil, nil, noLPTokenIn),
			expectedErr: invalidField("Amount"),
		},
		{
			name:        "fail - TfOneAssetWithdrawAll without Amount",
			tx:          withdraw(TfOneAssetWithdrawAll, nil, nil, nil, noLPTokenIn),
			expectedErr: ErrMissingField{Field: "Amount"},
		},
		{
			name:        "fail - TfOneAssetWithdrawAll with LPTokenIn",
			tx:          withdraw(TfOneAssetWithdrawAll, amount, nil, nil, lpTokenIn),
			expectedErr: invalidField("LPTokenIn"),
		},
		{
			name:        "fail - TfOneAssetWithdrawAll with Amount2",
			tx:          withdraw(TfOneAssetWithdrawAll, amount, amount2, nil, noLPTokenIn),
			expectedErr: invalidField("Amount2"),
		},
		{
			name:        "fail - TfOneAssetWithdrawAll with EPrice",
			tx:          withdraw(TfOneAssetWithdrawAll, amount, nil, ePrice, noLPTokenIn),
			expectedErr: invalidField("EPrice"),
		},
		{
			name:        "fail - TfSingleAsset without Amount",
			tx:          withdraw(TfSingleAsset, nil, nil, nil, lpTokenIn),
			expectedErr: ErrMissingField{Field: "Amount"},
		},
		{
			name:        "fail - TfSingleAsset with LPTokenIn",
			tx:          withdraw(TfSingleAsset, amount, nil, nil, lpTokenIn),
			expectedErr: invalidField("LPTokenIn"),
		},
		{
			name:        "fail - TfSingleAsset with Amount2",
			tx:          withdraw(TfSingleAsset, amount, amount2, nil, noLPTokenIn),
			expectedErr: invalidField("Amount2"),
		},
		{
			name:        "fail - TfSingleAsset with EPrice",
			tx:          withdraw(TfSingleAsset, amount, nil, ePrice, noLPTokenIn),
			expectedErr: invalidField("EPrice"),
		},
		{
			name:        "fail - TfTwoAsset without Amount",
			tx:          withdraw(TfTwoAsset, nil, nil, nil, lpTokenIn),
			expectedErr: ErrMissingField{Field: "Amount"},
		},
		{
			name:        "fail - TfTwoAsset without Amount2",
			tx:          withdraw(TfTwoAsset, amount, nil, nil, noLPTokenIn),
			expectedErr: ErrMissingField{Field: "Amount2"},
		},
		{
			name:        "fail - TfTwoAsset with LPTokenIn",
			tx:          withdraw(TfTwoAsset, amount, amount2, nil, lpTokenIn),
			expectedErr: invalidField("LPTokenIn"),
		},
		{
			name:        "fail - TfTwoAsset with EPrice",
			tx:          withdraw(TfTwoAsset, amount, amount2, ePrice, noLPTokenIn),
			expectedErr: invalidField("EPrice"),
		},
		{
			name:        "fail - TfOneAssetLPToken without Amount",
			tx:          withdraw(TfOneAssetLPToken, nil, nil, nil, lpTokenIn),
			expectedErr: ErrMissingField{Field: "Amount"},
		},
		{
			name:        "fail - TfOneAssetLPToken without LPTokenIn",
			tx:          withdraw(TfOneAssetLPToken, amount, nil, nil, noLPTokenIn),
			expectedErr: ErrMissingField{Field: "LPTokenIn"},
		},
		{
			name:        "fail - TfOneAssetLPToken with Amount2",
			tx:          withdraw(TfOneAssetLPToken, amount, amount2, nil, lpTokenIn),
			expectedErr: invalidField("Amount2"),
		},
		{
			name:        "fail - TfOneAssetLPToken with EPrice",
			tx:          withdraw(TfOneAssetLPToken, amount, nil, ePrice, lpTokenIn),
			expectedErr: invalidField("EPrice"),
		},
		{
			name:        "fail - TfLimitLPToken without Amount",
			tx:          withdraw(TfLimitLPToken, nil, nil, nil, lpTokenIn),
			expectedErr: ErrMissingField{Field: "Amount"},
		},
		{
			name:        "fail - TfLimitLPToken without EPrice",
			tx:          withdraw(TfLimitLPToken, amount, nil, nil, noLPTokenIn),
			expectedErr: ErrMissingField{Field: "EPrice"},
		},
		{
			name:        "fail - TfLimitLPToken with LPTokenIn",
			tx:          withdraw(TfLimitLPToken, amount, nil, ePrice, lpTokenIn),
			expectedErr: invalidField("LPTokenIn"),
		},
		{
			name:        "fail - TfLimitLPToken with Amount2",
			tx:          withdraw(TfLimitLPToken, amount, amount2, ePrice, noLPTokenIn),
			expectedErr: invalidField("Amount2"),
		},
		{
			name: "fail - same canonical pool issue",
			tx: withAssets(withdraw(TfWithdrawAll, nil, nil, nil, noLPTokenIn),
				ledger.Asset{Currency: "TST", Issuer: issuer},
				ledger.Asset{Currency: canonicalAmount.Currency, Issuer: canonicalAmount.Issuer}),
			expectedErr: ErrAMMInvalidTokens,
		},
		{
			name: "fail - same MPT pool issue with different hex case",
			tx: withAssets(withdraw(TfWithdrawAll, nil, nil, nil, noLPTokenIn),
				mptAsset, ledger.Asset{MPTIssuanceID: mptAmount.MPTIssuanceID}),
			expectedErr: ErrAMMInvalidTokens,
		},
		{
			name: "fail - both pool assets are XRP",
			tx: withAssets(withdraw(TfWithdrawAll, nil, nil, nil, noLPTokenIn),
				ledger.Asset{Currency: "XRP"}, ledger.Asset{Currency: "XRP"}),
			expectedErr: ErrAMMInvalidTokens,
		},
		{
			name:        "fail - same canonical amount issues",
			tx:          withdraw(TfTwoAsset, amount, canonicalAmount, nil, noLPTokenIn),
			expectedErr: ErrAMMInvalidTokens,
		},
		{
			name:        "fail - Amount outside pool",
			tx:          withdraw(TfSingleAsset, lpTokenIn, nil, nil, noLPTokenIn),
			expectedErr: ErrAMMInvalidTokens,
		},
		{
			name:        "fail - Amount2 outside pool",
			tx:          withdraw(TfTwoAsset, amount, lpTokenIn, nil, noLPTokenIn),
			expectedErr: ErrAMMInvalidTokens,
		},
		{
			name:        "fail - zero LPTokenIn",
			tx:          withdraw(TfLPToken, nil, nil, nil, zeroLPTokenIn),
			expectedErr: ErrInvalidTokenValue,
		},
		{
			name:        "fail - negative LPTokenIn",
			tx:          withdraw(TfLPToken, nil, nil, nil, negativeLPTokenIn),
			expectedErr: ErrInvalidTokenValue,
		},
		{
			name:        "fail - invalid LPTokenIn",
			tx:          withdraw(TfLPToken, nil, nil, nil, types.IssuedCurrencyAmount{Value: "12", Issuer: issuer}),
			expectedErr: ErrInvalidTokenFields,
		},
		{
			name:        "fail - zero Amount without EPrice",
			tx:          withdraw(TfSingleAsset, types.XRPCurrencyAmount(0), nil, nil, noLPTokenIn),
			expectedErr: ErrInvalidTokenValue,
		},
		{
			name:        "fail - zero Amount2",
			tx:          withdraw(TfTwoAsset, amount, types.XRPCurrencyAmount(0), nil, noLPTokenIn),
			expectedErr: ErrInvalidTokenValue,
		},
		{
			name:        "fail - zero EPrice",
			tx:          withdraw(TfLimitLPToken, amount, nil, types.XRPCurrencyAmount(0), noLPTokenIn),
			expectedErr: ErrInvalidTokenValue,
		},
		{
			name:        "fail - MPT EPrice",
			tx:          withdraw(TfLimitLPToken, amount, nil, mptAmount, noLPTokenIn),
			expectedErr: invalidField("EPrice"),
		},
		{
			name: "fail - invalid Asset",
			tx: func() *AMMWithdraw {
				tx := withdraw(TfWithdrawAll, nil, nil, nil, noLPTokenIn)
				tx.Asset.Currency = ""
				return tx
			}(),
			expectedErr: ErrMissingAssetCurrency,
		},
		{
			name: "fail - invalid Asset2",
			tx: func() *AMMWithdraw {
				tx := withdraw(TfWithdrawAll, nil, nil, nil, noLPTokenIn)
				tx.Asset2.Issuer = issuer
				return tx
			}(),
			expectedErr: ErrInvalidAssetIssuer,
		},
		{
			name:        "fail - invalid Amount",
			tx:          withdraw(TfSingleAsset, types.IssuedCurrencyAmount{Value: "1"}, nil, nil, noLPTokenIn),
			expectedErr: ErrInvalidTokenFields,
		},
		{
			name:        "fail - invalid Amount2",
			tx:          withdraw(TfTwoAsset, amount, types.IssuedCurrencyAmount{Value: "1"}, nil, noLPTokenIn),
			expectedErr: ErrInvalidTokenFields,
		},
		{
			name:        "fail - invalid EPrice",
			tx:          withdraw(TfLimitLPToken, amount, nil, types.IssuedCurrencyAmount{Value: "1"}, noLPTokenIn),
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
