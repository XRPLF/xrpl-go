package vault

import (
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/assert"
)

func TestVaultInfoRequest_VaultID(t *testing.T) {
	s := InfoRequest{
		VaultID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
	}

	j := `{
	"vault_id": "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestVaultInfoRequest_OwnerAndSeq(t *testing.T) {
	seq := uint32(1)
	s := InfoRequest{
		Owner: "rfmDuhDyLGgx94qiwf3YF8BUV5j6KSvE8",
		Seq:   &seq,
	}

	j := `{
	"owner": "rfmDuhDyLGgx94qiwf3YF8BUV5j6KSvE8",
	"seq": 1
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestInfoRequest_Validate(t *testing.T) {
	seq := uint32(1)

	tests := []struct {
		name     string
		request  InfoRequest
		expected error
	}{
		{
			name: "pass - valid VaultID",
			request: InfoRequest{
				VaultID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
			},
			expected: nil,
		},
		{
			name: "pass - valid Owner and Seq",
			request: InfoRequest{
				Owner: "rfmDuhDyLGgx94qiwf3YF8BUV5j6KSvE8",
				Seq:   &seq,
			},
			expected: nil,
		},
		{
			name:     "fail - no lookup params",
			request:  InfoRequest{},
			expected: ErrMissingLookupParam,
		},
		{
			name: "fail - Owner without Seq",
			request: InfoRequest{
				Owner: "rfmDuhDyLGgx94qiwf3YF8BUV5j6KSvE8",
			},
			expected: ErrOwnerRequiresSeq,
		},
		{
			name: "fail - Seq without Owner",
			request: InfoRequest{
				Seq: &seq,
			},
			expected: ErrSeqRequiresOwner,
		},
		{
			name: "fail - VaultID with Owner and Seq",
			request: InfoRequest{
				VaultID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Owner:   "rfmDuhDyLGgx94qiwf3YF8BUV5j6KSvE8",
				Seq:     &seq,
			},
			expected: ErrConflictingLookupParams,
		},
		{
			name: "fail - invalid VaultID format",
			request: InfoRequest{
				VaultID: "INVALIDID",
			},
			expected: ErrInvalidVaultID,
		},
		{
			name: "fail - invalid Owner address",
			request: InfoRequest{
				Owner: "not-a-valid-address",
				Seq:   &seq,
			},
			expected: ErrInvalidOwner,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.expected == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.expected)
			}
		})
	}
}

func TestVaultInfoResponse(t *testing.T) {
	withdrawalPolicy := uint8(0)
	flags := uint32(0)

	s := Response{
		Vault: Vault{
			Account: "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
			Asset: ledger.Asset{
				Currency: "USD",
				Issuer:   "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			},
			AssetsAvailable:   "500000",
			AssetsTotal:       "1000000",
			LedgerEntryType:   "Vault",
			Owner:             "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
			PreviousTxnID:     "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
			PreviousTxnLgrSeq: 28991004,
			Sequence:          1,
			Index:             "20B136D7BF6D2E3D610E28E3E6BE09F5C8F4F0241BBF6E2D072AE1BACB1388F5",
			Shares: Shares{
				Issuer:            "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
				LedgerEntryType:   "MPTokenIssuance",
				OutstandingAmount: "1000000",
				PreviousTxnID:     "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
				PreviousTxnLgrSeq: 28991004,
				Sequence:          1,
				Index:             "5A92F6ED33FDA68FB4B9FD140EA38C056CD2BA9673ECA5B4CEF40F2166BB6F0C",
				OwnerNode:         "0000000000000000",
			},
			ShareMPTID:       "00000000000000000000000000000000000000000000000000000000",
			WithdrawalPolicy: &withdrawalPolicy,
			OwnerNode:        "0000000000000000",
			Flags:            &flags,
		},
		LedgerHash: "4C99E5F63C0D0B1C2283B4F5DCE2239F80CE92E8B1A6AED1E110C198FC96E659",
		Validated:  true,
	}

	j := `{
	"vault": {
		"Account": "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
		"Asset": {
			"currency": "USD",
			"issuer": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
		},
		"AssetsAvailable": "500000",
		"AssetsTotal": "1000000",
		"LedgerEntryType": "Vault",
		"Owner": "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
		"PreviousTxnID": "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
		"PreviousTxnLgrSeq": 28991004,
		"Sequence": 1,
		"index": "20B136D7BF6D2E3D610E28E3E6BE09F5C8F4F0241BBF6E2D072AE1BACB1388F5",
		"shares": {
			"Issuer": "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
			"LedgerEntryType": "MPTokenIssuance",
			"OutstandingAmount": "1000000",
			"PreviousTxnID": "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
			"PreviousTxnLgrSeq": 28991004,
			"Sequence": 1,
			"index": "5A92F6ED33FDA68FB4B9FD140EA38C056CD2BA9673ECA5B4CEF40F2166BB6F0C",
			"OwnerNode": "0000000000000000"
		},
		"OwnerNode": "0000000000000000",
		"ShareMPTID": "00000000000000000000000000000000000000000000000000000000",
		"WithdrawalPolicy": 0,
		"Flags": 0
	},
	"ledger_hash": "4C99E5F63C0D0B1C2283B4F5DCE2239F80CE92E8B1A6AED1E110C198FC96E659",
	"validated": true
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}
