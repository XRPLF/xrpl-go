package types

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMPTAmount(t *testing.T) {
	amount := MPTAmount(10000)

	require.Equal(t, uint64(10000), amount.Uint64())
	require.Equal(t, "10000", amount.String())
	require.Equal(t, "10000", amount.Flatten())
	require.False(t, amount.IsZero())
	require.True(t, amount.IsValid())
	require.True(t, MPTAmount(0).IsZero())
	require.False(t, MPTAmount(math.MaxInt64+1).IsValid())
}

func TestMPTAmountJSON(t *testing.T) {
	t.Run("marshal maximum", func(t *testing.T) {
		encoded, err := json.Marshal(MaxMPTAmount)
		require.NoError(t, err)
		require.JSONEq(t, `"9223372036854775807"`, string(encoded))
	})

	t.Run("marshal out of range", func(t *testing.T) {
		_, err := json.Marshal(MPTAmount(math.MaxInt64 + 1))
		require.ErrorIs(t, err, ErrInvalidMPTAmount)
	})

	for _, value := range []string{"0", "1", "10000", "9223372036854775807"} {
		t.Run("unmarshal "+value, func(t *testing.T) {
			var amount MPTAmount
			err := json.Unmarshal([]byte(`"`+value+`"`), &amount)
			require.NoError(t, err)
			require.Equal(t, value, amount.String())
		})
	}

	for _, input := range []string{
		`1`,
		`""`,
		`"-1"`,
		`"+1"`,
		`"1.0"`,
		`"0x10"`,
		`"9223372036854775808"`,
	} {
		t.Run("reject "+input, func(t *testing.T) {
			var amount MPTAmount
			err := json.Unmarshal([]byte(input), &amount)
			require.ErrorIs(t, err, ErrInvalidMPTAmount)
		})
	}
}
