# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

#### xrpl

- `EncodeMPTokenMetadata`, `DecodeMPTokenMetadata` and `ValidateMPTokenMetadata` utils to encode, decode and validate MPTokenMetadata as per XLS-89 standard.
- `AuthorizeChannel` to authorize a payment channel.

### Fixed

#### xrpl

- `rpc` client timeout fetched from config.

### Refactored

#### xrpl

- `TxResponse` `Meta` field type changed to `TxMetadataBuilder`, enabling custom parsing for specific transactions metadata such as `Payment`, `NFTokenMint`, etc.

## [v0.1.13]

### Added

#### binary-codec

- `Number` and `AssetScale` fields to `definitions.json`.

#### xrpl

- `PermissionedDEX` support (XLS-81d).

### Fixed

#### xrpl

- `OracleSet` transaction to Flatten correctly and `Oracle` PriceDataSeries array.

#### binary-codec

- `definitions.json` where `LastUpdatedTime` had a typo issue.

### Refactored

#### xrpl

- Replaced `bip32` and `bip39` dependencies due to repository deletion and, therefore, dependency outdated.

## [v0.1.12]

### Added

#### xrpl

- Adds `PermissionedDomain` ledger entry type (XLS-80d).
- Adds `TokenEscrow` support (XLS-85).

### Fixed

- Flatten function in Escrow transaction types for Destination and Owner fields.

## [v0.1.11]

### BREAKING CHANGES

#### xrpl

- Moved `Signers` type from `github.com/Peersyst/xrpl-go/xrpl/transaction` package to `github.com/Peersyst/xrpl-go/xrpl/transaction/types`.

### Added

#### binary-codec

- Added `MPToken` definitions.
- Added `Hash192` type.
- Added functions to serialize and deserialize `MPTCurrencyAmount`.
- Added `GranularPermissions` and `DelegatablePermissions` entries to definitions.
- Added `PermissionValue` serialized type with custom serializer routing.
- Added`EncodeForSigningBatch` function.

#### xrpl

- Added `AMMClawback` transaction type.
- Added `MPTokenAuthorize`, `MPTokenIssuanceCreate`, `MPTokenIssuanceDestroy`, `MPTokenIssuanceSet` transactions. It also adds the `types.Holder`, `types.AssetScale`, `types.MPTokenMetadata` and `types.TransferFee` types to represent the holder of the token, the asset scale, the metadata and the transfer fee of the token respectively.
- Added `NFTokenMintOffer` support by adding `Amount`, `Expiration`, and `Destination` fields to `NFTokenMint` transaction. Also add `NFTokenMintMetadata` struct to handle transaction metadata with `nftoken_id` and `offer_id` fields.
- Added `MPTCurrencyAmount` for currency kinds.
- Added unit tests for `MPTCurrencyAmount`.
- Added `NFTokenModify` transaction type.

##### Account Permission Delegation (XLS-74d, XLS-75d)

- Added `DelegateSet` transaction type (XLS-74d) with validation and error support.
- Added `Delegate` ledger entry type (XLS-74d).
- Added `PermissionValue` and `Permission` types for delegated permissions.
- Added integration tests for `DelegateSet` submission and delegated `Payment` execution (XLS-75d).

##### Batch (XLS-56d)

- Added `Batch` transaction type.
- Added `CombineBatchSigners` function to combine the batch signers of a set of transactions into a single transaction.
- Added `SignMultiBatch` function to sign a multi-account Batch transaction.
- Added `TfInnerBatchTxn` flag.

## Changed

### binary-codec

- Refactored `Issue` codec type to support `Currency` and `Issuer` fields.

### Dependencies

- Bumped Go version to 1.23.0.

## Fixed

### xrpl

- Fixed some flatten fields with the `Flatten` function for `NFTokenMint`, `NFTokenCancel`, `NFTokenCreate`, `NFTokenBurn`

## [v0.1.10]

### BREAKING CHANGES

#### xrpl

- `Submit` client method is renamed to `SubmitTxBlob` in both clients.
- `SubmitAndWait` client method is renamed to `SubmitTxBlobAndWait` in both clients.

### Added

