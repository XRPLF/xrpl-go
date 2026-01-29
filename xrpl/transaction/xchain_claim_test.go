package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestXChainClaim_TxType(t *testing.T) {
	x := &XChainClaim{}
	require.Equal(t, x.TxType(), XChainClaimTx)
}

func TestXChainClaim_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *XChainClaim
		expected FlatTransaction
	}{
		{
			name: "pass - only base tx",
			tx: &XChainClaim{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
			},
			expected: FlatTransaction{
				"Account":         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"TransactionType": XChainClaimTx,
			},
		},
		{
			name: "pass - all fields set",
			tx: &XChainClaim{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				Amount:         types.XRPCurrencyAmount(1000000000),
				Destination:    "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				DestinationTag: types.DestinationTag(1),
				XChainBridge: types.XChainBridge{
					LockingChainDoor:  "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					LockingChainIssue: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					IssuingChainDoor:  "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					IssuingChainIssue: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				XChainClaimID: "1234567890",
			},
			expected: FlatTransaction{
				"TransactionType": XChainClaimTx,
				"Account":         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Amount":          types.XRPCurrencyAmount(1000000000).Flatten(),
				"Destination":     "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"DestinationTag":  uint32(1),
				"XChainBridge": types.FlatXChainBridge{
					"LockingChainDoor":  "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					"LockingChainIssue": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					"IssuingChainDoor":  "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					"IssuingChainIssue": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				"XChainClaimID": "1234567890",
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			require.Equal(t, testcase.tx.Flatten(), testcase.expected)
		})
	}
}

func TestXChainClaim_Validate(t *testing.T) {
	testcases := []struct {
		name        string
		tx          *XChainClaim
		expected    bool
		expectedErr error
	}{
		{
			name:        "fail - missing base tx",
			tx:          &XChainClaim{},
			expected:    false,
			expectedErr: ErrInvalidAccount,
		},
		{
			name: "fail - invalid amount",
			tx: &XChainClaim{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: XChainClaimTx,
				},
			},
			expected:    false,
			expectedErr: ErrMissingField{Field: "Amount"},
		},
		{
			name: "fail - invalid destination",
			tx: &XChainClaim{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: XChainClaimTx,
				},
				Amount:      types.XRPCurrencyAmount(1000000000),
				Destination: "invalid",
			},
			expected:    false,
			expectedErr: ErrInvalidDestinationAddress,
		},
		{
			name: "fail - invalid xchain bridge",
			tx: &XChainClaim{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: XChainClaimTx,
				},
				Amount:        types.XRPCurrencyAmount(1000000000),
				Destination:   "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				XChainClaimID: "1234567890",
			},
			expected:    false,
			expectedErr: types.ErrInvalidIssuingChainDoorAddress,
		},
		{
			name: "fail - missing xchain claim id",
			tx: &XChainClaim{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: XChainClaimTx,
				},
				Amount:      types.XRPCurrencyAmount(1000000000),
				Destination: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
			},
			expected:    false,
			expectedErr: ErrMissingXChainClaimID,
		},
		{
			name: "pass - all fields set",
			tx: &XChainClaim{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: XChainClaimTx,
				},
				Amount:         types.XRPCurrencyAmount(1000000000),
				Destination:    "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				DestinationTag: types.DestinationTag(1),
				XChainBridge: types.XChainBridge{
					LockingChainDoor:  "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					LockingChainIssue: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					IssuingChainDoor:  "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					IssuingChainIssue: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				XChainClaimID: "1234567890",
			},
			expected: true,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			ok, err := testcase.tx.Validate()
			if testcase.expectedErr != nil {
				require.Equal(t, err, testcase.expectedErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, ok, testcase.expected)
		})
	}
}
