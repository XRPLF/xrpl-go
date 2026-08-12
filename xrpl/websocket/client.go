// Package websocket provides a client for connecting to an XRPL WebSocket server.
//
// A Client discovers network identity on every connection unless a trusted
// identity was configured. Explicit and automatic reconnects reject a change
// from the previous discovered network ID.
package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/hash"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	transaction "github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/go-viper/mapstructure/v2"

	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
	streamtypes "github.com/Peersyst/xrpl-go/xrpl/queries/subscription/types"
	requests "github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket/interfaces"
	wstypes "github.com/Peersyst/xrpl-go/xrpl/websocket/types"
	ws "github.com/gorilla/websocket"

	commonconstants "github.com/Peersyst/xrpl-go/xrpl/common"
	"github.com/Peersyst/xrpl-go/xrpl/currency"
	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/internal/clientconfig"
)

const (
	// DefaultFeeCushion is the default cushion factor for fee calculations.
	DefaultFeeCushion float64 = 1.2
	// DefaultMaxFeeXRP is the default maximum fee in XRP.
	DefaultMaxFeeXRP = "2"

	// RestrictedNetworks is the largest network ID for which transactions omit NetworkID.
	RestrictedNetworks = clientinternal.RestrictedNetworks
	// RequiredNetworkIDVersion is the first rippled version that enforces NetworkID.
	RequiredNetworkIDVersion = clientinternal.RequiredNetworkIDVersion
)

var (
	fundWalletMaxAttempts  = 20
	fundWalletPollInterval = 1 * time.Second
)

type pendingResponseResult struct {
	response *ClientResponse
	err      error
}

type pendingResponse struct {
	result chan pendingResponseResult
	socket websocketConnection
	once   sync.Once
}

func newPendingResponse(socket websocketConnection) *pendingResponse {
	return &pendingResponse{
		result: make(chan pendingResponseResult, 1),
		socket: socket,
	}
}

func (p *pendingResponse) complete(result pendingResponseResult) {
	p.once.Do(func() {
		p.result <- result
	})
}

func (p *pendingResponse) cancel() {
	p.once.Do(func() {})
}

// Client is a WebSocket client for interacting with an XRPL server.
//
// Each On* handler is invoked on its own per-stream delivery goroutine, not on
// the socket reader goroutine. Calls to one handler are serialized in wire
// order, while handlers for different streams may run concurrently and have no
// global ordering guarantee. Delivery uses an unbuffered handoff with no event
// queue: while a handler is running, the reader can continue until the next
// event for that same handler, at which point that event applies backpressure
// to the shared reader and can delay all stream and request dispatch. Handlers
// should offload long-running work when that backpressure is undesirable.
// Automatic reconnect preserves On* handler registrations but does not replay
// server-side subscriptions. Callers must resubscribe after reconnect.
type Client struct {
	cfg  ClientConfig
	conn *Connection

	// Channels
	errorStream        lifecycleStream[error]
	ledgerClosedStream lifecycleStream[*streamtypes.LedgerStream]
	validationStream   lifecycleStream[*streamtypes.ValidationStream]
	transactionStream  lifecycleStream[*streamtypes.TransactionStream]
	peerStatusStream   lifecycleStream[*streamtypes.PeerStatusStream]
	orderBookStream    lifecycleStream[*streamtypes.OrderBookStream]
	bookChangesStream  lifecycleStream[*streamtypes.BookChangesStream]
	consensusStream    lifecycleStream[*streamtypes.ConsensusStream]

	// streamHandlerStateMu protects ctx, cancel, detachedHandlerRunners, and
	// coordinated start/reset operations on the registered lifecycleStream runners.
	streamHandlerStateMu sync.Mutex
	// detachedHandlerRunners tracks canceled runners that can still be executing
	// callbacks after Disconnect returns.
	detachedHandlerRunners []<-chan struct{}
	// streamHandlerResetMu serializes full lifecycle resets while old stream
	// handler runners are waited on outside streamHandlerStateMu.
	streamHandlerResetMu sync.Mutex
	// connectionHandshakeMu keeps ordinary requests off a new socket until
	// network identity discovery and socket publication complete.
	connectionHandshakeMu sync.RWMutex
	ctx                   context.Context
	cancel                context.CancelFunc
	pendingResponsesMu    sync.Mutex
	pendingResponses      map[uint64]*pendingResponse

	idCounter atomic.Uint64

	identity networkIdentityState
}

