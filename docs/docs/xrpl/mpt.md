# MPT operations and metadata

Use the core `xrpl/transaction` and `xrpl/transaction/types` packages for MPT transactions. Start with [Transactions](/docs/xrpl/transaction) for the common submission flow. These examples are fragments: supply the account addresses, issuance ID, and client for your application.

## Dynamic MPT

Use `ImmutableFlags` to make issuance capabilities or fields permanently immutable. Capability enablement on `MPTokenIssuanceSet` uses the normal transaction `Flags` field.

### Create an issuance

`MaximumAmount` is a quoted base-10 `types.MPTAmount` from `1` through `2^63-1`.

```go
maximumAmount := types.MPTAmount(1000000)
immutableFlags := types.ImmutableFlags(
	transaction.TifMPTCanLock |
		transaction.TifMPTMetadata,
)

create := transaction.MPTokenIssuanceCreate{
	BaseTx: transaction.BaseTx{
		Account: types.Address(issuer),
	},
	MaximumAmount:  &maximumAmount,
	ImmutableFlags: immutableFlags,
}

create.SetMPTCanLockFlag()
```

`TfMPTCanHoldConfidentialBalance` enables confidential balances. A non-zero transfer fee cannot be combined with that capability. See the [confidential guide](/docs/confidential) for the XLS-96 workflow this capability unlocks.

### Update an issuance

Use `TfMPTSet*` flags or the matching setters to enable a capability. Use `ImmutableFlags` only to prevent later changes.

```go
set := transaction.MPTokenIssuanceSet{
	BaseTx: transaction.BaseTx{
		Account: types.Address(issuer),
	},
	MPTokenIssuanceID: issuanceID,
}

set.SetMPTRequireAuthFlag()
set.SetMPTRequireAuthImmutableFlag()
```

Lock and unlock operations use `TfMPTLock` and `TfMPTUnlock`. A `Holder`-only transaction is a no-op and fails validation. Pair `Holder` with one of these flags. Lock and unlock operations cannot include capability, metadata, transfer-fee, or immutability mutations.

`IssuerEncryptionKey` and `AuditorEncryptionKey` register the ElGamal public keys XLS-96 confidential transfers need. Both are 33-byte compressed secp256k1 points, hex-encoded. `MPTokenIssuanceCreate` does not carry them, so an issuance enables the capability at creation and registers the keys with a later `MPTokenIssuanceSet`.

```go
issuerKey := issuerKeypair.PubKeyHex

set := transaction.MPTokenIssuanceSet{
	BaseTx: transaction.BaseTx{
		Account: types.Address(issuer),
	},
	MPTokenIssuanceID:   issuanceID,
	IssuerEncryptionKey: &issuerKey,
}
```

`AuditorEncryptionKey` requires `IssuerEncryptionKey` in the same transaction, and neither can be combined with `Holder`.

### Claw back MPT balances

`Clawback` accepts an MPT amount and requires the target `Holder`. The transaction `Account` must match the issuer encoded in the MPT issuance ID.

```go
clawback := transaction.Clawback{
	BaseTx: transaction.BaseTx{
		Account: types.Address(issuer),
	},
	Amount: types.MPTCurrencyAmount{
		MPTIssuanceID: issuanceID,
		Value:         "25",
	},
	Holder: types.Address(holder),
}
```

The `Holder` must be omitted for an issued-currency clawback. Tagged holder X-addresses are rejected.

## Ledger values

MPT ledger amount fields use quoted base-10 strings. This includes `MPToken.MPTAmount`, `MPToken.LockedAmount`, and the `MaximumAmount`, `OutstandingAmount`, and `LockedAmount` fields on `MPTokenIssuance`. `OwnerNode` fields are hexadecimal strings.

`MPTokenIssuance.ImmutableFlags` contains permanent restrictions and uses the `LsifMPT*` constants.

```go
if issuance.ImmutableFlags&ledger.LsifMPTMetadata != 0 {
	// Metadata can no longer change.
}
```

## Encode, validate, and decode metadata

Use `types.ParsedMPTokenMetadata` to construct metadata. Encoding serializes it to hex. **Encoding is not validation**: call `ValidateMPTokenMetadata` to check the schema before attaching it to a transaction.

This offline round-trip uses real encoded output, not a truncated placeholder:

```go
package main

import (
	"fmt"
	"log"

	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

func main() {
	subclass := "treasury"
	metadata := types.ParsedMPTokenMetadata{
		Ticker:        "TBILL",
		Name:          "Example Treasury Token",
		Icon:          "https://example.org/token.png",
		AssetClass:    "rwa",
		AssetSubclass: &subclass,
		IssuerName:    "Example Issuer",
	}
	encoded, err := types.EncodeMPTokenMetadata(metadata)
	if err != nil {
		log.Fatal(err)
	}
	if err := types.ValidateMPTokenMetadata(encoded); err != nil {
		log.Fatal(err)
	}
	decoded, err := types.DecodeMPTokenMetadata(encoded)
	if err != nil {
		log.Fatal(err)
	}
	create := transaction.MPTokenIssuanceCreate{
		BaseTx: transaction.BaseTx{
			Account: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		},
		MPTokenMetadata: types.MPTokenMetadata(encoded),
	}
	fmt.Println("Ticker:", decoded.Ticker)
	fmt.Println("Transaction type:", create.TxType())
}
```

This constructs a transaction but does not sign or submit it. Continue with the normal [transaction flow](/docs/xrpl/transaction).

Decoding accepts both long-form and compact-form JSON keys. Encoding uses compact keys with stable ordering. `ValidateMPTokenMetadata` checks hex/JSON syntax, the 1024-byte limit, required fields, allowed keys and values, and field-count limits. It returns `MPTokenMetadataValidationErrors` containing the validation failures.

See the [Go metadata types](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/transaction/types#ParsedMPTokenMetadata) for fields and the [XLS-89 schema](https://xls.xrpl.org/xls/XLS-0089-multi-purpose-token-metadata-schema.html) for requirements and enum values. In particular, `AssetClass: "rwa"` requires an asset subclass.

## Protocol reference

Use the XRPL references for [MPTokenIssuanceCreate](https://xrpl.org/docs/references/protocol/transactions/types/mptokenissuancecreate), [MPTokenIssuanceSet](https://xrpl.org/docs/references/protocol/transactions/types/mptokenissuanceset), and [Clawback](https://xrpl.org/docs/references/protocol/transactions/types/clawback). SDK support does not establish amendment activation on your network.
