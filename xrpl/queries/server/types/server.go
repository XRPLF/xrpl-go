// Package types provides data structures for server query responses.
// revive:disable:var-naming
package types

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

// Info represents the server info response, including load, ledger, and network metrics.
type Info struct {
	AmendmentBlocked         bool                 `json:"amendment_blocked,omitempty"`
	BuildVersion             string               `json:"build_version"`
	CompleteLedgers          string               `json:"complete_ledgers"`
	ClosedLedger             ClosedLedger         `json:"closed_ledger,omitzero"`
	HostID                   string               `json:"hostid"`
	IOLatencyMS              uint                 `json:"io_latency_ms"`
	JQTransOverflow          string               `json:"jq_trans_overflow"`
	LastClose                ServerClose          `json:"last_close"`
	Load                     ServerLoad           `json:"load,omitzero"`
	LoadFactor               uint                 `json:"load_factor"`
	NetworkID                uint                 `json:"network_id,omitempty"`
	LoadFactorLocal          uint                 `json:"load_factor_local,omitempty"`
	LoadFactorNet            uint                 `json:"load_factor_net,omitempty"`
	LoadFactorCluster        uint                 `json:"load_factor_cluster,omitempty"`
	LoadFactorFeeEscelation  uint                 `json:"load_factor_fee_escelation,omitempty"`
	LoadFactorFeeQueue       uint                 `json:"load_factor_fee_queue,omitempty"`
	LoadFactorServer         uint                 `json:"load_factor_server,omitempty"`
	PeerDisconnects          string               `json:"peer_disconnects,omitempty"`
	PeerDisconnectsResources string               `json:"peer_disconnects_resources,omitempty"`
	NetworkLedger            string               `json:"network_ledger,omitempty"`
	Peers                    uint                 `json:"peers,omitempty"`
	Ports                    []ServerPort         `json:"ports,omitempty"`
	PubkeyNode               string               `json:"pubkey_node"`
	PubkeyValidator          string               `json:"pubkey_validator,omitempty"`
	ServerState              string               `json:"server_state"`
	ServerStateDurationUS    string               `json:"server_state_duration_us"`
	StateAccounting          StateAccountingFinal `json:"state_accounting"`
	Time                     string               `json:"time"`
	Uptime                   uint                 `json:"uptime"`
	ValidatedLedger          ClosedLedger         `json:"validated_ledger,omitzero"`
	ValidationQuorum         uint                 `json:"validation_quorum"`
	ValidatorListExpires     string               `json:"validator_list_expires,omitempty"`
	ValidatorList            ServerValidatorList  `json:"validator_list,omitzero"`
}

// ServerValidatorList holds the count, expiration, and status of the server's validator list.
type ServerValidatorList struct {
	Count      uint   `json:"count"`
	Expiration string `json:"expiration"`
	Status     string `json:"status"`
}

// ServerLoad contains metrics about current server job types and thread usage.
type ServerLoad struct {
	JobTypes []JobType `json:"job_types"`
	Threads  uint      `json:"threads"`
}

// ServerClose holds details about the last ledger close, including converge time and number of proposers.
type ServerClose struct {
	ConvergeTimeS float32 `json:"converge_time_s"`
	Proposers     uint    `json:"proposers"`
}

// State represents a summary of the server's operational state, including load and ledger statistics.
type State struct {
	AmendmentBlocked        bool                 `json:"amendment_blocked,omitempty"`
	BuildVersion            string               `json:"build_version"`
	CompleteLedgers         string               `json:"complete_ledgers"`
	ClosedLedger            ClosedLedgerState    `json:"closed_ledger,omitzero"`
	IOLatencyMS             uint                 `json:"io_latency_ms"`
	JQTransOverflow         string               `json:"jq_trans_overflow"`
	LastClose               CloseState           `json:"last_close"`
	Load                    ServerLoad           `json:"load,omitzero"`
	LoadBase                int                  `json:"load_base"`
	LoadFactor              uint                 `json:"load_factor"`
	LoadFactorFeeEscelation uint                 `json:"load_factor_fee_escalation,omitempty"`
	LoadFactorFeeQueue      uint                 `json:"load_factor_fee_queue,omitempty"`
	LoadFactorFeeReference  uint                 `json:"load_factor_fee_reference,omitempty"`
	LoadFactorServer        uint                 `json:"load_factor_server,omitempty"`
	Peers                   uint                 `json:"peers,omitempty"`
	PubkeyNode              string               `json:"pubkey_node"`
	PubkeyValidator         string               `json:"pubkey_validator,omitempty"`
	ServerState             string               `json:"server_state"`
	ServerStateDurationUS   string               `json:"server_state_duration_us"`
	StateAccounting         StateAccountingFinal `json:"state_accounting"`
	Time                    string               `json:"time"`
	Uptime                  uint                 `json:"uptime"`
	ValidatedLedger         LedgerState          `json:"validated_ledger,omitzero"`
	ValidationQuorum        uint                 `json:"validation_quorum"`
	ValidatorListExpires    string               `json:"validator_list_expires,omitempty"`
}

// ClosedLedgerState contains metadata for a closed ledger, such as age, fees, and sequence.
type ClosedLedgerState struct {
	Age         uint          `json:"age"`
	BaseFee     float32       `json:"base_fee"`
	Hash        types.Hash256 `json:"hash"`
	ReserveBase float32       `json:"reserve_base"`
	ReserveInc  float32       `json:"reserve_inc"`
	Seq         uint          `json:"seq"`
}

// LedgerState represents the state of a validated ledger in the server state response.
type LedgerState struct {
	Age         uint   `json:"age,omitempty"`
	BaseFee     uint   `json:"base_fee"`
	CloseTime   uint   `json:"close_time"`
	Hash        string `json:"hash"`
	ReserveBase uint   `json:"reserve_base"`
	ReserveInc  uint   `json:"reserve_inc"`
	Seq         uint   `json:"seq"`
}

// CloseState describes metrics of a ledger close, including converge time and proposer count.
type CloseState struct {
	ConvergeTime uint `json:"converge_time"`
	Proposers    uint `json:"proposers"`
}
