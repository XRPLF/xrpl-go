## Use chan struct{} for Notification Channels, Not chan bool

**Impact: LOW**

When a channel carries no meaningful data — only the fact that a signal was sent — use `chan struct{}` (empty struct), not `chan bool` or `chan int`. This makes the intent clear and uses zero memory.

**Incorrect:**

```go
// Using bool is ambiguous: what does false mean?
disconnectCh := make(chan bool)

// Receiver must wonder: does true mean "disconnected" and false mean "reconnected"?
// What if false is never sent? Should I expect it?
case connected := <-disconnectCh:
    if !connected { ... }  // Confusing

// Using int is even worse — what do the values mean?
doneCh := make(chan int)
```

**Correct:**

```go
// chan struct{} clearly signals "an event occurred" with no ambiguity about value
disconnectCh := make(chan struct{})

// Sending a notification:
disconnectCh <- struct{}{}

// Closing to broadcast to all receivers (most common for done signals):
close(disconnectCh)

// Receiver only cares that the channel fired, not what value it holds
case <-disconnectCh:
    fmt.Println("disconnected")  // No confusion about what the value means

// Examples from stdlib using this pattern:
// context.Context.Done() returns <-chan struct{}
// sync.WaitGroup uses this concept internally
// time.After returns <-chan time.Time, but done/quit channels use chan struct{}

// Empty struct costs zero bytes:
var s struct{}
fmt.Println(unsafe.Sizeof(s))  // 0

// Contrast with bool: 1 byte; interface{}: 16 bytes
```

Why this matters: `chan bool` implies the value of the boolean matters to the receiver, creating confusion about what `true` vs `false` conveys. `chan struct{}` is the idiomatic Go way to express "signal only, no data." It's zero-sized, so sending a `struct{}{}` is cheap. The pattern appears throughout the standard library (e.g., `context.Done()` channels). Use it for done signals, quit signals, and any channel where the fact of receipt is the message.
