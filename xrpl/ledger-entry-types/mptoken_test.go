package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestMPToken_EntryType(t *testing.T) {
	mpToken := &MPToken{}
	require.Equal(t, mpToken.EntryType(), MPTokenEntry)
}

func TestMPToken_SetLsfMPTLocked(t *testing.T) {
	mpToken := &MPToken{}
	mpToken.SetLsfMPTLocked()
	require.Equal(t, mpToken.Flags, lsfMPTLocked)
}

func TestMPToken_SetLsfMPTAuthorized(t *testing.T) {
	mpToken := &MPToken{}
	mpToken.SetLsfMPTAuthorized()
	require.Equal(t, mpToken.Flags, lsfMPTAuthorized)
}

func TestMPTokenSerialization(t *testing.T) {
	tests := []struct {
		name     string
		mpToken  *MPToken
		expected string
	}{
		{
			name: "pass - valid MPToken with lsfMPTLocked",
			mpToken: &MPToken{
				Index:             types.Hash256("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
				LedgerEntryType:   MPTokenEntry,
				Flags:             lsfMPTLocked,
				Account:           types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				MPTokenIssuanceID: types.Hash192("rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1"),
				MPTAmount:         1000000,
				LockedAmount:      1,
				PreviousTxnID:     types.Hash256("8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB"),
				PreviousTxnLgrSeq: 234644,
				OwnerNode:         1,
			},
			expected: `{
	"index": "A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9",
	"LedgerEntryType": "MPToken",
	"Flags": 1,
	"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"MPTokenIssuanceID": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
	"MPTAmount": 1000000,
	"LockedAmount": 1,
	"PreviousTxnID": "8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB",
	"PreviousTxnLgrSeq": 234644,
	"OwnerNode": 1
}`,
		},
		{
			name: "pass - valid MPToken with lsfMPTAuthorized",
			mpToken: &MPToken{
				Index:             types.Hash256("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
				LedgerEntryType:   MPTokenEntry,
				Flags:             lsfMPTAuthorized,
				Account:           types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				MPTokenIssuanceID: types.Hash192("rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1"),
				MPTAmount:         1000000,
				LockedAmount:      1,
				PreviousTxnID:     types.Hash256("8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB"),
				PreviousTxnLgrSeq: 234644,
				OwnerNode:         1,
			},
			expected: `{
	"index": "A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9",
	"LedgerEntryType": "MPToken",
	"Flags": 2,
	"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"MPTokenIssuanceID": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
	"MPTAmount": 1000000,
	"LockedAmount": 1,
	"PreviousTxnID": "8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB",
	"PreviousTxnLgrSeq": 234644,
	"OwnerNode": 1
}`,
		},
		{
			name: "pass - valid MPToken with lsfMPTLocked and lsfMPTAuthorized",
			mpToken: &MPToken{
				Index:             types.Hash256("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
				LedgerEntryType:   MPTokenEntry,
				Flags:             lsfMPTLocked | lsfMPTAuthorized,
				Account:           types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				MPTokenIssuanceID: types.Hash192("rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1"),
				MPTAmount:         1000000,
				LockedAmount:      1,
				PreviousTxnID:     types.Hash256("8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB"),
				PreviousTxnLgrSeq: 234644,
				OwnerNode:         1,
			},
			expected: `{
	"index": "A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9",
	"LedgerEntryType": "MPToken",
	"Flags": 3,
	"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"MPTokenIssuanceID": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
	"MPTAmount": 1000000,
	"LockedAmount": 1,
	"PreviousTxnID": "8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB",
	"PreviousTxnLgrSeq": 234644,
	"OwnerNode": 1
}`,
		},
		{
			name: "pass - valid MPToken LockedAmount at 0",
			mpToken: &MPToken{
				Index:             types.Hash256("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
				LedgerEntryType:   MPTokenEntry,
				Flags:             lsfMPTLocked,
				Account:           types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				MPTokenIssuanceID: types.Hash192("rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1"),
				MPTAmount:         1000000,
				LockedAmount:      0,
				PreviousTxnID:     types.Hash256("8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB"),
				PreviousTxnLgrSeq: 234644,
				OwnerNode:         1,
			},
			expected: `{
	"index": "A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9",
	"LedgerEntryType": "MPToken",
	"Flags": 1,
	"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"MPTokenIssuanceID": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
	"MPTAmount": 1000000,
	"PreviousTxnID": "8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB",
	"PreviousTxnLgrSeq": 234644,
	"OwnerNode": 1
}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := testutil.SerializeAndDeserialize(t, test.mpToken, test.expected); err != nil {
				t.Error(err)
			}
		})
	}
}
