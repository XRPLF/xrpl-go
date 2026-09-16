package transaction

import (
	"encoding/json"
	"strings"
	"testing"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestVaultDeleteMemoDataJSON(t *testing.T) {
	empty, metadata := "", "ABCD"
	tests := []struct {
		name     string
		memo     *string
		wantJSON string
	}{
		{
			name:     "absent",
			wantJSON: `{"Account":"","TransactionType":"","VaultID":""}`,
		},
		{
			name:     "explicit empty preserved",
			memo:     &empty,
			wantJSON: `{"Account":"","TransactionType":"","VaultID":"","MemoData":""}`,
		},
		{
			name:     "metadata",
			memo:     &metadata,
			wantJSON: `{"Account":"","TransactionType":"","VaultID":"","MemoData":"ABCD"}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			want := VaultDelete{MemoData: tc.memo}
			var decoded VaultDelete
			require.NoError(t, json.Unmarshal([]byte(tc.wantJSON), &decoded))
			require.Equal(t, want, decoded)

			data, err := json.Marshal(want)
			require.NoError(t, err)
			require.JSONEq(t, tc.wantJSON, string(data))
		})
	}
}

func TestVaultDeleteMemoDataFlattenPresence(t *testing.T) {
	tx := VaultDelete{}
	require.NotContains(t, tx.Flatten(), "MemoData")

	empty := ""
	tx.MemoData = &empty
	require.Equal(t, FlatTransaction{
		"TransactionType": "VaultDelete",
		"VaultID":         "",
		"MemoData":        "",
	}, tx.Flatten())
}

func TestVaultDeleteMemoDataRoundTrip(t *testing.T) {
	metadata := strings.Repeat("AB", 256)
	tx := VaultDelete{
		BaseTx:   BaseTx{Account: "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es", Memos: []types.MemoWrapper{{Memo: types.Memo{MemoData: "4344"}}}},
		VaultID:  "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
		MemoData: &metadata,
	}
	flat := tx.Flatten()
	require.Equal(t, metadata, flat["MemoData"])
	blob, err := binarycodec.Encode(flat)
	require.NoError(t, err)
	decoded, err := binarycodec.Decode(blob)
	require.NoError(t, err)
	require.Equal(t, metadata, decoded["MemoData"])
	require.Equal(t, []any{map[string]any{"Memo": map[string]any{"MemoData": "4344"}}}, decoded["Memos"])
	signing, err := binarycodec.EncodeForSigning(flat)
	require.NoError(t, err)
	require.Equal(t, "53545800"+blob, signing)
	delete(flat, "MemoData")
	withoutMetadata, err := binarycodec.EncodeForSigning(flat)
	require.NoError(t, err)
	require.NotEqual(t, withoutMetadata, signing)
}
