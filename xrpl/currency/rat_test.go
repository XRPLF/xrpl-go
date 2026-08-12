package currency

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNativeAmountRatRejectsExpensiveInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		ok    bool
	}{
		{
			name:  "valid scientific notation",
			value: "1e-6",
			ok:    true,
		},
		{
			name:  "valid exponent boundary",
			value: "1e17",
			ok:    true,
		},
		{
			name:  "valid non-canonical native amount",
			value: "100000000000000000.000000",
			ok:    true,
		},
		{
			name:  "valid input length boundary",
			value: strings.Repeat("1", maxDecimalRatInputLen),
			ok:    true,
		},
		{
			name:  "input too long",
			value: strings.Repeat("1", maxDecimalRatInputLen+1),
			ok:    false,
		},
		{
			name:  "positive exponent too large",
			value: "1e18",
			ok:    false,
		},
		{
			name:  "negative exponent too large",
			value: "1e-18",
			ok:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			rat, ok := nativeAmountRat(test.value)
			require.Equal(t, test.ok, ok)
			require.Equal(t, test.ok, rat != nil)
		})
	}
}
