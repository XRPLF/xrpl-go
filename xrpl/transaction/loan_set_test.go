package transaction

import (
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoanSet_TxType(t *testing.T) {
	tx := &LoanSet{}
	assert.Equal(t, LoanSetTx, tx.TxType())
}

func TestLoanSet_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *LoanSet
		expected FlatTransaction
	}{
		{
			name: "pass - empty",
			tx:   &LoanSet{},
			expected: FlatTransaction{
				"TransactionType":    LoanSetTx.String(),
				"LoanBrokerID":       "",
				"PrincipalRequested": "",
			},
		},
		{
			name: "pass - explicit zero PaymentTotal preserves presence",
			tx: &LoanSet{
				PaymentTotal: func() *types.PaymentTotal { value := types.PaymentTotal(0); return &value }(),
			},
			expected: FlatTransaction{
				"TransactionType":    LoanSetTx.String(),
				"LoanBrokerID":       "",
				"PrincipalRequested": "",
				"PaymentTotal":       uint32(0),
			},
		},
		{
			name: "pass - explicit zero PaymentInterval and GracePeriod preserve presence",
			tx: &LoanSet{
				PaymentInterval: func() *types.PaymentInterval { v := types.PaymentInterval(0); return &v }(),
				GracePeriod:     func() *types.GracePeriod { v := types.GracePeriod(0); return &v }(),
			},
			expected: FlatTransaction{
				"TransactionType":    LoanSetTx.String(),
				"LoanBrokerID":       "",
				"PrincipalRequested": "",
				"PaymentInterval":    uint32(0),
				"GracePeriod":        uint32(0),
			},
		},
		{
			name: "pass - complete",
			tx: &LoanSet{
				BaseTx: BaseTx{
					Account:            "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					Fee:                1000000,
					Sequence:           1,
					LastLedgerSequence: 3000000,
				},
				LoanBrokerID:       "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				PrincipalRequested: types.XRPLNumber("100000"),
				Counterparty:       func() *types.Address { v := types.Address("rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds"); return &v }(),
				InterestRate:       func() *types.InterestRate { v := types.InterestRate(5000); return &v }(),
				PaymentInterval:    func() *types.PaymentInterval { v := types.PaymentInterval(2592000); return &v }(),
			},
			expected: FlatTransaction{
				"TransactionType":    LoanSetTx.String(),
				"Account":            "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
				"Fee":                "1000000",
				"Sequence":           uint32(1),
				"LastLedgerSequence": uint32(3000000),
				"LoanBrokerID":       "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				"PrincipalRequested": "100000",
				"Counterparty":       "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
				"InterestRate":       uint32(5000),
				"PaymentInterval":    uint32(2592000),
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.expected, testcase.tx.Flatten())
		})
	}
}

