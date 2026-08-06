package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"strings"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	account "github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	server "github.com/Peersyst/xrpl-go/xrpl/queries/server"
	requests "github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"

	jsoniter "github.com/json-iterator/go"

	commonconstants "github.com/Peersyst/xrpl-go/xrpl/common"
	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
)

const (
	// RestrictedNetworks is the largest network ID for which transactions omit NetworkID.
	RestrictedNetworks = clientinternal.RestrictedNetworks
	// RequiredNetworkIDVersion is the first rippled version that enforces NetworkID.
	RequiredNetworkIDVersion = clientinternal.RequiredNetworkIDVersion
)

// CreateRequest formats the parameters and method name ready for sending request
// Params will have been serialised if required and added to request struct before being passed to this method
func createRequest(reqParams XRPLRequest) ([]byte, error) {
	var body Request

	reqParams.SetAPIVersion(
		reqParams.APIVersion(),
	)

	body = Request{
		Method: reqParams.Method(),
		// each param object will have a struct with json serialising tags
		Params: [1]any{reqParams},
	}

	// Omit the Params field if method doesn't require any
	paramBytes, err := jsoniter.Marshal(body.Params)
	if err != nil {
		return nil, err
	}
	paramString := string(paramBytes)
	if strings.Compare(paramString, "[{}]") == 0 {
		// need to remove params field from the body if it is empty
		body = Request{
			Method: reqParams.Method(),
		}

		jsonBytes, err := jsoniter.Marshal(body)
		if err != nil {
			return nil, err
		}

		return jsonBytes, nil
	}

	jsonBytes, err := jsoniter.Marshal(body)
	if err != nil {
		return nil, ErrFailedToMarshalJSONRPCRequest{
			Method: reqParams.Method(),
			Params: reqParams,
			Err:    err,
		}
	}

	return jsonBytes, nil
}

// checkForError reads the http response and formats the error if it exists
func checkForError(res *http.Response, maxResponseSize int64) (Response, error) {
	var jr Response

	b, err := readResponseBody(res.Body, maxResponseSize)
	if err != nil || b == nil {
		return jr, err
	}

	// In case a different error code is returned
	if res.StatusCode != 200 {
		return jr, &ClientError{ErrorString: string(b)}
	}

	jDec := json.NewDecoder(bytes.NewReader(b))
	jDec.UseNumber()
	err = jDec.Decode(&jr)
	if err != nil {
		return jr, err
	}

	// result will have 'error' if error response
	if _, ok := jr.Result["error"]; ok {
		return jr, &ClientError{ErrorString: jr.Result["error"].(string)}
	}

	return jr, nil
}

func readResponseBody(body io.Reader, maxResponseSize int64) ([]byte, error) {
	if maxResponseSize == 0 {
		return io.ReadAll(body)
	}
	if maxResponseSize < 0 {
		maxResponseSize = defaultMaxResponseSize
	}

	limit := maxResponseSize
	if maxResponseSize < math.MaxInt64 {
		limit++
	}

	b, err := io.ReadAll(io.LimitReader(body, limit))
	if err != nil {
		return nil, err
	}
	// Deliberately do not drain the remaining body on oversize: closing without
	// draining costs one TCP reconnect, but draining would defeat the memory cap.
	if int64(len(b)) > maxResponseSize {
		return nil, ErrResponseTooLarge
	}

	return b, nil
}

// setValidTransactionAddresses applies the shared X-address and tag policy.
func (c *Client) setValidTransactionAddresses(tx *transaction.FlatTransaction) error {
	return clientinternal.SetValidAddresses(*tx)
}

// Sets the next valid sequence number for a given transaction.
func (c *Client) setTransactionNextValidSequenceNumber(tx *transaction.FlatTransaction) error {
	if _, ok := (*tx)["Account"].(string); !ok {
		return ErrMissingAccountInTransaction
	}
	res, err := c.GetAccountInfo(&account.InfoRequest{
		Account:     types.Address((*tx)["Account"].(string)),
		LedgerIndex: common.LedgerTitle("current"),
	})
	if err != nil {
		return err
	}

	(*tx)["Sequence"] = uint32(res.AccountData.Sequence)
	return nil
}

