package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestXChainAddClaimAttestation_TxType(t *testing.T) {
	tx := &XChainAddClaimAttestation{}
	require.Equal(t, XChainAddClaimAttestationTx, tx.TxType())
}

func TestXChainAddClaimAttestation_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *XChainAddClaimAttestation
		expected FlatTransaction
	}{
		{
			name: "pass - empty",
			tx:   &XChainAddClaimAttestation{},
			expected: FlatTransaction{
				"TransactionType":     "XChainAddClaimAttestation",
				"WasLockingChainSend": uint8(0),
			},
		},
		{
			name: "pass - full",
			tx: &XChainAddClaimAttestation{
				BaseTx: BaseTx{
					Account: "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: types.Address("rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es"),
				AttestationSignerAccount: types.Address("rEziJZmeZzsJvGVUmpUTey7qxQLKYxaK9f"),
				Destination:              types.Address("rKT9gDkaedAosiHyHZTjyZs2HvXpzuiGmC"),
				OtherChainSource:         types.Address("rnJmYAiqEVngtnb5ckRroXLtCbWC7CRUBx"),
				PublicKey:                "03ADB44CA8E56F78A0096825E5667C450ABD5C24C34E027BC1AAF7E5BD114CB5B5",
				Signature:                "3044022036C8B90F85E8073C465F00625248A72D4714600F98EBBADBAD3B7ED226109A3A02204C5A0AE12D169CF790F66541F3DB59C289E0D9CA7511FDFE352BB601F667A260",
				WasLockingChainSend:      1,
				XChainBridge: types.XChainBridge{
					LockingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainDoor:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					LockingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					IssuingChainIssue: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
				XChainClaimID: "0000000000000001",
			},
			expected: FlatTransaction{
				"Account":                  "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
				"TransactionType":          "XChainAddClaimAttestation",
				"Amount":                   "100",
				"AttestationRewardAccount": "rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es",
				"AttestationSignerAccount": "rEziJZmeZzsJvGVUmpUTey7qxQLKYxaK9f",
				"Destination":              "rKT9gDkaedAosiHyHZTjyZs2HvXpzuiGmC",
				"OtherChainSource":         "rnJmYAiqEVngtnb5ckRroXLtCbWC7CRUBx",
				"PublicKey":                "03ADB44CA8E56F78A0096825E5667C450ABD5C24C34E027BC1AAF7E5BD114CB5B5",
				"Signature":                "3044022036C8B90F85E8073C465F00625248A72D4714600F98EBBADBAD3B7ED226109A3A02204C5A0AE12D169CF790F66541F3DB59C289E0D9CA7511FDFE352BB601F667A260",
				"WasLockingChainSend":      uint8(1),
				"XChainBridge": types.FlatXChainBridge{
					"LockingChainDoor":  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"IssuingChainDoor":  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"LockingChainIssue": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"IssuingChainIssue": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
				"XChainClaimID": "0000000000000001",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.tx.Flatten())
		})
	}
}

func TestXChainAddClaimAttestation_Validate(t *testing.T) {
	testcases := []struct {
		name        string
		tx          *XChainAddClaimAttestation
		expected    bool
		expectedErr error
	}{
		{
			name:        "pass - empty",
			tx:          &XChainAddClaimAttestation{},
			expected:    false,
			expectedErr: ErrInvalidAccount,
		},
		{
			name: "fail - invalid amount",
			tx: &XChainAddClaimAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddClaimAttestationTx,
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
			tx: &XChainAddClaimAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddClaimAttestationTx,
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: types.Address("invalid"),
			},
			expected:    false,
			expectedErr: ErrInvalidAttestationRewardAccount,
		},
		{
			name: "fail - invalid attestation signer account",
			tx: &XChainAddClaimAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddClaimAttestationTx,
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: types.Address("rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es"),
				AttestationSignerAccount: types.Address("invalid"),
			},
			expected:    false,
			expectedErr: ErrInvalidAttestationSignerAccount,
		},
		{
			name: "fail - invalid destination",
			tx: &XChainAddClaimAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddClaimAttestationTx,
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: types.Address("rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es"),
				AttestationSignerAccount: types.Address("rEziJZmeZzsJvGVUmpUTey7qxQLKYxaK9f"),
				Destination:              types.Address("invalid"),
			},
			expected:    false,
			expectedErr: ErrInvalidDestination,
		},
		{
			name: "fail - invalid other chain source",
			tx: &XChainAddClaimAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddClaimAttestationTx,
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: types.Address("rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es"),
				AttestationSignerAccount: types.Address("rEziJZmeZzsJvGVUmpUTey7qxQLKYxaK9f"),
				Destination:              types.Address("rKT9gDkaedAosiHyHZTjyZs2HvXpzuiGmC"),
				OtherChainSource:         types.Address("invalid"),
			},
			expected:    false,
			expectedErr: ErrInvalidOtherChainSource,
		},
		{
			name: "fail - invalid public key",
			tx: &XChainAddClaimAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddClaimAttestationTx,
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: types.Address("rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es"),
				AttestationSignerAccount: types.Address("rEziJZmeZzsJvGVUmpUTey7qxQLKYxaK9f"),
				Destination:              types.Address("rKT9gDkaedAosiHyHZTjyZs2HvXpzuiGmC"),
				OtherChainSource:         types.Address("rnJmYAiqEVngtnb5ckRroXLtCbWC7CRUBx"),
				PublicKey:                "ABCD",
			},
			expected:    false,
			expectedErr: ErrInvalidPublicKey,
		},
		{
			name: "fail - invalid signature",
			tx: &XChainAddClaimAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddClaimAttestationTx,
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: types.Address("rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es"),
				AttestationSignerAccount: types.Address("rEziJZmeZzsJvGVUmpUTey7qxQLKYxaK9f"),
				Destination:              types.Address("rKT9gDkaedAosiHyHZTjyZs2HvXpzuiGmC"),
				OtherChainSource:         types.Address("rnJmYAiqEVngtnb5ckRroXLtCbWC7CRUBx"),
				PublicKey:                "03ADB44CA8E56F78A0096825E5667C450ABD5C24C34E027BC1AAF7E5BD114CB5B5",
				Signature:                "not-hex",
			},
			expected:    false,
			expectedErr: ErrInvalidSignature,
		},
		{
			name: "fail - invalid was locking chain send",
			tx: &XChainAddClaimAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddClaimAttestationTx,
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: types.Address("rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es"),
				AttestationSignerAccount: types.Address("rEziJZmeZzsJvGVUmpUTey7qxQLKYxaK9f"),
				Destination:              types.Address("rKT9gDkaedAosiHyHZTjyZs2HvXpzuiGmC"),
				OtherChainSource:         types.Address("rnJmYAiqEVngtnb5ckRroXLtCbWC7CRUBx"),
				PublicKey:                "03ADB44CA8E56F78A0096825E5667C450ABD5C24C34E027BC1AAF7E5BD114CB5B5",
				Signature:                "3044022036C8B90F85E8073C465F00625248A72D4714600F98EBBADBAD3B7ED226109A3A02204C5A0AE12D169CF790F66541F3DB59C289E0D9CA7511FDFE352BB601F667A260",
				WasLockingChainSend:      2,
			},
			expected:    false,
			expectedErr: ErrInvalidWasLockingChainSend,
		},
		{
			name: "fail - invalid xchain claim id",
			tx: &XChainAddClaimAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddClaimAttestationTx,
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: types.Address("rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es"),
				AttestationSignerAccount: types.Address("rEziJZmeZzsJvGVUmpUTey7qxQLKYxaK9f"),
				Destination:              types.Address("rKT9gDkaedAosiHyHZTjyZs2HvXpzuiGmC"),
				OtherChainSource:         types.Address("rnJmYAiqEVngtnb5ckRroXLtCbWC7CRUBx"),
				PublicKey:                "03ADB44CA8E56F78A0096825E5667C450ABD5C24C34E027BC1AAF7E5BD114CB5B5",
				Signature:                "3044022036C8B90F85E8073C465F00625248A72D4714600F98EBBADBAD3B7ED226109A3A02204C5A0AE12D169CF790F66541F3DB59C289E0D9CA7511FDFE352BB601F667A260",
				WasLockingChainSend:      1,
				XChainClaimID:            "",
			},
			expected:    false,
			expectedErr: ErrInvalidXChainClaimID,
		},
		{
			name: "fail - invalid xchain bridge",
			tx: &XChainAddClaimAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddClaimAttestationTx,
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: types.Address("rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es"),
				AttestationSignerAccount: types.Address("rEziJZmeZzsJvGVUmpUTey7qxQLKYxaK9f"),
				Destination:              types.Address("rKT9gDkaedAosiHyHZTjyZs2HvXpzuiGmC"),
				OtherChainSource:         types.Address("rnJmYAiqEVngtnb5ckRroXLtCbWC7CRUBx"),
				PublicKey:                "03ADB44CA8E56F78A0096825E5667C450ABD5C24C34E027BC1AAF7E5BD114CB5B5",
				Signature:                "3044022036C8B90F85E8073C465F00625248A72D4714600F98EBBADBAD3B7ED226109A3A02204C5A0AE12D169CF790F66541F3DB59C289E0D9CA7511FDFE352BB601F667A260",
				WasLockingChainSend:      1,
				XChainClaimID:            "0000000000000001",
			},
			expected:    false,
			expectedErr: types.ErrInvalidIssuingChainDoorAddress,
		},
		{
			name: "pass - valid tx",
			tx: &XChainAddClaimAttestation{
				BaseTx: BaseTx{
					Account:         "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
					TransactionType: XChainAddClaimAttestationTx,
				},
				Amount:                   types.XRPCurrencyAmount(100),
				AttestationRewardAccount: types.Address("rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es"),
				AttestationSignerAccount: types.Address("rEziJZmeZzsJvGVUmpUTey7qxQLKYxaK9f"),
				Destination:              types.Address("rKT9gDkaedAosiHyHZTjyZs2HvXpzuiGmC"),
				OtherChainSource:         types.Address("rnJmYAiqEVngtnb5ckRroXLtCbWC7CRUBx"),
				PublicKey:                "03ADB44CA8E56F78A0096825E5667C450ABD5C24C34E027BC1AAF7E5BD114CB5B5",
				Signature:                "3044022036C8B90F85E8073C465F00625248A72D4714600F98EBBADBAD3B7ED226109A3A02204C5A0AE12D169CF790F66541F3DB59C289E0D9CA7511FDFE352BB601F667A260",
				WasLockingChainSend:      1,
				XChainClaimID:            "0000000000000001",
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
