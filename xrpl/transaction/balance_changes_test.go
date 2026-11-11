package transaction

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/internal/decimal"
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
			name: "exact XRP delta",
			meta: &TxObjMeta{AffectedNodes: []AffectedNode{{ModifiedNode: &ModifiedNode{
				LedgerEntryType: ledger.AccountRootEntry,
				FinalFields:     ledger.FlatLedgerObject{"Account": "rAccount", "Balance": "20000000009"},
				PreviousFields:  map[string]any{"Balance": "10000000000"},
			}}}},
			expected: []AccountBalanceChanges{{Account: "rAccount", Balances: []Balance{{Currency: "XRP", Value: "10000.000009"}}}},
		},
		{
			name: "exact issued decimal exponent",
			meta: &TxObjMeta{AffectedNodes: []AffectedNode{{ModifiedNode: &ModifiedNode{
				LedgerEntryType: ledger.RippleStateEntry,
				FinalFields: ledger.FlatLedgerObject{
					"Balance":   map[string]any{"currency": "USD", "value": "1.2E-7"},
					"LowLimit":  map[string]any{"issuer": "rLow"},
					"HighLimit": map[string]any{"issuer": "rHigh"},
				},
				PreviousFields: map[string]any{"Balance": map[string]any{"value": "1e-7"}},
			}}}},
			expected: []AccountBalanceChanges{
				{Account: "rLow", Balances: []Balance{{Currency: "USD", Issuer: "rHigh", Value: "2e-8"}}},
				{Account: "rHigh", Balances: []Balance{{Currency: "USD", Issuer: "rLow", Value: "-2e-8"}}},
			},
		},
		{
			name: "negative zero trustline is skipped",
			meta: &TxObjMeta{AffectedNodes: []AffectedNode{{CreatedNode: &CreatedNode{
				LedgerEntryType: ledger.RippleStateEntry,
				NewFields: ledger.FlatLedgerObject{
					"Balance":   map[string]any{"currency": "USD", "value": "-0"},
					"LowLimit":  map[string]any{"issuer": "rLow"},
					"HighLimit": map[string]any{"issuer": "rHigh"},
				},
			}}}},
			expected: []AccountBalanceChanges{},
		},
		{
			name: "pass - account delete",
			meta: &TxObjMeta{
				AffectedNodes: []AffectedNode{
					{
						ModifiedNode: &ModifiedNode{
							FinalFields: ledger.FlatLedgerObject{
								"Account": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
								"Balance": "1039483986",
							},
							LedgerEntryType: ledger.AccountRootEntry,
							PreviousFields: map[string]any{
								"Balance": "1026818191",
							},
						},
					},
					{
						DeletedNode: &DeletedNode{
							FinalFields: ledger.FlatLedgerObject{
								"Account": "rf2c74F1Z2BrvL1dV7WMHYo6Jyaw446Fre",
								"Balance": "0",
							},
							LedgerEntryType: ledger.AccountRootEntry,
							PreviousFields: map[string]any{
								"Balance": "14665795",
							},
						},
					},
				},
			},
			expected: []AccountBalanceChanges{
				{
					Account: "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
					Balances: []Balance{
						{
							Value:    "12.665795",
							Currency: "XRP",
						},
					},
				},
				{
					Account: "rf2c74F1Z2BrvL1dV7WMHYo6Jyaw446Fre",
					Balances: []Balance{
						{
							Value:    "-14.665795",
							Currency: "XRP",
						},
					},
				},
			},
		},
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
			name: "pass - skip trustline limit change without balance change",
			meta: &TxObjMeta{
				AffectedNodes: []AffectedNode{
					{
						ModifiedNode: &ModifiedNode{
							FinalFields: ledger.FlatLedgerObject{
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
								"LowLimit": map[string]any{
									"currency": "USD",
									"issuer":   "rLDYrujdKUfVx28T9vRDAbyJ7G2WVXKo4K",
									"value":    "200",
								},
							},
							LedgerEntryType: ledger.RippleStateEntry,
							PreviousFields: map[string]any{
								"LowLimit": map[string]any{
									"currency": "USD",
									"issuer":   "rLDYrujdKUfVx28T9vRDAbyJ7G2WVXKo4K",
									"value":    "100",
								},
							},
						},
					},
					{
						ModifiedNode: &ModifiedNode{
							FinalFields: ledger.FlatLedgerObject{
								"Account":  "rLDYrujdKUfVx28T9vRDAbyJ7G2WVXKo4K",
								"Balance":  "99884302",
								"Sequence": 11,
							},
							LedgerEntryType: ledger.AccountRootEntry,
							PreviousFields: map[string]any{
								"Balance":  "99896302",
								"Sequence": 10,
							},
						},
					},
				},
			},
			expected: []AccountBalanceChanges{
				{
					Account: "rLDYrujdKUfVx28T9vRDAbyJ7G2WVXKo4K",
					Balances: []Balance{
						{
							// 99884302 - 99896302 = -12000 drops = -0.012 XRP.
							Value:    "-0.012",
							Currency: "XRP",
						},
					},
				},
			},
		},
		{
			name: "pass - skip account root owner count change without balance change",
			meta: &TxObjMeta{
				AffectedNodes: []AffectedNode{
					{
						ModifiedNode: &ModifiedNode{
							FinalFields: ledger.FlatLedgerObject{
								"Account":    "rLDYrujdKUfVx28T9vRDAbyJ7G2WVXKo4K",
								"Balance":    "239567992",
								"Flags":      0,
								"OwnerCount": 2,
								"Sequence":   37,
							},
							LedgerEntryType: ledger.AccountRootEntry,
							PreviousFields: map[string]any{
								"OwnerCount": 1,
							},
						},
					},
					{
						ModifiedNode: &ModifiedNode{
							FinalFields: ledger.FlatLedgerObject{
								"Account":  "rKmBGxocj9Abgy25J51Mk1iqFzW9aVF9Tc",
								"Balance":  "239555992",
								"Sequence": 38,
							},
							LedgerEntryType: ledger.AccountRootEntry,
							PreviousFields: map[string]any{
								"Balance":  "239567992",
								"Sequence": 37,
							},
						},
					},
				},
			},
			expected: []AccountBalanceChanges{
				{
					Account: "rKmBGxocj9Abgy25J51Mk1iqFzW9aVF9Tc",
					Balances: []Balance{
						{
							// 239555992 - 239567992 = -12000 drops = -0.012 XRP.
							Value:    "-0.012",
							Currency: "XRP",
						},
					},
				},
			},
		},
		{
			name: "pass - skip deleted zero balance trustline without balance change",
			meta: &TxObjMeta{
				AffectedNodes: []AffectedNode{
					{
						DeletedNode: &DeletedNode{
							FinalFields: ledger.FlatLedgerObject{
								"Balance": map[string]any{
									"currency": "USD",
									"issuer":   "rrrrrrrrrrrrrrrrrrrrBZbvji",
									"value":    "0",
								},
								"HighLimit": map[string]any{
									"currency": "USD",
									"issuer":   "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
									"value":    "0",
								},
								"LowLimit": map[string]any{
									"currency": "USD",
									"issuer":   "rLDYrujdKUfVx28T9vRDAbyJ7G2WVXKo4K",
									"value":    "0",
								},
							},
							LedgerEntryType: ledger.RippleStateEntry,
							PreviousFields: map[string]any{
								"Flags": 65536,
							},
						},
					},
					{
						ModifiedNode: &ModifiedNode{
							FinalFields: ledger.FlatLedgerObject{
								"Account":  "rLDYrujdKUfVx28T9vRDAbyJ7G2WVXKo4K",
								"Balance":  "99752302",
								"Sequence": 22,
							},
							LedgerEntryType: ledger.AccountRootEntry,
							PreviousFields: map[string]any{
								"Balance":  "99764302",
								"Sequence": 21,
							},
						},
					},
				},
			},
			expected: []AccountBalanceChanges{
				{
					Account: "rLDYrujdKUfVx28T9vRDAbyJ7G2WVXKo4K",
					Balances: []Balance{
						{
							// 99752302 - 99764302 = -12000 drops = -0.012 XRP.
							Value:    "-0.012",
							Currency: "XRP",
						},
					},
				},
			},
		},
		{
			name: "pass - skip account root zero balance delta",
			meta: &TxObjMeta{
				AffectedNodes: []AffectedNode{
					{
						ModifiedNode: &ModifiedNode{
							FinalFields: ledger.FlatLedgerObject{
								"Account": "rKmBGxocj9Abgy25J51Mk1iqFzW9aVF9Tc",
								"Balance": "239567992",
							},
							LedgerEntryType: ledger.AccountRootEntry,
							PreviousFields: map[string]any{
								"Balance": "239567992",
							},
						},
					},
				},
			},
			expected: []AccountBalanceChanges{},
		},
		{
			name: "pass - skip ripple state zero balance delta",
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
								"HighLimit": map[string]any{
									"currency": "USD",
									"issuer":   "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
									"value":    "0",
								},
								"LowLimit": map[string]any{
									"currency": "USD",
									"issuer":   "rKmBGxocj9Abgy25J51Mk1iqFzW9aVF9Tc",
									"value":    "1000000000",
								},
							},
							LedgerEntryType: ledger.RippleStateEntry,
							PreviousFields: map[string]any{
								"Balance": map[string]any{
									"currency": "USD",
									"issuer":   "rrrrrrrrrrrrrrrrrrrrBZbvji",
									"value":    "1.545330905250352",
								},
							},
						},
					},
				},
			},
			expected: []AccountBalanceChanges{},
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

