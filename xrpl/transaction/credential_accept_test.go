package transaction

import (
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestCredentialAccept_TxType(t *testing.T) {
	tx := &CredentialAccept{}
	require.Equal(t, CredentialAcceptTx, tx.TxType())
}

func TestCredentialAccept_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		input    *CredentialAccept
		expected FlatTransaction
	}{
		{
			name: "pass - valid CredentialAccept",
			input: &CredentialAccept{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: CredentialAcceptTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
				},
				Issuer:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				CredentialType: "6D795F63726564656E7469616C", // "my_credential" in hex
			},
			expected: FlatTransaction{
				"Account":         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
				"TransactionType": "CredentialAccept",
				"Fee":             "1",
				"Sequence":        uint32(1234),
				"Issuer":          "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"CredentialType":  "6D795F63726564656E7469616C",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flattened := tt.input.Flatten()
			require.Equal(t, tt.expected, flattened, "Flatten result differs from expected")
		})
	}
}

func TestCredentialAccept_Validate(t *testing.T) {
	tests := []struct {
		name     string
		input    *CredentialAccept
		expected bool
	}{
		{
			name: "pass - valid CredentialAccept",
			input: &CredentialAccept{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Issuer:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				CredentialType: "6D795F63726564656E7469616C",
			},
			expected: true,
		},
		{
			name: "fail - CredentialAccept with an invalid Issuer",
			input: &CredentialAccept{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Issuer:         "invalid_address",
				CredentialType: types.CredentialType("6D795F63726564656E7469616C"),
			},
			expected: false,
		},
		{
			name: "fail - CredentialAccept with an invalid CredentialType (empty)",
			input: &CredentialAccept{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Issuer:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				CredentialType: types.CredentialType(""),
			},
			expected: false,
		},
		{
			name: "fail - CredentialAccept with an invalid CredentialType (not hex)",
			input: &CredentialAccept{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Issuer:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				CredentialType: types.CredentialType("not hexadecimal value"),
			},
			expected: false,
		},
		{
			name: "fail - CredentialCreate with an invalid CredentialType (too long)",
			input: &CredentialAccept{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Issuer:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				CredentialType: types.CredentialType(strings.Repeat("0", types.MaxCredentialTypeLength+1)),
			},
			expected: false,
		},
		{
			name: "fail - CredentialCreate with an invalid CredentialType (too short)",
			input: &CredentialAccept{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Issuer:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				CredentialType: types.CredentialType(strings.Repeat("0", types.MinCredentialTypeLength-1)),
			},
			expected: false,
		},
		{
			name: "fail - CredentialAccept with an invalid BaseTx",
			input: &CredentialAccept{
				BaseTx: BaseTx{
					Account:         "invalid",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Issuer:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				CredentialType: types.CredentialType("6D795F63726564656E7469616C"),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.input.Validate()
			require.Equal(t, tt.expected, valid, "Validation result mismatch")
			if tt.expected {
				require.NoError(t, err, "Expected no error for valid case")
			} else {
				require.Error(t, err, "Expected error for invalid case")
			}
		})
	}
}
