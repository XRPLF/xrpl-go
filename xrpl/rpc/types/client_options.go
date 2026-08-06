// Package types contains data structures for RPC client configuration and options.
//
//revive:disable:var-naming
package types

import (
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

// SubmitOptions specifies options for submitting a transaction via RPC.
// A nil *SubmitOptions and a zero-value SubmitOptions are equivalent: Autofill
// and FailHard are false and no wallet is supplied. Set Autofill explicitly to
// true to populate missing fields before signing. AccountDelete submissions
// force fail_hard regardless of FailHard.
type SubmitOptions struct {
	// Autofill populates missing transaction fields before signing when true.
	Autofill bool
	// Wallet signs an otherwise unsigned transaction.
	Wallet *wallet.Wallet
	// FailHard requests fail_hard submission. AccountDelete always enables it.
	FailHard bool
}
