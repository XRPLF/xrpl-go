package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestAccountRoot(t *testing.T) {
	var s Object = &AccountRoot{
		Account:           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		AccountTxnID:      "0D5FB50FA65C9FE1538FD7E398FFFE9D1908DFA4576D8D7A020040686F93C77D",
		Balance:           types.XRPCurrencyAmount(148446663),
		Domain:            "6D64756F31332E636F6D",
		EmailHash:         "98B4375E1D753E5B91627516F6D70977",
		Flags:             8388608,
		LedgerEntryType:   AccountRootEntry,
		MessageKey:        "0000000000000000000000070000000300",
		OwnerCount:        3,
		PreviousTxnID:     "0D5FB50FA65C9FE1538FD7E398FFFE9D1908DFA4576D8D7A020040686F93C77D",
		PreviousTxnLgrSeq: 14091160,
		Sequence:          336,
		TransferRate:      1004999999,
	}
	j := `{
	"Flags": 8388608,
	"LedgerEntryType": "AccountRoot",
	"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"AccountTxnID": "0D5FB50FA65C9FE1538FD7E398FFFE9D1908DFA4576D8D7A020040686F93C77D",
	"Balance": "148446663",
	"Domain": "6D64756F31332E636F6D",
	"EmailHash": "98B4375E1D753E5B91627516F6D70977",
	"MessageKey": "0000000000000000000000070000000300",
	"OwnerCount": 3,
	"PreviousTxnID": "0D5FB50FA65C9FE1538FD7E398FFFE9D1908DFA4576D8D7A020040686F93C77D",
	"PreviousTxnLgrSeq": 14091160,
	"Sequence": 336,
	"TransferRate": 1004999999
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestAccountRoot_EntryType(t *testing.T) {
	ar := &AccountRoot{}
	require.Equal(t, AccountRootEntry, ar.EntryType())
}

func TestAccountRoot_SetLsfAllowTrustLineClawback(t *testing.T) {
	ar := &AccountRoot{}
	ar.SetLsfAllowTrustLineClawback()
	require.Equal(t, LsfAllowTrustLineClawback, ar.Flags)
}

func TestAccountRoot_SetLsfDefaultRipple(t *testing.T) {
	ar := &AccountRoot{}
	ar.SetLsfDefaultRipple()
	require.Equal(t, LsfDefaultRipple, ar.Flags)
}

func TestAccountRoot_SetLsfDepositAuth(t *testing.T) {
	ar := &AccountRoot{}
	ar.SetLsfDepositAuth()
	require.Equal(t, LsfDepositAuth, ar.Flags)
}

func TestAccountRoot_SetLsfDisableMaster(t *testing.T) {
	ar := &AccountRoot{}
	ar.SetLsfDisableMaster()
	require.Equal(t, LsfDisableMaster, ar.Flags)
}

func TestAccountRoot_SetLsfDisallowIncomingCheck(t *testing.T) {
	ar := &AccountRoot{}
	ar.SetLsfDisallowIncomingCheck()
	require.Equal(t, LsfDisallowIncomingCheck, ar.Flags)
}

func TestAccountRoot_SetLsfDisallowIncomingNFTokenOffer(t *testing.T) {
	ar := &AccountRoot{}
	ar.SetLsfDisallowIncomingNFTokenOffer()
	require.Equal(t, LsfDisallowIncomingNFTokenOffer, ar.Flags)
}

func TestAccountRoot_SetLsfDisallowIncomingPayChan(t *testing.T) {
	ar := &AccountRoot{}
	ar.SetLsfDisallowIncomingPayChan()
	require.Equal(t, LsfDisallowIncomingPayChan, ar.Flags)
}

func TestAccountRoot_SetLsfDisallowIncomingTrustline(t *testing.T) {
	ar := &AccountRoot{}
	ar.SetLsfDisallowIncomingTrustline()
	require.Equal(t, LsfDisallowIncomingTrustline, ar.Flags)
}

func TestAccountRoot_SetLsfDisallowXRP(t *testing.T) {
	ar := &AccountRoot{}
	ar.SetLsfDisallowXRP()
	require.Equal(t, LsfDisallowXRP, ar.Flags)
}

func TestAccountRoot_SetLsfGlobalFreeze(t *testing.T) {
	ar := &AccountRoot{}
	ar.SetLsfGlobalFreeze()
	require.Equal(t, LsfGlobalFreeze, ar.Flags)
}

func TestAccountRoot_SetLsfNoFreeze(t *testing.T) {
	ar := &AccountRoot{}
	ar.SetLsfNoFreeze()
	require.Equal(t, LsfNoFreeze, ar.Flags)
}

func TestAccountRoot_SetLsfPasswordSpent(t *testing.T) {
	ar := &AccountRoot{}
	ar.SetLsfPasswordSpent()
	require.Equal(t, LsfPasswordSpent, ar.Flags)
}

func TestAccountRoot_SetLsfRequireAuth(t *testing.T) {
	ar := &AccountRoot{}
	ar.SetLsfRequireAuth()
	require.Equal(t, LsfRequireAuth, ar.Flags)
}

func TestAccountRoot_SetLsfRequireDestTag(t *testing.T) {
	ar := &AccountRoot{}
	ar.SetLsfRequireDestTag()
	require.Equal(t, LsfRequireDestTag, ar.Flags)
}

func TestAccountRoot_SetLsfAllowTrustLineLocking(t *testing.T) {
	ar := &AccountRoot{}
	ar.SetLsfAllowTrustLineLocking()
	require.Equal(t, LsfAllowTrustLineLocking, ar.Flags)
}
