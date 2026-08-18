package builder

import (
	"testing"

	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	ledgerentries "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestGetProvableIssuanceRejectsMalformedIssuanceID(t *testing.T) {
	q := &mockQuerier{}

	_, err := getProvableIssuance(snapshotFor(q), "invalid")
	require.ErrorIs(t, err, ErrInvalidIssuanceID)
	require.NotErrorIs(t, err, ErrLedgerQuery)
	require.Zero(t, q.queryCalls, "a malformed issuance ID must fail before any ledger read")
}

func TestGetProvableIssuanceRejectsWrongEntryType(t *testing.T) {
	index, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)

	for _, test := range wrongEntryTypeNodes("IssuerEncryptionKey") {
		t.Run(test.Name, func(t *testing.T) {
			q := &mockQuerier{entries: map[string]ledgerentries.FlatLedgerObject{index: test.Node}}
			_, err := getProvableIssuance(snapshotFor(q), testIssuanceID)
			require.ErrorIs(t, err, ErrInvalidLedgerState)
			require.ErrorContains(t, err, "LedgerEntryType")
		})
	}
}

// TestGetProvableIssuanceReadsCapabilities pins the fields a proving builder consumes from a
// fully enabled issuance.
func TestGetProvableIssuanceReadsCapabilities(t *testing.T) {
	index, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)
	q := &mockQuerier{entries: map[string]ledgerentries.FlatLedgerObject{
		index: buildIssuanceEntry("issuerKey", ""),
	}}

	issuance, err := getProvableIssuance(snapshotFor(q), testIssuanceID)
	require.NoError(t, err)
	require.Equal(t, "issuerKey", issuance.issuerKey)
	require.True(t, issuance.canTransfer())
	require.True(t, issuance.canClawback())
	require.False(t, issuance.isLocked())
	require.False(t, issuance.requiresAuth())
	require.False(t, issuance.hasAuditor())
	require.Zero(t, issuance.transferFee)
	require.Equal(t, uint64(types.MaxMPTAmount), issuance.confidentialOutstanding)
}

// TestGetProvableIssuanceRejectsUnusableIssuance covers every issuance a proving builder
// refuses to build against, each mirroring a transactor preclaim that would otherwise cost
// a fee.
func TestGetProvableIssuanceRejectsUnusableIssuance(t *testing.T) {
	index, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)

	tests := []struct {
		name    string
		mutate  func(ledgerentries.FlatLedgerObject)
		wantErr error
	}{
		{
			name:    "confidential balances not enabled",
			mutate:  func(e ledgerentries.FlatLedgerObject) { e["Flags"] = float64(ledgerentries.LsfMPTCanTransfer) },
			wantErr: ErrConfidentialDisabled,
		},
		{
			name:    "flags absent",
			mutate:  func(e ledgerentries.FlatLedgerObject) { delete(e, "Flags") },
			wantErr: ErrConfidentialDisabled,
		},
		{
			name:    "malformed flags",
			mutate:  func(e ledgerentries.FlatLedgerObject) { e["Flags"] = "160" },
			wantErr: ErrInvalidLedgerState,
		},
		{
			name:    "malformed outstanding amount",
			mutate:  func(e ledgerentries.FlatLedgerObject) { e["ConfidentialOutstandingAmount"] = "nope" },
			wantErr: ErrInvalidLedgerState,
		},
		{
			name:    "malformed issuer key",
			mutate:  func(e ledgerentries.FlatLedgerObject) { e["IssuerEncryptionKey"] = 1 },
			wantErr: ErrInvalidLedgerState,
		},
		{
			name:    "issuer key missing",
			mutate:  func(e ledgerentries.FlatLedgerObject) { delete(e, "IssuerEncryptionKey") },
			wantErr: ErrEncryptionKeyNotSet,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := buildIssuanceEntry("issuerKey", "")
			test.mutate(entry)
			q := &mockQuerier{entries: map[string]ledgerentries.FlatLedgerObject{index: entry}}

			_, err := getProvableIssuance(snapshotFor(q), testIssuanceID)
			require.ErrorIs(t, err, test.wantErr)
		})
	}
}

