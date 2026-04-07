package integration

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// TxFieldUint32 reads the field from a transaction where the expected uint32 is not certain
func TxFieldUint32(t *testing.T, tx map[string]any, field string) uint32 {
	t.Helper()
	switch v := tx[field].(type) {
	case float64:
		return uint32(v)
	case json.Number:
		n, err := v.Float64()
		require.NoError(t, err)
		return uint32(n)
	default:
		t.Fatalf("unexpected type for tx field %q: %T", field, tx[field])
		return 0
	}
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