// NewClient creates a new WebSocket client using the provided ClientConfig.
// This client will open and close a websocket connection for each request.
func NewClient(cfg ClientConfig) *Client {
	clientconfig.WarnIfInsecureScheme("websocket", cfg.host)
	if cfg.reconnectBaseDelay <= 0 {
		cfg.reconnectBaseDelay = defaultReconnectBaseDelay
	}
	if cfg.reconnectMaxDelay <= 0 {
		cfg.reconnectMaxDelay = defaultReconnectMaxDelay
	}

	// Pre-canceled so handlers registered before Connect are deferred to the
	// first lifecycle reset, and any stray reportError before Connect is dropped.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	networkID := clientinternal.CloneNetworkID(cfg.networkID)
	trustedIdentity := networkID != nil && cfg.buildVersion != ""
	return &Client{
		cfg:              cfg,
		pendingResponses: make(map[uint64]*pendingResponse),
		conn:             newConnection(cfg.host, cfg.maxResponseSize),
		ctx:              ctx,
		cancel:           cancel,
		identity: networkIdentityState{
			ready:   trustedIdentity,
			trusted: trustedIdentity,
			current: clientinternal.NetworkIdentity{
				NetworkID:    clientinternal.CloneNetworkID(networkID),
				BuildVersion: cfg.buildVersion,
			},
		},
	}
}

// resetLifecycle cancels the current lifecycle, waits for existing handler
// runners to exit, then starts a fresh lifecycle. Lock order is
// streamHandlerResetMu -> streamHandlerStateMu, streamHandlerStateMu is
// released across waitForHandlerRunners so Report callers (which acquire
// streamHandlerStateMu) cannot deadlock the runners they are draining.
func (c *Client) resetLifecycle() context.Context {
	c.streamHandlerResetMu.Lock()
	defer c.streamHandlerResetMu.Unlock()

	c.streamHandlerStateMu.Lock()

	c.cancel()
	doneChannels := append([]<-chan struct{}{}, c.detachedHandlerRunners...)
	c.detachedHandlerRunners = nil
	doneChannels = append(doneChannels, c.resetHandlerRunners()...)
	c.streamHandlerStateMu.Unlock()

	waitForHandlerRunners(doneChannels)

	c.streamHandlerStateMu.Lock()
	defer c.streamHandlerStateMu.Unlock()
	c.ctx, c.cancel = context.WithCancel(context.Background())
	ctx := c.ctx
	c.startRegisteredHandlers(ctx)

	return ctx
}

// lifecycleContext returns the current lifecycle context under
// streamHandlerStateMu.
func (c *Client) lifecycleContext() context.Context {
	c.streamHandlerStateMu.Lock()
	defer c.streamHandlerStateMu.Unlock()

	return c.ctx
}

// cancelLifecycle cancels the current lifecycle and detaches registered handler
// runners under streamHandlerStateMu. It does not wait for runners to exit
// because Disconnect is supported from inside a stream handler, where waiting
// would deadlock the calling runner. Completion channels remain tracked so the
// next lifecycle reset waits before it starts replacement runners.
func (c *Client) cancelLifecycle() {
	c.streamHandlerStateMu.Lock()
	defer c.streamHandlerStateMu.Unlock()

	c.cancel()
	for _, done := range c.resetHandlerRunners() {
		if done != nil {
			c.detachedHandlerRunners = append(c.detachedHandlerRunners, done)
		}
	}
}

