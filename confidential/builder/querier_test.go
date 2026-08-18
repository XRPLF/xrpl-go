package builder

import (
	"errors"
	"fmt"
	"testing"

	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	ledgerentries "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/stretchr/testify/require"
)

type ledgerQuerierStub struct {
	accountInfo func(*account.InfoRequest) (*account.InfoResponse, error)
	ledgerEntry func(*ledger.EntryRequest) (*ledger.EntryResponse, error)
}

func (s ledgerQuerierStub) GetAccountInfo(req *account.InfoRequest) (*account.InfoResponse, error) {
	return s.accountInfo(req)
}

func (s ledgerQuerierStub) GetLedgerEntry(req *ledger.EntryRequest) (*ledger.EntryResponse, error) {
	return s.ledgerEntry(req)
}

// stubEntry returns a querier that answers every ledger_entry request with one response.
func stubEntry(resp *ledger.EntryResponse) ledgerQuerierStub {
	return ledgerQuerierStub{
		ledgerEntry: func(*ledger.EntryRequest) (*ledger.EntryResponse, error) {
			return resp, nil
		},
	}
}

func TestMPTokenIndexRejectsMalformedIssuanceID(t *testing.T) {
	_, err := mpTokenIndex("invalid", testAccount)
	require.ErrorIs(t, err, ErrInvalidIssuanceID)
	require.NotErrorIs(t, err, ErrLedgerQuery)
}

func TestBeginBuildRejectsUnvalidatedState(t *testing.T) {
	tests := []struct {
		name     string
		response *account.InfoResponse
	}{
		{name: "nil response"},
		{name: "unvalidated response", response: &account.InfoResponse{LedgerIndex: mockLedgerIndex}},
		{name: "missing ledger index", response: &account.InfoResponse{Validated: true}},
		{name: "missing sequence", response: &account.InfoResponse{Validated: true, LedgerIndex: mockLedgerIndex}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q := ledgerQuerierStub{
				accountInfo: func(*account.InfoRequest) (*account.InfoResponse, error) {
					return test.response, nil
				},
			}

			_, _, err := beginBuild(q, testAccount)
			require.ErrorIs(t, err, ErrInvalidLedgerState)
		})
	}
}

// TestBeginBuildReadsSequenceFromOpenLedger pins that the sequence and the snapshot come
// from different ledgers. The sequence must count transactions the account already submitted
// but that no validated ledger holds yet, or a second build inside one ledger interval
// reuses a sequence and fails tefPAST_SEQ, while the state the proofs consume must stay on
// one validated ledger.
func TestBeginBuildReadsSequenceFromOpenLedger(t *testing.T) {
	const validatedSequence, openSequence uint32 = 7, 9
	q := ledgerQuerierStub{
		accountInfo: func(req *account.InfoRequest) (*account.InfoResponse, error) {
			resp := &account.InfoResponse{Validated: true, LedgerIndex: mockLedgerIndex}
			resp.AccountData.Sequence = validatedSequence
			if req.LedgerIndex == common.LedgerSpecifier(common.Current) {
				resp.AccountData.Sequence = openSequence
				resp.LedgerIndex = mockLedgerIndex + 1
			}
			return resp, nil
		},
	}

	sequence, snapshot, err := beginBuild(q, testAccount)
	require.NoError(t, err)
	require.Equal(t, openSequence, sequence)
	require.Equal(t, mockLedgerIndex, snapshot.index)
}

// TestBeginBuildPreservesOpenLedgerQueryError pins that a failure on the second account_info
// stays a transport error rather than surfacing as missing ledger state.
func TestBeginBuildPreservesOpenLedgerQueryError(t *testing.T) {
	cause := errors.New("account_info unavailable")
	q := ledgerQuerierStub{
		accountInfo: func(req *account.InfoRequest) (*account.InfoResponse, error) {
			if req.LedgerIndex == common.LedgerSpecifier(common.Current) {
				return nil, cause
			}
			resp := &account.InfoResponse{Validated: true, LedgerIndex: mockLedgerIndex}
			resp.AccountData.Sequence = 1
			return resp, nil
		},
	}

	_, _, err := beginBuild(q, testAccount)
	require.ErrorIs(t, err, ErrLedgerQuery)
	require.ErrorIs(t, err, cause)
}

