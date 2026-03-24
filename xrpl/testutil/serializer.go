package testutil

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

// Deserialize unmarshals the JSON string d into s and verifies the result matches the original.
func Deserialize(s any, d string) error {
	decode := reflect.New(reflect.TypeOf(s))
	err := json.Unmarshal([]byte(d), decode.Interface())
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(s, decode.Elem().Interface()) {
		return fmt.Errorf("json decoding does not match expected struct")
	}
	return nil
}

// Serialize marshals s into JSON and asserts it equals the expected string d.
func Serialize(t *testing.T, s any, d string) error {
	j, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		return err
	}
	require.Equal(t, d, string(j), "json encoding does not match expected string")
	return nil
}

// SerializeAndDeserialize runs Serialize then Deserialize for s and d.
func SerializeAndDeserialize(t *testing.T, s any, d string) error {
	if err := Serialize(t, s, d); err != nil {
		return err
	}
	return Deserialize(s, d)
}
