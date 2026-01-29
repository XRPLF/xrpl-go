package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

// TestPermissionedDomainSet_TxType ensures the transaction type is set correctly.
func TestPermissionedDomainSet_TxType(t *testing.T) {
	tx := &PermissionedDomainSet{}
	require.Equal(t, PermissionedDomainSetTx, tx.TxType())
}

// TestPermissionedDomainSet_Flatten verifies the flattened output of the transaction.
func TestPermissionedDomainSet_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		tx       *PermissionedDomainSet
		expected string
	}{
		{
			name: "pass - without DomainID",
			tx: &PermissionedDomainSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PermissionedDomainSetTx,
				},
				AcceptedCredentials: []types.AuthorizeCredential{
					{
						// Using a sample issuer and credential type.
						Credential: types.Credential{
							Issuer:         "rPDXxSZcuVL3ZWoyU82bcde3zwvmShkRyF",
							CredentialType: types.CredentialType("1234"),
						},
					},
				},
			},
			expected: `{
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"AcceptedCredentials": [
					{
						"Credential": {
							"Issuer": "rPDXxSZcuVL3ZWoyU82bcde3zwvmShkRyF",
							"CredentialType": "1234"
						}
					}
				],
				"TransactionType": "PermissionedDomainSet"
			}`,
		},
		{
			name: "pass - with DomainID and multiple credentials",
			tx: &PermissionedDomainSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PermissionedDomainSetTx,
				},
				DomainID: "domain123",
				AcceptedCredentials: []types.AuthorizeCredential{
					{
						Credential: types.Credential{
							Issuer:         "rIssuer2",
							CredentialType: types.CredentialType("abcd"),
						},
					},
					{
						Credential: types.Credential{
							Issuer:         "rIssuer3",
							CredentialType: types.CredentialType("ef01"),
						},
					},
				},
			},
			expected: `{
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"TransactionType": "PermissionedDomainSet",
				"DomainID": "domain123",
				"AcceptedCredentials": [
					{
						"Credential": {
							"Issuer": "rIssuer2",
							"CredentialType": "abcd"
						}
					},
					{
						"Credential": {
							"Issuer": "rIssuer3",
							"CredentialType": "ef01"
						}
					}
				]
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testutil.CompareFlattenAndExpected(tt.tx.Flatten(), []byte(tt.expected))
			if err != nil {
				t.Error(err)
			}
		})
	}
}

// TestPermissionedDomainSet_Validate checks various cases for validating the transaction.
func TestPermissionedDomainSet_Validate(t *testing.T) {
	tests := []struct {
		name        string
		tx          *PermissionedDomainSet
		wantValid   bool
		wantErr     bool
		expectedErr error
	}{
		{
			name: "pass - valid transaction",
			tx: &PermissionedDomainSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PermissionedDomainSetTx,
				},
				AcceptedCredentials: []types.AuthorizeCredential{
					{
						Credential: types.Credential{
							Issuer:         "rPDXxSZcuVL3ZWoyU82bcde3zwvmShkRyF",
							CredentialType: types.CredentialType("1234"),
						},
					},
				},
			},
			wantValid:   true,
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name: "fail - missing BaseTx Account",
			tx: &PermissionedDomainSet{
				BaseTx: BaseTx{
					TransactionType: PermissionedDomainSetTx,
				},
				AcceptedCredentials: []types.AuthorizeCredential{
					{
						Credential: types.Credential{
							Issuer:         "rPDXxSZcuVL3ZWoyU82bcde3zwvmShkRyF",
							CredentialType: types.CredentialType("1234"),
						},
					},
				},
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidAccount,
		},
		{
			name: "fail - empty credentials list",
			tx: &PermissionedDomainSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PermissionedDomainSetTx,
				},
				AcceptedCredentials: types.AuthorizeCredentialList{},
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: types.ErrEmptyCredentials,
		},
		{
			name: "fail - duplicate credentials",
			tx: &PermissionedDomainSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PermissionedDomainSetTx,
				},
				AcceptedCredentials: types.AuthorizeCredentialList{
					types.AuthorizeCredential{
						Credential: types.Credential{
							Issuer:         "rPDXxSZcuVL3ZWoyU82bcde3zwvmShkRyF",
							CredentialType: types.CredentialType("1234"),
						},
					},
					types.AuthorizeCredential{
						Credential: types.Credential{
							Issuer:         "rPDXxSZcuVL3ZWoyU82bcde3zwvmShkRyF",
							CredentialType: types.CredentialType("1234"),
						},
					},
				},
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: types.ErrDuplicateCredentials,
		},
		{
			name: "fail - invalid credential (empty issuer)",
			tx: &PermissionedDomainSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PermissionedDomainSetTx,
				},
				AcceptedCredentials: []types.AuthorizeCredential{
					{
						Credential: types.Credential{
							Issuer:         "",
							CredentialType: types.CredentialType("1234"),
						},
					},
				},
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: types.ErrInvalidCredentialIssuer,
		},
		{
			name: "fail - invalid credential type",
			tx: &PermissionedDomainSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PermissionedDomainSetTx,
				},
				AcceptedCredentials: []types.AuthorizeCredential{
					{
						Credential: types.Credential{
							Issuer:         "rPDXxSZcuVL3ZWoyU82bcde3zwvmShkRyF",
							CredentialType: types.CredentialType("invalid"),
						},
					},
				},
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: types.ErrInvalidCredentialType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.tx.Validate()
			require.Equal(t, tt.wantValid, valid)
			if err != nil && tt.expectedErr != nil {
				// Compare error messages for equality.
				require.Equal(t, tt.expectedErr.Error(), err.Error())
			}
			require.Equal(t, tt.wantErr, err != nil)
		})
	}
}
