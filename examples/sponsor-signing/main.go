package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/hash"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

func main() {
	network := flag.String("network", "devnet", "test network: devnet or localnet")
	flag.Parse()
	if err := run(*network); err != nil {
		log.Fatal(err)
	}
}

func run(network string) error {
	var cfg *rpc.Config
	var err error
	switch network {
	case "devnet":
		cfg, err = rpc.NewClientConfig("https://s.devnet.rippletest.net:51234/",
			rpc.WithFaucetProvider(faucet.NewDevnetFaucetProvider()))
	case "localnet":
		cfg, err = rpc.NewClientConfig("http://localhost:5005")
	default:
		return fmt.Errorf("unknown network %q: use devnet or localnet", network)
	}
	if err != nil {
		return err
	}
	client := rpc.NewClient(cfg)
	info, err := client.GetServerInfo(&server.InfoRequest{})
	if err != nil {
		return fmt.Errorf("connect to %s: %w", network, err)
	}
	fmt.Printf("Network: %s, rippled: %s\n", network, info.Info.BuildVersion)
	features, err := client.GetAllFeatures(&server.FeatureAllRequest{})
	if err != nil {
		return fmt.Errorf("check amendments: %w", err)
	}
	for _, name := range []string{"Sponsor", "fixCleanup3_4_0"} {
		enabled := false
		for _, feature := range features.Features {
			if feature.Name == name && feature.Enabled {
				enabled = true
			}
		}
		if !enabled {
			if network != "localnet" {
				return fmt.Errorf("%s requires enabled amendment %s: use a compatible localnet", network, name)
			}
			// Standalone [features] rules are not reported as enabled by the API.
			fmt.Printf("%s is not reported enabled. Localnet must force it in [features].\n", name)
		}
	}

	// Only the account and sponsor need funding. The sponsor is also the
	// payment destination. Multisigners can authorize with unfunded accounts.
	account, err := wallet.New(crypto.ED25519())
	if err != nil {
		return err
	}
	sponsor, err := wallet.New(crypto.ED25519())
	if err != nil {
		return err
	}
	for _, w := range []*wallet.Wallet{&account, &sponsor} {
		fmt.Println("Funding:", w.ClassicAddress)
		if network == "devnet" {
			err = client.FundWallet(w)
		} else {
			// Public standalone genesis seed. Never use it on a public network.
			var genesis wallet.Wallet
			genesis, err = wallet.FromSeed("snoPBrXtMeMyMHUVTgbuqAfg1SUTb", "")
			if err == nil {
				payment := transaction.Payment{
					BaseTx:      transaction.BaseTx{Account: genesis.ClassicAddress},
					Destination: w.ClassicAddress, Amount: types.XRPCurrencyAmount(100_000_000),
				}
				err = submitSetup(client, genesis, payment.Flatten(), "Fund account")
			}
		}
		if err != nil {
			return fmt.Errorf("fund %s: %w", w.ClassicAddress, err)
		}
	}

	// 1. The sponsor authorizes the fee with its own signature.
	if err := sponsoredPayment(client, account, sponsor, "cosigned", nil); err != nil {
		return err
	}

	// 2. Install a 2-of-2 signer list on the sponsor, not on the account.
	var signers []wallet.Wallet
	var entries []ledger.SignerEntryWrapper
	for range 2 {
		signer, err := wallet.New(crypto.ED25519())
		if err != nil {
			return err
		}
		signers = append(signers, signer)
		entries = append(entries, ledger.SignerEntryWrapper{SignerEntry: ledger.SignerEntry{
			Account: signer.ClassicAddress, SignerWeight: 1,
		}})
	}
	list := transaction.SignerListSet{
		BaseTx:       transaction.BaseTx{Account: sponsor.ClassicAddress},
		SignerQuorum: 2, SignerEntries: entries,
	}
	if err := submitSetup(client, sponsor, list.Flatten(), "Set sponsor signer list"); err != nil {
		return err
	}
	if err := sponsoredPayment(client, account, sponsor, "multisigned", signers); err != nil {
		return err
	}

	// 3. Deposit 1 XRP into a fee pool. No require-sign flags are set, so the
	// account can use it without a sponsor signature. There is no typed
	// SponsorshipSet model yet, so use the supported flat wire fields.
	pool := transaction.FlatTransaction{
		"TransactionType": "SponsorshipSet", "Account": sponsor.ClassicAddress.String(),
		"Sponsee": account.ClassicAddress.String(), "FeeAmountDelta": "1000000",
	}
	if err := submitSetup(client, sponsor, pool, "Create pre-funded sponsorship"); err != nil {
		return err
	}
	if err := sponsoredPayment(client, account, sponsor, "prefunded", nil); err != nil {
		return err
	}
	fmt.Println("All three sponsor payments validated with tesSUCCESS.")
	return nil
}

