package builder

import (
	"errors"
	"strconv"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	ledgerentries "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/stretchr/testify/require"
)

// testHolderPrivKey is a fixed valid scalar so these tests run without CGo.
const testHolderPrivKey = "1111111111111111111111111111111111111111111111111111111111111111"

// balanceQuerier returns a querier holding the given MPToken and issuance; nil omits either.
func balanceQuerier(t *testing.T, mptoken, issuance ledgerentries.FlatLedgerObject) *mockQuerier {
	t.Helper()

	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)
	mptokenIndex, err := xrplhash.MPToken(testIssuanceID, testAccount)
	require.NoError(t, err)

	entries := map[string]ledgerentries.FlatLedgerObject{}
	if mptoken != nil {
		entries[mptokenIndex] = mptoken
	}
	if issuance != nil {
		entries[issuanceIndex] = issuance
	}
	return &mockQuerier{entries: entries}
}

// spendingBalanceParams returns valid params for testAccount.
func spendingBalanceParams() SpendingBalanceParams {
	return SpendingBalanceParams{
		Holder:        testAccount,
		IssuanceID:    testIssuanceID,
		HolderPrivKey: testHolderPrivKey,
		BalanceRange:  elgamal.AmountRange{Low: 0, High: 1000},
	}
}

// unspendableCiphertext is a well-formed ciphertext that fails to decrypt.
const unspendableCiphertext = "not-a-ciphertext"

// spendingMPToken returns an MPToken carrying a spending ciphertext.
func spendingMPToken(balanceCt string) ledgerentries.FlatLedgerObject {
	return buildMPTokenEntry(mptokenFields{holderKey: testHolderKey, balanceCt: balanceCt, inboxCt: testInboxCt})
}

func TestGetSpendingBalanceRejectsInvalidParams(t *testing.T) {
	tests := []struct {
		name    string
		params  func(SpendingBalanceParams) SpendingBalanceParams
		wantErr error
	}{
		{
			name:    "missing holder",
			params:  func(p SpendingBalanceParams) SpendingBalanceParams { p.Holder = ""; return p },
			wantErr: ErrMissingHolder,
		},
		{
			name:    "malformed holder",
			params:  func(p SpendingBalanceParams) SpendingBalanceParams { p.Holder = "not-an-address"; return p },
			wantErr: ErrInvalidHolder,
		},
		{
			name:    "zero holder",
			params:  func(p SpendingBalanceParams) SpendingBalanceParams { p.Holder = zeroClassicAccount; return p },
			wantErr: ErrInvalidHolder,
		},
		{
			name:    "missing issuance ID",
			params:  func(p SpendingBalanceParams) SpendingBalanceParams { p.IssuanceID = ""; return p },
			wantErr: ErrMissingIssuanceID,
		},
		{
			name:    "malformed issuance ID",
			params:  func(p SpendingBalanceParams) SpendingBalanceParams { p.IssuanceID = "not-an-issuance"; return p },
			wantErr: ErrInvalidIssuanceID,
		},
		{
			name:    "holder is the issuer",
			params:  func(p SpendingBalanceParams) SpendingBalanceParams { p.IssuanceID = testIssuerIssuanceID; return p },
			wantErr: ErrIssuerNotAllowed,
		},
		{
			name:    "missing private key",
			params:  func(p SpendingBalanceParams) SpendingBalanceParams { p.HolderPrivKey = ""; return p },
			wantErr: ErrMissingHolderKey,
		},
		{
			name:    "malformed private key",
			params:  func(p SpendingBalanceParams) SpendingBalanceParams { p.HolderPrivKey = "abcd"; return p },
			wantErr: ErrInvalidPrivKey,
		},
		{
			name: "inverted range",
			params: func(p SpendingBalanceParams) SpendingBalanceParams {
				p.BalanceRange = elgamal.AmountRange{Low: 10, High: 1}
				return p
			},
			wantErr: elgamal.ErrInvalidAmountRange,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q := balanceQuerier(t, spendingMPToken(unspendableCiphertext), buildIssuanceEntry(testHolderKey, ""))

			_, err := GetSpendingBalance(q, test.params(spendingBalanceParams()))
			require.ErrorIs(t, err, test.wantErr)
			require.Zero(t, q.queryCalls, "params must be rejected before any ledger query")
		})
	}
}

func TestGetSpendingBalanceReturnsZeroWithoutSpendingCiphertext(t *testing.T) {
	tests := []struct {
		name    string
		mptoken ledgerentries.FlatLedgerObject
	}{
		{name: "no confidential state", mptoken: buildMPTokenEntry(mptokenFields{})},
		{name: "registered but never converted", mptoken: buildMPTokenEntry(mptokenFields{holderKey: testHolderKey})},
		{name: "unmerged inbox only", mptoken: buildMPTokenEntry(mptokenFields{holderKey: testHolderKey, inboxCt: testInboxCt})},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q := balanceQuerier(t, test.mptoken, buildIssuanceEntry(testHolderKey, ""))

			balance, err := GetSpendingBalance(q, spendingBalanceParams())
			require.NoError(t, err)
			require.Zero(t, balance)
			require.Len(t, q.entryRequests, 1, "an absent ciphertext answers from the MPToken alone")
			require.Empty(t, q.accountRequests, "reading a balance needs no account sequence")
		})
	}
}

