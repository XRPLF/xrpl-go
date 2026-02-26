package transaction

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// The NFTokenBurn transaction is used to remove a NFToken object from the NFTokenPage in which it is being held, effectively removing the token from the ledger (burning it).
//
// The sender of this transaction must be the owner of the NFToken to burn; or, if the NFToken has the LsfBurnable flag enabled, can be the issuer or the issuer's authorized NFTokenMinter account instead.
//
// If this operation succeeds, the corresponding NFToken is removed.
// If this operation empties the NFTokenPage holding the NFToken or results in consolidation, thus removing a NFTokenPage, the owner’s reserve requirement is reduced by one.
//
// Example:
//
// ```json
//
//	{
//		"TransactionType": "NFTokenBurn",
//		"Account": "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
//		"Owner": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
//		"Fee": "10",
//		"NFTokenID": "000B013A95F14B0044F78A264E41713C64B5F89242540EE208C3098E00000D65"
//	  }
//
// ```
type NFTokenBurn struct {
	BaseTx
	// The NFToken to be removed by this transaction.
	NFTokenID types.NFTokenID
	// (Optional) The owner of the NFToken to burn. Only used if that owner is different than the account sending this transaction.
	// The issuer or authorized minter can use this field to burn NFTs that have the LsfBurnable flag enabled.
	Owner types.Address `json:",omitempty"`
}

// TxType returns the type of the transaction (NFTokenBurn).
func (*NFTokenBurn) TxType() TxType {
	return NFTokenBurnTx
}

// Flatten returns a map of the NFTokenBurn transaction fields.
func (n *NFTokenBurn) Flatten() FlatTransaction {
	flattened := n.BaseTx.Flatten()

	flattened["TransactionType"] = "NFTokenBurn"

	if n.Owner != "" {
		flattened["Owner"] = n.Owner.String()
	}

	if n.NFTokenID != "" {
		flattened["NFTokenID"] = n.NFTokenID.String()
	}

	return flattened
}

// Validate checks the validity of the NFTokenBurn fields.
func (n *NFTokenBurn) Validate() (bool, error) {
	ok, err := n.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	// check owner is a valid xrpl address
	if n.Owner != "" && !addresscodec.IsValidAddress(n.Owner.String()) {
		return false, ErrInvalidOwner
	}

	// check NFTokenID is a valid hexadecimal string
	if !typecheck.IsHex(n.NFTokenID.String()) {
		return false, ErrInvalidNFTokenID
	}

	return true, nil
}
