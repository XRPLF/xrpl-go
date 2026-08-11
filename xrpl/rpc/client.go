// Package rpc provides RPC client functionality for interacting with XRPL servers.
//
// A Client discovers network identity before its first identity-dependent
// operation and caches the first successful discovery for the client lifetime.
// A failed discovery is returned and a later operation retries it. A trusted
// identity configured with WithNetworkIdentity bypasses discovery only when its
// build version is non-empty.
package rpc

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	commonconstants "github.com/Peersyst/xrpl-go/xrpl/common"
	"github.com/Peersyst/xrpl-go/xrpl/hash"
	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	requests "github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	rpctypes "github.com/Peersyst/xrpl-go/xrpl/rpc/types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"

	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

var (
	fundWalletMaxAttempts  = 20
	fundWalletPollInterval = 1 * time.Second
)

// maxDrainBytes caps how much of an error response body is drained before
// retrying. A small bounded drain lets the HTTP transport reuse the
// connection via keep-alive when the body fits, and prevents a hostile
// upstream from forcing an unbounded read when it doesn't.
const maxDrainBytes = 4 << 10 // 4 KiB

// Client is an XRPL RPC client for sending requests and managing transactions.
type Client struct {
	cfg *Config

	networkID    *uint32
	buildVersion string

	identity networkIdentityState
}

// NewClient creates a new RPC Client with the given configuration.
func NewClient(cfg *Config) *Client {
	networkID := clientinternal.CloneNetworkID(cfg.networkID)
	return &Client{
		cfg:          cfg,
		networkID:    networkID,
		buildVersion: cfg.buildVersion,
		identity: networkIdentityState{
			ready: networkID != nil && cfg.buildVersion != "",
		},
	}
}

// Request sends a request to the XRPL server and returns the response and any error encountered.
func (c *Client) Request(reqParams XRPLRequest) (XRPLResponse, error) {
	return c.request(context.Background(), reqParams)
}

func (c *Client) request(ctx context.Context, reqParams XRPLRequest) (XRPLResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := reqParams.Validate(); err != nil {
		return nil, err
	}

	body, err := createRequest(reqParams)
	if err != nil {
		return nil, err
	}

	const maxAttempts = 4 // 1 initial attempt + 3 retries
	backoffDuration := c.cfg.retryDelay

	var (
		response   *http.Response
		cancelFunc context.CancelFunc
	)

	// cfg.timeout bounds a single attempt, not the full retry window.
	for attempt := range maxAttempts {
		requestCtx, cancel := context.WithTimeout(ctx, c.cfg.timeout)

		req, err := http.NewRequestWithContext(
			requestCtx,
			http.MethodPost,
			c.cfg.URL,
			bytes.NewReader(body),
		)
		if err != nil {
			cancel()
			return nil, err
		}

		req.Header = c.cfg.Headers

		response, err = c.cfg.HTTPClient.Do(req)
		if err != nil {
			// net/http documents response as nil when Do returns an error, but
			// custom HTTPClient impls may not follow that contract.
			if response != nil {
				_ = response.Body.Close()
			}
			cancel()
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, err
		}

		// HTTPClient is an interface, custom impls may return (nil, nil),
		// violating net/http's contract. Standard *http.Client never hits
		// this branch.
		if response == nil {
			cancel()
			return nil, &ClientError{ErrorString: "nil response from server"}
		}

		if response.StatusCode != http.StatusServiceUnavailable {
			cancelFunc = cancel
			break
		}

		// Drain and close the 503 response body before retrying so the connection
		// can be reused by the HTTP client.
		_, _ = io.CopyN(io.Discard, response.Body, maxDrainBytes)
		_ = response.Body.Close()
		cancel()

		if attempt == maxAttempts-1 {
			return nil, &ClientError{ErrorString: "Server is overloaded, rate limit exceeded"}
		}

		if err := clientinternal.Wait(ctx, backoffDuration); err != nil {
			return nil, err
		}
		backoffDuration *= 2
	}
	defer cancelFunc()
	defer func() {
		_ = response.Body.Close()
	}()

	var jr Response
	jr, err = checkForError(response, c.cfg.maxResponseSize)
	if err != nil {
		return nil, err
	}

	return &jr, nil
}

