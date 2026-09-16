# Sponsor signing

Run all three fee-sponsorship flows with one command. The example creates fresh wallets, funds the account and sponsor, submits transactions, and checks that each transaction is validated with `tesSUCCESS`. No seeds, addresses, signer lists, or sponsorship objects need to be supplied.

Use test networks only. Each run uses fresh test accounts and leaves their ledger objects in place. It does not print or save private keys.

## Run on devnet

From the repository root, with the Go version required by `go.mod` installed:

```sh
go run ./examples/sponsor-signing -network devnet
```

Devnet is also the default if you omit `-network`. Internet access to the devnet RPC server and faucet is required. Funding and ledger validation can take a minute or more. Faucet failures stop the example. Retry later or use localnet.

On devnet, the example prints the node version and checks that both `Sponsor` and `fixCleanup3_4_0` are enabled **before funding wallets**. A version number alone does not prove that these amendments are enabled. Devnet was verified with `3.4.0-rc6`, with both amendments enabled.

## Run on localnet

Docker is required. From the repository root:

```sh
make run-localnet-linux/amd64
# On ARM64, use make run-localnet-linux/arm64 instead.
go run ./examples/sponsor-signing -network localnet
make stop-localnet
```

Wait for the node to start before running the example. The Makefile starts a standalone node with automatic ledger closing. The example connects to `http://localhost:5005` and funds wallets from the public standalone genesis account. It does not use an external faucet. The repository config forces `Sponsor` and `fixCleanup3_4_0` through its `[features]` section. These standalone rules are not reported as enabled by the feature API. The example prints a notice and lets the submitted transactions verify compatibility instead. Localnet was verified with `3.4.0-rc1`. If you already have a compatible standalone node at that address, do not start a second one. Restart an existing node after updating the config.

## What happens

1. **Co-signed:** The account signs a 1-drop payment to the sponsor. `SignAsSponsorBlob` adds the sponsor's single signature. No `Sponsorship` object is needed.
2. **Multisigned:** The example installs a 2-of-2 signer list on the sponsor. Both signer wallets sign the **same account-signed blob** independently. `CombineSponsorSigners` combines the fragments. The signer wallets do not need funding.
3. **Pre-funded:** The sponsor submits `SponsorshipSet` with `FeeAmountDelta: "1000000"` to deposit 1 XRP into a fee pool. No require-sign flags are set. `AddPreFundedSponsor` adds the sponsor fields before account signing. No sponsor signature is needed for this payment.

The account pays the payment amount. Sponsorship covers only the transaction fee, not the payment amount or reserves in this example. The sponsor also pays for its signer-list and sponsorship setup transactions. `SponsorshipSet` uses a flat map because a typed model is not yet available.

Fees are calculated before signing with `AutofillMultisigned`. The account single-signs, so only the two sponsor multisigners add to the planned signer count. A single sponsor signature adds no multisigner surcharge. If your account also multisigns, include its planned signer count too. Leave `Fee` absent so autofill can calculate it.

The final blob is submitted unchanged. Do not autofill or sign it again.

## Expected result

The output includes transaction hashes, fees, and these result lines:

```text
cosigned: validated=true, result=tesSUCCESS
Set sponsor signer list: validated=true, result=tesSUCCESS
multisigned: validated=true, result=tesSUCCESS
Create pre-funded sponsorship: validated=true, result=tesSUCCESS
prefunded: validated=true, result=tesSUCCESS
All three sponsor payments validated with tesSUCCESS.
```

A failed setup, signature, submission, or validated transaction result stops the example with a nonzero exit status. `validated=true` alone is not treated as success.
