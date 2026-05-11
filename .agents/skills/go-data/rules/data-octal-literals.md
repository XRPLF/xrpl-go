## Use Clear Octal Literal Syntax

**Impact: LOW**

In Go, an integer literal starting with `0` (zero) is an octal (base-8) number. This is easy to miss at a glance and causes subtle bugs when you expect decimal values. Go 1.13 introduced the `0o` prefix specifically to make octal intent explicit.

**Incorrect:**

```go
// 0755 looks like decimal 755 but is octal 493
os.Mkdir("secrets", 0755)     // permission 755 octal = rwxr-xr-x

// Reading this constant, would you know it's octal?
const permissions = 0600      // Actually 384 decimal

// Easy to make a mistake
timeout := 010                // "ten seconds"? No — this is 8!
```

**Correct:**

```go
// 0o prefix makes octal explicit and unmistakable
os.Mkdir("secrets", 0o755)    // Clearly octal file permissions

const permissions = 0o600     // Reader immediately knows: octal

// Never use leading zero for decimal integers
timeout := 10                 // This is decimal 10

// Binary and hex also have clear prefixes
flags := 0b10110011           // binary
mask := 0xFF                  // hexadecimal
```

Why this matters: `010` and `10` look nearly identical but have different values (8 vs 10). This is especially dangerous with file permissions where `0777`, `0755`, `0644` are common values — developers may type them as `777`, `755`, `644` (decimal) and get completely wrong permission bits.

The Go vet tool does not warn about leading-zero octals, so there's no automatic safety net. Use `0o` for all octal literals and code reviewers will immediately recognize the intent.
