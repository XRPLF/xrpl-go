# Transactions

Use `xrpl/transaction` to construct typed transactions. Use a client to prepare network fields, a wallet to sign locally, and the client to submit the signed blob.

```text
construct -> validate -> flatten -> autofill -> sign -> submit -> check result
```

[Go transaction types](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/transaction) · [XRPL transaction reference](https://xrpl.org/docs/references/protocol/transactions)

## Send test XRP

This complete example creates and funds two temporary Testnet wallets, then sends 1 XRP between them. It needs network access and faucet availability. It does not print or save credentials. **Do not use it with production funds.**

After [installing the SDK](/docs/installation), save it as `main.go`, then run `go mod tidy` and `go run .`:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := rpc.NewClientConfig(
		"https://s.altnet.rippletest.net:51234/",
		rpc.WithFaucetProvider(faucet.NewTestnetFaucetProvider()),
	)
	if err != nil {
		return err
	}
	client := rpc.NewClient(cfg)

	sender, err := wallet.New(crypto.ED25519())
	if err != nil {
		return err
	}
	receiver, err := wallet.New(crypto.ED25519())
	if err != nil {
		return err
	}
	for _, w := range []*wallet.Wallet{&sender, &receiver} {
		if err := client.FundWallet(w); err != nil {
			return err
		}
	}

	payment := transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account:         sender.GetAddress(),
			TransactionType: transaction.PaymentTx,
		},
		Destination: receiver.GetAddress(),
		Amount:      types.XRPCurrencyAmount(1_000_000),
	}
	if _, err := payment.Validate(); err != nil {
		return err
	}
	flat := payment.Flatten()
	if err := client.Autofill(&flat); err != nil {
		return err
	}
	blob, hash, err := sender.Sign(flat)
	if err != nil {
		return err
	}
	fmt.Println("Transaction hash:", hash)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	response, err := client.SubmitTxBlobAndWaitContext(ctx, blob, false)
	if err != nil {
		return err
	}
	if response.Meta.TransactionResult != "tesSUCCESS" {
		return fmt.Errorf("transaction result: %s", response.Meta.TransactionResult)
	}
	fmt.Println("Payment validated successfully")
	return nil
}
```

The amount is in **drops**, not XRP. See [Currency amounts](/docs/xrpl/currency) for exact conversions. The context above bounds submission and monitoring, not the earlier faucet or autofill calls.

## What each step does

| Step | Responsibility |
| --- | --- |
| Construct | Set transaction-specific fields, `BaseTx.Account`, and `BaseTx.TransactionType` |
| Validate | Check local field rules, not ledger authorization or available funds |
| Flatten | Produce the `FlatTransaction` map used by clients and signing |
| Autofill | Fill missing network fields and normalize supported input forms |
| Sign | Produce a signed blob and its transaction hash without a network call |
| Submit and wait | Send the blob and monitor for validation or expiry |
| Check result | Distinguish a successful transaction from a validated failure |

`Validate()` requires `BaseTx.TransactionType` on the struct and returns `transaction.ErrInvalidTransactionType` when it is empty. `Flatten()` writes the type into the flat map for concrete transaction structs either way, so it is not a replacement for validation. Do not change or autofill a transaction after signing it.

Both clients also offer `SubmitTxAndWaitContext` with a wallet and explicit `Autofill: true`. See [Submission and finality](/docs/xrpl/submission) for this convenience path, retry behavior, cancellation, and recovery after an uncertain outcome.

## Choose a transaction type

Use the [Go API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/transaction) to find the struct. Use the [XRPL transaction reference](https://xrpl.org/docs/references/protocol/transactions/types) for the meaning of fields, flags, and transaction results.

A transaction type being present in the SDK does not mean its amendment is enabled on your target network.

- [Wallets and signing](/docs/xrpl/wallet): account signatures, multisigning, and payment-channel claims.
- [Sponsorship](/docs/xrpl/sponsorship): sponsor fields, signatures, and optional preflight.
- [MPT operations and metadata](/docs/xrpl/mpt): Go-specific MPT construction and metadata encoding.
- [Confidential builders](/docs/confidential/builders): transactions with encrypted fields and proofs.

## Credential IDs

`types.CredentialIDs.IsValid()` requires 1 to 8 distinct, nonzero, 256-bit hexadecimal IDs. Each ID must contain exactly 64 hex characters. Hex letter case does not affect uniqueness. All transaction types with `CredentialIDs` use this validator and return `transaction.ErrInvalidCredentialIDs` when validation fails.

Leave an optional `CredentialIDs` field nil to omit it. An explicitly empty list fails transaction validation. The validator checks only the list format, not credential existence, ownership, expiry, or authorization. It rejects zero IDs offline, without checking whether the target network has enabled `fixCleanup3_4_0`.
