package transaction

import (
	"encoding/json"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountSetTfFlags(t *testing.T) {
	tests := []struct {
		name     string
		setter   func(*AccountSet)
		expected uint32
	}{
		{
			name: "pass - SetRequireDestTag",
			setter: func(s *AccountSet) {
				s.SetRequireDestTag()
			},
			expected: TfRequireDestTag,
		},
		{
			name: "pass - SetRequireAuth",
			setter: func(s *AccountSet) {
				s.SetRequireAuth()
			},
			expected: TfRequireAuth,
		},
		{
			name: "pass - SetDisallowXRP",
			setter: func(s *AccountSet) {
				s.SetDisallowXRP()
			},
			expected: TfDisallowXRP,
		},
		{
			name: "pass - SetOptionalDestTag",
			setter: func(s *AccountSet) {
				s.SetOptionalDestTag()
			},
			expected: TfOptionalDestTag,
		},
		{
			name: "pass - SetRequireDestTag and SetRequireAuth",
			setter: func(s *AccountSet) {
				s.SetRequireDestTag()
				s.SetRequireAuth()
			},
			expected: TfRequireDestTag | TfRequireAuth,
		},
		{
			name: "pass - SetDisallowXRP and SetOptionalDestTag",
			setter: func(s *AccountSet) {
				s.SetDisallowXRP()
				s.SetOptionalDestTag()
			},
			expected: TfDisallowXRP | TfOptionalDestTag,
		},
		{
			name: "pass - SetRequireDestTag, SetRequireAuth, and SetDisallowXRP",
			setter: func(s *AccountSet) {
				s.SetRequireDestTag()
				s.SetRequireAuth()
				s.SetDisallowXRP()
			},
			expected: TfRequireDestTag | TfRequireAuth | TfDisallowXRP,
		},
		{
			name: "pass - All flags",
			setter: func(s *AccountSet) {
				s.SetRequireDestTag()
				s.SetRequireAuth()
				s.SetDisallowXRP()
				s.SetOptionalDestTag()
				s.SetOptionalAuth()
				s.SetAllowXRP()
			},
			expected: TfRequireDestTag | TfRequireAuth | TfDisallowXRP | TfOptionalDestTag | TfOptionalAuth | TfAllowXRP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &AccountSet{}
			tt.setter(s)
			if s.Flags != tt.expected {
				t.Errorf("Expected Flags to be %d, got %d", tt.expected, s.Flags)
			}
		})
	}
}

func TestAccountSetAsfFlags(t *testing.T) {
	tests := []struct {
		name     string
		setter   func(*AccountSet)
		expected uint32
	}{
		{
			name: "pass - SetAsfRequireDest",
			setter: func(s *AccountSet) {
				s.SetAsfRequireDest()
			},
			expected: AsfRequireDest,
		},
		{
			name: "pass - SetAsfRequireAuth",
			setter: func(s *AccountSet) {
				s.SetAsfRequireAuth()
			},
			expected: AsfRequireAuth,
		},
		{
			name: "pass - SetAsfDisallowXRP",
			setter: func(s *AccountSet) {
				s.SetAsfDisallowXRP()
			},
			expected: AsfDisallowXRP,
		},
		{
			name: "pass - SetAsfDisableMaster",
			setter: func(s *AccountSet) {
				s.SetAsfDisableMaster()
			},
			expected: AsfDisableMaster,
		},
		{
			name: "pass - SetAsfAccountTxnID",
			setter: func(s *AccountSet) {
				s.SetAsfAccountTxnID()
			},
			expected: AsfAccountTxnID,
		},
		{
			name: "pass - SetAsfNoFreeze",
			setter: func(s *AccountSet) {
				s.SetAsfNoFreeze()
			},
			expected: AsfNoFreeze,
		},
		{
			name: "pass - SetAsfGlobalFreeze",
			setter: func(s *AccountSet) {
				s.SetAsfGlobalFreeze()
			},
			expected: AsfGlobalFreeze,
		},
		{
			name: "pass - SetAsfDefaultRipple",
			setter: func(s *AccountSet) {
				s.SetAsfDefaultRipple()
			},
			expected: AsfDefaultRipple,
		},
		{
			name: "pass - SetAsfDepositAuth",
			setter: func(s *AccountSet) {
				s.SetAsfDepositAuth()
			},
			expected: AsfDepositAuth,
		},
		{
			name: "pass - SetAsfAuthorizedNFTokenMinter",
			setter: func(s *AccountSet) {
				s.SetAsfAuthorizedNFTokenMinter()
			},
			expected: AsfAuthorizedNFTokenMinter,
		},
		{
			name: "pass - SetAsfDisallowIncomingNFTokenOffer",
			setter: func(s *AccountSet) {
				s.SetAsfDisallowIncomingNFTokenOffer()
			},
			expected: AsfDisallowIncomingNFTokenOffer,
		},
		{
			name: "pass - SetAsfDisallowIncomingCheck",
			setter: func(s *AccountSet) {
				s.SetAsfDisallowIncomingCheck()
			},
			expected: AsfDisallowIncomingCheck,
		},
		{
			name: "pass - SetAsfDisallowIncomingPayChan",
			setter: func(s *AccountSet) {
				s.SetAsfDisallowIncomingPayChan()
			},
			expected: AsfDisallowIncomingPayChan,
		},
		{
			name: "pass - SetAsfDisallowIncomingTrustLine",
			setter: func(s *AccountSet) {
				s.SetAsfDisallowIncomingTrustLine()
			},
			expected: AsfDisallowIncomingTrustLine,
		},
		{
			name: "pass - SetAsfAllowTrustLineClawback",
			setter: func(s *AccountSet) {
				s.SetAsfAllowTrustLineClawback()
			},
			expected: AsfAllowTrustLineClawback,
		},
		{
			name: "pass - SetAsfAllowTrustLineLocking",
			setter: func(s *AccountSet) {
				s.SetAsfAllowTrustLineLocking()
			},
			expected: AsfAllowTrustLineLocking,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &AccountSet{}
			tt.setter(s)
			if s.SetFlag != tt.expected {
				t.Errorf("Expected Flags to be %d, got %d", tt.expected, s.Flags)
			}
		})
	}
}

