# Timestamps

Use `xrpl/time` to convert XRPL timestamps. Alias the import as `xrpltime` when also using Go's standard `time` package.

## Convert to a Go time

This offline example converts the start of the Ripple epoch to a Go `time.Time`, then converts it back:

```go
package main

import (
	"fmt"
	"time"

	xrpltime "github.com/Peersyst/xrpl-go/xrpl/time"
)

func main() {
	const rippleSeconds int64 = 0
	t := time.Unix(xrpltime.RippleTimeToUnixSeconds(rippleSeconds), 0).UTC()
	fmt.Println(t.Format(time.RFC3339))
	fmt.Println(xrpltime.UnixTimeToRippleTime(t.Unix()))
}
```

Expected output:

```text
2000-01-01T00:00:00Z
0
```

## Check the units

| Function | Input | Output |
| --- | --- | --- |
| `RippleTimeToUnixSeconds` | Ripple seconds | Unix seconds |
| `RippleTimeToUnixTime` | Ripple seconds | Unix milliseconds |
| `UnixTimeToRippleTime` | Unix seconds | Ripple seconds |
| `RippleTimeToISOTime` | Ripple seconds | UTC timestamp with millisecond formatting |
| `IsoTimeToRippleTime` | RFC3339 timestamp string | Ripple seconds, or a parsing error |

Despite their similar names, `RippleTimeToUnixTime` and `UnixTimeToRippleTime` are not unit-matched inverses. Use `RippleTimeToUnixSeconds` with `time.Unix`, or use `time.UnixMilli` for the milliseconds result. Subsecond precision is not retained in Ripple seconds.

See [XRPL time representation](https://xrpl.org/docs/references/protocol/data-types/basic-data-types#specifying-time) for the protocol epoch and the [Go API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/time) for all helpers.
