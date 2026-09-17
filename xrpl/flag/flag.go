// Package flag provides utility functions for working with bitwise flags.
package flag

// Contains checks if all bits of flag are present in currentFlag.
// Returns false if flag is 0.
// Note: It should be taken into account that the comparison is based on the flag value as a uint32.
// Different contexts may use same values, and they will return ok as the value matches.
func Contains(currentFlag, flag uint32) bool {
	return flag != 0 && (currentFlag&flag) == flag
}

// ContainsAny reports whether any bits in mask are set in currentFlags.
// Returns false if mask is zero.
func ContainsAny(currentFlags, mask uint32) bool {
	return currentFlags&mask != 0
}

// ContainsOnly checks if currentFlag contains no bits outside allowedFlags.
func ContainsOnly(currentFlag, allowedFlags uint32) bool {
	return currentFlag&^allowedFlags == 0
}
