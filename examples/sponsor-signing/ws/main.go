package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
)

func main() {
	w, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Println(err)
		return
	}

	sponsor, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Println(err)
		return
	}

	// Requires a network with Sponsor and fixCleanup3_4_0 enabled.
	client := websocket.NewClient(
		websocket.NewClientConfig().
			WithHost("wss://s.devnet.rippletest.net:51233").
			WithFaucetProvider(faucet.NewDevnetFaucetProvider()),
	)
	defer func() {
		if err := client.Disconnect(); err != nil {
			fmt.Println("Error disconnecting:", err)
		}
	}()

	fmt.Println("⏳ Connecting to server...")
	if err := client.Connect(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("✅ Connected to server")
	fmt.Println()

	fmt.Println("⏳ Funding wallets...")
	if err := client.FundWallet(&w); err != nil {
		fmt.Println(err)
		return
	}
	if err := client.FundWallet(&sponsor); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("💸 Wallets funded")

	// The account sends 1 drop. The sponsor pays the fee and receives the payment.
	payment := transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account: w.ClassicAddress,
		},
		Destination: sponsor.ClassicAddress,
		Amount:      types.XRPCurrencyAmount(1),
	}

	fmt.Println("⏳ Sending co-signed sponsor payment...")
	if err := sendCosignedPayment(client, w, sponsor, payment); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("⏳ Sending multisigned sponsor payment...")
	if err := sendMultisignedPayment(client, w, sponsor, payment); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("⏳ Sending pre-funded sponsor payment...")
	if err := sendPrefundedPayment(client, w, sponsor, payment); err != nil {
		fmt.Println(err)
		return
	}
}

func sendCosignedPayment(client *websocket.Client, account, sponsor wallet.Wallet, payment transaction.Payment) error {
	flatTx := payment.Flatten()
	flatTx["Sponsor"] = sponsor.ClassicAddress.String()
	flatTx["SponsorFlags"] = types.SpfSponsorFee

	if err := client.Autofill(&flatTx); err != nil {
		return err
	}

	txBlob, _, err := account.Sign(flatTx)
	if err != nil {
		return err
	}

	// Add the sponsor signature after the account signs. Do not autofill again.
	_, txBlob, _, err = wallet.SignAsSponsorBlob(sponsor, txBlob, nil)
	if err != nil {
		return err
	}

	return submitBlob(client, txBlob, "Co-signed payment")
}

func sendMultisignedPayment(client *websocket.Client, account, sponsor wallet.Wallet, payment transaction.Payment) error {
	// Install a 2-of-2 signer list on the sponsor. The signers need no funding.
	var signers []wallet.Wallet
	var entries []ledger.SignerEntryWrapper
	for range 2 {
		signer, err := wallet.New(crypto.ED25519())
		if err != nil {
			return err
		}
		signers = append(signers, signer)
		entries = append(entries, ledger.SignerEntryWrapper{
			SignerEntry: ledger.SignerEntry{
				Account:      signer.ClassicAddress,
				SignerWeight: 1,
			},
		})
	}

	list := transaction.SignerListSet{
		BaseTx: transaction.BaseTx{
			Account: sponsor.ClassicAddress,
		},
		SignerQuorum:  2,
		SignerEntries: entries,
	}
	if err := submitSetup(client, sponsor, list.Flatten(), "Sponsor signer list"); err != nil {
		return err
	}

	flatTx := payment.Flatten()
	flatTx["Sponsor"] = sponsor.ClassicAddress.String()
	flatTx["SponsorFlags"] = types.SpfSponsorFee

	// Include the two sponsor signers in the fee calculation before signing.
	if err := client.AutofillMultisigned(&flatTx, uint64(len(signers))); err != nil {
		return err
	}

	txBlob, _, err := account.Sign(flatTx)
	if err != nil {
		return err
	}

	var fragments []transaction.FlatTransaction
	for _, signer := range signers {
		// Each sponsor signer signs the same account-signed blob.
		fragment, _, _, err := wallet.SignAsSponsorBlob(signer, txBlob, &wallet.SignAsSponsorOptions{Multisign: true})
		if err != nil {
			return err
		}
		fragments = append(fragments, fragment)
	}

	_, txBlob, err = wallet.CombineSponsorSigners(fragments)
	if err != nil {
		return err
	}

	return submitBlob(client, txBlob, "Multisigned payment")
}

func sendPrefundedPayment(client *websocket.Client, account, sponsor wallet.Wallet, payment transaction.Payment) error {
	// Deposit 1 XRP into a fee pool without require-sign flags.
	feeAmountDelta := "1000000"
	pool := transaction.SponsorshipSet{
		BaseTx: transaction.BaseTx{
			Account: sponsor.ClassicAddress,
		},
		Sponsee:        account.ClassicAddress,
		FeeAmountDelta: &feeAmountDelta,
	}
	if err := submitSetup(client, sponsor, pool.Flatten(), "Pre-funded sponsorship"); err != nil {
		return err
	}

	flatTx, err := wallet.AddPreFundedSponsor(payment.Flatten(), sponsor.ClassicAddress, types.SpfSponsorFee)
	if err != nil {
		return err
	}

	if err := client.Autofill(&flatTx); err != nil {
		return err
	}

	// Only the account signs. The fee pool supplies the sponsor's authorization.
	txBlob, _, err := account.Sign(flatTx)
	if err != nil {
		return err
	}

	return submitBlob(client, txBlob, "Pre-funded payment")
}

func submitSetup(client *websocket.Client, signer wallet.Wallet, flatTx transaction.FlatTransaction, label string) error {
	if err := client.Autofill(&flatTx); err != nil {
		return err
	}

	txBlob, _, err := signer.Sign(flatTx)
	if err != nil {
		return err
	}

	return submitBlob(client, txBlob, label)
}

func submitBlob(client *websocket.Client, txBlob, label string) error {
	// Submit the final blob unchanged. Never autofill or sign it again.
	response, err := client.SubmitTxBlobAndWait(txBlob, true)
	if err != nil {
		return err
	}

	// A validated transaction can still have a failure result.
	if !response.Validated || response.Meta.TransactionResult != transaction.TesSUCCESS.String() {
		return fmt.Errorf("%s failed: validated=%t, result=%s", label, response.Validated, response.Meta.TransactionResult)
	}

	fmt.Printf("✅ %s submitted\n", label)
	fmt.Printf("🌐 Hash: %s\n", response.Hash.String())
	fmt.Printf("🌐 Validated: %t\n", response.Validated)
	return nil
}
