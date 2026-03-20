package types

import "testing"

func TestIssuedCurrency_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		currency IssuedCurrency
		want     map[string]any
	}{
		{
			name: "pass - basic issued currency",
			currency: IssuedCurrency{
				Currency: "FOO",
				Issuer:   "rPdYxU9dNkbzC5Y2h4jLbVJ3rMRrk7WVRL",
			},
			want: map[string]any{
				"currency": "FOO",
				"issuer":   "rPdYxU9dNkbzC5Y2h4jLbVJ3rMRrk7WVRL",
			},
		},
		{
			name: "pass - empty currency",
			currency: IssuedCurrency{
				Currency: "",
				Issuer:   "rPdYxU9dNkbzC5Y2h4jLbVJ3rMRrk7WVRL",
			},
			want: map[string]any{
				"currency": "",
				"issuer":   "rPdYxU9dNkbzC5Y2h4jLbVJ3rMRrk7WVRL",
			},
		},
		{
			name: "pass - empty issuer",
			currency: IssuedCurrency{
				Currency: "BAR",
				Issuer:   "",
			},
			want: map[string]any{
				"currency": "BAR",
				"issuer":   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.currency.Flatten()
			if len(got) != len(tt.want) {
				t.Errorf("IssuedCurrency.Flatten() map length = %v, want %v", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("IssuedCurrency.Flatten() got[%v] = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}
