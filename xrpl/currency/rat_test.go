package currency

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecimalRatRejectsExpensiveInputs(t *testing.T) {
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
			name:  "valid input length boundary",
			value: "100000000000000000.000000",
			ok:    true,
		},
		{
			name:  "input too long",
			value: "10000000000000000000000000000000000000000000000000000000000000000",
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

			actual, ok := decimalRat(test.value)
			require.Equal(t, test.ok, ok)
			require.Equal(t, test.ok, actual != nil)
		})
	}
}
