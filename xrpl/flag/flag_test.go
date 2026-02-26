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