func (c *Client) requestResult(ctx context.Context, req XRPLRequest, result any) error {
	response, err := c.request(ctx, req)
	if err != nil {
		return err
	}
	return response.GetResult(result)
}

// SubmitTxBlob sends a pre-signed transaction blob to the server.
// Its preflight validates only the structure of signing fields. rippled remains
// authoritative for cryptographic signature validity. AccountDelete always uses
// fail_hard as required by reliable-submission safety guidance.
func (c *Client) SubmitTxBlob(txBlob string, failHard bool) (*requests.SubmitResponse, error) {
	tx, err := clientinternal.DecodeTransactionBlob(txBlob)
	if err != nil {
		return nil, err
	}
	return c.submitTxBlob(context.Background(), txBlob, tx, failHard)
}

func (c *Client) submitTxBlob(
	ctx context.Context,
	txBlob string,
	tx map[string]any,
	failHard bool,
) (*requests.SubmitResponse, error) {
	signingType, err := clientinternal.InspectSignedTransaction(tx, false)
	if err != nil {
		return nil, err
	}
	if err := clientinternal.InspectSignedBatchInners(tx); err != nil {
		return nil, err
	}
	if signingType == clientinternal.UnsignedTransaction {
		return nil, ErrMissingTxSignatureOrSigningPubKey
	}

	return c.submitRequest(ctx, &requests.SubmitRequest{
		TxBlob:   txBlob,
		FailHard: clientinternal.SubmissionFailHard(tx, failHard),
	})
}

// SubmitTxBlobAndWait submits a pre-signed transaction and waits for an
// authoritative validated-ledger result. LastLedgerSequence is required and
// expiry occurs only after the validated ledger passes it.
func (c *Client) SubmitTxBlobAndWait(txBlob string, failHard bool) (*requests.TxResponse, error) {
	return c.SubmitTxBlobAndWaitContext(context.Background(), txBlob, failHard)
}

// SubmitTxBlobAndWaitContext is SubmitTxBlobAndWait with caller cancellation.
// Context cancellation is returned as ctx.Err and is never reported as a
// transaction failure or expiry.
func (c *Client) SubmitTxBlobAndWaitContext(
	ctx context.Context,
	txBlob string,
	failHard bool,
) (*requests.TxResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := clientinternal.ValidateFinalityMonitoring(c.cfg.retryDelay, c.cfg.maxRetries); err != nil {
		return nil, err
	}
	return c.submitTxBlobAndWait(ctx, txBlob, failHard)
}

func (c *Client) submitTxBlobAndWait(
	ctx context.Context,
	txBlob string,
	failHard bool,
) (*requests.TxResponse, error) {
	tx, err := clientinternal.DecodeTransactionBlob(txBlob)
	if err != nil {
		return nil, err
	}

	lastLedgerSequence, ok := tx["LastLedgerSequence"].(uint32)
	if !ok {
		return nil, ErrMissingLastLedgerSequenceInTransaction
	}
	if err := clientinternal.ValidateLastLedgerSequence(lastLedgerSequence); err != nil {
		return nil, err
	}

	submitResponse, err := c.submitTxBlob(ctx, txBlob, tx, failHard)
	if err != nil {
		return nil, err
	}
	if err := clientinternal.ValidatePreliminaryResult(
		submitResponse.EngineResult,
		submitResponse.EngineResultMessage,
	); err != nil {
		return nil, err
	}

	txHash, err := hash.SignTx(tx)
	if err != nil {
		return nil, err
	}

	return c.waitForTransaction(
		ctx,
		txHash,
		lastLedgerSequence,
		submitResponse.EngineResult,
	)
}

