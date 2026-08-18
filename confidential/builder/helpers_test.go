package builder

import (
	"errors"
	"strconv"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	ledgerentries "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

const (
	// testIssuanceID is issued by an account other than testAccount, so holder-submitted
	// transactions (convert, send, convert back, merge inbox) accept it.
	testIssuanceID = "000004C463C52827307480341E3CB23A0710CC839EB58A0A"
	// testIssuerIssuanceID embeds testAccount as the issuer. ConfidentialMPTClawback is
	// issuer-submitted, so it requires an issuance ID that testAccount itself issued.
	testIssuerIssuanceID = "000004C4B5F762798A53D543A014CAF8B297CFF8F2F937E8"
	testAccount          = "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	testDestination      = "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP"
)

// xAddressOf returns the mainnet X-address encoding of a classic address. Tests use it to
// pin that the builder rejects X-addresses outright, because the xrplhash and proof code it
// feeds decodes classic addresses only.
func xAddressOf(t *testing.T, classic string) string {
	t.Helper()
	_, accountID, err := addresscodec.DecodeClassicAddressToAccountID(classic)
	require.NoError(t, err)
	x, err := addresscodec.EncodeXAddress(accountID, 0, false, false)
	require.NoError(t, err)
	return x
}

// TestFixtureIssuanceIDs pins the issuer relationship the other suites depend on. Without
// it, changing testAccount would surface as an opaque role error in unrelated tests.
func TestFixtureIssuanceIDs(t *testing.T) {
	require.True(t, transaction.IsMPTokenIssuer(testIssuerIssuanceID, types.Address(testAccount)),
		"testIssuerIssuanceID must embed testAccount as its issuer")
	require.False(t, transaction.IsMPTokenIssuer(testIssuanceID, types.Address(testAccount)),
		"testIssuanceID must be issued by an account other than testAccount")
}

// mockQuerier implements LedgerQuerier for testing.
type mockQuerier struct {
	accountSeq uint32
	accountErr error // when set, GetAccountInfo returns this error
	entries    map[string]ledgerentries.FlatLedgerObject
	entryErrs  map[string]error
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
	if err := m.entryErrs[req.Index]; err != nil {
		return nil, err
	}
	node, ok := m.entries[req.Index]
	if !ok {
		// Mirror rippled, which reports a missing ledger entry as exactly "entryNotFound".
		return nil, errors.New(ledgerEntryNotFound)
	}
	return &ledger.EntryResponse{Node: node}, nil
}

// buildIssuanceEntry builds a mock MPTokenIssuance flat entry that is fully enabled for
// confidential transfers. Tests that exercise a capability check mutate the returned entry.
func buildIssuanceEntry(issuerKey, auditorKey string) ledgerentries.FlatLedgerObject {
	entry := ledgerentries.FlatLedgerObject{
		"IssuerEncryptionKey":           issuerKey,
		"Flags":                         float64(ledgerentries.LsfMPTCanHoldConfidentialBalance | ledgerentries.LsfMPTCanTransfer),
		"ConfidentialOutstandingAmount": strconv.FormatUint(uint64(types.MaxMPTAmount), 10),
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
