package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestXChainModifyBridge_TxType(t *testing.T) {
	tx := &XChainModifyBridge{}
	require.Equal(t, tx.TxType(), XChainModifyBridgeTx)
}

func TestXChainModifyBridge_SetClearAccountCreateAmount(t *testing.T) {
	tx := &XChainModifyBridge{}
	tx.SetClearAccountCreateAmount()
	require.Equal(t, tx.Flags, TfClearAccountCreateAmount)
}

func TestXChainModifyBridge_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *XChainModifyBridge
		expected FlatTransaction
	}{
		{
			name: "pass - valid tx",
			tx: &XChainModifyBridge{
				BaseTx: BaseTx{
					Account: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				},
				Flags:                  TfClearAccountCreateAmount,
				MinAccountCreateAmount: types.XRPCurrencyAmount(100),
				SignatureReward:        types.XRPCurrencyAmount(10),
				XChainBridge: types.XChainBridge{
					LockingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					LockingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
			},
			expected: FlatTransaction{
				"Account":                "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"TransactionType":        "XChainModifyBridge",
				"Flags":                  uint32(TfClearAccountCreateAmount),
				"MinAccountCreateAmount": "100",
				"SignatureReward":        "10",
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
			require.Equal(t, tc.tx.Flatten(), tc.expected)
		})
	}
}

func TestXChainModifyBridge_Validate(t *testing.T) {
	testcases := []struct {
		name        string
		tx          *XChainModifyBridge
		expected    bool
		expectedErr error
	}{
		{
			name:        "fail - invalid account",
			tx:          &XChainModifyBridge{},
			expected:    false,
			expectedErr: ErrInvalidAccount,
		},
		{
			name: "fail - invalid flags",
			tx: &XChainModifyBridge{
				BaseTx: BaseTx{
					Account:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					TransactionType: XChainModifyBridgeTx,
				},
			},
			expected:    false,
			expectedErr: ErrInvalidFlags,
		},
		{
			name: "fail - invalid min account create amount",
			tx: &XChainModifyBridge{
				BaseTx: BaseTx{
					Account:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					TransactionType: XChainModifyBridgeTx,
				},
				Flags: TfClearAccountCreateAmount,
				MinAccountCreateAmount: types.IssuedCurrencyAmount{
					Currency: "XRP",
					Value:    "100",
				},
			},
			expected:    false,
			expectedErr: ErrInvalidTokenFields,
		},
		{
			name: "fail - invalid signature reward",
			tx: &XChainModifyBridge{
				BaseTx: BaseTx{
					Account:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					TransactionType: XChainModifyBridgeTx,
				},
				Flags:                  TfClearAccountCreateAmount,
				MinAccountCreateAmount: types.XRPCurrencyAmount(100),
				SignatureReward: types.IssuedCurrencyAmount{
					Currency: "XRP",
					Value:    "100",
				},
			},
			expected:    false,
			expectedErr: ErrInvalidTokenFields,
		},
		{
			name: "fail - invalid xchain bridge",
			tx: &XChainModifyBridge{
				BaseTx: BaseTx{
					Account:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					TransactionType: XChainModifyBridgeTx,
				},
				Flags: TfClearAccountCreateAmount,
			},
			expected:    false,
			expectedErr: types.ErrInvalidIssuingChainDoorAddress,
		},
		{
			name: "pass - valid tx",
			tx: &XChainModifyBridge{
				BaseTx: BaseTx{
					Account:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					TransactionType: XChainModifyBridgeTx,
				},
				Flags: TfClearAccountCreateAmount,
				XChainBridge: types.XChainBridge{
					LockingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					LockingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
			},
			expected: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := tc.tx.Validate()
			if tc.expectedErr != nil {
				require.Equal(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, ok, tc.expected)
			}
		})
	}
}
