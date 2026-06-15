# Contributing to xrpl-go

## How to contribute

You can contribute by:

- Reporting bugs
- Suggesting enhancements
- Implementing features
- Writing documentation
- Writing tests

## Reporting bugs

Before opening an issue, check whether it has already been reported. Include:

- A clear title
- Steps to reproduce
- Expected and actual behavior
- Environment details, including Go version and OS
- Relevant logs, screenshots, or transaction data

## Suggesting enhancements

Before opening an enhancement request, check for duplicates. Include the use case, expected behavior, and any compatibility or security considerations.

## Development setup

### Prerequisites

- Go `1.24.3` or later, matching `go.mod`
- `make`
- Docker, for localnet integration tests
- Yarn, for the documentation site

`make lint` installs the pinned `golangci-lint` version from the [`Makefile`](Makefile).

### Clone and install

```bash
git clone https://github.com/XRPLF/xrpl-go
cd xrpl-go
go mod tidy
```

If you need vendored dependencies:

```bash
go mod vendor
```

### Lint and test

```bash
make lint
make test-ci
```

Useful focused checks:

```bash
go test ./xrpl/transaction
go test -run TestPayment ./xrpl/transaction
make test-binary-codec
make test-address-codec
make test-keypairs
make test-xrpl
```

Integration tests require a target network. For localnet:

```bash
make run-localnet-linux/amd64
make test-integration-localnet
```

Use `make run-localnet-linux/arm64` on arm64 machines.

## Documentation

The documentation site lives in [`docs/`](docs/).

```bash
cd docs
yarn
yarn start
yarn build
```

Published docs are hosted at <https://xrplf.github.io/xrpl-go/>.

## Pull requests

1. Fork the repository.
2. Create a branch for your change.
3. Make the smallest focused change that solves the issue.
4. Add or update tests for code changes.
5. Update documentation when behavior or user-facing APIs change.
6. Update `CHANGELOG.md` under `[Unreleased]` for code changes.
7. Run the relevant checks before opening the pull request.

Use conventional commits, for example:

```text
docs: update contributing guide
fix: validate account delete metadata
```

## Code style

- Match the existing package style.
- Keep changes surgical.
- Prefer simple, readable code over new abstractions.
- Add comments only when they explain non-obvious behavior.

## Licensing

By contributing, you agree that your contributions will be licensed under the MIT license.
