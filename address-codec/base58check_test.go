package addresscodec

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var testcases = []struct {
	prefix []byte
	in     []byte
	out    string
}{
	{
		[]byte{0x00},
		[]byte{136, 165, 165, 124, 130, 159, 64, 242, 94, 168, 51, 133, 187, 222, 108, 61, 139, 76, 160, 130},
		"rDTXLQ7ZKZVKz33zJbHjgVShjsBnqMBhmN",
	},
	{
		[]byte{FamilySeedPrefix},
		[]byte("testingsomething"),
		"shKMVJjV52uudwfS7HzzaiwmZqVeP",
	},
}

func TestCheckEncode(t *testing.T) {
	for x, tc := range testcases {
		// test encoding
		if res := Base58CheckEncode([]byte(tc.in), tc.prefix...); res != tc.out {
			t.Errorf("CheckEncode test #%d failed: got %s, want: %s", x, res, tc.out)
		}
	}
}

func TestCheckDecode(t *testing.T) {
	for x, tc := range testcases {

		res, err := Base58CheckDecode(tc.out)
		require.NoError(t, err, "CheckDecode test #%d", x)
		require.NotEmpty(t, res, "CheckDecode test #%d", x)
		require.Equal(t, tc.in, res[1:], "CheckDecode test #%d", x)

		// test the two decoding failure cases
		// case 1: checksum error
		_, err = Base58CheckDecode("3MNQE1Y")
		require.ErrorIs(t, err, ErrChecksum)
		// case 2: invalid formats (string lengths below 5 mean the version byte and/or the checksum
		// bytes are missing).
		for i := range 4 {
			_, err = Base58CheckDecode(strings.Repeat("x", i+1))
			require.ErrorIs(t, err, ErrInvalidFormat)
		}
	}
}
