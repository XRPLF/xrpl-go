//revive:disable:var-naming
package types

// Signer represents a transaction signer wrapper, containing signer data.
type Signer struct {
	SignerData SignerData `json:"Signer"`
}

// Flatten returns a JSON-like map for the Signer, embedding its SignerData.
func (s *Signer) Flatten() map[string]any {
	flattened := make(map[string]any)
	flattened["Signer"] = s.SignerData.Flatten()
	return flattened
}

// SignerData holds the account, signature, and public key fields for a signer.
type SignerData struct {
	Account       Address
	TxnSignature  string
	SigningPubKey string
}

// FlatSignerData is a flattened map representation of SignerData for JSON serialization.
type FlatSignerData map[string]any

// Flatten returns a map[string]interface{} containing the populated SignerData fields.
func (sd *SignerData) Flatten() map[string]any {
	flattened := make(map[string]any)
	if sd.Account != "" {
		flattened["Account"] = sd.Account.String()
	}
	if sd.TxnSignature != "" {
		flattened["TxnSignature"] = sd.TxnSignature
	}
	if sd.SigningPubKey != "" {
		flattened["SigningPubKey"] = sd.SigningPubKey
	}
	return flattened
}
