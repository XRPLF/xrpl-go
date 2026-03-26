//go:build cgo

package mptcrypto_test

import (
	"math"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/stretchr/testify/require"
)

// testAccountID returns a deterministic 20-byte account ID for testing.
func testAccountID(seed byte) [mptcrypto.AccountIDSize]byte {
	var id [mptcrypto.AccountIDSize]byte
	for i := range id {
		id[i] = seed + byte(i)
	}
	return id
}

// testIssuanceID returns a deterministic 24-byte issuance ID for testing.
func testIssuanceID() [mptcrypto.IssuanceIDSize]byte {
	var id [mptcrypto.IssuanceIDSize]byte
	for i := range id {
		id[i] = byte(i + 0x10)
	}
	return id
}

// region ElGamal
func TestGenerateKeypair(t *testing.T) {
	priv, pub, err := mptcrypto.GenerateKeypair()
	require.NoError(t, err)
	require.NotEqual(t, [mptcrypto.PrivKeySize]byte{}, priv, "privkey is all zeros")
	// compressed secp256k1 pubkey starts with 0x02 or 0x03
	require.Contains(t, []byte{0x02, 0x03}, pub[0], "unexpected pubkey prefix: 0x%02x", pub[0])
}

func TestGenerateBlindingFactor(t *testing.T) {
	bf1, err := mptcrypto.GenerateBlindingFactor()
	require.NoError(t, err)
	require.NotEqual(t, [mptcrypto.BlindingFactorSize]byte{}, bf1, "blinding factor is all zeros")

	// two calls should produce different values (non-deterministic RNG)
	bf2, err := mptcrypto.GenerateBlindingFactor()
	require.NoError(t, err)
	require.NotEqual(t, bf1, bf2, "two consecutive blinding factors are identical")
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	tests := []struct {
		name   string
		amount uint64
		// skipOnDecryptErr skips the subtest instead of failing when the C
		// library cannot recover the plaintext (BSGS table limitation).
		skipOnDecryptErr bool
	}{
		{"pass - zero", 0, false},
		{"pass - small value", 42, false},
		{"pass - one million", 1_000_000, false},
		{"pass - max uint64", math.MaxUint64, true},
	}

	priv, pub, err := mptcrypto.GenerateKeypair()
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf, err := mptcrypto.GenerateBlindingFactor()
			require.NoError(t, err)

			ct, err := mptcrypto.EncryptAmount(tt.amount, pub, bf)
			require.NoError(t, err)
			require.NotEqual(t, [mptcrypto.CiphertextSize]byte{}, ct, "ciphertext is all zeros")

			got, err := mptcrypto.DecryptAmount(ct, priv)
			if err != nil && tt.skipOnDecryptErr {
				t.Skipf("DecryptAmount not supported for this value: %v", err)
			}
			require.NoError(t, err)
			require.Equal(t, tt.amount, got)
		})
	}
}

// endregion

// region Context hashes
type contextHashFn func() ([mptcrypto.HashOutputSize]byte, error)

func TestContextHashes(t *testing.T) {
	account := testAccountID(0x01)
	account2 := testAccountID(0x20)
	iss := testIssuanceID()

	tests := []struct {
		name  string
		hash  contextHashFn
		other contextHashFn
	}{
		{
			"pass - Convert",
			func() ([mptcrypto.HashOutputSize]byte, error) { return mptcrypto.ConvertContextHash(account, iss, 1) },
			func() ([mptcrypto.HashOutputSize]byte, error) { return mptcrypto.ConvertContextHash(account, iss, 2) },
		},
		{
			"pass - ConvertBack",
			func() ([mptcrypto.HashOutputSize]byte, error) {
				return mptcrypto.ConvertBackContextHash(account, iss, 1, 1)
			},
			func() ([mptcrypto.HashOutputSize]byte, error) {
				return mptcrypto.ConvertBackContextHash(account, iss, 1, 2)
			},
		},
		{
			"pass - Send",
			func() ([mptcrypto.HashOutputSize]byte, error) {
				return mptcrypto.SendContextHash(account, iss, 1, account2, 1)
			},
			func() ([mptcrypto.HashOutputSize]byte, error) {
				return mptcrypto.SendContextHash(account, iss, 1, testAccountID(0x30), 1)
			},
		},
		{
			"pass - Clawback",
			func() ([mptcrypto.HashOutputSize]byte, error) {
				return mptcrypto.ClawbackContextHash(account, iss, 1, account2)
			},
			func() ([mptcrypto.HashOutputSize]byte, error) {
				return mptcrypto.ClawbackContextHash(account, iss, 1, testAccountID(0x30))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := tt.hash()
			require.NoError(t, err)
			require.NotEqual(t, [mptcrypto.HashOutputSize]byte{}, hash)

			// deterministic
			hash2, err := tt.hash()
			require.NoError(t, err)
			require.Equal(t, hash, hash2)

			// different input -> different hash
			hash3, err := tt.other()
			require.NoError(t, err)
			require.NotEqual(t, hash, hash3)
		})
	}
}

