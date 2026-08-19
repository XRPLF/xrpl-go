// Package main shows how to construct a confidential MPT offline.
// The example does not connect, sign, or submit.
//
// Confidential MPT needs a CGo-enabled build. Under CGO_ENABLED=0 this compiles and then
// fails at the first key generation.
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Peersyst/xrpl-go/confidential/builder"
	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	issuerAddress = "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	holderAddress = "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe"
	// The account submitting the merge on the holder's behalf, per XLS-75.
	delegateAddress = "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"
	// The ID the create in step 1 would produce: its sequence, then the issuer AccountID.
	// Read the real one from validated MPTokenIssuanceCreate metadata.
	issuanceID = "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8"
)

func main() {
	issuerKey, err := elgamal.GenerateKeypair()
	if err != nil {
		log.Fatal(err)
	}
	holderKey, err := elgamal.GenerateKeypair()
	if err != nil {
		log.Fatal(err)
	}
	// Back up the private keys. Do not print them or send them to the ledger.

	create := &transaction.MPTokenIssuanceCreate{
		BaseTx: transaction.BaseTx{
			Account:  types.Address(issuerAddress),
			Sequence: 1, // Placeholder: use current validated account state.
		},
	}
	create.SetMPTCanHoldConfidentialBalanceFlag()
	create.SetMPTCanTransferFlag()

	issuerPublicKey := issuerKey.PubKeyHex
	setKeys := &transaction.MPTokenIssuanceSet{
		BaseTx: transaction.BaseTx{
			Account:  types.Address(issuerAddress),
			Sequence: 2, // Submit after the create transaction is validated.
		},
		MPTokenIssuanceID:   issuanceID,
		IssuerEncryptionKey: &issuerPublicKey,
	}

	// First, use the standard MPT flow to authorize and fund the holder.
	// The convert registers the holder key and credits the inbox.
	convert, err := builder.PrepareConvert(builder.ConvertParams{
		BuildConvertParams: builder.BuildConvertParams{
			// The first-time form binds this sequence into its Schnorr proof, so it must be
			// the holder's real one. A later autofill would leave a proof the network rejects.
			TxOptions:     builder.TxOptions{Sequence: 1},
			Account:       holderAddress,
			IssuanceID:    issuanceID,
			Amount:        100,
			HolderPrivKey: holderKey.PrivKeyHex,
			HolderPubKey:  holderKey.PubKeyHex,
		},
		IssuerPubKey: issuerKey.PubKeyHex,
		FirstTime:    true,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Merge the inbox and wait for validation before a spend. The merge is the one step here
	// that can carry both options: it carries no proof of its own, and ConfidentialMPTMergeInbox
	// is delegable, unlike the ConfidentialMPTConvert above. The Ticket is spent from the
	// holder's account root, not the delegate's, and exists only after the holder has landed a
	// TicketCreate reserving it. A merge bumps ConfidentialBalanceVersion even so, which is why
	// it must be validated before a send or convert-back binds that version into a proof.
	merge, err := builder.PrepareMergeInbox(builder.MergeInboxParams{
		BuildMergeInboxParams: builder.BuildMergeInboxParams{
			TxOptions: builder.TxOptions{
				TicketSequence: 2,
				Delegate:       delegateAddress,
			},
			Account:    holderAddress,
			IssuanceID: issuanceID,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	printTransaction("1. Create the confidential issuance", create.Flatten())
	printTransaction("2. Register the issuer key", setKeys.Flatten())
	printTransaction("3. Convert public value", convert.Flatten())
	printTransaction("4. Merge the inbox", merge.Flatten())
}

func printTransaction(label string, value transaction.FlatTransaction) {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n%s\n", label, encoded)
}
