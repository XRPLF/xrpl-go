package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/assert"
)

func TestAssetFlatten(t *testing.T) {
	t.Run("IOU asset with currency and issuer", func(t *testing.T) {
		asset := Asset{
			Currency: "USD",
			Issuer:   "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		}

		expected := `{
	"currency": "USD",
	"issuer": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"
}`

		if err := testutil.CompareFlattenAndExpected(asset.Flatten(), []byte(expected)); err != nil {
			t.Error(err)
		}
	})

	t.Run("XRP asset with currency only", func(t *testing.T) {
		asset := Asset{
			Currency: "XRP",
		}

		expected := `{
	"currency": "XRP"
}`

		if err := testutil.CompareFlattenAndExpected(asset.Flatten(), []byte(expected)); err != nil {
			t.Error(err)
		}
	})

	t.Run("MPT asset with mpt_issuance_id only", func(t *testing.T) {
		asset := Asset{
			MPTIssuanceID: "0000000169F415C9F1AB6796AB9224CE635818AFD74F8175",
		}

		expected := `{
	"mpt_issuance_id": "0000000169F415C9F1AB6796AB9224CE635818AFD74F8175"
}`

		if err := testutil.CompareFlattenAndExpected(asset.Flatten(), []byte(expected)); err != nil {
			t.Error(err)
		}
	})
}

func TestAssetKind(t *testing.T) {
	t.Run("XRP asset - currency only, no issuer", func(t *testing.T) {
		asset := Asset{Currency: "XRP"}
		assert.Equal(t, AssetXRP, asset.Kind())
	})

	t.Run("IOU asset - currency and issuer", func(t *testing.T) {
		asset := Asset{
			Currency: "USD",
			Issuer:   "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		}
		assert.Equal(t, AssetIOU, asset.Kind())
	})

	t.Run("MPT asset - mpt_issuance_id only", func(t *testing.T) {
		asset := Asset{
			MPTIssuanceID: "0000000169F415C9F1AB6796AB9224CE635818AFD74F8175",
		}
		assert.Equal(t, AssetMPT, asset.Kind())
	})
}
