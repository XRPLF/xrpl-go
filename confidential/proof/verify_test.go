//go:build cgo

package proof_test

import (
	"encoding/hex"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/commitment"
	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	"github.com/stretchr/testify/require"
)

func TestVerifyRevealedAmountWithoutAuditor(t *testing.T) {
	const amount uint64 = 42

	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)

	holderCt, err := elgamal.Encrypt(amount, holderKP.PubKeyHex, bf)
	require.NoError(t, err)
	issuerCt, err := elgamal.Encrypt(amount, issuerKP.PubKeyHex, bf)
	require.NoError(t, err)

	holder := proof.HexParticipant{PubKeyHex: holderKP.PubKeyHex, CiphertextHex: holderCt}
	issuer := proof.HexParticipant{PubKeyHex: issuerKP.PubKeyHex, CiphertextHex: issuerCt}

	err = proof.VerifyRevealedAmount(amount, bf, holder, issuer, nil)
	require.NoError(t, err)
}

func TestVerifyRevealedAmountWithAuditor(t *testing.T) {
	const amount uint64 = 42

	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	auditorKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)

	holderCt, err := elgamal.Encrypt(amount, holderKP.PubKeyHex, bf)
	require.NoError(t, err)
	issuerCt, err := elgamal.Encrypt(amount, issuerKP.PubKeyHex, bf)
	require.NoError(t, err)
	auditorCt, err := elgamal.Encrypt(amount, auditorKP.PubKeyHex, bf)
	require.NoError(t, err)

	holder := proof.HexParticipant{PubKeyHex: holderKP.PubKeyHex, CiphertextHex: holderCt}
	issuer := proof.HexParticipant{PubKeyHex: issuerKP.PubKeyHex, CiphertextHex: issuerCt}
	auditor := proof.HexParticipant{PubKeyHex: auditorKP.PubKeyHex, CiphertextHex: auditorCt}

	err = proof.VerifyRevealedAmount(amount, bf, holder, issuer, &auditor)
	require.NoError(t, err)
}

func TestVerifyRevealedAmountWrongAmount(t *testing.T) {
	const amount uint64 = 42

	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)

	holderCt, err := elgamal.Encrypt(amount, holderKP.PubKeyHex, bf)
	require.NoError(t, err)
	issuerCt, err := elgamal.Encrypt(amount, issuerKP.PubKeyHex, bf)
	require.NoError(t, err)

	holder := proof.HexParticipant{PubKeyHex: holderKP.PubKeyHex, CiphertextHex: holderCt}
	issuer := proof.HexParticipant{PubKeyHex: issuerKP.PubKeyHex, CiphertextHex: issuerCt}

	err = proof.VerifyRevealedAmount(999, bf, holder, issuer, nil)
	require.ErrorIs(t, err, proof.ErrProofVerificationFailed)
}

// setupLinkageScenario creates the crypto state needed to test linkage and range proof verifiers.
func setupLinkageScenario(t *testing.T, amount, balance uint64) (
	senderKP elgamal.Keypair,
	txBF, balanceBF string,
	amountCt, balanceCt string,
	amountCommitHex, balanceCommitHex string,
	ctxHash string,
) {
	t.Helper()

	senderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	txBF, err = elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	balanceBF, err = elgamal.GenerateBlindingFactor()
	require.NoError(t, err)

	amountCt, err = elgamal.Encrypt(amount, senderKP.PubKeyHex, txBF)
	require.NoError(t, err)
	balanceCt, err = elgamal.Encrypt(balance, senderKP.PubKeyHex, balanceBF)
	require.NoError(t, err)

	amountCommitHex, err = commitment.Create(amount, txBF)
	require.NoError(t, err)
	balanceCommitHex, err = commitment.Create(balance, balanceBF)
	require.NoError(t, err)

	ctxHash, err = proof.SendContextHash(testAccount, testIssuanceID, 1, testDest, 0)
	require.NoError(t, err)
	return
}

func TestVerifyAmountLinkageRoundtrip(t *testing.T) {
	const amount uint64 = 500
	const balance uint64 = 1000

	senderKP, txBF, _, amountCt, _, amountCommitHex, _, ctxHash := setupLinkageScenario(t, amount, balance)

	// Generate the linkage proof at the mptcrypto level (exposed for testing).
	pubBytes, _ := hex.DecodeString(senderKP.PubKeyHex)
	bfBytes, _ := hex.DecodeString(txBF)
	ctBytes, _ := hex.DecodeString(amountCt)
	commitBytes, _ := hex.DecodeString(amountCommitHex)
	hashBytes, _ := hex.DecodeString(ctxHash)

	var pub [mptcrypto.PubKeySize]byte
	var bf [mptcrypto.BlindingFactorSize]byte
	var ct [mptcrypto.CiphertextSize]byte
	var com [mptcrypto.CommitmentSize]byte
	var hash [mptcrypto.HashOutputSize]byte
	copy(pub[:], pubBytes)
	copy(bf[:], bfBytes)
	copy(ct[:], ctBytes)
	copy(com[:], commitBytes)
	copy(hash[:], hashBytes)

	params := mptcrypto.PedersenProofParams{
		Commitment:     com,
		Amount:         amount,
		Ciphertext:     ct,
		BlindingFactor: bf,
	}

	proofBytes, err := mptcrypto.GenerateAmountLinkageProof(pub, bf, hash, params)
	require.NoError(t, err)

	proofHex := hex.EncodeToString(proofBytes[:])
	err = proof.VerifyAmountLinkage(proofHex, amountCt, senderKP.PubKeyHex, amountCommitHex, ctxHash)
	require.NoError(t, err)
}

