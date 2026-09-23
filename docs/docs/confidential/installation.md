# Install confidential helpers

Install this optional module when you need confidential builders or cryptographic helpers. Normal wallet signing and confidential transaction models remain in [core](/docs/installation).

The versioned commands below use core `v0.3.1` and confidential `v0.1.0`. To work on an unreleased checkout, use the [development workspace](#development-workspace).

## Add confidential helpers

```bash
go get github.com/Peersyst/xrpl-go/confidential@v0.1.0
```

Go also selects the required core module. You do not need a separate core installation command.

Package import paths have not changed:

```go
import (
	"github.com/Peersyst/xrpl-go/confidential/builder"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
)
```

After adding imports, run `go mod tidy` to record the dependencies and checksums used by your application.

### Native build requirements

The helpers use cgo to call `XRPLF/mpt-crypto`. Native operations require:

- Go `1.25.13` or later.
- `CGO_ENABLED=1`.
- Linux or macOS, on `amd64` or `arm64`.
- A C/C++ compiler and linker toolchain. On Linux, the linker also needs the system zlib development library.

For example, Ubuntu uses `build-essential` and `zlib1g-dev`. On macOS, install the Xcode command-line tools.

```bash
CGO_ENABLED=1 go build ./...
```

The module includes headers and static libraries for all four supported targets. Users do not need Conan or a separate `mpt-crypto` installation. `go get` does not install a compiler. Cross-compilation requires a matching target toolchain.

With cgo disabled, or on an unsupported target, the packages still build. Native cryptographic operations return `mptcrypto.ErrCgoRequired`. This fallback does not provide a pure-Go implementation of the cryptography.

### Troubleshoot the build

| Symptom | Check |
| --- | --- |
| `mptcrypto.ErrCgoRequired` | `go env CGO_ENABLED GOOS GOARCH`, supported targets, and build tags |
| Compiler not found | Install the C/C++ toolchain and check `CC`/`CXX` |
| Linux linker cannot find zlib | Install the target's zlib development package |
| Cross-compilation fails | Use a compiler and system libraries for the target, not the host |

`js`, `wasip1`, TinyGo, and go-fuzz builds use the unavailable-backend implementation.

## Versions and updates

The first confidential release requires core `v0.3.1` or later. Go selects one version of each module for the application:

| Existing core requirement | After adding confidential `v0.1.0` |
| --- | --- |
| None | Adds core `v0.3.1` |
| `v0.3.0` | Upgrades core to `v0.3.1` |
| A newer compatible core version | Keeps the newer version |

A minimum dependency is not an exact version lock or a guarantee that every future version is compatible. Check the [core changelog](/changelog/v0.3.x/changelog) and [confidential changelog](/changelog/confidential/v0.1.x/changelog) before upgrading.

Update the modules separately:

```bash
go get github.com/Peersyst/xrpl-go@latest
go get github.com/Peersyst/xrpl-go/confidential@latest
```

A confidential update upgrades core only when its dependency requirements need that upgrade. A core update does not automatically update confidential.

The Go command uses a version without the directory prefix. The Git tag includes the prefix:

| Module | `go get` version | Git tag |
| --- | --- | --- |
| Core | `@v0.3.1` | `v0.3.1` |
| Confidential | `@v0.1.0` | `confidential/v0.1.0` |

### Migration from the combined module

Starting with core `v0.3.1`, the helpers are a separate module, initially versioned `v0.1.0`. Their versions and releases are independent. Core's Go module archive no longer includes native headers and libraries. A repository clone or GitHub source archive still contains both modules. Older downloads and caches are not changed by the split.

Update applications that used confidential packages from `v0.3.1-mpt.0` to core `v0.3.1` and add the confidential module:

```bash
go get github.com/Peersyst/xrpl-go@v0.3.1 github.com/Peersyst/xrpl-go/confidential@v0.1.0
go mod tidy
```

Keep the same imports. Adding the new confidential module also selects its minimum core dependency, so a separate core installation is not normally needed. The command above makes both sides of the migration explicit.

Do not use a local `replace` directive to force an older combined core release. Both modules would then contain the same helper package paths, which can cause ambiguous-import errors.

## Development workspace

For an unreleased repository checkout, use the [contributor guide](https://github.com/XRPLF/xrpl-go/blob/main/CONTRIBUTING.md#work-on-confidential-helpers). It covers `make workspace`, local dependency selection, tests, and examples. Published-module users do not need a workspace or a local `replace` directive.