func TestLoanSet_Validate(t *testing.T) {
	paymentInterval := func(v uint32) *types.PaymentInterval { p := types.PaymentInterval(v); return &p }
	gracePeriod := func(v uint32) *types.GracePeriod { g := types.GracePeriod(v); return &g }
	schedule := func(interval *types.PaymentInterval, grace *types.GracePeriod) *LoanSet {
		return &LoanSet{
			BaseTx: BaseTx{
				Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
				TransactionType: LoanSetTx,
			},
			LoanBrokerID:       "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
			PrincipalRequested: types.XRPLNumber("100000"),
			PaymentInterval:    interval,
			GracePeriod:        grace,
		}
	}
	amounts := func(principal, origination, service, late, closeFee types.XRPLNumber) *LoanSet {
		tx := schedule(nil, nil)
		tx.PrincipalRequested = principal
		tx.LoanOriginationFee = &origination
		tx.LoanServiceFee = &service
		tx.LatePaymentFee = &late
		tx.ClosePaymentFee = &closeFee
		return tx
	}
	testcases := []struct {
		name     string
		tx       *LoanSet
		expected error
	}{
		{
			name: "fail - zero principal", tx: amounts("0", "", "", "", ""),
			expected: ErrLoanSetPrincipalRequestedInvalid,
		},
		{
			name: "fail - negative principal", tx: amounts("-1", "", "", "", ""),
			expected: ErrLoanSetPrincipalRequestedInvalid,
		},
		{
			name: "fail - negative origination fee", tx: amounts("1", "-1", "", "", ""),
			expected: ErrLoanSetLoanOriginationFeeInvalid,
		},
		{
			name: "fail - negative service fee", tx: amounts("1", "", "-1", "", ""),
			expected: ErrLoanSetLoanServiceFeeInvalid,
		},
		{
			name: "fail - negative late payment fee", tx: amounts("1", "", "", "-1", ""),
			expected: ErrLoanSetLatePaymentFeeInvalid,
		},
		{
			name: "fail - negative close payment fee", tx: amounts("1", "", "", "", "-1"),
			expected: ErrLoanSetClosePaymentFeeInvalid,
		},
		{
			name: "fail - origination fee above principal", tx: amounts("1", "1.000000000000000001", "", "", ""),
			expected: ErrLoanSetLoanOriginationFeeInvalid,
		},
		{
			name: "pass - origination fee that encodes to the principal", tx: amounts("1", "1.0000000000000000001", "", "", ""),
		},
		{
			name: "fail - scientific origination fee above principal", tx: amounts("1e-1000", "2e-1000", "", "", ""),
			expected: ErrLoanSetLoanOriginationFeeInvalid,
		},
		{
			name: "fail - invalid fee syntax", tx: amounts("1", "", "invalid", "", ""),
			expected: ErrLoanSetLoanServiceFeeInvalid,
		},
		{
			name: "fail - principal exponent too large to encode", tx: amounts("1e40000", "", "", "", ""),
			expected: ErrLoanSetPrincipalRequestedInvalid,
		},
		{
			name: "fail - principal exponent too small to encode", tx: amounts("1e-40000", "", "", "", ""),
			expected: ErrLoanSetPrincipalRequestedInvalid,
		},
		{
			name: "fail - principal without integer digits", tx: amounts(".5", "", "", "", ""),
			expected: ErrLoanSetPrincipalRequestedInvalid,
		},
		{
			name: "fail - principal exponent outside int32", tx: amounts("1e4294967296", "", "", "", ""),
			expected: ErrLoanSetPrincipalRequestedInvalid,
		},
		{
			name: "fail - origination fee exponent too large to encode", tx: amounts("1", "1e40000", "", "", ""),
			expected: ErrLoanSetLoanOriginationFeeInvalid,
		},
		{
			name: "fail - service fee exponent too small to encode", tx: amounts("1", "", "1e-40000", "", ""),
			expected: ErrLoanSetLoanServiceFeeInvalid,
		},
		{
			name: "fail - late payment fee without integer digits", tx: amounts("1", "", "", ".5", ""),
			expected: ErrLoanSetLatePaymentFeeInvalid,
		},
		{
			name: "fail - close payment fee exponent outside int32", tx: amounts("1", "", "", "", "1e4294967296"),
			expected: ErrLoanSetClosePaymentFeeInvalid,
		},
		{
			name: "fail - zero LoanBrokerID", tx: func() *LoanSet {
				tx := schedule(nil, nil)
				tx.LoanBrokerID = strings.Repeat("0", 64)
				return tx
			}(),
			expected: ErrLoanSetLoanBrokerIDInvalid,
		},
		{
			name: "pass - principal one", tx: amounts("1", "", "", "", ""),
		},
		{
			name: "pass - origination fee equals principal", tx: amounts("1", "1", "", "", ""),
		},
		{
			name: "pass - scientific fee equals principal", tx: amounts("+1e1000", "10e999", "", "", ""),
		},
		{
			name: "pass - fees at zero", tx: amounts("1", "0", "-0", "0e5", "0.0"),
		},
		{
			name: "fail - base tx invalid",
			tx: &LoanSet{
				BaseTx: BaseTx{
					TransactionType: LoanSetTx,
				},
			},
			expected: ErrInvalidAccount,
		},
		{
			name: "fail - LoanBrokerID required",
			tx: &LoanSet{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanSetTx,
				},
				PrincipalRequested: types.XRPLNumber("100000"),
			},
			expected: ErrLoanSetLoanBrokerIDRequired,
		},
		{
			name: "fail - PrincipalRequested required",
			tx: &LoanSet{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanSetTx,
				},
				LoanBrokerID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
			},
			expected: ErrLoanSetPrincipalRequestedRequired,
		},
		{
			name: "fail - LoanBrokerID invalid",
			tx: &LoanSet{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanSetTx,
				},
				LoanBrokerID:       "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F43",
				PrincipalRequested: types.XRPLNumber("100000"),
			},
			expected: ErrLoanSetLoanBrokerIDInvalid,
		},
		{
			name: "fail - PrincipalRequested invalid",
			tx: &LoanSet{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanSetTx,
				},
				LoanBrokerID:       "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				PrincipalRequested: types.XRPLNumber("invalid"),
			},
			expected: ErrLoanSetPrincipalRequestedInvalid,
		},
		{
			name: "fail - Data too long",
			tx: &LoanSet{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanSetTx,
				},
				LoanBrokerID:       "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				PrincipalRequested: types.XRPLNumber("100000"),
				Data:               func() *types.Data { v := types.Data("A" + strings.Repeat("B", 512)); return &v }(),
			},
			expected: ErrLoanSetDataInvalid,
		},
		{
			name: "fail - OverpaymentFee too high",
			tx: &LoanSet{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanSetTx,
				},
				LoanBrokerID:       "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				PrincipalRequested: types.XRPLNumber("100000"),
				OverpaymentFee:     func() *uint32 { v := uint32(100001); return &v }(),
			},
			expected: ErrLoanSetOverpaymentFeeInvalid,
		},
		{
			name: "fail - PaymentInterval too low",
			tx: &LoanSet{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanSetTx,
				},
				LoanBrokerID:       "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				PrincipalRequested: types.XRPLNumber("100000"),
				PaymentInterval:    func() *types.PaymentInterval { v := types.PaymentInterval(59); return &v }(),
			},
			expected: ErrLoanSetPaymentIntervalInvalid,
		},
		{
			name:     "pass - PaymentInterval omitted",
			tx:       schedule(nil, nil),
			expected: nil,
		},
		{
			name:     "fail - explicit zero PaymentInterval",
			tx:       schedule(paymentInterval(0), nil),
			expected: ErrLoanSetPaymentIntervalInvalid,
		},
		{
			name:     "pass - PaymentInterval at minimum",
			tx:       schedule(paymentInterval(60), nil),
			expected: nil,
		},
		{
			name:     "fail - explicit zero GracePeriod",
			tx:       schedule(paymentInterval(120), gracePeriod(0)),
			expected: ErrLoanSetGracePeriodInvalid,
		},
		{
			name:     "fail - GracePeriod below minimum",
			tx:       schedule(paymentInterval(120), gracePeriod(59)),
			expected: ErrLoanSetGracePeriodInvalid,
		},
		{
			name:     "pass - GracePeriod at minimum",
			tx:       schedule(paymentInterval(120), gracePeriod(60)),
			expected: nil,
		},
		{
			name:     "pass - GracePeriod equal to PaymentInterval",
			tx:       schedule(paymentInterval(120), gracePeriod(120)),
			expected: nil,
		},
		{
			name:     "fail - GracePeriod greater than PaymentInterval",
			tx:       schedule(paymentInterval(120), gracePeriod(121)),
			expected: ErrLoanSetGracePeriodInvalid,
		},
		{
			name:     "pass - GracePeriod 60 with PaymentInterval omitted",
			tx:       schedule(nil, gracePeriod(60)),
			expected: nil,
		},
		{
			name:     "fail - GracePeriod above the default PaymentInterval",
			tx:       schedule(nil, gracePeriod(61)),
			expected: ErrLoanSetGracePeriodInvalid,
		},
		{
			name: "fail - explicit zero PaymentTotal",
			tx: &LoanSet{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanSetTx,
				},
				LoanBrokerID:       "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				PrincipalRequested: types.XRPLNumber("100000"),
				PaymentTotal:       func() *types.PaymentTotal { v := types.PaymentTotal(0); return &v }(),
			},
			expected: ErrLoanSetPaymentTotalInvalid,
		},
		{
			name: "pass - complete",
			tx: &LoanSet{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanSetTx,
				},
				LoanBrokerID:       "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				PrincipalRequested: types.XRPLNumber("100000"),
				InterestRate:       func() *types.InterestRate { v := types.InterestRate(5000); return &v }(),
				PaymentInterval:    func() *types.PaymentInterval { v := types.PaymentInterval(2592000); return &v }(),
			},
			expected: nil,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			ok, err := testcase.tx.Validate()
			assert.Equal(t, ok, testcase.expected == nil)
			if testcase.expected != nil {
				assert.ErrorIs(t, err, testcase.expected)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoanSet_ValidateCounterpartySignature(t *testing.T) {
	counterparty := types.Address("rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds")
	validSigner := types.Signer{SignerData: types.SignerData{
		Account:       counterparty,
		TxnSignature:  "ABCD",
		SigningPubKey: testPublicKey,
	}}
	malformedSigner := validSigner
	malformedSigner.SignerData.TxnSignature = "not-hex"

	testcases := []struct {
		name                  string
		counterparty          *types.Address
		counterpartySignature *CounterpartySignature
		flags                 uint32
		expected              error
	}{
		{
			name:                  "pass - single counterparty signature",
			counterpartySignature: &CounterpartySignature{SigningPubKey: testPublicKey, TxnSignature: "ABCD"},
		},
		{
			name:                  "pass - single counterparty signature with empty Signers",
			counterpartySignature: &CounterpartySignature{SigningPubKey: testPublicKey, TxnSignature: "ABCD", Signers: []types.Signer{}},
		},
		{
			name:                  "fail - empty outer counterparty signature with nil Signers",
			counterpartySignature: &CounterpartySignature{},
			expected:              ErrLoanSetCounterpartySignatureInvalid,
		},
		{
			name:                  "fail - empty Signers without single-sign fields",
			counterpartySignature: &CounterpartySignature{Signers: []types.Signer{}},
			expected:              ErrLoanSetCounterpartySignatureInvalid,
		},
		{
			name:                  "fail - incomplete counterparty signature",
			counterpartySignature: &CounterpartySignature{SigningPubKey: testPublicKey},
			expected:              ErrLoanSetCounterpartySignatureInvalid,
		},
		{
			name:                  "fail - malformed counterparty public key",
			counterpartySignature: &CounterpartySignature{SigningPubKey: "ABCD", TxnSignature: "ABCD"},
			expected:              ErrLoanSetCounterpartySignatureInvalid,
		},
		{
			name: "fail - mixed counterparty signature forms",
			counterpartySignature: &CounterpartySignature{
				SigningPubKey: testPublicKey,
				TxnSignature:  "ABCD",
				Signers:       []types.Signer{validSigner},
			},
			expected: ErrLoanSetCounterpartySignatureInvalid,
		},
		{
			name:                  "pass - multisigned counterparty signature",
			counterpartySignature: &CounterpartySignature{Signers: []types.Signer{validSigner}},
		},
		{
			name:                  "fail - malformed multisigner signature",
			counterpartySignature: &CounterpartySignature{Signers: []types.Signer{malformedSigner}},
			expected:              ErrLoanSetCounterpartySignatureInvalid,
		},
		{
			name:                  "fail - duplicate multisigners",
			counterpartySignature: &CounterpartySignature{Signers: []types.Signer{validSigner, validSigner}},
			expected:              errDuplicateTransactionSigner,
		},
		{
			name:     "fail - inner Batch missing Counterparty",
			flags:    types.TfInnerBatchTxn,
			expected: ErrLoanSetInnerCounterpartyRequired,
		},
		{
			name:                  "fail - inner Batch has counterparty signature fields",
			flags:                 types.TfInnerBatchTxn,
			counterparty:          &counterparty,
			counterpartySignature: &CounterpartySignature{SigningPubKey: testPublicKey, TxnSignature: "ABCD"},
			expected:              ErrLoanSetInnerCounterpartySignature,
		},
		{
			name:         "pass - unsigned inner Batch with Counterparty",
			flags:        types.TfInnerBatchTxn,
			counterparty: &counterparty,
		},
		{
			name:                  "pass - inner Batch counterparty signature with nil Signers",
			flags:                 types.TfInnerBatchTxn,
			counterparty:          &counterparty,
			counterpartySignature: &CounterpartySignature{},
		},
		{
			name:                  "pass - inner Batch counterparty signature with empty Signers",
			flags:                 types.TfInnerBatchTxn,
			counterparty:          &counterparty,
			counterpartySignature: &CounterpartySignature{Signers: []types.Signer{}},
		},
		{
			name:                  "fail - inner Batch counterparty signature with nonempty Signers",
			flags:                 types.TfInnerBatchTxn,
			counterparty:          &counterparty,
			counterpartySignature: &CounterpartySignature{Signers: []types.Signer{validSigner}},
			expected:              ErrLoanSetInnerCounterpartySignature,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			tx := &LoanSet{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanSetTx,
					Flags:           testcase.flags,
				},
				LoanBrokerID:          "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				PrincipalRequested:    types.XRPLNumber("100000"),
				Counterparty:          testcase.counterparty,
				CounterpartySignature: testcase.counterpartySignature,
			}
			var before *CounterpartySignature
			if tx.CounterpartySignature != nil {
				signature := *tx.CounterpartySignature
				before = &signature
			}

			ok, err := tx.Validate()
			require.Equal(t, before, tx.CounterpartySignature, "validation must preserve nil versus empty Signers")
			if testcase.expected == nil {
				require.True(t, ok)
				require.NoError(t, err)
				return
			}
			require.False(t, ok)
			require.ErrorIs(t, err, testcase.expected)
		})
	}
}

func TestLoanSet_Flags(t *testing.T) {
	tests := []struct {
		name     string
		setter   func(*LoanSet)
		expected uint32
	}{
		{
			name: "pass - SetLoanOverpaymentFlag",
			setter: func(ls *LoanSet) {
				ls.SetLoanOverpaymentFlag()
			},
			expected: TfLoanOverpayment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := &LoanSet{}
			tt.setter(ls)
			if ls.Flags != tt.expected {
				t.Errorf("Expected LoanSet Flags to be %d, got %d", tt.expected, ls.Flags)
			}
		})
	}
}
