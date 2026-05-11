## Prevent Memory Leaks When Slicing Large Slices

**Impact: HIGH**

When you slice a large slice or array, the result shares the original backing array. Even if the original slice is no longer referenced, the GC cannot reclaim the unused elements because the sub-slice still holds a reference to the entire backing array. This causes significant memory leaks in long-running programs.

**Incorrect:**

```go
// Receives large messages (e.g., 1 million bytes each)
func getMessageType(msg []byte) []byte {
    return msg[:5]  // 5-byte result holds reference to entire 1M backing array!
}

func consumeMessages() {
    var messageTypes [][]byte
    for {
        msg := receiveMessage()  // 1 million bytes
        // Each stored type holds a 1M-byte backing array in memory
        messageTypes = append(messageTypes, getMessageType(msg))
        // After 1,000 messages: ~1 GB held in memory instead of ~5 KB
    }
}

// Also leaks with structs containing pointer fields
func keepFirstTwo(foos []Foo) []Foo {
    return foos[:2]  // foos[2:] elements with pointer fields aren't GC'd
}
```

**Correct:**

```go
// Solution 1: Copy the sub-slice to break the backing array reference
func getMessageType(msg []byte) []byte {
    msgType := make([]byte, 5)
    copy(msgType, msg)
    return msgType  // Independent 5-byte slice; original msg can be GC'd
}

// Solution 2: For structs with pointer fields, nil out unused elements
func keepFirstTwo(foos []Foo) []Foo {
    // Nil out pointer fields so GC can collect referenced memory
    for i := 2; i < len(foos); i++ {
        foos[i].data = nil  // Or foos[i] = Foo{} to zero the whole struct
    }
    return foos[:2]
}

// Note: full slice expression msg[:5:5] does NOT fix the leak —
// the backing array is still referenced and won't be GC'd
```

Why this matters: A sub-slice's capacity retains the entire original backing array in memory. With 1,000 messages of 1 MB each, storing only the first 5 bytes via slicing uses ~1 GB instead of ~5 KB. Always copy when you need to retain a small portion of a large slice. For slices of structs with pointer fields, nil out the excluded elements' pointer fields.
