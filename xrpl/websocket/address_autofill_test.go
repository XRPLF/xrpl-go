package websocket

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/stretchr/testify/require"
)

func TestClientSetValidTransactionAddressesRejectsTaggedUnsupportedField(t *testing.T) {
	const taggedOwner = "T719a5UwUCnEs54UsxG9CJYYDhwmFCvqXVCALUGJGSbNV3x"

	tx := transaction.FlatTransaction{"Owner": taggedOwner}
	cl := &Client{}

	err := cl.setValidTransactionAddresses(&tx)

	require.ErrorIs(t, err, ErrAccountIDTagNotAllowed)
	require.Equal(t, transaction.FlatTransaction{"Owner": taggedOwner}, tx)
}
