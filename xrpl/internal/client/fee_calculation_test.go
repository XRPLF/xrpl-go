package client

import (
	"cmp"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Peersyst/xrpl-go/xrpl/currency"
	ledgertypes "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	transactiontypes "github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// defaultTestMaxFeeXRP is the fee cap of cases that do not exercise capping.
const defaultTestMaxFeeXRP = "2"

const (
	testCounterparty    = "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH"
	testLoanBrokerOwner = "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1"
)

// errRequest is the transport failure the request stubs report.
var errRequest = errors.New("request failed")

func float64Pointer(value float64) *float64 {
	return &value
}

func uint64Pointer(value uint64) *uint64 {
	return &value
}

// signerLists builds the account_info signer list data for an account with the
// given number of signer entries.
func signerLists(entries int) []ledgertypes.SignerList {
	return []ledgertypes.SignerList{{SignerEntries: make([]ledgertypes.SignerEntryWrapper, entries)}}
}

// feeLedger holds the canned validated-ledger data that the fee queries read,
// so a case declares only the values it depends on.
type feeLedger struct {
	baseFeeXRP *float64
	loadFactor float64
	reserveInc *uint64
	// loanBrokerOwner is the Owner that the LoanBroker ledger entry reports.
	loanBrokerOwner string
	// accountSignerLists is the signer list data that account_info reports.
	accountSignerLists []ledgertypes.SignerList
	// signerListAccount records the account whose signer lists were requested.
	signerListAccount transactiontypes.Address
}

// requestFunc answers every fee query from the canned ledger data. Fee queries
// must read the validated ledger, so the stub asserts that selector too.
func (l *feeLedger) requestFunc(t *testing.T) RequestResultFunc {
	t.Helper()
	return func(_ context.Context, req Request, result any) error {
		switch request := req.(type) {
		case *server.InfoRequest:
			response, ok := result.(*server.InfoResponse)
			require.True(t, ok)
			response.Info.ValidatedLedger.BaseFeeXRP = l.baseFeeXRP
			response.Info.LoadFactor = l.loadFactor
		case *server.StateRequest:
			response, ok := result.(*server.StateResponse)
			require.True(t, ok)
			response.State.ValidatedLedger.ReserveInc = l.reserveInc
		case *ledger.EntryRequest:
			require.Equal(t, common.LedgerTitle("validated"), request.LedgerIndex)
			response, ok := result.(*ledger.EntryResponse)
			require.True(t, ok)
			response.Node = ledgertypes.FlatLedgerObject{"Owner": l.loanBrokerOwner}
		case *account.InfoRequest:
			require.Equal(t, common.LedgerTitle("validated"), request.LedgerIndex)
			require.True(t, request.SignerLists)
			l.signerListAccount = request.Account
			response, ok := result.(*account.InfoResponse)
			require.True(t, ok)
			response.SignerLists = l.accountSignerLists
		default:
			t.Fatalf("unexpected request %T", req)
		}
		return nil
	}
}

// failingRequestFunc fails the query that uses the given method and answers
// every other query from the canned ledger data.
func (l *feeLedger) failingRequestFunc(t *testing.T, method string) RequestResultFunc {
	t.Helper()
	answer := l.requestFunc(t)
	return func(ctx context.Context, req Request, result any) error {
		if req.Method() == method {
			return errRequest
		}
		return answer(ctx, req, result)
	}
}

func TestCalculateFee(t *testing.T) {
	t.Parallel()

	// A sixteen byte fulfillment, which adds one base fee to the EscrowFinish cost.
	const sixteenByteFulfillment = "00000000000000000000000000000000"

	tests := []struct {
		name        string
		tx          transaction.FlatTransaction
		nSigners    uint64
		cushion     float64
		maxFeeXRP   string
		ledger      feeLedger
		expectedFee string
	}{
		{
			name:        "ordinary transaction",
			tx:          transaction.FlatTransaction{"TransactionType": transaction.PaymentTx},
			cushion:     1,
			ledger:      feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1},
			expectedFee: "10",
		},
		{
			name:        "ordinary transaction with two signers",
			tx:          transaction.FlatTransaction{"TransactionType": transaction.PaymentTx},
			nSigners:    2,
			cushion:     1,
			ledger:      feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1},
			expectedFee: "30",
		},
		{
			name:        "cushion scales the network fee",
			tx:          transaction.FlatTransaction{"TransactionType": transaction.PaymentTx},
			cushion:     1.2,
			ledger:      feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1},
			expectedFee: "12",
		},
		{
			name:        "maximum fee caps an ordinary transaction",
			tx:          transaction.FlatTransaction{"TransactionType": transaction.PaymentTx},
			nSigners:    2,
			cushion:     1,
			maxFeeXRP:   "0.000015",
			ledger:      feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1},
			expectedFee: "15",
		},
		{
			name:        "confidential MPT transaction",
			tx:          transaction.FlatTransaction{"TransactionType": transaction.ConfidentialMPTSendTx},
			cushion:     1,
			ledger:      feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1},
			expectedFee: "100",
		},
		{
			name:        "confidential MPT transaction with two signers",
			tx:          transaction.FlatTransaction{"TransactionType": transaction.ConfidentialMPTSendTx},
			nSigners:    2,
			cushion:     1,
			ledger:      feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1},
			expectedFee: "120",
		},
		{
			name:        "fractional network fee rounds once",
			tx:          transaction.FlatTransaction{"TransactionType": transaction.PaymentTx},
			cushion:     1,
			ledger:      feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1.24},
			expectedFee: "12",
		},
		{
			name:        "signers do not amplify a rounded drop",
			tx:          transaction.FlatTransaction{"TransactionType": transaction.PaymentTx},
			nSigners:    2,
			cushion:     1,
			ledger:      feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1.24},
			expectedFee: "37",
		},
		{
			name:        "confidential multiplier does not amplify a rounded drop",
			tx:          transaction.FlatTransaction{"TransactionType": transaction.ConfidentialMPTSendTx},
			cushion:     1,
			ledger:      feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1.24},
			expectedFee: "124",
		},
		{
			name:        "confidential MPT transaction with signers on a fractional network fee",
			tx:          transaction.FlatTransaction{"TransactionType": transaction.ConfidentialMPTSendTx},
			nSigners:    2,
			cushion:     1,
			ledger:      feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1.24},
			expectedFee: "149",
		},
		{
			name:        "EscrowFinish without a fulfillment",
			tx:          transaction.FlatTransaction{"TransactionType": transaction.EscrowFinishTx},
			cushion:     1,
			ledger:      feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1},
			expectedFee: "10",
		},
		{
			name: "EscrowFinish with an empty fulfillment",
			tx: transaction.FlatTransaction{
				"TransactionType": transaction.EscrowFinishTx,
				"Fulfillment":     "",
			},
			cushion:     1,
			ledger:      feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1},
			expectedFee: "330",
		},
		{
			name: "EscrowFinish scales with the fulfillment size",
			tx: transaction.FlatTransaction{
				"TransactionType": transaction.EscrowFinishTx,
				"Fulfillment":     sixteenByteFulfillment,
			},
			cushion:     1,
			ledger:      feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1},
			expectedFee: "340",
		},
		{
			name:    "AccountDelete charges the owner reserve above the maximum fee",
			tx:      transaction.FlatTransaction{"TransactionType": transaction.AccountDeleteTx},
			cushion: 1,
			ledger: feeLedger{
				baseFeeXRP: float64Pointer(0.00001),
				loadFactor: 1,
				reserveInc: uint64Pointer(5000000),
			},
			expectedFee: "5000000",
		},
		{
			name:     "AccountDelete adds one base fee per signer",
			tx:       transaction.FlatTransaction{"TransactionType": transaction.AccountDeleteTx},
			nSigners: 1,
			cushion:  1,
			ledger: feeLedger{
				baseFeeXRP: float64Pointer(0.00001),
				loadFactor: 1,
				reserveInc: uint64Pointer(2000000),
			},
			expectedFee: "2000010",
		},
		{
			name:    "AMMCreate charges the owner reserve",
			tx:      transaction.FlatTransaction{"TransactionType": transaction.AMMCreateTx},
			cushion: 1,
			ledger: feeLedger{
				baseFeeXRP: float64Pointer(0.00001),
				loadFactor: 1,
				reserveInc: uint64Pointer(2000000),
			},
			expectedFee: "2000000",
		},
		{
			name: "LoanSet with an explicit counterparty",
			tx: transaction.FlatTransaction{
				"TransactionType": transaction.LoanSetTx,
				"Counterparty":    testCounterparty,
			},
			cushion: 1,
			ledger: feeLedger{
				baseFeeXRP:         float64Pointer(0.00001),
				loadFactor:         1,
				accountSignerLists: signerLists(3),
			},
			expectedFee: "40",
		},
		{
			name: "LoanSet resolves the counterparty from the LoanBroker",
			tx: transaction.FlatTransaction{
				"TransactionType": transaction.LoanSetTx,
				"LoanBrokerID":    "ABC",
			},
			cushion: 1,
			ledger: feeLedger{
				baseFeeXRP:      float64Pointer(0.00001),
				loadFactor:      1,
				loanBrokerOwner: testLoanBrokerOwner,
			},
			expectedFee: "20",
		},
		{
			name: "LoanSet adds one base fee per signer",
			tx: transaction.FlatTransaction{
				"TransactionType": transaction.LoanSetTx,
				"Counterparty":    testCounterparty,
			},
			nSigners: 2,
			cushion:  1,
			ledger: feeLedger{
				baseFeeXRP:         float64Pointer(0.00001),
				loadFactor:         1,
				accountSignerLists: signerLists(3),
			},
			expectedFee: "60",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			settings := FeeSettings{
				Cushion:   test.cushion,
				MaxFeeXRP: cmp.Or(test.maxFeeXRP, defaultTestMaxFeeXRP),
			}

			fee, err := CalculateFee(context.Background(), test.ledger.requestFunc(t), &test.tx, test.nSigners, settings)
			require.NoError(t, err)
			require.Equal(t, test.expectedFee, test.tx["Fee"])

			returnedFee, err := fee.WholeString()
			require.NoError(t, err)
			require.Equal(t, test.expectedFee, returnedFee)
		})
	}
}