func TestGetBalanceChangesTrustlineValues(t *testing.T) {
	for _, tc := range []struct {
		name    string
		balance string
		low     string
		high    string
	}{
		{name: "explicit plus sign", balance: "+5", low: "5", high: "-5"},
		{name: "1e-6 plain", balance: "0.000001", low: "0.000001", high: "-0.000001"},
		{name: "1e-7 exponential", balance: "0.0000001", low: "1e-7", high: "-1e-7"},
		{name: "1e20 plain", balance: "1e20", low: "100000000000000000000", high: "-100000000000000000000"},
		{name: "1e21 exponential", balance: "1e21", low: "1e+21", high: "-1e+21"},
		{name: "exponent 256", balance: "1e256", low: "1e+256", high: "-1e+256"},
		{name: "exponent -256", balance: "1.5e-256", low: "1.5e-256", high: "-1.5e-256"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			meta := &TxObjMeta{AffectedNodes: []AffectedNode{{CreatedNode: &CreatedNode{
				LedgerEntryType: ledger.RippleStateEntry,
				NewFields: ledger.FlatLedgerObject{
					"Balance":   map[string]any{"currency": "USD", "value": tc.balance},
					"LowLimit":  map[string]any{"issuer": "rLow"},
					"HighLimit": map[string]any{"issuer": "rHigh"},
				},
			}}}}
			expected := []AccountBalanceChanges{
				{Account: "rLow", Balances: []Balance{{Currency: "USD", Issuer: "rHigh", Value: tc.low}}},
				{Account: "rHigh", Balances: []Balance{{Currency: "USD", Issuer: "rLow", Value: tc.high}}},
			}
			balanceChanges, err := GetBalanceChanges(meta)
			require.NoError(t, err)
			require.ElementsMatch(t, expected, balanceChanges)
		})
	}
}

