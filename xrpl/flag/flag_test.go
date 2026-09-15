package flag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	dummyFlagA = 262144
	dummyFlagB = 131072
	dummyFlagC = 65536
)

func TestFlag_Contains(t *testing.T) {
	testCases := []struct {
		name        string
		currentFlag uint32
		flag        uint32
		expected    bool
	}{
		{
			name:        "pass - same flag",
			currentFlag: dummyFlagA,
			flag:        dummyFlagA,
			expected:    true,
		},
		{
			name:        "pass - containing flag",
			currentFlag: dummyFlagA | dummyFlagB,
			flag:        dummyFlagA,
			expected:    true,
		},
		{
			name:        "pass - not containing flag",
			currentFlag: dummyFlagA | dummyFlagB,
			flag:        dummyFlagC,
			expected:    false,
		},
		{
			name:        "pass - zero flag",
			currentFlag: dummyFlagA,
			flag:        0,
			expected:    false,
		},
		{
			name:        "pass - zero current flag",
			currentFlag: 0,
			flag:        0,
			expected:    false,
		},
		{
			name:        "pass - partial overlap multi-bit flag",
			currentFlag: dummyFlagA,
			flag:        dummyFlagA | dummyFlagB,
			expected:    false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := Contains(tc.currentFlag, tc.flag)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestFlag_ContainsAny(t *testing.T) {
	testCases := []struct {
		name         string
		currentFlags uint32
		mask         uint32
		expected     bool
	}{
		{
			name:         "same flag",
			currentFlags: dummyFlagA,
			mask:         dummyFlagA,
			expected:     true,
		},
		{
			name:         "partial overlap",
			currentFlags: dummyFlagA | dummyFlagC,
			mask:         dummyFlagA | dummyFlagB,
			expected:     true,
		},
		{
			name:         "all selected bits",
			currentFlags: dummyFlagA | dummyFlagB | dummyFlagC,
			mask:         dummyFlagA | dummyFlagB,
			expected:     true,
		},
		{
			name:         "no overlap",
			currentFlags: dummyFlagC,
			mask:         dummyFlagA | dummyFlagB,
			expected:     false,
		},
		{
			name:         "zero mask",
			currentFlags: dummyFlagA,
			mask:         0,
			expected:     false,
		},
		{
			name:         "zero current flags",
			currentFlags: 0,
			mask:         dummyFlagA,
			expected:     false,
		},
		{
			name:         "zero values",
			currentFlags: 0,
			mask:         0,
			expected:     false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ContainsAny(tc.currentFlags, tc.mask))
		})
	}
}

func TestFlag_ContainsOnly(t *testing.T) {
	testCases := []struct {
		name         string
		currentFlag  uint32
		allowedFlags uint32
		expected     bool
	}{
		{
			name:         "pass - all flags are allowed",
			currentFlag:  dummyFlagA | dummyFlagB,
			allowedFlags: dummyFlagA | dummyFlagB | dummyFlagC,
			expected:     true,
		},
		{
			name:         "pass - unsupported flag",
			currentFlag:  dummyFlagA | dummyFlagC,
			allowedFlags: dummyFlagA | dummyFlagB,
			expected:     false,
		},
		{
			name:         "pass - no flags",
			currentFlag:  0,
			allowedFlags: dummyFlagA,
			expected:     true,
		},
		{
			name:         "pass - no flags allowed",
			currentFlag:  dummyFlagA,
			allowedFlags: 0,
			expected:     false,
		},
		{
			name:         "pass - zero values",
			currentFlag:  0,
			allowedFlags: 0,
			expected:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := ContainsOnly(tc.currentFlag, tc.allowedFlags)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
