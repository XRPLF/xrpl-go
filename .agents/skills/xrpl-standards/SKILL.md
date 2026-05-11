---
name: xrpl-standards
description: "Reference for any XRPL Standard (XLS-N) when implementing or reviewing XRPL protocol features. Trigger on: XLS number (XLS-30, XLS-70), amendment name (AMM, Credentials, MPT, DID, NFToken, Batch, Escrow, Clawback, Firewall, Permissioned DEX), transaction type (AMMCreate, CredentialCreate, NFTokenMint, DelegateSet, PermissionedDomainSet, BatchSubmit), or ledger object name (Credential, AMM, MPToken, Delegate, Oracle, Bridge)."
---

# XRPL Standards

Raw specification files for all 77 XRPL Standards (XLS-1 through XLS-102), organized by topic. Read the relevant file to get the full spec — field definitions, transaction formats, ledger objects, failure conditions, invariants, and RPC changes.

## How to Use

The path depends on your environment. Find this skill's install directory and read the spec file directly:

```
Read <skill-dir>/references/<topic>/<file>.md
```

Common locations:
- **claude.ai**: `/mnt/skills/user/xrpl-standards/`
- **Claude Code** (`npx skills add`): `.claude/skills/xrpl-standards/` (relative to project root)

To auto-detect, search for the skill directory:
```bash
find /mnt/skills ~/.claude -name "INDEX.md" -path "*/xrpl-standards/references/*" 2>/dev/null | head -1 | sed 's|/references/INDEX.md||'
```

To list or fetch a spec not yet in refs:
```bash
bash <skill-dir>/scripts/list-xls.sh
bash <skill-dir>/scripts/fetch-xls.sh <number>
```

---

## identity — DID, credentials, sign-in

| XLS | Title | Status | File |
|-----|-------|--------|------|
| 40 | Decentralized Identity — `DID` obj, `DIDSet`, `DIDDelete` | Final | `references/identity/xls-0040.md` |
| 63 | SignIn Transaction — off-chain auth, `sfData` field | Stagnant | `references/identity/xls-0063.md` |
| 70 | On-Chain Credentials — `Credential` obj, `CredentialCreate/Accept/Delete`, `DepositPreauth` ext | Final | `references/identity/xls-0070.md` |

## tokens — NFT, MPT, URIToken

| XLS | Title | Status | File |
|-----|-------|--------|------|
| 20 | Non-Fungible Tokens — `NFToken`, `NFTokenMint/Burn/CreateOffer/AcceptOffer/CancelOffer` | Final | `references/tokens/xls-0020.md` |
| 33 | Multi-Purpose Tokens (MPT) — `MPTokenIssuance`, `MPTokenIssuanceCreate/Destroy/Set`, `MPTokenAuthorize` | Final | `references/tokens/xls-0033.md` |
| 35 | URITokens — `URIToken` obj, `URITokenMint/Burn/CreateOffer/BuyOffer/CancelOffer` | Draft | `references/tokens/xls-0035.md` |
| 46 | Dynamic NFTs — mutable metadata extension to NFToken | Final | `references/tokens/xls-0046.md` |
| 51 | NFToken Escrows | Stagnant | `references/tokens/xls-0051.md` |
| 52 | NFTokenMintOffer — combined mint + offer tx | Final | `references/tokens/xls-0052.md` |
| 54 | NFTokenOffer Destination Tag | Stagnant | `references/tokens/xls-0054.md` |
| 61 | Cross-Currency NFTokenAcceptOffer | Stagnant | `references/tokens/xls-0061.md` |
| 89 | MPT Metadata Schema | Final | `references/tokens/xls-0089.md` |
| 94 | Dynamic MPT — mutable MPToken fields | Draft | `references/tokens/xls-0094.md` |
| 96 | Confidential MPT — `ConfidentialMPTConvert/Send/MergeInbox/ConvertBack/Clawback` | Draft | `references/tokens/xls-0096.md` |
| 10 | Non-Transferable Token standard | Stagnant | `references/tokens/xls-0010.md` |
| 16 | NFT Metadata v1 | Stagnant | `references/tokens/xls-0016.md` |
| 24 | NFT Metadata v2 | Final | `references/tokens/xls-0024.md` |
| 26 | IOU Token Metadata via xrp-ledger.toml | Draft | `references/tokens/xls-0026.md` |

