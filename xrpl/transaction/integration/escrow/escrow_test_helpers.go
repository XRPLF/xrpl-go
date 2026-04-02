package escrow

import (
	"encoding/json"
	"testing"
	"time"

	queriesCommon "github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

func getLedgerCloseTime(t *testing.T, client integration.Client) int64 {
	t.Helper()
	req := &ledger.Request{
		LedgerIndex: queriesCommon.Validated,
	}
	var res *ledger.Response
	var err error
	switch c := client.(type) {
	case *websocket.Client:
		res, err = c.GetLedger(req)
	case *rpc.Client:
		res, err = c.GetLedger(req)
	default:
		t.Fatal("unsupported client type for getLedgerCloseTime")
	}
	require.NoError(t, err)
	return int64(res.Ledger.CloseTime)
}

func waitForLedgerTime(t *testing.T, client integration.Client, target int64) {
	t.Helper()
	const rippleOffset int64 = 946684800
	currentRippleTime := time.Now().Unix() - rippleOffset
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

func txFieldUint32(t *testing.T, tx map[string]any) uint32 {
	t.Helper()
	switch v := tx["Sequence"].(type) {
	case float64:
		return uint32(v)
	case json.Number:
		n, err := v.Float64()
		require.NoError(t, err)
		return uint32(n)
	default:
		t.Fatalf("unexpected type for tx field %q: %T", "Sequence", tx["Sequence"])
		return 0
	}
}
