//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package builder

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	ledgerentries "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

// TestClawbackBaseValidation verifies shared malformed-input validation through both entry points.
func TestClawbackBaseValidation(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	cases := []struct {
		name    string
		base    BuildClawbackParams
		wantErr error
	}{
		{name: "fail - missing account", base: BuildClawbackParams{Holder: testDestination, IssuanceID: testIssuerIssuanceID, IssuerPrivKey: kp.PrivKeyHex}, wantErr: ErrMissingAccount},
		{name: "fail - invalid account", base: BuildClawbackParams{Account: "notanaddress", Holder: testDestination, IssuanceID: testIssuerIssuanceID, IssuerPrivKey: kp.PrivKeyHex}, wantErr: ErrInvalidAccount},
		{name: "fail - missing holder", base: BuildClawbackParams{Account: testAccount, IssuanceID: testIssuerIssuanceID, IssuerPrivKey: kp.PrivKeyHex}, wantErr: ErrMissingHolder},
		{name: "fail - invalid holder", base: BuildClawbackParams{Account: testAccount, Holder: "notanaddress", IssuanceID: testIssuerIssuanceID, IssuerPrivKey: kp.PrivKeyHex}, wantErr: ErrInvalidHolder},
		{name: "fail - self clawback", base: BuildClawbackParams{Account: testAccount, Holder: testAccount, IssuanceID: testIssuerIssuanceID, IssuerPrivKey: kp.PrivKeyHex}, wantErr: ErrSelfClawback},
		{name: "fail - ACCOUNT_ZERO account", base: BuildClawbackParams{Account: zeroClassicAccount, Holder: testDestination, IssuanceID: testIssuerIssuanceID, IssuerPrivKey: kp.PrivKeyHex}, wantErr: ErrInvalidAccount},
		{name: "fail - ACCOUNT_ZERO holder", base: BuildClawbackParams{Account: testAccount, Holder: xAddressOf(t, zeroClassicAccount), IssuanceID: testIssuerIssuanceID, IssuerPrivKey: kp.PrivKeyHex}, wantErr: ErrInvalidHolder},
		{name: "fail - self clawback across address forms", base: BuildClawbackParams{Account: testAccount, Holder: xAddressOf(t, testAccount), IssuanceID: testIssuerIssuanceID, IssuerPrivKey: kp.PrivKeyHex}, wantErr: ErrSelfClawback},
		{name: "fail - tagged X-address holder", base: BuildClawbackParams{Account: testAccount, Holder: taggedXAddressOf(t, testDestination, 42), IssuanceID: testIssuerIssuanceID, IssuerPrivKey: kp.PrivKeyHex}, wantErr: transaction.ErrAccountIDTagNotAllowed},
		{name: "fail - account is not the issuance issuer", base: BuildClawbackParams{Account: testAccount, Holder: testDestination, IssuanceID: testIssuanceID, IssuerPrivKey: kp.PrivKeyHex}, wantErr: ErrNotIssuer},
		{name: "fail - missing issuance ID", base: BuildClawbackParams{Account: testAccount, Holder: testDestination, IssuerPrivKey: kp.PrivKeyHex}, wantErr: ErrMissingIssuanceID},
		{name: "fail - invalid issuance ID (not hex)", base: BuildClawbackParams{Account: testAccount, Holder: testDestination, IssuanceID: "ZZZZ", IssuerPrivKey: kp.PrivKeyHex}, wantErr: ErrInvalidIssuanceID},
		{name: "fail - invalid issuance ID (wrong length)", base: BuildClawbackParams{Account: testAccount, Holder: testDestination, IssuanceID: "aabb", IssuerPrivKey: kp.PrivKeyHex}, wantErr: ErrInvalidIssuanceID},
		{name: "fail - missing issuer priv key", base: BuildClawbackParams{Account: testAccount, Holder: testDestination, IssuanceID: testIssuerIssuanceID}, wantErr: ErrMissingIssuerKey},
		{name: "fail - invalid issuer priv key (not hex)", base: BuildClawbackParams{Account: testAccount, Holder: testDestination, IssuanceID: testIssuerIssuanceID, IssuerPrivKey: "ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ"}, wantErr: ErrInvalidPrivKey},
		{name: "fail - invalid issuer priv key (wrong length)", base: BuildClawbackParams{Account: testAccount, Holder: testDestination, IssuanceID: testIssuerIssuanceID, IssuerPrivKey: "aabb"}, wantErr: ErrInvalidPrivKey},
		{name: "fail - zero issuer private scalar", base: BuildClawbackParams{Account: testAccount, Holder: testDestination, IssuanceID: testIssuerIssuanceID, IssuerPrivKey: strings.Repeat("0", 64)}, wantErr: ErrInvalidPrivKey},
		{name: "fail - issuer private scalar equals curve order", base: BuildClawbackParams{Account: testAccount, Holder: testDestination, IssuanceID: testIssuerIssuanceID, IssuerPrivKey: "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141"}, wantErr: ErrInvalidPrivKey},
		{name: "fail - issuer private scalar above curve order", base: BuildClawbackParams{Account: testAccount, Holder: testDestination, IssuanceID: testIssuerIssuanceID, IssuerPrivKey: "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364142"}, wantErr: ErrInvalidPrivKey},
	}

	blindingFactor, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	validCiphertext, err := elgamal.Encrypt(1, kp.PubKeyHex, blindingFactor)
	require.NoError(t, err)

	t.Run("fail - validation PrepareClawback", func(t *testing.T) {
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := PrepareClawback(ClawbackParams{BuildClawbackParams: tc.base, Amount: 1, IssuerPubKey: kp.PubKeyHex, IssuerCiphertext: validCiphertext})
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
			IssuanceID:    testIssuerIssuanceID,
			IssuerPrivKey: issuerKP.PrivKeyHex,
		},
		Amount:           amount,
		IssuerPubKey:     issuerKP.PubKeyHex,
		IssuerCiphertext: issuerCt,
		Sequence:         1,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, transaction.ConfidentialMPTClawbackTx, result.TxType())
	require.NotEmpty(t, result.ZKProof)

	ctxHash, err := proof.ClawbackContextHash(testAccount, testIssuerIssuanceID, uint32(1), testDestination)
	require.NoError(t, err)
	err = proof.VerifyClawbackProof(result.ZKProof, amount, issuerKP.PubKeyHex, issuerCt, ctxHash)
	require.NoError(t, err)

	ok, err := result.Validate()
	require.NoError(t, err)
	require.True(t, ok)
}