func TestAccountClearAsfFlags(t *testing.T) {
	tests := []struct {
		name     string
		setter   func(*AccountSet)
		expected uint32
	}{
		{
			name: "pass - ClearAsfRequireDest",
			setter: func(s *AccountSet) {
				s.ClearAsfRequireDest()
			},
			expected: AsfRequireDest,
		},
		{
			name: "pass - ClearAsfRequireAuth",
			setter: func(s *AccountSet) {
				s.ClearAsfRequireAuth()
			},
			expected: AsfRequireAuth,
		},
		{
			name: "pass - ClearAsfDisallowXRP",
			setter: func(s *AccountSet) {
				s.ClearAsfDisallowXRP()
			},
			expected: AsfDisallowXRP,
		},
		{
			name: "pass - ClearAsfDisableMaster",
			setter: func(s *AccountSet) {
				s.ClearAsfDisableMaster()
			},
			expected: AsfDisableMaster,
		},
		{
			name: "pass - ClearAsfAccountTxnID",
			setter: func(s *AccountSet) {
				s.ClearAsfAccountTxnID()
			},
			expected: AsfAccountTxnID,
		},
		{
			name: "pass - AsfNoFreeze",
			setter: func(s *AccountSet) {
				s.ClearAsfNoFreeze()
			},
			expected: AsfNoFreeze,
		},
		{
			name: "pass - AsfGlobalFreeze",
			setter: func(s *AccountSet) {
				s.ClearAsfGlobalFreeze()
			},
			expected: AsfGlobalFreeze,
		},
		{
			name: "pass - ClearAsfDefaultRipple",
			setter: func(s *AccountSet) {
				s.ClearAsfDefaultRipple()
			},
			expected: AsfDefaultRipple,
		},
		{
			name: "pass - ClearAsfDepositAuth",
			setter: func(s *AccountSet) {
				s.ClearAsfDepositAuth()
			},
			expected: AsfDepositAuth,
		},
		{
			name: "pass - ClearAsfAuthorizedNFTokenMinter",
			setter: func(s *AccountSet) {
				s.ClearAsfAuthorizedNFTokenMinter()
			},
			expected: AsfAuthorizedNFTokenMinter,
		},
		{
			name: "pass - ClearAsfDisallowIncomingNFTokenOffer",
			setter: func(s *AccountSet) {
				s.ClearAsfDisallowIncomingNFTokenOffer()
			},
			expected: AsfDisallowIncomingNFTokenOffer,
		},
		{
			name: "pass - ClearAsfDisallowIncomingCheck",
			setter: func(s *AccountSet) {
				s.ClearAsfDisallowIncomingCheck()
			},
			expected: AsfDisallowIncomingCheck,
		},
		{
			name: "pass - ClearAsfDisallowIncomingPayChan",
			setter: func(s *AccountSet) {
				s.ClearAsfDisallowIncomingPayChan()
			},
			expected: AsfDisallowIncomingPayChan,
		},
		{
			name: "pass - ClearAsfDisallowIncomingTrustLine",
			setter: func(s *AccountSet) {
				s.ClearAsfDisallowIncomingTrustLine()
			},
			expected: AsfDisallowIncomingTrustLine,
		},
		{
			name: "pass - ClearAsfAllowTrustLineClawback",
			setter: func(s *AccountSet) {
				s.ClearAsfAllowTrustLineClawback()
			},
			expected: AsfAllowTrustLineClawback,
		},
		{
			name: "pass - ClearAsfAllowTrustLineLocking",
			setter: func(s *AccountSet) {
				s.ClearAsfAllowTrustLineLocking()
			},
			expected: AsfAllowTrustLineLocking,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &AccountSet{}
			tt.setter(s)
			if s.ClearFlag != tt.expected {
				t.Errorf("Expected Flags to be %d, got %d", tt.expected, s.Flags)
			}
		})
	}
}

