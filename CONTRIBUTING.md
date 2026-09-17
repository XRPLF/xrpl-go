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

- Go `1.25.13` or later, matching `go.mod`
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

### Lint and test core

The repository contains a core module at the root and an optional module in `confidential/`. Root `./...` commands do not cross into the nested module.

```bash
GOWORK=off make lint
GOWORK=off make test-ci
```

### Work on confidential helpers

Use a local workspace to test the checked-out core and confidential modules together:

```bash
make workspace
make lint-confidential
make test-confidential
make test-confidential-nocgo
```

The workspace and its version-specific core replacement are ignored by Git. They allow development before the required core version is published. Keep replacements out of both published `go.mod` files.

Native tests need a C/C++ toolchain and cgo on Linux or macOS, with amd64 or arm64. Linux also needs the zlib development library. `test-confidential-nocgo` builds all optional packages without cgo and tests the unavailable-backend contract.

Confidential examples live in `confidential/examples/`, and their integration tests live in `confidential/integration/`. Keep code that imports the optional helpers inside this module, including tests and examples. Otherwise, root `go mod tidy` can add the optional module as a core dependency.

### Focused tests

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
make workspace
make run-localnet-linux/amd64
make test-integration-localnet
```

Use `make run-localnet-linux/arm64` on arm64 machines.

Localnet targets run both modules by default. Set `INTEGRATION_MODULE=core` or `INTEGRATION_MODULE=confidential` to run only one module, including with `make integration-localnet`. Devnet and testnet core targets exclude the slower confidential scenarios. To run only the confidential scenarios:

```bash
make test-integration-confidential-localnet
make test-integration-confidential-devnet
```

## CI scope

[Core CI](.github/workflows/ci-core.yml) and [Confidential CI](.github/workflows/ci-confidential.yml) use GitHub's native path filters. Each workflow runs its module's tests, lint, vulnerability scan, API check, and localnet suite. Confidential also runs its four-platform native test matrix.

- Core changes run core checks only.
- Confidential changes run confidential checks only. Core is built as a dependency, but its tests do not run.
- Shared build inputs, such as `Makefile`, `.golangci.yml`, or localnet configuration, run both.
- Documentation-only changes skip Go CI. Markdown under `testdata/` is treated as a test fixture.

Each workflow has a manual trigger. Run **Confidential CI** manually to check a core change against the optional helpers. Weekly vulnerability scans still check both modules.

There are no custom change-detection scripts or aggregate status jobs. Do not require these path-filtered checks on every PR: GitHub can leave skipped workflows pending. Update any existing branch-protection rules that require the old checks. Native path filters also have GitHub's changed-file limits, so run the relevant workflows manually for large changes that GitHub skips.

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
6. Update the affected module's changelog: `CHANGELOG.md` for core or `confidential/CHANGELOG.md` for the optional helpers. Use `[Unreleased]` for pending changes, creating the section when needed. During release preparation, update the target version section instead. Omit `[Unreleased]` when it is empty.
7. Run the relevant checks before opening the pull request.

Use conventional commits, for example:

```text
docs: update contributing guide
fix: validate account delete metadata
```

## Releases

See [RELEASING.md](RELEASING.md) for independent versioning, the first-release order, and the GitHub Actions module picker. Release checks use `GOWORK=off` to verify the declared published dependencies rather than the local workspace.

## Code style

- Match the existing package style.
- Keep changes surgical.
- Prefer simple, readable code over new abstractions.
- Add comments only when they explain non-obvious behavior.

## Licensing

By contributing, you agree that your contributions will be licensed under the MIT license.