func TestLedgerSnapshotPinsLaterReadsByHash(t *testing.T) {
	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)
	holderIndex, err := mpTokenIndex(testIssuanceID, testAccount)
	require.NoError(t, err)
	q := &mockQuerier{
		accountSeq: 7,
		entries: map[string]ledgerentries.FlatLedgerObject{
			issuanceIndex: buildIssuanceEntry("issuer-key", ""),
			holderIndex:   buildMPTokenEntry(mptokenFields{holderKey: "holder-key"}),
		},
	}

	sequence, snapshot, err := beginBuild(q, testAccount)
	require.NoError(t, err)
	require.Equal(t, uint32(7), sequence)
	require.Len(t, q.accountRequests, 2)
	require.Equal(t, common.LedgerSpecifier(common.Validated), q.accountRequests[0].LedgerIndex)
	require.Equal(t, common.LedgerSpecifier(common.Current), q.accountRequests[1].LedgerIndex)
	_, err = getProvableIssuance(snapshot, testIssuanceID)
	require.NoError(t, err)
	_, err = getMPTokenConvertState(snapshot, issuanceState{}, testIssuanceID, testAccount)
	require.NoError(t, err)

	require.Len(t, q.entryRequests, 2)
	require.Equal(t, mockLedgerIndex, q.entryRequests[0].LedgerIndex)
	require.Empty(t, q.entryRequests[0].LedgerHash)
	require.Nil(t, q.entryRequests[1].LedgerIndex)
	require.Equal(t, mockLedgerHash, q.entryRequests[1].LedgerHash)
	require.Equal(t, mockLedgerHash, snapshot.hash)
}

// TestReadEntryRejectsInvalidResponse covers every response readEntry refuses to build on:
// one drawn from a ledger other than the selected snapshot, and one carrying no usable node.
// Both are the same failure, because neither yields state a proof may be bound to.
func TestReadEntryRejectsInvalidResponse(t *testing.T) {
	const objectIndex = "ABCDEF"
	otherHash := common.LedgerHash("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	validNode := ledgerentries.FlatLedgerObject{"LedgerEntryType": "MPToken"}
	tests := []struct {
		name     string
		response *ledger.EntryResponse
	}{
		{name: "nil response"},
		{name: "unvalidated", response: &ledger.EntryResponse{Index: objectIndex, LedgerHash: mockLedgerHash, LedgerIndex: mockLedgerIndex, Node: validNode}},
		{name: "wrong ledger index", response: &ledger.EntryResponse{Index: objectIndex, LedgerHash: mockLedgerHash, LedgerIndex: mockLedgerIndex + 1, Node: validNode, Validated: true}},
		{name: "missing ledger hash", response: &ledger.EntryResponse{Index: objectIndex, LedgerIndex: mockLedgerIndex, Node: validNode, Validated: true}},
		{name: "malformed ledger hash", response: &ledger.EntryResponse{Index: objectIndex, LedgerHash: "not-a-hash", LedgerIndex: mockLedgerIndex, Node: validNode, Validated: true}},
		{name: "wrong ledger hash", response: &ledger.EntryResponse{Index: objectIndex, LedgerHash: otherHash, LedgerIndex: mockLedgerIndex, Node: validNode, Validated: true}},
		{name: "wrong object index", response: &ledger.EntryResponse{Index: "other", LedgerHash: mockLedgerHash, LedgerIndex: mockLedgerIndex, Node: validNode, Validated: true}},
		{name: "empty JSON node", response: &ledger.EntryResponse{Index: objectIndex, LedgerHash: mockLedgerHash, LedgerIndex: mockLedgerIndex, Validated: true}},
		{name: "binary only", response: &ledger.EntryResponse{Index: objectIndex, LedgerHash: mockLedgerHash, LedgerIndex: mockLedgerIndex, NodeBinary: "00", Validated: true}},
		{name: "JSON and binary", response: &ledger.EntryResponse{Index: objectIndex, LedgerHash: mockLedgerHash, LedgerIndex: mockLedgerIndex, Node: validNode, NodeBinary: "00", Validated: true}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot := &ledgerSnapshot{q: stubEntry(test.response), index: mockLedgerIndex, hash: mockLedgerHash}
			_, err := snapshot.readEntry(objectIndex)
			require.ErrorIs(t, err, ErrInvalidLedgerState)
		})
	}
}

// TestReadCurrentEntryRejectsInvalidResponse pins the open-ledger read to the same entry
// identity checks the validated read makes. The open ledger is deliberately unbound, so
// none of the snapshot's ledger checks apply here, but requireCurrentBalanceVersion
// compares the version this returns against the snapshot's. A response about a different
// entry would decide staleness from state belonging to someone else.
func TestReadCurrentEntryRejectsInvalidResponse(t *testing.T) {
	const objectIndex = "ABCDEF"
	validNode := ledgerentries.FlatLedgerObject{"LedgerEntryType": "MPToken"}
	tests := []struct {
		name     string
		response *ledger.EntryResponse
	}{
		{name: "nil response"},
		{name: "wrong object index", response: &ledger.EntryResponse{Index: "other", Node: validNode}},
		{name: "empty JSON node", response: &ledger.EntryResponse{Index: objectIndex}},
		{name: "binary only", response: &ledger.EntryResponse{Index: objectIndex, NodeBinary: "00"}},
		{name: "JSON and binary", response: &ledger.EntryResponse{Index: objectIndex, Node: validNode, NodeBinary: "00"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot := &ledgerSnapshot{q: stubEntry(test.response), index: mockLedgerIndex, hash: mockLedgerHash}
			_, err := snapshot.readCurrentEntry(objectIndex)
			require.ErrorIs(t, err, ErrInvalidLedgerState)
		})
	}
}