// TestCalculateFeeBatch pins the Batch total and the zeroed inner fees that a
// Batch requires, which no other case observes.
func TestCalculateFeeBatch(t *testing.T) {
	t.Parallel()

	tx := transaction.FlatTransaction{
		"TransactionType": transaction.BatchTx,
		"RawTransactions": []map[string]any{
			{"RawTransaction": map[string]any{"TransactionType": transaction.ConfidentialMPTSendTx}},
			{"RawTransaction": map[string]any{"TransactionType": transaction.PaymentTx}},
		},
	}
	data := feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1}
	settings := FeeSettings{Cushion: 1, MaxFeeXRP: defaultTestMaxFeeXRP}

	_, err := CalculateFee(context.Background(), data.requestFunc(t), &tx, 0, settings)
	require.NoError(t, err)
	// Two base fees for the Batch itself, plus the 100 and 10 drop inner fees.
	require.Equal(t, "130", tx["Fee"])

	rawTransactions, ok := tx["RawTransactions"].([]map[string]any)
	require.True(t, ok)
	for _, rawTx := range rawTransactions {
		innerTx, ok := rawTx["RawTransaction"].(map[string]any)
		require.True(t, ok)
		require.Equal(t, "0", innerTx["Fee"])
	}
}

// TestCalculateFeeBatchSumsInnerFeesExactly pins that the Batch total sums the
// exact inner fees and rounds once. Rounding every inner fee first discards its
// fractional drop, which the inner count then multiplies into a shortfall.
func TestCalculateFeeBatchSumsInnerFeesExactly(t *testing.T) {
	t.Parallel()

	rawTransactions := make([]map[string]any, 0, 8)
	for range 8 {
		rawTransactions = append(rawTransactions, map[string]any{
			"RawTransaction": map[string]any{"TransactionType": transaction.PaymentTx},
		})
	}
	tx := transaction.FlatTransaction{
		"TransactionType": transaction.BatchTx,
		"RawTransactions": rawTransactions,
	}
	// A ten drop base fee under this load factor is 10.4 drops per base fee.
	data := feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1.04}
	settings := FeeSettings{Cushion: 1, MaxFeeXRP: defaultTestMaxFeeXRP}

	_, err := CalculateFee(context.Background(), data.requestFunc(t), &tx, 0, settings)
	require.NoError(t, err)
	// Ten base fees: two for the Batch itself and one per inner transaction.
	require.Equal(t, "104", tx["Fee"])
}