## defi — AMM, DEX, vault, lending, options

| XLS | Title | Status | File |
|-----|-------|--------|------|
| 30 | Automated Market Maker — `AMM` obj, `AMMCreate/Deposit/Withdraw/Vote/Bid/Delete` | Final | `references/defi/xls-0030.md` |
| 73 | AMM Clawback — `AMMClawback` tx | Draft | `references/defi/xls-0073.md` |
| 81 | Permissioned DEX — `OfferCreate`/`Payment` extensions, `PermissionedDomain` gating | Final | `references/defi/xls-0081.md` |
| 82 | MPT Integration into DEX, `MPTVersion2` amendment extending XLS-33/XLS-30 to MPT | Draft | `references/defi/xls-0082.md` |
| 65 | Single Asset Vault — `Vault` obj, `VaultCreate/Set/Delete/Deposit/Withdraw` | Draft | `references/defi/xls-0065.md` |
| 66 | Lending Protocol — `LoanBroker`, `Loan` objs, 9 new transactions | Draft | `references/defi/xls-0066.md` |
| 98 | Standard Metadata for Vaults, JSON convention for `Vault.Data` | Draft | `references/defi/xls-0098.md` |
| 62 | Options — option contract transactions | Stagnant | `references/defi/xls-0062.md` |
| 60 | Default AutoBridge | Stagnant | `references/defi/xls-0060.md` |

## payments — escrow, channels, batch, remit

| XLS | Title | Status | File |
|-----|-------|--------|------|
| 56 | Atomic/Batch Transactions — `Batch` tx, inner transaction handling | Final | `references/payments/xls-0056.md` |
| 85 | Token-Enabled Escrows — escrow/channel support for IOU and MPT | Final | `references/payments/xls-0085.md` |
| 100 | Smart Escrows — programmable escrow conditions | Draft | `references/payments/xls-0100.md` |
| 55 | Remit — `Remit` tx for bundled transfers | Final | `references/payments/xls-0055.md` |
| 34 | Token Payment Channels & Escrow (Withdrawn, superseded by XLS-85) | Withdrawn | `references/payments/xls-0034.md` |
| 67 | Charge | Stagnant | `references/payments/xls-0067.md` |
| 76 | Min Incoming Amount (Deprecated) | Deprecated | `references/payments/xls-0076.md` |

## accounts — permissions, freeze, signer lists, reserves

| XLS | Title | Status | File |
|-----|-------|--------|------|
| 74 | Account Permissions — granular tx-level permission flags | Final | `references/accounts/xls-0074.md` |
| 75 | Permission Delegation — `Delegate` obj, `DelegateSet` tx | Final | `references/accounts/xls-0075.md` |
| 80 | Permissioned Domains — `PermissionedDomain` obj, `PermissionedDomainSet/Delete` | Final | `references/accounts/xls-0080.md` |
| 86 | Firewall — account-level transaction filtering | Draft | `references/accounts/xls-0086.md` |
| 77 | Deep Freeze — enhanced freeze for trust lines | Final | `references/accounts/xls-0077.md` |
| 39 | Clawback — issuer clawback of IOU balances | Final | `references/accounts/xls-0039.md` |
| 68 | Sponsored Fees and Reserves — `Sponsorship` obj, fee/reserve delegation | Draft | `references/accounts/xls-0068.md` |
| 49 | Multiple Signer Lists | Draft | `references/accounts/xls-0049.md` |
| 64 | Pseudo-Account | Draft | `references/accounts/xls-0064.md` |
| 23 | Lite Accounts | Stagnant | `references/accounts/xls-0023.md` |
| 7  | Deletable Accounts | Final | `references/accounts/xls-0007.md` |
| 71 | Initial Owner Reserve Exemption | Stagnant | `references/accounts/xls-0071.md` |

