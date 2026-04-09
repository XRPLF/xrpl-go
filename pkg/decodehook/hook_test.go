package decodehook

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// customField is a test type that implements json.Unmarshaler.
// It accepts both JSON strings and numbers, always storing the value as a string.
type customField struct {
	Value string
}

func (c *customField) UnmarshalJSON(data []byte) error {
	if data[0] == '"' {
		return json.Unmarshal(data, &c.Value)
	}
	c.Value = string(data)
	return nil
}

func TestJSON(t *testing.T) {
	type target struct {
		Field customField `json:"field"`
	}

	tests := []struct {
		name      string
		input     map[string]any
		expected  string
		expectErr bool
	}{
		{
			name:     "delegates to UnmarshalJSON with string value",
			input:    map[string]any{"field": "hello"},
			expected: "hello",
		},
		{
			name:     "delegates to UnmarshalJSON with numeric value",
			input:    map[string]any{"field": json.Number("42")},
			expected: "42",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var result target
			dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				TagName:    "json",
				Result:     &result,
				DecodeHook: JSON(),
			})
			require.NoError(t, err)
			err = dec.Decode(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result.Field.Value)
		})
	}

	t.Run("skips types that do not implement json.Unmarshaler", func(t *testing.T) {
		hook := JSON()
		from := reflect.ValueOf("plain")
		to := reflect.New(reflect.TypeFor[string]()).Elem()
		out, err := hook(from, to)
		require.NoError(t, err)
		assert.Equal(t, "plain", out)
	})

	t.Run("returns error when source cannot be marshaled", func(t *testing.T) {
		hook := JSON()
		from := reflect.ValueOf(make(chan int))
		to := reflect.New(reflect.TypeFor[customField]()).Elem()
		_, err := hook(from, to)
		require.Error(t, err)
	})
}
