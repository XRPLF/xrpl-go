# XRPL-GO

## Commands

Run the core checks with `GOWORK=off`. A `go.work` file from `make workspace` can change the dependency versions that the root module uses. Root `./...` commands skip `confidential/`.

```bash
GOWORK=off make lint
GOWORK=off make lint-fix
GOWORK=off make test-ci
GOWORK=off go test -run TestPayment_TxType ./xrpl/transaction
```

For changes in `confidential/`, create the workspace first. The native tests need cgo.

```bash
make workspace
make lint-confidential
make test-confidential
make test-confidential-nocgo
```

For integration tests, localnet, and CI scope, read the "Development setup" and "CI scope" sections of [CONTRIBUTING.md](CONTRIBUTING.md).

## Module boundaries

The root module owns protocol types, codecs, clients, and wallet signing. The nested `confidential/` module owns builders, cryptographic helpers, native dependencies, its examples, and its integration tests. Dependencies go from confidential to core only. Keep code that imports confidential inside that module, including tests and examples, so root `go mod tidy` remains independent.

In the root module, a package imports only the layers after it in this order: `xrpl/` -> `keypairs/` and `binary-codec/` -> `address-codec/` -> `pkg/`. The low-level packages never import `xrpl/`.

Before changing release automation, module version requirements, or release notes, read [RELEASING.md](RELEASING.md).

## Protocol parity

xrpld (formerly rippled) is the reference for the protocol. Its source code is in the [XRPLF/rippled](https://github.com/XRPLF/rippled) repository. Use the source code, not only the documentation. The library must be correct compared to xrpld:

- Field definitions, flags, limits, binary encoding, signing payloads, and hashes match xrpld exactly.
- Validation accepts every transaction that can succeed on xrpld, unless a client policy that the library documents forbids it. Example: a tagged X-address in a field that has no tag field.

[xrpl.js](https://github.com/XRPLF/xrpl.js) and [xrpl-py](https://github.com/XRPLF/xrpl-py) are the most used client libraries, and they are the reference for client behavior. Before you implement a feature, read how they do it and keep parity with their behavior, so that users get the same experience in all client libraries. Match the behavior, not the API shape, and use Go idioms for the API.

When you implement or change a rule, and xrpld, xrpl.js, xrpl-py, and xrpl-go do not agree on it, the user chooses the behavior. Show the user a parity table with one row for each of the four. For each row, give the behavior and the location of the code. Then ask the user which behavior to implement.

Checks that prevent real problems for the user are welcome. In xrpld, the stateless rules are in the preflight stage of each transactor (`preflight`, `getFlagsMask`, and related functions). The rules that read the ledger are in `preclaim` and `doApply`.

- Stateless check: add it when it stops a real problem before the user signs or submits the transaction, for example a value that xrpld always rejects in preflight. When the rule is clear and the same in all cases, add the check without the parity table, even when xrpl.js and xrpl-py do not have it. When the rule is not clear, for example when it depends on an amendment, show the parity table and ask.
- Stateful check: it reads the ledger, so it costs requests to the server, and xrpld does the same check when it processes the transaction. Add it when its value to the user is more than its cost. Many ledger requests that only repeat an xrpld check cost more than they give. When xrpl.js or xrpl-py does the check, it can be a good idea to add it too, for parity.

## Tests

- Use table-driven tests when all cases use the same setup, action, and assertions, and only inputs or expected results change. When a new case fits an existing table, add it to that table.
- Use separate focused tests when setup or assertions differ. Keep per-case assertion callbacks, conditional checks, and one-off setup hooks out of tables.
- Use `golang/mock` for mocks.
- Integration tests are in `xrpl/transaction/integration/` and `confidential/integration/`. They skip unless `INTEGRATION` is `localnet`, `devnet`, or `testnet`, so a run without it tests nothing.

## Changelog

Before editing either module's changelog, compare the final branch with its base branch. Entries must describe the net effect of that final diff, not intermediate changes made during development. Follow the [contributor guidance](CONTRIBUTING.md#pull-requests) for section placement and omitting empty `[Unreleased]` sections.

- Add or update an entry only when the final branch has a changelog-worthy difference from the base branch
- If a feature changes and then returns to its base-branch behavior, remove its entry or add no entry
- If later work changes the effect of an existing entry, edit that entry to describe the final result instead of adding an entry for the intermediate state

Follow the existing format: group entries under `### BREAKING CHANGES`, `### Added`, `### Changed`, or `### Fixed`, with a `#### <package>` subheading. Use `### BREAKING CHANGES` for source-incompatible API changes and list it first. Keep entries concise but descriptive enough that users understand the impact. Use `confidential/CHANGELOG.md` for changes to the optional module and the root changelog for core changes.
