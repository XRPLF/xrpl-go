package transaction

import (
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestConfidentialMPTMergeInbox_TxType(t *testing.T) {
	tx := &ConfidentialMPTMergeInbox{}
	require.Equal(t, ConfidentialMPTMergeInboxTx, tx.TxType())
}

func TestConfidentialMPTMergeInbox_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		tx       *ConfidentialMPTMergeInbox
		expected FlatTransaction
	}{
		{
			name: "pass - all fields",
			tx: &ConfidentialMPTMergeInbox{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:     types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID: "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
			},
			expected: FlatTransaction{
				"Account":           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee":               "12",
				"TransactionType":   "ConfidentialMPTMergeInbox",
				"MPTokenIssuanceID": "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flattened := tt.tx.Flatten()
			require.Equal(t, tt.expected, flattened)
		})
	}
}

func TestConfidentialMPTMergeInbox_TaggedAccountEncoding(t *testing.T) {
	const (
		account   = "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"
		sourceTag = uint32(7)
	)
	taggedAccount, err := addresscodec.ClassicAddressToXAddress(account, sourceTag, true, false)
	require.NoError(t, err)
	tx := &ConfidentialMPTMergeInbox{
		BaseTx:            BaseTx{Account: types.Address(taggedAccount), TransactionType: ConfidentialMPTMergeInboxTx, Fee: 12},
		MPTokenIssuanceID: "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
	}
	valid, err := tx.Validate()
	require.True(t, valid)
	require.NoError(t, err)

	encoded, err := binarycodec.Encode(tx.Flatten())
	require.NoError(t, err)
	decoded, err := binarycodec.Decode(encoded)
	require.NoError(t, err)
	require.Equal(t, account, decoded["Account"])
	require.EqualValues(t, sourceTag, decoded["SourceTag"])
}

func TestConfidentialMPTMergeInbox_BinaryRoundTrip(t *testing.T) {
	const expected = "120056240000000581145B812C9D57731E27A2DA8B1830195F88EF32A3B60115000004C463C52827307480341125DA0577DEFC38405B0E3E"
	tx := &ConfidentialMPTMergeInbox{
		BaseTx: BaseTx{
			Account:         "r9LqNeG6qHxjeUocjvVki2XR35weJ9mZgQ",
			TransactionType: ConfidentialMPTMergeInboxTx,
			Sequence:        5,
		},
		MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
	}

	encoded, err := binarycodec.Encode(tx.Flatten())
	require.NoError(t, err)
	require.Equal(t, expected, encoded)
	decoded, err := binarycodec.Decode(encoded)
	require.NoError(t, err)
	require.Equal(t, map[string]any(tx.Flatten()), decoded)
}

func TestConfidentialMPTMergeInbox_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tx      *ConfidentialMPTMergeInbox
		wantErr error
	}{
		{
			name: "pass - valid transaction",
			tx: &ConfidentialMPTMergeInbox{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTMergeInboxTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID: "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
			},
			wantErr: nil,
		},
		{
			name: "fail - empty MPTokenIssuanceID",
			tx: &ConfidentialMPTMergeInbox{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTMergeInboxTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID: "",
			},
			wantErr: ErrConfidentialMPTInvalidIssuanceID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.tx.Validate()
			if tt.wantErr != nil {
				require.False(t, valid)
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.True(t, valid)
				require.NoError(t, err)
			}
		})
	}
}
