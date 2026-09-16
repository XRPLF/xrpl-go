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
	const maxMetadataBytes = 256
	deletionMetadata := strings.Repeat("AB", maxMetadataBytes)
	tx := VaultDelete{
		BaseTx: BaseTx{
			Account: "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
			Memos: []types.MemoWrapper{
				{Memo: types.Memo{MemoData: "4344"}},
			},
		},
		VaultID:  "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
		MemoData: &deletionMetadata,
	}

	// Top-level deletion metadata must remain separate from transaction memos.
	flat := tx.Flatten()
	require.Equal(t, deletionMetadata, flat["MemoData"])

	encodedTx, err := binarycodec.Encode(flat)
	require.NoError(t, err)
	decodedTx, err := binarycodec.Decode(encodedTx)
	require.NoError(t, err)

	wantMemos := []any{
		map[string]any{
			"Memo": map[string]any{"MemoData": "4344"},
		},
	}
	require.Equal(t, deletionMetadata, decodedTx["MemoData"])
	require.Equal(t, wantMemos, decodedTx["Memos"])

	// All fields in this fixture belong in the signing payload.
	const signingPrefix = "53545800" // "STX\x00", the single-signing prefix.
	signingData, err := binarycodec.EncodeForSigning(flat)
	require.NoError(t, err)
	require.Equal(t, signingPrefix+encodedTx, signingData)

	// Removing deletion metadata must change the data that is signed.
	tx.MemoData = nil
	signingDataWithoutMetadata, err := binarycodec.EncodeForSigning(tx.Flatten())
	require.NoError(t, err)
	require.NotEqual(t, signingData, signingDataWithoutMetadata)
}
