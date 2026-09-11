package transaction

import (
	"strings"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

var testClawbackProof = strings.Repeat("AB", ClawbackProofLen/2)

func TestConfidentialMPTClawback_TxType(t *testing.T) {
	tx := &ConfidentialMPTClawback{}
	require.Equal(t, ConfidentialMPTClawbackTx, tx.TxType())
}

func TestConfidentialMPTClawback_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		tx       *ConfidentialMPTClawback
		expected FlatTransaction
	}{
		{
			name: "pass - all fields",
			tx: &ConfidentialMPTClawback{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:     types.XRPCurrencyAmount(12),
				},
				Holder:            "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID: "00000001D28B177E48D9A8D057E70F7E464B498367281B98",
				MPTAmount:         types.MPTPlainAmount(1000),
				ZKProof:           testClawbackProof,
			},
			expected: FlatTransaction{
				"Account":           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee":               "12",
				"TransactionType":   "ConfidentialMPTClawback",
				"Holder":            "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				"MPTokenIssuanceID": "00000001D28B177E48D9A8D057E70F7E464B498367281B98",
				"MPTAmount":         "1000",
				"ZKProof":           testClawbackProof,
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

func TestConfidentialMPTClawback_TaggedAccountEncoding(t *testing.T) {
	const (
		account   = "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"
		sourceTag = uint32(7)
	)
	taggedAccount, err := addresscodec.ClassicAddressToXAddress(account, sourceTag, true, false)
	require.NoError(t, err)
	tx := &ConfidentialMPTClawback{
		BaseTx:            BaseTx{Account: types.Address(taggedAccount), TransactionType: ConfidentialMPTClawbackTx, Fee: 12},
		Holder:            "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
		MPTokenIssuanceID: "00000001D28B177E48D9A8D057E70F7E464B498367281B98",
		MPTAmount:         1,
		ZKProof:           testClawbackProof,
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

func TestConfidentialMPTClawback_BinaryRoundTrip(t *testing.T) {
	const expected = "1200592400000004301A0000000000000064702540ABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABAB81145B812C9D57731E27A2DA8B1830195F88EF32A3B68B14B5F762798A53D543A014CAF8B297CFF8F2F937E80115000004C463C52827307480341125DA0577DEFC38405B0E3E"
	tx := &ConfidentialMPTClawback{
		BaseTx: BaseTx{
			Account:         "r9LqNeG6qHxjeUocjvVki2XR35weJ9mZgQ",
			TransactionType: ConfidentialMPTClawbackTx,
			Sequence:        4,
		},
		MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
		Holder:            "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		MPTAmount:         100,
		ZKProof:           strings.Repeat("AB", ClawbackProofLen/2),
	}

	encoded, err := binarycodec.Encode(tx.Flatten())
	require.NoError(t, err)
	require.Equal(t, expected, encoded)
	decoded, err := binarycodec.Decode(encoded)
	require.NoError(t, err)
	require.Equal(t, map[string]any(tx.Flatten()), decoded)
}

func TestConfidentialMPTClawback_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tx      *ConfidentialMPTClawback
		wantErr error
	}{
		{
			name: "pass - valid transaction",
			tx: &ConfidentialMPTClawback{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTClawbackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Holder:            "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID: "00000001D28B177E48D9A8D057E70F7E464B498367281B98",
				MPTAmount:         types.MPTPlainAmount(1000),
				ZKProof:           testClawbackProof,
			},
			wantErr: nil,
		},
		{
			name: "pass - maximum amount",
			tx: &ConfidentialMPTClawback{
				BaseTx:            BaseTx{Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD", TransactionType: ConfidentialMPTClawbackTx, Fee: types.XRPCurrencyAmount(12)},
				Holder:            "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID: "00000001D28B177E48D9A8D057E70F7E464B498367281B98",
				MPTAmount:         types.MPTPlainAmount(types.MaxMPTAmount),
				ZKProof:           testClawbackProof,
			},
		},
		{
			name: "fail - empty MPTokenIssuanceID",
			tx: &ConfidentialMPTClawback{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTClawbackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Holder:            "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID: "",
				MPTAmount:         types.MPTPlainAmount(1000),
				ZKProof:           testClawbackProof,
			},
			wantErr: ErrConfidentialMPTInvalidIssuanceID,
		},
		{
			name: "fail - invalid Holder address",
			tx: &ConfidentialMPTClawback{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTClawbackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Holder:            "invalidAddress",
				MPTokenIssuanceID: "00000001D28B177E48D9A8D057E70F7E464B498367281B98",
				MPTAmount:         types.MPTPlainAmount(1000),
				ZKProof:           testClawbackProof,
			},
			wantErr: ErrConfidentialClawbackInvalidHolder,
		},
		{
			// ACCOUNT_ZERO decodes cleanly in either form but can never hold the MPToken
			// a clawback targets, so the field sentinel names it and the cause survives.
			name: "fail - ACCOUNT_ZERO Holder",
			tx: &ConfidentialMPTClawback{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTClawbackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Holder:            testAddrZero,
				MPTokenIssuanceID: "00000001D28B177E48D9A8D057E70F7E464B498367281B98",
				MPTAmount:         types.MPTPlainAmount(1000),
				ZKProof:           testClawbackProof,
			},
			wantErr: ErrZeroAccountID,
		},
		{
			name: "fail - Holder same as Account",
			tx: &ConfidentialMPTClawback{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTClawbackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Holder:            "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				MPTokenIssuanceID: "00000001D28B177E48D9A8D057E70F7E464B498367281B98",
				MPTAmount:         types.MPTPlainAmount(1000),
				ZKProof:           testClawbackProof,
			},
			wantErr: ErrConfidentialClawbackSelfClawback,
		},
		{
			name: "fail - Holder is the X-address of Account",
			tx: &ConfidentialMPTClawback{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTClawbackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Holder:            testXAddressAccount,
				MPTokenIssuanceID: "00000001D28B177E48D9A8D057E70F7E464B498367281B98",
				MPTAmount:         types.MPTPlainAmount(1000),
				ZKProof:           testClawbackProof,
			},
			wantErr: ErrConfidentialClawbackSelfClawback,
		},
		{
			name: "fail - Holder X-address carries a tag",
			tx: &ConfidentialMPTClawback{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTClawbackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Holder:            testXAddressTaggedDestination,
				MPTokenIssuanceID: "00000001D28B177E48D9A8D057E70F7E464B498367281B98",
				MPTAmount:         types.MPTPlainAmount(1000),
				ZKProof:           testClawbackProof,
			},
			wantErr: ErrConfidentialClawbackHolderTagNotAllowed,
		},
		{
			name: "fail - zero MPTAmount",
			tx: &ConfidentialMPTClawback{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTClawbackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Holder:            "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID: "00000001D28B177E48D9A8D057E70F7E464B498367281B98",
				MPTAmount:         types.MPTPlainAmount(0),
				ZKProof:           testClawbackProof,
			},
			wantErr: ErrConfidentialClawbackInvalidAmount,
		},
		{
			name: "fail - short ZKProof",
			tx: &ConfidentialMPTClawback{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTClawbackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Holder:            "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID: "00000001D28B177E48D9A8D057E70F7E464B498367281B98",
				MPTAmount:         types.MPTPlainAmount(1000),
				ZKProof:           "A1B2C3D4",
			},
			wantErr: ErrConfidentialClawbackBadProof,
		},
		{
			name: "fail - empty ZKProof",
			tx: &ConfidentialMPTClawback{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTClawbackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Holder:            "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID: "00000001D28B177E48D9A8D057E70F7E464B498367281B98",
				MPTAmount:         types.MPTPlainAmount(1000),
				ZKProof:           "",
			},
			wantErr: ErrConfidentialClawbackBadProof,
		},
		{
			name: "fail - invalid hex ZKProof",
			tx: &ConfidentialMPTClawback{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTClawbackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Holder:            "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID: "00000001D28B177E48D9A8D057E70F7E464B498367281B98",
				MPTAmount:         types.MPTPlainAmount(1000),
				ZKProof:           strings.Repeat("AA", ClawbackProofLen/2-1) + "ZZ",
			},
			wantErr: ErrConfidentialClawbackBadProof,
		},
		{
			name: "fail - long ZKProof",
			tx: &ConfidentialMPTClawback{
				BaseTx:            BaseTx{Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD", TransactionType: ConfidentialMPTClawbackTx, Fee: types.XRPCurrencyAmount(12)},
				Holder:            "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID: "00000001D28B177E48D9A8D057E70F7E464B498367281B98",
				MPTAmount:         types.MPTPlainAmount(1000),
				ZKProof:           strings.Repeat("AA", ClawbackProofLen/2+1),
			},
			wantErr: ErrConfidentialClawbackBadProof,
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
