package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestPermissionedDomain_EntryType(t *testing.T) {
	permissionedDomain := &PermissionedDomain{}
	require.Equal(t, permissionedDomain.EntryType(), PermissionedDomainEntry)
}

func TestPermissionedDomain(t *testing.T) {

	tests := []struct {
		name               string
		permissionedDomain *PermissionedDomain
		expected           string
	}{
		{
			name: "pass - valid PermissionedDomain",
			permissionedDomain: &PermissionedDomain{
				Index:           types.Hash256("3DFA1DDEA27AF7E466DE395CCB16158E07ECA6BC4EB5580F75EBD39DE833645F"),
				LedgerEntryType: PermissionedDomainEntry,
				Fee:             types.XRPCurrencyAmount(10),
				Flags:           0,
				Owner:           types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				OwnerNode:       "0000000000000000",
				Sequence:        390,
				AcceptedCredentials: types.AuthorizeCredentialList{
					{
						Credential: types.Credential{
							Issuer:         types.Address("rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1"),
							CredentialType: types.CredentialType("6D795F63726564656E7469616C"),
						},
					},
				},
				PreviousTxnID:     types.Hash256("E7E3F2BBAAF48CF893896E48DC4A02BDA0C747B198D5AE18BC3D7567EE64B904"),
				PreviousTxnLgrSeq: 8734523,
			},
			expected: `{
	"index": "3DFA1DDEA27AF7E466DE395CCB16158E07ECA6BC4EB5580F75EBD39DE833645F",
	"LedgerEntryType": "PermissionedDomain",
	"Fee": "10",
	"Flags": 0,
	"Owner": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"OwnerNode": "0000000000000000",
	"Sequence": 390,
	"AcceptedCredentials": [
		{
			"Credential": {
				"Issuer": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
				"CredentialType": "6D795F63726564656E7469616C"
			}
		}
	],
	"PreviousTxnID": "E7E3F2BBAAF48CF893896E48DC4A02BDA0C747B198D5AE18BC3D7567EE64B904",
	"PreviousTxnLgrSeq": 8734523
}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := testutil.SerializeAndDeserialize(t, test.permissionedDomain, test.expected); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestPermissionedDomain_Flatten(t *testing.T) {
	tests := []struct {
		name               string
		permissionedDomain *PermissionedDomain
		expected           string
	}{
		{
			name: "pass - valid PermissionedDomain",
			permissionedDomain: &PermissionedDomain{
				Index:           types.Hash256("3DFA1DDEA27AF7E466DE395CCB16158E07ECA6BC4EB5580F75EBD39DE833645F"),
				LedgerEntryType: PermissionedDomainEntry,
				Fee:             types.XRPCurrencyAmount(10),
				Flags:           0,
				Owner:           types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				OwnerNode:       "0000000000000000",
				Sequence:        390,
				AcceptedCredentials: types.AuthorizeCredentialList{
					{
						Credential: types.Credential{
							Issuer:         types.Address("rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1"),
							CredentialType: types.CredentialType("6D795F63726564656E7469616C"),
						},
					},
				},
				PreviousTxnID:     types.Hash256("E7E3F2BBAAF48CF893896E48DC4A02BDA0C747B198D5AE18BC3D7567EE64B904"),
				PreviousTxnLgrSeq: 8734523,
			},
			expected: `{
				"index": "3DFA1DDEA27AF7E466DE395CCB16158E07ECA6BC4EB5580F75EBD39DE833645F",
				"LedgerEntryType": "PermissionedDomain",
				"Fee": "10",
				"Flags": 0,
				"Owner": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"OwnerNode": "0000000000000000",
				"Sequence": 390,
				"AcceptedCredentials": [
					{
						"Credential": {
							"Issuer": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
							"CredentialType": "6D795F63726564656E7469616C"
						}
					}
				],
				"PreviousTxnID": "E7E3F2BBAAF48CF893896E48DC4A02BDA0C747B198D5AE18BC3D7567EE64B904",
				"PreviousTxnLgrSeq": 8734523
			}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := testutil.CompareFlattenAndExpected(test.permissionedDomain.Flatten(), []byte(test.expected)); err != nil {
				t.Error(err)
			}
		})
	}
}
