//revive:disable:var-naming
package types

// Oracle represents the oracle query response, containing an account and document ID.
type Oracle struct {
	Account          string `json:"account"`
	OracleDocumentID any    `json:"oracle_document_id"`
}
