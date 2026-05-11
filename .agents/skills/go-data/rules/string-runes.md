## Understand Runes: len() Returns Bytes, Not Characters

**Impact: MEDIUM**

In Go, a `rune` is a Unicode code point (alias for `int32`). Strings are sequences of bytes, not characters. `len(s)` returns the number of bytes, not the number of characters. Multi-byte characters (like Chinese, emoji, or accented letters) can span 2-4 bytes each, causing off-by-one errors and corrupted output when you treat strings as byte arrays.

**Incorrect:**

```go
s := "hêllo"  // ê is a 2-byte UTF-8 character

fmt.Println(len(s))    // 6, not 5! (ê takes 2 bytes)
fmt.Println(s[1])      // 195 (raw byte), not 'ê'

// Iterating by index gives bytes, not runes
for i := 0; i < len(s); i++ {
    fmt.Printf("%c", s[i])  // Prints garbled output: hÃllo
}

// Truncating at byte index corrupts multi-byte characters
short := s[:3]  // Cuts ê in the middle — invalid UTF-8!
```

**Correct:**

```go
s := "hêllo"

// Get rune count (not byte count)
fmt.Println(utf8.RuneCountInString(s))  // 5 ✓

// Iterate over runes using range — i is byte position, r is the rune
for i, r := range s {
    fmt.Printf("position %d: %c\n", i, r)
}
// position 0: h
// position 1: ê  (starts at byte 1, takes 2 bytes)
// position 3: l  (starts at byte 3)
// ...

// Access the nth rune: convert to []rune first
runes := []rune(s)
fmt.Printf("%c\n", runes[1])  // ê ✓

// Safe substring by rune index
sub := string([]rune(s)[:3])  // "hêl" ✓
```

Why this matters: `len(s)` and indexing `s[i]` operate on bytes. A `range` loop over a string produces rune values with their starting byte positions. When working with text that may contain non-ASCII characters (names, addresses, international content), always use rune-based operations. Use `unicode/utf8` package for byte-level UTF-8 manipulation.
