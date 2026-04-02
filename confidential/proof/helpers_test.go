//go:build cgo

package proof_test

import "strings"

const (
	testAccount    = "rDTXLQ7ZKZVKz33zJbHjgVShjsBnqMBhmN"
	testDest       = "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59"
	testHolder     = "rJKhsipKHooQbtS3v5Jro6N5Q7TMNPkoAs"
	testIssuanceID = "000000000000000000000000000000000000000000000001"
)

// zeroHex returns a hex string of n zero bytes (2*n hex chars).
func zeroHex(n int) string {
	return strings.Repeat("00", n)
}
