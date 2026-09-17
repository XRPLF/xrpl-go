package transaction

import (
	"strings"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestConfidentialMPTPointValidation(t *testing.T) {
	offCurvePoint := "02" + strings.Repeat("00", 32)
	invalidPrefix := "04" + testCompressedPoint1[2:]

	tests := []struct {
		name  string
		value string
		valid bool
	}{
		{name: "valid point", value: testCompressedPoint1, valid: true},
		{name: "alternate valid point", value: testCompressedPoint2, valid: true},
		{name: "valid odd-y point", value: "03" + testCompressedPoint1[2:], valid: true},
		{name: "invalid prefix", value: invalidPrefix},
		{name: "off curve", value: offCurvePoint},
		{name: "invalid hex", value: "02" + strings.Repeat("ZZ", 32)},
		{name: "wrong length", value: "02"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.valid, IsValidCompressedEncryptionKey(test.value))
			require.Equal(t, test.valid, IsValidCommitment(test.value))
		})
	}
}

func TestConfidentialMPTCiphertextValidation(t *testing.T) {
	offCurvePoint := "02" + strings.Repeat("00", 32)

	tests := []struct {
		name  string
		value string
		valid bool
	}{
		{name: "valid points", value: testCiphertext, valid: true},
		{name: "invalid first point", value: offCurvePoint + testCompressedPoint1},
		{name: "invalid second point", value: testCompressedPoint1 + offCurvePoint},
		{name: "invalid hex", value: strings.Repeat("ZZ", CiphertextLen/2)},
		{name: "wrong length", value: testCompressedPoint1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.valid, IsValidCiphertext(test.value))
		})
	}
}

// TestConfidentialMPTProtocolLengths pins each hex-encoded length constant to the byte
// size XLS-96 specifies, and drives the corresponding validator with a
// genuinely valid blob of that exact length plus the same blob one byte short and one
// byte long. Using a valid fixture rather than filler is what makes the rejection
// assertions meaningful: filler hex is not a curve point, so the point and ciphertext
// rows would otherwise be rejected for their content and pass regardless of the length.
func TestConfidentialMPTProtocolLengths(t *testing.T) {
	tests := []struct {
		name      string
		hexLen    int
		specBytes int
		valid     string
		isValid   func(string) bool
	}{
		{name: "blinding factor", hexLen: BlindingFactorLen, specBytes: 32, valid: testBlindingFactor, isValid: IsValidBlindingFactor},
		{name: "schnorr proof", hexLen: SchnorrProofLen, specBytes: 64, valid: strings.Repeat("AB", 64), isValid: IsValidSchnorrProof},
		{name: "send proof", hexLen: SendProofLen, specBytes: 192 + 754, valid: strings.Repeat("AB", 946), isValid: IsValidSendProof},
		{name: "convert back proof", hexLen: ConvertBackProofLen, specBytes: 128 + 688, valid: strings.Repeat("AB", 816), isValid: IsValidConvertBackProof},
		{name: "clawback proof", hexLen: ClawbackProofLen, specBytes: 64, valid: strings.Repeat("AB", 64), isValid: IsValidClawbackProof},
		{name: "compressed point", hexLen: CompressedPointLen, specBytes: 33, valid: testCompressedPoint1, isValid: IsValidCompressedEncryptionKey},
		{name: "ciphertext", hexLen: CiphertextLen, specBytes: 66, valid: testCiphertext, isValid: IsValidCiphertext},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.specBytes*2, test.hexLen, "hex length must be twice the XLS-96 byte size")
			require.Len(t, test.valid, test.hexLen, "fixture must be exactly the constant length")

			require.True(t, test.isValid(test.valid), "must accept a valid blob of the exact length")
			require.False(t, test.isValid(test.valid[:test.hexLen-2]), "must reject one byte short")
			require.False(t, test.isValid(test.valid+"AB"), "must reject one byte long")
		})
	}
}

// TestValidateConfidentialMPTBaseRejectsUndecodableAccount pins the fail-closed behaviour
// directly, because BaseTx.Validate rejects an undecodable Account before any of the five
// Validate methods reach this helper.
func TestValidateConfidentialMPTBaseRejectsUndecodableAccount(t *testing.T) {
	tx := &BaseTx{Account: "not-an-xrpl-address"}

	accountID, err := validateConfidentialMPTBase(tx)

	require.ErrorIs(t, err, ErrInvalidAccount)
	require.Nil(t, accountID)
}

// TestConfidentialMPTRejectsDuplicateAccountTag pins that every confidential MPT
// transaction rejects an Account pairing a tagged X-address with an explicit SourceTag.
// BaseTx.Validate deliberately allows the pairing, because client autofill resolves it,
// so each of these models has to reach validateConfidentialMPTBase for the rule to apply.
func TestConfidentialMPTRejectsDuplicateAccountTag(t *testing.T) {
	taggedAccount, err := addresscodec.ClassicAddressToXAddress(testAddrAccount, 42, true, false)
	require.NoError(t, err)

	base := func(txType TxType) BaseTx {
		return BaseTx{
			Account:         types.Address(taggedAccount),
			TransactionType: txType,
			Fee:             types.XRPCurrencyAmount(10),
			SourceTag:       7,
		}
	}

	type validator interface{ Validate() (bool, error) }

	testcases := []struct {
		name string
		tx   validator
	}{
		{name: "convert", tx: &ConfidentialMPTConvert{BaseTx: base(ConfidentialMPTConvertTx)}},
		{name: "convert back", tx: &ConfidentialMPTConvertBack{BaseTx: base(ConfidentialMPTConvertBackTx)}},
		{name: "send", tx: &ConfidentialMPTSend{BaseTx: base(ConfidentialMPTSendTx)}},
		{name: "clawback", tx: &ConfidentialMPTClawback{BaseTx: base(ConfidentialMPTClawbackTx)}},
		{name: "merge inbox", tx: &ConfidentialMPTMergeInbox{BaseTx: base(ConfidentialMPTMergeInboxTx)}},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			valid, err := tc.tx.Validate()
			require.False(t, valid)
			// The pairing must surface before any transaction-specific field error,
			// so the model never does proof or issuance work on a doubled tag.
			require.ErrorIs(t, err, ErrDuplicateXAddressTag)
		})
	}
}