#### xrpl

- Added `SubmitTx` and `SubmitTxAndWait` client methods to both clients.
- Added support for the Credential fields in the following transaction types:
  - Payment
  - DepositPreauth
  - AccountDelete
  - PaymentChannelClaim
  - EscrowFinish
- Added the `credential` ledger entry for the `account_objects` request.
- Added tec/tef/tel/tem/ter TxResult codes.
- Added `XLS-80d` support with `PermissionedDomain` transaction types:
  - `PermissionedDomainSet`
  - `PermissionedDomainDelete`

### Fixed

#### binary-codec

- Added native `uint8` type support for `Uint8` type.

#### big-decimal

- Fixed `BigDecimal` precision.

## [v0.1.9]

### Added

#### xrpl

- Added support for all the Credential transaction types:
  - CredentialCreate
  - CredentialAccept
  - CredentialDelete

### Fixed

#### big-decimal

- Amounts transcoding fix for large values.

## [v0.1.8]

### Added

#### xrpl

- Added `BalanceChanges` to the `Transaction` type.

### Changed

#### xrpl

- Updated `AffectedNode` type fields to be a pointer to allow nil values.
- Fixed `BaseLedger` field in `ledger` response (v1 and v2). BaseLedger.Transactions is now an array of interfaces instead of a slice of `FlatTransaction` due to `Expand` field in the request.

## [v0.1.7]

### Added

#### xrpl

- Added support for websocket client subscriptions. Now you can subscribe to streams like `ledgerClosed`, `transaction`, `consensus`, `peerStatusChange`, `validationReceived`, etc.

## [v0.1.6]

### Added

#### xrpl

- Configurable timeout for the RPC client. New default timeout of 5 seconds instead of 1 second.

### Fixed

#### xrpl

- Updates some fields in AccountSet and Payment related transactions to a pointer to allow 0 or "" values. For example:

  - `DestinationTag`
  - `TickSize`
  - `Domain`
  - `WalletLocator`
  - `WalletSize`
  - `TransferRate`

- Adds more tests for setting some `asf` flags in `AccountSet`.
- Fixed `Transaction` field in `account_tx` response.
- Fixed `Ledger` field in `ledger` response. LedgerIndex is now an uint32 instead of a string.

## [v0.1.5]

### Added

#### xrpl

Support for the XLS-77d (deep freeze)

## [v0.1.4]

### Added

#### xrpl

- Added `GatewayBalances` and `GetAggregatePrice` queries.

### Fixed

#### xrpl

- Updated SignerQuorum in SignerListSet to be an interface{} with uint32 type assertion instead of a value (uint32).
  - This allows distinguishing between an unset (nil) and an explicitly set value, including 0 to delete a signer list.
  - Ensures SignerQuorum is only included in the Flatten() output when explicitly defined.
  - Updates the `Validate` method to make sure `SignerEntries` is not set when `SignerQuorum` is set to 0

## [v0.1.3]

### Added

- Added `APIVersion` field to the `Client` struct.
- Added `RippledAPIV1` and `RippledAPIV2` constants.
- Added missing `ctid` field on `TxRequest` v1 query.
- Added missing `NoRippleCheck` query (v1 & v2 support).

### Changed

- RippledAPIV2 is set as default API version. Queries and transactions are now compatible with Rippled v2 by default. V1 is still supported. In order to use v1, you need to use the `v1` package of each query type.

## [v0.1.2]

### Fixed

#### xrpl

