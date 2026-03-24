package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestAccountDelete_TxType(t *testing.T) {
	tx := &AccountDelete{}
	require.Equal(t, AccountDeleteTx, tx.TxType())
}

func TestAccountDelete_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *AccountDelete
		expected FlatTransaction
	}{
		{
			name: "pass - basic",
			tx:   &AccountDelete{},
			expected: FlatTransaction{
				"TransactionType": AccountDeleteTx.String(),
			},
		},
		{
			name: "pass - with destination",
			tx: &AccountDelete{
				BaseTx: BaseTx{
					Account: types.Address("rWYkbWkCeg8dP6rXALnjgZSjjLyih5NXm"),
				},
				Destination:    types.Address("rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe"),
				DestinationTag: 13,
			},
			expected: FlatTransaction{
				"Account":         "rWYkbWkCeg8dP6rXALnjgZSjjLyih5NXm",
				"TransactionType": AccountDeleteTx.String(),
				"Destination":     "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
				"DestinationTag":  uint32(13),
			},
		},
		{
			name: "pass - with credential IDs",
			tx: &AccountDelete{
				BaseTx: BaseTx{
					Account: types.Address("rWYkbWkCeg8dP6rXALnjgZSjjLyih5NXm"),
				},
				Destination:   types.Address("rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe"),
				CredentialIDs: types.CredentialIDs{"1234567890abcdef"},
			},
			expected: FlatTransaction{
				"Account":         "rWYkbWkCeg8dP6rXALnjgZSjjLyih5NXm",
				"TransactionType": AccountDeleteTx.String(),
				"Destination":     "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
				"CredentialIDs":   []string{"1234567890abcdef"},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			require.Equal(t, testcase.expected, testcase.tx.Flatten())
		})
	}
}

func TestAccountDelete_Validate(t *testing.T) {
	testcases := []struct {
		name  string
		tx    *AccountDelete
		valid bool
		err   error
	}{
		{
			name: "fail - invalid base tx",
			tx: &AccountDelete{
				BaseTx: BaseTx{
					TransactionType: AccountDeleteTx,
				},
			},
			valid: false,
			err:   ErrInvalidAccount,
		},
		{
			name: "fail - invalid destination",
			tx: &AccountDelete{
				BaseTx: BaseTx{
					Account:         types.Address("rWYkbWkCeg8dP6rXALnjgZSjjLyih5NXm"),
					TransactionType: AccountDeleteTx,
				},
				Destination: types.Address("invalid"),
			},
			valid: false,
			err:   ErrInvalidDestinationAddress,
		},
		{
			name: "fail - invalid CredentialIDs",
			tx: &AccountDelete{
				BaseTx: BaseTx{
					Account:         types.Address("rWYkbWkCeg8dP6rXALnjgZSjjLyih5NXm"),
					TransactionType: AccountDeleteTx,
				},
				CredentialIDs: types.CredentialIDs{"invalid"},
			},
			valid: false,
			err:   ErrInvalidCredentialIDs,
		},
		{
			name: "pass - basic",
			tx: &AccountDelete{
				BaseTx: BaseTx{
					Account:         types.Address("rWYkbWkCeg8dP6rXALnjgZSjjLyih5NXm"),
					TransactionType: AccountDeleteTx,
				},
				Destination: types.Address("rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe"),
			},
			valid: true,
			err:   nil,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			valid, err := testcase.tx.Validate()
			if testcase.err != nil {
				require.Error(t, err)
				require.Equal(t, err, testcase.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, valid, testcase.valid)
			}
		})
	}
}
