//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package mptcrypto_test

import (
	"fmt"
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
	require.NotEqual(t, mptcrypto.PrivateKey{}, priv, "privkey is all zeros")
	// compressed secp256k1 pubkey starts with 0x02 or 0x03
	require.Contains(t, []byte{0x02, 0x03}, pub[0], "unexpected pubkey prefix: 0x%02x", pub[0])
}

func TestGenerateBlindingFactor(t *testing.T) {
	bf1, err := mptcrypto.GenerateBlindingFactor()
	require.NoError(t, err)
	require.NotEqual(t, mptcrypto.BlindingFactor{}, bf1, "blinding factor is all zeros")

	// two calls should produce different values (non-deterministic RNG)
	bf2, err := mptcrypto.GenerateBlindingFactor()
	require.NoError(t, err)
	require.NotEqual(t, bf1, bf2, "two consecutive blinding factors are identical")
}

func TestDecryptAmountBounds(t *testing.T) {
	const amount uint64 = 42

	privateKey, publicKey, err := mptcrypto.GenerateKeypair()
	require.NoError(t, err)
	blindingFactor, err := mptcrypto.GenerateBlindingFactor()
	require.NoError(t, err)
	ciphertext, err := mptcrypto.EncryptAmount(amount, publicKey, blindingFactor)
	require.NoError(t, err)

	tests := []struct {
		name      string
		rangeLow  uint64
		rangeHigh uint64
		wantErr   bool
	}{
		{name: "pass - found at lower bound", rangeLow: amount, rangeHigh: 50},
		{name: "pass - found at upper bound", rangeLow: 40, rangeHigh: amount},
		{name: "pass - found inside range", rangeLow: 40, rangeHigh: 50},
		{name: "pass - single-value interval", rangeLow: amount, rangeHigh: amount},
		{name: "fail - amount below range", rangeLow: 43, rangeHigh: 50, wantErr: true},
		{name: "fail - amount above range", rangeLow: 0, rangeHigh: 41, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mptcrypto.DecryptAmount(ciphertext, privateKey, tt.rangeLow, tt.rangeHigh)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, amount, got)
		})
	}
}

func TestDecryptAmountInvalidRange(t *testing.T) {
	tests := []struct {
		name      string
		rangeLow  uint64
		rangeHigh uint64
	}{
		{name: "fail - low exceeds high", rangeLow: 2, rangeHigh: 1},
		{name: "fail - high is max uint64", rangeLow: 0, rangeHigh: math.MaxUint64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mptcrypto.DecryptAmount(mptcrypto.Ciphertext{}, mptcrypto.PrivateKey{}, tt.rangeLow, tt.rangeHigh)
			require.ErrorIs(t, err, mptcrypto.ErrInvalidAmountRange)
		})
	}
}

// endregion

// region Context hashes
type contextHashFn func() (mptcrypto.ContextHash, error)

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
			func() (mptcrypto.ContextHash, error) { return mptcrypto.ConvertContextHash(account, iss, 1) },
			func() (mptcrypto.ContextHash, error) { return mptcrypto.ConvertContextHash(account, iss, 2) },
		},
		{
			"pass - ConvertBack",
			func() (mptcrypto.ContextHash, error) {
				return mptcrypto.ConvertBackContextHash(account, iss, 1, 1)
			},
			func() (mptcrypto.ContextHash, error) {
				return mptcrypto.ConvertBackContextHash(account, iss, 1, 2)
			},
		},
		{
			"pass - Send",
			func() (mptcrypto.ContextHash, error) {
				return mptcrypto.SendContextHash(account, iss, 1, account2, 1)
			},
			func() (mptcrypto.ContextHash, error) {
				return mptcrypto.SendContextHash(account, iss, 1, testAccountID(0x30), 1)
			},
		},
		{
			"pass - Clawback",
			func() (mptcrypto.ContextHash, error) {
				return mptcrypto.ClawbackContextHash(account, iss, 1, account2)
			},
			func() (mptcrypto.ContextHash, error) {
				return mptcrypto.ClawbackContextHash(account, iss, 1, testAccountID(0x30))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := tt.hash()
			require.NoError(t, err)
			require.NotEqual(t, mptcrypto.ContextHash{}, hash)

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

	tests := []struct {
		name      string
		first     uint64
		second    uint64
		wantEqual bool
	}{
		{name: "pass - same inputs are deterministic", first: 42, second: 42, wantEqual: true},
		{name: "pass - different amounts", first: 42, second: 99},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first, err := mptcrypto.PedersenCommitment(tt.first, bf)
			require.NoError(t, err)
			require.Contains(t, []byte{0x02, 0x03}, first[0], "unexpected commitment prefix: 0x%02x", first[0])

			second, err := mptcrypto.PedersenCommitment(tt.second, bf)
			require.NoError(t, err)
			if tt.wantEqual {
				require.Equal(t, first, second)
				return
			}
			require.NotEqual(t, first, second)
		})
	}
}

