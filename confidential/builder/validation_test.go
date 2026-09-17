package builder

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// stubTransaction stands in for a prepared transaction. No Prepare* input reaches the
// rejection path today, because every field the confidential MPT transactions validate is
// already screened by the builder's own validation. validatePreparedTransaction is the
// backstop for a transaction that later gains a field the builders do not screen, so it is
// exercised here against the interface it consumes.
type stubTransaction struct {
	valid bool
	err   error
}

func (s stubTransaction) Validate() (bool, error) { return s.valid, s.err }

func TestValidatePreparedTransaction(t *testing.T) {
	cause := errors.New("field is malformed")

	tests := []struct {
		name    string
		tx      stubTransaction
		wantErr bool
	}{
		{name: "accepted", tx: stubTransaction{valid: true}},
		{name: "rejected", tx: stubTransaction{err: cause}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validatePreparedTransaction(test.tx)
			if !test.wantErr {
				require.NoError(t, err)
				return
			}
			// The caller needs the sentinel to classify the failure and the cause to know
			// which field the transaction rejected.
			require.ErrorIs(t, err, ErrInvalidTransaction)
			require.ErrorIs(t, err, cause)
		})
	}
}

// TestValidatePreparedTransactionRejectsBoolOnlyFailure covers the half of the Validate
// contract no transaction in the repository exercises today. It has its own test because
// there is no cause to unwrap, which is the assertion the table above is built around.
func TestValidatePreparedTransactionRejectsBoolOnlyFailure(t *testing.T) {
	err := validatePreparedTransaction(stubTransaction{valid: false})
	require.ErrorIs(t, err, ErrInvalidTransaction)
}