// SubmitTx signs the transaction (if necessary) and submits it to the server
// via a submission request. It applies the provided submit options to decide whether
// to autofill missing fields and enforce failHard mode during submission.
func (c *Client) SubmitTx(tx transaction.FlatTransaction, opts *rpctypes.SubmitOptions) (*requests.SubmitResponse, error) {
	if opts == nil {
		opts = &rpctypes.SubmitOptions{}
	}
	txBlob, err := c.getSignedTx(context.Background(), tx, opts.Autofill, opts.Wallet)
	if err != nil {
		return nil, err
	}

	return c.SubmitTxBlob(txBlob, opts.FailHard)
}

// SubmitTxAndWait prepares, submits, and monitors a transaction until its
// validated-ledger outcome is authoritative.
func (c *Client) SubmitTxAndWait(tx transaction.FlatTransaction, opts *rpctypes.SubmitOptions) (*requests.TxResponse, error) {
	return c.SubmitTxAndWaitContext(context.Background(), tx, opts)
}

// SubmitTxAndWaitContext is SubmitTxAndWait with caller cancellation for
// transaction preparation, submission, and finality monitoring.
func (c *Client) SubmitTxAndWaitContext(
	ctx context.Context,
	tx transaction.FlatTransaction,
	opts *rpctypes.SubmitOptions,
) (*requests.TxResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := clientinternal.ValidateFinalityMonitoring(c.cfg.retryDelay, c.cfg.maxRetries); err != nil {
		return nil, err
	}
	if opts == nil {
		opts = &rpctypes.SubmitOptions{}
	}
	txBlob, err := c.getSignedTx(ctx, tx, opts.Autofill, opts.Wallet)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return c.submitTxBlobAndWait(ctx, txBlob, opts.FailHard)
}

// SubmitMultisigned submits a structurally complete multisigned transaction blob.
// rippled remains authoritative for cryptographic signature validity.
func (c *Client) SubmitMultisigned(txBlob string, failHard bool) (*requests.SubmitMultisignedResponse, error) {
	tx, err := clientinternal.DecodeTransactionBlob(txBlob)
	if err != nil {
		return nil, err
	}
	signingType, err := clientinternal.InspectSignedTransaction(tx, false)
	if err != nil {
		return nil, err
	}
	if err := clientinternal.InspectSignedBatchInners(tx); err != nil {
		return nil, err
	}
	if signingType != clientinternal.MultiSignedTransaction {
		return nil, ErrTransactionNotMultisigned
	}

	return c.submitMultisignedRequest(&requests.SubmitMultisignedRequest{
		Tx:       tx,
		FailHard: clientinternal.SubmissionFailHard(tx, failHard),
	})
}

// Autofill fills missing fields in a transaction. It commits all changes to the
// caller's map only after autofill succeeds. If autofill returns an error, the
// caller's map is unchanged.
func (c *Client) Autofill(tx *transaction.FlatTransaction) error {
	if tx == nil || *tx == nil {
		return ErrNilTransaction
	}
	working := transaction.FlatTransaction(clientinternal.CloneTransaction(*tx))
	if err := c.autofill(context.Background(), &working, 0); err != nil {
		return err
	}
	clientinternal.ReplaceTransactionContents(*tx, working)
	return nil
}

