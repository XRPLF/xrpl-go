// Package main runs a full XLS-96 confidential MPT lifecycle against devnet over JSON-RPC.
//
// It creates a confidential-capable issuance, registers the issuer encryption key, opts two
// holders in, and then moves value through every confidential transaction type: convert,
// merge inbox, send, convert back, and clawback.
//
// Confidential MPT needs a CGo-enabled build. Under CGO_ENABLED=0 this compiles and then
// fails at the first key generation with mptcrypto.ErrCgoRequired.
package main

import (
	"encoding/json"
	"fmt"

	"github.com/Peersyst/xrpl-go/confidential/builder"
	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	ledgerentries "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

const (
	// issuedAmount is the public MPT balance the issuer pays the holder, which the convert
	// then moves into confidential form.
	issuedAmount = 100
	// sentAmount is what the holder sends confidentially to the receiver.
	sentAmount = 40
)

// balanceSearch bounds every decryption in this example. ElGamal decryption searches the
// interval linearly, so a real application should use the narrowest range it can justify.
var balanceSearch = elgamal.AmountRange{Low: 0, High: issuedAmount}

func main() {
	// Configure the client
	fmt.Println("⏳ Setting up client...")
	cfg, err := rpc.NewClientConfig(
		"https://s.devnet.rippletest.net:51234",
		rpc.WithFaucetProvider(faucet.NewDevnetFaucetProvider()),
	)
	if err != nil {
		fmt.Println("❌ Error: client config:", err)
		return
	}
	client := rpc.NewClient(cfg)
	fmt.Println("✅ Client configured!")
	fmt.Println()

	// Every account that touches a confidential balance needs an ElGamal keypair, which is
	// unrelated to the XRPL keypair that signs. The private keys never reach the ledger.
	issuerKey, err := elgamal.GenerateKeypair()
	if err != nil {
		fmt.Println("❌ Error: issuer keypair:", err)
		return
	}
	holderKey, err := elgamal.GenerateKeypair()
	if err != nil {
		fmt.Println("❌ Error: holder keypair:", err)
		return
	}
	receiverKey, err := elgamal.GenerateKeypair()
	if err != nil {
		fmt.Println("❌ Error: receiver keypair:", err)
		return
	}

	fmt.Println("⏳ Funding wallets...")
	// Create and fund the issuer wallet
	issuer, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Println("❌ Error creating issuer wallet:", err)
		return
	}
	if err := client.FundWallet(&issuer); err != nil {
		fmt.Println("❌ Error funding issuer wallet:", err)
		return
	}
	fmt.Println("💸 Issuer wallet funded!")
	// Create and fund the holder wallet
	holder, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Println("❌ Error creating holder wallet:", err)
		return
	}
	if err := client.FundWallet(&holder); err != nil {
		fmt.Println("❌ Error funding holder wallet:", err)
		return
	}
	fmt.Println("💸 Holder wallet funded!")
	// Create and fund the receiver wallet
	receiver, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Println("❌ Error creating receiver wallet:", err)
		return
	}
	if err := client.FundWallet(&receiver); err != nil {
		fmt.Println("❌ Error funding receiver wallet:", err)
		return
	}
	fmt.Println("💸 Receiver wallet funded!")
	fmt.Println()

	// 1. Create an issuance that can hold confidential balances. The capability has to be
	// set here or by a later MPTokenIssuanceSet, and it can never be cleared once set.
	maximumAmount := types.MPTAmount(issuedAmount)
	create := &transaction.MPTokenIssuanceCreate{
		BaseTx:        transaction.BaseTx{Account: issuer.GetAddress()},
		MaximumAmount: &maximumAmount,
	}
	create.SetMPTCanTransferFlag()
	create.SetMPTCanHoldConfidentialBalanceFlag()
	create.SetMPTCanClawbackFlag()
	if err := submit(client, &issuer, create.Flatten(), "MPTokenIssuanceCreate"); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	// The issuance ID only exists once the create is validated, so it is read back rather
	// than predicted.
	issuanceID, err := issuanceID(client, issuer.GetAddress())
	if err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	fmt.Println("🆔 Issuance:", issuanceID)

	// 2. Register the issuer encryption key. Every confidential amount is also encrypted to
	// this key, which is what gives the issuer an auditable mirror of each holder balance.
	// MPTokenIssuanceCreate has no field for it, so it always takes a second transaction.
	setKeys := &transaction.MPTokenIssuanceSet{
		BaseTx:              transaction.BaseTx{Account: issuer.GetAddress()},
		MPTokenIssuanceID:   issuanceID,
		IssuerEncryptionKey: &issuerKey.PubKeyHex,
	}
	if err := submit(client, &issuer, setKeys.Flatten(), "MPTokenIssuanceSet"); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	// 3. Both holders need an MPToken before they can hold anything, confidential or not.
	// The issuance does not require authorization, so opting in is one-sided.
	for _, w := range []*wallet.Wallet{&holder, &receiver} {
		authorize := &transaction.MPTokenAuthorize{
			BaseTx:            transaction.BaseTx{Account: w.GetAddress()},
			MPTokenIssuanceID: issuanceID,
		}
		if err := submit(client, w, authorize.Flatten(), "MPTokenAuthorize"); err != nil {
			fmt.Println("❌ Error:", err)
			return
		}
	}

	// 4. Give the holder a public balance to convert.
	payment := &transaction.Payment{
		BaseTx:      transaction.BaseTx{Account: issuer.GetAddress()},
		Destination: holder.GetAddress(),
		Amount: types.MPTCurrencyAmount{
			MPTIssuanceID: issuanceID,
			Value:         types.MPTAmount(issuedAmount).String(),
		},
	}
	if err := submit(client, &issuer, payment.Flatten(), "Payment"); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	// 5. Convert the public balance into confidential form. The first convert also registers
	// the holder encryption key and carries the Schnorr proof that proves ownership of it,
	// which BuildConvert detects and assembles on its own.
	convert, err := builder.BuildConvert(client, builder.BuildConvertParams{
		Account:       holder.ClassicAddress.String(),
		IssuanceID:    issuanceID,
		Amount:        issuedAmount,
		HolderPrivKey: holderKey.PrivKeyHex,
		HolderPubKey:  holderKey.PubKeyHex,
	})
	if err != nil {
		fmt.Println("❌ Error: build convert:", err)
		return
	}
	if err := submit(client, &holder, convert.Flatten(), "ConfidentialMPTConvert"); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	// A convert credits the inbox, not the spending balance, so the holder cannot spend it
	// yet. The split is what keeps an incoming credit from invalidating a proof already in
	// flight against the spending balance.
	if err := report(client, holder.GetAddress(), holderKey, "holder after convert"); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	// 6. Merge the inbox into the spending balance. This carries no proof and no ciphertext,
	// so it is the cheapest confidential transaction, but it does bump the balance version.
	if err := mergeInbox(client, &holder, issuanceID); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	if err := report(client, holder.GetAddress(), holderKey, "holder after merge"); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	// 7. The receiver registers its own encryption key with a zero-value convert. A send
	// requires the destination to already carry a holder key and an inbox, so this is how a
	// holder opts in to receive before it holds anything.
	optIn, err := builder.BuildConvert(client, builder.BuildConvertParams{
		Account:       receiver.ClassicAddress.String(),
		IssuanceID:    issuanceID,
		Amount:        0,
		HolderPrivKey: receiverKey.PrivKeyHex,
		HolderPubKey:  receiverKey.PubKeyHex,
	})
	if err != nil {
		fmt.Println("❌ Error: build receiver opt-in:", err)
		return
	}
	if err := submit(client, &receiver, optIn.Flatten(), "ConfidentialMPTConvert (opt-in)"); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	// 8. Send confidentially. BuildSend reads and decrypts the holder's current spending
	// balance within BalanceRange, then encrypts the amount for the sender, the receiver,
	// and the issuer, and proves the whole set consistent without revealing it.
	send, err := builder.BuildSend(client, builder.BuildSendParams{
		Account:       holder.ClassicAddress.String(),
		Destination:   receiver.ClassicAddress.String(),
		IssuanceID:    issuanceID,
		Amount:        sentAmount,
		SenderPrivKey: holderKey.PrivKeyHex,
		SenderPubKey:  holderKey.PubKeyHex,
		BalanceRange:  balanceSearch,
	})
	if err != nil {
		fmt.Println("❌ Error: build send:", err)
		return
	}
	if err := submit(client, &holder, send.Flatten(), "ConfidentialMPTSend"); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	// 9. The receiver merges what it was sent before it can spend or convert it back.
	if err := mergeInbox(client, &receiver, issuanceID); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	if err := report(client, holder.GetAddress(), holderKey, "holder after send"); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	if err := report(client, receiver.GetAddress(), receiverKey, "receiver after merge"); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	// 10. Convert the holder's remainder back into a public balance. The amount becomes
	// public here, and the proof shows the confidential balance covered it.
	convertBack, err := builder.BuildConvertBack(client, builder.BuildConvertBackParams{
		Account:       holder.ClassicAddress.String(),
		IssuanceID:    issuanceID,
		Amount:        issuedAmount - sentAmount,
		HolderPrivKey: holderKey.PrivKeyHex,
		HolderPubKey:  holderKey.PubKeyHex,
		BalanceRange:  balanceSearch,
	})
	if err != nil {
		fmt.Println("❌ Error: build convert back:", err)
		return
	}
	if err := submit(client, &holder, convertBack.Flatten(), "ConfidentialMPTConvertBack"); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	if err := report(client, holder.GetAddress(), holderKey, "holder after convert back"); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	// 11. Claw back the receiver's whole confidential balance. The issuer never learns the
	// balance from the ledger alone, so BuildClawback derives it by decrypting the issuer
	// mirror ciphertext, and proves the revealed amount matches what that mirror holds.
	clawback, err := builder.BuildClawback(client, builder.BuildClawbackParams{
		Account:       issuer.ClassicAddress.String(),
		Holder:        receiver.ClassicAddress.String(),
		IssuanceID:    issuanceID,
		IssuerPrivKey: issuerKey.PrivKeyHex,
		BalanceRange:  balanceSearch,
	})
	if err != nil {
		fmt.Println("❌ Error: build clawback:", err)
		return
	}
	if err := submit(client, &issuer, clawback.Flatten(), "ConfidentialMPTClawback"); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	if err := report(client, receiver.GetAddress(), receiverKey, "receiver after clawback"); err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	fmt.Println("🎉 Confidential lifecycle complete")
}

