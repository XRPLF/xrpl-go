package oracle

import (
	"encoding/json"
	"testing"

	queriesCommon "github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

func getLedgerCloseTime(t *testing.T, client integration.Client) uint64 {
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
	return uint64(res.Ledger.CloseTime)
}

func txFieldUint32(t *testing.T, tx map[string]any, field string) uint32 {
	t.Helper()
	switch v := tx[field].(type) {
	case float64:
		return uint32(v)
	case json.Number:
		n, err := v.Int64()
		require.NoError(t, err)
		return uint32(n)
	default:
		t.Fatalf("unexpected type for tx field %q: %T", field, tx[field])
		return 0
	}
}

func txFieldFloat64(t *testing.T, m map[string]any, field string) float64 {
	t.Helper()
	switch v := m[field].(type) {
	case float64:
		return v
	case json.Number:
		n, err := v.Float64()
		require.NoError(t, err)
		return n
	default:
		t.Fatalf("unexpected type for field %q: %T", field, m[field])
		return 0
	}
}
