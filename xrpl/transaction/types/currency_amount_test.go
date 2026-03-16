package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssuedCurrencyAmount_IsZero(t *testing.T) {
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
			if got := tt.ica.IsZero(); got != tt.want {
				t.Errorf("IssuedCurrencyAmount.IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMPTCurrencyAmount_Kind(t *testing.T) {
	testcases := []struct {
		name     string
		mpt      MPTCurrencyAmount
		expected CurrencyKind
		expPass  bool
	}{
		{
			name:     "pass - mpt kind",
			mpt:      MPTCurrencyAmount{},
			expected: MPT,
			expPass:  true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.mpt.Kind()
			require.Equal(t, tc.expected, actual)
			if tc.expPass {
				require.NoError(t, nil)
			} else {
				require.Error(t, nil)
			}
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
