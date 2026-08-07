package binarycodec

import (
	"maps"
	"testing"

	"github.com/Peersyst/xrpl-go/binary-codec/types"
	"github.com/stretchr/testify/require"
)

func TestTransactionRawFieldValueOverrides(t *testing.T) {
	overrides := transactionRawFieldValueOverrides(map[string]any{"TransactionType": "UNLModify"})
	require.Equal(t, types.RawFieldValueOverrides{"Account": []byte{}}, overrides)

	tests := []struct {
		name            string
		transactionType any
	}{
		{name: "other string", transactionType: "Payment"},
		{name: "other pseudo-transaction string", transactionType: "SetFee"},
		{name: "another pseudo-transaction string", transactionType: "EnableAmendment"},
		{name: "numeric code", transactionType: uint16(102)},
		{name: "nil", transactionType: nil},
		{name: "uncomparable slice", transactionType: []any{"UNLModify"}},
		{name: "uncomparable map", transactionType: map[string]any{"name": "UNLModify"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var overrides types.RawFieldValueOverrides
			require.NotPanics(t, func() {
				overrides = transactionRawFieldValueOverrides(map[string]any{
					"TransactionType": test.transactionType,
				})
			})
			require.Nil(t, overrides)
		})
	}
}

func TestEncodeUNLModifyAccountOverride(t *testing.T) {
	// This vector is mainnet transaction 80CDD04AC3C26F02C678881C546280C116648C9B116F87320B1CE68490F13907:
	// https://livenet.xrpl.org/transactions/80CDD04AC3C26F02C678881C546280C116648C9B116F87320B1CE68490F13907
	const expected = "120066240000000026040B52006840000000000000007300701321EDB6FC8E803EE8EDC2793F1EC917B2EE41D35255618DEB91D3F9B1FC89B75D4539810000101101"

	base := map[string]any{
		"Fee":                "0",
		"LedgerSequence":     uint32(67850752),
		"Sequence":           uint32(0),
		"SigningPubKey":      "",
		"TransactionType":    "UNLModify",
		"UNLModifyDisabling": uint8(1),
		"UNLModifyValidator": "EDB6FC8E803EE8EDC2793F1EC917B2EE41D35255618DEB91D3F9B1FC89B75D4539",
	}

	for _, account := range []string{"", "rrrrrrrrrrrrrrrrrrrrrhoLvTp"} {
		t.Run("Account "+account, func(t *testing.T) {
			tx := make(map[string]any, len(base)+1)
			maps.Copy(tx, base)
			tx["Account"] = account

			actual, err := Encode(tx)
			require.NoError(t, err)
			require.Equal(t, expected, actual)
		})
	}
}

func TestEncodeOtherPseudoTransactions(t *testing.T) {
	tests := []struct {
		name     string
		tx       map[string]any
		expected string
	}{
		{
			// Mainnet transaction CA4562711E4679FE9317DD767871E90A404C7A8B84FAFD35EC2CF0231F1F6DAF:
			// https://livenet.xrpl.org/transactions/CA4562711E4679FE9317DD767871E90A404C7A8B84FAFD35EC2CF0231F1F6DAF
			name: "EnableAmendment",
			tx: map[string]any{
				"Account":         "rrrrrrrrrrrrrrrrrrrrrhoLvTp",
				"Amendment":       "AE35ABDEFBDE520372B31C957020B34A7A4A9DC3115A69803A44016477C84D6E",
				"Fee":             "0",
				"LedgerSequence":  uint32(84206081),
				"Sequence":        uint32(0),
				"SigningPubKey":   "",
				"TransactionType": "EnableAmendment",
			},
			expected: "1200642400000000260504E2015013AE35ABDEFBDE520372B31C957020B34A7A4A9DC3115A69803A44016477C84D6E684000000000000000730081140000000000000000000000000000000000000000",
		},
		{
			// Mainnet transaction 1C15FEA3E1D50F96B6598607FC773FF1F6E0125F30160144BE0C5CBC52F5151B:
			// https://livenet.xrpl.org/transactions/1C15FEA3E1D50F96B6598607FC773FF1F6E0125F30160144BE0C5CBC52F5151B
			name: "SetFee",
			tx: map[string]any{
				"Account":           "rrrrrrrrrrrrrrrrrrrrrhoLvTp",
				"BaseFee":           "000000000000000A",
				"Fee":               "0",
				"ReferenceFeeUnits": uint32(10),
				"ReserveBase":       uint32(20000000),
				"ReserveIncrement":  uint32(5000000),
				"Sequence":          uint32(0),
				"SigningPubKey":     "",
				"TransactionType":   "SetFee",
			},
			expected: "1200652400000000201E0000000A201F01312D002020004C4B4035000000000000000A684000000000000000730081140000000000000000000000000000000000000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := Encode(tt.tx)
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}
