package builder

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	ledgerentries "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/stretchr/testify/require"
)

// TestClawbackBaseValidation verifies all validateClawbackBase branches through both entry points.
func TestClawbackBaseValidation(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	cases := []struct {
		name    string
		base    BuildClawbackParams
		wantErr error
	}{
		{name: "fail - missing account", base: BuildClawbackParams{Holder: testDestination, IssuanceID: testIssuanceID, Amount: 1, IssuerPrivKey: kp.PrivKeyHex}, wantErr: ErrMissingAccount},
		{name: "fail - missing holder", base: BuildClawbackParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: 1, IssuerPrivKey: kp.PrivKeyHex}, wantErr: ErrMissingHolder},
		{name: "fail - missing issuance ID", base: BuildClawbackParams{Account: testAccount, Holder: testDestination, Amount: 1, IssuerPrivKey: kp.PrivKeyHex}, wantErr: ErrMissingIssuanceID},
		{name: "fail - self clawback", base: BuildClawbackParams{Account: testAccount, Holder: testAccount, IssuanceID: testIssuanceID, Amount: 1, IssuerPrivKey: kp.PrivKeyHex}, wantErr: ErrSelfClawback},
		{name: "fail - zero amount", base: BuildClawbackParams{Account: testAccount, Holder: testDestination, IssuanceID: testIssuanceID, Amount: 0, IssuerPrivKey: kp.PrivKeyHex}, wantErr: ErrZeroAmount},
		{name: "fail - missing issuer priv key", base: BuildClawbackParams{Account: testAccount, Holder: testDestination, IssuanceID: testIssuanceID, Amount: 1}, wantErr: ErrMissingIssuerKey},
	}

	t.Run("fail - validation PrepareClawback", func(t *testing.T) {
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := PrepareClawback(ClawbackParams{BuildClawbackParams: tc.base, IssuerPubKey: kp.PubKeyHex, IssuerCiphertext: "aa"})
				require.ErrorIs(t, err, tc.wantErr)
			})
		}
	})

	t.Run("fail - validation BuildClawback", func(t *testing.T) {
		q := &mockQuerier{}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := BuildClawback(q, tc.base)
				require.ErrorIs(t, err, tc.wantErr)
			})
		}
	})
}

func TestPrepareClawback_Pass(t *testing.T) {
	const amount uint64 = 500
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	issuerCt, err := elgamal.Encrypt(amount, issuerKP.PubKeyHex, bf)
	require.NoError(t, err)

	result, err := PrepareClawback(ClawbackParams{
		BuildClawbackParams: BuildClawbackParams{
			Account:       testAccount,
			Holder:        testDestination,
			IssuanceID:    testIssuanceID,
			Amount:        amount,
			IssuerPrivKey: issuerKP.PrivKeyHex,
		},
		IssuerPubKey:     issuerKP.PubKeyHex,
		IssuerCiphertext: issuerCt,
		Sequence:         1,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, transaction.ConfidentialMPTClawbackTx, result.TxType())
	require.NotEmpty(t, result.ZKProof)

	ctxHash, err := proof.ClawbackContextHash(testAccount, testIssuanceID, uint32(1), testDestination)
	require.NoError(t, err)
	err = proof.VerifyClawbackProof(result.ZKProof, amount, issuerKP.PubKeyHex, issuerCt, ctxHash)
	require.NoError(t, err)

	ok, err := result.Validate()
	require.NoError(t, err)
	require.True(t, ok)
}

func TestPrepareClawback_FailValidation(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	base := BuildClawbackParams{Account: testAccount, Holder: testDestination, IssuanceID: testIssuanceID, Amount: 1, IssuerPrivKey: kp.PrivKeyHex}

	tests := []struct {
		name    string
		params  ClawbackParams
		wantErr error
	}{
		{name: "fail - missing issuer pub key", params: ClawbackParams{BuildClawbackParams: base, IssuerCiphertext: "aa"}, wantErr: ErrMissingIssuerKey},
		{name: "fail - missing ciphertext", params: ClawbackParams{BuildClawbackParams: base, IssuerPubKey: kp.PubKeyHex}, wantErr: ErrMissingCiphertext},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PrepareClawback(tt.params)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestBuildClawback_Pass(t *testing.T) {
	const clawbackAmount uint64 = 500
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	issuerCt, err := elgamal.Encrypt(clawbackAmount, issuerKP.PubKeyHex, bf)
	require.NoError(t, err)

	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)
	mptokenIndex, err := xrplhash.MPToken(testIssuanceID, testDestination)
	require.NoError(t, err)

	q := &mockQuerier{
		accountSeq: 10,
		entries: map[string]ledgerentries.FlatLedgerObject{
			issuanceIndex: buildIssuanceEntry(issuerKP.PubKeyHex, ""),
			mptokenIndex:  buildMPTokenEntry("", "", 0, issuerCt),
		},
	}

	result, err := BuildClawback(q, BuildClawbackParams{
		Account:       testAccount,
		Holder:        testDestination,
		IssuanceID:    testIssuanceID,
		Amount:        clawbackAmount,
		IssuerPrivKey: issuerKP.PrivKeyHex,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, uint32(10), result.Sequence)
	require.NotEmpty(t, result.ZKProof)
}
