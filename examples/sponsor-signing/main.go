package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Peersyst/xrpl-go/xrpl/hash"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

// This example needs pre-existing accounts and a compatible node. It does not
// fund accounts, set signer lists, or create a ledger Sponsorship object.
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	mode := os.Getenv("SPONSOR_MODE")
	if mode != "cosigned" && mode != "multisigned" && mode != "prefunded" {
		return fmt.Errorf("set SPONSOR_MODE to cosigned, multisigned, or prefunded")
	}
	cfg, err := rpc.NewClientConfig(os.Getenv("XRPL_RPC_URL"))
	if err != nil {
		return err
	}
	client := rpc.NewClient(cfg)
	account, err := wallet.FromSeed(os.Getenv("ACCOUNT_SEED"), "")
	if err != nil {
		return fmt.Errorf("account wallet: %w", err)
	}
	payment := transaction.Payment{
		BaseTx:      transaction.BaseTx{Account: account.ClassicAddress},
		Destination: types.Address(os.Getenv("DESTINATION")),
		Amount:      types.XRPCurrencyAmount(1),
	}
	tx := payment.Flatten()
	sponsorAddress := types.Address(os.Getenv("SPONSOR_ADDRESS"))
	var sponsors []wallet.Wallet
	if mode == "prefunded" {
		tx, err = wallet.AddPreFundedSponsor(tx, sponsorAddress, types.SpfSponsorFee)
		if err != nil {
			return err
		}
	} else {
		seedNames := []string{"SPONSOR_SEED"}
		if mode == "multisigned" {
			seedNames = append(seedNames, "SECOND_SPONSOR_SIGNER_SEED")
		}
		for _, name := range seedNames {
			w, err := wallet.FromSeed(os.Getenv(name), "")
			if err != nil {
				return fmt.Errorf("%s wallet: %w", name, err)
			}
			sponsors = append(sponsors, w)
		}
		tx["Sponsor"], tx["SponsorFlags"] = sponsorAddress.String(), types.SpfSponsorFee
		if err := transaction.ValidateSponsorFields(tx); err != nil {
			return err
		}
	}

	// This example single-signs the account. A single sponsor adds zero extra
	// signers. For account multisigning, also add its planned signer count here.
	var extraSigners uint64
	if mode == "multisigned" {
		extraSigners = uint64(len(sponsors))
	}
	// Fee is absent. AutofillMultisigned does not replace a pre-existing fee.
	if err := client.AutofillMultisigned(&tx, extraSigners); err != nil {
		return err
	}
	accountBlob, txHash, err := account.Sign(tx)
	if err != nil {
		return err
	}
	finalBlob := accountBlob
	switch mode {
	case "cosigned":
		_, finalBlob, txHash, err = wallet.SignAsSponsorBlob(sponsors[0], accountBlob, nil)
	case "multisigned":
		var fragments []transaction.FlatTransaction
		for _, sponsor := range sponsors {
			// Every sponsor signer uses the SAME account-signed blob, not the
			// previous signer's result. Their signer list must meet ledger quorum.
			fragment, _, _, signErr := wallet.SignAsSponsorBlob(sponsor, accountBlob, &wallet.SignAsSponsorOptions{Multisign: true})
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
		return err
	}
	fmt.Println("Final transaction hash:", txHash)
	fmt.Println("Final transaction blob:", finalBlob)
	if os.Getenv("SUBMIT") == "1" {
		// Submit the final blob unchanged. Never autofill or sign it again.
		result, err := client.SubmitTxBlobAndWait(finalBlob, false)
		if err != nil {
			return err
		}
		fmt.Println("Validated:", result.Validated)
	}
	return nil
}