// TestReadCurrentEntryAcceptsUnvalidatedResponse pins that the open-ledger read does not
// inherit the snapshot's ledger checks. The open ledger is never validated and reports a
// different index, which is the whole reason this read exists.
func TestReadCurrentEntryAcceptsUnvalidatedResponse(t *testing.T) {
	const objectIndex = "ABCDEF"
	snapshot := &ledgerSnapshot{
		q: stubEntry(&ledger.EntryResponse{
			Index:       objectIndex,
			LedgerIndex: mockLedgerIndex + 7,
			Node:        ledgerentries.FlatLedgerObject{"LedgerEntryType": "MPToken"},
		}),
		index: mockLedgerIndex,
		hash:  mockLedgerHash,
	}

	resp, err := snapshot.readCurrentEntry(objectIndex)
	require.NoError(t, err)
	require.Equal(t, objectIndex, resp.Index)
}

// TestReadEntryRejectsUnselectedSnapshot pins that a snapshot with no ledger selected reads
// nothing, so a builder that skipped beginBuild cannot silently read the current ledger.
func TestReadEntryRejectsUnselectedSnapshot(t *testing.T) {
	_, err := (&ledgerSnapshot{}).readEntry("ABCDEF")
	require.ErrorIs(t, err, ErrInvalidLedgerState)
}

func TestGetMPTokenEntryRejectsWrongEntryType(t *testing.T) {
	index, err := mpTokenIndex(testIssuanceID, testAccount)
	require.NoError(t, err)

	for _, test := range wrongEntryTypeNodes("HolderEncryptionKey") {
		t.Run(test.Name, func(t *testing.T) {
			q := &mockQuerier{entries: map[string]ledgerentries.FlatLedgerObject{index: test.Node}}
			_, err := getMPTokenConvertState(snapshotFor(q), issuanceState{}, testIssuanceID, testAccount)
			require.ErrorIs(t, err, ErrInvalidLedgerState)
			require.ErrorContains(t, err, "LedgerEntryType")
			require.NotErrorIs(t, err, ErrMPTokenNotFound)
		})
	}
}

func TestClassifyEntryError(t *testing.T) {
	index, err := mpTokenIndex(testIssuanceID, testAccount)
	require.NoError(t, err)
	transportErr := errors.New("transport failed")
	entryNotFoundErr := errors.New(ledgerEntryNotFound)
	inexactErr := errors.New("transport mentioned entryNotFound")

	tests := []struct {
		name       string
		queryErr   error
		wantErr    error
		notErr     error
		wrappedErr error
	}{
		{name: "rippled entryNotFound", queryErr: entryNotFoundErr, wantErr: ErrMPTokenNotFound, notErr: ErrLedgerQuery, wrappedErr: entryNotFoundErr},
		{name: "wrapped rippled entryNotFound", queryErr: fmt.Errorf("request failed: %w", entryNotFoundErr), wantErr: ErrMPTokenNotFound, notErr: ErrLedgerQuery, wrappedErr: entryNotFoundErr},
		{name: "joined rippled entryNotFound", queryErr: errors.Join(errors.New("other error"), entryNotFoundErr), wantErr: ErrMPTokenNotFound, notErr: ErrLedgerQuery, wrappedErr: entryNotFoundErr},
		{name: "message is not exact", queryErr: inexactErr, wantErr: ErrLedgerQuery, notErr: ErrMPTokenNotFound, wrappedErr: inexactErr},
		{name: "transport error", queryErr: transportErr, wantErr: ErrLedgerQuery, notErr: ErrMPTokenNotFound, wrappedErr: transportErr},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q := &mockQuerier{entryErrs: map[string]error{index: test.queryErr}}
			_, err := getMPTokenState(snapshotFor(q), issuanceState{}, testIssuanceID, testAccount)
			require.ErrorIs(t, err, test.wantErr)
			require.NotErrorIs(t, err, test.notErr)
			require.ErrorIs(t, err, test.wrappedErr)
		})
	}
}