func sponsoredPayment(client *rpc.Client, account, sponsor wallet.Wallet, mode string, signers []wallet.Wallet) error {
	payment := transaction.Payment{
		BaseTx:      transaction.BaseTx{Account: account.ClassicAddress},
		Destination: sponsor.ClassicAddress, Amount: types.XRPCurrencyAmount(1),
	}
	tx := payment.Flatten()
	var err error
	if mode == "prefunded" {
		tx, err = wallet.AddPreFundedSponsor(tx, sponsor.ClassicAddress, types.SpfSponsorFee)
		if err != nil {
			return err
		}
	} else {
		tx["Sponsor"], tx["SponsorFlags"] = sponsor.ClassicAddress.String(), types.SpfSponsorFee
	}

	// Fee is absent. Autofill does not replace a pre-existing fee. The account
	// single-signs, so only sponsor multisigners add to the planned signer count.
	if err := client.AutofillMultisigned(&tx, uint64(len(signers))); err != nil {
		return fmt.Errorf("%s autofill: %w", mode, err)
	}
	accountBlob, txHash, err := account.Sign(tx)
	if err != nil {
		return err
	}
	finalBlob := accountBlob
	switch mode {
	case "cosigned":
		_, finalBlob, txHash, err = wallet.SignAsSponsorBlob(sponsor, accountBlob, nil)
	case "multisigned":
		var fragments []transaction.FlatTransaction
		for _, signer := range signers {
			// Every signer uses the SAME account-signed blob, not the previous result.
			fragment, _, _, signErr := wallet.SignAsSponsorBlob(signer, accountBlob, &wallet.SignAsSponsorOptions{Multisign: true})
			if signErr != nil {
				return signErr
			}
			fragments = append(fragments, fragment)
		}
		_, finalBlob, err = wallet.CombineSponsorSigners(fragments)
		if err == nil {
			txHash, err = hash.SignTxBlob(finalBlob)
		}
	}
	if err != nil {
		return fmt.Errorf("%s signing: %w", mode, err)
	}
	fmt.Printf("%s: fee=%v drops, hash=%s\n", mode, tx["Fee"], txHash)
	// Submit the final blob unchanged. Never autofill or sign it again.
	return submitBlob(client, finalBlob, mode)
}

func submitSetup(client *rpc.Client, signer wallet.Wallet, tx transaction.FlatTransaction, label string) error {
	if err := client.Autofill(&tx); err != nil {
		return fmt.Errorf("%s autofill: %w", label, err)
	}
	blob, _, err := signer.Sign(tx)
	if err != nil {
		return fmt.Errorf("%s signing: %w", label, err)
	}
	return submitBlob(client, blob, label)
}

func submitBlob(client *rpc.Client, blob, label string) error {
	result, err := client.SubmitTxBlobAndWait(blob, true)
	if err != nil {
		return fmt.Errorf("%s submission: %w", label, err)
	}
	// Validated alone does not mean success. A validated tec result is a failure.
	if !result.Validated || result.Meta.TransactionResult != transaction.TesSUCCESS.String() {
		return fmt.Errorf("%s: validated=%t, result=%s", label, result.Validated, result.Meta.TransactionResult)
	}
	fmt.Printf("%s: validated=true, result=%s\n", label, result.Meta.TransactionResult)
	return nil
}