// endregion

// region Pedersen commitment
func TestPedersenCommitment(t *testing.T) {
	bf, err := mptcrypto.GenerateBlindingFactor()
	require.NoError(t, err)

	t.Run("pass - valid prefix and deterministic", func(t *testing.T) {
		commit, err := mptcrypto.PedersenCommitment(42, bf)
		require.NoError(t, err)
		require.Contains(t, []byte{0x02, 0x03}, commit[0], "unexpected commitment prefix: 0x%02x", commit[0])

		commit2, err := mptcrypto.PedersenCommitment(42, bf)
		require.NoError(t, err)
		require.Equal(t, commit, commit2)
	})

	t.Run("pass - different amounts produce different commitments", func(t *testing.T) {
		commit, err := mptcrypto.PedersenCommitment(42, bf)
		require.NoError(t, err)
		commit2, err := mptcrypto.PedersenCommitment(99, bf)
		require.NoError(t, err)
		require.NotEqual(t, commit, commit2)
	})
}

// endregion

// region Proof generation
func TestConvertProofRoundtrip(t *testing.T) {
	priv, pub, err := mptcrypto.GenerateKeypair()
	require.NoError(t, err)

	ctxHash, err := mptcrypto.ConvertContextHash(testAccountID(0x01), testIssuanceID(), 1)
	require.NoError(t, err)

	proof, err := mptcrypto.GenerateConvertProof(pub, priv, ctxHash)
	require.NoError(t, err)

	t.Run("pass - valid proof verifies", func(t *testing.T) {
		require.NotEqual(t, [mptcrypto.SchnorrProofSize]byte{}, proof)
		err := mptcrypto.VerifyConvertProof(proof, pub, ctxHash)
		require.NoError(t, err)
	})

	t.Run("fail - wrong key rejected", func(t *testing.T) {
		_, wrongPub, err := mptcrypto.GenerateKeypair()
		require.NoError(t, err)
		err = mptcrypto.VerifyConvertProof(proof, wrongPub, ctxHash)
		require.Error(t, err)
	})
}

func TestConvertBackProofRoundtrip(t *testing.T) {
	priv, pub, err := mptcrypto.GenerateKeypair()
	require.NoError(t, err)

	totalAmount := uint64(100)
	bf, err := mptcrypto.GenerateBlindingFactor()
	require.NoError(t, err)

	ct, err := mptcrypto.EncryptAmount(totalAmount, pub, bf)
	require.NoError(t, err)

	totalCommit, err := mptcrypto.PedersenCommitment(totalAmount, bf)
	require.NoError(t, err)

	convertBackAmount := uint64(30)

	balanceCommit, err := mptcrypto.ComputeConvertBackRemainder(totalCommit, convertBackAmount)
	require.NoError(t, err)

	ctxHash, err := mptcrypto.ConvertBackContextHash(testAccountID(0x01), testIssuanceID(), 1, 1)
	require.NoError(t, err)

	balanceParams := mptcrypto.PedersenProofParams{
		Commitment:     totalCommit,
		Amount:         totalAmount,
		Ciphertext:     ct,
		BlindingFactor: bf,
	}

	proof, err := mptcrypto.GenerateConvertBackProof(priv, pub, ctxHash, convertBackAmount, balanceParams)
	require.NoError(t, err)

	require.NotEqual(t, [mptcrypto.ConvertBackProofSize]byte{}, proof)
	err = mptcrypto.VerifyConvertBackProof(proof, pub, ct, balanceCommit, convertBackAmount, ctxHash)
	require.NoError(t, err)
}

