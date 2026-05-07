package transaction

import (
	"fmt"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
)

const (
	// MinSigners is the minimum number of signers required in the SignerList.
	MinSigners = 1

	// MaxSigners is the maximum number of signers allowed in the SignerList.
	MaxSigners = 32
)

// SignerListSet creates, replaces, or removes a list of signers that can be used to multi-sign a transaction.
// This transaction type was introduced by the MultiSign amendment.
//
// Example:
//
// ```json
//
//	{
//	    "Flags": 0,
//	    "TransactionType": "SignerListSet",
//	    "Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
//	    "Fee": "12",
//	    "SignerQuorum": 3,
//	    "SignerEntries": [
//	        {
//	            "SignerEntry": {
//	                "Account": "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
//	                "SignerWeight": 2
//	            }
//	        },
//	        {
//	            "SignerEntry": {
//	                "Account": "rUpy3eEg8rqjqfUoLeBnZkscbKbFsKXC3v",
//	                "SignerWeight": 1
//	            }
//	        },
//	        {
//	            "SignerEntry": {
//	                "Account": "raKEEVSGnKSD9Zyvxu4z6Pqpm4ABH8FS6n",
//	                "SignerWeight": 1
//	            }
//	        }
//	    ]
//	}
//
// `
type SignerListSet struct {
	BaseTx
	// A target number for the signer weights. A multi-signature from this list is valid only if the sum weights of the signatures provided is greater than or equal to this value.
	// To delete a signer list, use the value 0. Needs to be an uint32.
	SignerQuorum any
	// (Omitted when deleting) Array of SignerEntry objects, indicating the addresses and weights of signers in this list.
	// This signer list must have at least 1 member and no more than 32 members.
	// No address may appear more than once in the list, nor may the Account submitting the transaction appear in the list.
	SignerEntries []ledger.SignerEntryWrapper
}

// TxType returns the transaction type for this transaction (SignerListSet).
func (*SignerListSet) TxType() TxType {
	return SignerListSetTx
}

// Flatten returns the flattened map of the SignerListSet transaction.
func (s *SignerListSet) Flatten() FlatTransaction {
	flattened := s.BaseTx.Flatten()

	flattened["TransactionType"] = "SignerListSet"

	if s.SignerQuorum != nil {
		flattened["SignerQuorum"] = s.SignerQuorum
	}

	if len(s.SignerEntries) > 0 {
		signerEntries := make([]any, len(s.SignerEntries))
		for i, entry := range s.SignerEntries {
			signerEntry := make(map[string]any)

			signerEntry["SignerEntry"] = entry.SignerEntry.Flatten()
			signerEntries[i] = signerEntry
		}
		flattened["SignerEntries"] = signerEntries
	}

	return flattened
}

// Validate checks if the SignerListSet struct is valid.
func (s *SignerListSet) Validate() (bool, error) {
	ok, err := s.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	sq, ok := s.SignerQuorum.(uint32)
	zeroQuorum := ((ok && sq == uint32(0)) || s.SignerQuorum == nil)

	// All other checks are for if SignerQuorum is greater than 0
	if zeroQuorum && len(s.SignerEntries) == 0 {
		return true, nil
	}

	if zeroQuorum && len(s.SignerEntries) > 0 {
		return false, ErrInvalidQuorumAndEntries
	}

	// Check if SignerEntries has at least 1 entry and no more than 32 entries
	if len(s.SignerEntries) < MinSigners || len(s.SignerEntries) > MaxSigners {
		return false, fmt.Errorf("%w: must have at least %d entry and no more than %d entries",
			ErrInvalidSignerEntries, MinSigners, MaxSigners)
	}

	seenSignerAccounts := make(map[string]struct{}, len(s.SignerEntries))
	sumSignerWeights := uint64(0)
	txAccount := s.Account.String()
	for _, signerEntry := range s.SignerEntries {
		// Check if WalletLocator is an hexadecimal string for each SignerEntry
		if signerEntry.SignerEntry.WalletLocator != "" && !typecheck.IsHex(signerEntry.SignerEntry.WalletLocator.String()) {
			return false, ErrInvalidWalletLocator
		}

		// Check if Account is a valid xrpl address for each SignerEntry
		signerAccount := signerEntry.SignerEntry.Account.String()
		if !addresscodec.IsValidAddress(signerAccount) {
			return false, ErrInvalidAccount
		}

		if signerAccount == txAccount {
			return false, ErrSignerAccountMatchesAccount
		}

		if _, ok := seenSignerAccounts[signerAccount]; ok {
			return false, ErrDuplicateSignerAccount
		}
		seenSignerAccounts[signerAccount] = struct{}{}

		if signerEntry.SignerEntry.SignerWeight == 0 {
			return false, ErrInvalidSignerWeight
		}

		sumSignerWeights += uint64(signerEntry.SignerEntry.SignerWeight)
	}

	// Check SignerQuorum is less than or equal to the sum of all SignerWeights
	if uint64(sq) > sumSignerWeights {
		return false, ErrSignerQuorumGreaterThanSumOfSignerWeights
	}

	return true, nil
}