// Calculates the current transaction fee for the ledger.
func (c *Client) getFeeXrp(cushion float64) (string, error) {
	res, err := c.GetServerInfo(&server.InfoRequest{})
	if err != nil {
		return "", err
	}

	baseFeeXRP := res.Info.ValidatedLedger.BaseFeeXRP
	if baseFeeXRP == nil {
		return "", ErrCouldNotGetBaseFeeXrp
	}

	return clientinternal.NetworkFeeXRP(
		*baseFeeXRP,
		res.Info.LoadFactor,
		cushion,
		c.cfg.maxFeeXRP,
	)
}

// calculateFeePerTransactionType calculates the fee for a transaction,
// including special costs for EscrowFinish, owner-reserve transactions, Batch,
// LoanSet, and multisigning.
func (c *Client) calculateFeePerTransactionType(tx *transaction.FlatTransaction, nSigners uint64) error {
	netFeeXRP, err := c.getFeeXrp(c.cfg.feeCushion)
	if err != nil {
		return err
	}

	netFee, err := clientinternal.NewFeeFromXRP(netFeeXRP)
	if err != nil {
		return err
	}
	baseFee := netFee

	transactionType, _ := (*tx)["TransactionType"].(string)
	isSpecialTxCost := transactionType == "AccountDelete" || transactionType == "AMMCreate" || transactionType == "VaultCreate"

	switch transactionType {
	case "EscrowFinish":
		if fulfillment, ok := (*tx)["Fulfillment"].(string); ok {
			fulfillmentBytesSize := (len(fulfillment) + 1) / 2
			baseFee, err = netFee.MultiplyFraction(33*16+uint64(fulfillmentBytesSize), 16)
			if err != nil {
				return err
			}
		}
	case "AccountDelete", "AMMCreate", "VaultCreate":
		reserveFee, reserveErr := c.fetchOwnerReserveFee()
		if reserveErr != nil {
			return reserveErr
		}
		baseFee = clientinternal.NewFeeFromUint64(reserveFee)
	case "Batch":
		rawTxFees, batchErr := c.calculateBatchFees(tx)
		if batchErr != nil {
			return batchErr
		}
		baseFee = netFee.Multiply(2).Add(rawTxFees)
	case "LoanSet":
		counterPartySignersCount, signerErr := c.fetchCounterPartySignersCount(*tx)
		if signerErr != nil {
			return signerErr
		}
		baseFee = netFee.Multiply(1 + counterPartySignersCount)
	}

	if nSigners > 0 {
		baseFee = baseFee.Add(netFee.Multiply(nSigners))
	}

	maxFee, err := clientinternal.NewFeeFromXRP(c.cfg.maxFeeXRP)
	if err != nil {
		return err
	}
	totalFee := baseFee
	if !isSpecialTxCost {
		totalFee = baseFee.Min(maxFee)
	}

	(*tx)["Fee"] = totalFee.CeilDrops()
	return nil
}

// Sets the latest validated ledger sequence for the transaction.
// Modifies the `LastLedgerSequence` field in the tx.
func (c *Client) setLastLedgerSequence(tx *transaction.FlatTransaction) error {
	index, err := c.GetLedgerIndex()
	if err != nil {
		return err
	}

	(*tx)["LastLedgerSequence"] = index.Uint32() + commonconstants.LedgerOffset
	return err
}

// Checks for any blockers that prevent the deletion of an account.
// Returns nil if there are no blockers, otherwise returns an error.
func (c *Client) checkAccountDeleteBlockers(address types.Address) error {
	accObjects, err := c.GetAccountObjects(&account.ObjectsRequest{
		Account:              address,
		LedgerIndex:          common.LedgerTitle("validated"),
		DeletionBlockersOnly: true,
	})
	if err != nil {
		return err
	}

	if len(accObjects.AccountObjects) > 0 {
		return ErrAccountCannotBeDeleted
	}
	return nil
}

