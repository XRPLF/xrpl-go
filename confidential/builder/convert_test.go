//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package builder

import (
	"errors"
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

func requireConvertCiphertexts(t *testing.T, result *transaction.ConfidentialMPTConvert, amount uint64, holderKey, issuerKey, auditorKey string) {
	t.Helper()
	holderCiphertext, err := elgamal.Encrypt(amount, holderKey, result.BlindingFactor)
	require.NoError(t, err)
	require.Equal(t, holderCiphertext, result.HolderEncryptedAmount)
	issuerCiphertext, err := elgamal.Encrypt(amount, issuerKey, result.BlindingFactor)
	require.NoError(t, err)
	require.Equal(t, issuerCiphertext, result.IssuerEncryptedAmount)
	if auditorKey == "" {
		require.Nil(t, result.AuditorEncryptedAmount)
		return
	}
	require.NotNil(t, result.AuditorEncryptedAmount)
	auditorCiphertext, err := elgamal.Encrypt(amount, auditorKey, result.BlindingFactor)
	require.NoError(t, err)
	require.Equal(t, auditorCiphertext, *result.AuditorEncryptedAmount)
}

// TestConvertBaseValidation verifies shared malformed-input validation through both entry points.
func TestConvertBaseValidation(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	cases := []struct {
		name    string
		base    BuildConvertParams
		wantErr error
	}{
		{name: "fail - missing account", base: BuildConvertParams{IssuanceID: testIssuanceID, Amount: 1, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: kp.PubKeyHex}, wantErr: ErrMissingAccount},
		{name: "fail - invalid account", base: BuildConvertParams{Account: "notanaddress", IssuanceID: testIssuanceID, Amount: 1, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: kp.PubKeyHex}, wantErr: ErrInvalidAccount},
		{name: "fail - ACCOUNT_ZERO account", base: BuildConvertParams{Account: zeroClassicAccount, IssuanceID: testIssuanceID}, wantErr: ErrInvalidAccount},
		{name: "fail - missing issuance ID", base: BuildConvertParams{Account: testAccount, Amount: 1, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: kp.PubKeyHex}, wantErr: ErrMissingIssuanceID},
		{name: "fail - invalid issuance ID (not hex)", base: BuildConvertParams{Account: testAccount, IssuanceID: strings.Repeat("GG", 24), Amount: 1, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: kp.PubKeyHex}, wantErr: ErrInvalidIssuanceID},
		{name: "fail - invalid issuance ID (wrong length)", base: BuildConvertParams{Account: testAccount, IssuanceID: "aabb", Amount: 1, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: kp.PubKeyHex}, wantErr: ErrInvalidIssuanceID},
		{name: "fail - account is the issuance issuer", base: BuildConvertParams{Account: testAccount, IssuanceID: testIssuerIssuanceID, Amount: 1, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: kp.PubKeyHex}, wantErr: ErrIssuerNotAllowed},
		{name: "fail - amount above protocol maximum", base: BuildConvertParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: math.MaxUint64, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: kp.PubKeyHex}, wantErr: ErrAmountTooLarge},
		{name: "fail - invalid holder priv key (not hex)", base: BuildConvertParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: 1, HolderPrivKey: strings.Repeat("ZZ", 32), HolderPubKey: kp.PubKeyHex}, wantErr: ErrInvalidPrivKey},
		{name: "fail - invalid holder priv key (wrong length)", base: BuildConvertParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: 1, HolderPrivKey: "aabb", HolderPubKey: kp.PubKeyHex}, wantErr: ErrInvalidPrivKey},
		{name: "fail - missing holder pub key", base: BuildConvertParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: 1, HolderPrivKey: kp.PrivKeyHex}, wantErr: ErrMissingHolderKey},
		{name: "fail - invalid holder pub key (not hex)", base: BuildConvertParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: 1, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: strings.Repeat("ZZ", 33)}, wantErr: ErrInvalidPubKey},
		{name: "fail - invalid holder pub key (wrong length)", base: BuildConvertParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: 1, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: "aabb"}, wantErr: ErrInvalidPubKey},
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
	tests := []struct {
		name   string
		amount uint64
	}{
		{name: "non-zero amount", amount: 1000},
		{name: "zero amount registers the key", amount: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			holderKP, err := elgamal.GenerateKeypair()
			require.NoError(t, err)
			issuerKP, err := elgamal.GenerateKeypair()
			require.NoError(t, err)

			result, err := PrepareConvert(ConvertParams{
				BuildConvertParams: BuildConvertParams{
					TxOptions:     TxOptions{TicketSequence: 11},
					Account:       testAccount,
					IssuanceID:    testIssuanceID,
					Amount:        test.amount,
					HolderPrivKey: holderKP.PrivKeyHex,
					HolderPubKey:  holderKP.PubKeyHex,
				},
				IssuerPubKey: issuerKP.PubKeyHex,
				FirstTime:    true,
			})
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, transaction.ConfidentialMPTConvertTx, result.TxType())
			require.EqualValues(t, test.amount, result.MPTAmount)
			requireTicketOptions(t, result.BaseTx, 11, "")

			// First time: key and proof must be set.
			require.NotNil(t, result.HolderEncryptionKey)
			require.Equal(t, holderKP.PubKeyHex, *result.HolderEncryptionKey)
			require.NotNil(t, result.ZKProof)

			// Verify the Schnorr proof cryptographically. It commits to the ticket sequence,
			// the sequence proxy the transaction spends.
			ctxHash, err := proof.ConvertContextHash(testAccount, testIssuanceID, uint32(11))
			require.NoError(t, err)
			err = proof.VerifyConvertProof(*result.ZKProof, holderKP.PubKeyHex, ctxHash)
			require.NoError(t, err)

			requireConvertCiphertexts(t, result, test.amount, holderKP.PubKeyHex, issuerKP.PubKeyHex, "")

			// Transaction must validate.
			ok, err := result.Validate()
			require.NoError(t, err)
			require.True(t, ok)
		})
	}
}

