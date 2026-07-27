package integration

import (
	"testing"

	querycommon "github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/stretchr/testify/require"
)

type ledgerAcceptRequest struct {
	querycommon.BaseRequest
}

func (*ledgerAcceptRequest) Method() string {
	return "ledger_accept"
}

func (*ledgerAcceptRequest) Validate() error {
	return nil
}

func acceptLedger(t *testing.T, client *rpc.Client) {
	t.Helper()

	_, err := client.Request(&ledgerAcceptRequest{})
	require.NoError(t, err)
}
