package builder

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	ledgerentries "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/stretchr/testify/require"
)

const (
	testIssuanceID  = "000004C463C52827307480341E3CB23A0710CC839EB58A0A"
	testAccount     = "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	testDestination = "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP"
)

// mockQuerier implements LedgerQuerier for testing.
type mockQuerier struct {
	accountSeq uint32
	accountErr error // when set, GetAccountInfo returns this error
	entries    map[string]ledgerentries.FlatLedgerObject
	// queryCalls counts ledger requests so tests can assert failures occur before unnecessary ledger access.
	queryCalls int
}

func (m *mockQuerier) GetAccountInfo(_ *account.InfoRequest) (*account.InfoResponse, error) {
	m.queryCalls++
	if m.accountErr != nil {
		return nil, m.accountErr
	}
	return &account.InfoResponse{
		AccountData: ledgerentries.AccountRoot{Sequence: m.accountSeq},
	}, nil
}

func (m *mockQuerier) GetLedgerEntry(req *ledger.EntryRequest) (*ledger.EntryResponse, error) {
	m.queryCalls++
	node, ok := m.entries[req.Index]
	if !ok {
		return nil, ErrMPTokenNotFound
	}
	return &ledger.EntryResponse{Node: node}, nil
}

// buildIssuanceEntry builds a mock MPTokenIssuance flat entry.
func buildIssuanceEntry(issuerKey, auditorKey string) ledgerentries.FlatLedgerObject {
	entry := ledgerentries.FlatLedgerObject{
		"IssuerEncryptionKey": issuerKey,
	}
	if auditorKey != "" {
		entry["AuditorEncryptionKey"] = auditorKey
	}
	return entry
}

// buildMPTokenEntry builds a mock MPToken flat entry for a holder.
func buildMPTokenEntry(holderKey, balanceCt string, balanceVersion float64, issuerCt string) ledgerentries.FlatLedgerObject {
	entry := ledgerentries.FlatLedgerObject{}
	if holderKey != "" {
		entry["HolderEncryptionKey"] = holderKey
	}
	if balanceCt != "" {
		entry["ConfidentialBalanceSpending"] = balanceCt
	}
	if balanceVersion != 0 {
		entry["ConfidentialBalanceVersion"] = balanceVersion
	}
	if issuerCt != "" {
		entry["IssuerEncryptedBalance"] = issuerCt
	}
	return entry
}

func newBalanceLedgerFixture(t *testing.T, sequence uint32, balanceVersion float64, balance uint64) (elgamal.Keypair, *mockQuerier) {
	t.Helper()

	ownerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	blindingFactor, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	balanceCt, err := elgamal.Encrypt(balance, ownerKP.PubKeyHex, blindingFactor)
	require.NoError(t, err)

	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)
	mptokenIndex, err := xrplhash.MPToken(testIssuanceID, testAccount)
	require.NoError(t, err)

	return ownerKP, &mockQuerier{
		accountSeq: sequence,
		entries: map[string]ledgerentries.FlatLedgerObject{
			issuanceIndex: buildIssuanceEntry(issuerKP.PubKeyHex, ""),
			mptokenIndex:  buildMPTokenEntry(ownerKP.PubKeyHex, balanceCt, balanceVersion, ""),
		},
	}
}
