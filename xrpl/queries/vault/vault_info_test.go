package vault

import (
	"encoding/json"
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				require.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.expected)
			}
		})
	}
}

func TestVaultInfoResponse_LedgerMetadata(t *testing.T) {
	currentIndex := uint32(3280200)
	zeroIndex := uint32(0)
	validatedIndex := uint32(3280199)

	tests := []struct {
		name     string
		response Response
	}{
		{
			name: "open ledger",
			response: Response{
				LedgerCurrentIndex: &currentIndex,
			},
		},
		{
			name: "present zero current index",
			response: Response{
				LedgerCurrentIndex: &zeroIndex,
			},
		},
		{
			name:     "omitted ledger metadata",
			response: Response{},
		},
		{
			name: "validated ledger",
			response: Response{
				LedgerHash:  "4C99E5F63C0D0B1C2283B4F5DCE2239F80CE92E8B1A6AED1E110C198FC96E659",
				LedgerIndex: &validatedIndex,
				Validated:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.response)
			require.NoError(t, err)

			var encoded map[string]any
			require.NoError(t, json.Unmarshal(data, &encoded))
			if tt.response.LedgerCurrentIndex == nil {
				require.NotContains(t, encoded, "ledger_current_index")
			} else {
				require.EqualValues(t, *tt.response.LedgerCurrentIndex, encoded["ledger_current_index"])
			}

			var decoded Response
			require.NoError(t, json.Unmarshal(data, &decoded))
			require.Equal(t, tt.response, decoded)
		})
	}
}

func TestVaultInfoResponse(t *testing.T) {
	withdrawalPolicy := types.VaultWithdrawalPolicy(0)
	flags := uint32(0)
	ledgerIndex := uint32(1234)

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
			PreviousTxnID:     types.Hash256("C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7"),
			PreviousTxnLgrSeq: 28991004,
			Sequence:          1,
			Index:             "20B136D7BF6D2E3D610E28E3E6BE09F5C8F4F0241BBF6E2D072AE1BACB1388F5",
			Shares: Shares{
				Issuer:            "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
				LedgerEntryType:   "MPTokenIssuance",
				OutstandingAmount: "1000000",
				PreviousTxnID:     types.Hash256("C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7"),
				PreviousTxnLgrSeq: 28991004,
				Sequence:          1,
				Index:             "5A92F6ED33FDA68FB4B9FD140EA38C056CD2BA9673ECA5B4CEF40F2166BB6F0C",
				OwnerNode:         "0",
				Flags:             &flags,
				AssetScale:        6,
				MaximumAmount:     "9223372036854775807",
				TransferFee:       50000,
				MPTokenMetadata:   "534841524553",
				LockedAmount:      "9007199254740993",
				ReferenceHolding:  types.Hash256("13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8"),
			},
			ShareMPTID:       "00000000000000000000000000000000000000000000000000000000",
			WithdrawalPolicy: &withdrawalPolicy,
			OwnerNode:        "0",
			Flags:            &flags,
		},
		LedgerHash:  "4C99E5F63C0D0B1C2283B4F5DCE2239F80CE92E8B1A6AED1E110C198FC96E659",
		LedgerIndex: &ledgerIndex,
		Validated:   true,
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
			"OwnerNode": "0",
			"Flags": 0,
			"AssetScale": 6,
			"MaximumAmount": "9223372036854775807",
			"TransferFee": 50000,
			"MPTokenMetadata": "534841524553",
			"LockedAmount": "9007199254740993",
			"ReferenceHolding": "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8"
		},
		"OwnerNode": "0",
		"ShareMPTID": "00000000000000000000000000000000000000000000000000000000",
		"WithdrawalPolicy": 0,
		"Flags": 0
	},
	"ledger_hash": "4C99E5F63C0D0B1C2283B4F5DCE2239F80CE92E8B1A6AED1E110C198FC96E659",
	"ledger_index": 1234,
	"validated": true
}`
	// Exercise the response contract, not a specific on-chain combination of share fields.
	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestSharesOptionalFields(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{name: "absent", json: `{}`},
		{name: "explicit zero scale and fee", json: `{"AssetScale":0,"TransferFee":0}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var shares Shares
			require.NoError(t, json.Unmarshal([]byte(tt.json), &shares))
			require.Equal(t, Shares{}, shares)
		})
	}

	encoded, err := json.Marshal(Shares{})
	require.NoError(t, err)
	var fields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(encoded, &fields))
	for _, field := range []string{"AssetScale", "MaximumAmount", "TransferFee", "MPTokenMetadata", "LockedAmount", "ReferenceHolding"} {
		require.NotContains(t, fields, field)
	}
}

func TestSharesNumericLimits(t *testing.T) {
	// These are JSON type limits, not protocol validation. Zero is covered above.
	tests := []struct {
		name    string
		json    string
		want    Shares
		wantErr bool
	}{
		{
			name: "maximum unsigned values",
			json: `{"AssetScale":255,"TransferFee":65535}`,
			want: Shares{AssetScale: 255, TransferFee: 65535},
		},
		{
			name:    "negative scale",
			json:    `{"AssetScale":-1}`,
			wantErr: true,
		},
		{
			name:    "negative transfer fee",
			json:    `{"TransferFee":-1}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var shares Shares
			err := json.Unmarshal([]byte(tt.json), &shares)
			if tt.wantErr {
				var typeErr *json.UnmarshalTypeError
				require.ErrorAs(t, err, &typeErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, shares)
		})
	}
}