func (c *Client) autofill(ctx context.Context, tx *transaction.FlatTransaction, nSigners uint64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := tx.RequireTransactionType(); err != nil {
		return err
	}
	if err := tx.NormalizeFlags(); err != nil {
		return err
	}
	if err := c.checkPaymentAmounts(tx); err != nil {
		return err
	}

	identity, err := c.ensureNetworkIdentity(ctx)
	if err != nil {
		return err
	}
	if err := c.setValidTransactionAddresses(tx); err != nil {
		return err
	}
	if err := clientinternal.ApplyNetworkIDPolicy(*tx, identity); err != nil {
		return err
	}
	if _, ok := (*tx)["Sequence"]; !ok {
		if err := c.setTransactionNextValidSequenceNumber(ctx, tx); err != nil {
			return err
		}
	}
	if _, ok := (*tx)["Fee"]; !ok {
		if err := c.calculateFeePerTransactionType(ctx, tx, nSigners); err != nil {
			return err
		}
	}
	if _, ok := (*tx)["LastLedgerSequence"]; !ok {
		if err := c.setLastLedgerSequence(ctx, tx); err != nil {
			return err
		}
	}
	txType := tx.TxType()
	if txType == transaction.AccountDeleteTx {
		accountAddress, ok := typecheck.ToString((*tx)["Account"])
		if !ok {
			return ErrMissingAccountInTransaction
		}
		if err := c.checkAccountDeleteBlockers(ctx, types.Address(accountAddress)); err != nil {
			return err
		}
	}
	if txType == transaction.BatchTx {
		if err := c.autofillRawTransactions(ctx, tx); err != nil {
			return err
		}
	}
	return nil
}

// AutofillMultisigned fills in the missing fields in a multisigned transaction.
// This function is used to fill in the missing fields in a multisigned transaction.
// It fills in the missing fields in the transaction and calculates the fee per number of signers.
func (c *Client) AutofillMultisigned(tx *transaction.FlatTransaction, nSigners uint64) error {
	if tx == nil || *tx == nil {
		return ErrNilTransaction
	}
	working := transaction.FlatTransaction(clientinternal.CloneTransaction(*tx))
	if err := c.autofill(context.Background(), &working, nSigners); err != nil {
		return err
	}
	clientinternal.ReplaceTransactionContents(*tx, working)
	return nil
}

// FaucetProvider returns the faucet provider for the client.
func (c *Client) FaucetProvider() commonconstants.FaucetProvider {
	return c.cfg.faucetProvider
}

// FundWallet funds a wallet with the client's faucet provider and polls the
// validated ledger until the account's balance increases. It returns
// ErrFundWalletBalanceNotUpdated if the balance fails to update within the
// poll window.
func (c *Client) FundWallet(wallet *wallet.Wallet) error {
	if wallet.ClassicAddress == "" {
		return ErrCannotFundWalletWithoutClassicAddress
	}

	// Starting balance. An error here (typically actNotFound for a
	// brand-new account) is treated as a zero balance so polling can still
	// detect the faucet deposit.
	startBalance, err := c.getXrpDropsBalance(wallet.ClassicAddress, common.Validated)
	if err != nil && !isFundWalletActNotFound(err) {
		return err
	}

	if err := c.cfg.faucetProvider.FundWallet(wallet.ClassicAddress); err != nil {
		return err
	}

	for range fundWalletMaxAttempts {
		time.Sleep(fundWalletPollInterval)
		balance, err := c.getXrpDropsBalance(wallet.ClassicAddress, common.Validated)
		if err != nil {
			if isFundWalletActNotFound(err) {
				continue
			}
			return err
		}
		if balance > startBalance {
			return nil
		}
	}

	return ErrFundWalletBalanceNotUpdated
}

func isFundWalletActNotFound(err error) bool {
	var clientErr *ClientError
	return errors.As(err, &clientErr) && clientErr.ErrorString == actNotFound
}

func (c *Client) autofillRawTransactions(
	ctx context.Context,
	tx *transaction.FlatTransaction,
) error {
	return clientinternal.AutofillBatchRawTransactions(*tx, func(accountAddress string) (uint32, error) {
		var accountInfo account.InfoResponse
		if err := c.requestResult(ctx, &account.InfoRequest{
			Account: types.Address(accountAddress),
		}, &accountInfo); err != nil {
			return 0, err
		}
		return accountInfo.AccountData.Sequence, nil
	})
}
