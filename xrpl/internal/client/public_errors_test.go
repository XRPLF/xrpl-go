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
	require.ErrorIs(t, rpc.ErrInvalidLastLedgerSequence, websocket.ErrInvalidLastLedgerSequence)
	require.ErrorIs(t, rpc.ErrSignerDataIsEmpty, rpc.ErrTransactionNotMultisigned)
	require.ErrorIs(t, websocket.ErrSignerDataIsEmpty, websocket.ErrTransactionNotMultisigned)
	require.ErrorIs(t, rpc.ErrAccountCannotBeDeleted, websocket.ErrAccountCannotBeDeleted)
	require.ErrorIs(t, rpc.ErrAccountHasSponsorshipObligations, websocket.ErrAccountHasSponsorshipObligations)
	require.ErrorIs(t, rpc.ErrAccountDeleteSponsorMismatch, websocket.ErrAccountDeleteSponsorMismatch)
}

func TestAccountDeleteErrorCompatibility(t *testing.T) {
	tests := []struct {
		name        string
		objects     error
		obligations error
		mismatch    error
	}{
		{"rpc", rpc.ErrAccountCannotBeDeleted, rpc.ErrAccountHasSponsorshipObligations, rpc.ErrAccountDeleteSponsorMismatch},
		{"websocket", websocket.ErrAccountCannotBeDeleted, websocket.ErrAccountHasSponsorshipObligations, websocket.ErrAccountDeleteSponsorMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.EqualError(t, tt.objects, "account cannot be deleted; there are Escrows, PayChannels, RippleStates, or Checks associated with the account")
			require.NotErrorIs(t, tt.obligations, tt.objects)
			require.NotErrorIs(t, tt.mismatch, tt.objects)
		})
	}
}
