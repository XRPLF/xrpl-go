//revive:disable:var-naming
package types

// IssuedCurrency represents an amount of a non-XRP currency issued by an account.
type IssuedCurrency struct {
	Currency string  `json:"currency"`
	Issuer   Address `json:"issuer"`
}

// Flatten returns a JSON-like map representing the IssuedCurrency fields.
func (i *IssuedCurrency) Flatten() map[string]any {
	flattened := make(map[string]any)
	flattened["currency"] = i.Currency
	flattened["issuer"] = i.Issuer.String()
	return flattened
}
