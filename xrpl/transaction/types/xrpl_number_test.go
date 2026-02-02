package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestXRPLNumber_String(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "empty string",
			value: "",
			want:  "",
		},
		{
			name:  "integer value",
			value: "100000",
			want:  "100000",
		},
		{
			name:  "decimal value",
			value: "1000.50",
			want:  "1000.50",
		},
		{
			name:  "small decimal",
			value: "0.000001",
			want:  "0.000001",
		},
		{
			name:  "scientific notation",
			value: "1e6",
			want:  "1e6",
		},
		{
			name:  "scientific notation with decimal",
			value: "1.5e10",
			want:  "1.5e10",
		},
		{
			name:  "negative number",
			value: "-1000",
			want:  "-1000",
		},
		{
			name:  "zero",
			value: "0",
			want:  "0",
		},
		{
			name:  "large number",
			value: "999999999999999999",
			want:  "999999999999999999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := XRPLNumber(tt.value)
			require.Equal(t, tt.want, (&result).String())
		})
	}
}