func TestCalculateFeeRejectsIncompleteData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		tx          transaction.FlatTransaction
		maxFeeXRP   string
		ledger      feeLedger
		expectedErr error
	}{
		{
			name:        "invalid maximum fee",
			tx:          transaction.FlatTransaction{"TransactionType": transaction.PaymentTx},
			maxFeeXRP:   "invalid",
			ledger:      feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1},
			expectedErr: ErrInvalidFeeValue,
		},
		{
			name:        "absent base fee",
			tx:          transaction.FlatTransaction{"TransactionType": transaction.PaymentTx},
			ledger:      feeLedger{loadFactor: 1},
			expectedErr: ErrCouldNotGetBaseFeeXrp,
		},
		{
			name:        "absent owner reserve",
			tx:          transaction.FlatTransaction{"TransactionType": transaction.AccountDeleteTx},
			ledger:      feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1},
			expectedErr: ErrCouldNotFetchOwnerReserve,
		},
		{
			name:        "LoanSet without a counterparty or a LoanBroker",
			tx:          transaction.FlatTransaction{"TransactionType": transaction.LoanSetTx},
			ledger:      feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1},
			expectedErr: ErrLoanBrokerIDRequired,
		},
		{
			name: "LoanSet with a LoanBroker that has no owner",
			tx: transaction.FlatTransaction{
				"TransactionType": transaction.LoanSetTx,
				"LoanBrokerID":    "ABC",
			},
			ledger:      feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1},
			expectedErr: ErrCouldNotFetchLoanBrokerOwner,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			settings := FeeSettings{
				Cushion:   1,
				MaxFeeXRP: cmp.Or(test.maxFeeXRP, defaultTestMaxFeeXRP),
			}

			_, err := CalculateFee(context.Background(), test.ledger.requestFunc(t), &test.tx, 0, settings)
			require.ErrorIs(t, err, test.expectedErr)
		})
	}
}

func TestCalculateFeePropagatesRequestFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		tx     transaction.FlatTransaction
		ledger feeLedger
		failOn string
	}{
		{
			name:   "server_info",
			tx:     transaction.FlatTransaction{"TransactionType": transaction.PaymentTx},
			ledger: feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1},
			failOn: "server_info",
		},
		{
			name:   "server_state",
			tx:     transaction.FlatTransaction{"TransactionType": transaction.AccountDeleteTx},
			ledger: feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1},
			failOn: "server_state",
		},
		{
			name: "ledger_entry",
			tx: transaction.FlatTransaction{
				"TransactionType": transaction.LoanSetTx,
				"LoanBrokerID":    "ABC",
			},
			ledger: feeLedger{
				baseFeeXRP:      float64Pointer(0.00001),
				loadFactor:      1,
				loanBrokerOwner: testLoanBrokerOwner,
			},
			failOn: "ledger_entry",
		},
		{
			name: "account_info",
			tx: transaction.FlatTransaction{
				"TransactionType": transaction.LoanSetTx,
				"Counterparty":    testCounterparty,
			},
			ledger: feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1},
			failOn: "account_info",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			settings := FeeSettings{Cushion: 1, MaxFeeXRP: defaultTestMaxFeeXRP}
			request := test.ledger.failingRequestFunc(t, test.failOn)

			_, err := CalculateFee(context.Background(), request, &test.tx, 0, settings)
			require.ErrorIs(t, err, errRequest)
		})
	}
}

func TestNetworkFeeDropsFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		baseFeeXRP    *float64
		loadFactor    float64
		cushion       float64
		maxFeeXRP     string
		expectedDrops string
		expectedErr   error
	}{
		{name: "absent base fee", loadFactor: 1, cushion: 1, expectedErr: ErrCouldNotGetBaseFeeXrp},
		{name: "explicit zero base fee", baseFeeXRP: float64Pointer(0), loadFactor: 1, cushion: 1, expectedDrops: "0"},
		{name: "default base fee", baseFeeXRP: float64Pointer(0.00001), loadFactor: 1, cushion: 1, expectedDrops: "10"},
		{name: "load factor scales the base fee", baseFeeXRP: float64Pointer(0.00001), loadFactor: 2, cushion: 1, expectedDrops: "20"},
		{name: "cushion scales the base fee", baseFeeXRP: float64Pointer(0.00001), loadFactor: 1, cushion: 1.2, expectedDrops: "12"},
		{name: "maximum fee caps the network fee", baseFeeXRP: float64Pointer(1), loadFactor: 1000, cushion: 1, maxFeeXRP: "0.123456", expectedDrops: "123456"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			data := feeLedger{baseFeeXRP: test.baseFeeXRP, loadFactor: test.loadFactor}
			maxFee, err := ParseFeeXRP(cmp.Or(test.maxFeeXRP, defaultTestMaxFeeXRP))
			require.NoError(t, err)

			actual, err := networkFeeDropsFor(context.Background(), data.requestFunc(t), test.cushion, maxFee)
			if test.expectedErr != nil {
				require.ErrorIs(t, err, test.expectedErr)
				return
			}
			require.NoError(t, err)
			actualDrops, err := actual.WholeString()
			require.NoError(t, err)
			require.Equal(t, test.expectedDrops, actualDrops)
		})
	}
}

// TestNetworkFeeDropsForKeepsFractionalDrops pins the contract callers rely on:
// the query returns the exact network fee, so the base-fee factor multiplies an
// unrounded value.
func TestNetworkFeeDropsForKeepsFractionalDrops(t *testing.T) {
	t.Parallel()

	data := feeLedger{baseFeeXRP: float64Pointer(0.00001), loadFactor: 1.24}
	maxFee, err := ParseFeeXRP(defaultTestMaxFeeXRP)
	require.NoError(t, err)

	netFee, err := networkFeeDropsFor(context.Background(), data.requestFunc(t), 1, maxFee)
	require.NoError(t, err)
	require.False(t, netFee.IsWhole())

	// Ten drops under a load factor of 1.24 is 12.4 drops.
	tenthDrops, err := netFee.Mul(10).WholeString()
	require.NoError(t, err)
	require.Equal(t, "124", tenthDrops)
}

func TestConfidentialFeeFactor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		txType   transaction.TxType
		expected uint64
	}{
		{name: "Clawback", txType: transaction.ConfidentialMPTClawbackTx, expected: 10},
		{name: "Convert", txType: transaction.ConfidentialMPTConvertTx, expected: 10},
		{name: "ConvertBack", txType: transaction.ConfidentialMPTConvertBackTx, expected: 10},
		{name: "MergeInbox", txType: transaction.ConfidentialMPTMergeInboxTx, expected: 10},
		{name: "Send", txType: transaction.ConfidentialMPTSendTx, expected: 10},
		{name: "ordinary transaction", txType: transaction.PaymentTx, expected: 1},
		{name: "special variable-fee transaction", txType: transaction.BatchTx, expected: 1},
		{name: "pseudo-transaction", txType: transaction.EnableAmendmentTx, expected: 1},
		{name: "unknown transaction", txType: transaction.TxType("FutureTransaction"), expected: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, test.expected, confidentialFeeFactor(test.txType))
		})
	}
}

