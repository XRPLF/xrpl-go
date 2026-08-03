package ledger

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

const maxUInt64 = uint64(18446744073709551615)

func TestQuotedUInt64JSONMarshal(t *testing.T) {
	tests := []struct {
		name  string
		value quotedUInt64
		want  string
	}{
		{name: "pass - zero", value: 0, want: `"0"`},
		{name: "pass - maximum", value: quotedUInt64(maxUInt64), want: `"18446744073709551615"`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := json.Marshal(test.value)
			require.NoError(t, err)
			require.JSONEq(t, test.want, string(encoded))
		})
	}
}

func TestQuotedUInt64JSONUnmarshal(t *testing.T) {
	valid := []struct {
		name  string
		input string
		want  quotedUInt64
	}{
		{name: "pass - quoted zero", input: `"0"`, want: 0},
		{name: "pass - quoted maximum", input: `"18446744073709551615"`, want: quotedUInt64(maxUInt64)},
		{name: "pass - unsigned integer", input: `42`, want: 42},
		{name: "pass - maximum unsigned integer", input: `18446744073709551615`, want: quotedUInt64(maxUInt64)},
	}

	for _, test := range valid {
		t.Run(test.name, func(t *testing.T) {
			var got quotedUInt64
			require.NoError(t, json.Unmarshal([]byte(test.input), &got))
			require.Equal(t, test.want, got)
		})
	}

	invalid := []struct {
		name  string
		input string
	}{
		{name: "fail - empty string", input: `""`},
		{name: "fail - non-decimal string", input: `"invalid"`},
		{name: "fail - negative string", input: `"-1"`},
		{name: "fail - fractional string", input: `"1.5"`},
		{name: "fail - too many decimal digits", input: `"000000000000000000001"`},
		{name: "fail - quoted overflow", input: `"18446744073709551616"`},
		{name: "fail - negative number", input: `-1`},
		{name: "fail - fractional number", input: `1.5`},
		{name: "fail - numeric overflow", input: `18446744073709551616`},
	}

	for _, test := range invalid {
		t.Run(test.name, func(t *testing.T) {
			var got quotedUInt64
			require.Error(t, json.Unmarshal([]byte(test.input), &got))
		})
	}
}

func TestHexUInt64JSONMarshal(t *testing.T) {
	tests := []struct {
		name  string
		value hexUInt64
		want  string
	}{
		{name: "pass - zero", value: 0, want: `"0000000000000000"`},
		{name: "pass - uppercase and zero padded", value: 10, want: `"000000000000000A"`},
		{name: "pass - maximum", value: hexUInt64(maxUInt64), want: `"FFFFFFFFFFFFFFFF"`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := json.Marshal(test.value)
			require.NoError(t, err)
			require.JSONEq(t, test.want, string(encoded))
		})
	}
}

func TestHexUInt64JSONUnmarshal(t *testing.T) {
	valid := []struct {
		name  string
		input string
		want  hexUInt64
	}{
		{name: "pass - short rippled form", input: `"a"`, want: 10},
		{name: "pass - zero-padded canonical form", input: `"000000000000000A"`, want: 10},
		{name: "pass - quoted maximum", input: `"FFFFFFFFFFFFFFFF"`, want: hexUInt64(maxUInt64)},
		{name: "pass - unsigned integer", input: `42`, want: 42},
		{name: "pass - maximum unsigned integer", input: `18446744073709551615`, want: hexUInt64(maxUInt64)},
	}

	for _, test := range valid {
		t.Run(test.name, func(t *testing.T) {
			var got hexUInt64
			require.NoError(t, json.Unmarshal([]byte(test.input), &got))
			require.Equal(t, test.want, got)
		})
	}

	invalid := []struct {
		name  string
		input string
	}{
		{name: "fail - empty string", input: `""`},
		{name: "fail - non-hex string", input: `"invalid"`},
		{name: "fail - negative string", input: `"-1"`},
		{name: "fail - fractional string", input: `"1.5"`},
		{name: "fail - too many hexadecimal digits", input: `"00000000000000001"`},
		{name: "fail - negative number", input: `-1`},
		{name: "fail - fractional number", input: `1.5`},
		{name: "fail - numeric overflow", input: `18446744073709551616`},
	}

	for _, test := range invalid {
		t.Run(test.name, func(t *testing.T) {
			var got hexUInt64
			require.Error(t, json.Unmarshal([]byte(test.input), &got))
		})
	}
}