func TestGetBalanceChangesPreservesDeletedNodePreviousFieldsFromJSON(t *testing.T) {
	const accountDeleteMetaJSON = `{
		"AffectedNodes": [
			{
				"DeletedNode": {
					"FinalFields": {
						"Account": "rf2c74F1Z2BrvL1dV7WMHYo6Jyaw446Fre",
						"Balance": "0"
					},
					"LedgerEntryType": "AccountRoot",
					"PreviousFields": {
						"Balance": "14665795"
					}
				}
			}
		],
		"TransactionIndex": 0,
		"TransactionResult": "tesSUCCESS"
	}`

	var meta TxObjMeta
	err := json.Unmarshal([]byte(accountDeleteMetaJSON), &meta)
	require.NoError(t, err)
	require.NotNil(t, meta.AffectedNodes[0].DeletedNode)
	require.Equal(t, "14665795", meta.AffectedNodes[0].DeletedNode.PreviousFields["Balance"])

	balanceChanges, err := GetBalanceChanges(&meta)

	require.NoError(t, err)
	require.ElementsMatch(t, []AccountBalanceChanges{
		{
			Account: "rf2c74F1Z2BrvL1dV7WMHYo6Jyaw446Fre",
			Balances: []Balance{
				{
					Value:    "-14.665795",
					Currency: "XRP",
				},
			},
		},
	}, balanceChanges)
}

