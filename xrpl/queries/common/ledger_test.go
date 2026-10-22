package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmarshalLedgerSpecifier_InvalidTitle(t *testing.T) {
	got, err := UnmarshalLedgerSpecifier([]byte(`"foo"`))
	require.ErrorIs(t, err, ErrInvalidLedgerSpecifier)
	require.EqualError(t, err, `invalid ledger specifier: "foo"`)
	require.Nil(t, got)
}
