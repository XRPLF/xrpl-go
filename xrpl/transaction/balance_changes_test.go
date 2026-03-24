package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/stretchr/testify/require"
)

func TestGetBalanceChanges(t *testing.T) {
	tt := []struct {
		name     string
		meta     *TxObjMeta
		expected []AccountBalanceChanges
	}{
		{
			name: "pass - USD payment to account with no USD",
			meta: &TxObjMeta{
				AffectedNodes: []AffectedNode{
					{
						ModifiedNode: &ModifiedNode{
							FinalFields: ledger.FlatLedgerObject{
								"Balance": map[string]any{
									"currency": "USD",
									"issuer":   "rrrrrrrrrrrrrrrrrrrrBZbvji",
									"value":    "1.535330905250352",
								},
								"Flags": 1114112,
								"HighLimit": map[string]any{
									"currency": "USD",
									"issuer":   "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
									"value":    "0",
								},
								"HighNode": "00000000000001E8",
								"LowLimit": map[string]any{
									"currency": "USD",
									"issuer":   "rKmBGxocj9Abgy25J51Mk1iqFzW9aVF9Tc",
									"value":    "1000000000",
								},
								"LowNode": "0000000000000000",
							},
							LedgerEntryType: ledger.RippleStateEntry,
							LedgerIndex:     "2F323020B4288ACD4066CC64C89DAD2E4D5DFC2D44571942A51C005BF79D6E25",
							PreviousFields: map[string]any{
								"Balance": map[string]any{
									"currency": "USD",
									"issuer":   "rrrrrrrrrrrrrrrrrrrrBZbvji",
									"value":    "1.545330905250352",
								},
							},
							PreviousTxnID:     "CEB7B6040C2989B9849C8D7E49F710457EDDE1D95ECDF1E298FD30CF2AC5BE11",
							PreviousTxnLgrSeq: 10424776,
						},
					},
					{
						ModifiedNode: &ModifiedNode{
							FinalFields: ledger.FlatLedgerObject{
								"Balance": map[string]any{
									"currency": "USD",
									"issuer":   "rrrrrrrrrrrrrrrrrrrrBZbvji",
									"value":    "0.01",
								},
								"Flags": 1114112,
								"HighLimit": map[string]any{
									"currency": "USD",
									"issuer":   "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
									"value":    "0",
								},
								"HighNode": "00000000000001E8",
								"LowLimit": map[string]any{
									"currency": "USD",
									"issuer":   "rLDYrujdKUfVx28T9vRDAbyJ7G2WVXKo4K",
									"value":    "1000000000",
								},
								"LowNode": "0000000000000000",
							},
							LedgerEntryType: ledger.RippleStateEntry,
							LedgerIndex:     "AAE13AF5192EFBFD49A8EEE5869595563FEB73228C0B38FED9CC3D20EE74F399",
							PreviousFields: map[string]any{
								"Balance": map[string]any{
									"currency": "USD",
									"issuer":   "rrrrrrrrrrrrrrrrrrrrBZbvji",
									"value":    "0",
								},
							},
							PreviousTxnID:     "A788447CF5FD7108CBF49416E2335F95ED3F5A9FC016686C8F9EFB34BBEA613A",
							PreviousTxnLgrSeq: 10425088,
						},
					},
					{
						ModifiedNode: &ModifiedNode{
							FinalFields: ledger.FlatLedgerObject{
								"Account":    "rKmBGxocj9Abgy25J51Mk1iqFzW9aVF9Tc",
								"Balance":    "239807992",
								"Flags":      0,
								"OwnerCount": 1,
								"Sequence":   17,
							},
							LedgerEntryType: ledger.AccountRootEntry,
							LedgerIndex:     "E9A39B0BA8703D5FFD05D9EAD01EE6C0E7A15CF33C2C6B7269107BD2BD535818",
							PreviousFields: map[string]any{
								"Balance":  "239819992",
								"Sequence": 16,
							},
							PreviousTxnID:     "3109F5A0F891CCA20B4D891EB7437973F40A7664C5176092EB2E5C0A949992AD",
							PreviousTxnLgrSeq: 10424942,
						},
					},
				},
				TransactionIndex:  3,
				TransactionResult: "tesSUCCESS",
			},
			expected: []AccountBalanceChanges{
				{
					Account: "rKmBGxocj9Abgy25J51Mk1iqFzW9aVF9Tc",
					Balances: []Balance{
						{
							Value:    "-0.01",
							Currency: "USD",
							Issuer:   "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
						},
						{
							Value:    "-0.012",
							Currency: "XRP",
						},
					},
				},
				{
					Account: "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
					Balances: []Balance{
						{
							Value:    "0.01",
							Currency: "USD",
							Issuer:   "rKmBGxocj9Abgy25J51Mk1iqFzW9aVF9Tc",
						},
						{
							Value:    "-0.01",
							Currency: "USD",
							Issuer:   "rLDYrujdKUfVx28T9vRDAbyJ7G2WVXKo4K",
						},
					},
				},
				{
					Account: "rLDYrujdKUfVx28T9vRDAbyJ7G2WVXKo4K",
					Balances: []Balance{
						{
							Value:    "0.01",
							Currency: "USD",
							Issuer:   "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
						},
					},
				},
			},
		},
		{
			name: "pass - XRP create account",
			meta: &TxObjMeta{
				AffectedNodes: []AffectedNode{
					{
						CreatedNode: &CreatedNode{
							LedgerEntryType: ledger.AccountRootEntry,
							LedgerIndex:     "C24354B286600B8F28E51233B4AC41A3B4DDD0FDC9BCF96BB171573F6B40A4AE",
							NewFields: ledger.FlatLedgerObject{
								"Account":  "rLDYrujdKUfVx28T9vRDAbyJ7G2WVXKo4K",
								"Balance":  "100000000",
								"Sequence": 1,
							},
						},
					},
					{
						ModifiedNode: &ModifiedNode{
							FinalFields: ledger.FlatLedgerObject{
								"Account":    "rKmBGxocj9Abgy25J51Mk1iqFzW9aVF9Tc",
								"Balance":    "339903994",
								"Flags":      0,
								"OwnerCount": 0,
								"Sequence":   9,
							},
							LedgerEntryType: ledger.AccountRootEntry,
							LedgerIndex:     "E9A39B0BA8703D5FFD05D9EAD01EE6C0E7A15CF33C2C6B7269107BD2BD535818",
							PreviousFields: map[string]any{
								"Balance":  "439915994",
								"Sequence": 8,
							},
							PreviousTxnID:     "0E6CF1A13C6A804BE50B08C1D0446C7405D8461254CC6B62337CA9FEA4DF13EC",
							PreviousTxnLgrSeq: 10424064,
						},
					},
				},
			},
			expected: []AccountBalanceChanges{
				{
					Account: "rKmBGxocj9Abgy25J51Mk1iqFzW9aVF9Tc",
					Balances: []Balance{
						{
							Value:    "-100.012",
							Currency: "XRP",
						},
					},
				},
				{
					Account: "rLDYrujdKUfVx28T9vRDAbyJ7G2WVXKo4K",
					Balances: []Balance{
						{
							Value:    "100",
							Currency: "XRP",
						},
					},
				},
			},
		},
		{
			name: "pass - USD payment of all USD in source account",
			meta: &TxObjMeta{
				AffectedNodes: []AffectedNode{
					{
						ModifiedNode: &ModifiedNode{
							FinalFields: ledger.FlatLedgerObject{
								"Balance": map[string]any{
									"currency": "USD",
									"issuer":   "rrrrrrrrrrrrrrrrrrrrBZbvji",
									"value":    "1.545330905250352",
								},
								"Flags": 1114112,
								"HighLimit": map[string]any{
									"currency": "USD",
									"issuer":   "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
									"value":    "0",
								},
								"HighNode": "00000000000001E8",
								"LowLimit": map[string]any{
									"currency": "USD",
									"issuer":   "rKmBGxocj9Abgy25J51Mk1iqFzW9aVF9Tc",
									"value":    "1000000000",
								},
								"LowNode": "0000000000000000",
							},
							LedgerEntryType: ledger.RippleStateEntry,
							LedgerIndex:     "2F323020B4288ACD4066CC64C89DAD2E4D5DFC2D44571942A51C005BF79D6E25",
							PreviousFields: map[string]any{
								"Balance": map[string]any{
									"currency": "USD",
									"issuer":   "rrrrrrrrrrrrrrrrrrrrBZbvji",
									"value":    "1.345330905250352",
								},
							},
							PreviousTxnID:     "24525F80080EAC8857F1A29A47AEF23FD2B0A52DAF7DC3900A4E31831187FCB1",
							PreviousTxnLgrSeq: 10443886,
						},
					},
					{
						ModifiedNode: &ModifiedNode{
							FinalFields: ledger.FlatLedgerObject{
								"Balance": map[string]any{
									"currency": "USD",
									"issuer":   "rrrrrrrrrrrrrrrrrrrrBZbvji",
									"value":    "0",
								},
								"Flags": 1114112,
								"HighLimit": map[string]any{
									"currency": "USD",
									"issuer":   "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
									"value":    "0",
								},
								"HighNode": "00000000000001E8",
								"LowLimit": map[string]any{
									"currency": "USD",
									"issuer":   "rLDYrujdKUfVx28T9vRDAbyJ7G2WVXKo4K",
									"value":    "1000000000",
								},
								"LowNode": "0000000000000000",
							},
							LedgerEntryType: ledger.RippleStateEntry,
							LedgerIndex:     "AAE13AF5192EFBFD49A8EEE5869595563FEB73228C0B38FED9CC3D20EE74F399",
							PreviousFields: map[string]any{
								"Balance": map[string]any{
									"currency": "USD",
									"issuer":   "rrrrrrrrrrrrrrrrrrrrBZbvji",
									"value":    "0.2",
								},
							},
							PreviousTxnID:     "24525F80080EAC8857F1A29A47AEF23FD2B0A52DAF7DC3900A4E31831187FCB1",
							PreviousTxnLgrSeq: 10443886,
						},
					},
					{
						ModifiedNode: &ModifiedNode{
							FinalFields: ledger.FlatLedgerObject{
								"Account":    "rLDYrujdKUfVx28T9vRDAbyJ7G2WVXKo4K",
								"Balance":    "99976002",
								"Flags":      0,
								"OwnerCount": 1,
								"Sequence":   3,
							},
							LedgerEntryType: ledger.AccountRootEntry,
							LedgerIndex:     "C24354B286600B8F28E51233B4AC41A3B4DDD0FDC9BCF96BB171573F6B40A4AE",
							PreviousFields: map[string]any{
								"Balance":  "99988002",
								"Sequence": 2,
							},
							PreviousTxnID:     "A788447CF5FD7108CBF49416E2335F95ED3F5A9FC016686C8F9EFB34BBEA613A",
							PreviousTxnLgrSeq: 10425088,
						},
					},
				},
			},
			expected: []AccountBalanceChanges{
				{
					Account: "rKmBGxocj9Abgy25J51Mk1iqFzW9aVF9Tc",
					Balances: []Balance{
						{
							Value:    "0.2",
							Currency: "USD",
							Issuer:   "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
						},
					},
				},
				{
					Account: "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
					Balances: []Balance{
						{
							Value:    "-0.2",
							Currency: "USD",
							Issuer:   "rKmBGxocj9Abgy25J51Mk1iqFzW9aVF9Tc",
						},

						{
							Value:    "0.2",
							Currency: "USD",
							Issuer:   "rLDYrujdKUfVx28T9vRDAbyJ7G2WVXKo4K",
						},
					},
				},
				{
					Account: "rLDYrujdKUfVx28T9vRDAbyJ7G2WVXKo4K",
					Balances: []Balance{
						{
							Value:    "-0.2",
							Currency: "USD",
							Issuer:   "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
						},

						{
							Value:    "-0.012",
							Currency: "XRP",
							Issuer:   "",
						},
					},
				},
			},
		},
		{
			name: "pass - USD payment to account with USD",
			meta: &TxObjMeta{
				AffectedNodes: []AffectedNode{
					{
						ModifiedNode: &ModifiedNode{
							FinalFields: ledger.FlatLedgerObject{
								"Balance": map[string]any{
									"currency": "USD",
									"issuer":   "rrrrrrrrrrrrrrrrrrrrBZbvji",
									"value":    "1.525330905250352",
								},
								"Flags": 1114112,
								"HighLimit": map[string]any{
									"currency": "USD",
									"issuer":   "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
									"value":    "0",
								},
								"HighNode": "00000000000001E8",
								"LowLimit": map[string]any{
									"currency": "USD",
									"issuer":   "rKmBGxocj9Abgy25J51Mk1iqFzW9aVF9Tc",
									"value":    "1000000000",
								},
								"LowNode": "0000000000000000",
							},
							LedgerEntryType: ledger.RippleStateEntry,
							LedgerIndex:     "2F323020B4288ACD4066CC64C89DAD2E4D5DFC2D44571942A51C005BF79D6E25",
							PreviousFields: map[string]any{
								"Balance": map[string]any{
									"currency": "USD",
									"issuer":   "rrrrrrrrrrrrrrrrrrrrBZbvji",
									"value":    "1.535330905250352",
								},
							},
							PreviousTxnID:     "DC061E6F47B1B6E9A496A31B1AF87194B4CB24B2EBF8A59F35E31E12509238BD",
							PreviousTxnLgrSeq: 10459364,
						},
					},
					{
						ModifiedNode: &ModifiedNode{
							FinalFields: map[string]any{
								"Balance": map[string]any{
									"currency": "USD",
									"issuer":   "rrrrrrrrrrrrrrrrrrrrBZbvji",
									"value":    "0.02",
								},
								"Flags": 1114112,
								"HighLimit": map[string]any{
									"currency": "USD",
									"issuer":   "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
									"value":    "0",
								},
								"HighNode": "00000000000001E8",
								"LowLimit": map[string]any{
									"currency": "USD",
									"issuer":   "rLDYrujdKUfVx28T9vRDAbyJ7G2WVXKo4K",
									"value":    "1000000000",
								},
								"LowNode": "0000000000000000",
							},
							LedgerEntryType: ledger.RippleStateEntry,
							LedgerIndex:     "AAE13AF5192EFBFD49A8EEE5869595563FEB73228C0B38FED9CC3D20EE74F399",
							PreviousFields: map[string]any{
								"Balance": map[string]any{
									"currency": "USD",
									"issuer":   "rrrrrrrrrrrrrrrrrrrrBZbvji",
									"value":    "0.01",
								},
							},
							PreviousTxnID:     "DC061E6F47B1B6E9A496A31B1AF87194B4CB24B2EBF8A59F35E31E12509238BD",
							PreviousTxnLgrSeq: 10459364,
						},
					},
					{
						ModifiedNode: &ModifiedNode{
							FinalFields: map[string]any{
								"Account":    "rKmBGxocj9Abgy25J51Mk1iqFzW9aVF9Tc",
								"Balance":    "239555992",
								"Flags":      0,
								"OwnerCount": 1,
								"Sequence":   38,
							},
							LedgerEntryType: ledger.AccountRootEntry,
							LedgerIndex:     "E9A39B0BA8703D5FFD05D9EAD01EE6C0E7A15CF33C2C6B7269107BD2BD535818",
							PreviousFields: map[string]any{
								"Balance":  "239567992",
								"Sequence": 37,
							},
							PreviousTxnID:     "DC061E6F47B1B6E9A496A31B1AF87194B4CB24B2EBF8A59F35E31E12509238BD",
							PreviousTxnLgrSeq: 10459364,
						},
					},
				},
			},
			expected: []AccountBalanceChanges{
				{
					Account: "rKmBGxocj9Abgy25J51Mk1iqFzW9aVF9Tc",
					Balances: []Balance{
						{
							Value:    "-0.01",
							Currency: "USD",
							Issuer:   "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
						},
						{
							Value:    "-0.012",
							Currency: "XRP",
						},
					},
				},
				{
					Account: "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
					Balances: []Balance{
						{
							Issuer:   "rKmBGxocj9Abgy25J51Mk1iqFzW9aVF9Tc",
							Currency: "USD",
							Value:    "0.01",
						},
						{
							Issuer:   "rLDYrujdKUfVx28T9vRDAbyJ7G2WVXKo4K",
							Currency: "USD",
							Value:    "-0.01",
						},
					},
				},
				{
					Account: "rLDYrujdKUfVx28T9vRDAbyJ7G2WVXKo4K",
					Balances: []Balance{
						{
							Issuer:   "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
							Currency: "USD",
							Value:    "0.01",
						},
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			balanceChanges, err := GetBalanceChanges(tc.meta)
			require.NoError(t, err)
			require.ElementsMatch(t, tc.expected, balanceChanges)
		})
	}
}
