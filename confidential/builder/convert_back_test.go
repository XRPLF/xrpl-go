//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package builder

import (
	"math"
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	ledgerentries "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/stretchr/testify/require"
)

// TestConvertBackBaseValidation verifies shared malformed-input validation through both entry points.
func TestConvertBackBaseValidation(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	cases := []struct {
		name    string
		base    BuildConvertBackParams
		wantErr error
	}{
		{name: "fail - missing account", base: BuildConvertBackParams{IssuanceID: testIssuanceID, Amount: 1, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: kp.PubKeyHex}, wantErr: ErrMissingAccount},
		{name: "fail - invalid account", base: BuildConvertBackParams{Account: "notanaddress", IssuanceID: testIssuanceID, Amount: 1, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: kp.PubKeyHex}, wantErr: ErrInvalidAccount},
		{name: "fail - ACCOUNT_ZERO account", base: BuildConvertBackParams{Account: zeroClassicAccount, IssuanceID: testIssuanceID}, wantErr: ErrInvalidAccount},
		{name: "fail - missing issuance ID", base: BuildConvertBackParams{Account: testAccount, Amount: 1, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: kp.PubKeyHex}, wantErr: ErrMissingIssuanceID},
		{name: "fail - invalid issuance ID (not hex)", base: BuildConvertBackParams{Account: testAccount, IssuanceID: strings.Repeat("GG", 24), Amount: 1, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: kp.PubKeyHex}, wantErr: ErrInvalidIssuanceID},
		{name: "fail - invalid issuance ID (wrong length)", base: BuildConvertBackParams{Account: testAccount, IssuanceID: "aabb", Amount: 1, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: kp.PubKeyHex}, wantErr: ErrInvalidIssuanceID},
		{name: "fail - zero amount", base: BuildConvertBackParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: 0, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: kp.PubKeyHex}, wantErr: ErrZeroAmount},
		{name: "fail - account is the issuance issuer", base: BuildConvertBackParams{Account: testAccount, IssuanceID: testIssuerIssuanceID, Amount: 1, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: kp.PubKeyHex}, wantErr: ErrIssuerNotAllowed},
		{name: "fail - amount above protocol maximum", base: BuildConvertBackParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: math.MaxUint64, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: kp.PubKeyHex}, wantErr: ErrAmountTooLarge},
		{name: "fail - missing holder priv key", base: BuildConvertBackParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: 1, HolderPubKey: kp.PubKeyHex}, wantErr: ErrMissingHolderKey},
		{name: "fail - invalid holder priv key (not hex)", base: BuildConvertBackParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: 1, HolderPrivKey: strings.Repeat("ZZ", 32), HolderPubKey: kp.PubKeyHex}, wantErr: ErrInvalidPrivKey},
		{name: "fail - invalid holder priv key (wrong length)", base: BuildConvertBackParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: 1, HolderPrivKey: "aabb", HolderPubKey: kp.PubKeyHex}, wantErr: ErrInvalidPrivKey},
		{name: "fail - missing holder pub key", base: BuildConvertBackParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: 1, HolderPrivKey: kp.PrivKeyHex}, wantErr: ErrMissingHolderKey},
		{name: "fail - invalid holder pub key (not hex)", base: BuildConvertBackParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: 1, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: strings.Repeat("ZZ", 33)}, wantErr: ErrInvalidPubKey},
		{name: "fail - invalid holder pub key (wrong length)", base: BuildConvertBackParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: 1, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: "aabb"}, wantErr: ErrInvalidPubKey},
	}

	t.Run("fail - validation PrepareConvertBack", func(t *testing.T) {
		issKP, err := elgamal.GenerateKeypair()
		require.NoError(t, err)
		bf, err := elgamal.GenerateBlindingFactor()
		require.NoError(t, err)
		ct, err := elgamal.Encrypt(100, kp.PubKeyHex, bf)
		require.NoError(t, err)

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := PrepareConvertBack(ConvertBackParams{
					BuildConvertBackParams: tc.base,
					IssuerPubKey:           issKP.PubKeyHex,
					CurrentBalance:         100,
					CurrentBalanceCt:       ct,
				})
				require.ErrorIs(t, err, tc.wantErr)
			})
		}
	})

	t.Run("fail - validation BuildConvertBack", func(t *testing.T) {
		q := &mockQuerier{}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := BuildConvertBack(q, tc.base)
				require.ErrorIs(t, err, tc.wantErr)
			})
		}
	})
}

