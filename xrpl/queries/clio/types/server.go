// Package types contains data structures for CLIO server query types.
//
//revive:disable:var-naming
package types

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

// Counters holds metrics for RPC calls and active subscriptions in the CLIO server.
// It includes counts of forwarded RPCs and the number of active subscription types.
type Counters struct {
	RPC           map[string]RPC `json:"rpc"`
	Subscriptions Subscriptions  `json:"subscriptions"`
}

// RPC represents timing and status information for a single RPC call.
// It tracks start, finish, error, and forwarded timestamps and duration.
type RPC struct {
	Started    string `json:"started,omitempty"`
	Finished   string `json:"finished,omitempty"`
	Errored    string `json:"errored,omitempty"`
	Forwarded  string `json:"forwarded,omitempty"`
	DurationUS string `json:"duration_us,omitempty"`
}

// Subscriptions holds counters for different CLIO subscription streams.
// Each field represents the number of active subscribers for that stream.
type Subscriptions struct {
	Ledger               int `json:"ledger"`
	Transactions         int `json:"transactions"`
	TransactionsProposed int `json:"transactions_proposed"`
	Manifests            int `json:"manifests"`
	Validations          int `json:"validations"`
	Account              int `json:"account"`
	AccountsProposed     int `json:"accounts_proposed"`
	Books                int `json:"books"`
}

// LedgerInfo contains information about the current ledger state in CLIO.
// It includes sequence number, hash, fee metrics, and ledger age.
type LedgerInfo struct {
	Age            uint          `json:"age"`
	BaseFeeXRP     float64       `json:"base_fee_xrp"`
	Hash           types.Hash256 `json:"hash"`
	ReserveBaseXRP float32       `json:"reserve_base_xrp"`
	ReserveIncXRP  float32       `json:"reserve_inc_xrp"`
	Seq            uint          `json:"seq"`
}

// Cache describes the state of the CLIO cache, including size and fullness.
type Cache struct {
	Size            int  `json:"size"`
	IsFull          bool `json:"is_full"`
	LatestLedgerSeq int  `json:"latest_ledger_seq"`
}

// ETL holds status details for the CLIO ETL (Extract, Transform, Load) subsystem.
// It reports on source connections, writer status, and publication age.
type ETL struct {
	ETLSources            []ETLSource `json:"etl_sources"`
	IsWriter              bool        `json:"is_writer"`
	ReadOnly              bool        `json:"read_only"`
	LastPublishAgeSeconds string      `json:"last_publish_age_seconds"`
}

// ETLSource represents a single data source configured for CLIO ETL.
// It includes connection status, ports, and message age metrics.
type ETLSource struct {
	ValidatedRange    string `json:"validated_range"`
	IsConnected       string `json:"is_connected"`
	IP                string `json:"ip"`
	WSPort            string `json:"ws_port"`
	GRPCPort          string `json:"grpc_port"`
	LastMsgAgeSeconds string `json:"last_msg_age_seconds"`
}
