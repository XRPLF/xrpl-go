package types

import (
	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
)

type EntryAssetPair struct {
	Asset  ledger.Asset `json:"asset"`
	Asset2 ledger.Asset `json:"asset2"`
}

type EntryMPToken struct {
	MPTIssuanceID string `json:"mp_issuance_id"`
	Account       string `json:"account"`
}

type EntryCredential struct {
	Subject        string `json:"subject"`
	Issuer         string `json:"issuer"`
	CredentialType string `json:"credential_type"`
}

type EntryDepositPreauth struct {
	Owner      string `json:"owner"`
	Authorized string `json:"authorized"`
}

type EntryDirectory struct {
	SubIndex uint32 `json:"sub_index,omitempty"`
	DirRoot  string `json:"dir_root,omitempty"`
	Owner    string `json:"owner,omitempty"`
}

type EntryEscrow struct {
	Owner string `json:"owner"`
	Seq   string `json:"seq"`
}

type EntryOffer struct {
	Account string `json:"account"`
	Seq     string `json:"seq"`
}

type EntryRippleState struct {
	Accounts []string `json:"accounts"`
	Currency string   `json:"currency"`
}

type EntryTicket struct {
	Owner          string `json:"owner"`
	TicketSequence string `json:"ticket_sequence"`
}

type EntryXChainBridge struct {
	LockingChainDoor  string `json:"locking_chain_door"`
	LockingChainIssue string `json:"locking_chain_issue"`
	IssuingChainDoor  string `json:"issuing_chain_door"`
	IssuingChainIssue string `json:"issuing_chain_issue"`
}

type EntryXChainOwnedClaimID struct {
	EntryXChainBridge
	XChainOwnedClaimID interface{} `json:"xchain_owned_claim_id"`
}

type EntryXChainOwnedCreateAccountClaimID struct {
	EntryXChainBridge
	XChainOwnedCreateAccountClaimID interface{} `json:"xchain_owned_create_account_claim_id"`
}
