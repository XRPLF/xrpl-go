package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBatchAccount_String(t *testing.T) {
	value := "r9cZA1mLK5R5Am25ArfXFmqgNQW4RdFp"
	empty := BatchAccount{}
	ba := BatchAccount{
		value: value,
	}

	require.Empty(t, empty.String())
	require.Equal(t, ba.String(), value)
}
