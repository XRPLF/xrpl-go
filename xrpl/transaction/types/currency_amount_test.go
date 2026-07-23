package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssuedCurrencyAmount_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		ica  IssuedCurrencyAmount
		want bool
	}{
		{
			name: "Zero value",
			ica:  IssuedCurrencyAmount{},
			want: true,
		},
		{
			name: "Non-zero value",
			ica: IssuedCurrencyAmount{
				Issuer:   "rEXAMPLE",
				Currency: "USD",
				Value:    "100",
			},
			want: false,
		},
		{
			name: "Non-zero value, invalid only with issuer",
			ica: IssuedCurrencyAmount{
				Issuer: "rEXAMPLE",
			},
			want: false,
		},
		{
			name: "Non-zero value, invalid only with value",
			ica: IssuedCurrencyAmount{
				Value: "100",
			},
			want: false,
		},
		{
			name: "Non-zero value, invalid only with currency",
			ica: IssuedCurrencyAmount{
				Currency: "USD",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ica.IsEmpty(); got != tt.want {
				t.Errorf("IssuedCurrencyAmount.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCurrencyAmount_IsZero(t *testing.T) {
	tests := []struct {
		name   string
		amount CurrencyAmount
		want   bool
	}{
		{name: "XRP - zero", amount: XRPCurrencyAmount(0), want: true},
		{name: "XRP - non-zero", amount: XRPCurrencyAmount(1), want: false},
		{name: "IOU - zero", amount: IssuedCurrencyAmount{Value: "0"}, want: true},
		{name: "IOU - zero with sign and decimals", amount: IssuedCurrencyAmount{Value: "-0.000"}, want: true},
		{name: "IOU - non-zero", amount: IssuedCurrencyAmount{Value: "100"}, want: false},
		// 1e-400 underflows IEEE-754 to 0.0 but is non-zero in the textual decimal.
		{name: "IOU - underflow non-zero", amount: IssuedCurrencyAmount{Value: "1e-400"}, want: false},
		{name: "IOU - in-spec minimum 1e-96", amount: IssuedCurrencyAmount{Value: "1e-96"}, want: false},
		{name: "IOU - invalid value", amount: IssuedCurrencyAmount{Value: "not-a-number"}, want: false},
		{name: "MPT - zero", amount: MPTCurrencyAmount{Value: "0"}, want: true},
		{name: "MPT - non-zero", amount: MPTCurrencyAmount{Value: "42"}, want: false},
		{name: "MPT - invalid value", amount: MPTCurrencyAmount{Value: "abc"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.amount.IsZero())
		})
	}
}

func TestMPTCurrencyAmount_Kind(t *testing.T) {
	testcases := []struct {
		name     string
		mpt      MPTCurrencyAmount
		expected CurrencyKind
	}{
		{
			name:     "pass - mpt kind",
			mpt:      MPTCurrencyAmount{},
			expected: MPT,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.mpt.Kind()
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestMPTCurrencyAmount_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		mpt      MPTCurrencyAmount
		expected map[string]any
		err      error
		expPass  bool
	}{
		{
			name:     "pass - empty",
			mpt:      MPTCurrencyAmount{},
			expected: map[string]any{},
			err:      nil,
			expPass:  true,
		},
		{
			name: "pass - only issuance id",
			mpt: MPTCurrencyAmount{
				MPTIssuanceID: "00000000000000000000000000000000",
			},
			expected: map[string]any{
				"mpt_issuance_id": "00000000000000000000000000000000",
			},
			err:     nil,
			expPass: true,
		},
		{
			name: "pass - only value",
			mpt: MPTCurrencyAmount{
				Value: "100",
			},
			expected: map[string]any{
				"value": "100",
			},
			err:     nil,
			expPass: true,
		},
		{
			name: "pass - both fields",
			mpt: MPTCurrencyAmount{
				MPTIssuanceID: "00000000000000000000000000000000",
				Value:         "100",
			},
			expected: map[string]any{
				"mpt_issuance_id": "00000000000000000000000000000000",
				"value":           "100",
			},
			err:     nil,
			expPass: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.mpt.Flatten()
			require.Equal(t, tc.expected, actual)
			if tc.expPass {
				require.NoError(t, tc.err)
			} else {
				require.Error(t, tc.err)
			}
		})
	}
}

func TestMPTPlainAmount_UnmarshalJSON(t *testing.T) {
	t.Run("pass - valid JSON string", func(t *testing.T) {
		var a MPTPlainAmount
		err := json.Unmarshal([]byte(`"12345"`), &a)
		require.NoError(t, err)
		require.Equal(t, MPTPlainAmount(12345), a)
	})

	t.Run("pass - zero value", func(t *testing.T) {
		var a MPTPlainAmount
		err := json.Unmarshal([]byte(`"0"`), &a)
		require.NoError(t, err)
		require.Equal(t, MPTPlainAmount(0), a)
	})

	t.Run("fail - invalid string", func(t *testing.T) {
		var a MPTPlainAmount
		err := json.Unmarshal([]byte(`"notanumber"`), &a)
		require.Error(t, err)
	})

	t.Run("pass - round trip", func(t *testing.T) {
		original := MPTPlainAmount(9999)
		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded MPTPlainAmount
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)
		require.Equal(t, original, decoded)
	})
}

func TestUnmarshalCurrencyAmount_MPT(t *testing.T) {
	testcases := []struct {
		name     string
		input    []byte
		expected MPTCurrencyAmount
		err      error
		expPass  bool
	}{
		{
			name:  "pass - valid mpt json",
			input: []byte(`{"mpt_issuance_id":"issuance","value":"42"}`),
			expected: MPTCurrencyAmount{
				MPTIssuanceID: "issuance",
				Value:         "42",
			},
			err:     nil,
			expPass: true,
		},
		{
			name:     "fail - invalid json",
			input:    []byte(`{invalid}`),
			expected: MPTCurrencyAmount{},
			err:      nil,
			expPass:  false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := UnmarshalCurrencyAmount(tc.input)
			if tc.expPass {
				require.NoError(t, err)
				mpt, ok := actual.(MPTCurrencyAmount)
				require.True(t, ok, "expected MPTCurrencyAmount, got %T", actual)
				require.Equal(t, tc.expected, mpt)
			} else {
				require.Error(t, err)
			}
		})
	}
}
