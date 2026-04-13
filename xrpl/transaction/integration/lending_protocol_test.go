package integration

import (
	"strconv"
	"testing"

	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	xrplledger "github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

// stringFromFlatObject extracts a string field from a FlatLedgerObject.
func stringFromFlatObject(obj ledger.FlatLedgerObject, field string) (string, bool) {
	v, ok := obj[field]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// findLoanObject searches account objects for a Loan with the given index.
func findLoanObject(objects []ledger.FlatLedgerObject, loanObjectID string) ledger.FlatLedgerObject {
	for _, obj := range objects {
		if idx, ok := stringFromFlatObject(obj, "index"); ok && idx == loanObjectID {
			return obj
		}
	}
	return nil
}

// testIntegrationLendingProtocolSingleSigning tests the full lending protocol lifecycle
// with single signing: VaultCreate -> VaultDeposit -> LoanBrokerSet -> LoanSet with
// counterparty signing -> LoanPay (partial) -> LoanPay (full with tfLoanFullPayment flag).
func testIntegrationLendingProtocolSingleSigning(t *testing.T, client integration.Client) {
	t.Helper()

	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 3,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	// Wallet 0: Vault Owner / Loan Broker
	vaultOwner := runner.GetWallet(0)
	loanBroker := vaultOwner
	// Wallet 1: Depositor
	depositor := runner.GetWallet(1)
	// Wallet 2: Borrower
	borrower := runner.GetWallet(2)

	// ========== STEP 1: Create Vault ==========
	assetsMaximum := types.XRPLNumber("1e17")
	vaultCreateTx := &transaction.VaultCreate{
		BaseTx: transaction.BaseTx{
			Account: vaultOwner.GetAddress(),
		},
		Asset: ledger.Asset{
			Currency: "XRP",
		},
		AssetsMaximum: &assetsMaximum,
	}

	flatVaultCreateTx := vaultCreateTx.Flatten()
	vaultCreateResp, err := runner.TestTransaction(&flatVaultCreateTx, vaultOwner, "tesSUCCESS", nil)
	require.NoError(t, err)
	require.NotNil(t, vaultCreateResp)

	vaultCreateAccount, ok := vaultCreateResp.Tx["Account"].(string)
	require.True(t, ok)
	vaultCreateSequence := integration.TxFieldUint32(t, vaultCreateResp.Tx, "Sequence")

	vaultObjectID, err := xrplhash.Vault(vaultCreateAccount, vaultCreateSequence)
	require.NoError(t, err)

	// ========== STEP 2: Deposit Funds into Vault ==========
	vaultDepositTx := &transaction.VaultDeposit{
		BaseTx: transaction.BaseTx{
			Account: depositor.GetAddress(),
		},
		VaultID: types.Hash256(vaultObjectID),
		Amount:  types.XRPCurrencyAmount(10_000_000),
	}

	flatVaultDepositTx := vaultDepositTx.Flatten()
	_, err = runner.TestTransaction(&flatVaultDepositTx, depositor, "tesSUCCESS", nil)
	require.NoError(t, err)

	// ========== STEP 3: Create Loan Broker ==========
	debtMaximum := types.XRPLNumber("25000000")
	managementFeeRate := types.InterestRate(10000)
	loanBrokerSetTx := &transaction.LoanBrokerSet{
		BaseTx: transaction.BaseTx{
			Account: loanBroker.GetAddress(),
		},
		VaultID:           vaultObjectID,
		DebtMaximum:       &debtMaximum,
		ManagementFeeRate: &managementFeeRate,
	}

	flatLoanBrokerTx := loanBrokerSetTx.Flatten()
	loanBrokerTxResp, err := runner.TestTransaction(&flatLoanBrokerTx, loanBroker, "tesSUCCESS", nil)
	require.NoError(t, err)
	require.NotNil(t, loanBrokerTxResp)

	loanBrokerAccount, ok := loanBrokerTxResp.Tx["Account"].(string)
	require.True(t, ok)
	loanBrokerSequence := integration.TxFieldUint32(t, loanBrokerTxResp.Tx, "Sequence")

	loanBrokerObjectID, err := xrplhash.LoanBroker(loanBrokerAccount, loanBrokerSequence)
	require.NoError(t, err)

	// Verify LoanBroker object was created with correct fields
	loanBrokerObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account: loanBroker.GetAddress(),
		Type:    account.LoanBrokerObject,
	})
	require.NoError(t, err)
	require.Len(t, loanBrokerObjects.AccountObjects, 1)

	loanBrokerObj := loanBrokerObjects.AccountObjects[0]
	loanBrokerObjIndex, ok := stringFromFlatObject(loanBrokerObj, "index")
	require.True(t, ok)
	require.Equal(t, loanBrokerObjectID, loanBrokerObjIndex)

	loanBrokerDebtMaximum, ok := stringFromFlatObject(loanBrokerObj, "DebtMaximum")
	require.True(t, ok)
	require.Equal(t, debtMaximum.String(), loanBrokerDebtMaximum)

	// Get the LoanSequence before the loan is created (used to compute the Loan hash)
	loanBrokerLoanSequence := integration.TxFieldUint32(t, loanBrokerObj, "LoanSequence")

	// ========== STEP 4: Create Loan ==========
	counterparty := types.Address(borrower.GetAddress())
	paymentTotal := types.PaymentTotal(3)
	loanSetTx := &transaction.LoanSet{
		BaseTx: transaction.BaseTx{
			Account: loanBroker.GetAddress(),
		},
		LoanBrokerID:       loanBrokerObjectID,
		PrincipalRequested: types.XRPLNumber("5000000"),
		Counterparty:       &counterparty,
		PaymentTotal:       &paymentTotal,
	}

	flatLoanSetTx := loanSetTx.Flatten()
	err = client.Autofill(&flatLoanSetTx)
	require.NoError(t, err)

	brokerBlob, _, err := loanBroker.Sign(flatLoanSetTx)
	require.NoError(t, err)

	_, counterpartyBlob, _, err := wallet.SignLoanSetByCounterpartyBlob(*borrower, brokerBlob, nil)
	require.NoError(t, err)
	require.NotEmpty(t, counterpartyBlob)

	_, err = client.SubmitTxBlobAndWait(counterpartyBlob, true)
	require.NoError(t, err)

	// Compute the Loan hash using the LoanBroker's LoanSequence at creation time
	loanObjectID, err := xrplhash.Loan(loanBrokerObjectID, loanBrokerLoanSequence)
	require.NoError(t, err)

	// Verify Loan object was created with correct fields
	loanObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account: borrower.GetAddress(),
		Type:    account.LoanObject,
	})
	require.NoError(t, err)

	loanObj := findLoanObject(loanObjects.AccountObjects, loanObjectID)
	require.NotNil(t, loanObj, "loan object not found")

	loanObjIndex, ok := stringFromFlatObject(loanObj, "index")
	require.True(t, ok)
	require.Equal(t, loanObjectID, loanObjIndex)

	loanPrincipalOutstanding, ok := stringFromFlatObject(loanObj, "PrincipalOutstanding")
	require.True(t, ok)
	require.Equal(t, loanSetTx.PrincipalRequested.String(), loanPrincipalOutstanding)

	loanBrokerIDField, ok := stringFromFlatObject(loanObj, "LoanBrokerID")
	require.True(t, ok)
	require.Equal(t, loanBrokerObjectID, loanBrokerIDField)

	loanBorrower, ok := stringFromFlatObject(loanObj, "Borrower")
	require.True(t, ok)
	require.Equal(t, borrower.GetAddress().String(), loanBorrower)

	loanPaymentRemaining := integration.TxFieldUint32(t, loanObj, "PaymentRemaining")
	require.Equal(t, uint32(paymentTotal), loanPaymentRemaining)

	// ========== STEP 5: Make Partial Loan Payment ==========
	loanPayTx := &transaction.LoanPay{
		BaseTx: transaction.BaseTx{
			Account: borrower.GetAddress(),
		},
		LoanID: loanObjectID,
		Amount: types.XRPCurrencyAmount(2_500_000),
	}

	flatLoanPayTx := loanPayTx.Flatten()
	_, err = runner.TestTransaction(&flatLoanPayTx, borrower, "tesSUCCESS", nil)
	require.NoError(t, err)

	// Verify loan state after partial payment
	updatedLoanObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account: borrower.GetAddress(),
		Type:    account.LoanObject,
	})
	require.NoError(t, err)

	paidLoanObj := findLoanObject(updatedLoanObjects.AccountObjects, loanObjectID)
	require.NotNil(t, paidLoanObj, "loan object not found after partial payment")

	// Principal outstanding should have decreased after payment
	paidPrincipal, ok := stringFromFlatObject(paidLoanObj, "PrincipalOutstanding")
	require.True(t, ok)
	paidPrincipalInt, err := strconv.ParseInt(paidPrincipal, 10, 64)
	require.NoError(t, err)
	originalPrincipalInt, err := strconv.ParseInt(loanPrincipalOutstanding, 10, 64)
	require.NoError(t, err)
	require.Less(t, paidPrincipalInt, originalPrincipalInt, "principal should decrease after payment")

	// PaymentRemaining should be decremented by 1
	paidPaymentRemaining := integration.TxFieldUint32(t, paidLoanObj, "PaymentRemaining")
	require.Equal(t, loanPaymentRemaining-1, paidPaymentRemaining, "payment remaining should decrease by 1")

	// Principal should not be zero yet
	require.Positive(t, paidPrincipalInt, "principal should not be zero after partial payment")

	// ManagementFeeOutstanding should be unset when fee is zero
	_, hasMgmtFee := paidLoanObj["ManagementFeeOutstanding"]
	require.False(t, hasMgmtFee, "ManagementFeeOutstanding should be absent when zero")

	// ========== STEP 5B: Make Full Loan Payment with tfLoanFullPayment ==========
	fullPaymentTx := &transaction.LoanPay{
		BaseTx: transaction.BaseTx{
			Account: borrower.GetAddress(),
		},
		LoanID: loanObjectID,
		Amount: types.XRPCurrencyAmount(paidPrincipalInt),
	}
	fullPaymentTx.SetFullPaymentFlag()

	flatFullPaymentTx := fullPaymentTx.Flatten()
	_, err = runner.TestTransaction(&flatFullPaymentTx, borrower, "tesSUCCESS", nil)
	require.NoError(t, err)

	// Verify loan state after full payment — PrincipalOutstanding and PaymentRemaining should be cleared
	finalLoanObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account: borrower.GetAddress(),
		Type:    account.LoanObject,
	})
	require.NoError(t, err)

	finalLoanObj := findLoanObject(finalLoanObjects.AccountObjects, loanObjectID)
	require.NotNil(t, finalLoanObj, "loan object not found after full payment")

	_, hasPrincipal := finalLoanObj["PrincipalOutstanding"]
	require.False(t, hasPrincipal, "PrincipalOutstanding should be absent after full payment")

	_, hasPaymentRemaining := finalLoanObj["PaymentRemaining"]
	require.False(t, hasPaymentRemaining, "PaymentRemaining should be absent after full payment")
}

