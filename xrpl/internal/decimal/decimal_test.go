package decimal

import (
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
		ok    bool
	}{
		{name: "integer", value: "42", want: "42", ok: true},
		{name: "signed fraction", value: "-1.5", want: "-3/2", ok: true},
		{name: "explicit plus sign", value: "+5", want: "5", ok: true},
		{name: "leading point", value: ".5", want: "1/2", ok: true},
		{name: "trailing point", value: "5.", want: "5", ok: true},
		{name: "exponent boundary", value: "1E+2", want: "100", ok: true},
		{name: "negative exponent boundary", value: "1e-2", want: "1/100", ok: true},
		{name: "input length boundary", value: strings.Repeat("0", MaxInputLen-1) + "7", want: "7", ok: true},
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			rat, ok := Parse(test.value, 2)
			require.Equal(t, test.ok, ok)
			if ok {
				require.Equal(t, test.want, rat.RatString())
			}
		})
	}
}
