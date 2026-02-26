package transaction

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// NFTokenModify is used to change the URI field of an NFT to point to a different URI in order to update the supporting data for the NFT.
// The NFT must have been minted with the TfMutable flag set. See Dynamic Non-Fungible Tokens (https://xrpl.org/docs/concepts/tokens/nfts/dynamic-nfts).
//
// # Example
//
// ```json
//
//	{
//	  "TransactionType": "NFTokenModify",
//	  "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
//	  "Owner": "rogue5HnPRSszD9CWGSUz8UGHMVwSSKF6",
//	  "Fee": "10",
//	  "Sequence": 33,
//	  "NFTokenID": "0008C350C182B4F213B82CCFA4C6F59AD76F0AFCFBDF04D5A048C0A300000007",
//	  "URI": "697066733A2F2F62616679626569636D6E73347A736F6C686C6976346C746D6E356B697062776373637134616C70736D6C6179696970666B73746B736D3472746B652F5665742E706E67"
//	}
//
// ```
type NFTokenModify struct {
	BaseTx
	// (Optional) Address of the owner of the NFT. If the Account and Owner are the same address, omit this field.
	Owner types.Address `json:",omitempty"`
	// Composite field that uniquely identifies the token.
	NFTokenID types.NFTokenID
	// (Optional) Up to 256 bytes of arbitrary data. In JSON, this should be encoded as a string of hexadecimal.
	// You can use the xrpl.convertStringToHex utility to convert a URI to its hexadecimal equivalent.
	// This is intended to be a URI that points to the data or metadata associated with the NFT.
	// The contents could decode to an HTTP or HTTPS URL, an IPFS URI, a magnet link, immediate data encoded as an RFC 2379 "data" URL, or even an issuer-specific encoding.
	// The URI is not checked for validity. If you do not specify a URI, the existing URI is deleted.
	URI types.NFTokenURI `json:",omitempty"`
}

// TxType returns the type of the transaction (NFTokenModify).
func (*NFTokenModify) TxType() TxType {
	return NFTokenModifyTx
}

// Flatten returns a map of the NFTokenModify transaction fields.
func (n *NFTokenModify) Flatten() FlatTransaction {
	flattened := n.BaseTx.Flatten()

	flattened["TransactionType"] = "NFTokenModify"

	if n.Owner != "" {
		flattened["Owner"] = n.Owner.String()
	}

	flattened["NFTokenID"] = n.NFTokenID.String()

	if n.URI != "" {
		flattened["URI"] = n.URI.String()
	}

	return flattened
}

// Validate checks the NFTokenModify fields for correctness and returns an error if invalid.
func (n *NFTokenModify) Validate() (bool, error) {
	// Validate the base transaction
	_, err := n.BaseTx.Validate()
	if err != nil {
		return false, err
	}

	// Check if the owner and account are not equal
	if n.Account == n.Owner {
		return false, ErrOwnerAccountConflict
	}

	// Check that the Owner is a valid XRPL address
	if n.Owner != "" && !addresscodec.IsValidAddress(n.Owner.String()) {
		return false, ErrInvalidOwner
	}

	// Check that the URI is a valid hex string
	if n.URI != "" && !typecheck.IsHex(n.URI.String()) {
		return false, ErrInvalidURI
	}

	return true, nil
}