func TestGetSpendingBalanceReportsMissingMPToken(t *testing.T) {
	q := balanceQuerier(t, nil, buildIssuanceEntry(testHolderKey, ""))

	_, err := GetSpendingBalance(q, spendingBalanceParams())
	require.ErrorIs(t, err, ErrMPTokenNotFound)
	require.NotErrorIs(t, err, ErrLedgerQuery)
}

func TestGetSpendingBalanceReportsMissingIssuance(t *testing.T) {
	q := balanceQuerier(t, spendingMPToken(unspendableCiphertext), nil)

	_, err := GetSpendingBalance(q, spendingBalanceParams())
	require.ErrorIs(t, err, ErrIssuanceNotFound)
	require.NotErrorIs(t, err, ErrLedgerQuery)
}

func TestGetSpendingBalanceRejectsNonConfidentialIssuance(t *testing.T) {
	issuance := withIssuanceFlags(buildIssuanceEntry(testHolderKey, ""), ledgerentries.LsfMPTCanTransfer)
	q := balanceQuerier(t, spendingMPToken(unspendableCiphertext), issuance)

	_, err := GetSpendingBalance(q, spendingBalanceParams())
	require.ErrorIs(t, err, ErrConfidentialDisabled)
}

func TestGetSpendingBalanceRejectsMalformedLedgerState(t *testing.T) {
	validIssuance := buildIssuanceEntry(testHolderKey, "")
	validMPToken := spendingMPToken(unspendableCiphertext)

	tests := []struct {
		name     string
		mptoken  ledgerentries.FlatLedgerObject
		issuance ledgerentries.FlatLedgerObject
	}{
		{
			name:     "MPToken index holds another entry type",
			mptoken:  ledgerentries.FlatLedgerObject{"LedgerEntryType": "Offer", "ConfidentialBalanceSpending": unspendableCiphertext},
			issuance: validIssuance,
		},
		{
			name:     "MPToken without an entry type",
			mptoken:  ledgerentries.FlatLedgerObject{"ConfidentialBalanceSpending": unspendableCiphertext},
			issuance: validIssuance,
		},
		{
			name:     "spending ciphertext is not a string",
			mptoken:  ledgerentries.FlatLedgerObject{"LedgerEntryType": string(ledgerentries.MPTokenEntry), "ConfidentialBalanceSpending": float64(1)},
			issuance: validIssuance,
		},
		{
			name:     "spending ciphertext is empty",
			mptoken:  ledgerentries.FlatLedgerObject{"LedgerEntryType": string(ledgerentries.MPTokenEntry), "ConfidentialBalanceSpending": ""},
			issuance: validIssuance,
		},
		{
			name:     "issuance index holds another entry type",
			mptoken:  validMPToken,
			issuance: ledgerentries.FlatLedgerObject{"LedgerEntryType": string(ledgerentries.MPTokenEntry)},
		},
		{
			name:    "issuance flags are not a number",
			mptoken: validMPToken,
			issuance: ledgerentries.FlatLedgerObject{
				"LedgerEntryType": string(ledgerentries.MPTokenIssuanceEntry),
				"Flags":           "1",
			},
		},
		{
			name:    "confidential outstanding amount is not an integer string",
			mptoken: validMPToken,
			issuance: ledgerentries.FlatLedgerObject{
				"LedgerEntryType":               string(ledgerentries.MPTokenIssuanceEntry),
				"Flags":                         float64(confidentialIssuanceFlags),
				"ConfidentialOutstandingAmount": "not-a-number",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q := balanceQuerier(t, test.mptoken, test.issuance)

			_, err := GetSpendingBalance(q, spendingBalanceParams())
			require.ErrorIs(t, err, ErrInvalidLedgerState)
		})
	}
}

func TestGetSpendingBalanceRejectsUnvalidatedResponse(t *testing.T) {
	mptokenIndex, err := xrplhash.MPToken(testIssuanceID, testAccount)
	require.NoError(t, err)
	q := stubEntry(&ledger.EntryResponse{
		Index:       mptokenIndex,
		LedgerHash:  mockLedgerHash,
		LedgerIndex: mockLedgerIndex,
		Node:        spendingMPToken(unspendableCiphertext),
	})

	_, err = GetSpendingBalance(q, spendingBalanceParams())
	require.ErrorIs(t, err, ErrInvalidLedgerState)
}

