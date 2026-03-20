package time

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeConversion_RippleTimeToUnixTime(t *testing.T) {
	testCases := []struct {
		name       string
		rippleTime int64
		unixTime   int64
	}{
		{
			name:       "pass - ripple Time 0",
			rippleTime: 0,
			unixTime:   946684800000,
		},
		{
			name:       "pass - ripple Time 1",
			rippleTime: 1,
			unixTime:   946684801000,
		},
		{
			name:       "pass - ripple Time 100",
			rippleTime: 100,
			unixTime:   946684900000,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := RippleTimeToUnixTime(tc.rippleTime)
			assert.Equal(t, tc.unixTime, actual)
		})
	}
}

func TestTimeConversion_UnixTimeToRippleTime(t *testing.T) {
	testCases := []struct {
		name       string
		rippleTime int64
		unixTime   int64
	}{
		{
			name:       "pass - ripple Time 0",
			rippleTime: 0,
			unixTime:   946684800,
		},
		{
			name:       "pass - ripple Time 1",
			rippleTime: 1,
			unixTime:   946684801,
		},
		{
			name:       "pass - ripple Time 100",
			rippleTime: 100,
			unixTime:   946684900,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := UnixTimeToRippleTime(tc.unixTime)
			assert.Equal(t, tc.rippleTime, actual)
		})
	}
}

func TestTimeConversion_RippleTimeToIsoTime(t *testing.T) {
	testCases := []struct {
		name       string
		rippleTime int64
		isoTime    string
	}{
		{
			name:       "pass - ISO time 2000-01-01T00:00:00.000Z",
			rippleTime: 0,
			isoTime:    "2000-01-01T00:00:00.000Z",
		},
		{
			name:       "pass - ISO time 2030-01-01T00:00:00.000Z",
			rippleTime: 946771200,
			isoTime:    "2030-01-01T00:00:00.000Z",
		},
		{
			name:       "pass - ISO time 2001-01-01T00:00:00.000Z",
			rippleTime: 31622400,
			isoTime:    "2001-01-01T00:00:00.000Z",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := RippleTimeToISOTime(tc.rippleTime)
			assert.Equal(t, tc.isoTime, actual)
		})
	}
}

func TestTimeConversion_IsoTimeToRippleTime(t *testing.T) {
	testCases := []struct {
		name       string
		rippleTime int64
		isoTime    string
		wantErr    bool
	}{
		{
			name:       "pass - ISO time 2000-01-01T00:00:00.000Z",
			rippleTime: 0,
			isoTime:    "2000-01-01T00:00:00.000Z",
			wantErr:    false,
		},
		{
			name:       "pass - ISO time 2030-01-01T00:00:00.000Z",
			rippleTime: 946771200,
			isoTime:    "2030-01-01T00:00:00.000Z",
			wantErr:    false,
		},
		{
			name:       "pass - ISO time 2001-01-01T00:00:00.000Z",
			rippleTime: 31622400,
			isoTime:    "2001-01-01T00:00:00.000Z",
			wantErr:    false,
		},
		{
			name:       "fail - Invalid ISO time",
			rippleTime: 31622400,
			isoTime:    "Invalid",
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := IsoTimeToRippleTime(tc.isoTime)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.rippleTime, actual)
			}
		})
	}
}

func TestTimeConversion_ParseISO8601(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected time.Time
		hasError bool
	}{
		{
			name:     "pass - Valid ISO8601 time",
			input:    "2023-10-01T12:34:56.789Z",
			expected: time.Date(2023, 10, 1, 12, 34, 56, 789000000, time.UTC),
			hasError: false,
		},
		{
			name:     "pass - Valid ISO8601 time without milliseconds",
			input:    "2023-10-01T12:34:56Z",
			expected: time.Date(2023, 10, 1, 12, 34, 56, 0, time.UTC),
			hasError: false,
		},
		{
			name:     "fail - Invalid ISO8601 time",
			input:    "invalid-time",
			expected: time.Time{},
			hasError: true,
		},
		{
			name:     "fail - Empty string",
			input:    "",
			expected: time.Time{},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := parseISO8601(tc.input)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}