func TestPrepareConvertBack_Pass(t *testing.T) {
	const currentBalance uint64 = 1000
	const withdrawAmount uint64 = 100

	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	auditorKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	// Simulate existing balance state.
	balanceBF, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	balanceCt, err := elgamal.Encrypt(currentBalance, holderKP.PubKeyHex, balanceBF)
	require.NoError(t, err)

	result, err := PrepareConvertBack(ConvertBackParams{
		BuildConvertBackParams: BuildConvertBackParams{
			Account:       testAccount,
			IssuanceID:    testIssuanceID,
			Amount:        withdrawAmount,
			HolderPrivKey: holderKP.PrivKeyHex,
			HolderPubKey:  holderKP.PubKeyHex,
		},
		IssuerPubKey:     issuerKP.PubKeyHex,
		AuditorPubKey:    auditorKP.PubKeyHex,
		Sequence:         1,
		BalanceVersion:   0,
		CurrentBalance:   currentBalance,
		CurrentBalanceCt: balanceCt,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, transaction.ConfidentialMPTConvertBackTx, result.TxType())
	require.NotNil(t, result.AuditorEncryptedAmount)

	expectedHolderCt, err := elgamal.Encrypt(withdrawAmount, holderKP.PubKeyHex, result.BlindingFactor)
	require.NoError(t, err)
	require.Equal(t, expectedHolderCt, result.HolderEncryptedAmount)
	expectedIssuerCt, err := elgamal.Encrypt(withdrawAmount, issuerKP.PubKeyHex, result.BlindingFactor)
	require.NoError(t, err)
	require.Equal(t, expectedIssuerCt, result.IssuerEncryptedAmount)
	expectedAuditorCt, err := elgamal.Encrypt(withdrawAmount, auditorKP.PubKeyHex, result.BlindingFactor)
	require.NoError(t, err)
	require.Equal(t, expectedAuditorCt, *result.AuditorEncryptedAmount)

	// Verify the linkage + range proof cryptographically.
	ctxHash, err := proof.ConvertBackContextHash(testAccount, testIssuanceID, uint32(1), uint32(0))
	require.NoError(t, err)
	err = proof.VerifyConvertBackProof(result.ZKProof, holderKP.PubKeyHex, balanceCt, result.BalanceCommitment, withdrawAmount, ctxHash)
	require.NoError(t, err)

	ok, err := result.Validate()
	require.NoError(t, err)
	require.True(t, ok)
}

func TestPrepareConvertBack_MaxAmount(t *testing.T) {
	const maxAmount = uint64(math.MaxInt64)

	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	balanceBF, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	balanceCt, err := elgamal.Encrypt(maxAmount, holderKP.PubKeyHex, balanceBF)
	require.NoError(t, err)

	result, err := PrepareConvertBack(ConvertBackParams{
		BuildConvertBackParams: BuildConvertBackParams{
			Account:       testAccount,
			IssuanceID:    testIssuanceID,
			Amount:        maxAmount,
			HolderPrivKey: holderKP.PrivKeyHex,
			HolderPubKey:  holderKP.PubKeyHex,
		},
		IssuerPubKey:     issuerKP.PubKeyHex,
		Sequence:         1,
		CurrentBalance:   maxAmount,
		CurrentBalanceCt: balanceCt,
	})
	require.NoError(t, err)
	require.Len(t, result.ZKProof, transaction.ConvertBackProofLen)

	ctxHash, err := proof.ConvertBackContextHash(testAccount, testIssuanceID, 1, 0)
	require.NoError(t, err)
	require.NoError(t, proof.VerifyConvertBackProof(
		result.ZKProof,
		holderKP.PubKeyHex,
		balanceCt,
		result.BalanceCommitment,
		maxAmount,
		ctxHash,
	))
	valid, err := result.Validate()
	require.NoError(t, err)
	require.True(t, valid)
}

func TestPrepareConvertBack_MismatchedPrivateKey(t *testing.T) {
	const currentBalance uint64 = 100

	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	differentKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	balanceBF, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	balanceCt, err := elgamal.Encrypt(currentBalance, holderKP.PubKeyHex, balanceBF)
	require.NoError(t, err)

	_, err = PrepareConvertBack(ConvertBackParams{
		BuildConvertBackParams: BuildConvertBackParams{
			Account:       testAccount,
			IssuanceID:    testIssuanceID,
			Amount:        1,
			HolderPrivKey: differentKP.PrivKeyHex,
			HolderPubKey:  holderKP.PubKeyHex,
		},
		IssuerPubKey:     issuerKP.PubKeyHex,
		Sequence:         1,
		CurrentBalance:   currentBalance,
		CurrentBalanceCt: balanceCt,
	})
	require.ErrorIs(t, err, ErrCryptoFailed)
	require.ErrorIs(t, err, proof.ErrProofGenerationFailed)
}

func TestPrepareConvertBack_FailInsufficientBalance(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	ct, err := elgamal.Encrypt(100, kp.PubKeyHex, bf)
	require.NoError(t, err)

	_, err = PrepareConvertBack(ConvertBackParams{
		BuildConvertBackParams: BuildConvertBackParams{
			Account:       testAccount,
			IssuanceID:    testIssuanceID,
			Amount:        200, // More than CurrentBalance (100)
			HolderPrivKey: kp.PrivKeyHex,
			HolderPubKey:  kp.PubKeyHex,
		},
		IssuerPubKey:     issKP.PubKeyHex,
		Sequence:         1,
		CurrentBalance:   100,
		CurrentBalanceCt: ct,
	})
	require.ErrorIs(t, err, ErrInsufficientBalance)
}

func TestPrepareConvertBack_FailValidation(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	ct, err := elgamal.Encrypt(100, kp.PubKeyHex, bf)
	require.NoError(t, err)

	base := BuildConvertBackParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: 1, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: kp.PubKeyHex}

	tests := []struct {
		name    string
		params  ConvertBackParams
		wantErr error
	}{
		{name: "fail - missing issuer pub key", params: ConvertBackParams{BuildConvertBackParams: base, CurrentBalance: 100, CurrentBalanceCt: ct}, wantErr: ErrMissingIssuerKey},
		{name: "fail - invalid issuer pub key (not hex)", params: ConvertBackParams{BuildConvertBackParams: base, IssuerPubKey: strings.Repeat("ZZ", 33), CurrentBalance: 100, CurrentBalanceCt: ct}, wantErr: ErrInvalidPubKey},
		{name: "fail - invalid issuer pub key (wrong length)", params: ConvertBackParams{BuildConvertBackParams: base, IssuerPubKey: "aabb", CurrentBalance: 100, CurrentBalanceCt: ct}, wantErr: ErrInvalidPubKey},
		{name: "fail - invalid auditor pub key (not hex)", params: ConvertBackParams{BuildConvertBackParams: base, IssuerPubKey: issKP.PubKeyHex, AuditorPubKey: strings.Repeat("ZZ", 33), CurrentBalance: 100, CurrentBalanceCt: ct}, wantErr: ErrInvalidPubKey},
		{name: "fail - invalid auditor pub key (wrong length)", params: ConvertBackParams{BuildConvertBackParams: base, IssuerPubKey: issKP.PubKeyHex, AuditorPubKey: "aabb", CurrentBalance: 100, CurrentBalanceCt: ct}, wantErr: ErrInvalidPubKey},
		{name: "fail - missing balance state", params: ConvertBackParams{BuildConvertBackParams: base, IssuerPubKey: issKP.PubKeyHex, CurrentBalance: 100}, wantErr: ErrMissingSenderState},
		{name: "fail - invalid ciphertext (not hex)", params: ConvertBackParams{BuildConvertBackParams: base, IssuerPubKey: issKP.PubKeyHex, CurrentBalance: 100, CurrentBalanceCt: strings.Repeat("ZZ", 66)}, wantErr: ErrInvalidCiphertext},
		{name: "fail - invalid ciphertext (wrong length)", params: ConvertBackParams{BuildConvertBackParams: base, IssuerPubKey: issKP.PubKeyHex, CurrentBalance: 100, CurrentBalanceCt: "aabb"}, wantErr: ErrInvalidCiphertext},
		{name: "fail - missing final sequence", params: ConvertBackParams{BuildConvertBackParams: base, IssuerPubKey: issKP.PubKeyHex, CurrentBalance: 100, CurrentBalanceCt: ct}, wantErr: ErrMissingSequence},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PrepareConvertBack(tt.params)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestBuildConvertBack_BalanceRange(t *testing.T) {
	const currentBalance uint64 = 1000

	tests := []struct {
		name         string
		balanceRange elgamal.AmountRange
		wantErr      bool
	}{
		{name: "pass - balance in range", balanceRange: elgamal.AmountRange{Low: currentBalance, High: currentBalance}},
		{name: "fail - balance outside range", balanceRange: elgamal.AmountRange{Low: 0, High: currentBalance - 1}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			holderKP, q := newBalanceLedgerFixture(t, 3, 1, currentBalance)
			holderMPTIndex, err := xrplhash.MPToken(testIssuanceID, testAccount)
			require.NoError(t, err)
			q.entries[holderMPTIndex]["HolderEncryptionKey"] = strings.ToUpper(holderKP.PubKeyHex)
			result, err := BuildConvertBack(q, BuildConvertBackParams{
				Account:       testAccount,
				IssuanceID:    testIssuanceID,
				Amount:        100,
				HolderPrivKey: holderKP.PrivKeyHex,
				HolderPubKey:  holderKP.PubKeyHex,
				BalanceRange:  tt.balanceRange,
			})
			if tt.wantErr {
				require.ErrorIs(t, err, ErrCryptoFailed)
				require.ErrorIs(t, err, elgamal.ErrDecryptFailed)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, uint32(3), result.Sequence)

			balanceCiphertext, ok := q.entries[holderMPTIndex]["ConfidentialBalanceSpending"].(string)
			require.True(t, ok)
			contextHash, err := proof.ConvertBackContextHash(testAccount, testIssuanceID, result.Sequence, 1)
			require.NoError(t, err)
			require.NoError(t, proof.VerifyConvertBackProof(result.ZKProof, holderKP.PubKeyHex, balanceCiphertext, result.BalanceCommitment, 100, contextHash))
		})
	}
}

func TestBuildConvertBackRejectsMismatchedHolderKey(t *testing.T) {
	const currentBalance uint64 = 1000

	_, q := newBalanceLedgerFixture(t, 3, 1, currentBalance)
	differentKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	_, err = BuildConvertBack(q, BuildConvertBackParams{
		Account:       testAccount,
		IssuanceID:    testIssuanceID,
		Amount:        100,
		HolderPrivKey: differentKP.PrivKeyHex,
		HolderPubKey:  differentKP.PubKeyHex,
		BalanceRange:  elgamal.AmountRange{Low: currentBalance, High: currentBalance},
	})
	require.ErrorIs(t, err, ErrKeyMismatch)
	require.ErrorContains(t, err, "holder key")
}

func TestBuildConvertBackRejectsMissingHolderState(t *testing.T) {
	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)
	mptokenIndex, err := xrplhash.MPToken(testIssuanceID, testAccount)
	require.NoError(t, err)

	q := &mockQuerier{
		accountSeq: 5,
		entries: map[string]ledgerentries.FlatLedgerObject{
			issuanceIndex: buildIssuanceEntry(issuerKP.PubKeyHex, ""),
			mptokenIndex:  buildMPTokenEntry(mptokenFields{}),
		},
	}

	_, err = BuildConvertBack(q, BuildConvertBackParams{
		Account:       testAccount,
		IssuanceID:    testIssuanceID,
		Amount:        100,
		HolderPrivKey: holderKP.PrivKeyHex,
		HolderPubKey:  holderKP.PubKeyHex,
		BalanceRange:  elgamal.AmountRange{Low: 0, High: 1000},
	})
	require.ErrorIs(t, err, ErrMissingSenderState)
	require.NotErrorIs(t, err, ErrCryptoFailed)
}

