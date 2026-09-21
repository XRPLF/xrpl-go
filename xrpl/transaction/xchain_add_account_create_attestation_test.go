package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestXChainAddAccountCreateAttestation_TxType(t *testing.T) {
	tx := &XChainAddAccountCreateAttestation{}
	require.Equal(t, XChainAddAccountCreateAttestationTx, tx.TxType())
}

func TestXChainAddAccountCreateAttestation_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *XChainAddAccountCreateAttestation
		expected FlatTransaction
	}{
		{
			name: "pass - empty",
			tx:   &XChainAddAccountCreateAttestation{},
			expected: FlatTransaction{
				"TransactionType":     "XChainAddAccountCreateAttestation",
				"WasLockingChainSend": uint8(0),
			},
		},
		{
			name: "pass - complete tx",
			tx: &XChainAddAccountCreateAttestation{
				BaseTx: BaseTx{
					Account: "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: "rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es",
				AttestationSignerAccount: "rpWLegmW9WrFBzHUj7brhQNZzrxgLj9oxw",
				Destination:              "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
				OtherChainSource:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
				PublicKey:                "03ADB44CA8E56F78A0096825E5667C450ABD5C24C34E027BC1AAF7E5BD114CB5B5",
				Signature:                "3044022076FC043240000000000000000000000000000000000000000000000000000000022076FC04324000000000000000000000000000000000000000000000000000000000",
				SignatureReward:          types.XRPCurrencyAmount(204),
				WasLockingChainSend:      1,
				XChainAccountCreateCount: "2",
				XChainBridge: types.XChainBridge{
					LockingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					LockingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
			},
			expected: FlatTransaction{
				"Account":                  "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
				"TransactionType":          "XChainAddAccountCreateAttestation",
				"Amount":                   "100",
				"AttestationRewardAccount": "rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es",
				"AttestationSignerAccount": "rpWLegmW9WrFBzHUj7brhQNZzrxgLj9oxw",
				"Destination":              "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
				"OtherChainSource":         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
				"PublicKey":                "03ADB44CA8E56F78A0096825E5667C450ABD5C24C34E027BC1AAF7E5BD114CB5B5",
				"Signature":                "3044022076FC043240000000000000000000000000000000000000000000000000000000022076FC04324000000000000000000000000000000000000000000000000000000000",
				"SignatureReward":          "204",
				"WasLockingChainSend":      uint8(1),
				"XChainAccountCreateCount": "2",
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

func TestXChainAddAccountCreateAttestation_Validate(t *testing.T) {
	testcases := []struct {
		name        string
		tx          *XChainAddAccountCreateAttestation
		expected    bool
		expectedErr error
	}{
		{
			name:        "pass - empty",
			tx:          &XChainAddAccountCreateAttestation{},
			expected:    false,
			expectedErr: ErrInvalidAccount,
		},
		{
			name: "fail - invalid amount",
			tx: &XChainAddAccountCreateAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddAccountCreateAttestationTx,
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
			name: "fail - invalid attestation reward account",
			tx: &XChainAddAccountCreateAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddAccountCreateAttestationTx,
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: "invalid",
			},
			expected:    false,
			expectedErr: ErrInvalidAttestationRewardAccount,
		},
		{
			name: "fail - invalid attestation signer account",
			tx: &XChainAddAccountCreateAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddAccountCreateAttestationTx,
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: types.Address("rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es"),
				AttestationSignerAccount: "invalid",
			},
			expected:    false,
			expectedErr: ErrInvalidAttestationSignerAccount,
		},
		{
			name: "fail - invalid destination",
			tx: &XChainAddAccountCreateAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddAccountCreateAttestationTx,
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: types.Address("rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es"),
				AttestationSignerAccount: types.Address("rpWLegmW9WrFBzHUj7brhQNZzrxgLj9oxw"),
				Destination:              "invalid",
			},
			expected:    false,
			expectedErr: ErrInvalidDestination,
		},
		{
			name: "fail - invalid other chain source",
			tx: &XChainAddAccountCreateAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddAccountCreateAttestationTx,
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: types.Address("rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es"),
				AttestationSignerAccount: types.Address("rpWLegmW9WrFBzHUj7brhQNZzrxgLj9oxw"),
				Destination:              types.Address("rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo"),
				OtherChainSource:         "invalid",
			},
			expected:    false,
			expectedErr: ErrInvalidOtherChainSource,
		},
		{
			name: "fail - invalid public key",
			tx: &XChainAddAccountCreateAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddAccountCreateAttestationTx,
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: types.Address("rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es"),
				AttestationSignerAccount: types.Address("rpWLegmW9WrFBzHUj7brhQNZzrxgLj9oxw"),
				Destination:              types.Address("rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo"),
				OtherChainSource:         types.Address("rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo"),
				PublicKey:                "ABCD",
			},
			expected:    false,
			expectedErr: ErrInvalidPublicKey,
		},
		{
			name: "fail - invalid signature",
			tx: &XChainAddAccountCreateAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddAccountCreateAttestationTx,
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: types.Address("rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es"),
				AttestationSignerAccount: types.Address("rpWLegmW9WrFBzHUj7brhQNZzrxgLj9oxw"),
				Destination:              types.Address("rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo"),
				OtherChainSource:         types.Address("rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo"),
				PublicKey:                "03ADB44CA8E56F78A0096825E5667C450ABD5C24C34E027BC1AAF7E5BD114CB5B5",
				Signature:                "not-hex",
			},
			expected:    false,
			expectedErr: ErrInvalidSignature,
		},
		{
			name: "fail - invalid signature reward",
			tx: &XChainAddAccountCreateAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddAccountCreateAttestationTx,
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: types.Address("rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es"),
				AttestationSignerAccount: types.Address("rpWLegmW9WrFBzHUj7brhQNZzrxgLj9oxw"),
				Destination:              types.Address("rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo"),
				OtherChainSource:         types.Address("rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo"),
				PublicKey:                "03ADB44CA8E56F78A0096825E5667C450ABD5C24C34E027BC1AAF7E5BD114CB5B5",
				Signature:                "3044022076FC043240000000000000000000000000000000000000000000000000000000022076FC04324000000000000000000000000000000000000000000000000000000000",
				SignatureReward: types.IssuedCurrencyAmount{
					Value:    "100",
					Currency: "XRP",
				},
			},
			expected:    false,
			expectedErr: ErrInvalidTokenFields,
		},
		{
			name: "fail - invalid was locking chain send",
			tx: &XChainAddAccountCreateAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddAccountCreateAttestationTx,
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: types.Address("rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es"),
				AttestationSignerAccount: types.Address("rpWLegmW9WrFBzHUj7brhQNZzrxgLj9oxw"),
				Destination:              types.Address("rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo"),
				OtherChainSource:         types.Address("rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo"),
				PublicKey:                "03ADB44CA8E56F78A0096825E5667C450ABD5C24C34E027BC1AAF7E5BD114CB5B5",
				Signature:                "3044022076FC043240000000000000000000000000000000000000000000000000000000022076FC04324000000000000000000000000000000000000000000000000000000000",
				SignatureReward:          types.XRPCurrencyAmount(204),
				WasLockingChainSend:      2,
			},
			expected:    false,
			expectedErr: ErrInvalidWasLockingChainSend,
		},
		{
			name: "fail - invalid x chain account create count",
			tx: &XChainAddAccountCreateAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddAccountCreateAttestationTx,
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: types.Address("rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es"),
				AttestationSignerAccount: types.Address("rpWLegmW9WrFBzHUj7brhQNZzrxgLj9oxw"),
				Destination:              types.Address("rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo"),
				OtherChainSource:         types.Address("rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo"),
				PublicKey:                "03ADB44CA8E56F78A0096825E5667C450ABD5C24C34E027BC1AAF7E5BD114CB5B5",
				Signature:                "3044022076FC043240000000000000000000000000000000000000000000000000000000022076FC04324000000000000000000000000000000000000000000000000000000000",
				SignatureReward:          types.XRPCurrencyAmount(204),
				WasLockingChainSend:      1,
				XChainAccountCreateCount: "",
			},
			expected:    false,
			expectedErr: ErrInvalidXChainAccountCreateCount,
		},
		{
			name: "fail - invalid x chain bridge",
			tx: &XChainAddAccountCreateAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddAccountCreateAttestationTx,
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: types.Address("rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es"),
				AttestationSignerAccount: types.Address("rpWLegmW9WrFBzHUj7brhQNZzrxgLj9oxw"),
				Destination:              types.Address("rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo"),
				OtherChainSource:         types.Address("rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo"),
				PublicKey:                "03ADB44CA8E56F78A0096825E5667C450ABD5C24C34E027BC1AAF7E5BD114CB5B5",
				Signature:                "3044022076FC043240000000000000000000000000000000000000000000000000000000022076FC04324000000000000000000000000000000000000000000000000000000000",
				SignatureReward:          types.XRPCurrencyAmount(204),
				WasLockingChainSend:      1,
				XChainAccountCreateCount: "2",
			},
			expected:    false,
			expectedErr: types.ErrInvalidIssuingChainDoorAddress,
		},
		{
			name: "pass - complete tx",
			tx: &XChainAddAccountCreateAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddAccountCreateAttestationTx,
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: types.Address("rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es"),
				AttestationSignerAccount: types.Address("rpWLegmW9WrFBzHUj7brhQNZzrxgLj9oxw"),
				Destination:              types.Address("rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo"),
				OtherChainSource:         types.Address("rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo"),
				PublicKey:                "03ADB44CA8E56F78A0096825E5667C450ABD5C24C34E027BC1AAF7E5BD114CB5B5",
				Signature:                "3044022076FC043240000000000000000000000000000000000000000000000000000000022076FC04324000000000000000000000000000000000000000000000000000000000",
				SignatureReward:          types.XRPCurrencyAmount(204),
				WasLockingChainSend:      1,
				XChainAccountCreateCount: "2",
				XChainBridge: types.XChainBridge{
					LockingChainDoor:  types.Address("rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"),
					IssuingChainDoor:  types.Address("rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"),
					LockingChainIssue: types.Address("rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"),
					IssuingChainIssue: types.Address("rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"),
				},
			},
			expected: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := tc.tx.Validate()
			if tc.expectedErr != nil {
				require.Equal(t, tc.expectedErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, ok)
			}
		})
	}
}