// TestGetMPTokenStateReadsBalanceVersion pins the wiring the parser test cannot see: XLS-96
// 7.5.5 starts the counter at 0 and 9.3 wraps it back to 0, so an MPToken that omits the
// field is at version 0 rather than malformed.
func TestGetMPTokenStateReadsBalanceVersion(t *testing.T) {
	index, err := mpTokenIndex(testIssuanceID, testAccount)
	require.NoError(t, err)

	tests := []struct {
		name    string
		version float64
		absent  bool
		want    uint32
	}{
		{name: "absent is version zero", absent: true},
		{name: "zero", version: 0},
		{name: "present", version: 2, want: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := buildMPTokenEntry(spendable(testHolderKey, testSpendingCt, test.version))
			if test.absent {
				delete(node, "ConfidentialBalanceVersion")
			}
			q := &mockQuerier{entries: map[string]ledgerentries.FlatLedgerObject{index: node}}

			state, err := getMPTokenState(snapshotFor(q), issuanceState{}, testIssuanceID, testAccount)
			require.NoError(t, err)
			require.Equal(t, test.want, state.balanceVersion)
		})
	}
}

// TestGetMPTokenHolderKeyIgnoresBalanceVersion pins that the key-only reader never parses the
// version, so a receiver whose version the builder does not consume cannot fail a send.
func TestGetMPTokenHolderKeyIgnoresBalanceVersion(t *testing.T) {
	index, err := mpTokenIndex(testIssuanceID, testAccount)
	require.NoError(t, err)
	node := buildMPTokenEntry(spendable("key", "balance-ciphertext", 0))
	node["ConfidentialBalanceVersion"] = "not a number"
	q := &mockQuerier{entries: map[string]ledgerentries.FlatLedgerObject{index: node}}
	snapshot := snapshotFor(q)

	holder, err := getMPTokenConvertState(snapshot, issuanceState{}, testIssuanceID, testAccount)
	require.NoError(t, err)
	require.Equal(t, "key", holder.holderKey)

	_, err = getMPTokenState(snapshot, issuanceState{}, testIssuanceID, testAccount)
	require.ErrorIs(t, err, ErrInvalidLedgerState)
}

// TestGetMPTokenStateRejectsSupersededBalanceVersion pins the open-ledger reread. The proof
// binds the balance version the transactor reads at apply time, so a version the validated
// snapshot cannot see yet is a proof the network will reject, and the build must not spend a
// fee to find that out.
func TestGetMPTokenStateRejectsSupersededBalanceVersion(t *testing.T) {
	index, err := mpTokenIndex(testIssuanceID, testAccount)
	require.NoError(t, err)
	validated := buildMPTokenEntry(spendable(testHolderKey, testSpendingCt, 3))

	tests := []struct {
		name    string
		current map[string]ledgerentries.FlatLedgerObject
		wantErr error
	}{
		{
			name:    "open ledger agrees",
			current: map[string]ledgerentries.FlatLedgerObject{index: validated},
		},
		{
			name:    "in-flight transaction bumped the version",
			current: map[string]ledgerentries.FlatLedgerObject{index: buildMPTokenEntry(spendable(testHolderKey, testSpendingCt, 4))},
			wantErr: ErrStaleBalanceVersion,
		},
		{
			name:    "in-flight transaction deleted the MPToken",
			current: map[string]ledgerentries.FlatLedgerObject{},
			wantErr: ErrStaleBalanceVersion,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q := &mockQuerier{
				entries:        map[string]ledgerentries.FlatLedgerObject{index: validated},
				currentEntries: test.current,
			}
			state, err := getMPTokenState(snapshotFor(q), issuanceState{}, testIssuanceID, testAccount)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, uint32(3), state.balanceVersion)
		})
	}
}

// TestGetMPTokenStateReadsOpenLedgerVersionOnce pins that the reread targets the open ledger
// rather than the snapshot, because a reread bound to the snapshot could never observe the
// in-flight transaction it exists to catch.
func TestGetMPTokenStateReadsOpenLedgerVersionOnce(t *testing.T) {
	index, err := mpTokenIndex(testIssuanceID, testAccount)
	require.NoError(t, err)
	entry := buildMPTokenEntry(spendable(testHolderKey, testSpendingCt, 1))
	q := &mockQuerier{entries: map[string]ledgerentries.FlatLedgerObject{index: entry}}

	_, err = getMPTokenState(snapshotFor(q), issuanceState{}, testIssuanceID, testAccount)
	require.NoError(t, err)

	require.Len(t, q.entryRequests, 2)
	require.Equal(t, mockLedgerIndex, q.entryRequests[0].LedgerIndex)
	require.Equal(t, common.Current, q.entryRequests[1].LedgerIndex)
	require.Empty(t, q.entryRequests[1].LedgerHash)
}

