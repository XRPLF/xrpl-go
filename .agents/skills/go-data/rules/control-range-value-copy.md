## Range Loop Values Are Copies — Mutations Don't Affect the Original

**Impact: MEDIUM**

In a `range` loop, the value variable is a copy of the element, not a reference to it. Mutating the value variable does not modify the original slice or map element. This is one of the most common Go surprises for developers coming from other languages.

**Incorrect:**

```go
type Account struct {
    Balance float64
}

accounts := []Account{
    {Balance: 100},
    {Balance: 200},
    {Balance: 300},
}

// BUG: a is a copy of each element — mutations don't affect the slice
for _, a := range accounts {
    a.Balance += 1000  // Modifies the copy, not accounts[i]
}

fmt.Println(accounts)  // [{100} {200} {300}] — unchanged!
```

**Correct:**

```go
// Option 1: Use the index to access the original element
for i := range accounts {
    accounts[i].Balance += 1000
}

// Option 2: Classic for loop
for i := 0; i < len(accounts); i++ {
    accounts[i].Balance += 1000
}

// Option 3: Slice of pointers (pointer copy points to same struct)
accounts := []*Account{{Balance: 100}, {Balance: 200}, {Balance: 300}}
for _, a := range accounts {
    a.Balance += 1000  // a is a pointer copy — but still points to original
}
// Note: pointer slices can be less cache-friendly (CPU prefetching)
```

Why this matters: Go's rule is simple — everything is a copy when assigned. The range value variable is assigned a copy each iteration. For structs you want to mutate, use the index or a pointer. The same applies to maps: `for k, v := range m { v.Field = x }` won't modify the map values if they're structs.
