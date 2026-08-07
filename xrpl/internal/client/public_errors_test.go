package client_test

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

func TestPublicClientErrorIdentity(t *testing.T) {
	require.ErrorIs(t, rpc.ErrAmountAndDeliverMaxMustBeIdentical, websocket.ErrAmountAndDeliverMaxMustBeIdentical)
	require.ErrorIs(t, rpc.ErrTransactionNotMultisigned, websocket.ErrTransactionNotMultisigned)
	require.ErrorIs(t, rpc.ErrBatchRawTransactionsCount, websocket.ErrBatchRawTransactionsCount)
	require.ErrorIs(t, rpc.ErrSignerDataIsEmpty, rpc.ErrTransactionNotMultisigned)
	require.ErrorIs(t, websocket.ErrSignerDataIsEmpty, websocket.ErrTransactionNotMultisigned)
}
