package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	account "github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
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

	// A present error field must be a string. Treat null and other values as
	// malformed responses instead of accepting proxy-modified success payloads.
	if errorValue, ok := jr.Result["error"]; ok {
		errorString, ok := errorValue.(string)
		if !ok {
			return jr, fmt.Errorf("%w: got %T", ErrResponseErrorFieldIsNotAString, errorValue)
		}
		return jr, &ClientError{ErrorString: errorString}
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
func (c *Client) setTransactionNextValidSequenceNumber(
	ctx context.Context,
	tx *transaction.FlatTransaction,
) error {
	if _, ok := (*tx)["Account"].(string); !ok {
		return ErrMissingAccountInTransaction
	}
	var res account.InfoResponse
	if err := c.requestResult(ctx, &account.InfoRequest{
		Account:     types.Address((*tx)["Account"].(string)),
		LedgerIndex: common.LedgerTitle("current"),
	}, &res); err != nil {
		return err
	}

	(*tx)["Sequence"] = uint32(res.AccountData.Sequence)
	return nil
}

// feeRequest adapts the client transport to the shared fee helpers.
func (c *Client) feeRequest() clientinternal.RequestResultFunc {
	return func(ctx context.Context, req clientinternal.Request, result any) error {
		return c.requestResult(ctx, req, result)
	}
}

func (c *Client) feeSettings() clientinternal.FeeSettings {
	return clientinternal.FeeSettings{Cushion: c.cfg.feeCushion, MaxFeeXRP: c.cfg.maxFeeXRP}
}

// calculateFeePerTransactionType calculates the fee for a transaction,
// including special costs for EscrowFinish, owner-reserve transactions, Batch,
// confidential MPT transactions, LoanSet, and multisigning.
func (c *Client) calculateFeePerTransactionType(
	ctx context.Context,
	tx *transaction.FlatTransaction,
	nSigners uint64,
) error {
	_, err := clientinternal.CalculateFee(ctx, c.feeRequest(), tx, nSigners, c.feeSettings())
	return err
}

// Sets the latest validated ledger sequence for the transaction.
// Modifies the `LastLedgerSequence` field in the tx.
func (c *Client) setLastLedgerSequence(ctx context.Context, tx *transaction.FlatTransaction) error {
	var response ledger.Response
	if err := c.requestResult(ctx, &ledger.Request{
		LedgerIndex: common.LedgerTitle("validated"),
	}, &response); err != nil {
		return err
	}

	(*tx)["LastLedgerSequence"] = response.LedgerIndex.Uint32() + commonconstants.LedgerOffset
	return nil
}

// Checks for any blockers that prevent the deletion of an account.
// Returns nil if there are no blockers, otherwise returns an error.
func (c *Client) checkAccountDeleteBlockers(ctx context.Context, address types.Address) error {
	var accObjects account.ObjectsResponse
	if err := c.requestResult(ctx, &account.ObjectsRequest{
		Account:              address,
		LedgerIndex:          common.LedgerTitle("validated"),
		DeletionBlockersOnly: true,
	}, &accObjects); err != nil {
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
	return clientinternal.NormalizeDeliverMax(*tx)
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
//
// Even when autofill is disabled, this client submission path needs a discovered
// network identity, or trusted values from WithNetworkIdentity, before it signs.
// Call wallet.Sign directly when signing must be fully offline.
func (c *Client) getSignedTx(
	ctx context.Context,
	tx transaction.FlatTransaction,
	autofill bool,
	wallet *wallet.Wallet,
) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	working := transaction.FlatTransaction(clientinternal.CloneTransaction(tx))
	if working == nil {
		return "", ErrNilTransaction
	}
	if err := c.checkPaymentAmounts(&working); err != nil {
		return "", err
	}

	signingType, err := clientinternal.InspectSignedTransaction(working, false)
	if err != nil {
		return "", err
	}
	if signingType != clientinternal.UnsignedTransaction {
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
		if err := c.autofill(ctx, &working, 0); err != nil {
			return "", err
		}
	} else {
		identity, err := c.ensureNetworkIdentity(ctx)
		if err != nil {
			return "", err
		}
		if err := clientinternal.ApplyNetworkIDPolicy(working, identity); err != nil {
			return "", err
		}
	}

	if err := ctx.Err(); err != nil {
		return "", err
	}
	txBlob, _, err := wallet.Sign(working)
	if err != nil {
		return "", err
	}
	return txBlob, nil
}
