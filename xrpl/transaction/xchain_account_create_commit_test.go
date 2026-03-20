package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXChainAccountCreateCommit_TxType(t *testing.T) {
	tx := &XChainAccountCreateCommit{}
	assert.Equal(t, XChainAccountCreateCommitTx, tx.TxType())
}

func TestXChainAccountCreateCommit_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *XChainAccountCreateCommit
		expected FlatTransaction
	}{
		{
			name: "pass - empty",
			tx:   &XChainAccountCreateCommit{},
			expected: FlatTransaction{
				"TransactionType": "XChainAccountCreateCommit",
			},
		},
		{
			name: "pass - full",
			tx: &XChainAccountCreateCommit{
				BaseTx: BaseTx{
					Account: "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
				},
				Destination:     "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
				Amount:          types.XRPCurrencyAmount(100),
				SignatureReward: types.XRPCurrencyAmount(1),
				XChainBridge: types.XChainBridge{
					LockingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					LockingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
			},
			expected: FlatTransaction{
				"Account":         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
				"TransactionType": "XChainAccountCreateCommit",
				"Destination":     "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
				"Amount":          "100",
				"SignatureReward": "1",
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

func TestXChainAccountCreateCommit_Validate(t *testing.T) {
	testcases := []struct {
		name        string
		tx          *XChainAccountCreateCommit
		expected    bool
		expectedErr error
	}{
		{
			name:        "fail - invalid account",
			tx:          &XChainAccountCreateCommit{},
			expected:    false,
			expectedErr: ErrInvalidAccount,
		},
		{
			name: "fail - invalid amount",
			tx: &XChainAccountCreateCommit{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAccountCreateCommitTx,
				},
				Amount: types.IssuedCurrencyAmount{
					Value:    "100",
					Currency: "XRP",
				},
			},
			expected:    false,
			expectedErr: ErrInvalidTokenFields,
		},
		{
			name: "fail - invalid destination",
			tx: &XChainAccountCreateCommit{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAccountCreateCommitTx,
				},
				Amount:      types.XRPCurrencyAmount(100),
				Destination: "invalid",
			},
			expected:    false,
			expectedErr: ErrInvalidAccount,
		},
		{
			name: "fail - invalid signature reward",
			tx: &XChainAccountCreateCommit{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAccountCreateCommitTx,
				},
				Amount:      types.XRPCurrencyAmount(100),
				Destination: "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
				SignatureReward: types.IssuedCurrencyAmount{
					Value:    "1",
					Currency: "XRP",
				},
			},
			expected:    false,
			expectedErr: ErrInvalidTokenFields,
		},
		{
			name: "fail - invalid bridge",
			tx: &XChainAccountCreateCommit{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAccountCreateCommitTx,
				},
				Amount:          types.XRPCurrencyAmount(100),
				Destination:     "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
				SignatureReward: types.XRPCurrencyAmount(1),
			},
			expected:    false,
			expectedErr: types.ErrInvalidIssuingChainDoorAddress,
		},
		{
			name: "pass - valid",
			tx: &XChainAccountCreateCommit{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAccountCreateCommitTx,
				},
				Amount:          types.XRPCurrencyAmount(100),
				Destination:     "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
				SignatureReward: types.XRPCurrencyAmount(1),
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
				require.Error(t, err)
				require.Equal(t, tc.expectedErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, ok)
			}
		})
	}
}