// TestBuildConvertBackRejectsAmountAboveOutstanding pins the supply check. rippled fails a
// convert back above ConfidentialOutstandingAmount with tecINSUFFICIENT_FUNDS, and the
// issuance is already in hand, so the holder's MPToken is never read.
func TestBuildConvertBackRejectsAmountAboveOutstanding(t *testing.T) {
	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)

	issuance := buildIssuanceEntry(issuerKP.PubKeyHex, "")
	issuance["ConfidentialOutstandingAmount"] = "99"
	q := &mockQuerier{
		accountSeq: 5,
		entries: map[string]ledgerentries.FlatLedgerObject{
			issuanceIndex: issuance,
		},
	}

	_, err = BuildConvertBack(q, BuildConvertBackParams{
		Account:       testAccount,
		IssuanceID:    testIssuanceID,
		Amount:        100,
		HolderPrivKey: holderKP.PrivKeyHex,
		HolderPubKey:  holderKP.PubKeyHex,
		BalanceRange:  elgamal.AmountRange{Low: 0, High: 1000},
	})
	require.ErrorIs(t, err, ErrAmountExceedsOutstanding)
	require.NotErrorIs(t, err, ErrMPTokenNotFound)
}

func TestBuildConvertBack_InvalidRangeBeforeLedgerQueries(t *testing.T) {
	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	q := &mockQuerier{accountErr: ErrLedgerQuery}
	_, err = BuildConvertBack(q, BuildConvertBackParams{
		Account:       testAccount,
		IssuanceID:    testIssuanceID,
		Amount:        1,
		HolderPrivKey: holderKP.PrivKeyHex,
		HolderPubKey:  holderKP.PubKeyHex,
		BalanceRange:  elgamal.AmountRange{Low: 2, High: 1},
	})
	require.ErrorIs(t, err, elgamal.ErrInvalidAmountRange)
	require.NotErrorIs(t, err, ErrLedgerQuery)
	require.Zero(t, q.queryCalls, "invalid ranges must fail before ledger queries")
}
