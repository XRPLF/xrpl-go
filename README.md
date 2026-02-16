# XRPL-GO

[![Go Reference](https://pkg.go.dev/badge/github.com/Peersyst/xrpl-go.svg)](https://pkg.go.dev/github.com/Peersyst/xrpl-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/Peersyst/xrpl-go)](https://goreportcard.com/report/github.com/Peersyst/xrpl-go)
[![Release Card](https://img.shields.io/github/v/release/Peersyst/xrpl-go?include_prereleases)](https://github.com/Peersyst/xrpl-go/releases)


The `xrpl-go` library provides a Go implementation for interacting with the XRP Ledger. From serialization to signing transactions, the library allows users to work with the most
complex elements of the XRP Ledger. A full library of models for all transactions and core server rippled API objects are provided.

## Requirements

Requiring Go version `1.22.0` and later.
[Download latest Go version](https://go.dev/dl/)

## Packages

| Name | Description |
|---------|-------------|
| addresscodec | Provides functions for encoding and decoding XRP Ledger addresses |
| binarycodec | Implements binary serialization and deserialization of XRP Ledger objects |
| keypairs | Handles generation and management of cryptographic key pairs for XRP Ledger accounts |
| xrpl | Core package containing the main functionality for interacting with the XRP Ledger |
| examples | Contains example code demonstrating usage of the xrpl-go library |

## Quickstart

This guide covers everything you need to start contributing to XRPL-GO.

### Development Requirements

To work on this project, you'll need:

- **Go compiler** version `1.22.0` or later ([download](https://go.dev/doc/install) or use [gvm](https://github.com/moovweb/gvm))
- **golangci-lint** (see `GOLANGCI_LINT_VERSION` in `Makefile`)
- **make** command-line tool
- **Docker** (for running CI/CD workflows locally)
- **yarn** (for running the documentation site)

A Go debugger is also recommended for improved debugging experience.

### Getting Started

1. **Fork the repository**

   Fork [XRPLF/xrpl-go](https://github.com/XRPLF/xrpl-go) to your GitHub account to work on changes and open pull requests.

2. **Install dependencies**

   Install dependencies globally:
   ```bash
   go mod tidy
   ```

   Or install them locally in a `vendor` directory:
   ```bash
   go mod vendor
   ```

3. **Run the linter**

   ```bash
   make lint
   ```

   > **Note:** If `golangci-lint` is not installed, this command will automatically install it with the same version used in CI/CD workflows.

4. **Run tests**

   Run the full test suite as executed in CI/CD:
   ```bash
   make test-ci
   ```

   Verify that all tests pass. Check the `Makefile` for additional rules and feel free to propose new ones.

## Report an issue

If you find any issues, please report them to the [XRPL-GO GitHub repository](https://github.com/Peersyst/xrpl-go/issues).

## License
The `xrpl-go` library is licensed under the MIT License.
