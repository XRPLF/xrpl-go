package transaction

import (
	"strings"
	"testing"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
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
				PaymentTotal:       new(types.PaymentTotal(12)),
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
				"PaymentTotal":       uint32(12),
			},
		},
		{
			name: "pass - explicit zero PaymentTotal",
			tx:   &LoanSet{PaymentTotal: new(types.PaymentTotal(0))},
			expected: FlatTransaction{
				"TransactionType":    LoanSetTx.String(),
				"LoanBrokerID":       "",
				"PrincipalRequested": "",
				"PaymentTotal":       uint32(0),
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
	const (
		secpKey       = "0330E7FC9D56BB25D6893BA3F317AE5BCF33B3291BD63DB32654A313222F7FD020"
		edKey         = "ED5F5AC8B98974A3CA843326D9B88CEBD0560177B973EE0B149F782CFAA06DC66A"
		sig           = "3045022100D184EB4AE5956FF600E7536EE459345C7BBCF097A84CC61A93B9AF7197EDB98702201CEA8009B7BEEBAA2AACC0359B41C427C1C5B550A4CA4B80CF2174AF2D6D5DCE"
		signerAccount = types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD")
	)
	signerFor := func(account types.Address, pubKey, signature string) []types.Signer {
		return []types.Signer{{SignerData: types.SignerData{Account: account, SigningPubKey: pubKey, TxnSignature: signature}}}
	}
	signers := func(pubKey, signature string) []types.Signer {
		return signerFor(signerAccount, pubKey, signature)
	}
	// countersigned returns an otherwise valid LoanSet with the given flags and CounterpartySignature.
	countersigned := func(flags uint32, signature *CounterpartySignature) *LoanSet {
		return &LoanSet{
			BaseTx: BaseTx{
				Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
				TransactionType: LoanSetTx,
				Flags:           flags,
			},
			LoanBrokerID:          "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
			PrincipalRequested:    "100000",
			CounterpartySignature: signature,
		}
	}

	testcases := []struct {
		name     string
		tx       *LoanSet
		expected error
	}{
		// Base transaction and required fields

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

		// PaymentTotal

		{
			name: "fail - explicit zero PaymentTotal",
			tx: &LoanSet{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanSetTx,
				},
				LoanBrokerID:       "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				PrincipalRequested: types.XRPLNumber("100000"),
				PaymentTotal:       new(types.PaymentTotal(0)),
			},
			expected: ErrLoanSetPaymentTotalInvalid,
		},
		{
			name: "pass - PaymentTotal of one",
			tx: &LoanSet{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanSetTx,
				},
				LoanBrokerID:       "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				PrincipalRequested: types.XRPLNumber("100000"),
				PaymentTotal:       new(types.PaymentTotal(1)),
			},
			expected: nil,
		},

		// CounterpartySignature: unsigned and single signature

		{name: "pass - empty for simulation", tx: countersigned(0, &CounterpartySignature{})},
		{name: "pass - SigningPubKey without TxnSignature for simulation", tx: countersigned(0, &CounterpartySignature{SigningPubKey: secpKey})},
		{name: "pass - secp256k1 single signature", tx: countersigned(0, &CounterpartySignature{SigningPubKey: secpKey, TxnSignature: sig})},
		{name: "pass - ed25519 single signature", tx: countersigned(0, &CounterpartySignature{SigningPubKey: edKey, TxnSignature: sig})},
		{name: "pass - lowercase ed25519 prefix", tx: countersigned(0, &CounterpartySignature{SigningPubKey: "ed" + edKey[2:], TxnSignature: sig})},
		{name: "fail - uncompressed key prefix", tx: countersigned(0, &CounterpartySignature{SigningPubKey: "04" + secpKey[2:], TxnSignature: sig}), expected: ErrInvalidCounterpartySignature},
		{name: "fail - 32-byte key", tx: countersigned(0, &CounterpartySignature{SigningPubKey: secpKey[:64], TxnSignature: sig}), expected: ErrInvalidCounterpartySignature},
		{name: "fail - 34-byte key", tx: countersigned(0, &CounterpartySignature{SigningPubKey: secpKey + "00", TxnSignature: sig}), expected: ErrInvalidCounterpartySignature},
		{name: "fail - non-hex key", tx: countersigned(0, &CounterpartySignature{SigningPubKey: secpKey[:64] + "ZZ", TxnSignature: sig}), expected: ErrInvalidCounterpartySignature},
		{name: "fail - odd-length TxnSignature", tx: countersigned(0, &CounterpartySignature{SigningPubKey: secpKey, TxnSignature: sig[1:]}), expected: ErrInvalidCounterpartySignature},
		{name: "fail - non-hex TxnSignature", tx: countersigned(0, &CounterpartySignature{SigningPubKey: secpKey, TxnSignature: "ZZ" + sig[2:]}), expected: ErrInvalidCounterpartySignature},
		{name: "fail - TxnSignature without SigningPubKey", tx: countersigned(0, &CounterpartySignature{TxnSignature: sig}), expected: ErrInvalidCounterpartySignature},

		// CounterpartySignature: Signers

		{name: "pass - multisignature", tx: countersigned(0, &CounterpartySignature{Signers: signers(secpKey, sig)})},
		{name: "pass - simulated multisignature without signer signatures", tx: countersigned(0, &CounterpartySignature{Signers: signers(secpKey, "")})},
		{name: "pass - simulated multisignature without signer key or signature", tx: countersigned(0, &CounterpartySignature{Signers: signers("", "")})},
		{name: "pass - untagged X-address Signer Account", tx: countersigned(0, &CounterpartySignature{Signers: signerFor(xAddr(t, string(signerAccount), 0, false), secpKey, "")})},
		{name: "fail - SigningPubKey with Signers", tx: countersigned(0, &CounterpartySignature{SigningPubKey: secpKey, Signers: signers(secpKey, sig)}), expected: ErrInvalidCounterpartySignature},
		{name: "fail - TxnSignature with Signers", tx: countersigned(0, &CounterpartySignature{TxnSignature: sig, Signers: signers(secpKey, sig)}), expected: ErrInvalidCounterpartySignature},
		{name: "fail - odd-length Signer TxnSignature", tx: countersigned(0, &CounterpartySignature{Signers: signers(secpKey, sig[1:])}), expected: ErrInvalidCounterpartySignature},
		{name: "fail - non-hex Signer SigningPubKey", tx: countersigned(0, &CounterpartySignature{Signers: signers("ZZ"+secpKey[2:], sig)}), expected: ErrInvalidCounterpartySignature},
		{name: "fail - unsupported nonempty Signer SigningPubKey", tx: countersigned(0, &CounterpartySignature{Signers: signers("00", "AB")}), expected: ErrInvalidCounterpartySignature},
		{name: "fail - Signer TxnSignature without SigningPubKey", tx: countersigned(0, &CounterpartySignature{Signers: signers("", sig)}), expected: ErrInvalidCounterpartySignature},
		{name: "fail - missing Signer Account", tx: countersigned(0, &CounterpartySignature{Signers: signerFor("", secpKey, "")}), expected: ErrSignerAccountShouldBeString},
		{name: "fail - malformed Signer Account", tx: countersigned(0, &CounterpartySignature{Signers: signerFor("rInvalid", secpKey, "")}), expected: ErrSignerAccountShouldBeString},
		{name: "fail - tagged X-address Signer Account", tx: countersigned(0, &CounterpartySignature{Signers: signerFor(xAddr(t, string(signerAccount), 1, true), secpKey, "")}), expected: ErrSignerAccountTagNotAllowed},
		{name: "fail - ACCOUNT_ZERO Signer Account", tx: countersigned(0, &CounterpartySignature{Signers: signerFor("rrrrrrrrrrrrrrrrrrrrrhoLvTp", secpKey, "")}), expected: ErrSignerAccountZero},

		// CounterpartySignature: inner Batch transaction

		{name: "pass - empty on an inner Batch transaction", tx: countersigned(types.TfInnerBatchTxn, &CounterpartySignature{})},
		{name: "fail - SigningPubKey on an inner Batch transaction", tx: countersigned(types.TfInnerBatchTxn, &CounterpartySignature{SigningPubKey: secpKey}), expected: ErrInvalidCounterpartySignature},
		{name: "fail - TxnSignature on an inner Batch transaction", tx: countersigned(types.TfInnerBatchTxn, &CounterpartySignature{TxnSignature: sig}), expected: ErrInvalidCounterpartySignature},
		{name: "fail - Signers on an inner Batch transaction", tx: countersigned(types.TfInnerBatchTxn, &CounterpartySignature{Signers: signers(secpKey, "")}), expected: ErrInvalidCounterpartySignature},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			ok, err := testcase.tx.Validate()
			assert.Equal(t, ok, testcase.expected == nil)
			assert.ErrorIs(t, err, testcase.expected)
		})
	}
}

