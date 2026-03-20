package account

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

func TestAccountOffersRequest(t *testing.T) {
	s := OffersRequest{
		Account:     "abc",
		LedgerIndex: common.LedgerIndex(10),
		Marker:      "123",
	}
	j := `{
	"account": "abc",
	"ledger_index": 10,
	"marker": "123"
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}
