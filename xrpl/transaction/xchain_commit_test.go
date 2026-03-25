package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestXChainCommit_TxType(t *testing.T) {
	tx := &XChainCommit{}
	require.Equal(t, XChainCommitTx, tx.TxType())
}

func TestXChainCommit_Flatten(t *testing.T) {
	testcases := []struct {
		name        string
		tx          *XChainCommit
		expected    FlatTransaction
		expectedErr error
	}{
		{
			name: "pass - required fields",
			tx: &XChainCommit{
				BaseTx: BaseTx{
					Account: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				},
				Amount: types.XRPCurrencyAmount(10000),
				XChainBridge: types.XChainBridge{
					LockingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					LockingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
				XChainClaimID: "13f",
			},
			expected: FlatTransaction{
				"Account":         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"TransactionType": "XChainCommit",
				"Amount":          "10000",
				"XChainClaimID":   "13f",
				"XChainBridge": types.FlatXChainBridge{
					"LockingChainDoor":  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"LockingChainIssue": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"IssuingChainDoor":  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"IssuingChainIssue": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
			},
		},
		{
			name: "pass - optional xchain commit fields",
			tx: &XChainCommit{
				BaseTx: BaseTx{
					Account: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				},
				Amount: types.XRPCurrencyAmount(10000),
				XChainBridge: types.XChainBridge{
					LockingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					LockingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
				XChainClaimID:         "13f",
				OtherChainDestination: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			},
			expected: FlatTransaction{
				"Account":         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"TransactionType": "XChainCommit",
				"Amount":          "10000",
				"XChainClaimID":   "13f",
				"XChainBridge": types.FlatXChainBridge{
					"LockingChainDoor":  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"LockingChainIssue": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"IssuingChainDoor":  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"IssuingChainIssue": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
				"OtherChainDestination": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			flattened := testcase.tx.Flatten()
			require.Equal(t, testcase.expected, flattened)
		})
	}
}

func TestXChainCommit_Validate(t *testing.T) {
	testcases := []struct {
		name        string
		tx          *XChainCommit
		expected    bool
		expectedErr error
	}{
		{
			name:        "fail - missing required fields",
			tx:          &XChainCommit{},
			expected:    false,
			expectedErr: ErrInvalidAccount,
		},
		{
			name: "fail - missing amount",
			tx: &XChainCommit{
				BaseTx: BaseTx{
					Account:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					TransactionType: XChainCommitTx,
				},
			},
			expected:    false,
			expectedErr: ErrMissingField{Field: "Amount"},
		},
		{
			name: "fail - missing xchain bridge",
			tx: &XChainCommit{
				BaseTx: BaseTx{
					Account:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					TransactionType: XChainCommitTx,
				},
				Amount: types.XRPCurrencyAmount(10000),
			},
			expected:    false,
			expectedErr: types.ErrInvalidIssuingChainDoorAddress,
		},
		{
			name: "fail - missing xchain claim id",
			tx: &XChainCommit{
				BaseTx: BaseTx{
					Account:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					TransactionType: XChainCommitTx,
				},
				Amount: types.XRPCurrencyAmount(10000),
				XChainBridge: types.XChainBridge{
					LockingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					LockingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
			},
			expected:    false,
			expectedErr: ErrInvalidXChainClaimID,
		},
		{
			name: "pass - valid tx",
			tx: &XChainCommit{
				BaseTx: BaseTx{
					Account:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					TransactionType: XChainCommitTx,
				},
				Amount: types.XRPCurrencyAmount(10000),
				XChainBridge: types.XChainBridge{
					LockingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					LockingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
				XChainClaimID: "13f",
			},
			expected: true,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			valid, err := testcase.tx.Validate()
			if testcase.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, testcase.expectedErr, err)
			} else {
				require.NoError(t, err)
				require.True(t, valid)
			}
		})
	}
}