func TestClawbackProofRoundtrip(t *testing.T) {
	priv, pub, err := mptcrypto.GenerateKeypair()
	require.NoError(t, err)

	amount := uint64(42)
	bf, err := mptcrypto.GenerateBlindingFactor()
	require.NoError(t, err)

	ct, err := mptcrypto.EncryptAmount(amount, pub, bf)
	require.NoError(t, err)

	ctxHash, err := mptcrypto.ClawbackContextHash(testAccountID(0x01), testIssuanceID(), 1, testAccountID(0x20))
	require.NoError(t, err)

	proof, err := mptcrypto.GenerateClawbackProof(priv, pub, ctxHash, amount, ct)
	require.NoError(t, err)

	require.NotEqual(t, [mptcrypto.EqualityProofSize]byte{}, proof)
	err = mptcrypto.VerifyClawbackProof(proof, amount, pub, ct, ctxHash)
	require.NoError(t, err)
}

func TestSendProofRoundtrip(t *testing.T) {
	tests := []struct {
		name        string
		withAuditor bool
	}{
		{"pass - 3 participants", false},
		{"pass - 4 participants with auditor", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			senderPriv, senderPub, err := mptcrypto.GenerateKeypair()
			require.NoError(t, err)
			_, issuerPub, err := mptcrypto.GenerateKeypair()
			require.NoError(t, err)
			_, destPub, err := mptcrypto.GenerateKeypair()
			require.NoError(t, err)

			balanceAmount := uint64(100)
			balanceBF, err := mptcrypto.GenerateBlindingFactor()
			require.NoError(t, err)
			senderBalanceCT, err := mptcrypto.EncryptAmount(balanceAmount, senderPub, balanceBF)
			require.NoError(t, err)
			balanceCommit, err := mptcrypto.PedersenCommitment(balanceAmount, balanceBF)
			require.NoError(t, err)

			sendAmount := uint64(30)
			txBF, err := mptcrypto.GenerateBlindingFactor()
			require.NoError(t, err)

			senderAmountCT, err := mptcrypto.EncryptAmount(sendAmount, senderPub, txBF)
			require.NoError(t, err)
			issuerAmountCT, err := mptcrypto.EncryptAmount(sendAmount, issuerPub, txBF)
			require.NoError(t, err)
			destAmountCT, err := mptcrypto.EncryptAmount(sendAmount, destPub, txBF)
			require.NoError(t, err)

			participants := []mptcrypto.Participant{
				{PubKey: senderPub, Ciphertext: senderAmountCT},
				{PubKey: issuerPub, Ciphertext: issuerAmountCT},
				{PubKey: destPub, Ciphertext: destAmountCT},
			}

			if tt.withAuditor {
				_, auditorPub, err := mptcrypto.GenerateKeypair()
				require.NoError(t, err)
				auditorAmountCT, err := mptcrypto.EncryptAmount(sendAmount, auditorPub, txBF)
				require.NoError(t, err)
				participants = append(participants, mptcrypto.Participant{
					PubKey: auditorPub, Ciphertext: auditorAmountCT,
				})
			}

			amountCommit, err := mptcrypto.PedersenCommitment(sendAmount, txBF)
			require.NoError(t, err)

			amountParams := mptcrypto.PedersenProofParams{
				Commitment:     amountCommit,
				Amount:         sendAmount,
				Ciphertext:     senderAmountCT,
				BlindingFactor: txBF,
			}
			balanceParams := mptcrypto.PedersenProofParams{
				Commitment:     balanceCommit,
				Amount:         balanceAmount,
				Ciphertext:     senderBalanceCT,
				BlindingFactor: balanceBF,
			}

			ctxHash, err := mptcrypto.SendContextHash(testAccountID(0x01), testIssuanceID(), 1, testAccountID(0x20), 1)
			require.NoError(t, err)

			proof, err := mptcrypto.GenerateSendProof(senderPriv, sendAmount, participants, txBF, ctxHash, amountParams, balanceParams)
			require.NoError(t, err)
			require.NotEmpty(t, proof)

			err = mptcrypto.VerifySendProof(proof, participants, senderBalanceCT, amountCommit, balanceCommit, ctxHash)
			require.NoError(t, err)
		})
	}
}

func TestAmountLinkageProofRoundtrip(t *testing.T) {
	_, pub, err := mptcrypto.GenerateKeypair()
	require.NoError(t, err)

	amount := uint64(42)
	bf, err := mptcrypto.GenerateBlindingFactor()
	require.NoError(t, err)

	ct, err := mptcrypto.EncryptAmount(amount, pub, bf)
	require.NoError(t, err)

	commit, err := mptcrypto.PedersenCommitment(amount, bf)
	require.NoError(t, err)

	ctxHash, err := mptcrypto.ConvertContextHash(testAccountID(0x01), testIssuanceID(), 1)
	require.NoError(t, err)

	params := mptcrypto.PedersenProofParams{
		Commitment:     commit,
		Amount:         amount,
		Ciphertext:     ct,
		BlindingFactor: bf,
	}

	proof, err := mptcrypto.GenerateAmountLinkageProof(pub, bf, ctxHash, params)
	require.NoError(t, err)
	require.NotEqual(t, [mptcrypto.PedersenLinkSize]byte{}, proof)

	err = mptcrypto.VerifyAmountLinkage(proof, ct, pub, commit, ctxHash)
	require.NoError(t, err)
}