// cancelLifecycleForReplacement fails pending requests for the old connection
// and cancels its lifecycle. The caller holds connectionHandshakeMu exclusively,
// so no replacement request can register before the new socket is published.
// Handler runners stay tracked so resetLifecycle can wait for them.
func (c *Client) cancelLifecycleForReplacement() {
	c.failPendingResponses(ErrDisconnected)

	c.streamHandlerStateMu.Lock()
	defer c.streamHandlerStateMu.Unlock()

	c.cancel()
}

// Connect opens a websocket connection to the server. It completes network
// identity discovery before it starts reading messages in a goroutine. Do not
// call Connect synchronously from a stream or error handler. If a handler needs
// to reconnect, start Connect in a separate goroutine or coordinate it outside
// the handler callback.
func (c *Client) Connect() error {
	bufferedMessages, err := c.connect(context.Background(), c.cancelLifecycleForReplacement)
	if err != nil {
		return err
	}

	ctx := c.resetLifecycle()
	for _, message := range bufferedMessages {
		c.handleMessage(ctx, message)
	}
	go c.readMessages(ctx)
	return nil
}

// connect prepares a newly dialed socket before it becomes available to normal
// client requests. It returns stream messages read during identity discovery so
// the caller can replay them after the socket is published. onBeforePublish runs
// under connectionHandshakeMu after preparation succeeds and immediately before
// publication. Manual Connect uses it to cancel the old lifecycle context before
// the new socket becomes visible. Automatic reconnect keeps its current lifecycle.
func (c *Client) connect(ctx context.Context, onBeforePublish func()) ([][]byte, error) {
	c.connectionHandshakeMu.Lock()
	defer c.connectionHandshakeMu.Unlock()

	connectCtx := ctx
	cancel := func() {}
	if c.cfg.timeout > 0 {
		connectCtx, cancel = context.WithTimeout(ctx, c.cfg.timeout)
	}
	defer cancel()

	conn, err := c.conn.beginConnect(connectCtx)
	if err != nil {
		return nil, err
	}
	bufferedMessages, err := c.prepareNetworkIdentity(connectCtx, conn)
	if err != nil {
		if closeErr := c.conn.invalidateSocket(conn); closeErr != nil {
			return nil, errors.Join(err, closeErr)
		}
		return nil, err
	}
	if onBeforePublish != nil {
		onBeforePublish()
	}
	if err := c.conn.publishSocket(connectCtx, conn); err != nil {
		return nil, err
	}
	return bufferedMessages, nil
}

// Disconnect closes the websocket connection and cancels the current client
// lifecycle, including registered handler runners, even when no connection is
// currently open. Disconnect does not wait for runners or readMessages to
// exit: doing so would deadlock when Disconnect is called from inside a
// stream handler. The lifecycle context is canceled and handler runners are
// detached so they drain asynchronously, and the readMessages goroutine is
// unblocked by the socket close performed by the connection disconnect
// operation rather than by context cancellation. On* registrations persist across
// Disconnect, a subsequent successful Connect restarts handler runners
// against the new lifecycle (and resetLifecycle waits for the previous
// runners before starting fresh ones). Callers must serialize concurrent
// calls to Connect and Disconnect externally.
func (c *Client) Disconnect() error {
	c.failPendingResponses(ErrDisconnected)
	err := c.conn.disconnect(c.cancelLifecycle)
	// Reject requests that raced with the socket claim above. Their writes either
	// used the closing socket or failed because the connection was unavailable.
	c.failPendingResponses(ErrDisconnected)
	if errors.Is(err, ErrNotConnected) {
		return nil
	}
	return err
}

// IsConnected returns true if the client is connected to the server.
func (c *Client) IsConnected() bool {
	return c.conn.IsConnected()
}