- The `InfoRequest` for the `account_info` method had an incorrect field `signer_list` (an `s` was missing). The correct field is now `signer_lists`.  
  Link to the documentation [here](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/account-methods/account_info#request-format).

## [v0.1.1]

### Added

#### address-codec

- New `ErrInvalidAddressFormat` error.

### Fixed

#### binary-codec

- Fixed `AccountID` X-Address decoding/encoding support.

#### xrpl

- Replace `IsValidClassicAddress` with `IsValidAddress` on transactions `Validate` methods:
  - `AccountDelete`
  - `AMMBid`
  - `DepositPreauth`
  - `EscrowCancel`
  - `EscrowFinish`
  - `EscrowCancel`
  - `NFTokenBurn`
  - `NFTokenCreateOffer`
  - `NFTokenMint`
  - `NFTokenOffer`
  - `Payment`
  - `PaymentChannelCreate`
  - `SetRegularKey`
  - `SignerListSet`
  - `BaseTx`
  - `XChainBridge`
  - `XChainAccountCreateCommit`
  - `XChainAddAccountCreateAttestation`
  - `XChainAddClaimAttestation`
  - `XChainClaim`
  - `XChainCreateClaimID`
- Master address derivation on wallet `FromSeed` function.
- `NetworkID` field on `BaseTx` type.

## [v0.1.0]

### Added

#### binary-codec

- Updated `definitions`.
- New `DecodeLedgerData` function.
- `Quality` encoding/decoding functions.
- New `XChainBridge` and `Issue` types.

#### address-codec

- Address validation with `IsValidAddress`, `IsValidClassicAddress` and `IsValidXAddress`.
- Address conversion with `XAddressToClassicAddress` and `ClassicAddressToXAddress`.
- X-Address encoding/decoding with `EncodeXAddress` and `DecodeXAddress`.

#### keypairs

- New `DeriveNodeAddress` function.

#### xrpl

- New `AccountRoot`, `Amendments`, `Bridge`, `DID`, `DirectoryNode`, `Oracle`, `RippleState`, `XChainOwnedClaimID`, `XChainOwnedCreateAccountClaimID` ledger entry types.
- New `Multisign` utility function.
- New `NftHistory`, `NftsByIssuer`, `LedgerData`, `Check`, `BookOffers`, `PathFind`, `FeatureOne`, `FeatureAll` queries.
- New `SubmitMultisigned` request.
- New `AMMBid`, `AMMCreate`, `AMMDelete`, `AMMDeposit`, `AMMVote`, `AMMWithdraw` amm transactions.
- New `CheckCancel`, `CheckCash`, `CheckCreate` check transactions.
- New `DepositPreauth` transaction.
- New `DIDSet` and `DIDDelete` transactions.
- New `EscrowCreate`, `EscrowFinish`, `EscrowCancel` escrow transactions.
- New `OracleSet` and `OracleDelete` oracle transactions.
- New `XChainAccountCreateCommitment`, `XChainAddAccountCreateAttestation`, `XChainAddClaimAttestation`, `XChainClaim`, `XChainCommit`, `XChainCreateBridge`, `XChainCreateClaimID` and `XChainModifyBridge` cross-chain transactions.
- New `Multisign` wallet method.
- Ripple time conversion utility functions.
- Added query methods for websocket and rpc clients.
- New `SubmitMultisigned`, `AutofillMultisigned` and `SubmitTxBlobAndWait` methods for both clients.
- Added `Autofill` method for rpc client.
- New `MaxRetries` and `RetryDelay` config options for both clients.

#### Other

- Implemented `secp256k1` algorithm.

### Changed

#### binary-codec

- Exported `FieldInstance` type.
- Updated `NewBinaryParser` constructor to accept `definitions.Definitions` as a parameter.
- Updated `NewSerializer` to `NewBinarySerializer` constructor.
- Refactored `FieldIDCodec` to be a struct with `Encode` and `Decode` methods.
- `FromJson` methods to `FromJSON`.
- `ToJson` methods to `ToJSON`.

#### address-codec

No changes were made.

#### keypairs

- Decoupled `ed25519` and `secp256k1` algorithms from `keypairs` package.
- Decoupled `der` parsing from `keypairs` package.

#### xrpl

- Renamed `CurrencyStringToHex` to `ConvertStringToHex` and `CurrencyHexToString` to `ConvertHexToString`.
- Renamed `HashSignedTx` to `TxBlob`.
- Wallet API methods have been renamed for better usability.
- Renamed `SendRequest` to `Request` methods for websocket and rpc clients.

### Fixed

#### xrpl

- Some queries did not have proper fields. All queries have been updated with the fields that are required by the XRP Ledger.
- Some transaction types did not have proper fields. All transaction types have been updated with the fields that are required by the XRP Ledger.
