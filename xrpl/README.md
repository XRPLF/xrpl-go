# xrpl

This is the core package of the SDK. It contains everything needed to construct, sign, and submit transactions to the XRP Ledger, query ledger state, and manage accounts — built on top of the lower-level `address-codec`, `keypairs`, and `binary-codec` packages.

## Package Structure

```
xrpl/
├── transaction/        # All transaction types and shared transaction logic
├── wallet/             # Wallet creation, derivation, and offline signing
├── rpc/                # Synchronous JSON-RPC client
├── websocket/          # Asynchronous WebSocket client
├── queries/            # Request/response types for all rippled API methods
├── ledger-entry-types/ # Structs for ledger objects (Offer, AccountRoot, etc.)
├── currency/           # Currency amount utilities
├── hash/               # Transaction hash utilities
├── common/             # Shared constants and helpers
├── flag/               # Transaction flag definitions
├── time/               # XRPL epoch time utilities
├── multisign.go        # Multi-signature aggregation utility
└── interfaces/         # Shared interfaces (CryptoImplementation, etc.)
```

---

## transaction/

All XRPL transaction types live here, one file per type. Every transaction embeds `BaseTx` and implements the `Tx` interface:

```go
type Tx interface {
    TxType() TxType
}
```

`BaseTx` contains the fields common to every transaction:

| Field | Description |
|---|---|
| `Account` | Sender's classic address |
| `TransactionType` | e.g. `Payment`, `OfferCreate`, `TrustSet` |
| `Fee` | Cost in drops (must be set before signing) |
| `Sequence` | Account sequence number (0 if using a Ticket) |
| `LastLedgerSequence` | Expiry ledger — always set to avoid stuck transactions |
| `SigningPubKey` | Public key of the signer (set by `wallet.Sign`) |
| `TxnSignature` | Signature (set by `wallet.Sign`) |
| `Signers` | Multi-signature entries |
| `Memos` | Arbitrary attached data |
| `Flags` | Bitmask of transaction-specific flags |

Each transaction type has a `Flatten() FlatTransaction` method that converts the struct to a `map[string]interface{}` for JSON-RPC submission.

Available transaction types include: `Payment`, `AccountSet`, `AccountDelete`, `TrustSet`, `OfferCreate`, `OfferCancel`, `EscrowCreate/Finish/Cancel`, `PaymentChannelCreate/Fund/Claim`, `NFTokenMint/Burn/CreateOffer/CancelOffer/AcceptOffer`, `AMMCreate/Deposit/Withdraw/Vote/Bid/Delete`, `CheckCreate/Cash/Cancel`, `TicketCreate`, `SignerListSet`, `SetRegularKey`, `DepositPreauth`, `DIDSet/Delete`, `OracleSet/Delete`, `XChain*`, `Batch`, and more.

---

## wallet/

Provides the `Wallet` struct for key management and offline signing.

```go
type Wallet struct {
    PublicKey      string
    PrivateKey     string
    ClassicAddress types.Address
    Seed           string
}
```

### Creation

```go
// Random wallet (ED25519 or SECP256K1)
w, err := wallet.New(crypto.ED25519())

// From an existing seed
w, err := wallet.FromSeed(seed, "")
w, err := wallet.FromSecret(seed) // alias

// From a BIP-39 mnemonic (derives via m/44'/144'/0'/0/0)
w, err := wallet.FromMnemonic("word1 word2 ...")
```

### Signing

```go
// Single signature — sets SigningPubKey and TxnSignature on the tx map
txBlob, txHash, err := w.Sign(flatTx)

// Multi-signature — sets SigningPubKey = "" and appends a Signer entry
txBlob, txHash, err := w.Multisign(flatTx)
```

`Sign` internally calls `binarycodec.EncodeForSigning` to get the signing payload, signs it with `keypairs.Sign`, then calls `binarycodec.Encode` to produce the final blob.

---

## rpc/

A synchronous HTTP JSON-RPC client. Best for one-off queries and simple transaction submission.

```go
cfg := rpc.NewConfig("https://s.altnet.rippletest.net:51234")
client := rpc.NewClient(cfg)

// Submit a pre-signed blob
resp, err := client.SubmitTxBlob(txBlob, false)

// Submit and wait for ledger confirmation
txResp, err := client.SubmitTxBlobAndWait(txBlob, false)
```

The client automatically retries on HTTP 503 (up to 3 times with exponential backoff) and validates that submitted blobs contain a signature before sending.

---

## websocket/

An asynchronous WebSocket client. Best for subscriptions, real-time monitoring, and applications that need to react to ledger events.

```go
cfg := websocket.NewConfig("wss://s.altnet.rippletest.net:51233")
client, err := websocket.NewClient(cfg)

// Subscribe to ledger close events
err = client.SubscribeLedger()

// Read events from channels
ledger := <-client.GetLedgerClosedChannel()
tx     := <-client.GetTransactionChannel()
```

The WebSocket client manages connection lifecycle, request/response correlation by ID, and exposes typed channels for each stream type (ledger, transaction, validation, etc.).

---

## queries/

Typed request and response structs for every rippled API method, organized by category:

| Subdirectory | Methods |
|---|---|
| `account/` | `account_info`, `account_lines`, `account_offers`, `account_tx`, etc. |
| `ledger/` | `ledger`, `ledger_closed`, `ledger_current`, `ledger_data`, `ledger_entry` |
| `transactions/` | `submit`, `submit_multisigned`, `tx`, `transaction_entry` |
| `server/` | `server_info`, `server_state`, `fee` |
| `subscription/` | Stream types for ledger, transaction, and validation subscriptions |

---

## multisign.go

Top-level utility for combining multiple individual multi-signature blobs into a single transaction ready for submission:

```go
// Each blob must be produced by wallet.Multisign
finalBlob, err := xrpl.Multisign(blob1, blob2, blob3)
```

Signers are sorted by account ID bytes (ascending) as required by the XRPL protocol before the final blob is encoded.