func TestGetBalanceChangesReturnsInvalidBalanceErrors(t *testing.T) {
	for _, tc := range []struct {
		name    string
		balance any
	}{
		{"text", "not-a-balance"},
		{"number", 1},
		{"null", nil},
		{"missing value", map[string]any{}},
		{"number value", map[string]any{"value": 1}},
		{"null value", map[string]any{"value": nil}},
		{"too long", strings.Repeat("1", decimal.MaxInputLen+1)},
		{"infinity", "Inf"},
		{"large exponent", "1e999999"},
		{"exponent 257", "1e257"},
		{"exponent -257", "1e-257"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, meta := range []*TxObjMeta{
				{AffectedNodes: []AffectedNode{{CreatedNode: &CreatedNode{
					LedgerEntryType: ledger.AccountRootEntry,
					NewFields:       ledger.FlatLedgerObject{"Account": "rAccount", "Balance": tc.balance},
				}}}},
				{AffectedNodes: []AffectedNode{{ModifiedNode: &ModifiedNode{
					LedgerEntryType: ledger.AccountRootEntry,
					FinalFields:     ledger.FlatLedgerObject{"Account": "rAccount", "Balance": "1"},
					PreviousFields:  map[string]any{"Balance": tc.balance},
				}}}},
				{AffectedNodes: []AffectedNode{{ModifiedNode: &ModifiedNode{
					LedgerEntryType: ledger.AccountRootEntry,
					FinalFields:     ledger.FlatLedgerObject{"Account": "rAccount", "Balance": tc.balance},
					PreviousFields:  map[string]any{"Balance": "1"},
				}}}},
			} {
				_, err := GetBalanceChanges(meta)
				require.ErrorIs(t, err, errInvalidBalanceValue)
			}
		})
	}
}

func TestGetBalanceChangesReturnsInvalidFieldErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		fields ledger.FlatLedgerObject
		entry  ledger.EntryType
		want   error
	}{
		{"missing account", ledger.FlatLedgerObject{"Balance": "1"}, ledger.AccountRootEntry, errAccountNotFoundForXRPQuantity},
		{"null account", ledger.FlatLedgerObject{"Balance": "1", "Account": nil}, ledger.AccountRootEntry, errAccountNotFoundForXRPQuantity},
		{"number account", ledger.FlatLedgerObject{"Balance": "1", "Account": 1}, ledger.AccountRootEntry, errAccountNotFoundForXRPQuantity},
		{"missing low issuer", ledger.FlatLedgerObject{"Balance": map[string]any{"value": "1", "currency": "USD"}, "LowLimit": map[string]any{}}, ledger.RippleStateEntry, errLowLimitIssuerNotFound},
		{"number low issuer", ledger.FlatLedgerObject{"Balance": map[string]any{"value": "1", "currency": "USD"}, "LowLimit": map[string]any{"issuer": 1}}, ledger.RippleStateEntry, errLowLimitIssuerNotFound},
		{"null high issuer", ledger.FlatLedgerObject{"Balance": map[string]any{"value": "1", "currency": "USD"}, "LowLimit": map[string]any{"issuer": "rLow"}, "HighLimit": map[string]any{"issuer": nil}}, ledger.RippleStateEntry, errHighLimitIssuerNotFound},
		{"missing currency", ledger.FlatLedgerObject{"Balance": map[string]any{"value": "1"}, "LowLimit": map[string]any{"issuer": "rLow"}, "HighLimit": map[string]any{"issuer": "rHigh"}}, ledger.RippleStateEntry, errBalanceCurrencyNotFound},
		{"number currency", ledger.FlatLedgerObject{"Balance": map[string]any{"value": "1", "currency": 1}, "LowLimit": map[string]any{"issuer": "rLow"}, "HighLimit": map[string]any{"issuer": "rHigh"}}, ledger.RippleStateEntry, errBalanceCurrencyNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			meta := &TxObjMeta{AffectedNodes: []AffectedNode{{CreatedNode: &CreatedNode{LedgerEntryType: tc.entry, NewFields: tc.fields}}}}
			_, err := GetBalanceChanges(meta)
			require.ErrorIs(t, err, tc.want)
		})
	}
}
