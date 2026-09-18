# Contributing to xrpl-go

Help improve the SDK with bug reports, fixes, features, tests, or documentation.

[Development setup](#development-setup) · [Tests](#focused-tests) · [Documentation](#documentation) · [Pull requests](#pull-requests)

Run commands from the repository root unless a section says otherwise.

## Reporting bugs

Before opening an issue, check whether it has already been reported. Include:

- A clear title
- Steps to reproduce
- Expected and actual behavior
- Environment details, including Go version and OS
- Relevant logs or transaction data, with secrets removed

Never include seeds, private keys, or mnemonics in an issue.

## Suggesting enhancements

Before opening an enhancement request, check for duplicates. Include the use case, expected behavior, and any compatibility or security considerations.

## Development setup

### Prerequisites

For core development, install **Go 1.25.13 or later** and `make`.

Additional tools depend on your work:

| Work | Requirements |
| --- | --- |
| Confidential cryptography | [Native toolchain](confidential/README.md#installation) and cgo |
| Localnet integration tests | Docker |
| Documentation site | Node.js 20 or later and Yarn |

`make lint` installs the pinned `golangci-lint` version from the [`Makefile`](Makefile).

### Clone and install

```bash
git clone https://github.com/XRPLF/xrpl-go
cd xrpl-go
go mod tidy
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

Run the checks for the packages you change:

```bash
go test ./xrpl/transaction
go test -run TestPayment ./xrpl/transaction
make test-binary-codec
make test-address-codec
make test-keypairs
make test-xrpl
```

### Integration tests

Integration tests require a running ledger. For localnet:

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

- Core changes run core checks only, except for the shared inputs below.
- Confidential changes run confidential checks only. Core is built as a dependency, but its tests do not run.
- Shared build inputs, such as `Makefile`, `.golangci.yml`, or localnet configuration, run both. Changes to the wire-size constants in `pkg/mptsizes/` also run both, so the native bindings are checked against their headers.
- Documentation-only changes skip Go CI. Markdown under `testdata/` is treated as a test fixture.

Each workflow has a manual trigger. Run **Confidential CI** manually to check a core change against the optional helpers. Weekly vulnerability scans still check both modules.

For maintainers: do not require path-filtered checks on every PR. GitHub can leave skipped workflows pending, and there is no aggregate status job. Review branch-protection rules when changing these workflows. GitHub also limits the files evaluated by path filters, so run checks manually when a large change is skipped.

## Documentation

The site lives in [`docs/`](docs/) and is published at <https://xrplf.github.io/xrpl-go/>.

Install dependencies and start a local preview:

```bash
cd docs
yarn install --frozen-lockfile
yarn start
```

Stop the preview before checking the production build, or run this in a second terminal from `docs/`:

```bash
yarn build
```

The [deployment workflow](.github/workflows/docs-deploy.yml) publishes relevant changes pushed to `main`. README and package-guide changes do not require building the site.

## Pull requests

1. Fork the repository.
2. Create a branch for your change.
3. Make the smallest focused change that solves the issue.
4. Add or update tests for code changes.
5. Update documentation when behavior or user-facing APIs change.
6. Add a changelog entry when the final diff has a user-facing change worth recording. Use `CHANGELOG.md` for core or `confidential/CHANGELOG.md` for the optional helpers.
7. Run the relevant checks. In the PR, list the commands you ran and any checks you could not run.

Use `[Unreleased]` for pending changelog entries, or the target version section during release preparation. Describe the final effect relative to the base branch, not intermediate changes. Omit empty `[Unreleased]` sections.

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
