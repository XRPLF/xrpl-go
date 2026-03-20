// Package testutil provides utilities for testing JSON flattening and serialization.
package testutil

import (
	"encoding/json"
	"fmt"
)

// CompareFlattenAndExpected compares a flattened map and expected JSON bytes and returns an error if they differ.
func CompareFlattenAndExpected(flattened map[string]any, expected []byte) error {
	// Convert flattened to JSON
	flattenedJSON, err := json.Marshal(flattened)
	if err != nil {
		return fmt.Errorf("%w, error: %w", ErrMarshalingPaymentFlattened, err)
	}

	// Normalize expected JSON
	var expectedMap map[string]any
	if err := json.Unmarshal([]byte(expected), &expectedMap); err != nil {
		return fmt.Errorf("%w, error: %w", ErrUnmarshalingExpected, err)
	}
	expectedJSON, err := json.Marshal(expectedMap)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMarshalingExpectedPaymentObject, err)
	}

	// Compare JSON strings
	if string(flattenedJSON) != string(expectedJSON) {
		return fmt.Errorf("%w.\nGot:      %v\nExpected: %v", ErrFlattenedAndExpectedJSONNotEqual, string(flattenedJSON), string(expectedJSON))
	}

	return nil
}