// FaucetProvider returns the configured faucet provider for the client.
func (c *Client) FaucetProvider() commonconstants.FaucetProvider {
	return c.cfg.faucetProvider
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

	identity, err := c.networkIdentity()
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

// FundWallet funds a wallet with XRP from the faucet and polls the validated
// ledger until the account's balance increases. It returns
// ErrFundWalletBalanceNotUpdated if the balance fails to update within the
// poll window. If the wallet does not have a classic address, it returns an
// error.
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
	var wsErr *ErrorWebsocketClientXrplResponse
	return errors.As(err, &wsErr) && wsErr.Type == actNotFound
}

// Request sends a request to the server and returns the response.
// This function is used to send requests to the server.
// It returns the response from the server.
func (c *Client) Request(req interfaces.Request) (*ClientResponse, error) {
	return c.request(context.Background(), req)
}

func (c *Client) request(ctx context.Context, req interfaces.Request) (*ClientResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	requestCtx, cancel := context.WithTimeoutCause(ctx, c.cfg.timeout, ErrRequestTimedOut)
	defer cancel()

	id := c.idCounter.Add(1)
	msg, err := c.formatRequest(req, id, nil)
	if err != nil {
		return nil, err
	}

	c.connectionHandshakeMu.RLock()
	if cause := context.Cause(requestCtx); cause != nil {
		c.connectionHandshakeMu.RUnlock()
		return nil, cause
	}
	socket := c.conn.currentSocket()
	if socket == nil {
		c.connectionHandshakeMu.RUnlock()
		return nil, ErrNotConnectedToServer
	}
	pendingResponse := c.registerPendingResponse(id, socket)
	defer func() {
		pendingResponse.cancel()
		c.unregisterPendingResponse(id)
	}()

	err = c.conn.writeMessageTo(requestCtx, socket, msg, 0)
	c.connectionHandshakeMu.RUnlock()
	if err != nil {
		if requestCtx.Err() != nil {
			return nil, context.Cause(requestCtx)
		}
		c.failPendingResponsesForSocket(socket, ErrDisconnected)
		return nil, errors.Join(ErrDisconnected, err)
	}

	res, err := c.awaitResponse(requestCtx, pendingResponse)
	if err != nil {
		return nil, err
	}

	if err := res.CheckError(); err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) requestResult(ctx context.Context, req interfaces.Request, result any) error {
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

// SubmitTx signs the transaction (if necessary) and submits it to the server
// via a submission request. It applies the provided submit options to decide whether
// to autofill missing fields and enforce failHard mode during submission.
func (c *Client) SubmitTx(tx transaction.FlatTransaction, opts *wstypes.SubmitOptions) (*requests.SubmitResponse, error) {
	if opts == nil {
		opts = &wstypes.SubmitOptions{}
	}
	txBlob, err := c.getSignedTx(context.Background(), tx, opts.Autofill, opts.Wallet)
	if err != nil {
		return nil, err
	}

	return c.SubmitTxBlob(txBlob, opts.FailHard)
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

// SubmitTxAndWait prepares, submits, and monitors a transaction until its
// validated-ledger outcome is authoritative. Nil options retain their existing
// zero-value behavior, and AccountDelete still forces fail_hard.
func (c *Client) SubmitTxAndWait(tx transaction.FlatTransaction, opts *wstypes.SubmitOptions) (*requests.TxResponse, error) {
	return c.SubmitTxAndWaitContext(context.Background(), tx, opts)
}

// SubmitTxAndWaitContext is SubmitTxAndWait with caller cancellation for
// transaction preparation, submission, and finality monitoring.
func (c *Client) SubmitTxAndWaitContext(
	ctx context.Context,
	tx transaction.FlatTransaction,
	opts *wstypes.SubmitOptions,
) (*requests.TxResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := clientinternal.ValidateFinalityMonitoring(c.cfg.retryDelay, c.cfg.maxRetries); err != nil {
		return nil, err
	}
	if opts == nil {
		opts = &wstypes.SubmitOptions{}
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
	var responseErr *ErrorWebsocketClientXrplResponse
	return errors.As(err, &responseErr) && responseErr.Type == txnNotFound
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

func (c *Client) formatRequest(req interfaces.Request, id uint64, marker any) ([]byte, error) {
	m := make(map[string]any)
	if _, ok := req.(json.Marshaler); ok {
		requestJSON, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}
		// UseNumber preserves numeric fidelity through the map round trip
		// (plain Unmarshal would coerce every number to float64).
		dec := json.NewDecoder(bytes.NewReader(requestJSON))
		dec.UseNumber()
		if err := dec.Decode(&m); err != nil {
			return nil, err
		}
	} else {
		dec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: "json", Result: &m})
		if err := dec.Decode(req); err != nil {
			return nil, err
		}
	}

	m["id"] = id
	m["command"] = req.Method()
	m["api_version"] = req.APIVersion()
	if marker != nil {
		m["marker"] = marker
	}

	return json.Marshal(m)
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

// getFeeDrops calculates the current transaction fee for the ledger.
func (c *Client) getFeeDrops(
	ctx context.Context,
	cushion float64,
	maxFee currency.Drops,
) (currency.Drops, error) {
	var res server.InfoResponse
	if err := c.requestResult(ctx, &server.InfoRequest{}, &res); err != nil {
		return currency.Drops{}, err
	}

	baseFeeXRP := res.Info.ValidatedLedger.BaseFeeXRP
	if baseFeeXRP == nil {
		return currency.Drops{}, ErrCouldNotGetBaseFeeXrp
	}

	return clientinternal.NetworkFeeDrops(
		*baseFeeXRP,
		res.Info.LoadFactor,
		cushion,
		maxFee,
	)
}

// calculateFeePerTransactionType calculates the fee for a transaction,
// including special costs for EscrowFinish, owner-reserve transactions, Batch,
// LoanSet, and multisigning.
func (c *Client) calculateFeePerTransactionType(
	ctx context.Context,
	tx *transaction.FlatTransaction,
	nSigners uint64,
) error {
	maxFee, err := clientinternal.ParseFeeXRP(c.cfg.maxFeeXRP)
	if err != nil {
		return err
	}

	netFee, err := c.getFeeDrops(ctx, c.cfg.feeCushion, maxFee)
	if err != nil {
		return err
	}
	baseFee := netFee

	transactionType := tx.TxType()
	isSpecialTxCost := transactionType == transaction.AccountDeleteTx ||
		transactionType == transaction.AMMCreateTx

	switch transactionType { //nolint:exhaustive // Only transaction types with nonstandard fees need cases.
	case transaction.EscrowFinishTx:
		if fulfillment, ok := (*tx)["Fulfillment"].(string); ok {
			fulfillmentBytesSize := (len(fulfillment) + 1) / 2
			baseFee = netFee.Mul(33 + uint64(fulfillmentBytesSize)/16)
		}
	case transaction.AccountDeleteTx, transaction.AMMCreateTx:
		reserveFee, reserveErr := c.fetchOwnerReserveFee(ctx)
		if reserveErr != nil {
			return reserveErr
		}
		baseFee = currency.DropsFromUint64(reserveFee)
	case transaction.BatchTx:
		rawTxFees, batchErr := c.calculateBatchFees(ctx, tx)
		if batchErr != nil {
			return batchErr
		}
		baseFee = netFee.Mul(2).Add(rawTxFees)
	case transaction.LoanSetTx:
		counterPartySignersCount, signerErr := c.fetchCounterPartySignersCount(ctx, *tx)
		if signerErr != nil {
			return signerErr
		}
		baseFee = netFee.Mul(1 + counterPartySignersCount)
	}

	if nSigners > 0 {
		baseFee = baseFee.Add(netFee.Mul(nSigners))
	}

	totalFee := baseFee
	if !isSpecialTxCost {
		totalFee = baseFee.Min(maxFee)
	}

	fee, err := totalFee.Ceil().WholeString()
	if err != nil {
		return err
	}
	(*tx)["Fee"] = fee
	return nil
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

func (c *Client) registerPendingResponse(id uint64, sockets ...websocketConnection) *pendingResponse {
	var socket websocketConnection
	if len(sockets) > 0 {
		socket = sockets[0]
	}
	response := newPendingResponse(socket)

	c.pendingResponsesMu.Lock()
	defer c.pendingResponsesMu.Unlock()

	c.pendingResponses[id] = response

	return response
}

func (c *Client) unregisterPendingResponse(id uint64) {
	c.pendingResponsesMu.Lock()
	defer c.pendingResponsesMu.Unlock()

	delete(c.pendingResponses, id)
}

func (c *Client) lookupPendingResponse(id uint64) (*pendingResponse, bool) {
	c.pendingResponsesMu.Lock()
	defer c.pendingResponsesMu.Unlock()

	responseChan, ok := c.pendingResponses[id]
	return responseChan, ok
}

func (c *Client) failPendingResponses(err error) {
	c.pendingResponsesMu.Lock()
	pendingResponses := c.pendingResponses
	c.pendingResponses = make(map[uint64]*pendingResponse)
	c.pendingResponsesMu.Unlock()

	completePendingResponses(pendingResponses, err)
}

func (c *Client) failPendingResponsesForSocket(socket websocketConnection, err error) {
	c.pendingResponsesMu.Lock()
	pendingResponses := make(map[uint64]*pendingResponse)
	for id, response := range c.pendingResponses {
		if response.socket == socket {
			pendingResponses[id] = response
			delete(c.pendingResponses, id)
		}
	}
	c.pendingResponsesMu.Unlock()

	completePendingResponses(pendingResponses, err)
}

func completePendingResponses(pendingResponses map[uint64]*pendingResponse, err error) {
	result := pendingResponseResult{err: err}
	for _, response := range pendingResponses {
		response.complete(result)
	}
}

func (c *Client) awaitResponse(ctx context.Context, response *pendingResponse) (*ClientResponse, error) {
	select {
	case result := <-response.result:
		return result.response, result.err
	case <-ctx.Done():
		response.cancel()
		return nil, context.Cause(ctx)
	}
}

func (c *Client) handleMessage(ctx context.Context, message []byte) {
	var stream wstypes.Message
	if !c.unmarshalMessage(ctx, message, &stream) {
		return
	}
	if stream.IsRequest() {
		c.handleRequest(ctx, message)
	} else if stream.IsStream() {
		c.handleStream(ctx, stream.Type, message)
	}
}

func (c *Client) handleRequest(ctx context.Context, message []byte) {
	var res ClientResponse
	if !c.unmarshalMessage(ctx, message, &res) {
		return
	}
	response, ok := c.lookupPendingResponse(res.ID)
	if !ok {
		return
	}

	response.complete(pendingResponseResult{response: &res})
}

func (c *Client) unmarshalMessage(ctx context.Context, message []byte, v any) bool {
	if err := json.Unmarshal(message, v); err != nil {
		c.reportError(ctx, err)
		return false
	}
	return true
}

func (c *Client) handleStream(ctx context.Context, t streamtypes.Type, message []byte) {
	switch t {
	case streamtypes.LedgerStreamType:
		var ledger streamtypes.LedgerStream
		if c.unmarshalMessage(ctx, message, &ledger) {
			c.reportLedgerClosed(ctx, &ledger)
		}
	case streamtypes.TransactionStreamType:
		var transactionStream streamtypes.TransactionStream
		if !c.unmarshalMessage(ctx, message, &transactionStream) {
			return
		}
		c.reportTransaction(ctx, &transactionStream)
	case streamtypes.ValidationStreamType:
		var validation streamtypes.ValidationStream
		if c.unmarshalMessage(ctx, message, &validation) {
			c.reportValidationReceived(ctx, &validation)
		}
	case streamtypes.PeerStatusStreamType:
		var peerStatus streamtypes.PeerStatusStream
		if c.unmarshalMessage(ctx, message, &peerStatus) {
			c.reportPeerStatusChange(ctx, &peerStatus)
		}
	case streamtypes.BookChangesStreamType:
		var bookChanges streamtypes.BookChangesStream
		if c.unmarshalMessage(ctx, message, &bookChanges) {
			c.reportBookChanges(ctx, &bookChanges)
		}
	case streamtypes.ConsensusStreamType:
		var consensus streamtypes.ConsensusStream
		if c.unmarshalMessage(ctx, message, &consensus) {
			c.reportConsensusPhase(ctx, &consensus)
		}
	default:
		c.reportError(ctx, ErrUnknownStreamType{
			Type: t,
		})
	}
}

func (c *Client) readMessages(ctx context.Context) {
	retryCount := 0
	maxRetries := c.cfg.maxReconnects

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if c.conn == nil {
			return
		}
		message, failedSocket, err := c.conn.readMessageWithSocket(time.Time{})
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err != nil {
			if failedSocket != nil {
				wasCurrent, closeErr := c.conn.invalidateSocketState(failedSocket)
				if closeErr != nil {
					c.reportError(ctx, closeErr)
				}
				c.failPendingResponsesForSocket(failedSocket, ErrDisconnected)
				// A failed socket is stale only when a replacement is already
				// published. If no current socket exists, another operation, such
				// as a failed active write, already invalidated this socket and this
				// reader must still start reconnection.
				if !wasCurrent && c.IsConnected() {
					return
				}
			} else {
				c.failPendingResponses(ErrDisconnected)
			}
			if !ws.IsCloseError(err) && !ws.IsUnexpectedCloseError(err) {
				c.reportError(ctx, err)
			}
			if !c.reconnectWithBackoff(ctx, &retryCount, maxRetries) {
				return
			}
			continue
		}

		// Send the message to the channel.
		c.handleMessage(ctx, message)
		// Reset retry count on a successful message.
		retryCount = 0
	}
}

func (c *Client) disconnectConnection(ctx context.Context) {
	if err := c.conn.Disconnect(); err != nil && !errors.Is(err, ErrNotConnected) {
		c.reportError(ctx, err)
	}
}

// reconnectWithBackoff retries c.conn.Connect() with capped exponential
// backoff until it succeeds or the budget is exhausted. Returns true on
// success and false when the budget is exhausted or ctx is cancelled.
// retryCount is updated in place so it persists across disconnect events.
func (c *Client) reconnectWithBackoff(ctx context.Context, retryCount *int, maxRetries int) bool {
	var lastErr error
	for {
		if c.IsConnected() {
			return true
		}
		if *retryCount >= maxRetries {
			c.reportError(ctx, ErrMaxReconnectionAttemptsReached{
				Attempts: maxRetries,
				Err:      lastErr,
			})
			return false
		}
		*retryCount++

		timer := time.NewTimer(reconnectDelay(
			*retryCount,
			c.cfg.reconnectBaseDelay,
			c.cfg.reconnectMaxDelay,
		))
		select {
		case <-ctx.Done():
			timer.Stop()
			return false
		case <-timer.C:
		}

		bufferedMessages, connErr := c.connect(ctx, nil)
		if connErr != nil {
			if ctx.Err() != nil || errors.Is(connErr, context.Canceled) {
				return false
			}
			lastErr = connErr
			if errors.Is(connErr, ErrAlreadyConnected) && c.IsConnected() {
				return true
			}
			continue
		}
		for _, message := range bufferedMessages {
			c.handleMessage(ctx, message)
		}
		if ctx.Err() != nil {
			c.disconnectConnection(ctx)
			return false
		}
		return true
	}
}

// reconnectDelay returns baseDelay * 2^(attempt-1), capped at maxDelay.
// Attempt is 1-indexed.
func reconnectDelay(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	backoff := baseDelay
	for i := 1; i < attempt && backoff < maxDelay; i++ {
		backoff *= 2
	}
	if backoff > maxDelay {
		return maxDelay
	}
	return backoff
}

// getSignedTx ensures the transaction is fully signed and returns the transaction blob.
// Submission works on a deep copy, so autofill, address conversion, NetworkID policy,
// and DeliverMax normalization never mutate the caller-owned transaction map.
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
		identity, err := c.networkIdentity()
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

// fetchOwnerReserveFee fetches the owner reserve fee from the server state.
func (c *Client) fetchOwnerReserveFee(ctx context.Context) (uint64, error) {
	var response server.StateResponse
	if err := c.requestResult(ctx, &server.StateRequest{}, &response); err != nil {
		return 0, err
	}

	reserveInc := response.State.ValidatedLedger.ReserveInc
	if reserveInc == nil {
		return 0, ErrCouldNotFetchOwnerReserve
	}

	return *reserveInc, nil
}

// fetchCounterPartySignersCount fetches the number of signers for the counterparty account.
// For LoanSet transactions, if Counterparty is not provided, it fetches the LoanBroker and uses its Owner.
// Returns the number of signers in the counterparty's signer list, or 1 if no signer list exists.
func (c *Client) fetchCounterPartySignersCount(
	ctx context.Context,
	tx transaction.FlatTransaction,
) (uint64, error) {
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
		var res ledger.EntryResponse
		if err := c.requestResult(ctx, &ledger.EntryRequest{
			Index:       loanBrokerID,
			LedgerIndex: common.LedgerTitle("validated"),
		}, &res); err != nil {
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
	var accountInfo account.InfoResponse
	if err := c.requestResult(ctx, &account.InfoRequest{
		Account:     counterparty,
		LedgerIndex: common.LedgerTitle("validated"),
		SignerLists: true,
	}, &accountInfo); err != nil {
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
func (c *Client) calculateBatchFees(
	ctx context.Context,
	tx *transaction.FlatTransaction,
) (currency.Drops, error) {
	var totalFees currency.Drops

	// Get RawTransactions from the batch transaction
	rawTransactions, ok := (*tx)["RawTransactions"].([]map[string]any)
	if !ok {
		return currency.Drops{}, ErrRawTransactionsFieldMissing
	}

	// Iterate through each raw transaction
	for _, rawTx := range rawTransactions {
		// Extract the actual transaction from the wrapper
		innerTx, ok := rawTx["RawTransaction"].(map[string]any)
		if !ok {
			return currency.Drops{}, ErrRawTransactionFieldMissing
		}

		// Calculate fee for this inner transaction (no multi-signing for inner transactions)
		innerTxFlat := transaction.FlatTransaction(innerTx)
		if innerTxFlat.TxType() == transaction.BatchTx {
			return currency.Drops{}, types.ErrBatchNestedTransaction
		}
		err := c.calculateFeePerTransactionType(ctx, &innerTxFlat, 0)
		if err != nil {
			return currency.Drops{}, err
		}

		// Extract the calculated fee
		feeStr, ok := innerTx["Fee"].(string)
		if !ok {
			return currency.Drops{}, ErrFeeFieldMissing
		}

		innerTx["Fee"] = "0"

		innerFee, err := currency.DropsFromString(feeStr)
		if err != nil {
			return currency.Drops{}, ErrFailedToParseFee{
				Fee: feeStr,
				Err: err,
			}
		}

		totalFees = totalFees.Add(innerFee)
	}

	return totalFees, nil
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