// mergeInbox moves a holder's confidential inbox balance into its spending balance.
func mergeInbox(client *rpc.Client, holder *wallet.Wallet, issuanceID string) error {
	merge, err := builder.BuildMergeInbox(client, builder.BuildMergeInboxParams{
		Account:    holder.ClassicAddress.String(),
		IssuanceID: issuanceID,
	})
	if err != nil {
		return fmt.Errorf("build merge inbox: %w", err)
	}
	return submit(client, holder, merge.Flatten(), "ConfidentialMPTMergeInbox")
}

// report decrypts and prints a holder's own view of its confidential balances. Only the
// holder private key can do this, which is the point: the ledger stores ciphertexts.
func report(client *rpc.Client, address types.Address, key elgamal.Keypair, label string) error {
	token, err := mptoken(client, address)
	if err != nil {
		return err
	}

	inbox, err := decrypt(token.ConfidentialBalanceInbox, key.PrivKeyHex)
	if err != nil {
		return fmt.Errorf("%s inbox: %w", label, err)
	}
	spending, err := decrypt(token.ConfidentialBalanceSpending, key.PrivKeyHex)
	if err != nil {
		return fmt.Errorf("%s spending: %w", label, err)
	}

	// An absent MPTAmount means the public balance is zero, which the ledger omits.
	public := token.MPTAmount
	if public == "" {
		public = "0"
	}

	fmt.Printf("🔐 %s: public=%s inbox=%d spending=%d version=%d\n",
		label, public, inbox, spending, token.ConfidentialBalanceVersion)
	return nil
}

