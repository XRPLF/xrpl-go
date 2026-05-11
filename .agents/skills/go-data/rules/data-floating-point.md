## Handle Floating-Point Precision Correctly

**Impact: HIGH**

Floating-point numbers are approximations. They cannot represent all decimal values exactly, arithmetic operations accumulate rounding errors, and comparing float results with `==` produces unexpected results. Ignoring this leads to subtle bugs in financial calculations, physics simulations, and any domain requiring precise numeric results.

**Incorrect:**

```go
// Direct equality comparison of floats is unreliable
a := 0.1 + 0.2
if a == 0.3 {  // false! a is 0.30000000000000004
    fmt.Println("equal")
}

// Order of operations affects precision
func Sum(prices []float64) float64 {
    var total float64
    for _, p := range prices {
        total += p  // Adding many small values to a large total loses precision
    }
    return total
}

// Naive financial calculations
price := 1.005
rounded := math.Round(price*100) / 100  // May not give 1.01 as expected
```

**Correct:**

```go
// Use epsilon comparison for floats
const epsilon = 1e-9

func FloatEquals(a, b float64) bool {
    return math.Abs(a-b) < epsilon
}

a := 0.1 + 0.2
if FloatEquals(a, 0.3) {
    fmt.Println("equal")  // Works correctly
}

// For financial calculations, use integer arithmetic (cents) or decimal library
type Money struct {
    cents int64  // Store as integer cents to avoid float issues
}

func (m Money) Add(other Money) Money {
    return Money{cents: m.cents + other.cents}
}

// Pairwise summation or sorting before summing improves precision
func SumSorted(values []float64) float64 {
    sorted := make([]float64, len(values))
    copy(sorted, values)
    sort.Float64s(sorted)  // Sort ascending: add small values first
    var total float64
    for _, v := range sorted {
        total += v
    }
    return total
}
```

Why this matters: IEEE 754 float64 has 53 bits of mantissa, giving about 15-16 significant decimal digits. Operations like `0.1 + 0.2` don't yield exact results because 0.1 and 0.2 cannot be represented exactly in binary. For money: use integer cents. For comparisons: use epsilon. For summing many values: sort first or use compensated summation.
