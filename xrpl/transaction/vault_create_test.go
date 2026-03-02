package transaction

import (
	"strings"
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestVaultCreate_TxType(t *testing.T) {
	tx := &VaultCreate{}
	assert.Equal(t, tx.TxType(), VaultCreateTx)
}

func TestVaultCreate_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *VaultCreate
		expected map[string]interface{}
	}{
		{
			name: "pass - empty",
			tx:   &VaultCreate{},
			expected: map[string]interface{}{
				"TransactionType": VaultCreateTx.String(),
			},
		},
		{
			name: "pass - with XRP asset",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:            "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					Fee:                1000000,
					Sequence:           1,
					LastLedgerSequence: 3000000,
				},
				Asset: ledger.Asset{Currency: "XRP"},
			},
			expected: map[string]interface{}{
				"TransactionType":    VaultCreateTx.String(),
				"Account":            "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
				"Fee":                "1000000",
				"Sequence":           uint32(1),
				"LastLedgerSequence": uint32(3000000),
				"Asset":              map[string]interface{}{"currency": "XRP"},
			},
		},
		{
			name: "pass - with optional fields",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account: "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
				},
				Asset: ledger.Asset{
					Currency: "USD",
					Issuer:   "rXJSJiZMxaLuH3kQBUV5DLipnYtrE6iVb",
				},
				AssetsMaximum:    func() *types.XRPLNumber { v := types.XRPLNumber("1000000"); return &v }(),
				WithdrawalPolicy: func() *uint8 { v := ledger.VaultStrategyFirstComeFirstServe; return &v }(),
			},
			expected: map[string]interface{}{
				"TransactionType":  VaultCreateTx.String(),
				"Account":          "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
				"Asset":            map[string]interface{}{"currency": "USD", "issuer": types.Address("rXJSJiZMxaLuH3kQBUV5DLipnYtrE6iVb")},
				"AssetsMaximum":    "1000000",
				"WithdrawalPolicy": ledger.VaultStrategyFirstComeFirstServe,
			},
		},
		{
			name: "pass - with Scale, Data, DomainID and MPTokenMetadata",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account: "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					Flags:   TfVaultPrivate,
				},
				Asset: ledger.Asset{
					Currency: "USD",
					Issuer:   "rXJSJiZMxaLuH3kQBUV5DLipnYtrE6iVb",
				},
				Scale:           func() *uint8 { v := uint8(6); return &v }(),
				Data:            func() *types.Data { v := types.Data("DEADBEEF"); return &v }(),
				DomainID:        func() *string { v := "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"; return &v }(),
				MPTokenMetadata: func() *string { v := "AABBCCDD"; return &v }(),
			},
			expected: map[string]interface{}{
				"TransactionType": VaultCreateTx.String(),
				"Account":         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
				"Flags":           TfVaultPrivate,
				"Asset":           map[string]interface{}{"currency": "USD", "issuer": types.Address("rXJSJiZMxaLuH3kQBUV5DLipnYtrE6iVb")},
				"Scale":           uint8(6),
				"Data":            "DEADBEEF",
				"DomainID":        "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				"MPTokenMetadata": "AABBCCDD",
			},
		},
		{
			name: "pass - with MPT asset",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account: "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
				},
				Asset: ledger.Asset{
					MPTIssuanceID: "983F536DBB46D5BBF43A0B5890576874EE1CF48CE31CA508A529EC17CD1A90EF",
				},
			},
			expected: map[string]interface{}{
				"TransactionType": VaultCreateTx.String(),
				"Account":         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
				"Asset":           map[string]interface{}{"mpt_issuance_id": "983F536DBB46D5BBF43A0B5890576874EE1CF48CE31CA508A529EC17CD1A90EF"},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.tx.Flatten(), testcase.expected)
		})
	}
}

