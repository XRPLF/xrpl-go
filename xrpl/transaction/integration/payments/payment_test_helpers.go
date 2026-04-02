package payment

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func txFieldUint32(t *testing.T, tx map[string]any, field string) uint32 {
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
