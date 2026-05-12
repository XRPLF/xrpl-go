package transaction

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNFTokenModify_TxType(t *testing.T) {
	tx := &NFTokenModify{}
	assert.Equal(t, NFTokenModifyTx, tx.TxType())
}

func TestNFTokenModify_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		nft      *NFTokenModify
		expected FlatTransaction
	}{
		{
			name: "pass - all fields",
			nft: &NFTokenModify{
				BaseTx: BaseTx{
					TransactionType: NFTokenMintTx,
					Account:         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
				Owner:     "rogue5HnPRSszD9CWGSUz8UGHMVwSSKF6",
				NFTokenID: "0008C350C182B4F213B82CCFA4C6F59AD76F0AFCFBDF04D5A048C0A300000007",
				URI:       "697066733A2F2F62616679626569636D6E73347A736F6C686C6976346C746D6E356B697062776373637134616C70736D6C6179696970666B73746B736D3472746B652F5665742E706E67",
			},
			expected: FlatTransaction{
				"TransactionType": "NFTokenModify",
				"Account":         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				"Owner":           "rogue5HnPRSszD9CWGSUz8UGHMVwSSKF6",
				"NFTokenID":       "0008C350C182B4F213B82CCFA4C6F59AD76F0AFCFBDF04D5A048C0A300000007",
				"URI":             "697066733A2F2F62616679626569636D6E73347A736F6C686C6976346C746D6E356B697062776373637134616C70736D6C6179696970666B73746B736D3472746B652F5665742E706E67",
			},
		},
		{
			name: "pass - minimum required fields",
			nft: &NFTokenModify{
				BaseTx: BaseTx{
					TransactionType: NFTokenMintTx,
					Account:         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
				NFTokenID: "0008C350C182B4F213B82CCFA4C6F59AD76F0AFCFBDF04D5A048C0A300000007",
			},
			expected: FlatTransaction{
				"TransactionType": "NFTokenModify",
				"Account":         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				"NFTokenID":       "0008C350C182B4F213B82CCFA4C6F59AD76F0AFCFBDF04D5A048C0A300000007",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.nft.Flatten())
		})
	}
}

func TestNFTokenModify_Validate(t *testing.T) {
	tests := []struct {
		name        string
		nft         *NFTokenModify
		expectValid bool
		expectError bool
		errMessage  error
	}{
		{
			name: "pass - valid transaction",
			nft: &NFTokenModify{
				BaseTx: BaseTx{
					Account:         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					TransactionType: NFTokenModifyTx,
				},
				Owner:     "rogue5HnPRSszD9CWGSUz8UGHMVwSSKF6",
				NFTokenID: "0008C350C182B4F213B82CCFA4C6F59AD76F0AFCFBDF04D5A048C0A300000007",
				URI:       "697066733A2F2F62616679626569636D6E73347A736F6C686C6976346C746D6E356B697062776373637134616C70736D6C6179696970666B73746B736D3472746B652F5665742E706E67",
			},
			expectValid: true,
			expectError: false,
			errMessage:  nil,
		},
		{
			name: "fail - invalid NFTokenID (short)",
			nft: &NFTokenModify{
				BaseTx: BaseTx{
					Account:         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					TransactionType: NFTokenModifyTx,
				},
				Owner:     "rogue5HnPRSszD9CWGSUz8UGHMVwSSKF6",
				NFTokenID: "deadbeef",
			},
			expectValid: false,
			expectError: true,
			errMessage:  ErrInvalidNFTokenID,
		},
		{
			name: "fail - invalid NFTokenID (non-hex)",
			nft: &NFTokenModify{
				BaseTx: BaseTx{
					Account:         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					TransactionType: NFTokenModifyTx,
				},
				Owner:     "rogue5HnPRSszD9CWGSUz8UGHMVwSSKF6",
				NFTokenID: "ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ",
			},
			expectValid: false,
			expectError: true,
			errMessage:  ErrInvalidNFTokenID,
		},
		{
			name: "fail - invalid base transaction",
			nft: &NFTokenModify{
				BaseTx: BaseTx{
					Account:         "invalid",
					TransactionType: NFTokenModifyTx,
				},
				Owner:     "rogue5HnPRSszD9CWGSUz8UGHMVwSSKF6",
				NFTokenID: "0008C350C182B4F213B82CCFA4C6F59AD76F0AFCFBDF04D5A048C0A300000007",
				URI:       "697066733A2F2F62616679626569636D6E73347A736F6C686C6976346C746D6E356B697062776373637134616C70736D6C6179696970666B73746B736D3472746B652F5665742E706E67",
			},
			expectValid: false,
			expectError: true,
			errMessage:  ErrInvalidAccount,
		},
		{
			name: "fail - account and owner are the same",
			nft: &NFTokenModify{
				BaseTx: BaseTx{
					Account:         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					TransactionType: NFTokenModifyTx,
				},
				Owner:     "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				NFTokenID: "0008C350C182B4F213B82CCFA4C6F59AD76F0AFCFBDF04D5A048C0A300000007",
			},
			expectValid: false,
			expectError: true,
			errMessage:  ErrOwnerAccountConflict,
		},
		{
			name: "fail - invalid owner address",
			nft: &NFTokenModify{
				BaseTx: BaseTx{
					Account:         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					TransactionType: NFTokenModifyTx,
				},
				Owner:     "invalidAddress",
				NFTokenID: "0008C350C182B4F213B82CCFA4C6F59AD76F0AFCFBDF04D5A048C0A300000007",
			},
			expectValid: false,
			expectError: true,
			errMessage:  ErrInvalidOwner,
		},
		{
			name: "fail - invalid URI",
			nft: &NFTokenModify{
				BaseTx: BaseTx{
					Account:         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					TransactionType: NFTokenModifyTx,
				},
				Owner:     "rogue5HnPRSszD9CWGSUz8UGHMVwSSKF6",
				NFTokenID: "0008C350C182B4F213B82CCFA4C6F59AD76F0AFCFBDF04D5A048C0A300000007",
				URI:       "invalidHexURI",
			},
			expectValid: false,
			expectError: true,
			errMessage:  ErrInvalidURI,
		},
		{
			name: "pass - no owner and no URI",
			nft: &NFTokenModify{
				BaseTx: BaseTx{
					Account:         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					TransactionType: NFTokenModifyTx,
				},
				NFTokenID: "0008C350C182B4F213B82CCFA4C6F59AD76F0AFCFBDF04D5A048C0A300000007",
			},
			expectValid: true,
			expectError: false,
			errMessage:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.nft.Validate()
			require.Equal(t, tt.expectValid, valid)
			require.Equal(t, tt.expectError, err != nil)
			require.Equal(t, tt.errMessage, err)
		})
	}
}
