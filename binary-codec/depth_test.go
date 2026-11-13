package binarycodec

import (
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/binary-codec/types"
	"github.com/stretchr/testify/require"
)

func TestNestingDepth(t *testing.T) {
	for _, tc := range []struct {
		name        string
		objects     int
		array       bool
		scalar      bool
		expectedErr error
	}{
		{name: "scalar at depth 10", objects: 9, scalar: true},
		{name: "scalar at depth 11", objects: 10, scalar: true, expectedErr: types.ErrMaxNestingDepth},
		{name: "object at depth 10", objects: 10},
		{name: "object at depth 11", objects: 11, expectedErr: types.ErrMaxNestingDepth},
		{name: "array object at depth 10", objects: 8, array: true},
		{name: "array object at depth 11", objects: 9, array: true, expectedErr: types.ErrMaxNestingDepth},
		{name: "array scalar at depth 10", objects: 7, array: true, scalar: true},
		{name: "array scalar at depth 11", objects: 8, array: true, scalar: true, expectedErr: types.ErrMaxNestingDepth},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value := map[string]any{}
			binary := ""
			if tc.scalar {
				value["Sequence"] = uint32(1)
				binary = "2400000001"
			}
			if tc.array {
				value = map[string]any{"Memos": []any{map[string]any{"Memo": value}}}
				binary = "F9EA" + binary + "E1F1"
			}
			for range tc.objects {
				value = map[string]any{"TransactionMetaData": value}
				binary = "E2" + binary + "E1"
			}
			encoded, encodeErr := Encode(value)
			decoded, decodeErr := Decode(binary)
			if tc.expectedErr != nil {
				require.ErrorIs(t, encodeErr, tc.expectedErr)
				require.ErrorIs(t, decodeErr, tc.expectedErr)
				return
			}
			require.NoError(t, encodeErr)
			require.Equal(t, binary, encoded)
			require.NoError(t, decodeErr)
			require.Equal(t, value, decoded)
		})
	}
}

func TestDecodeDeepObjects(t *testing.T) {
	blob := strings.Repeat("EF", 16384) + strings.Repeat("E1", 16384)
	_, err := Decode(blob)
	require.ErrorIs(t, err, types.ErrMaxNestingDepth)
}
