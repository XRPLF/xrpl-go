# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Essential Commands

### Linting
```bash
make lint           # Run golangci-lint (auto-installs if missing)
make lint-fix       # Run gofmt to fix formatting issues
```

### Testing
```bash
# Unit tests
make test-ci        # Run all unit tests (CI mode, clean cache, parallel)
make test-all       # Run all unit tests (standard mode)

# Package-specific tests
make test-binary-codec   # Test binary-codec package
make test-address-codec  # Test address-codec package
make test-keypairs       # Test keypairs package
make test-xrpl           # Test xrpl package

# Integration tests (require running network)
make run-localnet-linux/amd64       # Start local rippled node (amd64)
make run-localnet-linux/arm64       # Start local rippled node (arm64)
make test-integration-localnet      # Run integration tests on localnet
make test-integration-devnet        # Run integration tests on devnet
make test-integration-testnet       # Run integration tests on testnet

# Coverage and benchmarks
make coverage-unit   # Generate coverage report (coverage.html)
make benchmark       # Run benchmarks
```

### Running Single Tests
```bash
# Run specific test file
go test ./xrpl/transaction/payment_test.go

# Run specific test function
go test -run TestPayment ./xrpl/transaction

# Run with verbose output
go test -v ./xrpl/transaction

# Run integration tests (set INTEGRATION env var)
INTEGRATION=localnet go test ./xrpl/transaction/integration -v
```

## Architecture Overview

### Package Structure

The repository follows a modular architecture with four primary layers:

1. **Low-level cryptographic primitives** (`address-codec/`, `keypairs/`, `binary-codec/`)
   - `address-codec/`: Encodes/decodes XRPL addresses using Base58Check
   - `keypairs/`: Cryptographic key generation and management (secp256k1, ed25519)
   - `binary-codec/`: Binary serialization/deserialization of XRPL objects
     - `definitions/`: Field type definitions from XRPL protocol
     - `serdes/`: Serialization/deserialization logic
     - `types/`: Type implementations for each XRPL field type

2. **Core XRPL functionality** (`xrpl/`)
   - `transaction/`: All transaction types and transaction handling
     - Each transaction type has its own file (e.g., `payment.go`, `account_set.go`)
     - `BaseTx` struct contains common fields for all transactions
     - All transactions implement `Tx` interface with `TxType()` method
     - `integration/` subdirectory contains integration tests
   - `rpc/`: JSON-RPC client for synchronous requests to rippled nodes
   - `websocket/`: WebSocket client for real-time subscriptions and async requests
   - `wallet/`: Wallet creation, derivation from seed/mnemonic, and offline signing
   - `queries/`: Query request/response types for ledger data
   - `ledger-entry-types/`: Types for different ledger objects
   - `hash/`, `time/`, `currency/`, `common/`: Utilities for XRPL-specific data types

3. **Internal utilities** (`pkg/`)
   - Shared utilities that don't depend on XRPL-specific logic
   - `crypto/`, `big-decimal/`, `map_utils/`, `random/`, `typecheck/`

4. **Examples and documentation** (`examples/`, `docs/`)
   - `examples/` contains working examples for each major feature
   - Each example is self-contained in its own directory

### Key Design Patterns

#### Interfaces and Abstractions
- Each major package has an `interfaces/` subdirectory defining contracts
- Enables dependency injection and testing with mocks
- See `xrpl/interfaces/`, `keypairs/interfaces/`, etc.

#### Test Organization
- `testutil/` directories provide test helpers and fixtures for each package
- Unit tests exclude: `faucet/`, `examples/`, `testutil/`, `interfaces/`
- Integration tests in `xrpl/transaction/integration/` require live network
- Use `golang/mock` for generating mocks

#### Type Safety
- Transaction types use specific Go types from `xrpl/transaction/types/`
- Currency amounts distinguish between XRP (drops) and issued currencies
- Address validation through `addresscodec` package

### Client Usage Pattern

Two primary client types for interacting with XRPL:

1. **RPC Client** (`xrpl/rpc/`): Synchronous JSON-RPC requests
   - Used for one-off queries and transaction submission
   - Methods in `queries.go` map to rippled API methods

2. **WebSocket Client** (`xrpl/websocket/`): Asynchronous connections
   - Used for subscriptions and real-time data
   - Supports streaming ledger updates, transaction monitoring
   - Methods in `queries.go` and subscription handlers in `subscription.go`

Both clients share similar query interfaces but differ in connection management.

### Transaction Lifecycle

1. **Create**: Construct transaction struct (e.g., `Payment`, `AccountSet`)
2. **Autofill**: Set `Fee`, `Sequence`, `LastLedgerSequence` (via RPC/WebSocket client helpers)
3. **Sign**: Use `wallet.Sign()` to add signature
4. **Encode**: Serialize with `binarycodec.Encode()` for submission
5. **Submit**: Send via `rpc.Submit()` or `websocket.Submit()`
6. **Monitor**: Wait for validation and check result

See `examples/send-xrp/` or `examples/send-payment/` for complete workflows.

## Important Development Notes

### Test Exclusions
The Makefile excludes certain packages from standard test runs:
- `faucet/` - Interacts with external testnet faucets
- `examples/` - Standalone example code, not library tests
- `testutil/` - Test helpers, not tests themselves
- `interfaces/` - Interface definitions only

### Integration Tests
Integration tests require `INTEGRATION` environment variable:
- `localnet`: Requires local rippled node (Docker)
- `devnet`: Uses XRPL devnet
- `testnet`: Uses XRPL testnet

Start localnet with `make run-localnet-linux/amd64` before running integration tests.

### Lint Configuration
- Uses `golangci-lint v2.2.2` (configured in Makefile)
- Config in `.golangci.yml` enables: govet, errcheck, staticcheck, gosec, etc.
- Excludes package-comments linter for examples/

### Binary Codec
The binary codec is critical for transaction signing and submission:
- Canonical field ordering defined in `binary-codec/definitions/`
- Type serializers in `binary-codec/types/`
- Always use `binarycodec.Encode()` for creating transaction blobs
- Use `binarycodec.EncodeForSigning()` when preparing transactions for signature

### Common Gotchas
- XRP amounts are always in "drops" (1 XRP = 1,000,000 drops)
- Transaction `Fee` must be set before signing
- `Sequence` numbers must be consecutive (or use Tickets)
- `LastLedgerSequence` prevents transaction from staying in queue indefinitely
- Address encoding differs between classic addresses and X-addresses