func TestPrepareConvert_PassNotFirstTime(t *testing.T) {
	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	result, err := PrepareConvert(ConvertParams{
		BuildConvertParams: BuildConvertParams{
			TxOptions:    TxOptions{Sequence: 2},
			Account:      testAccount,
			IssuanceID:   testIssuanceID,
			Amount:       500,
			HolderPubKey: holderKP.PubKeyHex,
		},
		IssuerPubKey: issuerKP.PubKeyHex,
		FirstTime:    false,
	})
	require.NoError(t, err)
	require.Nil(t, result.HolderEncryptionKey)
	require.Nil(t, result.ZKProof)
	requireConvertCiphertexts(t, result, 500, holderKP.PubKeyHex, issuerKP.PubKeyHex, "")

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
			TxOptions:     TxOptions{Sequence: 1},
			Account:       testAccount,
			IssuanceID:    testIssuanceID,
			Amount:        100,
			HolderPrivKey: holderKP.PrivKeyHex,
			HolderPubKey:  holderKP.PubKeyHex,
		},
		IssuerPubKey:  issuerKP.PubKeyHex,
		AuditorPubKey: auditorKP.PubKeyHex,
		FirstTime:     false,
	})
	require.NoError(t, err)
	require.NotNil(t, result.AuditorEncryptedAmount)
	require.Len(t, *result.AuditorEncryptedAmount, 132)
	requireConvertCiphertexts(t, result, 100, holderKP.PubKeyHex, issuerKP.PubKeyHex, auditorKP.PubKeyHex)

	ok, err := result.Validate()
	require.NoError(t, err)
	require.True(t, ok)
}

