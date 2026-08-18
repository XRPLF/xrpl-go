package builder

import (
	"errors"
	"strings"
	"testing"

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
		{name: "fail - missing issuance ID", base: BuildMergeInboxParams{Account: testAccount}, wantErr: ErrMissingIssuanceID},
		{name: "fail - invalid issuance ID (not hex)", base: BuildMergeInboxParams{Account: testAccount, IssuanceID: strings.Repeat("GG", 24)}, wantErr: ErrInvalidIssuanceID},
		{name: "fail - invalid issuance ID (wrong length)", base: BuildMergeInboxParams{Account: testAccount, IssuanceID: "aabb"}, wantErr: ErrInvalidIssuanceID},
		{name: "fail - account is an X-address", base: BuildMergeInboxParams{Account: xAddressOf(t, testAccount), IssuanceID: testIssuanceID}, wantErr: ErrInvalidAccount},
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

func TestBuildMergeInbox_Pass(t *testing.T) {
	q := &mockQuerier{accountSeq: 42}

	result, err := BuildMergeInbox(q, BuildMergeInboxParams{
		Account:    testAccount,
		IssuanceID: testIssuanceID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, uint32(42), result.Sequence)
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
