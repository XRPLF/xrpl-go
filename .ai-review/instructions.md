# Reviewing xrpl-go

Review changes for transaction correctness, security, and public API compatibility.
Use [AGENTS.md](../AGENTS.md) for project structure, commands, test conventions, and
changelog policy. Apply the checks below to the relevant changed code.

## Protocol evidence

- Verify wire fields, flags, and signing rules against the targeted rippled version
  and amendment. Use the relevant [XRPL Standards reference](../.agents/skills/xrpl-standards/references/INDEX.md)
  for specification context. A sibling SDK or draft specification alone is not proof
  that this implementation is wrong.
- Report a concrete failure path and its effect on callers. Check existing helpers,
  callers, and tests before reporting a missing check. Separate an observed defect
  from a protocol assumption that still needs verification.

## Amounts and Go values

- In `xrpl/transaction/types/currency_amount.go`, `XRPCurrencyAmount` is a `uint64`
  in Go but a quoted decimal string in JSON and `Flatten()`. Preserve both
  representations at their respective boundaries.
- `IssuedCurrencyAmount.Value` and `MPTCurrencyAmount.Value` are strings.
  `MPTAmount` has its own integer range check. Preserve exact values through parsing,
  arithmetic, and encoding. Check rounding, overflow, and underflow before accepting
  a conversion through `float64` or a narrower integer. Use the existing decimal
  helpers in `pkg/big-decimal/` or `math/big` where applicable.
- Check the concrete kind behind the `CurrencyAmount` interface. Preserve the
  distinction between an issued currency and an MPT.
  In particular, `UnmarshalCurrencyAmount` rejects mixed issuer/currency and MPT fields.
- At `map[string]any` boundaries, validate dynamic types and numeric ranges before
  assertions or casts. Preserve supported representations through `pkg/typecheck/`.
  A typed nil in an interface is not a nil interface.

## Transaction models and JSON

- Trace a changed field through the struct, `TxType()`, `Flatten()`, `Validate()`,
  and the binary definitions. `Tx` only requires `TxType()`; satisfying that
  interface does not prove that validation or serialization is complete. Use
  `xrpl/transaction/payment.go` as a representative model.
- `FlatTransaction` is a `map[string]any`. JSON tags alone do not change `Flatten()`.
  Check exact field names, nested wrappers, amount flattening, and flag/field
  dependencies. Preserve the base validation in account-submitted transactions;
  consensus-generated pseudo-transactions have a different contract.
- Distinguish absent, zero, empty, and nil values. Examples: a non-nil
  `Payment.DestinationTag` can carry zero; a ticket requires an explicit zero
  `Sequence`; an inner Batch transaction needs an explicit empty `SigningPubKey`.
  Check `omitempty` and hand-written flattening separately.
- For response model changes, check JSON decoding and both clients' `GetResult`
  paths. RPC and WebSocket use `pkg/decodehook/` plus the text-unmarshal hook.
  Preserve custom unmarshalling, API-version differences, and tolerance of extra
  response fields. Do not impose request validation on server responses.

## Binary encoding

- `binary-codec/definitions/definitions.json` and `STObject` determine field codes,
  types, ordinal order, variable-length encoding, and signing-field selection.
  Check changed metadata against protocol evidence. Go map iteration order must
  not determine the encoded bytes.
- New serialized types need `FromJSON`, `ToJSON`, and registration in
  `GetSerializedType` in `binary-codec/types/serialized_type.go`.
  Preserve parser bounds checks and matching variable-length prefixes.
- For supported canonical bytes, decoding and re-encoding must preserve the bytes.
  Compare JSON after the codec's canonicalization, not raw input spelling.
  Check an independent expected binary fixture as well: a round trip alone can
  hide matching encoder and decoder defects.
- Preserve field-specific JSON formats. `binary-codec/types/uint64.go` uses decimal
  strings for selected MPT amount fields and hexadecimal strings for other UInt64
  fields. Treating every UInt64 field as decimal changes the wire value.
- Preserve the `UNLModify` zero-length `Account` value override in
  `binary-codec/codec.go`; it reproduces rippled's default `STAccount` encoding.
  This is not an ordinary missing-account error.

## Signing and secrets

- Keep the signing payloads in `binary-codec/codec.go` separate: single signing,
  multisigning, payment-channel claims, and Batch signing use different prefixes
  or payloads. Regular multisigning requires an empty `SigningPubKey`; counterparty
  signing in `xrpl/wallet/counterparty_signer.go` preserves the first signer's key.
- Preserve the algorithm distinction in `pkg/crypto/`: secp256k1 signs a SHA-512-half
  digest and uses canonical DER signatures; Ed25519 signs the message bytes without
  that extra prehash. Check key lengths and prefixes before slicing key bytes.
- `Wallet.Sign` and `Wallet.Multisign` preserve the caller's map. Map and slice
  copies are shallow; changes to nested signing data must not mutate caller-owned
  values. Check each other helper's documented mutation contract separately.
- Production entropy must use cryptographically secure randomness, as in
  `pkg/random/`. Keep seeds, private keys, mnemonics, and authorization headers out
  of errors and logs. Distinguish documented public test keys from real secrets.

## Client errors and concurrency

- Compare RPC and WebSocket behavior when changing shared transaction logic in
  `xrpl/internal/client/`. A successful submission response is not final validation.
  Preserve the validated-ledger checks and expiry boundary in `WaitForFinality`;
  cancellation or a transport failure is not proof that a transaction failed.
- Preserve error identity where callers use `errors.Is` or `errors.As`. Check that
  malformed remote data returns an error rather than panicking, and that response
  bodies, retries, and waits remain bounded and cancellable where supported.
- For WebSocket changes, trace pending requests through completion, timeout,
  disconnect, and socket replacement. Check shared maps, channel ownership, and
  goroutine exit paths. Preserve per-stream callback order and the documented
  backpressure behavior in `xrpl/websocket/client.go`.
- `Disconnect` can run inside a stream callback; waiting synchronously for that
  callback can deadlock. Reconnect preserves handler registrations but does not
  replay server subscriptions. Judge changes against these lifecycle contracts.

## Compatibility and tests

- Assess exported API changes against the release baseline used by
  `.github/workflows/ci-api-compatibility.yml`. A new export is not automatically
  a breaking change. Check behavioral and JSON compatibility even when types compile.
  Keep the module import path from `go.mod`; the GitHub organization name is not
  a reason to rewrite imports.
- Flag incorrect assertions and missing regression tests for a specific changed
  behavior after checking existing coverage. Prioritize exact encoding vectors,
  malformed inputs, zero/omitted fields, and signing input immutability. Use focused
  race tests for concurrency changes; avoid general coverage or test-style demands.
- Integration tests need the environment and, for localnet, the build tag set by
  the Makefile targets. A unit-only run does not prove live-ledger behavior.
  Leave formatting and lint rules to `.golangci.yml` and the existing CI checks.