func TestPrepareClawback_RejectsAmountCiphertextMismatch(t *testing.T) {
	const encryptedAmount uint64 = 500
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	blindingFactor, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	issuerCiphertext, err := elgamal.Encrypt(encryptedAmount, issuerKP.PubKeyHex, blindingFactor)
	require.NoError(t, err)

	_, err = PrepareClawback(ClawbackParams{
		BuildClawbackParams: BuildClawbackParams{
			Account:       testAccount,
			Holder:        testDestination,
			IssuanceID:    testIssuerIssuanceID,
			IssuerPrivKey: issuerKP.PrivKeyHex,
		},
		Amount:           encryptedAmount - 1,
		IssuerPubKey:     issuerKP.PubKeyHex,
		IssuerCiphertext: issuerCiphertext,
		Sequence:         1,
	})
	require.ErrorIs(t, err, ErrCryptoFailed)
	require.ErrorIs(t, err, proof.ErrProofGenerationFailed)
}

func TestPrepareClawback_MaximumAmount(t *testing.T) {
	const amount uint64 = math.MaxInt64
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	blindingFactor, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	issuerCiphertext, err := elgamal.Encrypt(amount, issuerKP.PubKeyHex, blindingFactor)
	require.NoError(t, err)

	result, err := PrepareClawback(ClawbackParams{
		BuildClawbackParams: BuildClawbackParams{
			Account:       testAccount,
			Holder:        testDestination,
			IssuanceID:    testIssuerIssuanceID,
			IssuerPrivKey: issuerKP.PrivKeyHex,
		},
		Amount:           amount,
		IssuerPubKey:     issuerKP.PubKeyHex,
		IssuerCiphertext: issuerCiphertext,
		Sequence:         1,
	})
	require.NoError(t, err)

	contextHash, err := proof.ClawbackContextHash(testAccount, testIssuerIssuanceID, result.Sequence, testDestination)
	require.NoError(t, err)
	require.NoError(t, proof.VerifyClawbackProof(result.ZKProof, amount, issuerKP.PubKeyHex, issuerCiphertext, contextHash))
}

