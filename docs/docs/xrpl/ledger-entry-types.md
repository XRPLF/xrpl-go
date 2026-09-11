# ledger-entry-types

## Overview

The `ledger-entry-types` package contains types and functions to handle ledger objects. They are used by other packages, like [`transaction`](/docs/xrpl/transaction) to type the transaction's fields.

- [`AccountRoot`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/accountroot)
- [`Amendments`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/amendments)
- [`AMM`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/amm)
- [`Bridge`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/bridge)
- [`Check`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/check)
- [`Credential`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/credential)
- `Delegate`
- [`DepositPreauth`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/depositpreauth)
- [`Did`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/did)
- [`DirectoryNode`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/directorynode)
- [`Escrow`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/escrow)
- [`FeeSettings`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/feesettings)
- [`Hashes`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/ledgerhashes)
- [`Loan`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/loan)
- [`LoanBroker`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/loanbroker)
- [`MPToken`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/mptoken)
- [`MPTokenIssuance`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/mptokenissuance)
- [`NegativeUNL`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/negativeunl)
- [`NFTokenOffer`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/nftokenoffer)
- [`NFTokenPage`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/nftokenpage)
- [`Offer`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/offer)
- [`Oracle`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/oracle)
- [`PayChannel`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/paychannel)
- [`PermissionedDomain`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/permissioneddomain)
- [`RippleState`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/ripplestate)
- [`SignerList`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/signerlist)
- [`Ticket`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/ticket)
- [`Vault`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/vault)
- [`XChainOwnedClaimID`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/xchainownedclaimid)
- [`XChainOwnedCreateAccountClaimID`](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/xchainownedcreateaccountclaimid)

## MPT and Oracle values

MPT ledger amount fields use quoted base-10 strings. This includes `MPToken.MPTAmount`, `MPToken.LockedAmount`, and the `MaximumAmount`, `OutstandingAmount`, and `LockedAmount` fields on `MPTokenIssuance`. `OwnerNode` fields are hexadecimal strings.

`MPTokenIssuance.ImmutableFlags` contains permanent restrictions and uses the `LsifMPT*` constants.

```go
if issuance.ImmutableFlags&ledger.LsifMPTMetadata != 0 {
 // Metadata can no longer change.
}
```

`PriceData.AssetPrice` is a pointer so absent and explicit zero prices stay distinct. Oracle prices accept the XLS-47 `Scale` range from `0` through `20`. `Flatten` omits `Scale` when `AssetPrice` is absent, and `Validate` rejects a nonzero `Scale` without a price.

### Confidential MPT fields

XLS-96 adds confidential state to both MPT entries. `MPTokenIssuance` carries `IssuerEncryptionKey`, the optional `AuditorEncryptionKey`, and `ConfidentialOutstandingAmount`, the confidential supply, which is a quoted base-10 string like the other issuance amounts. `MPToken` carries `HolderEncryptionKey`, the `ConfidentialBalanceSpending` and `ConfidentialBalanceInbox` ciphertexts, the `IssuerEncryptedBalance` and optional `AuditorEncryptedBalance` mirror ciphertexts, and `ConfidentialBalanceVersion`.

`LsfMPTCanHoldConfidentialBalance` on `MPTokenIssuance.Flags` reports whether the issuance allows confidential balances, and `LsifMPTCanHoldConfidentialBalance` on `ImmutableFlags` reports whether that setting can still change. See the [confidential guide](/docs/confidential) for how these fields are produced and consumed.

## Usage

To import the package, you can use the following code:

```go
import "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
```