// endregion

// region Proof generation
func TestConvertProofRoundtrip(t *testing.T) {
	priv, pub, err := mptcrypto.GenerateKeypair()
	require.NoError(t, err)
	_, wrongPub, err := mptcrypto.GenerateKeypair()
	require.NoError(t, err)

	ctxHash, err := mptcrypto.ConvertContextHash(testAccountID(0x01), testIssuanceID(), 1)
	require.NoError(t, err)

	proof, err := mptcrypto.GenerateConvertProof(pub, priv, ctxHash)
	require.NoError(t, err)
	require.NotEqual(t, [mptcrypto.SchnorrProofSize]byte{}, proof)

	tests := []struct {
		name    string
		pubKey  mptcrypto.PublicKey
		wantErr bool
	}{
		{name: "pass - valid proof verifies", pubKey: pub},
		{name: "fail - wrong key rejected", pubKey: wrongPub, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mptcrypto.VerifyConvertProof(proof, tt.pubKey, ctxHash)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
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
	err = mptcrypto.VerifyConvertBackProof(proof, pub, ct, totalCommit, convertBackAmount, ctxHash)
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

	require.NotEqual(t, [mptcrypto.CompactClawbackProofSize]byte{}, proof)
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
			_, destPub, err := mptcrypto.GenerateKeypair()
			require.NoError(t, err)
			_, issuerPub, err := mptcrypto.GenerateKeypair()
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
			destAmountCT, err := mptcrypto.EncryptAmount(sendAmount, destPub, txBF)
			require.NoError(t, err)
			issuerAmountCT, err := mptcrypto.EncryptAmount(sendAmount, issuerPub, txBF)
			require.NoError(t, err)

			participants := []mptcrypto.Participant{
				{PubKey: senderPub, Ciphertext: senderAmountCT},
				{PubKey: destPub, Ciphertext: destAmountCT},
				{PubKey: issuerPub, Ciphertext: issuerAmountCT},
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

			proof, err := mptcrypto.GenerateSendProof(senderPriv, senderPub, sendAmount, participants, txBF, ctxHash, amountParams.Commitment, balanceParams)
			require.NoError(t, err)
			require.NotEmpty(t, proof)

			err = mptcrypto.VerifySendProof(proof, participants, senderBalanceCT, amountCommit, balanceCommit, ctxHash)
			require.NoError(t, err)
		})
	}
}

func TestVerifySendProofRejectsShortProof(t *testing.T) {
	shortProof := make([]byte, mptcrypto.SendProofSize-1)

	var senderCT mptcrypto.Ciphertext
	var amountCommit, balanceCommit mptcrypto.Commitment
	var ctxHash mptcrypto.ContextHash

	err := mptcrypto.VerifySendProof(shortProof, nil, senderCT, amountCommit, balanceCommit, ctxHash)
	require.EqualError(t, err, fmt.Sprintf("mptcrypto: proof must be %d bytes, got %d", mptcrypto.SendProofSize, len(shortProof)))
}

// endregion

// region Internal component verifiers
func TestVerifyRevealedAmount(t *testing.T) {
	const amount uint64 = 42

	bf, err := mptcrypto.GenerateBlindingFactor()
	require.NoError(t, err)

	_, holderPub, err := mptcrypto.GenerateKeypair()
	require.NoError(t, err)
	_, issuerPub, err := mptcrypto.GenerateKeypair()
	require.NoError(t, err)
	_, auditorPub, err := mptcrypto.GenerateKeypair()
	require.NoError(t, err)

	holderCT, err := mptcrypto.EncryptAmount(amount, holderPub, bf)
	require.NoError(t, err)
	issuerCT, err := mptcrypto.EncryptAmount(amount, issuerPub, bf)
	require.NoError(t, err)
	auditorCT, err := mptcrypto.EncryptAmount(amount, auditorPub, bf)
	require.NoError(t, err)

	holder := mptcrypto.Participant{PubKey: holderPub, Ciphertext: holderCT}
	issuer := mptcrypto.Participant{PubKey: issuerPub, Ciphertext: issuerCT}
	auditor := &mptcrypto.Participant{PubKey: auditorPub, Ciphertext: auditorCT}

	tests := []struct {
		name         string
		verifyAmount uint64
		auditor      *mptcrypto.Participant
		wantErr      bool
	}{
		{name: "pass - without auditor", verifyAmount: amount},
		{name: "pass - with auditor", verifyAmount: amount, auditor: auditor},
		{name: "fail - wrong amount", verifyAmount: 99, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mptcrypto.VerifyRevealedAmount(tt.verifyAmount, bf, holder, issuer, tt.auditor)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

// endregion

// region Utilities

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
