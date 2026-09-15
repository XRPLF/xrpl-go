package types

const (
	// SpfSponsorFee indicates that the sponsor pays the transaction fee.
	SpfSponsorFee uint32 = 1
	// SpfSponsorReserve indicates that the sponsor provides reserves.
	SpfSponsorReserve uint32 = 2
)

// SponsorSignature contains single-signature or multisignature authorization.
// Outside Batch, use both nonempty signature strings or a nonempty Signers
// list. Multisigning permits an absent or empty SigningPubKey, but no TxnSignature.
// Inner Batch transactions use only an explicitly empty SigningPubKey.
// Nil string pointers represent absent fields, not empty strings.
type SponsorSignature struct {
	SigningPubKey *string  `json:",omitempty"`
	TxnSignature  *string  `json:",omitempty"`
	Signers       []Signer `json:",omitempty"`
}

// Flatten returns the sponsor authorization fields for serialization.
func (s *SponsorSignature) Flatten() map[string]any {
	result := make(map[string]any)
	if s.SigningPubKey != nil {
		result["SigningPubKey"] = *s.SigningPubKey
	}
	if s.TxnSignature != nil {
		result["TxnSignature"] = *s.TxnSignature
	}
	if len(s.Signers) > 0 {
		signers := make([]map[string]any, len(s.Signers))
		for i := range s.Signers {
			signers[i] = s.Signers[i].Flatten()
		}
		result["Signers"] = signers
	}
	return result
}