// decrypt reads one confidential balance. An absent ciphertext means the balance was never
// created, which reads as zero rather than as an error.
func decrypt(ciphertext, privateKey string) (uint64, error) {
	if ciphertext == "" {
		return 0, nil
	}
	return elgamal.Decrypt(ciphertext, privateKey, balanceSearch)
}

// mptoken reads the MPToken entry an address owns for the single issuance in this example.
func mptoken(client *rpc.Client, address types.Address) (ledgerentries.MPToken, error) {
	objects, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account:     address,
		Type:        account.MPTokenObject,
		LedgerIndex: common.Validated,
	})
	if err != nil {
		return ledgerentries.MPToken{}, fmt.Errorf("account objects: %w", err)
	}
	if len(objects.AccountObjects) != 1 {
		return ledgerentries.MPToken{}, fmt.Errorf("expected 1 MPToken for %s, got %d", address, len(objects.AccountObjects))
	}
	return decodeMPToken(objects.AccountObjects[0])
}

// decodeMPToken converts a raw account_objects entry into the typed ledger entry. The
// query returns objects as maps because one request can mix entry types.
func decodeMPToken(object map[string]any) (ledgerentries.MPToken, error) {
	encoded, err := json.Marshal(object)
	if err != nil {
		return ledgerentries.MPToken{}, fmt.Errorf("encode MPToken: %w", err)
	}

	var token ledgerentries.MPToken
	if err := json.Unmarshal(encoded, &token); err != nil {
		return ledgerentries.MPToken{}, fmt.Errorf("decode MPToken: %w", err)
	}
	return token, nil
}

