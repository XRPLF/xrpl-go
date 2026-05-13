## Avoid Unintended Variable Shadowing

**Impact: MEDIUM**

Variable shadowing occurs when you declare a variable in an inner scope with the same name as an outer scope variable. This is legal in Go but often leads to confusing bugs where you think you're modifying one variable but are actually working with a different one.

**Incorrect:**

```go
func LoadConfig() (*Config, error) {
    config := &Config{Timeout: 30}

    if file, err := os.Open("config.json"); err == nil {
        defer file.Close()
        config, err := json.Unmarshal(data, &config)  // Shadows outer 'config'!
        if err != nil {
            return nil, err
        }
        // This config only exists in this block
    }

    return config, nil  // Returns partially initialized config!
}
```

**Correct:**

```go
func LoadConfig() (*Config, error) {
    config := &Config{Timeout: 30}

    file, err := os.Open("config.json")
    if err == nil {
        defer file.Close()
        err = json.Unmarshal(data, config)  // Uses outer 'config'
        if err != nil {
            return nil, err
        }
    }

    return config, nil
}
```

Why this matters: The incorrect version creates a new `config` variable inside the if block that shadows the outer one. Changes to this shadowed variable don't affect the outer `config`, so the function returns the partially initialized default config instead of the loaded one.

Prevention: Use `go vet` which detects some shadowing cases. Consider shorter scopes and unique variable names. The `:=` operator is convenient but dangerous in nested scopes—sometimes explicit `var` declarations make shadowing more obvious.