func TestOwnerReserveFee(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		reserveInc  *uint64
		expected    uint64
		expectedErr error
	}{
		{name: "absent reserve", expectedErr: ErrCouldNotFetchOwnerReserve},
		{name: "explicit zero reserve", reserveInc: uint64Pointer(0), expected: 0},
		{name: "positive reserve", reserveInc: uint64Pointer(2000000), expected: 2000000},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			data := feeLedger{reserveInc: test.reserveInc}

			actual, err := ownerReserveFee(context.Background(), data.requestFunc(t))
			if test.expectedErr != nil {
				require.ErrorIs(t, err, test.expectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestCounterPartySignersCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		tx              transaction.FlatTransaction
		ledger          feeLedger
		expectedCount   uint64
		expectedAccount transactiontypes.Address
	}{
		{
			name:            "explicit counterparty with a signer list",
			tx:              transaction.FlatTransaction{"Counterparty": testCounterparty},
			ledger:          feeLedger{accountSignerLists: signerLists(3)},
			expectedCount:   3,
			expectedAccount: testCounterparty,
		},
		{
			name:            "explicit counterparty without a signer list",
			tx:              transaction.FlatTransaction{"Counterparty": testCounterparty},
			expectedCount:   1,
			expectedAccount: testCounterparty,
		},
		{
			name:            "explicit counterparty with an empty signer list",
			tx:              transaction.FlatTransaction{"Counterparty": testCounterparty},
			ledger:          feeLedger{accountSignerLists: signerLists(0)},
			expectedCount:   0,
			expectedAccount: testCounterparty,
		},
		{
			name: "counterparty resolved from the LoanBroker",
			tx:   transaction.FlatTransaction{"LoanBrokerID": "ABC"},
			ledger: feeLedger{
				loanBrokerOwner:    testLoanBrokerOwner,
				accountSignerLists: signerLists(2),
			},
			expectedCount:   2,
			expectedAccount: testLoanBrokerOwner,
		},
		{
			name: "LoanBroker owner without a signer list",
			tx:   transaction.FlatTransaction{"LoanBrokerID": "ABC"},
			ledger: feeLedger{
				loanBrokerOwner: testLoanBrokerOwner,
			},
			expectedCount:   1,
			expectedAccount: testLoanBrokerOwner,
		},
		{
			name: "explicit counterparty takes precedence over the LoanBroker",
			tx: transaction.FlatTransaction{
				"Counterparty": testCounterparty,
				"LoanBrokerID": "ABC",
			},
			ledger:          feeLedger{loanBrokerOwner: testLoanBrokerOwner},
			expectedCount:   1,
			expectedAccount: testCounterparty,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual, err := counterPartySignersCount(context.Background(), test.ledger.requestFunc(t), test.tx)
			require.NoError(t, err)
			require.Equal(t, test.expectedCount, actual)
			require.Equal(t, test.expectedAccount, test.ledger.signerListAccount)
		})
	}
}

func TestCounterPartySignersCountRejectsMissingBrokerData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		tx          transaction.FlatTransaction
		ledger      feeLedger
		expectedErr error
	}{
		{
			name:        "no counterparty and no LoanBrokerID",
			tx:          transaction.FlatTransaction{},
			expectedErr: ErrLoanBrokerIDRequired,
		},
		{
			name:        "empty counterparty and empty LoanBrokerID",
			tx:          transaction.FlatTransaction{"Counterparty": "", "LoanBrokerID": ""},
			expectedErr: ErrLoanBrokerIDRequired,
		},
		{
			name:        "LoanBrokerID is not a string",
			tx:          transaction.FlatTransaction{"LoanBrokerID": 1},
			expectedErr: ErrLoanBrokerIDRequired,
		},
		{
			name:        "LoanBroker without an owner",
			tx:          transaction.FlatTransaction{"LoanBrokerID": "ABC"},
			expectedErr: ErrCouldNotFetchLoanBrokerOwner,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := counterPartySignersCount(context.Background(), test.ledger.requestFunc(t), test.tx)
			require.ErrorIs(t, err, test.expectedErr)
		})
	}
}

func TestBatchFeesRejectsMalformedBatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		tx          transaction.FlatTransaction
		expectedErr error
	}{
		{
			name:        "missing RawTransactions",
			tx:          transaction.FlatTransaction{},
			expectedErr: ErrRawTransactionsFieldMissing,
		},
		{
			name:        "missing RawTransaction wrapper",
			tx:          transaction.FlatTransaction{"RawTransactions": []map[string]any{{}}},
			expectedErr: ErrRawTransactionFieldMissing,
		},
		{
			name: "nested Batch",
			tx: transaction.FlatTransaction{
				"RawTransactions": []map[string]any{{
					"RawTransaction": map[string]any{
						"TransactionType": transaction.BatchTx,
					},
				}},
			},
			expectedErr: transactiontypes.ErrBatchNestedTransaction,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// Every case fails before issuing a request, so neither the
			// transport nor the fee values are reached.
			request := func(context.Context, Request, any) error {
				t.Fatal("unexpected request")
				return nil
			}
			_, err := batchFees(context.Background(), request, &test.tx, currency.Drops{}, currency.Drops{})
			require.ErrorIs(t, err, test.expectedErr)
		})
	}
}
