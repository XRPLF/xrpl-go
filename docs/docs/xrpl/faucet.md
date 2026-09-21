# Test funding

A faucet provides test XRP on Testnet or Devnet. Choose the provider for the same network as your client. Test XRP has no monetary value, and faucets do not fund Mainnet accounts.

## Fund your own test wallet

This complete example creates a temporary wallet and funds it on Testnet. It needs network access and faucet availability. It prints only the public address, not credentials. Do not use the temporary wallet for funds you need to keep.

```go
package main

import (
	"fmt"
	"log"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

func main() {
	cfg, err := rpc.NewClientConfig(
		"https://s.altnet.rippletest.net:51234/",
		rpc.WithFaucetProvider(faucet.NewTestnetFaucetProvider()),
	)
	if err != nil {
		log.Fatal(err)
	}
	client := rpc.NewClient(cfg)
	w, err := wallet.New(crypto.ED25519())
	if err != nil {
		log.Fatal(err)
	}
	if err := client.FundWallet(&w); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Funded Testnet address:", w.GetAddress())
}
```

For Devnet, use `https://s.devnet.rippletest.net:51234/` and `faucet.NewDevnetFaucetProvider()`. The WebSocket configuration also accepts `WithFaucetProvider`. `FundWallet` on a client requires this option. Without a provider the call panics instead of returning an error.

Faucets are external services and can be unavailable or rate-limited. Handle errors rather than assuming every new wallet is funded. Public test networks can reset their state.

## Custom providers

Implement `common.FaucetProvider` for a custom test network. You can also call a provider's `FundWallet(address)` directly. The client helper additionally handles account availability checks.

See the [Go faucet API](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/faucet), [XRPL test networks](https://xrpl.org/resources/dev-tools/xrp-faucets), and [Transactions](/docs/xrpl/transaction) for the next step.
