package builder

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Peersyst/xrpl-go/pkg/typecheck"
)

// Ledger entries reach the builders as decoded JSON maps, so every field starts as an any.
// rpc decodes with UseNumber and websocket decodes into any, which means a numeric field
// arrives as a json.Number or a float64 and never as an integer. The readers below turn one
// such field into the typed value a builder can use, and report anything unusable as
// ErrInvalidLedgerState naming the field it came from.

// requiredString reads a field the caller cannot proceed without.
func requiredString(node map[string]any, field string) (string, error) {
	value, err := optionalString(node, field)
	if err != nil {
		return "", err
	}
	if value == "" {
		return "", malformedLedgerField(field)
	}
	return value, nil
}

// optionalString reads a field that may be absent. A present but unusable value is
// still malformed, so only an absent field yields an empty string.
func optionalString(node map[string]any, field string) (string, error) {
	raw, present := node[field]
	if !present {
		return "", nil
	}
	value, ok := raw.(string)
	if !ok || value == "" {
		return "", malformedLedgerField(field)
	}
	return value, nil
}

// optionalUint32 reads a numeric field that an entry omits while it holds its default value.
// An absent field is zero, so only a present but unreadable value is malformed.
func optionalUint32(node map[string]any, field string) (uint32, error) {
	raw, present := node[field]
	if !present {
		return 0, nil
	}
	value, ok := typecheck.ToUint32(raw)
	if !ok {
		return 0, malformedLedgerField(field)
	}
	return value, nil
}

// optionalUint64 reads a ledger amount field that may be absent. MPT amounts are serialized
// as decimal strings, so a float64 cannot represent one exactly and is not accepted.
func optionalUint64(node map[string]any, field string) (uint64, error) {
	raw, present := node[field]
	if !present {
		return 0, nil
	}

	var text string
	switch amount := raw.(type) {
	case string:
		text = amount
	case json.Number:
		text = amount.String()
	default:
		return 0, malformedLedgerField(field)
	}

	value, err := strconv.ParseUint(text, 10, 64)
	if err != nil {
		return 0, malformedLedgerField(field)
	}
	return value, nil
}

func malformedLedgerField(field string) error {
	return fmt.Errorf("%w: field %s is missing or malformed", ErrInvalidLedgerState, field)
}