func TestBalanceLinkageProofRoundtrip(t *testing.T) {
	priv, pub, err := mptcrypto.GenerateKeypair()
	require.NoError(t, err)

	amount := uint64(100)
	bf, err := mptcrypto.GenerateBlindingFactor()
	require.NoError(t, err)

	ct, err := mptcrypto.EncryptAmount(amount, pub, bf)
	require.NoError(t, err)

	commit, err := mptcrypto.PedersenCommitment(amount, bf)
	require.NoError(t, err)

	ctxHash, err := mptcrypto.ConvertContextHash(testAccountID(0x01), testIssuanceID(), 1)
	require.NoError(t, err)

	params := mptcrypto.PedersenProofParams{
		Commitment:     commit,
		Amount:         amount,
		Ciphertext:     ct,
		BlindingFactor: bf,
	}

	proof, err := mptcrypto.GenerateBalanceLinkageProof(priv, pub, ctxHash, params)
	require.NoError(t, err)
	require.NotEqual(t, [mptcrypto.PedersenLinkSize]byte{}, proof)

	err = mptcrypto.VerifyBalanceLinkage(proof, ct, pub, commit, ctxHash)
	require.NoError(t, err)
}

// endregion

// region Internal component verifiers
func TestVerifyRevealedAmount(t *testing.T) {
	amount := uint64(42)
	bf, err := mptcrypto.GenerateBlindingFactor()
	require.NoError(t, err)

	_, holderPub, err := mptcrypto.GenerateKeypair()
	require.NoError(t, err)
	_, issuerPub, err := mptcrypto.GenerateKeypair()
	require.NoError(t, err)

	holderCT, err := mptcrypto.EncryptAmount(amount, holderPub, bf)
	require.NoError(t, err)
	issuerCT, err := mptcrypto.EncryptAmount(amount, issuerPub, bf)
	require.NoError(t, err)

	holder := mptcrypto.Participant{PubKey: holderPub, Ciphertext: holderCT}
	issuer := mptcrypto.Participant{PubKey: issuerPub, Ciphertext: issuerCT}

	t.Run("pass - without auditor", func(t *testing.T) {
		err := mptcrypto.VerifyRevealedAmount(amount, bf, holder, issuer, nil)
		require.NoError(t, err)
	})

	t.Run("pass - with auditor", func(t *testing.T) {
		_, auditorPub, err := mptcrypto.GenerateKeypair()
		require.NoError(t, err)
		auditorCT, err := mptcrypto.EncryptAmount(amount, auditorPub, bf)
		require.NoError(t, err)
		auditor := &mptcrypto.Participant{PubKey: auditorPub, Ciphertext: auditorCT}

		err = mptcrypto.VerifyRevealedAmount(amount, bf, holder, issuer, auditor)
		require.NoError(t, err)
	})

	t.Run("fail - wrong amount", func(t *testing.T) {
		err := mptcrypto.VerifyRevealedAmount(99, bf, holder, issuer, nil)
		require.Error(t, err)
	})
}

// endregion

// region Utilities
func TestGetSendProofSize(t *testing.T) {
	t.Run("pass - positive size for 2 participants", func(t *testing.T) {
		size := mptcrypto.GetSendProofSize(2)
		require.Greater(t, size, 0)
	})

	t.Run("pass - more participants produce larger proof", func(t *testing.T) {
		size2 := mptcrypto.GetSendProofSize(2)
		size3 := mptcrypto.GetSendProofSize(3)
		require.Greater(t, size3, size2)
	})
}

func TestComputeConvertBackRemainder(t *testing.T) {
	bf, err := mptcrypto.GenerateBlindingFactor()
	require.NoError(t, err)

	commit, err := mptcrypto.PedersenCommitment(100, bf)
	require.NoError(t, err)

	remainder, err := mptcrypto.ComputeConvertBackRemainder(commit, 30)
	require.NoError(t, err)
	require.Contains(t, []byte{0x02, 0x03}, remainder[0], "unexpected prefix: 0x%02x", remainder[0])
	require.NotEqual(t, commit, remainder)
}

// endregion
