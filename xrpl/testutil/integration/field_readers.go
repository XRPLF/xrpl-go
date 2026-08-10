package integration

import (
	"encoding/json"
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/stretchr/testify/require"
)

// DecodeLedgerObject converts a raw ledger response into its typed ledger-entry model.
func DecodeLedgerObject[T any](t *testing.T, object ledger.FlatLedgerObject) T {
	t.Helper()

	encoded, err := json.Marshal(object)
	require.NoError(t, err)

	var decoded T
	require.NoError(t, json.Unmarshal(encoded, &decoded))
	return decoded
}

// TxFieldUint32 reads the field from a transaction where the expected uint32 is not certain
func TxFieldUint32(t *testing.T, tx map[string]any, field string) uint32 {
	return uint32(TxFieldFloat64(t, tx, field))
}

// TxFieldFloat64 reads the field from a transaction where the expected float64 is not certain
func TxFieldFloat64(t *testing.T, m map[string]any, field string) float64 {
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
