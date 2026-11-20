package decimal

import (
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "integer", value: "42", want: "42"},
		{name: "signed fraction", value: "-1.5", want: "-3/2"},
		{name: "explicit plus sign", value: "+5", want: "5"},
		{name: "leading point", value: ".5", want: "1/2"},
		{name: "trailing point", value: "5.", want: "5"},
		{name: "exponent boundary", value: "1E+2", want: "100"},
		{name: "negative exponent boundary", value: "1e-2", want: "1/100"},
		{name: "input length boundary", value: strings.Repeat("0", MaxInputLen-1) + "7", want: "7"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			rat, ok := Parse(test.value, 2)
			require.True(t, ok)
			require.Equal(t, test.want, rat.RatString())
		})
	}
}

func TestParseRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
	}{
		{name: "input too long", value: strings.Repeat("0", MaxInputLen) + "7"},
		{name: "positive exponent too large", value: "1e3"},
		{name: "negative exponent too large", value: "1e-3"},
		{name: "exponent overflows int", value: "1e99999999999999999999"},
		{name: "empty", value: ""},
		{name: "text", value: "abc"},
		{name: "infinity", value: "Inf"},
		{name: "fraction", value: "1/2"},
		{name: "hexadecimal", value: "0x10"},
		{name: "underscore", value: "1_000"},
		{name: "inner sign", value: "1-2"},
		{name: "two points", value: "1.2.3"},
		{name: "missing exponent digits", value: "1e"},
		{name: "missing mantissa", value: "e2"},
		{name: "point only", value: "."},
		{name: "minus only", value: "-"},
		{name: "plus only", value: "+"},
		{name: "exponent sign without digits", value: "1e+"},
		{name: "two exponents", value: "1e5e5"},
		{name: "double sign", value: "--1"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			rat, ok := Parse(test.value, 2)
			require.False(t, ok)
			require.Nil(t, rat)
		})
	}
}

func TestParseAmountExponentLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		value  string
		wantOK bool
	}{
		{name: "largest exponent", value: "1e256", wantOK: true},
		{name: "smallest exponent", value: "1e-256", wantOK: true},
		{name: "exponent above limit", value: "1e257", wantOK: false},
		{name: "exponent below limit", value: "1e-257", wantOK: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, ok := ParseAmount(test.value)
			require.Equal(t, test.wantOK, ok)
		})
	}
}

func TestFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "zero", value: "0", want: "0"},
		{name: "negative zero", value: "-0", want: "0"},
		{name: "integer", value: "42", want: "42"},
		{name: "explicit plus sign", value: "+5", want: "5"},
		{name: "trailing fraction zeros", value: "1.50", want: "1.5"},
		{name: "negative fraction", value: "-123.456", want: "-123.456"},
		{name: "smallest plain exponent", value: "0.000001", want: "0.000001"},
		{name: "smallest plain exponent with digits", value: "0.0000015", want: "0.0000015"},
		{name: "first small exponential", value: "0.0000001", want: "1e-7"},
		{name: "largest small exponential", value: "0.00000099", want: "9.9e-7"},
		{name: "negative small exponential", value: "-0.00000015", want: "-1.5e-7"},
		{name: "largest plain exponent", value: "1e20", want: "100000000000000000000"},
		{name: "largest plain value", value: "999999999999999999999", want: "999999999999999999999"},
		{name: "first large exponential", value: "1e21", want: "1e+21"},
		{name: "large exponential with digits", value: "1234e18", want: "1.234e+21"},
		{name: "negative large exponential", value: "-1e21", want: "-1e+21"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			rat, ok := Parse(test.value, 21)
			require.True(t, ok)

			text, ok := Format(rat)
			require.True(t, ok)
			require.Equal(t, test.want, text)
		})
	}
}

func TestFormatRejectsNonTerminatingValue(t *testing.T) {
	t.Parallel()

	text, ok := Format(big.NewRat(1, 3))
	require.False(t, ok)
	require.Empty(t, text)
}

func FuzzParse(f *testing.F) {
	for _, seed := range []string{"", "0", "-1.5", "+5", "1e21", "1e-7", ".", "1e+", "1e5e5", "--1", "Inf"} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, text string) {
		rat, ok := ParseAmount(text)
		require.Equal(t, ok, rat != nil)
	})
}

func FuzzFormatRoundTrip(f *testing.F) {
	for _, seed := range []string{"0", "-1.5", "+5", "1e21", "1e-7", "0.000001", "123.456e-20", "999999999999999999999"} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, text string) {
		rat, ok := ParseAmount(text)
		if !ok {
			t.Skip()
		}

		formatted, ok := Format(rat)
		require.True(t, ok)
		if len(formatted) > MaxInputLen {
			t.Skip()
		}

		// The canonical exponent can exceed the exponent of the input by the
		// number of digits in the input.
		parsed, ok := Parse(formatted, maxAmountExponent+MaxInputLen)
		require.True(t, ok, formatted)
		require.Zero(t, rat.Cmp(parsed), formatted)
	})
}
