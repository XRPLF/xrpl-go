// Package rpc provides RPC client functionality for interacting with XRPL servers.
package rpc

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	commonconstants "github.com/Peersyst/xrpl-go/xrpl/common"
	"github.com/Peersyst/xrpl-go/xrpl/hash"
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

	NetworkID uint32
}

// NewClient creates a new RPC Client with the given configuration.
func NewClient(cfg *Config) *Client {
	return &Client{
		cfg: cfg,
	}
}

// Request sends a request to the XRPL server and returns the response and any error encountered.
func (c *Client) Request(reqParams XRPLRequest) (XRPLResponse, error) {
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
		ctx, cancel := context.WithTimeout(context.Background(), c.cfg.timeout)

		req, err := http.NewRequestWithContext(
			ctx,
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

		// time.Sleep is non-cancellable. If Request ever accepts a caller
		// context, switch to a select on ctx.Done() / time.After.
		time.Sleep(backoffDuration)
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

// SubmitTxBlob sends a pre-signed transaction blob to the server.
// It decodes the blob to confirm that it contains either a signature
// or a signing public key, and then submits it using a submission request.
// The failHard flag determines how strictly errors are handled.
func (c *Client) SubmitTxBlob(txBlob string, failHard bool) (*requests.SubmitResponse, error) {
	tx, err := binarycodec.Decode(txBlob)
	if err != nil {
		return nil, err
	}

	_, okTxSig := tx["TxSignature"].(string)
	_, okPubKey := tx["SigningPubKey"].(string)

	if !okTxSig && !okPubKey {
		return nil, ErrMissingTxSignatureOrSigningPubKey
	}

	return c.submitRequest(&requests.SubmitRequest{
		TxBlob:   txBlob,
		FailHard: failHard,
	})
}

// SubmitTxBlobAndWait sends a pre-signed transaction blob to the server,
// decodes it to retrieve the required LastLedgerSequence, submits the blob,
// and then waits until the transaction is confirmed in a ledger. It returns
// the transaction response if the submission is successful.
func (c *Client) SubmitTxBlobAndWait(txBlob string, failHard bool) (*requests.TxResponse, error) {
	tx, err := binarycodec.Decode(txBlob)
	if err != nil {
		return nil, err
	}

	lastLedgerSequence, ok := tx["LastLedgerSequence"].(uint32)
	if !ok {
		return nil, ErrMissingLastLedgerSequenceInTransaction
	}

	txResponse, err := c.SubmitTxBlob(txBlob, failHard)
	if err != nil {
		return nil, err
	}

	if txResponse.EngineResult != "tesSUCCESS" {
		return nil, &ClientError{ErrorString: "transaction failed to submit with engine result: " + txResponse.EngineResult}
	}

	txHash, err := hash.SignTxBlob(txBlob)
	if err != nil {
		return nil, err
	}

	return c.waitForTransaction(txHash, lastLedgerSequence)
}

// SubmitTx signs the transaction (if necessary) and submits it to the server
// via a submission request. It applies the provided submit options to decide whether
// to autofill missing fields and enforce failHard mode during submission.
func (c *Client) SubmitTx(tx transaction.FlatTransaction, opts *rpctypes.SubmitOptions) (*requests.SubmitResponse, error) {
	if opts == nil {
		opts = &rpctypes.SubmitOptions{}
	}
	txBlob, err := c.getSignedTx(tx, opts.Autofill, opts.Wallet)
	if err != nil {
		return nil, err
	}

	return c.submitRequest(&requests.SubmitRequest{
		TxBlob:   txBlob,
		FailHard: opts.FailHard,
	})
}

// SubmitTxAndWait prepares a transaction by ensuring it is fully signed,
// submits it to the server, and waits for ledger confirmation.
// It validates that the transaction's EngineResult is successful before returning
// the transaction response.
func (c *Client) SubmitTxAndWait(tx transaction.FlatTransaction, opts *rpctypes.SubmitOptions) (*requests.TxResponse, error) {
	if opts == nil {
		opts = &rpctypes.SubmitOptions{}
	}
	// Get the signed transaction blob.
	txBlob, err := c.getSignedTx(tx, opts.Autofill, opts.Wallet)
	if err != nil {
		return nil, err
	}

	// Delegate to SubmitTxBlobAndWait to handle submission, engine result check,
	// ledger sequence validation, and waiting for confirmation.
	return c.SubmitTxBlobAndWait(txBlob, opts.FailHard)
}

// SubmitMultisigned submits a multisigned transaction blob to the server and returns the response.
func (c *Client) SubmitMultisigned(txBlob string, failHard bool) (*requests.SubmitMultisignedResponse, error) {
	tx, err := binarycodec.Decode(txBlob)
	if err != nil {
		return nil, err
	}
	signers, okSigners := tx["Signers"].([]any)

	if okSigners && len(signers) > 0 {
		for _, sig := range signers {
			signer := sig.(map[string]any)
			signerData := signer["Signer"].(map[string]any)
			if signerData["SigningPubKey"] == "" && signerData["TxnSignature"] == "" {
				return nil, ErrSignerDataIsEmpty
			}
		}
	}

	return c.submitMultisignedRequest(&requests.SubmitMultisignedRequest{
		Tx:       tx,
		FailHard: failHard,
	})
}

// Autofill fills in the missing fields in a transaction.
func (c *Client) Autofill(tx *transaction.FlatTransaction) error {
	if err := c.setValidTransactionAddresses(tx); err != nil {
		return err
	}

	err := c.setTransactionFlags(tx)
	if err != nil {
		return err
	}

	if _, ok := (*tx)["NetworkID"]; !ok {
		if c.NetworkID != 0 {
			(*tx)["NetworkID"] = c.NetworkID
		}
	}
	if _, ok := (*tx)["Sequence"]; !ok {
		err := c.setTransactionNextValidSequenceNumber(tx)
		if err != nil {
			return err
		}
	}
	if _, ok := (*tx)["Fee"]; !ok {
		err := c.calculateFeePerTransactionType(tx, 0)
		if err != nil {
			return err
		}
	}
	if _, ok := (*tx)["LastLedgerSequence"]; !ok {
		err := c.setLastLedgerSequence(tx)
		if err != nil {
			return err
		}
	}
	if txType, ok := (*tx)["TransactionType"].(string); ok {
		if acc, ok := (*tx)["Account"].(types.Address); txType == transaction.AccountDeleteTx.String() && ok {
			err := c.checkAccountDeleteBlockers(acc)
			if err != nil {
				return err
			}
		}
		if txType == transaction.PaymentTx.String() {
			err := c.checkPaymentAmounts(tx)
			if err != nil {
				return err
			}
		}
		if txType == transaction.BatchTx.String() {
			err := c.autofillRawTransactions(tx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// AutofillMultisigned fills in the missing fields in a multisigned transaction.
// This function is used to fill in the missing fields in a multisigned transaction.
// It fills in the missing fields in the transaction and calculates the fee per number of signers.
func (c *Client) AutofillMultisigned(tx *transaction.FlatTransaction, nSigners uint64) error {
	err := c.Autofill(tx)
	if err != nil {
		return err
	}

	err = c.calculateFeePerTransactionType(tx, nSigners)
	if err != nil {
		return err
	}

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

type validatedInnerTx struct {
	rawTx   map[string]any
	account string
}

func (c *Client) autofillRawTransactions(tx *transaction.FlatTransaction) error {
	rawTxs, ok := (*tx)["RawTransactions"].([]map[string]any)
	if !ok {
		return ErrRawTransactionsFieldIsNotAnArray
	}

	var outerNetworkID *uint32
	if outer := (*tx)["NetworkID"]; outer != nil {
		outerNetworkIDUint, ok := outer.(uint32)
		if !ok {
			return ErrNetworkIDFieldIsNotAUint32
		}
		if outerNetworkIDUint != c.NetworkID {
			return ErrNetworkIDFieldMismatch
		}
		outerNetworkID = &outerNetworkIDUint
	}

	inners := make([]validatedInnerTx, 0, len(rawTxs))
	for _, rawTx := range rawTxs {
		innerRawTx, ok := rawTx["RawTransaction"].(map[string]any)
		if !ok {
			return ErrRawTransactionFieldIsNotAnObject
		}

		acc, ok := innerRawTx["Account"].(string)
		if !ok {
			return ErrAccountFieldIsNotAString
		}

		if fee := innerRawTx["Fee"]; fee != nil && fee != "0" {
			return types.ErrBatchInnerTransactionInvalid
		}

		if signingPubKey := innerRawTx["SigningPubKey"]; signingPubKey != nil && signingPubKey != "" {
			return ErrSigningPubKeyFieldMustBeEmpty
		}

		if innerRawTx["TxnSignature"] != nil {
			return ErrTxnSignatureFieldMustBeEmpty
		}
		if innerRawTx["Signers"] != nil {
			return ErrSignersFieldMustBeEmpty
		}

		if networkID := innerRawTx["NetworkID"]; networkID != nil {
			innerNetworkID, ok := networkID.(uint32)
			if !ok {
				return ErrNetworkIDFieldIsNotAUint32
			}
			if innerNetworkID != c.NetworkID {
				return ErrNetworkIDFieldMismatch
			}
			if outerNetworkID != nil && innerNetworkID != *outerNetworkID {
				return ErrNetworkIDFieldMismatch
			}
		}

		inners = append(inners, validatedInnerTx{rawTx: innerRawTx, account: acc})
	}

	needsNetworkID, err := c.txNeedsNetworkID()
	if err != nil {
		return err
	}

	accountSeq := make(map[string]uint32, len(inners))

	for _, inner := range inners {
		innerRawTx := inner.rawTx
		if innerRawTx["Fee"] == nil {
			innerRawTx["Fee"] = "0"
		}

		if innerRawTx["SigningPubKey"] == nil {
			innerRawTx["SigningPubKey"] = ""
		}

		if innerRawTx["NetworkID"] == nil && needsNetworkID {
			innerRawTx["NetworkID"] = c.NetworkID
		}

		if innerRawTx["Sequence"] == nil && innerRawTx["TicketSequence"] == nil {
			acc := inner.account

			if accountSeq[acc] != 0 {
				innerRawTx["Sequence"] = accountSeq[acc]
				accountSeq[acc]++
			} else {
				accountInfo, err := c.GetAccountInfo(&account.InfoRequest{
					Account: types.Address(acc),
				})
				if err != nil {
					return err
				}
				var seq uint32
				if innerRawTx["Account"] == (*tx)["Account"] {
					seq = accountInfo.AccountData.Sequence + 1
				} else {
					seq = accountInfo.AccountData.Sequence
				}
				accountSeq[acc] = seq + 1
				innerRawTx["Sequence"] = seq
			}
		}
	}

	return nil
}