func TestGetSpendingBalanceRejectsReadsFromTwoLedgers(t *testing.T) {
	const laterLedgerHash = common.LedgerHash("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")

	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)

	q := ledgerQuerierStub{
		ledgerEntry: func(req *ledger.EntryRequest) (*ledger.EntryResponse, error) {
			resp := &ledger.EntryResponse{
				Index:       req.Index,
				LedgerHash:  mockLedgerHash,
				LedgerIndex: mockLedgerIndex,
				Node:        spendingMPToken(unspendableCiphertext),
				Validated:   true,
			}
			if req.Index == issuanceIndex {
				resp.LedgerHash = laterLedgerHash
				resp.LedgerIndex = mockLedgerIndex + 1
				resp.Node = buildIssuanceEntry(testHolderKey, "")
			}
			return resp, nil
		},
	}

	_, err = GetSpendingBalance(q, spendingBalanceParams())
	require.ErrorIs(t, err, ErrInvalidLedgerState)
}

func TestGetSpendingBalancePreservesQueryErrors(t *testing.T) {
	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)
	mptokenIndex, err := xrplhash.MPToken(testIssuanceID, testAccount)
	require.NoError(t, err)

	tests := []struct {
		name  string
		index string
	}{
		{name: "MPToken read", index: mptokenIndex},
		{name: "issuance read", index: issuanceIndex},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cause := errors.New("ledger_entry unavailable")
			q := balanceQuerier(t, spendingMPToken(unspendableCiphertext), buildIssuanceEntry(testHolderKey, ""))
			q.entryErrs = map[string]error{test.index: cause}

			_, err := GetSpendingBalance(q, spendingBalanceParams())
			require.ErrorIs(t, err, ErrLedgerQuery)
			require.ErrorIs(t, err, cause)
		})
	}
}

func TestGetSpendingBalancePreservesDecryptionError(t *testing.T) {
	q := balanceQuerier(t, spendingMPToken(unspendableCiphertext), buildIssuanceEntry(testHolderKey, ""))

	_, err := GetSpendingBalance(q, spendingBalanceParams())
	require.ErrorIs(t, err, ErrCryptoFailed)
	require.ErrorIs(t, err, elgamal.ErrInvalidCiphertext)
	require.NotContains(t, err.Error(), testHolderPrivKey)
}

func TestGetSpendingBalanceReadsOneValidatedLedgerWithoutAccountQuery(t *testing.T) {
	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)
	mptokenIndex, err := xrplhash.MPToken(testIssuanceID, testAccount)
	require.NoError(t, err)
	q := balanceQuerier(t, spendingMPToken(unspendableCiphertext), buildIssuanceEntry(testHolderKey, ""))

	_, err = GetSpendingBalance(q, spendingBalanceParams())
	require.ErrorIs(t, err, ErrCryptoFailed)

	require.Empty(t, q.accountRequests)
	require.Equal(t, []ledger.EntryRequest{
		{Index: mptokenIndex, LedgerIndex: common.LedgerSpecifier(common.Validated)},
		{Index: issuanceIndex, LedgerHash: mockLedgerHash},
	}, q.entryRequests)
}

func TestGetSpendingBalanceRejectsRangeAboveOutstanding(t *testing.T) {
	const outstanding uint64 = 50

	issuance := buildIssuanceEntry(testHolderKey, "")
	issuance["ConfidentialOutstandingAmount"] = strconv.FormatUint(outstanding, 10)
	q := balanceQuerier(t, spendingMPToken(unspendableCiphertext), issuance)

	params := spendingBalanceParams()
	params.BalanceRange = elgamal.AmountRange{Low: outstanding + 1, High: outstanding + 100}

	_, err := GetSpendingBalance(q, params)
	require.ErrorIs(t, err, elgamal.ErrInvalidAmountRange)
	require.NotErrorIs(t, err, ErrCryptoFailed)
}

func TestBoundedBalanceRangeCapsHighAtOutstanding(t *testing.T) {
	const outstanding uint64 = 1000

	tests := []struct {
		name        string
		given       elgamal.AmountRange
		outstanding uint64
		want        elgamal.AmountRange
		wantErr     bool
	}{
		{
			name:        "range within the supply is left alone",
			given:       elgamal.AmountRange{Low: 10, High: 900},
			outstanding: outstanding,
			want:        elgamal.AmountRange{Low: 10, High: 900},
		},
		{
			name:        "high is capped at the supply",
			given:       elgamal.AmountRange{Low: 0, High: outstanding * 1000},
			outstanding: outstanding,
			want:        elgamal.AmountRange{Low: 0, High: outstanding},
		},
		{
			name:        "range ending at the supply is left alone",
			given:       elgamal.AmountRange{Low: 0, High: outstanding},
			outstanding: outstanding,
			want:        elgamal.AmountRange{Low: 0, High: outstanding},
		},
		{
			name:        "empty supply caps the search to zero",
			given:       elgamal.AmountRange{Low: 0, High: outstanding},
			outstanding: 0,
			want:        elgamal.AmountRange{Low: 0, High: 0},
		},
		{
			name:        "low above the supply is rejected",
			given:       elgamal.AmountRange{Low: outstanding + 1, High: outstanding + 2},
			outstanding: outstanding,
			wantErr:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := boundedBalanceRange(test.given, test.outstanding)
			if test.wantErr {
				require.ErrorIs(t, err, elgamal.ErrInvalidAmountRange)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}
}
