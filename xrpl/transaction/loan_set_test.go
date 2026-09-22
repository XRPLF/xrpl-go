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
			name: "explicit zero PaymentTotal preserves presence",
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
	testcases := []struct {
		name     string
		tx       *LoanSet
		expected error
	}{
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
				assert.Contains(t, err.Error(), testcase.expected.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoanSet_ValidateCounterpartySignature(t *testing.T) {
	const validPublicKey = "ED5F5AC8B98974A3CA843326D9B88CEBD0560177B973EE0B149F782CFAA06DC66A"
	counterparty := types.Address("rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds")
	validSigner := types.Signer{SignerData: types.SignerData{
		Account:       counterparty,
		TxnSignature:  "ABCD",
		SigningPubKey: validPublicKey,
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
			counterpartySignature: &CounterpartySignature{SigningPubKey: validPublicKey, TxnSignature: "ABCD"},
		},
		{
			name:                  "fail - incomplete counterparty signature",
			counterpartySignature: &CounterpartySignature{SigningPubKey: validPublicKey},
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
				SigningPubKey: validPublicKey,
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
			counterpartySignature: &CounterpartySignature{SigningPubKey: validPublicKey, TxnSignature: "ABCD"},
			expected:              ErrLoanSetInnerCounterpartySignature,
		},
		{
			name:         "pass - unsigned inner Batch with Counterparty",
			flags:        types.TfInnerBatchTxn,
			counterparty: &counterparty,
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

			ok, err := tx.Validate()
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
