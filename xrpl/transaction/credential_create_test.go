package transaction

import (
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestCredentialCreate_TxType(t *testing.T) {
	tx := &CredentialCreate{}
	require.Equal(t, CredentialCreateTx, tx.TxType())
}

func TestCredentialCreate_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		input    *CredentialCreate
		expected FlatTransaction
	}{
		{
			name: "pass - valid CredentialCreate",
			input: &CredentialCreate{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: CredentialCreateTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
				},
				Subject:        "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				Expiration:     123456,
				CredentialType: "6D795F63726564656E7469616C",                                   // "my_credential" in hex
				URI:            "687474703A2F2F636F6D70616E792E636F6D2F63726564656E7469616C73", // "http://company.com/credentials" in hex
			},
			expected: FlatTransaction{
				"Account":         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
				"TransactionType": "CredentialCreate",
				"Fee":             "1",
				"Sequence":        uint32(1234),
				"Subject":         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Expiration":      uint32(123456),
				"CredentialType":  "6D795F63726564656E7469616C",
				"URI":             "687474703A2F2F636F6D70616E792E636F6D2F63726564656E7469616C73",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flattened := tt.input.Flatten()
			require.Equal(t, tt.expected, flattened)
		})
	}
}

func TestCredentialCreate_Validate(t *testing.T) {
	tests := []struct {
		name     string
		input    *CredentialCreate
		expected bool
	}{
		{
			name: "pass - valid CredentialCreate",
			input: &CredentialCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Subject:        "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				CredentialType: "6D795F63726564656E7469616C",
				Expiration:     123456,
				URI:            "687474703A2F2F636F6D70616E792E636F6D2F63726564656E7469616C73",
			},
			expected: true,
		},
		{
			name: "pass - valid CredentialCreate with required fields only",
			input: &CredentialCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Subject:        "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				CredentialType: "6D795F63726564656E7469616C",
			},
			expected: true,
		},
		{
			name: "pass - valid CredentialCreate without URI",
			input: &CredentialCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Subject:        "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				CredentialType: types.CredentialType("6D795F63726564656E7469616C"),
				Expiration:     123456,
			},
			expected: true,
		},
		{
			name: "pass - valid CredentialCreate without URI",
			input: &CredentialCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Subject:        "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				CredentialType: types.CredentialType("6D795F63726564656E7469616C"),
				Expiration:     123456,
			},
			expected: true,
		},
		{
			name: "fail - CredentialCreate with an invalid Subject",
			input: &CredentialCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Subject:        "invalid_address",
				CredentialType: types.CredentialType("6D795F63726564656E7469616C"),
				Expiration:     123456,
			},
			expected: false,
		},
		{
			name: "pass - CredentialCreate with an Expiration of 0 (won't be pass to the flatten transaction)",
			input: &CredentialCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Subject:        "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				CredentialType: types.CredentialType("6D795F63726564656E7469616C"),
				Expiration:     0,
			},
			expected: true,
		},
		{
			name: "fail - CredentialCreate with an invalid CredentialType (empty)",
			input: &CredentialCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Subject:        "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				CredentialType: types.CredentialType(""),
			},
			expected: false,
		},
		{
			name: "fail - CredentialCreate with an invalid CredentialType (not hex)",
			input: &CredentialCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Subject:        "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				CredentialType: types.CredentialType("not hexadecimal value"),
			},
			expected: false,
		},
		{
			name: "fail - CredentialCreate with an invalid CredentialType (too long)",
			input: &CredentialCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Subject:        "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				CredentialType: types.CredentialType(strings.Repeat("0", types.MaxCredentialTypeLength+1)),
			},
			expected: false,
		},
		{
			name: "fail - CredentialCreate with an invalid CredentialType (too short)",
			input: &CredentialCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Subject:        "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				CredentialType: types.CredentialType(strings.Repeat("0", types.MinCredentialTypeLength-1)),
			},
			expected: false,
		},
		{
			name: "fail - CredentialCreate with an invalid BaseTx",
			input: &CredentialCreate{
				BaseTx: BaseTx{
					Account:         "invalid",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Subject:        "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				CredentialType: types.CredentialType("6D795F63726564656E7469616C"),
			},
			expected: false,
		},
		{
			name: "fail - CredentialCreate with a URI that is too short",
			input: &CredentialCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Subject:        "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				CredentialType: "6D795F63726564656E7469616C",
				Expiration:     123456,
				URI:            "0",
			},
			expected: false,
		},
		{
			name: "fail - CredentialCreate with a URI that is too long",
			input: &CredentialCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Subject:        "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				CredentialType: "6D795F63726564656E7469616C",
				Expiration:     123456,
				URI:            strings.Repeat("0", 513),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.input.Validate()
			require.Equal(t, tt.expected, valid)
			if tt.expected {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
