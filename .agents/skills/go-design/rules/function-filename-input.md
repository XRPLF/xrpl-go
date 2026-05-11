## Accept io.Reader Instead of Filename as Function Input

**Impact: MEDIUM**

Functions that accept a filename to read from are harder to test and less reusable than functions that accept an `io.Reader`. A filename forces callers to create actual files, while `io.Reader` accepts files, HTTP bodies, strings, byte buffers, or any other data source.

**Incorrect:**

```go
// Tightly coupled to the filesystem — hard to test, hard to reuse
func countEmptyLines(filename string) (int, error) {
    file, err := os.Open(filename)
    if err != nil {
        return 0, err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    var count int
    for scanner.Scan() {
        if scanner.Text() == "" {
            count++
        }
    }
    return count, scanner.Err()
}

// Testing requires creating actual files:
// os.WriteFile("test.txt", []byte("foo\n\nbar"), 0o644)
// count, err := countEmptyLines("test.txt")
// os.Remove("test.txt")
```

**Correct:**

```go
// Accepts any data source — testable and reusable
func countEmptyLines(r io.Reader) (int, error) {
    scanner := bufio.NewScanner(r)
    var count int
    for scanner.Scan() {
        if scanner.Text() == "" {
            count++
        }
    }
    return count, scanner.Err()
}

// Testing is trivial — no files needed:
func TestCountEmptyLines(t *testing.T) {
    input := strings.NewReader("foo\n\nbar\n\nbaz")
    count, err := countEmptyLines(input)
    // ...
}

// Works with files:
file, _ := os.Open("data.txt")
defer file.Close()
count, _ := countEmptyLines(file)

// Works with HTTP request body:
count, _ := countEmptyLines(r.Body)
```

Why this matters: `*os.File`, `http.Request.Body`, `strings.NewReader`, and `bytes.NewReader` all implement `io.Reader`. A function accepting `io.Reader` works with all of them transparently. Test cases don't need filesystem setup/teardown. The function is reusable across different data sources without duplication.

Exception: `os.Open` and similar filesystem-specific functions obviously need filenames. But functions that just *read* data should prefer `io.Reader`.
