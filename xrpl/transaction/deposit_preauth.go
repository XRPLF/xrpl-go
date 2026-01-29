package transaction

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// DepositPreauth gives pre-approval for another account to deliver payments to the sender. (Requires the DepositPreauth amendment)
//
// ```json
//
//	{
//	  "TransactionType" : "DepositPreauth",
//	  "Account" : "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
//	  "Authorize" : "rEhxGqkqPPSxQ3P25J66ft5TwpzV14k2de",
//	  "Fee" : "10",
//	  "Flags" : 2147483648,
//	  "Sequence" : 2
//	}
//
// ```
type DepositPreauth struct {
	BaseTx
	// (Optional) The XRP Ledger address of the sender to preauthorize.
	Authorize types.Address `json:",omitempty"`
	// A set of credentials to authorize.
	AuthorizeCredentials []types.AuthorizeCredentialsWrapper `json:",omitempty"`
	// (Optional) The XRP Ledger address of a sender whose preauthorization should be revoked.
	Unauthorize types.Address `json:",omitempty"`
	// A set of credentials whose preauthorization should be revoked.
	UnauthorizeCredentials []types.AuthorizeCredentialsWrapper `json:",omitempty"`
}

// TxType implements the TxType method for the DepositPreauth struct.
func (*DepositPreauth) TxType() TxType {
	return DepositPreauthTx
}

// Flatten implements the Flatten method for the DepositPreauth struct.
func (d *DepositPreauth) Flatten() FlatTransaction {
	flattened := d.BaseTx.Flatten()

	flattened["TransactionType"] = DepositPreauthTx.String()

	if d.Authorize != "" {
		flattened["Authorize"] = d.Authorize.String()
	}

	if d.Unauthorize != "" {
		flattened["Unauthorize"] = d.Unauthorize.String()
	}

	if len(d.AuthorizeCredentials) > 0 {
		flattenedAuthorizeCredentials := make([]any, 0, len(d.AuthorizeCredentials))
		for _, credential := range d.AuthorizeCredentials {
			flattenedAuthorizeCredential := credential.Flatten()
			if flattenedAuthorizeCredential != nil {
				flattenedAuthorizeCredentials = append(flattenedAuthorizeCredentials, flattenedAuthorizeCredential)
			}
		}
		flattened["AuthorizeCredentials"] = flattenedAuthorizeCredentials
	}

	if len(d.UnauthorizeCredentials) > 0 {
		flattenedUnauthorizeCredentials := make([]any, 0, len(d.UnauthorizeCredentials))
		for _, credential := range d.UnauthorizeCredentials {
			flattenedUnauthorizeCredential := credential.Flatten()
			if flattenedUnauthorizeCredential != nil {
				flattenedUnauthorizeCredentials = append(flattenedUnauthorizeCredentials, flattenedUnauthorizeCredential)
			}
		}
		flattened["UnauthorizeCredentials"] = flattenedUnauthorizeCredentials
	}

	return flattened
}

// Validate implements the Validate method for the DepositPreauth struct.
func (d *DepositPreauth) Validate() (bool, error) {
	_, err := d.BaseTx.Validate()
	if err != nil {
		return false, err
	}

	// check that one of the four fields (Authorize, AuthorizeCredentials, Unauthorize, UnauthorizeCredentials) only is set
	if !d.isOnlyOneFieldSet() {
		return false, ErrDepositPreauthMustSetOnlyOneField
	}

	if d.Authorize != "" && d.Authorize.String() == d.Account.String() {
		return false, ErrDepositPreauthAuthorizeCannotBeSender
	}

	if d.Unauthorize != "" && d.Unauthorize.String() == d.Account.String() {
		return false, ErrDepositPreauthUnauthorizeCannotBeSender
	}

	if d.Authorize != "" && !addresscodec.IsValidAddress(d.Authorize.String()) {
		return false, ErrDepositPreauthInvalidAuthorize
	}

	if d.Unauthorize != "" && !addresscodec.IsValidAddress(d.Unauthorize.String()) {
		return false, ErrDepositPreauthInvalidUnauthorize
	}

	if len(d.AuthorizeCredentials) > 0 {
		for _, credential := range d.AuthorizeCredentials {
			if !credential.Credential.IsValid() {
				return false, ErrDepositPreauthInvalidAuthorizeCredentials
			}
		}
	}

	if len(d.UnauthorizeCredentials) > 0 {
		for _, credential := range d.UnauthorizeCredentials {
			if !credential.Credential.IsValid() {
				return false, ErrDepositPreauthInvalidUnauthorizeCredentials
			}
		}
	}

	return true, nil
}

// isOnlyOneFieldSet returns true if only one field is set in the DepositPreauth struct.
func (d *DepositPreauth) isOnlyOneFieldSet() bool {
	var count int

	if d.Authorize != "" {
		count++
	}
	if len(d.AuthorizeCredentials) > 0 {
		count++
	}
	if d.Unauthorize != "" {
		count++
	}
	if len(d.UnauthorizeCredentials) > 0 {
		count++
	}

	return count == 1
}
