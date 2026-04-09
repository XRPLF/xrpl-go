package escrow

import (
	"testing"
	"time"

	queriesCommon "github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	xrpltime "github.com/Peersyst/xrpl-go/xrpl/time"
	"github.com/stretchr/testify/require"
)

func getLedgerCloseTime(t *testing.T, client integration.Client) int64 {
	t.Helper()
	res, err := client.GetLedger(&ledger.Request{
		LedgerIndex: queriesCommon.Validated,
	})
	require.NoError(t, err)
	return int64(res.Ledger.CloseTime)
}

func waitForLedgerTime(t *testing.T, client integration.Client, target int64) {
	t.Helper()

	currentRippleTime := xrpltime.UnixTimeToRippleTime(time.Now().UTC().Unix())
	if target > currentRippleTime {
		waitDuration := time.Duration(target-currentRippleTime) * time.Second
		time.Sleep(waitDuration)
	}
	for range 30 {
		if getLedgerCloseTime(t, client) > target {
			return
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("ledger close_time did not reach %d after 30 attempts", target)
}
