package builder

import (
	"errors"
	"strings"
	"testing"

	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	ledgerentries "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/stretchr/testify/require"
)

// TestMergeInboxBaseValidation verifies shared malformed-input validation through both entry points.
func TestMergeInboxBaseValidation(t *testing.T) {
	cases := []struct {
		name    string
		base    BuildMergeInboxParams
		wantErr error
	}{
		{name: "fail - missing account", base: BuildMergeInboxParams{IssuanceID: testIssuanceID}, wantErr: ErrMissingAccount},
		{name: "fail - invalid account", base: BuildMergeInboxParams{Account: "notanaddress", IssuanceID: testIssuanceID}, wantErr: ErrInvalidAccount},
		{name: "fail - ACCOUNT_ZERO account", base: BuildMergeInboxParams{Account: zeroClassicAccount, IssuanceID: testIssuanceID}, wantErr: ErrInvalidAccount},
		{name: "fail - missing issuance ID", base: BuildMergeInboxParams{Account: testAccount}, wantErr: ErrMissingIssuanceID},
		{name: "fail - invalid issuance ID (not hex)", base: BuildMergeInboxParams{Account: testAccount, IssuanceID: strings.Repeat("GG", 24)}, wantErr: ErrInvalidIssuanceID},
		{name: "fail - invalid issuance ID (wrong length)", base: BuildMergeInboxParams{Account: testAccount, IssuanceID: "aabb"}, wantErr: ErrInvalidIssuanceID},
		{name: "fail - account is the issuance issuer", base: BuildMergeInboxParams{Account: testAccount, IssuanceID: testIssuerIssuanceID}, wantErr: ErrIssuerNotAllowed},
	}

	t.Run("fail - validation PrepareMergeInbox", func(t *testing.T) {
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := PrepareMergeInbox(MergeInboxParams{BuildMergeInboxParams: tc.base})
				require.ErrorIs(t, err, tc.wantErr)
			})
		}
	})

	t.Run("fail - validation BuildMergeInbox", func(t *testing.T) {
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				q := &mockQuerier{}
				_, err := BuildMergeInbox(q, tc.base)
				require.ErrorIs(t, err, tc.wantErr)
				require.Zero(t, q.queryCalls)
			})
		}
	})
}

func TestPrepareMergeInbox_Pass(t *testing.T) {
	result, err := PrepareMergeInbox(MergeInboxParams{
		BuildMergeInboxParams: BuildMergeInboxParams{
			Account:    testAccount,
			IssuanceID: testIssuanceID,
		},
		Sequence: 42,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, transaction.ConfidentialMPTMergeInboxTx, result.TxType())
	require.Equal(t, testIssuanceID, result.MPTokenIssuanceID)

	ok, err := result.Validate()
	require.NoError(t, err)
	require.True(t, ok)
}

// newMergeInboxLedger builds a ledger where testAccount can merge its inbox.
func newMergeInboxLedger(t *testing.T, sequence uint32) *mockQuerier {
	t.Helper()

	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)
	mptokenIndex, err := xrplhash.MPToken(testIssuanceID, testAccount)
	require.NoError(t, err)

	return &mockQuerier{
		accountSeq: sequence,
		entries: map[string]ledgerentries.FlatLedgerObject{
			issuanceIndex: buildIssuanceEntry("issuer-key", ""),
			mptokenIndex:  buildMPTokenEntry(mergeable(testHolderKey)),
		},
	}
}

func TestBuildMergeInbox_Pass(t *testing.T) {
	q := newMergeInboxLedger(t, 42)

	result, err := BuildMergeInbox(q, BuildMergeInboxParams{
		Account:    testAccount,
		IssuanceID: testIssuanceID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, uint32(42), result.Sequence)
}

// TestBuildMergeInboxLedgerPreflight covers the state ConfidentialMPTMergeInbox rejects.
// The merge carries no proof, so these reads exist only to keep a doomed transaction from
// costing a fee and a sequence.
func TestBuildMergeInboxLedgerPreflight(t *testing.T) {
	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)
	mptokenIndex, err := xrplhash.MPToken(testIssuanceID, testAccount)
	require.NoError(t, err)

	tests := []struct {
		name    string
		mutate  func(*mockQuerier)
		wantErr error
	}{
		{
			name:    "fail - issuance not found",
			mutate:  func(q *mockQuerier) { delete(q.entries, issuanceIndex) },
			wantErr: ErrIssuanceNotFound,
		},
		{
			name: "fail - confidential balances not enabled",
			mutate: func(q *mockQuerier) {
				q.entries[issuanceIndex]["Flags"] = float64(ledgerentries.LsfMPTCanTransfer)
			},
			wantErr: ErrConfidentialDisabled,
		},
		{
			name:    "fail - MPToken not found",
			mutate:  func(q *mockQuerier) { delete(q.entries, mptokenIndex) },
			wantErr: ErrMPTokenNotFound,
		},
		{
			name:    "fail - inbox balance missing",
			mutate:  func(q *mockQuerier) { delete(q.entries[mptokenIndex], "ConfidentialBalanceInbox") },
			wantErr: ErrMissingSenderState,
		},
		{
			name:    "fail - spending balance missing",
			mutate:  func(q *mockQuerier) { delete(q.entries[mptokenIndex], "ConfidentialBalanceSpending") },
			wantErr: ErrMissingSenderState,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q := newMergeInboxLedger(t, 42)
			test.mutate(q)

			_, err := BuildMergeInbox(q, BuildMergeInboxParams{
				Account:    testAccount,
				IssuanceID: testIssuanceID,
			})
			require.ErrorIs(t, err, test.wantErr)
		})
	}
}

// TestBuildMergeInboxKeyRequirements separates the two encryption keys the merge is easily
// assumed to treat alike. ConfidentialMPTMergeInbox reads no issuer key, because it encrypts
// nothing, but it does reject a holder that never registered one.
func TestBuildMergeInboxKeyRequirements(t *testing.T) {
	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)
	mptokenIndex, err := xrplhash.MPToken(testIssuanceID, testAccount)
	require.NoError(t, err)

	tests := []struct {
		name    string
		index   string
		removed string
		wantErr error
	}{
		{name: "issuance without an issuer key", index: issuanceIndex, removed: "IssuerEncryptionKey"},
		{name: "holder without a holder key", index: mptokenIndex, removed: "HolderEncryptionKey", wantErr: ErrMissingSenderState},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q := newMergeInboxLedger(t, 42)
			delete(q.entries[test.index], test.removed)

			result, err := BuildMergeInbox(q, BuildMergeInboxParams{
				Account:    testAccount,
				IssuanceID: testIssuanceID,
			})
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, uint32(42), result.Sequence)
		})
	}
}

func TestBuildMergeInboxPreservesAccountQueryError(t *testing.T) {
	cause := errors.New("account_info unavailable")
	q := &mockQuerier{accountErr: cause}

	_, err := BuildMergeInbox(q, BuildMergeInboxParams{
		Account:    testAccount,
		IssuanceID: testIssuanceID,
	})
	require.ErrorIs(t, err, ErrLedgerQuery)
	require.ErrorIs(t, err, cause)
}