// TestRequireHolderUsable pins the client side of checkFrozen and requireAuth. Each case is
// a condition the transactor rejects after it has already charged a fee, except the two the
// builder deliberately declines to decide from the entries it reads.
func TestRequireHolderUsable(t *testing.T) {
	tests := []struct {
		name           string
		issuanceFlags  uint32
		issuanceDomain bool
		mptokenFlags   any
		wantErr        error
	}{
		{name: "unlocked and unauthenticated issuance"},
		{name: "issuance locked", issuanceFlags: ledgerentries.LsfMPTLocked, wantErr: ErrIssuanceLocked},
		{name: "holder locked", mptokenFlags: float64(ledgerentries.LsfMPTLocked), wantErr: ErrHolderLocked},
		{name: "auth required and holder unauthorized", issuanceFlags: ledgerentries.LsfMPTRequireAuth, wantErr: ErrHolderNotAuthorized},
		{
			name:          "auth required and holder authorized",
			issuanceFlags: ledgerentries.LsfMPTRequireAuth,
			mptokenFlags:  float64(ledgerentries.LsfMPTAuthorized),
		},
		{
			// A permissioned domain authorizes through credentials the builder never reads,
			// and that path accepts a holder without lsfMPTAuthorized, so rejecting here
			// would deny a build the network would have applied.
			name:           "auth required through a permissioned domain",
			issuanceFlags:  ledgerentries.LsfMPTRequireAuth,
			issuanceDomain: true,
		},
		{name: "malformed holder flags", mptokenFlags: "not-a-number", wantErr: ErrInvalidLedgerState},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := map[string]any{}
			if test.mptokenFlags != nil {
				node["Flags"] = test.mptokenFlags
			}
			err := requireHolderUsable(node, issuanceState{flags: test.issuanceFlags, hasDomain: test.issuanceDomain})
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

// TestGetIssuerCiphertextRequiresHolderKey pins the clawback's own field set.
// ConfidentialMPTClawback rejects a holder missing either the issuer mirror the equality
// proof consumes or the holder encryption key, and runs neither checkFrozen nor requireAuth,
// so a locked holder stays clawable.
func TestGetIssuerCiphertextRequiresHolderKey(t *testing.T) {
	index, err := mpTokenIndex(testIssuanceID, testAccount)
	require.NoError(t, err)

	tests := []struct {
		name    string
		fields  mptokenFields
		flags   any
		wantErr error
	}{
		{name: "holder key and issuer mirror present", fields: clawable(testIssuerMirrorCt)},
		{name: "missing holder key", fields: mptokenFields{issuerCt: testIssuerMirrorCt}, wantErr: ErrInvalidLedgerState},
		{name: "missing issuer mirror", fields: mptokenFields{holderKey: testHolderKey}, wantErr: ErrInvalidLedgerState},
		{name: "locked holder is still clawable", fields: clawable(testIssuerMirrorCt), flags: float64(ledgerentries.LsfMPTLocked)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := buildMPTokenEntry(test.fields)
			if test.flags != nil {
				entry["Flags"] = test.flags
			}
			q := &mockQuerier{entries: map[string]ledgerentries.FlatLedgerObject{index: entry}}

			ct, err := getIssuerCiphertext(snapshotFor(q), testIssuanceID, testAccount)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, testIssuerMirrorCt, ct)
		})
	}
}

// TestRequireCurrentBalanceVersionNeedsALeadingOpenLedger pins that a version difference is
// read as staleness only when the open ledger demonstrably leads the pinned validated one.
// rippled identifies an open ledger by ledger_current_index alone, and a pooled endpoint can
// answer the open read from a server that trails the snapshot. Rejecting on that would fail
// a build the network would have accepted, so the check stands down instead.
func TestRequireCurrentBalanceVersionNeedsALeadingOpenLedger(t *testing.T) {
	const (
		objectIndex             = "ABCDEF"
		validatedVersion uint32 = 3
	)
	// The open ledger disagrees with the validated one in every case, so only the leading
	// ledger may turn that disagreement into a rejection.
	node := buildMPTokenEntry(spendable(testHolderKey, testSpendingCt, 9))

	tests := []struct {
		name      string
		openIndex common.LedgerIndex
		wantErr   error
	}{
		{name: "open ledger leads", openIndex: mockOpenLedgerIndex, wantErr: ErrStaleBalanceVersion},
		{name: "open ledger trails", openIndex: mockLedgerIndex - 1},
		{name: "open ledger equals the snapshot", openIndex: mockLedgerIndex},
		{name: "open ledger index absent", openIndex: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot := &ledgerSnapshot{
				q: stubEntry(&ledger.EntryResponse{
					Index:              objectIndex,
					LedgerCurrentIndex: test.openIndex,
					Node:               node,
				}),
				index: mockLedgerIndex,
				hash:  mockLedgerHash,
			}

			err := requireCurrentBalanceVersion(snapshot, objectIndex, validatedVersion)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