// issuanceID reads back the ID of the single issuance the issuer created.
func issuanceID(client *rpc.Client, issuer types.Address) (string, error) {
	objects, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account:     issuer,
		Type:        account.MPTIssuanceObject,
		LedgerIndex: common.Validated,
	})
	if err != nil {
		return "", fmt.Errorf("account objects: %w", err)
	}
	if len(objects.AccountObjects) != 1 {
		return "", fmt.Errorf("expected 1 MPTokenIssuance, got %d", len(objects.AccountObjects))
	}

	id, ok := objects.AccountObjects[0]["mpt_issuance_id"].(string)
	if !ok {
		return "", fmt.Errorf("issuance response has no mpt_issuance_id")
	}
	return id, nil
}

// submit autofills, signs, and submits a transaction, then waits for it to validate.
// Autofill is what supplies the ten base fees a confidential transaction owes. It leaves
// the sequence a builder already resolved alone, which matters because every confidential
// proof binds that sequence.
func submit(client *rpc.Client, signer *wallet.Wallet, flat transaction.FlatTransaction, label string) error {
	fmt.Printf("⏳ Submitting %s...\n", label)
	if err := client.Autofill(&flat); err != nil {
		return fmt.Errorf("autofill %s: %w", label, err)
	}

	blob, _, err := signer.Sign(flat)
	if err != nil {
		return fmt.Errorf("sign %s: %w", label, err)
	}

	response, err := client.SubmitTxBlobAndWait(blob, false)
	if err != nil {
		return fmt.Errorf("submit %s: %w", label, err)
	}
	if !response.Validated || response.Meta.TransactionResult != "tesSUCCESS" {
		return fmt.Errorf("%s failed with %s", label, response.Meta.TransactionResult)
	}

	fmt.Printf("✅ %s validated in ledger %d!\n", label, response.LedgerIndex)
	fmt.Println("🌐 Transaction Hash:", response.Hash.String())
	fmt.Println()
	return nil
}
