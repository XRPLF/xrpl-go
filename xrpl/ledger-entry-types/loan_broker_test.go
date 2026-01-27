package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestLoanBroker(t *testing.T) {
	var s Object = &LoanBroker{
		LedgerEntryType:   LoanBrokerEntry,
		Flags:             0,
		Sequence:          3606,
		LoanSequence:      1,
		OwnerNode:         "0000000000000000",
		VaultNode:         "0000000000000000",
		VaultID:           "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
		Account:           "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
		Owner:             "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
		DebtMaximum:       types.XRPLNumber("1000000"),
		PreviousTxnID:     "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
		PreviousTxnLgrSeq: 28991004,
	}

	j := `{
	"LedgerEntryType": "LoanBroker",
	"Flags": 0,
	"Sequence": 3606,
	"LoanSequence": 1,
	"OwnerNode": "0000000000000000",
	"VaultNode": "0000000000000000",
	"VaultID": "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
	"Account": "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
	"Owner": "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
	"DebtMaximum": "1000000",
	"PreviousTxnID": "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
	"PreviousTxnLgrSeq": 28991004
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestLoanBroker_EntryType(t *testing.T) {
	s := &LoanBroker{}
	require.Equal(t, s.EntryType(), LoanBrokerEntry)
}

func TestLoanBroker_WithOptionalFields(t *testing.T) {
	ownerCount := types.OwnerCount(5)
	debtTotal := types.XRPLNumber("500000")
	coverAvailable := types.XRPLNumber("100000")
	coverRateMinimum := types.CoverRate(5000)
	coverRateLiquidation := types.CoverRate(1000)

	var s Object = &LoanBroker{
		LedgerEntryType:      LoanBrokerEntry,
		Flags:                0,
		Sequence:             3606,
		LoanSequence:         1,
		OwnerNode:            "0000000000000000",
		VaultNode:            "0000000000000000",
		VaultID:              "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
		Account:              "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
		Owner:                "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
		OwnerCount:           &ownerCount,
		DebtTotal:            &debtTotal,
		DebtMaximum:          types.XRPLNumber("1000000"),
		CoverAvailable:       &coverAvailable,
		CoverRateMinimum:     &coverRateMinimum,
		CoverRateLiquidation: &coverRateLiquidation,
		PreviousTxnID:        "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
		PreviousTxnLgrSeq:    28991004,
	}

	j := `{
	"LedgerEntryType": "LoanBroker",
	"Flags": 0,
	"Sequence": 3606,
	"LoanSequence": 1,
	"OwnerNode": "0000000000000000",
	"VaultNode": "0000000000000000",
	"VaultID": "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
	"Account": "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
	"Owner": "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
	"OwnerCount": 5,
	"DebtTotal": "500000",
	"DebtMaximum": "1000000",
	"CoverAvailable": "100000",
	"CoverRateMinimum": 5000,
	"CoverRateLiquidation": 1000,
	"PreviousTxnID": "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
	"PreviousTxnLgrSeq": 28991004
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}
