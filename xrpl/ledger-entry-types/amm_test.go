package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/assert"
)

func TestAssetFlatten(t *testing.T) {
	asset := Asset{
		Currency: "USD",
		Issuer:   "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	}

	json := `{
	"currency": "USD",
	"issuer": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"
}`

	if err := testutil.CompareFlattenAndExpected(asset.Flatten(), []byte(json)); err != nil {
		t.Error(err)
	}

	// 2nd test with issuer empty
	asset2 := Asset{
		Currency: "XRP",
	}

	json2 := `{
	"currency": "XRP"
}`

	if err := testutil.CompareFlattenAndExpected(asset2.Flatten(), []byte(json2)); err != nil {
		t.Error(err)
	}
}

func TestAMM_EntryType(t *testing.T) {
	entry := &AMM{}
	assert.Equal(t, AMMEntry, entry.EntryType())
}
func TestAuthAccounts_Flatten(t *testing.T) {
	authAccount := AuthAccount{
		Account: "rExampleAccount",
	}

	authAccounts := AuthAccounts{
		AuthAccount: authAccount,
	}

	expectedJSON := `{
	"AuthAccount": {
		"Account": "rExampleAccount"
	}
}`

	if err := testutil.CompareFlattenAndExpected(authAccounts.Flatten(), []byte(expectedJSON)); err != nil {
		t.Error(err)
	}
}

func TestAuthAccount_Flatten(t *testing.T) {
	authAccount := AuthAccount{
		Account: "rExampleAccount",
	}

	expectedJSON := `{
	"Account": "rExampleAccount"
}`

	if err := testutil.CompareFlattenAndExpected(authAccount.Flatten(), []byte(expectedJSON)); err != nil {
		t.Error(err)
	}
}
