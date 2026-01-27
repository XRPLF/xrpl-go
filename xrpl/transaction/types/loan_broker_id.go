// Package types provides core transaction types and helpers for the XRPL Go library.
//
//revive:disable:var-naming
package types

// LoanBrokerID represents a Loan Broker identifier.
// It must be a 64 characters hexadecimal string.
type LoanBrokerID string

// Value returns the string representation of LoanBrokerID.
func (l *LoanBrokerID) Value() string {
	return string(*l)
}
