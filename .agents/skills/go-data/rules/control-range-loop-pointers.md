## Don't Store Pointers to Range Loop Variables

**Impact: CRITICAL**

Range loop variables are reused across iterations. Taking their address creates pointers that all point to the same memory location, causing all your stored pointers to reference the last element.

**Incorrect:**

```go
func CollectUsers(names []string) []*User {
    var users []*User
    for _, name := range names {
        user := User{Name: name}
        users = append(users, &user)  // All point to same variable!
    }
    return users  // All users have the last name!
}
```

**Correct:**

```go
func CollectUsers(names []string) []*User {
    users := make([]*User, 0, len(names))
    for _, name := range names {
        user := User{Name: name}
        users = append(users, &user)  // Creates new user each iteration
    }
    return users
}

// Or better - use index:
func CollectUsers(names []string) []*User {
    users := make([]*User, len(names))
    for i := range names {
        users[i] = &User{Name: names[i]}
    }
    return users
}
```

Why this matters: This bug is subtle and common. The range loop reuses the same `user` variable for each iteration. Taking `&user` gives you the address of this shared variable, so all your pointers end up pointing to it. After the loop, it contains the last value, making all pointers reference the same data.

Solution: Declare variables inside the loop body (as shown in Correct version 1), use the index to access the original slice, or in Go 1.22+, the loop variable is implicitly copied per iteration. Always use the `-race` flag during testing to catch these issues.
