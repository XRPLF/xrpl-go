// Package time provides time conversion utilities for XRPL timestamps.
package time

import (
	"time"
)

// The rippled server and its APIs represent time as an unsigned integer.
// This number measures the number of seconds since the "Ripple Epoch" of January 1, 2000 (00:00 UTC).
// This is like the way the Unix epoch works, except the Ripple Epoch is 946684800 seconds after the Unix Epoch.
const (
	RippleEpochDiff int64  = 946684800
	ISO8601Format   string = "2006-01-02T15:04:05.000Z"
)

// RippleTimeToUnixTime converts a ripple timestamp to a unix timestamp.
//
// rpepoch is the number of seconds since January 1, 2000 (00:00 UTC).
//
// It returns the number of milliseconds since the Unix epoch (January 1, 1970 00:00 UTC).
func RippleTimeToUnixTime(rpepoch int64) int64 {
	return RippleTimeToUnixSeconds(rpepoch) * 1000
}

// RippleTimeToUnixSeconds converts a ripple timestamp to unix elapsed seconds.
//
// rpepoch is the number of seconds since January 1, 2000 (00:00 UTC).
//
// It returns the number of seconds since the Unix epoch (January 1, 1970 00:00 UTC).
func RippleTimeToUnixSeconds(rpepoch int64) int64 {
	return rpepoch + RippleEpochDiff
}

// UnixTimeToRippleTime converts a unix timestamp to a ripple timestamp.
//
// timestamp is the number of milliseconds since the Unix epoch (January 1, 1970 00:00 UTC).
//
// It returns the number of seconds since the Ripple epoch (January 1, 2000 00:00 UTC).
func UnixTimeToRippleTime(timestamp int64) int64 {
	return timestamp - RippleEpochDiff
}

// RippleTimeToISOTime converts a ripple timestamp to an ISO 8601 formatted time string.
//
// rpepoch is the number of seconds since January 1, 2000 (00:00 UTC).
//
// It returns the time formatted as an ISO 8601 string.
func RippleTimeToISOTime(rippleTime int64) string {
	unixTime := RippleTimeToUnixTime(rippleTime)
	return time.Unix(int64(unixTime/1000), 0).UTC().Format(ISO8601Format)
}

// IsoTimeToRippleTime converts an ISO8601 timestamp to a ripple timestamp.
//
// iso8601 is the ISO 8601 formatted string.
//
// It returns the seconds since ripple epoch (1/1/2000 GMT).
func IsoTimeToRippleTime(isoTime string) (int64, error) {
	t, err := parseISO8601(isoTime)
	if err != nil {
		return 0, err
	}
	return UnixTimeToRippleTime(t.Unix()), nil
}

// ParseISO8601 parses an ISO 8601 formatted string into a time.Time object.
//
// iso8601 is the ISO 8601 formatted string.
//
// It returns the parsed time.Time object and an error if the parsing fails.
func parseISO8601(iso8601 string) (time.Time, error) {
	return time.Parse(time.RFC3339, iso8601)
}
