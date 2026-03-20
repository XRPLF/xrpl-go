package server

import (
	"testing"

	servertypes "github.com/Peersyst/xrpl-go/xrpl/queries/server/types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

func TestServerInfoResponse(t *testing.T) {
	s := InfoResponse{
		Info: servertypes.Info{
			BuildVersion:    "1.9.4",
			CompleteLedgers: "32570-75801736",
			HostID:          "ARMY",
			IOLatencyMS:     1,
			JQTransOverflow: "2282",
			LastClose: servertypes.ServerClose{
				ConvergeTimeS: 3.002,
				Proposers:     35,
			},
			LoadFactor:            1,
			Peers:                 20,
			PubkeyNode:            "n9KKBZvwPZ95rQi4BP3an1MRctTyavYkZiLpQwasmFYTE6RYdeX3",
			ServerState:           "full",
			ServerStateDurationUS: "69205850392",
			StateAccounting: servertypes.StateAccountingFinal{
				Connected: servertypes.InfoAccounting{
					DurationUS:  "141058919",
					Transitions: "7",
				},
				Disconnected: servertypes.InfoAccounting{
					DurationUS:  "514136273",
					Transitions: "3",
				},
				Full: servertypes.InfoAccounting{
					DurationUS:  "4360230140761",
					Transitions: "32",
				},
				Syncing: servertypes.InfoAccounting{
					DurationUS:  "50606510",
					Transitions: "30",
				},
				Tracking: servertypes.InfoAccounting{
					DurationUS:  "40245486",
					Transitions: "34",
				},
			},
			Time:   "2022-Nov-16 21:50:22.711679 UTC",
			Uptime: 4360976,
			ValidatedLedger: servertypes.ClosedLedger{
				Age:            1,
				BaseFeeXRP:     0.00001,
				Hash:           "3147A41F5F013209581FCDCBBB7A87A4F01EF6842963E13B2B14C8565E00A22B",
				ReserveBaseXRP: 10,
				ReserveIncXRP:  2,
				Seq:            75801736,
			},
			ValidationQuorum: 28,
		},
	}

	j := `{
	"info": {
		"build_version": "1.9.4",
		"complete_ledgers": "32570-75801736",
		"hostid": "ARMY",
		"io_latency_ms": 1,
		"jq_trans_overflow": "2282",
		"last_close": {
			"converge_time_s": 3.002,
			"proposers": 35
		},
		"load_factor": 1,
		"peers": 20,
		"pubkey_node": "n9KKBZvwPZ95rQi4BP3an1MRctTyavYkZiLpQwasmFYTE6RYdeX3",
		"server_state": "full",
		"server_state_duration_us": "69205850392",
		"state_accounting": {
			"disconnected": {
				"duration_us": "514136273",
				"transitions": "3"
			},
			"connected": {
				"duration_us": "141058919",
				"transitions": "7"
			},
			"full": {
				"duration_us": "4360230140761",
				"transitions": "32"
			},
			"syncing": {
				"duration_us": "50606510",
				"transitions": "30"
			},
			"tracking": {
				"duration_us": "40245486",
				"transitions": "34"
			}
		},
		"time": "2022-Nov-16 21:50:22.711679 UTC",
		"uptime": 4360976,
		"validated_ledger": {
			"age": 1,
			"base_fee_xrp": 0.00001,
			"hash": "3147A41F5F013209581FCDCBBB7A87A4F01EF6842963E13B2B14C8565E00A22B",
			"reserve_base_xrp": 10,
			"reserve_inc_xrp": 2,
			"seq": 75801736
		},
		"validation_quorum": 28
	}
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}
