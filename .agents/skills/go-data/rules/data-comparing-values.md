## Know When == Works and When to Use reflect.DeepEqual

**Impact: MEDIUM**

Go's `==` operator only works on *comparable* types. Slices and maps are not comparable — using `==` on structs containing them causes a compile error, and comparing `any`-typed values containing them causes a runtime panic. Know the alternatives.

**Incorrect:**

```go
type Customer struct {
    ID         string
    Operations []float64  // slice — not comparable
}

cust1 := Customer{ID: "x", Operations: []float64{1.0}}
cust2 := Customer{ID: "x", Operations: []float64{1.0}}

// Compile error: invalid operation: cust1 == cust2
// (struct containing []float64 cannot be compared)
fmt.Println(cust1 == cust2)

// Worse: compiles but panics at runtime when using any
var a any = cust1
var b any = cust2
fmt.Println(a == b)  // panic: runtime error: comparing uncomparable type
```

**Correct:**

```go
// Option 1: reflect.DeepEqual — works on anything, ~100x slower than ==
fmt.Println(reflect.DeepEqual(cust1, cust2))  // true
// Note: DeepEqual distinguishes nil from empty slices

// Option 2: Custom comparison method — fast, explicit, correct
func (a Customer) Equal(b Customer) bool {
    if a.ID != b.ID {
        return false
    }
    if len(a.Operations) != len(b.Operations) {
        return false
    }
    for i := range a.Operations {
        if a.Operations[i] != b.Operations[i] {
            return false
        }
    }
    return true
}
// Custom method is ~96x faster than reflect.DeepEqual

// Option 3: Use go-cmp or testify in tests
// assert.Equal(t, cust1, cust2)  // testify handles deep equality
```

Comparable types (can use `==`): booleans, numbers, strings, pointers, channels, interfaces, arrays (of comparable elements), structs (where all fields are comparable).

Non-comparable types (cannot use `==`): slices, maps, functions.

Use `reflect.DeepEqual` for correctness in tests where performance isn't critical. Implement custom `Equal` methods for production code where performance matters. Check `bytes.Compare` and similar standard library functions before implementing your own.