func TestVerifyBalanceLinkageRoundtrip(t *testing.T) {
	const amount uint64 = 500
	const balance uint64 = 1000

	senderKP, _, balanceBF, _, balanceCt, _, balanceCommitHex, ctxHash := setupLinkageScenario(t, amount, balance)

	// Generate the balance linkage proof at the mptcrypto level.
	privBytes, _ := hex.DecodeString(senderKP.PrivKeyHex)
	pubBytes, _ := hex.DecodeString(senderKP.PubKeyHex)
	bfBytes, _ := hex.DecodeString(balanceBF)
	ctBytes, _ := hex.DecodeString(balanceCt)
	commitBytes, _ := hex.DecodeString(balanceCommitHex)
	hashBytes, _ := hex.DecodeString(ctxHash)

	var priv [mptcrypto.PrivKeySize]byte
	var pub [mptcrypto.PubKeySize]byte
	var bf [mptcrypto.BlindingFactorSize]byte
	var ct [mptcrypto.CiphertextSize]byte
	var com [mptcrypto.CommitmentSize]byte
	var hash [mptcrypto.HashOutputSize]byte
	copy(priv[:], privBytes)
	copy(pub[:], pubBytes)
	copy(bf[:], bfBytes)
	copy(ct[:], ctBytes)
	copy(com[:], commitBytes)
	copy(hash[:], hashBytes)

	params := mptcrypto.PedersenProofParams{
		Commitment:     com,
		Amount:         balance,
		Ciphertext:     ct,
		BlindingFactor: bf,
	}

	proofBytes, err := mptcrypto.GenerateBalanceLinkageProof(priv, pub, hash, params)
	require.NoError(t, err)

	proofHex := hex.EncodeToString(proofBytes[:])
	err = proof.VerifyBalanceLinkage(proofHex, balanceCt, senderKP.PubKeyHex, balanceCommitHex, ctxHash)
	require.NoError(t, err)
}

func TestVerifyRevealedAmountInvalidInputs(t *testing.T) {
	tests := []struct {
		name    string
		fn      func() error
		wantErr error
	}{
		{
			name: "fail - bad blinding factor",
			fn: func() error {
				return proof.VerifyRevealedAmount(42, "bad",
					proof.HexParticipant{PubKeyHex: zeroHex(33), CiphertextHex: zeroHex(66)},
					proof.HexParticipant{PubKeyHex: zeroHex(33), CiphertextHex: zeroHex(66)},
					nil)
			},
			wantErr: proof.ErrInvalidBlindingFactor,
		},
		{
			name: "fail - bad holder pubkey",
			fn: func() error {
				return proof.VerifyRevealedAmount(42, zeroHex(32),
					proof.HexParticipant{PubKeyHex: "zz", CiphertextHex: zeroHex(66)},
					proof.HexParticipant{PubKeyHex: zeroHex(33), CiphertextHex: zeroHex(66)},
					nil)
			},
			wantErr: proof.ErrInvalidPubKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestVerifyLinkageInvalidInputs(t *testing.T) {
	tests := []struct {
		name    string
		fn      func() error
		wantErr error
	}{
		{
			name: "fail - amount linkage bad proof",
			fn: func() error {
				return proof.VerifyAmountLinkage("zz", zeroHex(66), zeroHex(33), "02"+zeroHex(32), zeroHex(32))
			},
			wantErr: proof.ErrInvalidProof,
		},
		{
			name: "fail - amount linkage bad ciphertext",
			fn: func() error {
				return proof.VerifyAmountLinkage(zeroHex(195), "zz", zeroHex(33), "02"+zeroHex(32), zeroHex(32))
			},
			wantErr: proof.ErrInvalidCiphertext,
		},
		{
			name: "fail - balance linkage bad pubkey",
			fn: func() error {
				return proof.VerifyBalanceLinkage(zeroHex(195), zeroHex(66), "zz", "02"+zeroHex(32), zeroHex(32))
			},
			wantErr: proof.ErrInvalidPubKey,
		},
		{
			name: "fail - balance linkage bad commitment",
			fn: func() error {
				return proof.VerifyBalanceLinkage(zeroHex(195), zeroHex(66), zeroHex(33), "zz", zeroHex(32))
			},
			wantErr: proof.ErrInvalidCommitment,
		},
		{
			name: "fail - balance linkage bad ctx hash",
			fn: func() error {
				return proof.VerifyBalanceLinkage(zeroHex(195), zeroHex(66), zeroHex(33), "02"+zeroHex(32), "zz")
			},
			wantErr: proof.ErrInvalidContextHash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestVerifyEqualityProofInvalidInputs(t *testing.T) {
	tests := []struct {
		name    string
		fn      func() error
		wantErr error
	}{
		{
			name: "fail - bad proof hex",
			fn: func() error {
				return proof.VerifyEqualityProof("zzzz", nil, zeroHex(32))
			},
			wantErr: proof.ErrInvalidProof,
		},
		{
			name: "fail - bad ctx hash",
			fn: func() error {
				return proof.VerifyEqualityProof(zeroHex(32), nil, "zz")
			},
			wantErr: proof.ErrInvalidContextHash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestVerifySendRangeProofInvalidInputs(t *testing.T) {
	tests := []struct {
		name    string
		fn      func() error
		wantErr error
	}{
		{
			name: "fail - bad proof",
			fn: func() error {
				return proof.VerifySendRangeProof("zz", "02"+zeroHex(32), "02"+zeroHex(32), zeroHex(32))
			},
			wantErr: proof.ErrInvalidProof,
		},
		{
			name: "fail - bad amount commitment",
			fn: func() error {
				return proof.VerifySendRangeProof(zeroHex(754), "zz", "02"+zeroHex(32), zeroHex(32))
			},
			wantErr: proof.ErrInvalidCommitment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
