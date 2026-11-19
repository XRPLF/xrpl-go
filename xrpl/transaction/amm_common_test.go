package transaction

import (
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

// ammPoolValidators validate each AMM transaction that names a pool by its two assets.
// AMMCreate names the pool through the issues of its amounts.
var ammPoolValidators = map[string]func(asset, asset2 ledger.Asset) (bool, error){
	"AMMBid": func(asset, asset2 ledger.Asset) (bool, error) {
		tx := AMMBid{BaseTx: BaseTx{Account: "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1", TransactionType: AMMBidTx}, Asset: asset, Asset2: asset2}
		return tx.Validate()
	},
	"AMMVote": func(asset, asset2 ledger.Asset) (bool, error) {
		tx := AMMVote{BaseTx: BaseTx{Account: "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1", TransactionType: AMMVoteTx}, Asset: asset, Asset2: asset2}
		return tx.Validate()
	},
	"AMMWithdraw": func(asset, asset2 ledger.Asset) (bool, error) {
		tx := AMMWithdraw{
			BaseTx: BaseTx{Account: "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1", TransactionType: AMMWithdrawTx, Flags: TfWithdrawAll},
			Asset:  asset,
			Asset2: asset2,
		}
		return tx.Validate()
	},
	"AMMCreate": func(asset, asset2 ledger.Asset) (bool, error) {
		tx := AMMCreate{
			BaseTx:  BaseTx{Account: "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1", TransactionType: AMMCreateTx},
			Amount:  ammPoolAmount(asset),
			Amount2: ammPoolAmount(asset2),
		}
		return tx.Validate()
	},
}

// ammPoolAmount returns a positive amount of asset.
func ammPoolAmount(asset ledger.Asset) types.CurrencyAmount {
	switch asset.Kind() {
	case ledger.AssetIOU:
		return types.IssuedCurrencyAmount{Currency: asset.Currency, Issuer: asset.Issuer, Value: "100"}
	case ledger.AssetMPT:
		return types.MPTCurrencyAmount{MPTIssuanceID: asset.MPTIssuanceID, Value: "100"}
	}
	return types.XRPCurrencyAmount(100)
}

func TestAMMPoolAssetsMustDiffer(t *testing.T) {
	const issuer = types.Address("rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd")
	xIssuer, err := addresscodec.ClassicAddressToXAddress(issuer.String(), 0, false, false)
	require.NoError(t, err)
	tst := ledger.Asset{Currency: "TST", Issuer: issuer}
	mpt := ledger.Asset{MPTIssuanceID: "00000001ABCDEF0123456789ABCDEF0123456789ABCDEF01"}

	tests := []struct {
		name          string
		asset, asset2 ledger.Asset
		wantErr       error
	}{
		{
			name:    "fail - same IOU",
			asset:   tst,
			asset2:  tst,
			wantErr: ErrAMMInvalidTokens,
		},
		{
			name:    "fail - same IOU with text and hex currency",
			asset:   tst,
			asset2:  ledger.Asset{Currency: "0000000000000000000000005453540000000000", Issuer: issuer},
			wantErr: ErrAMMInvalidTokens,
		},
		{
			name:    "fail - same IOU with classic and X-address issuer",
			asset:   tst,
			asset2:  ledger.Asset{Currency: "TST", Issuer: types.Address(xIssuer)},
			wantErr: ErrAMMInvalidTokens,
		},
		{
			name:    "fail - same MPT with different hex case",
			asset:   mpt,
			asset2:  ledger.Asset{MPTIssuanceID: "00000001abcdef0123456789abcdef0123456789abcdef01"},
			wantErr: ErrAMMInvalidTokens,
		},
		{
			name:   "pass - IOU and XRP",
			asset:  tst,
			asset2: ledger.Asset{Currency: "XRP"},
		},
		{
			name:   "pass - same currency from different issuers",
			asset:  tst,
			asset2: ledger.Asset{Currency: "TST", Issuer: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"},
		},
		{
			name:   "pass - different currencies from one issuer",
			asset:  tst,
			asset2: ledger.Asset{Currency: "USD", Issuer: issuer},
		},
		{
			name:   "pass - different MPTs",
			asset:  mpt,
			asset2: ledger.Asset{MPTIssuanceID: "00000002ABCDEF0123456789ABCDEF0123456789ABCDEF01"},
		},
		{
			name:   "pass - MPT and IOU",
			asset:  mpt,
			asset2: tst,
		},
	}

	for _, tt := range tests {
		for txType, validate := range ammPoolValidators {
			t.Run(tt.name+"/"+txType, func(t *testing.T) {
				ok, err := validate(tt.asset, tt.asset2)
				if tt.wantErr == nil {
					require.NoError(t, err)
					require.True(t, ok)
					return
				}
				require.ErrorIs(t, err, tt.wantErr)
				require.False(t, ok)
			})
		}
	}
}

// AMMBid keeps its own error for an XRP/XRP pool. The other transactions report the pair as equal.
func TestAMMPoolAssetsBothXRP(t *testing.T) {
	xrp := ledger.Asset{Currency: "XRP"}
	for txType, validate := range ammPoolValidators {
		t.Run(txType, func(t *testing.T) {
			wantErr := ErrAMMInvalidTokens
			if txType == "AMMBid" {
				wantErr = ErrAMMAtLeastOneAssetMustBeNonXRP
			}
			ok, err := validate(xrp, xrp)
			require.ErrorIs(t, err, wantErr)
			require.False(t, ok)
		})
	}
}