func TestPrepareConvert_FailValidation(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	otherKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	base := BuildConvertParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: 1, HolderPrivKey: kp.PrivKeyHex, HolderPubKey: kp.PubKeyHex}

	tests := []struct {
		name    string
		params  ConvertParams
		wantErr error
	}{
		{name: "fail - missing issuer pub key", params: ConvertParams{BuildConvertParams: base}, wantErr: ErrMissingIssuerKey},
		{name: "fail - invalid issuer pub key (wrong length)", params: ConvertParams{BuildConvertParams: base, IssuerPubKey: "aabb"}, wantErr: ErrInvalidPubKey},
		{name: "fail - invalid issuer pub key (not hex)", params: ConvertParams{BuildConvertParams: base, IssuerPubKey: strings.Repeat("ZZ", 33)}, wantErr: ErrInvalidPubKey},
		{name: "fail - invalid auditor pub key (wrong length)", params: ConvertParams{BuildConvertParams: base, IssuerPubKey: kp.PubKeyHex, AuditorPubKey: "aabb"}, wantErr: ErrInvalidPubKey},
		{name: "fail - invalid auditor pub key (not hex)", params: ConvertParams{BuildConvertParams: base, IssuerPubKey: kp.PubKeyHex, AuditorPubKey: strings.Repeat("ZZ", 33)}, wantErr: ErrInvalidPubKey},
		{
			name: "fail - first-time conversion missing holder private key",
			params: ConvertParams{
				BuildConvertParams: BuildConvertParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: 1, HolderPubKey: kp.PubKeyHex},
				IssuerPubKey:       kp.PubKeyHex,
				FirstTime:          true,
			},
			wantErr: ErrMissingHolderKey,
		},
		{
			name: "fail - first-time conversion invalid holder private key",
			params: ConvertParams{
				BuildConvertParams: BuildConvertParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: 1, HolderPrivKey: "aabb", HolderPubKey: kp.PubKeyHex},
				IssuerPubKey:       kp.PubKeyHex,
				FirstTime:          true,
			},
			wantErr: ErrInvalidPrivKey,
		},
		{
			name:    "fail - missing final sequence on first-time form",
			params:  ConvertParams{BuildConvertParams: base, IssuerPubKey: kp.PubKeyHex, FirstTime: true},
			wantErr: ErrMissingSequence,
		},
		{
			name: "fail - mismatched holder keypair",
			params: ConvertParams{
				BuildConvertParams: BuildConvertParams{
					TxOptions:     TxOptions{Sequence: 1},
					Account:       testAccount,
					IssuanceID:    testIssuanceID,
					Amount:        1,
					HolderPrivKey: otherKP.PrivKeyHex,
					HolderPubKey:  kp.PubKeyHex,
				},
				IssuerPubKey: kp.PubKeyHex,
				FirstTime:    true,
			},
			wantErr: ErrCryptoFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PrepareConvert(tt.params)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestBuildConvert_PassFirstTime(t *testing.T) {
	tests := []struct {
		name   string
		amount uint64
	}{
		{name: "non-zero conversion", amount: 1000},
		{name: "maximum conversion", amount: math.MaxInt64},
		{name: "zero-amount key registration", amount: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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
					mptokenIndex:  buildMPTokenEntry(holdingPublicly(test.amount)),
				},
			}

			result, err := BuildConvert(q, BuildConvertParams{
				Account:       testAccount,
				IssuanceID:    testIssuanceID,
				Amount:        test.amount,
				HolderPrivKey: holderKP.PrivKeyHex,
				HolderPubKey:  holderKP.PubKeyHex,
			})
			require.NoError(t, err)
			require.Equal(t, uint32(5), result.Sequence)
			require.NotNil(t, result.HolderEncryptionKey)
			require.NotNil(t, result.ZKProof)
			requireConvertCiphertexts(t, result, test.amount, holderKP.PubKeyHex, issuerKP.PubKeyHex, "")

			ctxHash, err := proof.ConvertContextHash(testAccount, testIssuanceID, result.Sequence)
			require.NoError(t, err)
			require.NoError(t, proof.VerifyConvertProof(*result.ZKProof, holderKP.PubKeyHex, ctxHash))

			holderAmount, err := elgamal.Decrypt(result.HolderEncryptedAmount, holderKP.PrivKeyHex, elgamal.AmountRange{Low: test.amount, High: test.amount})
			require.NoError(t, err)
			require.Equal(t, test.amount, holderAmount)
			issuerAmount, err := elgamal.Decrypt(result.IssuerEncryptedAmount, issuerKP.PrivKeyHex, elgamal.AmountRange{Low: test.amount, High: test.amount})
			require.NoError(t, err)
			require.Equal(t, test.amount, issuerAmount)

			valid, err := result.Validate()
			require.NoError(t, err)
			require.True(t, valid)
		})
	}
}

