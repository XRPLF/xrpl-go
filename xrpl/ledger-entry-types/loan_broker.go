package ledger

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// LoanBroker represents a LoanBroker ledger entry that captures attributes of the Lending Protocol.
//
// ```json
//
//	{
//	  "LedgerEntryType": "LoanBroker",
//	  "Flags": 0,
//	  "Sequence": 3606,
//	  "LoanSequence": 1,
//	  "OwnerNode": "0000000000000000",
//	  "VaultNode": "0000000000000000",
//	  "VaultID": "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
//	  "Account": "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
//	  "Owner": "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
//	  "DebtMaximum": "1000000",
//	  "PreviousTxnID": "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
//	  "PreviousTxnLgrSeq": 28991004
//	}
//
// ```
type LoanBroker struct {
	// The unique ID for this ledger entry. In JSON, this field is represented with different names depending on the
	// context and API method. (Note, even though this is specified as "optional" in the code, every ledger entry
	// should have one unless it's legacy data from very early in the XRP Ledger's history.)
	Index types.Hash256 `json:"index,omitempty"`
	// The value "LoanBroker", mapped to the string LoanBroker, indicates that this object is a LoanBroker object.
	LedgerEntryType EntryType
	// Ledger object flags.
	Flags uint32
	// The transaction sequence number of LoanBrokerSet transaction that created this LoanBroker object.
	Sequence uint32
	// A sequential identifier for Loan objects, incremented each time a new Loan is created by this LoanBroker instance.
	LoanSequence uint32
	// Identifies the page where this item is referenced in the owner's directory.
	OwnerNode string
	// Identifies the page where this item is referenced in the Vault's pseudo-account owner's directory.
	VaultNode string
	// The ID of the Vault object associated with this Lending Protocol Instance.
	VaultID types.Hash256
	// The address of the LoanBroker pseudo-account.
	Account types.Address
	// The address of the Loan Broker account.
	Owner types.Address
	// The number of active Loans issued by the LoanBroker.
	OwnerCount *types.OwnerCount `json:",omitempty"`
	// The total asset amount the protocol owes the Vault, including interest.
	DebtTotal *types.XRPLNumber `json:",omitempty"`
	// The maximum amount the protocol can owe the Vault. The default value of 0 means there is no limit to the debt.
	DebtMaximum types.XRPLNumber
	// The total amount of first-loss capital deposited into the Lending Protocol.
	CoverAvailable *types.XRPLNumber `json:",omitempty"`
	// The 1/10th basis point of the DebtTotal that the first loss capital must cover.
	// Valid values are between 0 and 100000 inclusive. A value of 1 is equivalent to 1/10 bps or 0.001%.
	CoverRateMinimum *types.CoverRate `json:",omitempty"`
	// The 1/10th basis point of minimum required first loss capital that is liquidated to cover a Loan default.
	// Valid values are between 0 and 100000 inclusive. A value of 1 is equivalent to 1/10 bps or 0.001%.
	CoverRateLiquidation *types.CoverRate `json:",omitempty"`
	// The identifying hash of the transaction that most recently modified this entry.
	PreviousTxnID types.Hash256
	// The index of the ledger that contains the transaction that most recently modified this entry.
	PreviousTxnLgrSeq uint32
}

// EntryType returns the ledger entry type for LoanBroker.
func (*LoanBroker) EntryType() EntryType {
	return LoanBrokerEntry
}
