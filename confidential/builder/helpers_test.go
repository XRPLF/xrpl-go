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
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
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
	// zeroClassicAccount is ACCOUNT_ZERO, which decodes cleanly in either address form
	// but can never sign, so every builder account field rejects it.
	zeroClassicAccount = "rrrrrrrrrrrrrrrrrrrrrhoLvTp"
	// mockLedgerIndex and mockLedgerHash are the validated ledger the mock querier reports,
	// so tests can assert every read of one build is bound to the same ledger.
	mockLedgerIndex = common.LedgerIndex(123)
	mockLedgerHash  = common.LedgerHash("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	// mockOpenLedgerIndex is the open ledger the mock answers a current read from. It leads
	// the validated ledger, as a real open ledger always does.
	mockOpenLedgerIndex = mockLedgerIndex + 1
	// testIssuerMirrorCt and testInboxCt stand in for ciphertexts the builders require to be
	// present but never decrypt, so their contents are irrelevant.
	testIssuerMirrorCt  = "issuer-mirror-ciphertext"
	testAuditorMirrorCt = "auditor-mirror-ciphertext"
	testInboxCt         = "inbox-ciphertext"
	testSpendingCt      = "spending-ciphertext"
	// testHolderKey stands in where a builder only checks that a holder key is registered.
	testHolderKey = "holder-encryption-key"
	// confidentialIssuanceFlags is the fully enabled issuance: confidential balances,
	// transfers, and clawback. Tests that exercise one capability check clear the flag
	// they target.
	confidentialIssuanceFlags = ledgerentries.LsfMPTCanHoldConfidentialBalance |
		ledgerentries.LsfMPTCanTransfer | ledgerentries.LsfMPTCanClawback
)

// taggedXAddressOf returns the mainnet X-address encoding of a classic address with an
// embedded tag. Tests use it to pin where a tag has a companion field to land in and
// where it does not.
func taggedXAddressOf(t *testing.T, classic string, tag uint32) string {
	t.Helper()
	x, err := addresscodec.ClassicAddressToXAddress(classic, tag, true, false)
	require.NoError(t, err)
	return x
}

// xAddressOf returns the untagged mainnet X-address encoding of a classic address. Tests
// use it to pin that the builder treats the two spellings as the same account.
func xAddressOf(t *testing.T, classic string) string {
	t.Helper()
	x, err := addresscodec.ClassicAddressToXAddress(classic, 0, false, false)
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
	// currentEntries answers open-ledger reads when set, so a test can present state a
	// transaction still in flight has already changed. Nil means the open ledger and the
	// validated ledger agree, which is the ordinary case.
	currentEntries map[string]ledgerentries.FlatLedgerObject
	entryErrs      map[string]error
	// queryCalls counts ledger requests so tests can assert failures occur before unnecessary ledger access.
	queryCalls      int
	accountRequests []account.InfoRequest
	entryRequests   []ledger.EntryRequest
}

func (m *mockQuerier) GetAccountInfo(req *account.InfoRequest) (*account.InfoResponse, error) {
	m.queryCalls++
	m.accountRequests = append(m.accountRequests, *req)
	if m.accountErr != nil {
		return nil, m.accountErr
	}
	return &account.InfoResponse{
		AccountData: ledgerentries.AccountRoot{Sequence: m.accountSeq},
		LedgerIndex: mockLedgerIndex,
		Validated:   true,
	}, nil
}

func (m *mockQuerier) GetLedgerEntry(req *ledger.EntryRequest) (*ledger.EntryResponse, error) {
	m.queryCalls++
	m.entryRequests = append(m.entryRequests, *req)
	if err := m.entryErrs[req.Index]; err != nil {
		return nil, err
	}
	entries := m.entries
	if req.LedgerIndex == common.Current && m.currentEntries != nil {
		entries = m.currentEntries
	}
	node, ok := entries[req.Index]
	if !ok {
		// Mirror rippled, which reports a missing ledger entry as exactly "entryNotFound".
		return nil, errors.New(ledgerEntryNotFound)
	}
	// Mirror rippled, which identifies an open ledger by ledger_current_index alone and a
	// closed one by ledger_hash and ledger_index. The open ledger always leads the validated
	// one, which is what makes a version read from it comparable.
	if req.LedgerIndex == common.Current {
		return &ledger.EntryResponse{
			Index:              req.Index,
			LedgerCurrentIndex: mockOpenLedgerIndex,
			Node:               node,
		}, nil
	}
	return &ledger.EntryResponse{
		Index:       req.Index,
		LedgerHash:  mockLedgerHash,
		LedgerIndex: mockLedgerIndex,
		Node:        node,
		Validated:   true,
	}, nil
}

// snapshotFor returns an unbound snapshot on the mock querier's validated ledger, which is
// the state beginBuild hands a builder before its first read.
func snapshotFor(q LedgerQuerier) *ledgerSnapshot {
	return &ledgerSnapshot{q: q, index: mockLedgerIndex}
}

