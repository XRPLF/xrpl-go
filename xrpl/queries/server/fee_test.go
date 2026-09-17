package server_test

import (
	"encoding/json"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
	servertypes "github.com/Peersyst/xrpl-go/xrpl/queries/server/types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/require"
)

func feeResponseFixture() (server.FeeResponse, string) {
	s := server.FeeResponse{
		CurrentLedgerSize: "14",
		CurrentQueueSize:  "0",
		Drops: servertypes.FeeDrops{
			BaseFee:       10,
			MedianFee:     11000,
			MinimumFee:    10,
			OpenLedgerFee: 10,
		},
		ExpectedLedgerSize: "24",
		LedgerCurrentIndex: 26575101,
		Levels: servertypes.FeeLevels{
			MedianLevel:     281600,
			MinimumLevel:    256,
			OpenLedgerLevel: 256,
			ReferenceLevel:  256,
		},
		MaxQueueSize: "480",
	}
	j := `{
	"current_ledger_size": "14",
	"current_queue_size": "0",
	"drops": {
		"base_fee": "10",
		"median_fee": "11000",
		"minimum_fee": "10",
		"open_ledger_fee": "10"
	},
	"expected_ledger_size": "24",
	"ledger_current_index": 26575101,
	"levels": {
		"median_level": "281600",
		"minimum_level": "256",
		"open_ledger_level": "256",
		"reference_level": "256"
	},
	"max_queue_size": "480"
}`
	return s, j
}

func TestFeeResponseSerialize(t *testing.T) {
	value, payload := feeResponseFixture()
	require.NoError(t, testutil.Serialize(t, value, payload))
}

func TestFeeResponseJSONDecode(t *testing.T) {
	want, payload := feeResponseFixture()
	var got server.FeeResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestFeeResponseClientDecode(t *testing.T) {
	want, payload := feeResponseFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got server.FeeResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}