func (c *Client) checkPaymentAmounts(tx *transaction.FlatTransaction) error {
	if tx.TxType() != transaction.PaymentTx {
		return nil
	}
	if !clientinternal.NormalizeDeliverMax(*tx) {
		return ErrAmountAndDeliverMaxMustBeIdentical
	}
	return nil
}

func (c *Client) submitMultisignedRequest(req *requests.SubmitMultisignedRequest) (*requests.SubmitMultisignedResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var subRes requests.SubmitMultisignedResponse
	err = res.GetResult(&subRes)
	if err != nil {
		return nil, err
	}
	return &subRes, nil
}

func (c *Client) submitRequest(
	ctx context.Context,
	req *requests.SubmitRequest,
) (*requests.SubmitResponse, error) {
	res, err := c.request(ctx, req)
	if err != nil {
		return nil, err
	}
	var subRes requests.SubmitResponse
	if err := res.GetResult(&subRes); err != nil {
		return nil, err
	}
	return &subRes, nil
}

func (c *Client) waitForTransaction(
	ctx context.Context,
	txHash string,
	lastLedgerSequence uint32,
	preliminaryResult string,
) (*requests.TxResponse, error) {
	return clientinternal.WaitForFinality(
		ctx,
		clientinternal.FinalityConfig{
			LastLedgerSequence: lastLedgerSequence,
			PreliminaryResult:  preliminaryResult,
			PollInterval:       c.cfg.retryDelay,
			MaxAttempts:        c.cfg.maxRetries,
		},
		clientinternal.TxFinalityHooks(
			func(ctx context.Context) (clientinternal.ResponseDecoder, error) {
				return c.request(ctx, &requests.TxRequest{Transaction: txHash})
			},
			func(ctx context.Context) (clientinternal.ResponseDecoder, error) {
				return c.request(ctx, &ledger.Request{LedgerIndex: common.Validated})
			},
			isTransactionNotFoundError,
		),
	)
}

func isTransactionNotFoundError(err error) bool {
	var clientErr *ClientError
	return errors.As(err, &clientErr) && clientErr.ErrorString == txnNotFound
}

// getSignedTx ensures the transaction is fully signed and returns the transaction blob.
// Submission works on a deep copy, so autofill, address conversion, NetworkID policy,
// and DeliverMax normalization never mutate the caller-owned transaction map.
func (c *Client) getSignedTx(tx transaction.FlatTransaction, autofill bool, wallet *wallet.Wallet) (string, error) {
	working := transaction.FlatTransaction(clientinternal.CloneTransaction(tx))
	if working == nil {
		return "", ErrNilTransaction
	}
	if err := c.checkPaymentAmounts(&working); err != nil {
		return "", err
	}

	form, err := clientinternal.InspectSignedTransaction(working, false)
	if err != nil {
		return "", err
	}
	if form != clientinternal.UnsignedTransaction {
		blob, err := binarycodec.Encode(working)
		if err != nil {
			return "", err
		}
		return blob, nil
	}

	if wallet == nil {
		return "", ErrMissingWallet
	}

	if autofill {
		// working is already a private deep copy, so the unexported worker is
		// enough. The public Autofill wrapper would clone it a second time.
		if err := c.autofill(&working); err != nil {
			return "", err
		}
	} else {
		identity, err := c.ensureNetworkIdentity()
		if err != nil {
			return "", err
		}
		if err := clientinternal.ApplyNetworkIDPolicy(working, identity); err != nil {
			return "", err
		}
	}

	txBlob, _, err := wallet.Sign(working)
	if err != nil {
		return "", err
	}
	return txBlob, nil
}

