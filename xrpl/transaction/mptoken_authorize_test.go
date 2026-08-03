package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestMPTokenAuthorize_TxType(t *testing.T) {
	tx := &MPTokenAuthorize{}
	require.Equal(t, MPTokenAuthorizeTx, tx.TxType())
}

func TestMPTokenAuthorize_Flatten(t *testing.T) {
	holder := types.Address("rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh")
	tests := []struct {
		name     string
		tx       *MPTokenAuthorize
		expected string
	}{
		{
			name: "pass - BaseTx only MPTokenAuthorize",
			tx: &MPTokenAuthorize{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenAuthorizeTx,
				},
				MPTokenIssuanceID: "1234",
			},
			expected: `{
				"Account": "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
				"TransactionType": "MPTokenAuthorize",
				"MPTokenIssuanceID": "1234"
			}`,
		},
		{
			name: "pass - MPTokenAuthorize with Holder",
			tx: &MPTokenAuthorize{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenAuthorizeTx,
				},
				MPTokenIssuanceID: "1234",
				Holder:            &holder,
			},
			expected: `{
				"Account": "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2", 
				"TransactionType": "MPTokenAuthorize",
				"MPTokenIssuanceID": "1234",
				"Holder": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := testutil.CompareFlattenAndExpected(tt.tx.Flatten(), []byte(tt.expected)); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestMPTokenAuthorize_Validate(t *testing.T) {
	holder := types.Address("rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh")
	accountHolder := types.Address("rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2")
	tests := []struct {
		name       string
		tx         *MPTokenAuthorize
		wantValid  bool
		wantErr    bool
		errMessage error
	}{
		{
			name: "fail - empty MPTokenIssuanceID",
			tx: &MPTokenAuthorize{
				BaseTx: BaseTx{
					Account:         accountHolder,
					TransactionType: MPTokenAuthorizeTx,
				},
				MPTokenIssuanceID: "",
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrInvalidMPTokenIssuanceIDAuthorize,
		},
		{
			name: "fail - non-hex MPTokenIssuanceID",
			tx: &MPTokenAuthorize{
				BaseTx: BaseTx{
					Account:         accountHolder,
					TransactionType: MPTokenAuthorizeTx,
				},
				MPTokenIssuanceID: "not-a-hex-value!",
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrInvalidMPTokenIssuanceIDAuthorize,
		},
		{
			name: "fail - holder account conflict",
			tx: &MPTokenAuthorize{
				BaseTx: BaseTx{
					Account:         accountHolder,
					TransactionType: MPTokenAuthorizeTx,
				},
				MPTokenIssuanceID: "1234",
				Holder:            types.Holder(accountHolder),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrHolderAccountConflict,
		},
		{
			name: "pass - valid without holder",
			tx: &MPTokenAuthorize{
				BaseTx: BaseTx{
					Account:         accountHolder,
					TransactionType: MPTokenAuthorizeTx,
				},
				MPTokenIssuanceID: "1234",
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "pass - valid with holder",
			tx: &MPTokenAuthorize{
				BaseTx: BaseTx{
					Account:         accountHolder,
					TransactionType: MPTokenAuthorizeTx,
				},
				MPTokenIssuanceID: "1234",
				Holder:            types.Holder(holder),
			},
			wantValid: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.tx.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.errMessage, err)
				require.False(t, valid)
			} else {
				require.NoError(t, err)
				require.True(t, valid)
			}
		})
	}
}

func TestMPTokenAuthorize_Flags(t *testing.T) {
	tx := &MPTokenAuthorize{}

	tx.SetMPTUnauthorizeFlag()
	require.Equal(t, uint32(1), tx.Flags)
}
