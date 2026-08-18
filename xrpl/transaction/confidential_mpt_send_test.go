package transaction

import (
	"strings"
	"testing"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	bctypes "github.com/Peersyst/xrpl-go/binary-codec/types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

// Test helpers for ConfidentialMPTSend.
var (
	// testAuditorCiphertext is a fourth distinct ciphertext, for the auditor copy.
	testAuditorCiphertext = testCompressedPoint1 + testCompressedPoint3
	testSendProof         = strings.Repeat("11", SendProofLen/2)
	testCredentialID      = strings.Repeat("AA", 32)
)

func TestConfidentialMPTSend_TxType(t *testing.T) {
	tx := &ConfidentialMPTSend{}
	require.Equal(t, ConfidentialMPTSendTx, tx.TxType())
}

func TestConfidentialMPTSend_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		tx       *ConfidentialMPTSend
		expected FlatTransaction
	}{
		{
			name: "pass - without optional fields",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:     types.XRPCurrencyAmount(12),
				},
				Destination:                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID:          "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				SenderEncryptedAmount:      testCiphertext,
				DestinationEncryptedAmount: testCiphertext2,
				IssuerEncryptedAmount:      testCiphertext3,
				ZKProof:                    testSendProof,
				BalanceCommitment:          testCompressedPoint1,
				AmountCommitment:           testCompressedPoint3,
			},
			expected: FlatTransaction{
				"Account":                    "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee":                        "12",
				"TransactionType":            "ConfidentialMPTSend",
				"Destination":                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				"MPTokenIssuanceID":          "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				"SenderEncryptedAmount":      testCiphertext,
				"DestinationEncryptedAmount": testCiphertext2,
				"IssuerEncryptedAmount":      testCiphertext3,
				"ZKProof":                    testSendProof,
				"BalanceCommitment":          testCompressedPoint1,
				"AmountCommitment":           testCompressedPoint3,
			},
		},
		{
			name: "pass - with optional fields",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:     types.XRPCurrencyAmount(12),
				},
				Destination:                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				DestinationTag:             types.DestinationTag(0),
				MPTokenIssuanceID:          "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				SenderEncryptedAmount:      testCiphertext,
				DestinationEncryptedAmount: testCiphertext2,
				IssuerEncryptedAmount:      testCiphertext3,
				ZKProof:                    testSendProof,
				BalanceCommitment:          testCompressedPoint1,
				AmountCommitment:           testCompressedPoint3,
				AuditorEncryptedAmount:     types.HexBlob(testAuditorCiphertext),
				CredentialIDs:              types.CredentialIDs{testCredentialID},
			},
			expected: FlatTransaction{
				"Account":                    "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee":                        "12",
				"TransactionType":            "ConfidentialMPTSend",
				"Destination":                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				"DestinationTag":             uint32(0),
				"MPTokenIssuanceID":          "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				"SenderEncryptedAmount":      testCiphertext,
				"DestinationEncryptedAmount": testCiphertext2,
				"IssuerEncryptedAmount":      testCiphertext3,
				"ZKProof":                    testSendProof,
				"BalanceCommitment":          testCompressedPoint1,
				"AmountCommitment":           testCompressedPoint3,
				"AuditorEncryptedAmount":     testAuditorCiphertext,
				"CredentialIDs":              []string{testCredentialID},
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

// TestConfidentialMPTSend_TaggedDestinationEncoding covers the case the DestinationTag
// duplicate check exists to keep working: a tagged X-address Destination with no explicit
// DestinationTag expands to a classic Destination plus the embedded tag on encode.
func TestConfidentialMPTSend_TaggedDestinationEncoding(t *testing.T) {
	const (
		destination    = "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP"
		destinationTag = uint32(7)
	)
	tx := newValidConfidentialMPTSend()
	tx.Destination = types.Address(testXAddressTaggedDestination)

	valid, err := tx.Validate()
	require.True(t, valid)
	require.NoError(t, err)

	encoded, err := binarycodec.Encode(tx.Flatten())
	require.NoError(t, err)
	decoded, err := binarycodec.Decode(encoded)
	require.NoError(t, err)
	require.Equal(t, destination, decoded["Destination"])
	require.EqualValues(t, destinationTag, decoded["DestinationTag"])
}

// TestConfidentialMPTSend_BinaryRoundTrip pins the wire encoding of every ConfidentialMPTSend
// field, including DestinationTag. rippled declares {sfDestinationTag, SoeOptional} for
// ttCONFIDENTIAL_MPT_SEND in transactions.macro, though the XLS-96 field table omits it.
func TestConfidentialMPTSend_BinaryRoundTrip(t *testing.T) {
	const (
		prefix = "12005824000000032E000030397025C3F1"
		suffix = "70274202ABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABAB03CDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCD70284202ABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABAB03CDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCD70294202ABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABAB03CDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCD702B4202ABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABAB03CDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCD702D2102ABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABAB702E2102ABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABAB81145B812C9D57731E27A2DA8B1830195F88EF32A3B68314B5F762798A53D543A014CAF8B297CFF8F2F937E8051320EA85602C1B41F6F1F5E83C0E6B87142FB8957BD209469E4CC347BA2D0C26F66A0115000004C463C52827307480341125DA0577DEFC38405B0E3E"
	)
	point := "02" + strings.Repeat("AB", 32)
	ciphertext := point + "03" + strings.Repeat("CD", 32)
	tx := &ConfidentialMPTSend{
		BaseTx: BaseTx{
			Account:         "r9LqNeG6qHxjeUocjvVki2XR35weJ9mZgQ",
			TransactionType: ConfidentialMPTSendTx,
			Sequence:        3,
		},
		MPTokenIssuanceID:          "000004C463C52827307480341125DA0577DEFC38405B0E3E",
		Destination:                "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		DestinationTag:             types.DestinationTag(12_345),
		SenderEncryptedAmount:      ciphertext,
		DestinationEncryptedAmount: ciphertext,
		IssuerEncryptedAmount:      ciphertext,
		AuditorEncryptedAmount:     types.HexBlob(ciphertext),
		ZKProof:                    strings.Repeat("AB", SendProofLen/2),
		AmountCommitment:           point,
		BalanceCommitment:          point,
		CredentialIDs:              types.CredentialIDs{"EA85602C1B41F6F1F5E83C0E6B87142FB8957BD209469E4CC347BA2D0C26F66A"},
	}
	expected := prefix + strings.Repeat("AB", SendProofLen/2) + suffix

	encoded, err := binarycodec.Encode(tx.Flatten())
	require.NoError(t, err)
	require.Equal(t, expected, encoded)
	decoded, err := binarycodec.Decode(encoded)
	require.NoError(t, err)
	require.Equal(t, map[string]any(tx.Flatten()), decoded)
}

func newValidConfidentialMPTSend() *ConfidentialMPTSend {
	return &ConfidentialMPTSend{
		BaseTx: BaseTx{
			Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
			TransactionType: ConfidentialMPTSendTx,
			Fee:             types.XRPCurrencyAmount(12),
		},
		Destination:                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
		MPTokenIssuanceID:          "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
		SenderEncryptedAmount:      testCiphertext,
		DestinationEncryptedAmount: testCiphertext2,
		IssuerEncryptedAmount:      testCiphertext3,
		ZKProof:                    testSendProof,
		BalanceCommitment:          testCompressedPoint1,
		AmountCommitment:           testCompressedPoint3,
	}
}

func TestConfidentialMPTSend_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(tx *ConfidentialMPTSend)
		wantErr error
	}{
		{name: "valid transaction", mutate: func(*ConfidentialMPTSend) {}},
		{name: "credential IDs", mutate: func(tx *ConfidentialMPTSend) { tx.CredentialIDs = types.CredentialIDs{testCredentialID} }},
		{name: "auditor encrypted amount", mutate: func(tx *ConfidentialMPTSend) { tx.AuditorEncryptedAmount = types.HexBlob(testAuditorCiphertext) }},
		{name: "X-address destination without tag", mutate: func(tx *ConfidentialMPTSend) { tx.Destination = types.Address(testXAddressDestination) }},
		{name: "empty issuance ID", mutate: func(tx *ConfidentialMPTSend) { tx.MPTokenIssuanceID = "" }, wantErr: ErrConfidentialMPTInvalidIssuanceID},
		{name: "invalid destination", mutate: func(tx *ConfidentialMPTSend) { tx.Destination = "invalidAddress" }, wantErr: ErrConfidentialSendInvalidDestination},
		{name: "destination same as account", mutate: func(tx *ConfidentialMPTSend) { tx.Destination = tx.Account }, wantErr: ErrConfidentialSendSelfSend},
		// ACCOUNT_ZERO decodes cleanly in either form but can never hold the MPToken a
		// destination must hold, so the field sentinel names it and the cause survives.
		{name: "ACCOUNT_ZERO destination", mutate: func(tx *ConfidentialMPTSend) { tx.Destination = testAddrZero }, wantErr: ErrConfidentialSendInvalidDestination},
		{name: "ACCOUNT_ZERO destination reports its cause", mutate: func(tx *ConfidentialMPTSend) { tx.Destination = testAddrZero }, wantErr: ErrZeroAccountID},
		{name: "destination X-address of account", mutate: func(tx *ConfidentialMPTSend) { tx.Destination = types.Address(testXAddressAccount) }, wantErr: ErrConfidentialSendSelfSend},
		{name: "tagged X-address account with SourceTag", mutate: func(tx *ConfidentialMPTSend) {
			tx.Account = types.Address(testXAddressTaggedAccount)
			tx.SourceTag = 9
		}, wantErr: bctypes.ErrDuplicateXAddressTag},
		{name: "tagged X-address account without SourceTag", mutate: func(tx *ConfidentialMPTSend) {
			tx.Account = types.Address(testXAddressTaggedAccount)
		}},
		{name: "tagged X-address destination with DestinationTag", mutate: func(tx *ConfidentialMPTSend) {
			tag := uint32(7)
			tx.Destination = types.Address(testXAddressTaggedDestination)
			tx.DestinationTag = &tag
		}, wantErr: bctypes.ErrDuplicateXAddressTag},
		{name: "empty sender ciphertext", mutate: func(tx *ConfidentialMPTSend) { tx.SenderEncryptedAmount = "" }, wantErr: ErrConfidentialSendInvalidCiphertext},
		{name: "empty destination ciphertext", mutate: func(tx *ConfidentialMPTSend) { tx.DestinationEncryptedAmount = "" }, wantErr: ErrConfidentialSendInvalidCiphertext},
		{name: "empty issuer ciphertext", mutate: func(tx *ConfidentialMPTSend) { tx.IssuerEncryptedAmount = "" }, wantErr: ErrConfidentialSendInvalidCiphertext},
		{name: "short sender ciphertext", mutate: func(tx *ConfidentialMPTSend) { tx.SenderEncryptedAmount = "AABB" }, wantErr: ErrConfidentialSendInvalidCiphertext},
		{name: "short destination ciphertext", mutate: func(tx *ConfidentialMPTSend) { tx.DestinationEncryptedAmount = "BBCC" }, wantErr: ErrConfidentialSendInvalidCiphertext},
		{name: "short auditor ciphertext", mutate: func(tx *ConfidentialMPTSend) { tx.AuditorEncryptedAmount = types.HexBlob("AABB") }, wantErr: ErrConfidentialSendInvalidCiphertext},
		{name: "non-hex auditor ciphertext", mutate: func(tx *ConfidentialMPTSend) { tx.AuditorEncryptedAmount = types.HexBlob("not-hex!") }, wantErr: ErrConfidentialSendInvalidCiphertext},
		{name: "short proof", mutate: func(tx *ConfidentialMPTSend) { tx.ZKProof = strings.Repeat("AA", 10) }, wantErr: ErrConfidentialSendInvalidProof},
		{name: "non-hex proof", mutate: func(tx *ConfidentialMPTSend) { tx.ZKProof = strings.Repeat("AA", SendProofLen/2-1) + "ZZ" }, wantErr: ErrConfidentialSendInvalidProof},
		{name: "empty proof", mutate: func(tx *ConfidentialMPTSend) { tx.ZKProof = "" }, wantErr: ErrConfidentialSendInvalidProof},
		{name: "empty balance commitment", mutate: func(tx *ConfidentialMPTSend) { tx.BalanceCommitment = "" }, wantErr: ErrConfidentialSendInvalidCommitment},
		{name: "empty amount commitment", mutate: func(tx *ConfidentialMPTSend) { tx.AmountCommitment = "" }, wantErr: ErrConfidentialSendInvalidCommitment},
		{name: "short balance commitment", mutate: func(tx *ConfidentialMPTSend) { tx.BalanceCommitment = "EEFF" }, wantErr: ErrConfidentialSendInvalidCommitment},
		{name: "short amount commitment", mutate: func(tx *ConfidentialMPTSend) { tx.AmountCommitment = "FF11" }, wantErr: ErrConfidentialSendInvalidCommitment},
		{name: "invalid credential IDs", mutate: func(tx *ConfidentialMPTSend) { tx.CredentialIDs = types.CredentialIDs{"not-hex!"} }, wantErr: ErrInvalidCredentialIDs},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx := newValidConfidentialMPTSend()
			test.mutate(tx)

			valid, err := tx.Validate()

			if test.wantErr != nil {
				require.False(t, valid)
				require.ErrorIs(t, err, test.wantErr)
				return
			}
			require.True(t, valid)
			require.NoError(t, err)
		})
	}
}