func TestAccountSet_Validate(t *testing.T) {
	testCases := []struct {
		name       string
		accountSet *AccountSet
		valid      bool
	}{
		{
			name: "pass - Valid AccountSet",
			accountSet: &AccountSet{
				BaseTx: BaseTx{
					Account:         "r7dawf5hSG71faLnCrPiAQ5DkXfVxULPs",
					TransactionType: AccountSetTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
				ClearFlag:    1,
				SetFlag:      2,
				Domain:       types.Domain("A5B21758D2318FA2C"),
				EmailHash:    types.EmailHash("1234567890abcdef"),
				MessageKey:   types.MessageKey("messageKey"),
				TransferRate: types.TransferRate(1000000001),
				TickSize:     types.TickSize(5),
			},
			valid: true,
		},
		{
			name: "pass - Valid AccountSet without options, just the commons fields",
			accountSet: &AccountSet{
				BaseTx: BaseTx{
					Account:         "r7dawf5hSG71faLnCrPiAQ5DkXfVxULPs",
					TransactionType: AccountSetTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
			},
			valid: true,
		},
		{
			name: "fail - Invalid AccountSet with high SetFlag",
			accountSet: &AccountSet{
				BaseTx: BaseTx{
					Account:         "r7dawf5hSG71faLnCrPiAQ5DkXfVxULPs",
					TransactionType: AccountSetTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
				SetFlag: 18, // too high
			},
			valid: false,
		},
		{
			name: "fail - Invalid AccountSet with low TickSize",
			accountSet: &AccountSet{
				BaseTx: BaseTx{
					Account:         "r7dawf5hSG71faLnCrPiAQ5DkXfVxULPs",
					TransactionType: AccountSetTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
				TickSize: types.TickSize(2),
			},
			valid: false,
		},
		{
			name: "fail - Invalid AccountSet with high TickSize",
			accountSet: &AccountSet{
				BaseTx: BaseTx{
					Account:         "r7dawf5hSG71faLnCrPiAQ5DkXfVxULPs",
					TransactionType: AccountSetTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
				TickSize: types.TickSize(16),
			},
			valid: false,
		},
		{
			name: "pass - Valid AccountSet TickSize set to 0 to disable it",
			accountSet: &AccountSet{
				BaseTx: BaseTx{
					Account:         "r7dawf5hSG71faLnCrPiAQ5DkXfVxULPs",
					TransactionType: AccountSetTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
				TickSize: types.TickSize(0),
			},
			valid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valid, err := tc.accountSet.Validate()
			if valid != tc.valid {
				t.Errorf("Validation result for %s is incorrect. Expected: %v, Got: %v", tc.name, tc.valid, valid)
			}
			if err != nil && tc.valid {
				t.Errorf("Validation failed for %s: %s", tc.name, err)
			}
			if err == nil && !tc.valid {
				t.Errorf("Validation should have failed for %s", tc.name)
			}
		})
	}
}

func TestAccountSet_Flatten(t *testing.T) {
	tests := []struct {
		name       string
		accountSet *AccountSet
		expected   FlatTransaction
	}{
		{
			name: "pass - Flatten with all fields",
			accountSet: &AccountSet{
				BaseTx: BaseTx{
					Account:         "r7dawf5hSG71faLnCrPiAQ5DkXfVxULPs",
					TransactionType: AccountSetTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
				ClearFlag:     AsfRequireDest,
				Domain:        types.Domain("A5B21758D2318FA2C"),
				EmailHash:     types.EmailHash("1234567890abcdef"),
				MessageKey:    types.MessageKey("messagekey"),
				NFTokenMinter: types.NFTokenMinter("nftokenminter"),
				SetFlag:       AsfRequireAuth,
				TransferRate:  types.TransferRate(1000000001),
				TickSize:      types.TickSize(5),
				WalletLocator: types.WalletLocator("walletLocator"),
				WalletSize:    types.WalletSize(10),
			},
			expected: FlatTransaction{
				"Account":         "r7dawf5hSG71faLnCrPiAQ5DkXfVxULPs",
				"TransactionType": "AccountSet",
				"Fee":             "1",
				"Sequence":        uint32(1234),
				"SigningPubKey":   "ghijk",
				"TxnSignature":    "A1B2C3D4E5F6",
				"ClearFlag":       AsfRequireDest,
				"Domain":          "A5B21758D2318FA2C",
				"EmailHash":       "1234567890abcdef",
				"MessageKey":      "messagekey",
				"NFTokenMinter":   "nftokenminter",
				"SetFlag":         AsfRequireAuth,
				"TransferRate":    uint32(1000000001),
				"TickSize":        uint8(5),
				"WalletLocator":   "walletLocator",
				"WalletSize":      uint32(10),
			},
		},
		{
			name: "pass - Flatten with empty string or value set to 0 to remove/disable the fields",
			accountSet: &AccountSet{
				BaseTx: BaseTx{
					Account:         "r7dawf5hSG71faLnCrPiAQ5DkXfVxULPs",
					TransactionType: AccountSetTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
				Domain:        types.Domain(""),
				EmailHash:     types.EmailHash(""),
				TickSize:      types.TickSize(0),
				TransferRate:  types.TransferRate(0),
				NFTokenMinter: types.NFTokenMinter(""),
				WalletLocator: types.WalletLocator(""),
				WalletSize:    types.WalletSize(0),
			},
			expected: FlatTransaction{
				"Account":         "r7dawf5hSG71faLnCrPiAQ5DkXfVxULPs",
				"TransactionType": "AccountSet",
				"Fee":             "1",
				"Sequence":        uint32(1234),
				"SigningPubKey":   "ghijk",
				"TxnSignature":    "A1B2C3D4E5F6",
				"Domain":          "",
				"EmailHash":       "",
				"TickSize":        uint8(0),
				"TransferRate":    uint32(0),
				"NFTokenMinter":   "",
				"WalletLocator":   "",
				"WalletSize":      uint32(0),
			},
		},
		{
			name: "pass - Flatten with required strings only",
			accountSet: &AccountSet{
				BaseTx: BaseTx{
					Account:         "r7dawf5hSG71faLnCrPiAQ5DkXfVxULPs",
					TransactionType: AccountSetTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
			},
			expected: FlatTransaction{
				"Account":         "r7dawf5hSG71faLnCrPiAQ5DkXfVxULPs",
				"TransactionType": "AccountSet",
				"Fee":             "1",
				"Sequence":        uint32(1234),
				"SigningPubKey":   "ghijk",
				"TxnSignature":    "A1B2C3D4E5F6",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flattened := tc.accountSet.Flatten()
			require.Equal(t, tc.expected, flattened)
		})
	}
}

func TestAccountSet_TxType(t *testing.T) {
	entry := &AccountSet{}
	assert.Equal(t, AccountSetTx, entry.TxType())
}

func TestAccountSet_Unmarshal(t *testing.T) {
	tests := []struct {
		name                 string
		jsonData             string
		expectUnmarshalError bool
	}{
		{
			name: "pass - full AccountSet",
			jsonData: `{
				"TransactionType": "AccountSet",
				"Account": "rEXAMPLE123456789ABCDEFGHJKLMNPQRSTUVWXYZ",
				"Fee": "10",
				"Sequence": 1,
				"Flags": 2147483648,
				"LastLedgerSequence": 12345678,
				"SetFlag": 5,
				"ClearFlag": 6,
				"Domain": "6578616D706C652E636F6D", 
				"EmailHash": "98B4375E1D753E5A3F075C48A3C9AE0A",
				"MessageKey": "020000000000000000000000000000000000000000000000000000000000000001",
				"TransferRate": 1005000000,
				"TickSize": 5,
				"NFTokenMinter": "rNFTMINTERADDRESS123456789ABCDEFGHJKLMNPQRSTUVWXYZ",
				"NetworkID": 1024,
				"Memos": [
					{
						"Memo": {
							"MemoType": "74657374",
							"MemoData": "48656C6C6F2C20584D52"
						}
					}
				],
				"Signers": [
					{
					"Signer": {
						"Account": "rSIGNER123456789ABCDEFGHJKLMNPQRSTUVWXYZ",
						"SigningPubKey": "ED5F93AB1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF12345678",
						"TxnSignature": "3045022100D7F67A81F343...B87D"
					}
					}
				],
				"SourceTag": 12345,
				"SigningPubKey": "ED5F93AB1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF12345678",
				"TxnSignature": "3045022100D7F67A81F343...B87D"
			}`,
			expectUnmarshalError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var accountSet AccountSet
			err := json.Unmarshal([]byte(tt.jsonData), &accountSet)
			if (err != nil) != tt.expectUnmarshalError {
				t.Errorf("Unmarshal() error = %v, expectUnmarshalError %v", err, tt.expectUnmarshalError)
				return
			}
		})
	}
}