func TestLoanSet_ValidateInterestRateBounds(t *testing.T) {
	maxRate := types.InterestRate(100_000)
	aboveMax := maxRate + 1
	tests := []struct {
		name                                   string
		interest, late, closeRate, overpayment *types.InterestRate
		wantErr                                error
	}{
		{name: "all rates at maximum", interest: &maxRate, late: &maxRate, closeRate: &maxRate, overpayment: &maxRate},
		{name: "interest above maximum", interest: &aboveMax, wantErr: ErrLoanSetInterestRateInvalid},
		{name: "late interest above maximum", late: &aboveMax, wantErr: ErrLoanSetLateInterestRateInvalid},
		{name: "close interest above maximum", closeRate: &aboveMax, wantErr: ErrLoanSetCloseInterestRateInvalid},
		{name: "overpayment interest above maximum", overpayment: &aboveMax, wantErr: ErrLoanSetOverpaymentInterestRateInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := LoanSet{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanSetTx,
				},
				LoanBrokerID:            "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				PrincipalRequested:      "100000",
				InterestRate:            tt.interest,
				LateInterestRate:        tt.late,
				CloseInterestRate:       tt.closeRate,
				OverpaymentInterestRate: tt.overpayment,
			}
			ok, err := tx.Validate()
			require.Equal(t, tt.wantErr == nil, ok)
			require.ErrorIs(t, err, tt.wantErr)
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

func TestLoanSet_CounterpartySignersEncoding(t *testing.T) {
	const (
		secpKey = "0330E7FC9D56BB25D6893BA3F317AE5BCF33B3291BD63DB32654A313222F7FD020"
		sig     = "3045022100D184EB4AE5956FF600E7536EE459345C7BBCF097A84CC61A93B9AF7197EDB98702201CEA8009B7BEEBAA2AACC0359B41C427C1C5B550A4CA4B80CF2174AF2D6D5DCE"
		account = "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"
	)

	tests := []struct {
		name               string
		signerAccount      types.Address
		signingPubKey, txn string
	}{
		{name: "signed multisigner", signerAccount: account, signingPubKey: secpKey, txn: sig},
		{name: "simulated multisigner without signature", signerAccount: account, signingPubKey: secpKey},
		{name: "simulated multisigner without key or signature", signerAccount: account},
		{name: "untagged X-address multisigner", signerAccount: xAddr(t, account, 0, false), signingPubKey: secpKey, txn: sig},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := LoanSet{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanSetTx,
					Fee:             types.XRPCurrencyAmount(10),
					Sequence:        1,
				},
				LoanBrokerID:       "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				PrincipalRequested: "100000",
				CounterpartySignature: &CounterpartySignature{Signers: []types.Signer{{SignerData: types.SignerData{
					Account: tt.signerAccount, SigningPubKey: tt.signingPubKey, TxnSignature: tt.txn,
				}}}},
			}
			ok, err := tx.Validate()
			require.NoError(t, err)
			require.True(t, ok)

			flatSigner := map[string]any{"Account": tt.signerAccount.String(), "SigningPubKey": tt.signingPubKey, "TxnSignature": tt.txn}
			flat := tx.Flatten()
			require.Equal(t, map[string]any{"Signers": []map[string]any{{"Signer": flatSigner}}}, flat["CounterpartySignature"])

			encoded, err := binarycodec.Encode(flat)
			require.NoError(t, err)
			decoded, err := binarycodec.Decode(encoded)
			require.NoError(t, err)
			decodedSigner := map[string]any{"Account": account, "SigningPubKey": tt.signingPubKey, "TxnSignature": tt.txn}
			require.Equal(t, map[string]any{"Signers": []any{map[string]any{"Signer": decodedSigner}}}, decoded["CounterpartySignature"])
		})
	}
}
