package builder

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/stretchr/testify/require"
)

// TestMergeInboxBaseValidation verifies all validateMergeInboxBase branches through both entry points.
func TestMergeInboxBaseValidation(t *testing.T) {
	cases := []struct {
		name    string
		base    BuildMergeInboxParams
		wantErr error
	}{
		{name: "fail - missing account", base: BuildMergeInboxParams{IssuanceID: testIssuanceID}, wantErr: ErrMissingAccount},
		{name: "fail - missing issuance ID", base: BuildMergeInboxParams{Account: testAccount}, wantErr: ErrMissingIssuanceID},
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
		q := &mockQuerier{}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := BuildMergeInbox(q, tc.base)
				require.ErrorIs(t, err, tc.wantErr)
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
