# queries

## Overview

The `queries` package contains mainly request and response types for the [XRPL methods](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods). This package is used by the package clients [`rpc`](/docs/xrpl/rpc) and [`websocket`](/docs/xrpl/websocket) to send client queries to the ledger.

:::info

As a developer, you may be interested in calling the queries using the [`websocket`](/docs/xrpl/websocket) or [`rpc`](/docs/xrpl/rpc) clients. Both clients expose methods to call each query exposed by the `queries` package.

:::

Queries are grouped by different categories or packages:

- `account`: Methods to work with account info.
- `channel`: Methods to work with channels.
- `ledger`: Methods to retrieve ledger info.
- `transaction`: Submit and query ledger transactions.
- `path`: Methods to use paths and order books.
- `nft`: Methods to work with NFTs.
- `oracle`: Methods to work with oracles.
- `vault`: Methods to work with vaults.
- `clio`: Methods to use the Clio API, not [`rippled`](https://github.com/XRPLF/rippled).
- `server`: Methods to retrieve information about the current state of the [`rippled`](https://github.com/XRPLF/rippled) server.
- `utility`: Perform convenient tasks, such as ping and random number generation.

### API version

By default, all queries are meant to be used with the latest XRPL API version (currently `v2`). If you want to use a specific version, you will need to import the specific version queries package from each subpackage.

For example, if you want to use the XRPL API version `v1` queries of the `account` subpackage, you will need to import it this way:

```go
import accountv1 "github.com/Peersyst/xrpl-go/xrpl/queries/account/v1"
```

## Categories

### account

The `account` package contains methods to interact with XRPL accounts. These methods allow you to:

- Retrieve account information like balances, settings, and objects.
- Get account transaction history.
- Query account channels and escrows.
- Check account offers and payment channels.

The available methods correspond to the [Account Methods](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods#account-methods) in the XRPL API.

The account subpackage provides the following queries requests:

| Request                  | Method name                                                                                                                      | V1 support | V2 support |
| ------------------------ | -------------------------------------------------------------------------------------------------------------------------------- | ---------- | ---------- |
| `ChannelRequest`         | [account_channels](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/account-methods/account_channels)     | ✅         | ✅         |
| `CurrenciesRequest`      | [account_currencies](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/account-methods/account_currencies) | ✅         | ✅         |
| `GatewayBalancesRequest` | [gateway_balances](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/account-methods/gateway_balances)     | ❌         | ✅         |
| `InfoRequest`            | [account_info](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/account-methods/account_info)             | ✅         | ✅         |
| `LinesRequest`           | [account_lines](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/account-methods/account_lines)           | ✅         | ✅         |
| `NFTsRequest`            | [account_nfts](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/account-methods/account_nfts)             | ✅         | ✅         |
| `NoRippleCheckRequest`   | [noripple_check](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/account-methods/noripple_check)         | ✅         | ✅         |
| `ObjectsRequest`         | [account_objects](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/account-methods/account_objects)       | ✅         | ✅         |
| `OffersRequest`          | [account_offers](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/account-methods/account_offers)         | ✅         | ✅         |
| `TransactionsRequest`    | [account_tx](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/account-methods/account_tx)                 | ✅         | ✅         |

#### Usage

To use the `account` package, you need to import it in your project:

```go
import "github.com/Peersyst/xrpl-go/xrpl/queries/account"
```

### channel

The `channel` package contains methods to interact with XRPL channels. These methods allow you to:

- Verify the channel's state.

The available methods correspond to the [Payment Channel Methods](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/payment-channel-methods) in the XRPL API.

The `channel` subpackage provides the following queries requests:

| Request         | Method name                                                                                                                      | V1 support | V2 support |
| --------------- | -------------------------------------------------------------------------------------------------------------------------------- | ---------- | ---------- |
| `VerifyRequest` | [channel_verify](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/payment-channel-methods/channel_verify) | ✅         | ✅         |

#### Usage

To use the `channel` package, you need to import it in your project:

```go
import "github.com/Peersyst/xrpl-go/xrpl/queries/channel"
```

### ledger

The `ledger` package contains methods to interact with XRPL ledgers. These methods allow you to:

- Retrieve specific, current or closed ledger information.

The available methods correspond to the [Ledger Methods](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/ledger-methods) in the XRPL API.

The `ledger` subpackage provides the following queries requests:

| Request          | Method name                                                                                                             | V1 support | V2 support |
| ---------------- | ----------------------------------------------------------------------------------------------------------------------- | ---------- | ---------- |
| `Request`        | [ledger](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/ledger-methods/ledger)                 | ✅         | ✅         |
| `ClosedRequest`  | [ledger_closed](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/ledger-methods/ledger_closed)   | ✅         | ✅         |
| `CurrentRequest` | [ledger_current](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/ledger-methods/ledger_current) | ✅         | ✅         |
| `DataRequest`    | [ledger_data](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/ledger-methods/ledger_data)       | ✅         | ✅         |

#### Usage

To use the `ledger` package, you need to import it in your project:

```go
import "github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
```

### transaction

The `transaction` package contains methods to interact with XRPL transactions. These methods allow you to:

- Submit ledger transactions.
- Query ledger transactions.

The available methods correspond to the [Transaction Methods](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/transaction-methods) in the XRPL API.

The `transaction` subpackage provides the following queries requests:

| Request                    | Method name                                                                                                                          | V1 support | V1 support |
| -------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ | ---------- | ---------- |
| `SubmitRequest`            | [submit](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/transaction-methods/submit)                         | ✅         | ✅         |
| `SubmitMultisignedRequest` | [submit_multisigned](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/transaction-methods/submit_multisigned) | ✅         | ✅         |
| `EntryRequest`             | [transaction_entry](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/transaction-methods/transaction_entry)   | ✅         | ✅         |
| `TxRequest`                | [tx](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/transaction-methods/tx)                                 | ✅         | ✅         |

#### Usage

To use the `transaction` package, you need to import it in your project:

```go
import "github.com/Peersyst/xrpl-go/xrpl/queries/transaction"
```

### path, nft and oracle

The `path`, `nft` and `oracle` packages contain methods to interact with XRPL paths, NFTs and oracles. These methods allow you to:

- Retrieve paths and order books.
- Get NFTs buy and sell offers.

The available methods correspond to the [Path and Order Book Methods](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/path-and-order-book-methods) in the XRPL API.

The `path` subpackage provides the following queries requests:

| Request                                                      | Method name                                                                                                                                  | V1 support | V2 support |
| ------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------- | ---------- | ---------- |
| `BookOffersRequest`                                          | [book_offers](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/path-and-order-book-methods/book_offers)               | ✅         | ✅         |
| `DepositAuthorizedRequest`                                   | [deposit_authorized](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/path-and-order-book-methods/deposit_authorized) | ✅         | ✅         |
| `FindCreateRequest`, `FindCloseRequest`, `FindStatusRequest` | [path_find](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/path-and-order-book-methods/path_find)                   | ✅         | ✅         |
| `RipplePathFindRequest`                                      | [ripple_path_find](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/path-and-order-book-methods/ripple_path_find)     | ✅         | ✅         |

The `nft` subpackage provides the following queries requests:

| Request                    | Method name                                                                                                                            | V1 support | V2 support |
| -------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- | ---------- | ---------- |
| `NFTokenBuyOffersRequest`  | [nft_buy_offers](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/path-and-order-book-methods/nft_buy_offers)   | ✅         | ✅         |
| `NFTokenSellOffersRequest` | [nft_sell_offers](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/path-and-order-book-methods/nft_sell_offers) | ✅         | ✅         |

The `oracle` subpackage provides the following queries requests:

| Request                    | Method name                                                                                                                       | V1 support | V2 support |
| -------------------------- | --------------------------------------------------------------------------------------------------------------------------------- | ---------- | ---------- |
| `GetAggregatePriceRequest` | [get_aggregate_price](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/oracle-methods/get_aggregate_price) | ❌         | ✅         |

#### Usage

To use the `path` package, you need to import it in your project:

```go
import "github.com/Peersyst/xrpl-go/xrpl/queries/path"
```

To use the `oracle` package, you need to import it in your project:

```go
import "github.com/Peersyst/xrpl-go/xrpl/queries/oracle"
```

To use the `nft` package, you need to import it in your project:

```go
import "github.com/Peersyst/xrpl-go/xrpl/queries/nft"
```

### vault

The `vault` package contains methods to interact with XRPL vaults. These methods allow you to:

- Retrieve vault information by vault ID or owner and sequence number.

The `vault` subpackage provides the following queries requests:

| Request       | Method name                                                                                                                  | V1 support | V2 support |
| ------------- | ---------------------------------------------------------------------------------------------------------------------------- | ---------- | ---------- |
| `InfoRequest` | [vault_info](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/vault-methods/vault_info)               | ❌         | ✅         |

#### Usage

To use the `vault` package, you need to import it in your project:

```go
import "github.com/Peersyst/xrpl-go/xrpl/queries/vault"
```

### clio

The `clio` package contains methods to interact with the Clio API, not [`rippled`](https://github.com/XRPLF/rippled). These methods allow you to:

- Retrieve NFT history.
- Retrieve NFts information.

The available methods correspond to the [Clio Methods](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/clio-methods) in the XRPL API.

The `clio` subpackage provides the following queries requests:

| Request               | Method name                                                                                                           | V1 support | V2 support |
| --------------------- | --------------------------------------------------------------------------------------------------------------------- | ---------- | ---------- |
| `NFTHistoryRequest`   | [nft_history](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/clio-methods/nft_history)       | ✅         | ✅         |
| `NFTInfoRequest`      | [nft_info](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/clio-methods/nft_info)             | ✅         | ✅         |
| `NFTsByIssuerRequest` | [nfts_by_issuer](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/clio-methods/nfts_by_issuer) | ❌         | ✅         |

#### Usage

To use the `clio` package, you need to import it in your project:

```go
import "github.com/Peersyst/xrpl-go/xrpl/queries/clio"
```

### server

The `server` package contains methods to interact with the [`rippled`](https://github.com/XRPLF/rippled) server. These methods allow you to:

- Retrieve server information.
- Get fee information.
- Get the manifest.

The available methods correspond to the [Server Info Methods](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/server-info-methods) in the XRPL API.

The `server` subpackage provides the following queries requests:

| Request             | Method name                                                                                                              | V1 support | V2 support |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------ | ---------- | ---------- |
| `FeatureAllRequest` | [feature_all](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/server-info-methods/feature_all)   | ✅         | ✅         |
| `FeatureOneRequest` | [feature](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/server-info-methods/feature)           | ✅         | ✅         |
| `FeeRequest`        | [fee](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/server-info-methods/fee)                   | ✅         | ✅         |
| `ManifestRequest`   | [manifest](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/server-info-methods/manifest)         | ✅         | ✅         |
| `InfoRequest`       | [server_info](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/server-info-methods/server_info)   | ✅         | ✅         |
| `StateRequest`      | [server_state](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/server-info-methods/server_state) | ✅         | ✅         |

#### Usage

To use the `server` package, you need to import it in your project:

```go
import "github.com/Peersyst/xrpl-go/xrpl/queries/server"
```

### utility

The `utility` package contains methods to interact with the XRPL utility. These methods allow you to:

- Retrieve a random number.
- Ping the server.

The available methods correspond to the [Utility Methods](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/utility-methods) in the XRPL API.

The `utility` subpackage provides the following queries requests:

| Request         | Method name                                                                                              | V1 support | V2 support |
| --------------- | -------------------------------------------------------------------------------------------------------- | ---------- | ---------- |
| `RandomRequest` | [random](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/utility-methods/random) | ✅         | ✅         |
| `PingRequest`   | [ping](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/utility-methods/ping)     | ✅         | ✅         |

#### Usage

To use the `utility` package, you need to import it in your project:

```go
import "github.com/Peersyst/xrpl-go/xrpl/queries/utility"
```
