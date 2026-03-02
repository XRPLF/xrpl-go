package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/assert"
)

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
