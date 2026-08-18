package builder

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"testing"

	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	ledgerentries "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

// confidentialIssuanceFlags is the issuance flag set a confidential send requires.
const confidentialIssuanceFlags = ledgerentries.LsfMPTCanHoldConfidentialBalance | ledgerentries.LsfMPTCanTransfer

func TestQueryInputValidation(t *testing.T) {
	q := &mockQuerier{}

	_, err := getIssuance(q, "invalid")
	require.ErrorIs(t, err, ErrInvalidIssuanceID)
	require.NotErrorIs(t, err, ErrLedgerQuery)
	require.Zero(t, q.queryCalls)

	_, err = mpTokenIndex("invalid", testAccount)
	require.ErrorIs(t, err, ErrInvalidIssuanceID)
	require.NotErrorIs(t, err, ErrLedgerQuery)
}

func TestParseConfidentialBalanceVersion(t *testing.T) {
	index, err := mpTokenIndex(testIssuanceID, testAccount)
	require.NoError(t, err)
	tests := []struct {
		name        string
		present     bool
		value       any
		wantVersion uint32
		wantErr     bool
	}{
		{name: "absent"},
		{name: "RPC zero", present: true, value: json.Number("0")},
		{name: "RPC json number", present: true, value: json.Number("2"), wantVersion: 2},
		{name: "RPC maximum", present: true, value: json.Number("4294967295"), wantVersion: math.MaxUint32},
		{name: "WebSocket zero", present: true, value: float64(0)},
		{name: "WebSocket float", present: true, value: float64(3), wantVersion: 3},
		{name: "WebSocket maximum", present: true, value: float64(4294967295), wantVersion: math.MaxUint32},
		{name: "explicit null", present: true, wantErr: true},
		{name: "fractional JSON number", present: true, value: json.Number("2.5"), wantErr: true},
		{name: "negative JSON number", present: true, value: json.Number("-1"), wantErr: true},
		{name: "malformed JSON number", present: true, value: json.Number("nope"), wantErr: true},
		{name: "JSON number overflow", present: true, value: json.Number("4294967296"), wantErr: true},
		{name: "float overflow", present: true, value: float64(4294967296), wantErr: true},
		{name: "negative float", present: true, value: float64(-1), wantErr: true},
		{name: "fractional float", present: true, value: 2.5, wantErr: true},
		{name: "NaN float", present: true, value: math.NaN(), wantErr: true},
		{name: "positive infinity", present: true, value: math.Inf(1), wantErr: true},
		{name: "negative infinity", present: true, value: math.Inf(-1), wantErr: true},
		{name: "wrong type", present: true, value: "2", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := ledgerentries.FlatLedgerObject{}
			if test.present {
				node["ConfidentialBalanceVersion"] = test.value
			}
			q := &mockQuerier{entries: map[string]ledgerentries.FlatLedgerObject{index: node}}
			_, _, version, err := getMPTokenState(q, testIssuanceID, testAccount)
			if test.wantErr {
				require.ErrorIs(t, err, ErrLedgerQuery)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.wantVersion, version)
		})
	}
}

// TestGetMPTokenHolderKeyIgnoresBalanceVersion pins that the key-only reader never parses the
// version, so a receiver whose version the builder does not consume cannot fail a send.
func TestGetMPTokenHolderKeyIgnoresBalanceVersion(t *testing.T) {
	index, err := mpTokenIndex(testIssuanceID, testAccount)
	require.NoError(t, err)
	node := ledgerentries.FlatLedgerObject{
		"HolderEncryptionKey":        "key",
		"ConfidentialBalanceVersion": "not a number",
	}
	q := &mockQuerier{entries: map[string]ledgerentries.FlatLedgerObject{index: node}}

	holderKey, err := getMPTokenHolderKey(q, testIssuanceID, testAccount)
	require.NoError(t, err)
	require.Equal(t, "key", holderKey)

	_, _, _, err = getMPTokenState(q, testIssuanceID, testAccount)
	require.ErrorIs(t, err, ErrLedgerQuery)
}

func TestGetMPTokenStateErrorClassification(t *testing.T) {
	index, err := mpTokenIndex(testIssuanceID, testAccount)
	require.NoError(t, err)
	transportErr := errors.New("transport failed")
	entryNotFoundErr := errors.New(ledgerEntryNotFound)

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
		{name: "message is not exact", queryErr: errors.New("transport mentioned entryNotFound"), wantErr: ErrLedgerQuery, notErr: ErrMPTokenNotFound},
		{name: "transport error", queryErr: transportErr, wantErr: ErrLedgerQuery, notErr: ErrMPTokenNotFound, wrappedErr: transportErr},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q := &mockQuerier{entryErrs: map[string]error{index: test.queryErr}}
			_, _, _, err := getMPTokenState(q, testIssuanceID, testAccount)
			require.ErrorIs(t, err, test.wantErr)
			require.NotErrorIs(t, err, test.notErr)
			if test.wrappedErr != nil {
				require.ErrorIs(t, err, test.wrappedErr)
			}
		})
	}
}

func TestGetIssuanceCapabilities(t *testing.T) {
	index, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)

	tests := []struct {
		name    string
		mutate  func(ledgerentries.FlatLedgerObject)
		wantErr error
	}{
		{name: "fully enabled"},
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
			wantErr: ErrLedgerQuery,
		},
		{
			name:    "malformed outstanding amount",
			mutate:  func(e ledgerentries.FlatLedgerObject) { e["ConfidentialOutstandingAmount"] = "nope" },
			wantErr: ErrLedgerQuery,
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
			if test.mutate != nil {
				test.mutate(entry)
			}
			q := &mockQuerier{entries: map[string]ledgerentries.FlatLedgerObject{index: entry}}

			issuance, err := getIssuance(q, testIssuanceID)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, "issuerKey", issuance.issuerKey)
			require.True(t, issuance.canTransfer())
			require.Zero(t, issuance.transferFee)
			require.Equal(t, uint64(types.MaxMPTAmount), issuance.confidentialOutstanding)
		})
	}
}

func TestParseLedgerUint64(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    uint64
		wantErr bool
	}{
		{name: "decimal string", value: "1000", want: 1000},
		{name: "zero string", value: "0"},
		{name: "maximum string", value: strconv.FormatUint(math.MaxUint64, 10), want: math.MaxUint64},
		{name: "json number", value: json.Number("42"), want: 42},
		{name: "negative string", value: "-1", wantErr: true},
		{name: "malformed string", value: "nope", wantErr: true},
		{name: "float is not exact", value: float64(1000), wantErr: true},
		{name: "nil", value: nil, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseLedgerUint64("ConfidentialOutstandingAmount", test.value)
			if test.wantErr {
				require.ErrorIs(t, err, ErrLedgerQuery)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}
}