func TestBuildConvertRejectsMissingMPToken(t *testing.T) {
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

	_, err = BuildConvert(q, BuildConvertParams{
		Account:       testAccount,
		IssuanceID:    testIssuanceID,
		Amount:        1000,
		HolderPrivKey: holderKP.PrivKeyHex,
		HolderPubKey:  holderKP.PubKeyHex,
	})
	require.ErrorIs(t, err, ErrMPTokenNotFound)
}

// TestBuildConvertChecksPublicBalance pins the public-balance preflight.
// ConfidentialMPTConvert debits sfMPTAmount and rejects a convert it cannot fund with
// tecINSUFFICIENT_FUNDS, so the builder decides it from the MPToken it already read. A
// zero-amount convert funds nothing, which is what keeps the opt-in path open to a holder
// with no public balance at all.
func TestBuildConvertChecksPublicBalance(t *testing.T) {
	tests := []struct {
		name         string
		publicAmount uint64
		amount       uint64
		wantErr      error
	}{
		{name: "converts less than it holds", publicAmount: 1000, amount: 400},
		{name: "converts exactly what it holds", publicAmount: 1000, amount: 1000},
		{name: "converts more than it holds", publicAmount: 1000, amount: 1001, wantErr: ErrInsufficientBalance},
		{name: "holds nothing", amount: 1, wantErr: ErrInsufficientBalance},
		{name: "zero-amount opt-in while holding nothing", amount: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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
					mptokenIndex:  buildMPTokenEntry(holdingPublicly(test.publicAmount)),
				},
			}

			_, err = BuildConvert(q, BuildConvertParams{
				Account:       testAccount,
				IssuanceID:    testIssuanceID,
				Amount:        test.amount,
				HolderPrivKey: holderKP.PrivKeyHex,
				HolderPubKey:  holderKP.PubKeyHex,
			})
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestBuildConvertRejectsMalformedHolderPrivKeyBeforeLedgerQueries(t *testing.T) {
	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	q := &mockQuerier{accountSeq: 5}
	_, err = BuildConvert(q, BuildConvertParams{
		Account:       testAccount,
		IssuanceID:    testIssuanceID,
		Amount:        1000,
		HolderPrivKey: "aabb",
		HolderPubKey:  holderKP.PubKeyHex,
	})
	require.ErrorIs(t, err, ErrInvalidPrivKey)
	require.Zero(t, q.queryCalls, "malformed private keys must fail before ledger queries")
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
			mptokenIndex:  buildMPTokenEntry(mptokenFields{holderKey: strings.ToUpper(holderKP.PubKeyHex), balanceCt: balanceCt, publicAmount: 200}),
		},
	}

	result, err := BuildConvert(q, BuildConvertParams{
		Account:      testAccount,
		IssuanceID:   testIssuanceID,
		Amount:       200,
		HolderPubKey: holderKP.PubKeyHex,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Nil(t, result.HolderEncryptionKey)
	require.Nil(t, result.ZKProof)

	differentKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	_, err = BuildConvert(q, BuildConvertParams{
		Account:      testAccount,
		IssuanceID:   testIssuanceID,
		Amount:       200,
		HolderPubKey: differentKP.PubKeyHex,
	})
	require.ErrorIs(t, err, ErrKeyMismatch)
	require.ErrorContains(t, err, "holder key")
}

func TestBuildConvertPreservesMPTokenQueryError(t *testing.T) {
	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)
	mptokenIndex, err := xrplhash.MPToken(testIssuanceID, testAccount)
	require.NoError(t, err)
	transportErr := errors.New("transport failed")
	q := &mockQuerier{
		accountSeq: 5,
		entries: map[string]ledgerentries.FlatLedgerObject{
			issuanceIndex: buildIssuanceEntry(issuerKP.PubKeyHex, ""),
		},
		entryErrs: map[string]error{mptokenIndex: transportErr},
	}

	_, err = BuildConvert(q, BuildConvertParams{
		Account:       testAccount,
		IssuanceID:    testIssuanceID,
		Amount:        100,
		HolderPrivKey: holderKP.PrivKeyHex,
		HolderPubKey:  holderKP.PubKeyHex,
	})
	require.ErrorIs(t, err, ErrLedgerQuery)
	require.ErrorIs(t, err, transportErr)
	require.NotErrorIs(t, err, ErrMPTokenNotFound)
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
	mptokenIndex, err := xrplhash.MPToken(testIssuanceID, testAccount)
	require.NoError(t, err)

	q := &mockQuerier{
		accountSeq: 3,
		entries: map[string]ledgerentries.FlatLedgerObject{
			issuanceIndex: buildIssuanceEntry(issuerKP.PubKeyHex, auditorKP.PubKeyHex),
			mptokenIndex:  buildMPTokenEntry(holdingPublicly(100)),
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
	requireConvertCiphertexts(t, result, 100, holderKP.PubKeyHex, issuerKP.PubKeyHex, auditorKP.PubKeyHex)
}

// TestBuildConvertTicketBindsProofAndSkipsSequenceLookup pins the online first-time path with a
// caller-supplied ticket: no account query runs, and the Schnorr proof commits to the ticket
// sequence rather than to an account sequence the builder never read.
func TestBuildConvertTicketBindsProofAndSkipsSequenceLookup(t *testing.T) {
	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)
	mptokenIndex, err := xrplhash.MPToken(testIssuanceID, testAccount)
	require.NoError(t, err)

	q := &mockQuerier{
		accountErr: errors.New("the account sequence must not be read"),
		entries: map[string]ledgerentries.FlatLedgerObject{
			issuanceIndex: buildIssuanceEntry(issuerKP.PubKeyHex, ""),
			mptokenIndex:  buildMPTokenEntry(holdingPublicly(1000)),
		},
	}

	result, err := BuildConvert(q, BuildConvertParams{
		TxOptions:     TxOptions{TicketSequence: 5},
		Account:       testAccount,
		IssuanceID:    testIssuanceID,
		Amount:        1000,
		HolderPrivKey: holderKP.PrivKeyHex,
		HolderPubKey:  holderKP.PubKeyHex,
	})
	require.NoError(t, err)
	requireTicketOptions(t, result.BaseTx, 5, "")
	require.Empty(t, q.accountRequests)

	ctxHash, err := proof.ConvertContextHash(testAccount, testIssuanceID, 5)
	require.NoError(t, err)
	require.NoError(t, proof.VerifyConvertProof(*result.ZKProof, holderKP.PubKeyHex, ctxHash))
}

// TestBuildConvertRejectsDelegate pins that the non-delegable rule is reached before any ledger
// read, so a delegate xrpld would reject never costs a query.
func TestBuildConvertRejectsDelegate(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	q := &mockQuerier{accountSeq: 5}

	_, err = BuildConvert(q, BuildConvertParams{
		TxOptions:    TxOptions{Sequence: 1, Delegate: testDelegate},
		Account:      testAccount,
		IssuanceID:   testIssuanceID,
		Amount:       1,
		HolderPubKey: kp.PubKeyHex,
	})

	require.ErrorIs(t, err, ErrDelegateNotAllowed)
	require.Zero(t, q.queryCalls)
}

// TestPrepareConvertAllowsZeroSequenceOnRepeatForm pins that only the first-time form demands
// a final sequence. A repeat convert carries no proof, so nothing binds the sequence and the
// caller keeps the prepare-then-autofill flow, matching PrepareMergeInbox.
func TestPrepareConvertAllowsZeroSequenceOnRepeatForm(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	result, err := PrepareConvert(ConvertParams{
		BuildConvertParams: BuildConvertParams{
			Account:      testAccount,
			IssuanceID:   testIssuanceID,
			Amount:       100,
			HolderPubKey: kp.PubKeyHex,
		},
		IssuerPubKey: kp.PubKeyHex,
	})
	require.NoError(t, err)
	require.Zero(t, result.Sequence)
	require.Nil(t, result.ZKProof)
	require.Nil(t, result.HolderEncryptionKey)
}