// fetchOwnerReserveFee fetches the owner reserve fee from the server state.
func (c *Client) fetchOwnerReserveFee() (uint64, error) {
	response, err := c.GetServerState(&server.StateRequest{})
	if err != nil {
		return 0, err
	}

	reserveInc := response.State.ValidatedLedger.ReserveInc
	if reserveInc == nil {
		return 0, ErrCouldNotFetchOwnerReserve
	}

	return uint64(*reserveInc), nil
}

// fetchCounterPartySignersCount fetches the number of signers for the counterparty account.
// For LoanSet transactions, if Counterparty is not provided, it fetches the LoanBroker and uses its Owner.
// Returns the number of signers in the counterparty's signer list, or 1 if no signer list exists.
func (c *Client) fetchCounterPartySignersCount(tx transaction.FlatTransaction) (uint64, error) {
	var counterparty types.Address

	// Extract Counterparty from transaction if present
	if cp, ok := tx["Counterparty"]; ok {
		if cpStr, ok := cp.(string); ok && cpStr != "" {
			counterparty = types.Address(cpStr)
		}
	}

	// If Counterparty is not provided and transaction has LoanBrokerID, fetch LoanBroker
	if counterparty == "" {
		loanBrokerID, ok := tx["LoanBrokerID"].(string)
		if !ok || loanBrokerID == "" {
			return 0, ErrLoanBrokerIDRequired
		}

		// Make ledger_entry request
		res, err := c.GetLedgerEntry(&ledger.EntryRequest{
			Index:       loanBrokerID,
			LedgerIndex: common.LedgerTitle("validated"),
		})
		if err != nil {
			return 0, err
		}

		// Extract Owner from the LoanBroker FlatLedgerObject
		owner, ok := res.Node["Owner"].(string)
		if !ok || owner == "" {
			return 0, ErrCouldNotFetchLoanBrokerOwner
		}
		counterparty = types.Address(owner)
	}

	if counterparty == "" {
		return 0, ErrCounterpartyRequired
	}

	// Fetch account info with signer lists
	accountInfo, err := c.GetAccountInfo(&account.InfoRequest{
		Account:     counterparty,
		LedgerIndex: common.LedgerTitle("validated"),
		SignerLists: true,
	})
	if err != nil {
		return 0, err
	}

	// Extract the first signer list's SignerEntries length
	if len(accountInfo.SignerLists) > 0 {
		return uint64(len(accountInfo.SignerLists[0].SignerEntries)), nil
	}

	// Default to 1 if no signer list exists
	return 1, nil
}

// calculateBatchFees calculates the total fees for all inner transactions in a Batch.
func (c *Client) calculateBatchFees(tx *transaction.FlatTransaction) (*clientinternal.Fee, error) {
	totalFees := clientinternal.NewFeeFromUint64(0)

	// Get RawTransactions from the batch transaction
	rawTransactions, ok := (*tx)["RawTransactions"].([]map[string]any)
	if !ok {
		return nil, ErrRawTransactionsFieldMissing
	}

	// Iterate through each raw transaction
	for _, rawTx := range rawTransactions {
		// Extract the actual transaction from the wrapper
		innerTx, ok := rawTx["RawTransaction"].(map[string]any)
		if !ok {
			return nil, ErrRawTransactionFieldMissing
		}

		// Calculate fee for this inner transaction (no multi-signing for inner transactions)
		innerTxFlat := transaction.FlatTransaction(innerTx)
		err := c.calculateFeePerTransactionType(&innerTxFlat, 0)
		if err != nil {
			return nil, err
		}

		// Extract the calculated fee
		feeStr, ok := innerTx["Fee"].(string)
		if !ok {
			return nil, ErrFeeFieldMissing
		}

		innerTx["Fee"] = "0"

		innerFee, err := clientinternal.NewFeeFromDrops(feeStr)
		if err != nil {
			return nil, ErrFailedToParseFee{
				Fee: feeStr,
				Err: err,
			}
		}

		totalFees = totalFees.Add(innerFee)
	}

	return totalFees, nil
}
