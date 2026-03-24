// Package maputils provides utility functions for working with maps.
package maputils

// GetKeys returns the keys of the given map.
func GetKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