## data — oracles, subscriptions

| XLS | Title | Status | File |
|-----|-------|--------|------|
| 47 | Price Oracles — `Oracle` obj, `OracleSet/Delete`, price aggregation RPC | Final | `references/data/xls-0047.md` |
| 78 | Subscriptions — on-chain payment streams | Draft | `references/data/xls-0078.md` |

## cross-chain — bridge, proof of payment

| XLS | Title | Status | File |
|-----|-------|--------|------|
| 38 | Cross-Chain Bridge — `Bridge` obj, `XChainCreateBridge`, locking/claiming txs | Final | `references/cross-chain/xls-0038.md` |
| 41 | XPOP — proof-of-payment standard, off-chain verification format | Final | `references/cross-chain/xls-0041.md` |

## smart-contracts — plugins, WASM, contracts

| XLS | Title | Status | File |
|-----|-------|--------|------|
| 101 | Smart Contracts — `Contract/ContractSource/ContractData` objs, `ContractCreate/Call/Modify/Delete` | Draft | `references/smart-contracts/xls-0101.md` |
| 102 | WASM VM — virtual machine specification | Draft | `references/smart-contracts/xls-0102.md` |
| 42  | Plugins — plugin ledger objects | Stagnant | `references/smart-contracts/xls-0042.md` |

## core — protocol fundamentals

| XLS | Title | Status | File |
|-----|-------|--------|------|
| 13 | Tickets — `Ticket` obj, `TicketCreate` tx | Final | `references/core/xls-0013.md` |
| 37 | CTID — Concise Transaction Identifier encoding | Final | `references/core/xls-0037.md` |
| 69 | Simulate — `simulate` RPC for dry-run execution | Final | `references/core/xls-0069.md` |
| 97 | Server Definitions — `server_definitions` RPC enhancements | Final | `references/core/xls-0097.md` |
| 22 | API Versioning | Final | `references/core/xls-0022.md` |
| 45 | Prepublish Validator Lists | Final | `references/core/xls-0045.md` |
| 17 | XFL — developer-friendly balance representation | Final | `references/core/xls-0017.md` |
| 5  | Tagged Addresses | Final | `references/core/xls-0005.md` |
| 9  | Blinded Tags | Stagnant | `references/core/xls-0009.md` |
| 11 | Retiring Amendments | Final | `references/core/xls-0011.md` |
| 12 | Secret Numbers | Final | `references/core/xls-0012.md` |
| 18 | Bootstrapping XRPLD Networks | Stagnant | `references/core/xls-0018.md` |
| 21 | Asset Code Prefixes | Stagnant | `references/core/xls-0021.md` |
| 25 | Enhanced Secret Numbers | Final | `references/core/xls-0025.md` |
| 50 | Validator TOML Infrastructure | Final | `references/core/xls-0050.md` |
| 95 | Rename rippled → xrpld | Draft | `references/core/xls-0095.md` |
| 1  | XLS Process and Guidelines | Living | `references/core/xls-0001.md` |

## ecosystem — wallets, URIs, icons

| XLS | Title | Status | File |
|-----|-------|--------|------|
| 2  | Destination Information | Stagnant | `references/ecosystem/xls-0002.md` |
| 3  | Deeplink Signed Transactions | Stagnant | `references/ecosystem/xls-0003.md` |
| 4  | Trustline Add URI | Stagnant | `references/ecosystem/xls-0004.md` |
| 6  | Visual Account Icons | Final | `references/ecosystem/xls-0006.md` |

---

## Refreshing Refs

```bash
cd packages/skills-build
pnpm build:xrpl-standards
```
