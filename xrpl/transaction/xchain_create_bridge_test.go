package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestXChainCreateBridge_TxType(t *testing.T) {
	tx := &XChainCreateBridge{}
	require.Equal(t, XChainCreateBridgeTx, tx.TxType())
}

func TestXChainCreateBridge_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *XChainCreateBridge
		expected FlatTransaction
	}{
		{
			name: "pass - no optional fields",
			tx: &XChainCreateBridge{
				BaseTx: BaseTx{
					Account: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				},
				SignatureReward: types.XRPCurrencyAmount(0),
				XChainBridge: types.XChainBridge{
					LockingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					LockingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
			},
			expected: FlatTransaction{
				"Account":         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"TransactionType": "XChainCreateBridge",
				"SignatureReward": "0",
				"XChainBridge": types.FlatXChainBridge{
					"LockingChainDoor":  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"LockingChainIssue": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"IssuingChainDoor":  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"IssuingChainIssue": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
			},
		},
		{
			name: "pass - all fields present",
			tx: &XChainCreateBridge{
				BaseTx: BaseTx{
					Account: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				},
				MinAccountCreateAmount: types.XRPCurrencyAmount(10000),
				SignatureReward:        types.XRPCurrencyAmount(10000),
				XChainBridge: types.XChainBridge{
					LockingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					LockingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
			},
			expected: FlatTransaction{
				"Account":                "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"TransactionType":        "XChainCreateBridge",
				"MinAccountCreateAmount": "10000",
				"SignatureReward":        "10000",
				"XChainBridge": types.FlatXChainBridge{
					"LockingChainDoor":  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"LockingChainIssue": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"IssuingChainDoor":  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"IssuingChainIssue": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			require.Equal(t, testcase.expected, testcase.tx.Flatten())
		})
	}
}

func TestXChainCreateBridge_Validate(t *testing.T) {
	testcases := []struct {
		name        string
		tx          *XChainCreateBridge
		expected    bool
		expectedErr error
	}{
		{
			name:        "fail - missing required fields",
			tx:          &XChainCreateBridge{},
			expected:    false,
			expectedErr: ErrInvalidAccount,
		},
		{
			name: "fail - missing signature reward",
			tx: &XChainCreateBridge{
				BaseTx: BaseTx{
					Account:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					TransactionType: XChainCreateBridgeTx,
				},
			},
			expected:    false,
			expectedErr: ErrMissingField{Field: "SignatureReward"},
		},
		{
			name: "fail - missing xchain bridge",
			tx: &XChainCreateBridge{
				BaseTx: BaseTx{
					Account:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					TransactionType: XChainCreateBridgeTx,
				},
				SignatureReward: types.XRPCurrencyAmount(10000),
			},
			expected:    false,
			expectedErr: types.ErrInvalidIssuingChainDoorAddress,
		},
		{
			name: "fail - invalid min account create amount",
			tx: &XChainCreateBridge{
				BaseTx: BaseTx{
					Account:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					TransactionType: XChainCreateBridgeTx,
				},
				MinAccountCreateAmount: types.IssuedCurrencyAmount{
					Currency: "XRP",
					Value:    "test",
				},
			},
			expected:    false,
			expectedErr: ErrInvalidTokenFields,
		},
		{
			name: "pass - valid tx with required fields",
			tx: &XChainCreateBridge{
				BaseTx: BaseTx{
					Account:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					TransactionType: XChainCreateBridgeTx,
				},
				SignatureReward: types.XRPCurrencyAmount(10000),
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
		{
			name: "pass - valid tx with all fields",
			tx: &XChainCreateBridge{
				BaseTx: BaseTx{
					Account:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					TransactionType: XChainCreateBridgeTx,
				},
				MinAccountCreateAmount: types.XRPCurrencyAmount(10000),
				SignatureReward:        types.XRPCurrencyAmount(10000),
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

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			ok, err := testcase.tx.Validate()
			if testcase.expectedErr != nil {
				require.Equal(t, testcase.expectedErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, testcase.expected, ok)
			}
		})
	}
}