// wrongEntryTypeNodes returns the three ways a node's LedgerEntryType can fail to name the
// entry the reader asked for. Each node carries the companion field the reader needs, so the
// type check is what fails rather than a missing field.
func wrongEntryTypeNodes(companion string) []struct {
	Name string
	Node ledgerentries.FlatLedgerObject
} {
	return []struct {
		Name string
		Node ledgerentries.FlatLedgerObject
	}{
		{Name: "missing", Node: ledgerentries.FlatLedgerObject{companion: "key"}},
		{Name: "malformed", Node: ledgerentries.FlatLedgerObject{"LedgerEntryType": 1, companion: "key"}},
		{Name: "wrong", Node: ledgerentries.FlatLedgerObject{"LedgerEntryType": "Offer", companion: "key"}},
	}
}

// buildIssuanceEntry builds a mock MPTokenIssuance flat entry that is fully enabled for
// confidential transfers. Tests that exercise a capability check mutate the returned entry.
func buildIssuanceEntry(issuerKey, auditorKey string) ledgerentries.FlatLedgerObject {
	entry := ledgerentries.FlatLedgerObject{
		"LedgerEntryType":               string(ledgerentries.MPTokenIssuanceEntry),
		"Flags":                         float64(confidentialIssuanceFlags),
		"ConfidentialOutstandingAmount": strconv.FormatUint(uint64(types.MaxMPTAmount), 10),
	}
	// rippled omits an unset encryption key rather than writing an empty blob, so the
	// fixture omits it too and an empty argument means "no key registered".
	if issuerKey != "" {
		entry["IssuerEncryptionKey"] = issuerKey
	}
	if auditorKey != "" {
		entry["AuditorEncryptionKey"] = auditorKey
	}
	return entry
}

// mptokenFields describes a mock MPToken. The zero value is an MPToken that exists but was
// never initialized for confidential balances, which several failure tables rely on, so
// every field is written only when set.
type mptokenFields struct {
	holderKey      string
	balanceCt      string
	balanceVersion float64
	inboxCt        string
	issuerCt       string
	// publicAmount is MPTAmount, the public balance a convert debits. It is serialized as a
	// decimal string, as rippled does for a base-ten UInt64. Zero omits the field, which is
	// how an MPToken holding nothing appears.
	publicAmount uint64
}

// spendable describes an MPToken carrying every field a confidential spend requires.
func spendable(holderKey, balanceCt string, balanceVersion float64) mptokenFields {
	return mptokenFields{
		holderKey:      holderKey,
		balanceCt:      balanceCt,
		balanceVersion: balanceVersion,
		issuerCt:       testIssuerMirrorCt,
	}
}

// receivable describes an MPToken carrying every field a confidential send's destination
// requires. A receiver needs an inbox and the issuer mirror, but no spending balance.
func receivable(holderKey string) mptokenFields {
	return mptokenFields{
		holderKey: holderKey,
		inboxCt:   testInboxCt,
		issuerCt:  testIssuerMirrorCt,
	}
}

// withAuditorMirror adds the auditor mirror balance XLS-96 8.4 keeps on every holder's
// MPToken once the issuance registers an auditor key.
func withAuditorMirror(entry ledgerentries.FlatLedgerObject) ledgerentries.FlatLedgerObject {
	entry["AuditorEncryptedBalance"] = testAuditorMirrorCt
	return entry
}

// withIssuanceFlags narrows a fixture's capability flags so one gate can be tested alone.
func withIssuanceFlags(entry ledgerentries.FlatLedgerObject, flags uint32) ledgerentries.FlatLedgerObject {
	entry["Flags"] = float64(flags)
	return entry
}

// mergeable describes an MPToken initialized for an inbox merge: a registered key plus both
// confidential balances. The merge reads no value, so the ciphertexts are placeholders.
func mergeable(holderKey string) mptokenFields {
	return mptokenFields{holderKey: holderKey, balanceCt: testSpendingCt, inboxCt: testInboxCt}
}

// clawable describes a holder MPToken a clawback can target: the issuer mirror balance the
// equality proof consumes, plus the holder key the transactor requires alongside it.
func clawable(issuerCt string) mptokenFields {
	return mptokenFields{holderKey: testHolderKey, issuerCt: issuerCt}
}

// buildMPTokenEntry builds a mock MPToken flat entry for a holder.
func buildMPTokenEntry(f mptokenFields) ledgerentries.FlatLedgerObject {
	entry := ledgerentries.FlatLedgerObject{"LedgerEntryType": string(ledgerentries.MPTokenEntry)}
	if f.holderKey != "" {
		entry["HolderEncryptionKey"] = f.holderKey
	}
	if f.balanceCt != "" {
		entry["ConfidentialBalanceSpending"] = f.balanceCt
		entry["ConfidentialBalanceVersion"] = f.balanceVersion
	}
	if f.inboxCt != "" {
		entry["ConfidentialBalanceInbox"] = f.inboxCt
	}
	if f.issuerCt != "" {
		entry["IssuerEncryptedBalance"] = f.issuerCt
	}
	if f.publicAmount > 0 {
		entry["MPTAmount"] = strconv.FormatUint(f.publicAmount, 10)
	}
	return entry
}

// holdingPublicly describes an MPToken with a public balance and no confidential state,
// which is what a holder looks like before its first convert.
func holdingPublicly(publicAmount uint64) mptokenFields {
	return mptokenFields{publicAmount: publicAmount}
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
			mptokenIndex:  buildMPTokenEntry(spendable(ownerKP.PubKeyHex, balanceCt, balanceVersion)),
		},
	}
}
