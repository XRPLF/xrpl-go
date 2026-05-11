## Common JSON Handling Mistakes: Embedding, time.Time, and map[string]any

**Impact: MEDIUM**

Three common JSON mistakes: (1) embedding a type that implements `json.Marshaler` hijacks the parent's marshaling; (2) marshaling/unmarshaling `time.Time` loses the monotonic clock component; (3) `json.Unmarshal` into `map[string]any` converts all numbers to `float64`.

**Mistake 1: Embedded type hijacks JSON marshaling**

```go
// WRONG: time.Time implements json.Marshaler; embedding it hijacks Event's marshaling
type Event struct {
    ID   int
    time.Time  // Embedded — Event now "is" a json.Marshaler via time.Time
}

event := Event{ID: 1234, Time: time.Now()}
b, _ := json.Marshal(event)
// Output: "2021-05-18T21:15:08.381652+02:00"  — ID is LOST!

// CORRECT: Use a named field instead of embedding
type Event struct {
    ID   int
    Time time.Time  // Named field — Event uses default struct marshaling
}
// Output: {"ID":1234,"Time":"2021-05-18T21:15:08.381652+02:00"}
```

**Mistake 2: time.Time loses monotonic part after marshal/unmarshal**

```go
// time.Now() has both wall clock and monotonic reading
// JSON marshaling only preserves the wall clock
event1 := Event{Time: time.Now()}            // Has monotonic part (m=+0.000338660)
b, _ := json.Marshal(event1)
var event2 Event
json.Unmarshal(b, &event2)                   // event2.Time has NO monotonic part

fmt.Println(event1 == event2)  // FALSE — different structs due to monotonic difference

// CORRECT: Use time.Equal() for comparison, or strip monotonic before marshaling
fmt.Println(event1.Time.Equal(event2.Time))  // TRUE — ignores monotonic

// Or strip monotonic when creating:
event1 := Event{Time: time.Now().Truncate(0)}  // Truncate(0) removes monotonic part
```

**Mistake 3: map[string]any converts all numbers to float64**

```go
b := []byte(`{"id": 32, "name": "foo"}`)
var m map[string]any
json.Unmarshal(b, &m)

fmt.Printf("%T\n", m["id"])  // float64 — NOT int!
// This can cause goroutine panics if you assert m["id"].(int)

// CORRECT: Use a typed struct for known JSON schemas
type Message struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}
var msg Message
json.Unmarshal(b, &msg)
fmt.Printf("%T\n", msg.ID)  // int ✓
```

Why this matters: (1) Embedded types promote all methods including `MarshalJSON`; the embedded type's marshaler overrides the parent's default. (2) `time.Time.Equal()` ignores monotonic time, while `==` includes it — always use `Equal()` to compare times. (3) JSON numbers without decimal points are still parsed as `float64` in untyped maps — use typed structs when possible.
