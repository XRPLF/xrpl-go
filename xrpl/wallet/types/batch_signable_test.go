package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBatchSignable_Flatten(t *testing.T) {
	tc := []struct {
		name string
		bs   BatchSignable
		want map[string]any
	}{
		{
			name: "pass - empty batch signable",
			bs:   BatchSignable{},
			want: map[string]any{
				"flags": uint32(0),
			},
		},
		{
			name: "pass - batch signable with flags and txids",
			bs: BatchSignable{
				Flags: 0,
				TxIDs: []string{"tx1", "tx2"},
			},
			want: map[string]any{
				"flags": uint32(0),
				"txIDs": []string{"tx1", "tx2"},
			},
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.bs.Flatten()
			require.Equal(t, tt.want, got)
		})
	}
}