// testIntegrationLendingProtocolMultiSigning tests the lending protocol lifecycle
// where the counterparty (borrower) is a multisig account. Two signers each sign the LoanSet
// as counterparty, and their signatures are combined using CombineLoanSetCounterpartySigners.
// Also covers LoanBrokerCoverDeposit, LoanBrokerCoverWithdraw, LoanManage, LoanPay, LoanDelete,
// LoanBrokerCoverClawback, and LoanBrokerDelete.
func testIntegrationLendingProtocolMultiSigning(t *testing.T, client integration.Client) {
	t.Helper()

	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 6,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	// Wallet 0: Vault Owner / Loan Broker
	vaultOwner := runner.GetWallet(0)
	loanBroker := vaultOwner
	// Wallet 1: MPT issuer
	mptIssuer := runner.GetWallet(1)
	// Wallet 2: Depositor
	depositor := runner.GetWallet(2)
	// Wallet 3: Borrower (multisig account)
	borrower := runner.GetWallet(3)
	// Wallet 4 & 5: Signers for the borrower's multisig
	signer1 := runner.GetWallet(4)
	signer2 := runner.GetWallet(5)

	// Set up SignerList on borrower account (quorum = 2, each signer weight = 1)
	signerListSetTx := &transaction.SignerListSet{
		BaseTx: transaction.BaseTx{
			Account: borrower.GetAddress(),
		},
		SignerQuorum: uint32(2),
		SignerEntries: []ledger.SignerEntryWrapper{
			{
				SignerEntry: ledger.SignerEntry{
					Account:      signer1.GetAddress(),
					SignerWeight: 1,
				},
			},
			{
				SignerEntry: ledger.SignerEntry{
					Account:      signer2.GetAddress(),
					SignerWeight: 1,
				},
			},
		},
	}

	flatSignerListTx := signerListSetTx.Flatten()
	_, err = runner.TestTransaction(&flatSignerListTx, borrower, "tesSUCCESS", nil)
	require.NoError(t, err)

	// Create the MPT issuance
	mpttoken := &transaction.MPTokenIssuanceCreate{
		BaseTx: transaction.BaseTx{
			Account: mptIssuer.GetAddress(),
		},
	}
	mpttoken.SetMPTCanTransferFlag()
	mpttoken.SetMPTCanClawbackFlag()

	mptTokenFlat := mpttoken.Flatten()
	err = client.Autofill(&mptTokenFlat)
	require.NoError(t, err)

	mptTokenBlob, _, err := mptIssuer.Sign(mptTokenFlat)
	require.NoError(t, err)

	mptTokenTxResp, err := client.SubmitTxBlobAndWait(mptTokenBlob, true)
	require.NoError(t, err)

	mptTokenIssuanceID := mptTokenTxResp.Meta.MPTIssuanceID.String()
	require.NotEmpty(t, mptTokenIssuanceID)

	// Create an MPT-collateralized vault
	vaultCreateTx := &transaction.VaultCreate{
		BaseTx: transaction.BaseTx{
			Account: vaultOwner.GetAddress(),
		},
		Asset: ledger.Asset{
			MPTIssuanceID: mptTokenIssuanceID,
		},
	}

	flatVaultCreateTx := vaultCreateTx.Flatten()
	vaultCreateResp, err := runner.TestTransaction(&flatVaultCreateTx, vaultOwner, "tesSUCCESS", nil)
	require.NoError(t, err)
	require.NotNil(t, vaultCreateResp)

	vaultCreateAccount, ok := vaultCreateResp.Tx["Account"].(string)
	require.True(t, ok)
	vaultCreateSequence := integration.TxFieldUint32(t, vaultCreateResp.Tx, "Sequence")

	vaultObjectID, err := xrplhash.Vault(vaultCreateAccount, vaultCreateSequence)
	require.NoError(t, err)

	// MPTokenAuthorize for depositor
	mptAuthorizeTx := &transaction.MPTokenAuthorize{
		BaseTx: transaction.BaseTx{
			Account: depositor.GetAddress(),
		},
		MPTokenIssuanceID: mptTokenIssuanceID,
	}
	flatMPTAuthorizeTx := mptAuthorizeTx.Flatten()
	_, err = runner.TestTransaction(&flatMPTAuthorizeTx, depositor, "tesSUCCESS", nil)
	require.NoError(t, err)

	// Fund the depositor with MPT tokens
	paymentTx := &transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account: mptIssuer.GetAddress(),
		},
		Destination: depositor.GetAddress(),
		Amount: types.MPTCurrencyAmount{
			MPTIssuanceID: mptTokenIssuanceID,
			Value:         "500000",
		},
	}
	flatPaymentTx := paymentTx.Flatten()
	_, err = runner.TestTransaction(&flatPaymentTx, mptIssuer, "tesSUCCESS", nil)
	require.NoError(t, err)

	// MPTokenAuthorize for loanBroker
	loanBrokerMptAuthorizeTx := &transaction.MPTokenAuthorize{
		BaseTx: transaction.BaseTx{
			Account: loanBroker.GetAddress(),
		},
		MPTokenIssuanceID: mptTokenIssuanceID,
	}
	flatLoanBrokerMptAuthorizeTx := loanBrokerMptAuthorizeTx.Flatten()
	_, err = runner.TestTransaction(&flatLoanBrokerMptAuthorizeTx, loanBroker, "tesSUCCESS", nil)
	require.NoError(t, err)

	// Fund the loan broker with MPT tokens
	loanBrokerPaymentTx := &transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account: mptIssuer.GetAddress(),
		},
		Destination: loanBroker.GetAddress(),
		Amount: types.MPTCurrencyAmount{
			MPTIssuanceID: mptTokenIssuanceID,
			Value:         "500000",
		},
	}
	flatLoanBrokerPaymentTx := loanBrokerPaymentTx.Flatten()
	_, err = runner.TestTransaction(&flatLoanBrokerPaymentTx, mptIssuer, "tesSUCCESS", nil)
	require.NoError(t, err)

	// Deposit MPT into vault
	vaultDepositTx := &transaction.VaultDeposit{
		BaseTx: transaction.BaseTx{
			Account: depositor.GetAddress(),
		},
		VaultID: types.Hash256(vaultObjectID),
		Amount: types.MPTCurrencyAmount{
			MPTIssuanceID: mptTokenIssuanceID,
			Value:         "200000",
		},
	}
	flatVaultDepositTx := vaultDepositTx.Flatten()
	_, err = runner.TestTransaction(&flatVaultDepositTx, depositor, "tesSUCCESS", nil)
	require.NoError(t, err)

	// ========== Create the LoanBroker ==========
	debtMaximum := types.XRPLNumber("100000")
	loanBrokerSetTx := &transaction.LoanBrokerSet{
		BaseTx: transaction.BaseTx{
			Account: loanBroker.GetAddress(),
		},
		VaultID:     vaultObjectID,
		DebtMaximum: &debtMaximum,
	}

	flatLoanBrokerTx := loanBrokerSetTx.Flatten()
	loanBrokerTxResp, err := runner.TestTransaction(&flatLoanBrokerTx, loanBroker, "tesSUCCESS", nil)
	require.NoError(t, err)
	require.NotNil(t, loanBrokerTxResp)

	loanBrokerAccount, ok := loanBrokerTxResp.Tx["Account"].(string)
	require.True(t, ok)
	loanBrokerSequence := integration.TxFieldUint32(t, loanBrokerTxResp.Tx, "Sequence")

	loanBrokerObjectID, err := xrplhash.LoanBroker(loanBrokerAccount, loanBrokerSequence)
	require.NoError(t, err)

	// Verify LoanBroker object was created with correct fields
	loanBrokerObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account: loanBroker.GetAddress(),
		Type:    account.LoanBrokerObject,
	})
	require.NoError(t, err)
	require.Len(t, loanBrokerObjects.AccountObjects, 1)

	loanBrokerObj := loanBrokerObjects.AccountObjects[0]
	loanBrokerObjIndex, ok := stringFromFlatObject(loanBrokerObj, "index")
	require.True(t, ok)
	require.Equal(t, loanBrokerObjectID, loanBrokerObjIndex)

	loanBrokerDebtMaximum, ok := stringFromFlatObject(loanBrokerObj, "DebtMaximum")
	require.True(t, ok)
	require.Equal(t, debtMaximum.String(), loanBrokerDebtMaximum)

	// Get the LoanSequence before the loan is created (used to compute the Loan hash)
	loanBrokerLoanSequence := integration.TxFieldUint32(t, loanBrokerObj, "LoanSequence")

	// ========== Create a Loan (negative test: expect temBAD_SIGNER without counterparty sig) ==========
	counterparty := borrower.GetAddress()
	paymentTotal := types.PaymentTotal(1)
	interestRate := types.InterestRate(0)
	loanSetTx := &transaction.LoanSet{
		BaseTx: transaction.BaseTx{
			Account: loanBroker.GetAddress(),
		},
		LoanBrokerID:       loanBrokerObjectID,
		PrincipalRequested: types.XRPLNumber("100000"),
		InterestRate:       &interestRate,
		Counterparty:       &counterparty,
		PaymentTotal:       &paymentTotal,
	}

	flatLoanSetTx := loanSetTx.Flatten()
	_, err = runner.TestTransaction(&flatLoanSetTx, loanBroker, "temBAD_SIGNER", nil)
	require.NoError(t, err)

	// Broker signs the LoanSet
	err = client.Autofill(&flatLoanSetTx)
	require.NoError(t, err)

	brokerBlob, _, err := loanBroker.Sign(flatLoanSetTx)
	require.NoError(t, err)

	// Each signer signs the LoanSet as counterparty multisig
	multisignOpts := &wallet.SignLoanSetByCounterpartyOptions{
		Multisign: true,
	}

	signer1Tx, signer1Blob, _, err := wallet.SignLoanSetByCounterpartyBlob(*signer1, brokerBlob, multisignOpts)
	require.NoError(t, err)
	require.NotEmpty(t, signer1Blob)

	signer2Tx, signer2Blob, _, err := wallet.SignLoanSetByCounterpartyBlob(*signer2, brokerBlob, multisignOpts)
	require.NoError(t, err)
	require.NotEmpty(t, signer2Blob)

	// Combine the two counterparty multisig signatures
	combinedTx, combinedBlob, err := wallet.CombineLoanSetCounterpartySigners([]transaction.FlatTransaction{signer1Tx, signer2Tx})
	require.NoError(t, err)
	require.NotNil(t, combinedTx)
	require.NotEmpty(t, combinedBlob)

	_, err = client.SubmitTxBlobAndWait(combinedBlob, true)
	require.NoError(t, err)

	// Compute the Loan hash using the LoanBroker's LoanSequence at creation time
	loanObjectID, err := xrplhash.Loan(loanBrokerObjectID, loanBrokerLoanSequence)
	require.NoError(t, err)

	// Verify Loan object exists with correct fields
	loanObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account: borrower.GetAddress(),
		Type:    account.LoanObject,
	})
	require.NoError(t, err)

	loanObj := findLoanObject(loanObjects.AccountObjects, loanObjectID)
	require.NotNil(t, loanObj, "loan object not found")

	loanObjIndex, ok := stringFromFlatObject(loanObj, "index")
	require.True(t, ok)
	require.Equal(t, loanObjectID, loanObjIndex)

	loanPrincipalOutstanding, ok := stringFromFlatObject(loanObj, "PrincipalOutstanding")
	require.True(t, ok)
	require.Equal(t, loanSetTx.PrincipalRequested.String(), loanPrincipalOutstanding)

	loanBrokerIDField, ok := stringFromFlatObject(loanObj, "LoanBrokerID")
	require.True(t, ok)
	require.Equal(t, loanBrokerObjectID, loanBrokerIDField)

	loanBorrowerField, ok := stringFromFlatObject(loanObj, "Borrower")
	require.True(t, ok)
	require.Equal(t, borrower.GetAddress().String(), loanBorrowerField)

	loanPaymentRemaining := integration.TxFieldUint32(t, loanObj, "PaymentRemaining")
	require.Equal(t, uint32(paymentTotal), loanPaymentRemaining)

	// ========== LoanBrokerCoverDeposit ==========
	loanBrokerCoverDepositTx := &transaction.LoanBrokerCoverDeposit{
		BaseTx: transaction.BaseTx{
			Account: loanBroker.GetAddress(),
		},
		LoanBrokerID: loanBrokerObjectID,
		Amount: types.MPTCurrencyAmount{
			MPTIssuanceID: mptTokenIssuanceID,
			Value:         "50000",
		},
	}

	flatLoanBrokerCoverDeposit := loanBrokerCoverDepositTx.Flatten()
	_, err = runner.TestTransaction(&flatLoanBrokerCoverDeposit, loanBroker, "tesSUCCESS", nil)
	require.NoError(t, err)

	// Assert LoanBroker object has updated CoverAvailable
	coverDepositResult, err := client.GetLedgerEntry(&xrplledger.EntryRequest{
		Index: loanBrokerObjectID,
	})
	require.NoError(t, err)

	coverAvailableAfterDeposit, ok := stringFromFlatObject(coverDepositResult.Node, "CoverAvailable")
	require.True(t, ok)
	require.Equal(t, (loanBrokerCoverDepositTx.Amount.(types.MPTCurrencyAmount)).Value, coverAvailableAfterDeposit)

	// ========== LoanBrokerCoverWithdraw ==========
	withdrawDestination := loanBroker.GetAddress()
	destinationTag := uint32(10)
	loanBrokerCoverWithdrawTx := &transaction.LoanBrokerCoverWithdraw{
		BaseTx: transaction.BaseTx{
			Account: loanBroker.GetAddress(),
		},
		LoanBrokerID: loanBrokerObjectID,
		Amount: types.MPTCurrencyAmount{
			MPTIssuanceID: mptTokenIssuanceID,
			Value:         "25000",
		},
		Destination:    &withdrawDestination,
		DestinationTag: &destinationTag,
	}

	flatLoanBrokerCoverWithdrawTx := loanBrokerCoverWithdrawTx.Flatten()
	_, err = runner.TestTransaction(&flatLoanBrokerCoverWithdrawTx, loanBroker, "tesSUCCESS", nil)
	require.NoError(t, err)

	// Assert LoanBroker object has updated CoverAvailable
	coverWithdrawResult, err := client.GetLedgerEntry(&xrplledger.EntryRequest{
		Index: loanBrokerObjectID,
	})
	require.NoError(t, err)

	coverAvailableAfterWithdraw, ok := stringFromFlatObject(coverWithdrawResult.Node, "CoverAvailable")
	require.True(t, ok)

	depositAmt, err := strconv.ParseInt(coverAvailableAfterDeposit, 10, 64)
	require.NoError(t, err)
	withdrawAmt, err := strconv.ParseInt((loanBrokerCoverWithdrawTx.Amount.(types.MPTCurrencyAmount)).Value, 10, 64)
	require.NoError(t, err)
	require.Equal(t, strconv.FormatInt(depositAmt-withdrawAmt, 10), coverAvailableAfterWithdraw)

	// ========== LoanManage — mark loan as impaired ==========
	loanManageTx := &transaction.LoanManage{
		BaseTx: transaction.BaseTx{
			Account: loanBroker.GetAddress(),
		},
		LoanID: loanObjectID,
	}
	loanManageTx.SetLoanImpairFlag()

	flatLoanManageTx := loanManageTx.Flatten()
	_, err = runner.TestTransaction(&flatLoanManageTx, loanBroker, "tesSUCCESS", nil)
	require.NoError(t, err)

	// Assert Loan object is impaired
	loanAfterManageObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account: borrower.GetAddress(),
		Type:    account.LoanObject,
	})
	require.NoError(t, err)

	impairedLoanObj := findLoanObject(loanAfterManageObjects.AccountObjects, loanObjectID)
	require.NotNil(t, impairedLoanObj, "loan object not found after impair")

	loanFlags := integration.TxFieldUint32(t, impairedLoanObj, "Flags")
	require.Equal(t, ledger.LsfLoanImpaired, loanFlags)

	// ========== LoanPay — full payment ==========
	loanPayTx := &transaction.LoanPay{
		BaseTx: transaction.BaseTx{
			Account: borrower.GetAddress(),
		},
		LoanID: loanObjectID,
		Amount: types.MPTCurrencyAmount{
			MPTIssuanceID: mptTokenIssuanceID,
			Value:         "100000",
		},
	}

	flatLoanPayTx := loanPayTx.Flatten()
	_, err = runner.TestTransaction(&flatLoanPayTx, borrower, "tesSUCCESS", nil)
	require.NoError(t, err)

	// Verify loan state after full payment
	loanAfterPayResult, err := client.GetLedgerEntry(&xrplledger.EntryRequest{
		Index: loanObjectID,
	})
	require.NoError(t, err)

	// Loan gets un-impaired when a payment is made
	loanFlagsAfterPay := integration.TxFieldUint32(t, loanAfterPayResult.Node, "Flags")
	require.Equal(t, uint32(0), loanFlagsAfterPay)

	// Entire loan is paid off — TotalValueOutstanding should be absent
	_, hasTotalValue := loanAfterPayResult.Node["TotalValueOutstanding"]
	require.False(t, hasTotalValue, "TotalValueOutstanding should be absent after full payment")

	// ========== LoanDelete ==========
	loanDeleteTx := &transaction.LoanDelete{
		BaseTx: transaction.BaseTx{
			Account: borrower.GetAddress(),
		},
		LoanID: loanObjectID,
	}

	flatLoanDeleteTx := loanDeleteTx.Flatten()
	_, err = runner.TestTransaction(&flatLoanDeleteTx, borrower, "tesSUCCESS", nil)
	require.NoError(t, err)

	// Assert Loan object is deleted
	loanAfterDeleteObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account: borrower.GetAddress(),
		Type:    account.LoanObject,
	})
	require.NoError(t, err)
	require.Empty(t, loanAfterDeleteObjects.AccountObjects)

	// ========== LoanBrokerCoverClawback ==========
	loanBrokerIDForClawback := types.LoanBrokerID(loanBrokerObjectID)
	loanBrokerCoverClawbackTx := &transaction.LoanBrokerCoverClawback{
		BaseTx: transaction.BaseTx{
			Account: mptIssuer.GetAddress(),
		},
		LoanBrokerID: &loanBrokerIDForClawback,
		Amount: types.MPTCurrencyAmount{
			MPTIssuanceID: mptTokenIssuanceID,
			Value:         "10000",
		},
	}

	flatLoanBrokerCoverClawbackTx := loanBrokerCoverClawbackTx.Flatten()
	_, err = runner.TestTransaction(&flatLoanBrokerCoverClawbackTx, mptIssuer, "tesSUCCESS", nil)
	require.NoError(t, err)

	// Assert LoanBroker object has updated CoverAvailable
	clawbackResult, err := client.GetLedgerEntry(&xrplledger.EntryRequest{
		Index: loanBrokerObjectID,
	})
	require.NoError(t, err)

	coverAvailableAfterClawback, ok := stringFromFlatObject(clawbackResult.Node, "CoverAvailable")
	require.True(t, ok)

	remainingCover, err := strconv.ParseInt(coverAvailableAfterWithdraw, 10, 64)
	require.NoError(t, err)
	clawbackAmt, err := strconv.ParseInt((loanBrokerCoverClawbackTx.Amount.(types.MPTCurrencyAmount)).Value, 10, 64)
	require.NoError(t, err)
	require.Equal(t, strconv.FormatInt(remainingCover-clawbackAmt, 10), coverAvailableAfterClawback)

	// ========== LoanBrokerDelete ==========
	loanBrokerDeleteTx := &transaction.LoanBrokerDelete{
		BaseTx: transaction.BaseTx{
			Account: loanBroker.GetAddress(),
		},
		LoanBrokerID: loanBrokerObjectID,
	}

	flatLoanBrokerDelete := loanBrokerDeleteTx.Flatten()
	_, err = runner.TestTransaction(&flatLoanBrokerDelete, loanBroker, "tesSUCCESS", nil)
	require.NoError(t, err)

	// Assert LoanBroker object is deleted
	loanBrokerAfterDeleteObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account: loanBroker.GetAddress(),
		Type:    account.LoanBrokerObject,
	})
	require.NoError(t, err)
	require.Empty(t, loanBrokerAfterDeleteObjects.AccountObjects)
}

func TestIntegrationLendingProtocolSingleSigning_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationLendingProtocolSingleSigning(t, client)
}

func TestIntegrationLendingProtocolSingleSigning_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationLendingProtocolSingleSigning(t, client)
}

func TestIntegrationLendingProtocolMultiSigning_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationLendingProtocolMultiSigning(t, client)
}

func TestIntegrationLendingProtocolMultiSigning_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationLendingProtocolMultiSigning(t, client)
}
