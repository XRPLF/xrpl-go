package builder

import (
	"encoding/json"
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRequiredString covers a field the caller cannot proceed without. Absent, empty, and
// wrongly typed are the same failure, because none of them yields a value to use.
func TestRequiredString(t *testing.T) {
	tests := []struct {
		name    string
		node    map[string]any
		want    string
		wantErr bool
	}{
		{name: "value", node: map[string]any{"Field": "value"}, want: "value"},
		{name: "missing", node: map[string]any{}, wantErr: true},
		{name: "empty", node: map[string]any{"Field": ""}, wantErr: true},
		{name: "wrong type", node: map[string]any{"Field": 1}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := requiredString(test.node, "Field")
			if test.wantErr {
				require.ErrorIs(t, err, ErrInvalidLedgerState)
				require.ErrorContains(t, err, "Field")
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.want, value)
		})
	}
}

// TestOptionalString covers a field that may be absent. Only absence is allowed: a field the
// server did write but that carries nothing usable is malformed, not optional.
func TestOptionalString(t *testing.T) {
	tests := []struct {
		name    string
		node    map[string]any
		want    string
		wantErr bool
	}{
		{name: "value", node: map[string]any{"Field": "value"}, want: "value"},
		{name: "missing", node: map[string]any{}},
		{name: "empty", node: map[string]any{"Field": ""}, wantErr: true},
		{name: "wrong type", node: map[string]any{"Field": 1}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := optionalString(test.node, "Field")
			if test.wantErr {
				require.ErrorIs(t, err, ErrInvalidLedgerState)
				require.ErrorContains(t, err, "Field")
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.want, value)
		})
	}
}

// TestOptionalUint32 covers a numeric field an entry omits while it holds its default. rpc
// decodes with UseNumber and websocket decodes into any, so both encodings of a JSON number
// must be accepted and anything that is not an exact whole number in range must not be.
func TestOptionalUint32(t *testing.T) {
	tests := []struct {
		name    string
		node    map[string]any
		want    uint32
		wantErr bool
	}{
		{name: "absent", node: map[string]any{}},
		{name: "RPC zero", node: map[string]any{"Field": json.Number("0")}},
		{name: "RPC json number", node: map[string]any{"Field": json.Number("2")}, want: 2},
		{name: "RPC maximum", node: map[string]any{"Field": json.Number("4294967295")}, want: math.MaxUint32},
		{name: "WebSocket zero", node: map[string]any{"Field": float64(0)}},
		{name: "WebSocket float", node: map[string]any{"Field": float64(3)}, want: 3},
		{name: "WebSocket maximum", node: map[string]any{"Field": float64(4294967295)}, want: math.MaxUint32},
		{name: "explicit null", node: map[string]any{"Field": nil}, wantErr: true},
		{name: "fractional JSON number", node: map[string]any{"Field": json.Number("2.5")}, wantErr: true},
		{name: "negative JSON number", node: map[string]any{"Field": json.Number("-1")}, wantErr: true},
		{name: "malformed JSON number", node: map[string]any{"Field": json.Number("nope")}, wantErr: true},
		{name: "JSON number overflow", node: map[string]any{"Field": json.Number("4294967296")}, wantErr: true},
		{name: "float overflow", node: map[string]any{"Field": float64(4294967296)}, wantErr: true},
		{name: "negative float", node: map[string]any{"Field": float64(-1)}, wantErr: true},
		{name: "fractional float", node: map[string]any{"Field": 2.5}, wantErr: true},
		{name: "NaN float", node: map[string]any{"Field": math.NaN()}, wantErr: true},
		{name: "positive infinity", node: map[string]any{"Field": math.Inf(1)}, wantErr: true},
		{name: "negative infinity", node: map[string]any{"Field": math.Inf(-1)}, wantErr: true},
		{name: "wrong string type", node: map[string]any{"Field": "2"}, wantErr: true},
		{name: "negative integer type", node: map[string]any{"Field": -1}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := optionalUint32(test.node, "Field")
			if test.wantErr {
				require.ErrorIs(t, err, ErrInvalidLedgerState)
				require.ErrorContains(t, err, "Field")
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.want, value)
		})
	}
}

// TestOptionalUint64 covers a ledger amount field. MPT amounts run past the range a float64
// represents exactly, so they are serialized as decimal strings and a float is not accepted.
func TestOptionalUint64(t *testing.T) {
	tests := []struct {
		name    string
		node    map[string]any
		want    uint64
		wantErr bool
	}{
		{name: "absent", node: map[string]any{}},
		{name: "decimal string", node: map[string]any{"Field": "1000"}, want: 1000},
		{name: "zero string", node: map[string]any{"Field": "0"}},
		{name: "maximum string", node: map[string]any{"Field": strconv.FormatUint(math.MaxUint64, 10)}, want: math.MaxUint64},
		{name: "json number", node: map[string]any{"Field": json.Number("42")}, want: 42},
		{name: "negative string", node: map[string]any{"Field": "-1"}, wantErr: true},
		{name: "malformed string", node: map[string]any{"Field": "nope"}, wantErr: true},
		{name: "float is not exact", node: map[string]any{"Field": float64(1000)}, wantErr: true},
		{name: "explicit null", node: map[string]any{"Field": nil}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := optionalUint64(test.node, "Field")
			if test.wantErr {
				require.ErrorIs(t, err, ErrInvalidLedgerState)
				require.ErrorContains(t, err, "Field")
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.want, value)
		})
	}
}
