## Use sync.Cond to Broadcast Notifications to Multiple Goroutines

**Impact: LOW**

When multiple goroutines need to wait for the same repeating condition, `sync.Cond` is the right tool. A channel can only deliver a message to one goroutine at a time; only a channel closure broadcasts to all, but closing is a one-shot action. `sync.Cond.Broadcast()` wakes all waiting goroutines each time a condition changes.

**Incorrect:**

```go
// Busy loop wastes CPU — checking condition repeatedly
for donation.balance < goal {
    // spinning without sleeping burns CPU at 100%
}

// Channel approach: each message goes to ONE goroutine (round-robin)
// Multiple listeners miss notifications
ch <- balance  // Only one listener receives each update
```

**Correct:**

```go
type Donation struct {
    cond    *sync.Cond
    balance int
}

donation := &Donation{
    cond: sync.NewCond(&sync.Mutex{}),
}

// Listener goroutines — wait for condition to be met
go func(goal int) {
    donation.cond.L.Lock()
    for donation.balance < goal {
        donation.cond.Wait()  // Atomically: unlock, suspend, re-lock when woken
    }
    fmt.Printf("$%d goal reached\n", donation.balance)
    donation.cond.L.Unlock()
}(10)

go func(goal int) {
    donation.cond.L.Lock()
    for donation.balance < goal {
        donation.cond.Wait()
    }
    fmt.Printf("$%d goal reached\n", donation.balance)
    donation.cond.L.Unlock()
}(15)

// Updater goroutine — broadcasts every time the condition changes
for {
    time.Sleep(time.Second)
    donation.cond.L.Lock()
    donation.balance++
    donation.cond.L.Unlock()
    donation.cond.Broadcast()  // Wakes ALL waiting goroutines to re-check condition
}
```

Why this matters: `sync.Cond.Wait()` atomically releases the lock and suspends the goroutine, then re-acquires the lock when woken — no busy loop, no wasted CPU. `Broadcast()` wakes all goroutines waiting on the condition; `Signal()` wakes one. Always check the condition in a `for` loop (not `if`) after `Wait` returns, because spurious wakeups can occur. Use `sync.Cond` when multiple goroutines need repeated broadcast notifications about a shared state change.
