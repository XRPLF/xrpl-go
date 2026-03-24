package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestXChainCreateClaimID_TxType(t *testing.T) {
	tx := &XChainCreateClaimID{}
	require.Equal(t, XChainCreateClaimIDTx, tx.TxType())
}

func TestXChainCreateClaimID_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *XChainCreateClaimID
		expected FlatTransaction
	}{
		{
			name: "pass - valid tx",
			tx: &XChainCreateClaimID{
				BaseTx: BaseTx{
					Account: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				},
				OtherChainSource: "rMTi57fNy2UkUb4RcdoUeJm7gjxVQvxzUo",
				SignatureReward:  types.XRPCurrencyAmount(100),
				XChainBridge: types.XChainBridge{
					LockingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					LockingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
			},
			expected: FlatTransaction{
				"Account":          "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"TransactionType":  "XChainCreateClaimID",
				"OtherChainSource": "rMTi57fNy2UkUb4RcdoUeJm7gjxVQvxzUo",
				"SignatureReward":  "100",
				"XChainBridge": types.FlatXChainBridge{
					"LockingChainDoor":  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"IssuingChainDoor":  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"LockingChainIssue": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"IssuingChainIssue": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.tx.Flatten())
		})
	}
}

func TestXChainCreateClaimID_Validate(t *testing.T) {
	testcases := []struct {
		name        string
		tx          *XChainCreateClaimID
		expected    bool
		expectedErr error
	}{
		{
			name:        "fail - missing required fields",
			tx:          &XChainCreateClaimID{},
			expected:    false,
			expectedErr: ErrInvalidAccount,
		},
		{
			name: "fail - missing  other chain source",
			tx: &XChainCreateClaimID{
				BaseTx: BaseTx{
					Account:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					TransactionType: XChainCreateClaimIDTx,
				},
			},
			expected:    false,
			expectedErr: ErrInvalidAccount,
		},
		{
			name: "fail - missing signature reward",
			tx: &XChainCreateClaimID{
				BaseTx: BaseTx{
					Account:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					TransactionType: XChainCreateClaimIDTx,
				},
				OtherChainSource: "rMTi57fNy2UkUb4RcdoUeJm7gjxVQvxzUo",
			},
			expected:    false,
			expectedErr: ErrMissingField{Field: "SignatureReward"},
		},
		{
			name: "fail - invalid xchain bridge",
			tx: &XChainCreateClaimID{
				BaseTx: BaseTx{
					Account:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					TransactionType: XChainCreateClaimIDTx,
				},
				OtherChainSource: "rMTi57fNy2UkUb4RcdoUeJm7gjxVQvxzUo",
				SignatureReward:  types.XRPCurrencyAmount(100),
			},
			expected:    false,
			expectedErr: types.ErrInvalidIssuingChainDoorAddress,
		},
		{
			name: "pass - valid tx",
			tx: &XChainCreateClaimID{
				BaseTx: BaseTx{
					Account:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					TransactionType: XChainCreateClaimIDTx,
				},
				OtherChainSource: "rMTi57fNy2UkUb4RcdoUeJm7gjxVQvxzUo",
				SignatureReward:  types.XRPCurrencyAmount(100),
				XChainBridge: types.XChainBridge{
					LockingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					LockingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
			},
			expected:    true,
			expectedErr: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := tc.tx.Validate()
			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, ok)
			}
		})
	}
}