func TestPrepareClawback_FailValidation(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	base := BuildClawbackParams{Account: testAccount, Holder: testDestination, IssuanceID: testIssuerIssuanceID, IssuerPrivKey: kp.PrivKeyHex}
	blindingFactor, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	validCiphertext, err := elgamal.Encrypt(1, kp.PubKeyHex, blindingFactor)
	require.NoError(t, err)

	tests := []struct {
		name    string
		params  ClawbackParams
		wantErr error
	}{
		{name: "fail - missing issuer pub key", params: ClawbackParams{BuildClawbackParams: base, Amount: 1, IssuerCiphertext: validCiphertext}, wantErr: ErrMissingIssuerKey},
		{name: "fail - invalid issuer pub key (wrong length)", params: ClawbackParams{BuildClawbackParams: base, Amount: 1, IssuerPubKey: "aabb", IssuerCiphertext: validCiphertext}, wantErr: ErrInvalidPubKey},
		{name: "fail - invalid issuer pub key (not hex)", params: ClawbackParams{BuildClawbackParams: base, Amount: 1, IssuerPubKey: strings.Repeat("ZZ", 33), IssuerCiphertext: validCiphertext}, wantErr: ErrInvalidPubKey},
		{name: "fail - missing ciphertext", params: ClawbackParams{BuildClawbackParams: base, Amount: 1, IssuerPubKey: kp.PubKeyHex}, wantErr: ErrMissingCiphertext},
		{name: "fail - invalid ciphertext (wrong length)", params: ClawbackParams{BuildClawbackParams: base, Amount: 1, IssuerPubKey: kp.PubKeyHex, IssuerCiphertext: "aabb"}, wantErr: ErrInvalidCiphertext},
		{name: "fail - invalid ciphertext (not hex)", params: ClawbackParams{BuildClawbackParams: base, Amount: 1, IssuerPubKey: kp.PubKeyHex, IssuerCiphertext: strings.Repeat("ZZ", 66)}, wantErr: ErrInvalidCiphertext},
		{name: "fail - missing final sequence", params: ClawbackParams{BuildClawbackParams: base, Amount: 1, IssuerPubKey: kp.PubKeyHex, IssuerCiphertext: validCiphertext}, wantErr: ErrMissingSequence},
		{name: "fail - zero amount", params: ClawbackParams{BuildClawbackParams: base, Amount: 0, IssuerPubKey: kp.PubKeyHex, IssuerCiphertext: validCiphertext}, wantErr: ErrZeroAmount},
		{name: "fail - amount above protocol maximum", params: ClawbackParams{BuildClawbackParams: base, Amount: math.MaxUint64, IssuerPubKey: kp.PubKeyHex, IssuerCiphertext: validCiphertext}, wantErr: ErrAmountTooLarge},
		{name: "fail - amount above protocol maximum (boundary)", params: ClawbackParams{BuildClawbackParams: base, Amount: uint64(math.MaxInt64) + 1, IssuerPubKey: kp.PubKeyHex, IssuerCiphertext: validCiphertext}, wantErr: ErrAmountTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PrepareClawback(tt.params)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestBuildClawback_FailLedgerQueries(t *testing.T) {
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuerIssuanceID)
	require.NoError(t, err)
	mptokenIndex, err := xrplhash.MPToken(testIssuerIssuanceID, testDestination)
	require.NoError(t, err)

	entryNotFoundErr := errors.New(ledgerEntryNotFound)
	sequenceErr := errors.New("account query failed")
	transportErr := errors.New("transport failed")
	validParams := BuildClawbackParams{
		Account:       testAccount,
		Holder:        testDestination,
		IssuanceID:    testIssuerIssuanceID,
		IssuerPrivKey: issuerKP.PrivKeyHex,
		BalanceRange:  elgamal.AmountRange{Low: 0, High: 1000},
	}

	tests := []struct {
		name    string
		querier *mockQuerier
		wantErr error
		cause   error
	}{
		{
			name:    "fail - beginBuild error",
			querier: &mockQuerier{accountErr: sequenceErr},
			wantErr: ErrLedgerQuery,
			cause:   sequenceErr,
		},
		{
			name:    "fail - issuance entry not found",
			querier: &mockQuerier{accountSeq: 1, entries: map[string]ledgerentries.FlatLedgerObject{}},
			wantErr: ErrIssuanceNotFound,
		},
		{
			name: "fail - confidential balances not enabled",
			querier: &mockQuerier{accountSeq: 1, entries: map[string]ledgerentries.FlatLedgerObject{
				issuanceIndex: {"LedgerEntryType": "MPTokenIssuance"}, // entry exists but carries no capability flags
			}},
			wantErr: ErrConfidentialDisabled,
		},
		{
			name: "fail - clawback not enabled",
			querier: &mockQuerier{accountSeq: 1, entries: map[string]ledgerentries.FlatLedgerObject{
				issuanceIndex: withIssuanceFlags(buildIssuanceEntry(issuerKP.PubKeyHex, ""),
					ledgerentries.LsfMPTCanHoldConfidentialBalance),
			}},
			wantErr: ErrClawbackDisabled,
		},
		{
			name: "fail - issuer encryption key not set",
			querier: &mockQuerier{accountSeq: 1, entries: map[string]ledgerentries.FlatLedgerObject{
				issuanceIndex: buildIssuanceEntry("", ""), // entry exists but no IssuerEncryptionKey
			}},
			wantErr: ErrEncryptionKeyNotSet,
		},
		{
			name: "fail - MPToken not found",
			querier: &mockQuerier{accountSeq: 1, entries: map[string]ledgerentries.FlatLedgerObject{
				issuanceIndex: buildIssuanceEntry(issuerKP.PubKeyHex, ""),
				// no MPToken entry
			}},
			wantErr: ErrMPTokenNotFound,
		},
		{
			name: "fail - rippled entryNotFound",
			querier: &mockQuerier{
				accountSeq: 1,
				entries: map[string]ledgerentries.FlatLedgerObject{
					issuanceIndex: buildIssuanceEntry(issuerKP.PubKeyHex, ""),
				},
				entryErrs: map[string]error{mptokenIndex: entryNotFoundErr},
			},
			wantErr: ErrMPTokenNotFound,
			cause:   entryNotFoundErr,
		},
		{
			name: "fail - wrapped rippled entryNotFound",
			querier: &mockQuerier{
				accountSeq: 1,
				entries: map[string]ledgerentries.FlatLedgerObject{
					issuanceIndex: buildIssuanceEntry(issuerKP.PubKeyHex, ""),
				},
				entryErrs: map[string]error{mptokenIndex: fmt.Errorf("request failed: %w", entryNotFoundErr)},
			},
			wantErr: ErrMPTokenNotFound,
			cause:   entryNotFoundErr,
		},
		{
			name: "fail - MPToken transport error",
			querier: &mockQuerier{
				accountSeq: 1,
				entries: map[string]ledgerentries.FlatLedgerObject{
					issuanceIndex: buildIssuanceEntry(issuerKP.PubKeyHex, ""),
				},
				entryErrs: map[string]error{mptokenIndex: transportErr},
			},
			wantErr: ErrLedgerQuery,
			cause:   transportErr,
		},
		{
			name: "fail - IssuerEncryptedBalance missing",
			querier: &mockQuerier{accountSeq: 1, entries: map[string]ledgerentries.FlatLedgerObject{
				issuanceIndex: buildIssuanceEntry(issuerKP.PubKeyHex, ""),
				mptokenIndex:  buildMPTokenEntry(mptokenFields{holderKey: testHolderKey}), // no IssuerEncryptedBalance
			}},
			wantErr: ErrInvalidLedgerState,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildClawback(tt.querier, validParams)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.cause != nil {
				require.ErrorIs(t, err, tt.cause)
			}
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

	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuerIssuanceID)
	require.NoError(t, err)
	mptokenIndex, err := xrplhash.MPToken(testIssuerIssuanceID, testDestination)
	require.NoError(t, err)

	q := &mockQuerier{
		accountSeq: 10,
		entries: map[string]ledgerentries.FlatLedgerObject{
			issuanceIndex: buildIssuanceEntry(issuerKP.PubKeyHex, ""),
			mptokenIndex:  buildMPTokenEntry(clawable(issuerCt)),
		},
	}

	result, err := BuildClawback(q, BuildClawbackParams{
		Account:       testAccount,
		Holder:        testDestination,
		IssuanceID:    testIssuerIssuanceID,
		IssuerPrivKey: issuerKP.PrivKeyHex,
		BalanceRange:  elgamal.AmountRange{Low: 0, High: clawbackAmount},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, transaction.ConfidentialMPTClawbackTx, result.TxType())
	require.Equal(t, transaction.ConfidentialMPTClawbackTx, result.TransactionType)
	require.Equal(t, testAccount, result.Account.String())
	require.Equal(t, testDestination, result.Holder.String())
	require.Equal(t, testIssuerIssuanceID, result.MPTokenIssuanceID)
	require.EqualValues(t, clawbackAmount, result.MPTAmount)
	require.Equal(t, uint32(10), result.Sequence)

	contextHash, err := proof.ClawbackContextHash(testAccount, testIssuerIssuanceID, result.Sequence, testDestination)
	require.NoError(t, err)
	require.NoError(t, proof.VerifyClawbackProof(result.ZKProof, clawbackAmount, issuerKP.PubKeyHex, issuerCt, contextHash))

	valid, err := result.Validate()
	require.NoError(t, err)
	require.True(t, valid)
}

// TestBuildClawbackDerivesAmountFromLedger pins XLS-96 11.1: the issuer decrypts the holder's
// IssuerEncryptedBalance to obtain the amount. The caller never supplies it, so a balance the
// caller could not have guessed must still reach MPTAmount and verify against the ciphertext.
func TestBuildClawbackDerivesAmountFromLedger(t *testing.T) {
	const holderBalance uint64 = 8237

	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	issuerCt, err := elgamal.Encrypt(holderBalance, issuerKP.PubKeyHex, bf)
	require.NoError(t, err)

	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuerIssuanceID)
	require.NoError(t, err)
	mptokenIndex, err := xrplhash.MPToken(testIssuerIssuanceID, testDestination)
	require.NoError(t, err)

	q := &mockQuerier{
		accountSeq: 4,
		entries: map[string]ledgerentries.FlatLedgerObject{
			issuanceIndex: buildIssuanceEntry(issuerKP.PubKeyHex, ""),
			mptokenIndex:  buildMPTokenEntry(clawable(issuerCt)),
		},
	}

	result, err := BuildClawback(q, BuildClawbackParams{
		Account:       testAccount,
		Holder:        testDestination,
		IssuanceID:    testIssuerIssuanceID,
		IssuerPrivKey: issuerKP.PrivKeyHex,
		BalanceRange:  elgamal.AmountRange{Low: 0, High: 10000},
	})
	require.NoError(t, err)
	require.EqualValues(t, holderBalance, result.MPTAmount)

	contextHash, err := proof.ClawbackContextHash(testAccount, testIssuerIssuanceID, result.Sequence, testDestination)
	require.NoError(t, err)
	require.NoError(t, proof.VerifyClawbackProof(result.ZKProof, holderBalance, issuerKP.PubKeyHex, issuerCt, contextHash))
}

func TestBuildClawback_InvalidRangeBeforeLedgerQueries(t *testing.T) {
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	q := &mockQuerier{accountErr: ErrLedgerQuery}
	_, err = BuildClawback(q, BuildClawbackParams{
		Account:       testAccount,
		Holder:        testDestination,
		IssuanceID:    testIssuerIssuanceID,
		IssuerPrivKey: issuerKP.PrivKeyHex,
		BalanceRange:  elgamal.AmountRange{Low: 2, High: 1},
	})
	require.ErrorIs(t, err, elgamal.ErrInvalidAmountRange)
	require.NotErrorIs(t, err, ErrLedgerQuery)
	require.Zero(t, q.queryCalls, "invalid ranges must fail before ledger queries")
}

// TestBuildClawbackBoundsSearchByOutstandingAmount pins that ConfidentialOutstandingAmount
// caps the decryption search. rippled fails a clawback above it with tecINSUFFICIENT_FUNDS,
// and no holder balance can exceed the confidential supply, so lowering the ceiling never
// hides a balance the transaction could have clawed back.
func TestBuildClawbackBoundsSearchByOutstandingAmount(t *testing.T) {
	const holderBalance uint64 = 500

	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	issuerCt, err := elgamal.Encrypt(holderBalance, issuerKP.PubKeyHex, bf)
	require.NoError(t, err)

	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuerIssuanceID)
	require.NoError(t, err)
	mptokenIndex, err := xrplhash.MPToken(testIssuerIssuanceID, testDestination)
	require.NoError(t, err)

	tests := []struct {
		name        string
		outstanding string
		low         uint64
		wantErr     error
	}{
		{name: "outstanding covers the balance", outstanding: "500"},
		{name: "outstanding above the balance", outstanding: "100000"},
		{
			name:        "outstanding below the balance",
			outstanding: "100",
			wantErr:     ErrCryptoFailed,
		},
		{
			// A range that starts above the confidential supply cannot contain any holder
			// balance, which is a problem with the bounds the caller supplied rather than the
			// protocol bound ErrAmountExceedsOutstanding reports.
			name:        "range starts above the outstanding amount",
			outstanding: "100",
			low:         200,
			wantErr:     elgamal.ErrInvalidAmountRange,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			issuance := buildIssuanceEntry(issuerKP.PubKeyHex, "")
			issuance["ConfidentialOutstandingAmount"] = test.outstanding
			q := &mockQuerier{
				accountSeq: 10,
				entries: map[string]ledgerentries.FlatLedgerObject{
					issuanceIndex: issuance,
					mptokenIndex:  buildMPTokenEntry(clawable(issuerCt)),
				},
			}

			result, err := BuildClawback(q, BuildClawbackParams{
				Account:       testAccount,
				Holder:        testDestination,
				IssuanceID:    testIssuerIssuanceID,
				IssuerPrivKey: issuerKP.PrivKeyHex,
				BalanceRange:  elgamal.AmountRange{Low: test.low, High: math.MaxUint32},
			})
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, types.MPTPlainAmount(holderBalance), result.MPTAmount)
		})
	}
}
