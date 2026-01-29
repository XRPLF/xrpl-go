package transaction

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOracleDelete_TxType(t *testing.T) {
	tx := &OracleDelete{}
	assert.Equal(t, OracleDeleteTx, tx.TxType())
}

func TestOracleDelete_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *OracleDelete
		expected FlatTransaction
	}{
		{
			name: "pass - empty",
			tx:   &OracleDelete{},
			expected: FlatTransaction{
				"TransactionType":  "OracleDelete",
				"OracleDocumentID": uint32(0),
			},
		},
		{
			name: "pass - complete",
			tx: &OracleDelete{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: OracleDeleteTx,
				},
				OracleDocumentID: 34,
			},
			expected: FlatTransaction{
				"Account":          "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
				"TransactionType":  "OracleDelete",
				"OracleDocumentID": uint32(34),
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			actual := testcase.tx.Flatten()
			assert.Equal(t, testcase.expected, actual)
		})
	}
}

func TestOracleDelete_Validate(t *testing.T) {
	testcases := []struct {
		name string
		tx   *OracleDelete
		err  error
	}{
		{
			name: "fail - missing account",
			tx: &OracleDelete{
				BaseTx: BaseTx{
					TransactionType: OracleDeleteTx,
				},
			},
			err: ErrInvalidAccount,
		},
		{
			name: "pass - complete",
			tx: &OracleDelete{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: OracleDeleteTx,
				},
			},
			err: nil,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			ok, err := testcase.tx.Validate()
			assert.Equal(t, testcase.err, err)
			assert.Equal(t, ok, testcase.err == nil)
		})
	}
}