func TestVaultCreate_Validate(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *VaultCreate
		expected error
	}{
		{
			name: "fail - base tx invalid",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					TransactionType: VaultCreateTx,
				},
			},
			expected: ErrInvalidAccount,
		},
		{
			name: "fail - Asset required",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
				},
			},
			expected: ErrVaultCreateAssetRequired,
		},
		{
			name: "pass - MPT asset",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
				},
				Asset: ledger.Asset{
					MPTIssuanceID: "983F536DBB46D5BBF43A0B5890576874EE1CF48CE31CA508A529EC17CD1A90EF",
				},
			},
			expected: nil,
		},
		{
			name: "pass - IOU asset",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
				},
				Asset: ledger.Asset{
					Currency: "USD",
					Issuer:   "rXJSJiZMxaLuH3kQBUV5DLipnYtrE6iVb",
				},
			},
			expected: nil,
		},
		{
			name: "fail - AssetsMaximum invalid",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
				},
				Asset:         ledger.Asset{Currency: "XRP"},
				AssetsMaximum: func() *types.XRPLNumber { v := types.XRPLNumber("notanumber"); return &v }(),
			},
			expected: ErrVaultCreateAssetsMaximumInvalid,
		},
		{
			name: "fail - Data not hex",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
				},
				Asset: ledger.Asset{Currency: "XRP"},
				Data:  func() *types.Data { v := types.Data("zznothex"); return &v }(),
			},
			expected: ErrVaultCreateDataInvalid,
		},
		{
			name: "fail - Data too large (> 256 bytes)",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
				},
				Asset: ledger.Asset{Currency: "XRP"},
				Data:  func() *types.Data { v := types.Data(strings.Repeat("AB", 257)); return &v }(),
			},
			expected: ErrVaultCreateDataInvalid,
		},
		{
			name: "pass - valid Data",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
				},
				Asset: ledger.Asset{Currency: "XRP"},
				Data:  func() *types.Data { v := types.Data("DEADBEEF"); return &v }(),
			},
			expected: nil,
		},
		{
			name: "fail - MPTokenMetadata not hex",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
				},
				Asset:           ledger.Asset{Currency: "XRP"},
				MPTokenMetadata: func() *string { v := "zznothex"; return &v }(),
			},
			expected: ErrVaultCreateMPTokenMetadataInvalid,
		},
		{
			name: "fail - MPTokenMetadata too large (> 1024 bytes)",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
				},
				Asset:           ledger.Asset{Currency: "XRP"},
				MPTokenMetadata: func() *string { v := strings.Repeat("AB", 1025); return &v }(),
			},
			expected: ErrVaultCreateMPTokenMetadataInvalid,
		},
		{
			name: "pass - valid MPTokenMetadata",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
				},
				Asset:           ledger.Asset{Currency: "XRP"},
				MPTokenMetadata: func() *string { v := "AABBCCDD"; return &v }(),
			},
			expected: nil,
		},
		{
			name: "fail - Scale requires IOU asset (MPT)",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
				},
				Asset: ledger.Asset{
					MPTIssuanceID: "983F536DBB46D5BBF43A0B5890576874EE1CF48CE31CA508A529EC17CD1A90EF",
				},
				Scale: func() *uint8 { v := uint8(5); return &v }(),
			},
			expected: ErrVaultCreateScaleRequiresIOU,
		},
		{
			name: "pass - IOU asset with Scale 0",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
				},
				Asset: ledger.Asset{
					Currency: "USD",
					Issuer:   "rXJSJiZMxaLuH3kQBUV5DLipnYtrE6iVb",
				},
				Scale: func() *uint8 { v := uint8(0); return &v }(),
			},
			expected: nil,
		},
		{
			name: "pass - IOU asset with Scale 18",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
				},
				Asset: ledger.Asset{
					Currency: "USD",
					Issuer:   "rXJSJiZMxaLuH3kQBUV5DLipnYtrE6iVb",
				},
				Scale: func() *uint8 { v := uint8(18); return &v }(),
			},
			expected: nil,
		},
		{
			name: "pass - IOU asset with Scale 10",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
				},
				Asset: ledger.Asset{
					Currency: "USD",
					Issuer:   "rXJSJiZMxaLuH3kQBUV5DLipnYtrE6iVb",
				},
				Scale: func() *uint8 { v := uint8(10); return &v }(),
			},
			expected: nil,
		},
		{
			name: "pass - IOU asset without Scale",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
				},
				Asset: ledger.Asset{
					Currency: "USD",
					Issuer:   "rXJSJiZMxaLuH3kQBUV5DLipnYtrE6iVb",
				},
			},
			expected: nil,
		},
		{
			name: "fail - Scale invalid (> 18)",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
				},
				Asset: ledger.Asset{
					Currency: "USD",
					Issuer:   "rXJSJiZMxaLuH3kQBUV5DLipnYtrE6iVb",
				},
				Scale: func() *uint8 { v := uint8(19); return &v }(),
			},
			expected: ErrVaultCreateScaleInvalid,
		},
		{
			name: "fail - Scale requires IOU asset",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
				},
				Asset: ledger.Asset{Currency: "XRP"},
				Scale: func() *uint8 { v := uint8(6); return &v }(),
			},
			expected: ErrVaultCreateScaleRequiresIOU,
		},
		{
			name: "fail - DomainID requires TfVaultPrivate flag",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
				},
				Asset:    ledger.Asset{Currency: "XRP"},
				DomainID: func() *string { v := "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"; return &v }(),
			},
			expected: ErrVaultCreateDomainIDRequiresPrivateFlag,
		},
		{
			name: "pass - XRP asset",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
				},
				Asset: ledger.Asset{Currency: "XRP"},
			},
			expected: nil,
		},
		{
			name: "pass - IOU asset with Scale",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
				},
				Asset: ledger.Asset{
					Currency: "USD",
					Issuer:   "rXJSJiZMxaLuH3kQBUV5DLipnYtrE6iVb",
				},
				Scale: func() *uint8 { v := uint8(6); return &v }(),
			},
			expected: nil,
		},
		{
			name: "pass - with TfVaultPrivate and DomainID",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
					Flags:           TfVaultPrivate,
				},
				Asset:    ledger.Asset{Currency: "XRP"},
				DomainID: func() *string { v := "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"; return &v }(),
			},
			expected: nil,
		},
		{
			name: "fail - DomainID invalid (too short)",
			tx: &VaultCreate{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultCreateTx,
					Flags:           TfVaultPrivate,
				},
				Asset:    ledger.Asset{Currency: "XRP"},
				DomainID: func() *string { v := "TOOSHORT"; return &v }(),
			},
			expected: ErrVaultCreateDomainIDInvalid,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			ok, err := testcase.tx.Validate()
			assert.Equal(t, ok, testcase.expected == nil)
			if testcase.expected != nil {
				assert.Contains(t, err.Error(), testcase.expected.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
