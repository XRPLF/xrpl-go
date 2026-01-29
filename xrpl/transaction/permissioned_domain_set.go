package transaction

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// PermissionedDomainSet creates a permissioned domain, or modifies one that you own.
// (Requires the PermissionedDomains amendment)
//
// ```json
//
//	{
//	  "TransactionType": "PermissionedDomainSet",
//	  "Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
//	  "Fee": "10",
//	  "Sequence": 390,
//	  "AcceptedCredentials": [
//	    {
//	        "Credential": {
//	            "Issuer": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
//	            "CredentialType": "6D795F63726564656E7469616C"
//	        }
//	    }
//	  ]
//	}
//
// ```
type PermissionedDomainSet struct {
	BaseTx
	// DomainID is the ledger entry ID of an existing permissioned domain to modify.
	// When omitted, it creates a new permissioned domain.
	DomainID string `json:",omitempty"`
	// AcceptedCredentials is a list of credentials that grant access to the domain.
	// An empty array indicates deletion of the field.
	AcceptedCredentials types.AuthorizeCredentialList
}

// TxType returns the type of the transaction.
func (p *PermissionedDomainSet) TxType() TxType {
	return PermissionedDomainSetTx
}

// Flatten returns a flattened map representation of the PermissionedDomainSet transaction.
func (p *PermissionedDomainSet) Flatten() FlatTransaction {
	flattened := p.BaseTx.Flatten()
	flattened["TransactionType"] = p.TxType().String()

	if p.DomainID != "" {
		flattened["DomainID"] = p.DomainID
	}

	flattened["AcceptedCredentials"] = p.AcceptedCredentials.Flatten()

	return flattened
}

// Validate validates the PermissionedDomainSet transaction.
func (p *PermissionedDomainSet) Validate() (bool, error) {
	if ok, err := p.BaseTx.Validate(); !ok {
		return false, err
	}

	if err := p.AcceptedCredentials.Validate(); err != nil {
		return false, err
	}

	return true, nil
}
