# Changelog

Changes to the optional `github.com/Peersyst/xrpl-go/confidential` module are recorded here. Changes before the module split are recorded in the repository's root changelog.

The first independent release is planned as `v0.1.0`, tagged `confidential/v0.1.0`. Core releases and confidential releases have independent versions.

## [Unreleased]

### Added

#### confidential

- Added an independent Go module for the existing builders, encryption, commitments, proofs, and cgo bindings. It requires core `github.com/Peersyst/xrpl-go v0.3.1` or later and keeps all existing package import paths.
- Included the native headers and static libraries for Linux and macOS on amd64 and arm64 in this optional module, together with its examples and integration tests.

#### confidential/builder

- Added `GetSpendingBalance()`, which reads a holder's `ConfidentialBalanceSpending` and decrypts it with that holder's ElGamal private key. It takes the same `LedgerQuerier` the builders do, so `rpc.Client` and `websocket.Client` share one reader. Both reads come from one validated ledger, no account sequence is queried, the unspendable `ConfidentialBalanceInbox` is excluded, a missing `MPToken` reports `ErrMPTokenNotFound`, and an `MPToken` with no spending ciphertext reads as zero without decrypting. The search is bounded by the caller's `BalanceRange`, capped at the issuance `ConfidentialOutstandingAmount` as `BuildClawback` already does.
