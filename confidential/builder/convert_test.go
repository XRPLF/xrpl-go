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

// TestConvertBaseValidation verifies all validateConvertBase branches through both entry points.
func TestConvertBaseValidation(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	cases := []struct {
		name    string
		base    BuildConvertParams
		wantErr error
	}{
		{name: "fail - missing account", base: BuildConvertParams{IssuanceID: testIssuanceID, Amount: 1, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: kp.PubKeyHex}, wantErr: ErrMissingAccount},
		{name: "fail - missing issuance ID", base: BuildConvertParams{Account: testAccount, Amount: 1, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: kp.PubKeyHex}, wantErr: ErrMissingIssuanceID},
		{name: "fail - missing holder key", base: BuildConvertParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: 1}, wantErr: ErrMissingHolderKey},
	}

	t.Run("fail - validation PrepareConvert", func(t *testing.T) {
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := PrepareConvert(ConvertParams{BuildConvertParams: tc.base, IssuerPubKey: kp.PubKeyHex})
				require.ErrorIs(t, err, tc.wantErr)
			})
		}
	})

	t.Run("fail - validation BuildConvert", func(t *testing.T) {
		q := &mockQuerier{}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := BuildConvert(q, tc.base)
				require.ErrorIs(t, err, tc.wantErr)
			})
		}
	})
}

func TestPrepareConvert_PassFirstTime(t *testing.T) {
	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	result, err := PrepareConvert(ConvertParams{
		BuildConvertParams: BuildConvertParams{
			Account:       testAccount,
			IssuanceID:    testIssuanceID,
			Amount:        1000,
			HolderPrivKey: holderKP.PrivKeyHex,
			HolderPubKey:  holderKP.PubKeyHex,
		},
		IssuerPubKey: issuerKP.PubKeyHex,
		Sequence:     1,
		FirstTime:    true,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, transaction.ConfidentialMPTConvertTx, result.TxType())

	// First time: key and proof must be set.
	require.NotNil(t, result.HolderEncryptionKey)
	require.Equal(t, holderKP.PubKeyHex, *result.HolderEncryptionKey)
	require.NotNil(t, result.ZKProof)

	// Verify the Schnorr proof cryptographically.
	ctxHash, err := proof.ConvertContextHash(testAccount, testIssuanceID, uint32(1))
	require.NoError(t, err)
	err = proof.VerifyConvertProof(*result.ZKProof, holderKP.PubKeyHex, ctxHash)
	require.NoError(t, err)

	// Transaction must validate.
	ok, err := result.Validate()
	require.NoError(t, err)
	require.True(t, ok)
}

func TestPrepareConvert_PassNotFirstTime(t *testing.T) {
	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	result, err := PrepareConvert(ConvertParams{
		BuildConvertParams: BuildConvertParams{
			Account:       testAccount,
			IssuanceID:    testIssuanceID,
			Amount:        500,
			HolderPrivKey: holderKP.PrivKeyHex,
			HolderPubKey:  holderKP.PubKeyHex,
		},
		IssuerPubKey: issuerKP.PubKeyHex,
		Sequence:     2,
		FirstTime:    false,
	})
	require.NoError(t, err)
	require.Nil(t, result.HolderEncryptionKey)
	require.Nil(t, result.ZKProof)

	ok, err := result.Validate()
	require.NoError(t, err)
	require.True(t, ok)
}

func TestPrepareConvert_PassWithAuditor(t *testing.T) {
	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	auditorKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	result, err := PrepareConvert(ConvertParams{
		BuildConvertParams: BuildConvertParams{
			Account:       testAccount,
			IssuanceID:    testIssuanceID,
			Amount:        100,
			HolderPrivKey: holderKP.PrivKeyHex,
			HolderPubKey:  holderKP.PubKeyHex,
		},
		IssuerPubKey:  issuerKP.PubKeyHex,
		AuditorPubKey: auditorKP.PubKeyHex,
		Sequence:      1,
		FirstTime:     false,
	})
	require.NoError(t, err)
	require.NotNil(t, result.AuditorEncryptedAmount)
	require.Len(t, *result.AuditorEncryptedAmount, 132)

	ok, err := result.Validate()
	require.NoError(t, err)
	require.True(t, ok)
}

func TestPrepareConvert_FailValidation(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	t.Run("fail - missing issuer pub key", func(t *testing.T) {
		_, err := PrepareConvert(ConvertParams{
			BuildConvertParams: BuildConvertParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: 1, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: kp.PubKeyHex},
		})
		require.ErrorIs(t, err, ErrMissingIssuerKey)
	})
}

func TestBuildConvert_PassFirstTime(t *testing.T) {
	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)

	q := &mockQuerier{
		accountSeq: 5,
		entries: map[string]ledgerentries.FlatLedgerObject{
			issuanceIndex: buildIssuanceEntry(issuerKP.PubKeyHex, ""),
		},
	}

	result, err := BuildConvert(q, BuildConvertParams{
		Account:       testAccount,
		IssuanceID:    testIssuanceID,
		Amount:        1000,
		HolderPrivKey: holderKP.PrivKeyHex,
		HolderPubKey:  holderKP.PubKeyHex,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, uint32(5), result.Sequence)
	require.NotNil(t, result.HolderEncryptionKey)
	require.NotNil(t, result.ZKProof)
}

func TestBuildConvert_PassNotFirstTime(t *testing.T) {
	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)
	mptokenIndex, err := xrplhash.MPToken(testIssuanceID, testAccount)
	require.NoError(t, err)

	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	balanceCt, err := elgamal.Encrypt(500, holderKP.PubKeyHex, bf)
	require.NoError(t, err)

	q := &mockQuerier{
		accountSeq: 7,
		entries: map[string]ledgerentries.FlatLedgerObject{
			issuanceIndex: buildIssuanceEntry(issuerKP.PubKeyHex, ""),
			mptokenIndex:  buildMPTokenEntry(holderKP.PubKeyHex, balanceCt, 0, ""),
		},
	}

	result, err := BuildConvert(q, BuildConvertParams{
		Account:       testAccount,
		IssuanceID:    testIssuanceID,
		Amount:        200,
		HolderPrivKey: holderKP.PrivKeyHex,
		HolderPubKey:  holderKP.PubKeyHex,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Nil(t, result.HolderEncryptionKey)
	require.Nil(t, result.ZKProof)
}

func TestBuildConvert_PassWithAuditor(t *testing.T) {
	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	auditorKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)

	q := &mockQuerier{
		accountSeq: 3,
		entries: map[string]ledgerentries.FlatLedgerObject{
			issuanceIndex: buildIssuanceEntry(issuerKP.PubKeyHex, auditorKP.PubKeyHex),
		},
	}

	result, err := BuildConvert(q, BuildConvertParams{
		Account:       testAccount,
		IssuanceID:    testIssuanceID,
		Amount:        100,
		HolderPrivKey: holderKP.PrivKeyHex,
		HolderPubKey:  holderKP.PubKeyHex,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.AuditorEncryptedAmount)
}
