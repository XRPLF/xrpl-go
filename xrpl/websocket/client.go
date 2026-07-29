// Package websocket provides a client for connecting to an XRPL WebSocket server.
package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
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

	// streamHandlerStateMu protects ctx, cancel, and coordinated start/reset
	// operations on the registered lifecycleStream runners.
	streamHandlerStateMu sync.Mutex
	// streamHandlerResetMu serializes full lifecycle resets while old stream
	// handler runners are waited on outside streamHandlerStateMu.
	streamHandlerResetMu sync.Mutex
	ctx                  context.Context
	cancel               context.CancelFunc
	pendingResponsesMu   sync.Mutex
	pendingResponses     map[uint64]*pendingResponse

	idCounter atomic.Uint64

	// NetworkID is the discovered network identity or a compare-mode override.
	// Nil means unknown. A pointer to zero is mainnet. Configure overrides before
	// Connect and do not mutate identity fields concurrently. Use
	// WithNetworkIdentity for an explicit trusted discovery bypass.
	NetworkID *uint32
	// BuildVersion is the discovered rippled version used for NetworkID policy.
	// Configure it before Connect and do not mutate it concurrently.
	BuildVersion string

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
		NetworkID:        networkID,
		BuildVersion:     cfg.buildVersion,
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
	doneChannels := c.resetHandlerRunners()
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

// cancelLifecycle cancels the current lifecycle and clears registered handler
// runners under streamHandlerStateMu. It does not wait for runners to exit:
// Disconnect is supported from inside a stream handler, where waiting would
// deadlock the calling runner. Orphaned runners exit asynchronously when they
// observe ctx.Done.
func (c *Client) cancelLifecycle() {
	c.streamHandlerStateMu.Lock()
	defer c.streamHandlerStateMu.Unlock()

	c.cancel()
	c.resetHandlerRunners()
}

// Connect opens a websocket connection to the server. It starts reading messages in a goroutine.
// Do not call Connect synchronously from a stream or error handler. If a handler
// needs to reconnect, start Connect in a separate goroutine or coordinate it
// outside the handler callback.
func (c *Client) Connect() error {
	if err := c.connect(context.Background()); err != nil {
		return err
	}

	ctx := c.resetLifecycle()
	go c.readMessages(ctx)
	return nil
}

// connect prepares a newly dialed socket before it becomes available to normal
// client requests.
func (c *Client) connect(ctx context.Context) error {
	conn, err := c.conn.beginConnect(ctx)
	if err != nil {
		return err
	}
	if err := c.prepareNetworkIdentity(ctx, conn); err != nil {
		if closeErr := c.conn.invalidateSocket(conn); closeErr != nil {
			return errors.Join(err, closeErr)
		}
		return err
	}
	return c.conn.publishSocket(ctx, conn)
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
	if err := c.autofill(&working); err != nil {
		return err
	}
	clientinternal.ReplaceTransactionContents(*tx, working)
	return nil
}

func (c *Client) autofill(tx *transaction.FlatTransaction) error {
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
		if err := c.setTransactionNextValidSequenceNumber(tx); err != nil {
			return err
		}
	}
	if _, ok := (*tx)["Fee"]; !ok {
		if err := c.calculateFeePerTransactionType(tx, 0); err != nil {
			return err
		}
	}
	if _, ok := (*tx)["LastLedgerSequence"]; !ok {
		if err := c.setLastLedgerSequence(tx); err != nil {
			return err
		}
	}

	txType := tx.TxType()
	if txType == transaction.AccountDeleteTx {
		account, ok := (*tx)["Account"].(string)
		if !ok {
			return ErrMissingAccountInTransaction
		}
		if err := c.checkAccountDeleteBlockers(types.Address(account)); err != nil {
			return err
		}
	}
	if txType == transaction.BatchTx {
		if err := c.autofillRawTransactions(tx); err != nil {
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
	if err := c.autofill(&working); err != nil {
		return err
	}
	if err := c.calculateFeePerTransactionType(&working, nSigners); err != nil {
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
	socket := c.conn.currentSocket()
	if socket == nil {
		return nil, ErrNotConnectedToServer
	}

	requestCtx, cancel := context.WithTimeoutCause(ctx, c.cfg.timeout, ErrRequestTimedOut)
	defer cancel()

	id := c.idCounter.Add(1)

	msg, err := c.formatRequest(req, id, nil)
	if err != nil {
		return nil, err
	}

	pendingResponse := c.registerPendingResponse(id, socket)
	defer func() {
		pendingResponse.cancel()
		c.unregisterPendingResponse(id)
	}()

	err = c.conn.writeMessageTo(requestCtx, socket, msg, 0)
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

// SubmitTxBlob sends a pre-signed transaction blob to the server.
// Its preflight validates only the structure of signing fields. rippled remains
// authoritative for cryptographic signature validity. AccountDelete always uses
// fail_hard as required by reliable-submission safety guidance.
func (c *Client) SubmitTxBlob(txBlob string, failHard bool) (*requests.SubmitResponse, error) {
	return c.submitTxBlob(context.Background(), txBlob, failHard)
}

func (c *Client) submitTxBlob(ctx context.Context, txBlob string, failHard bool) (*requests.SubmitResponse, error) {
	tx, err := clientinternal.DecodeTransactionBlob(txBlob)
	if err != nil {
		return nil, err
	}

	form, err := clientinternal.InspectSignedTransaction(tx, false)
	if err != nil {
		return nil, err
	}
	if err := clientinternal.InspectSignedBatchInners(tx); err != nil {
		return nil, err
	}
	if form == clientinternal.UnsignedTransaction {
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
	txBlob, err := c.getSignedTx(tx, opts.Autofill, opts.Wallet)
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
	form, err := clientinternal.InspectSignedTransaction(tx, false)
	if err != nil {
		return nil, err
	}
	if err := clientinternal.InspectSignedBatchInners(tx); err != nil {
		return nil, err
	}
	if form != clientinternal.MultiSignedTransaction {
		return nil, ErrSignerDataIsEmpty
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
	if err := clientinternal.ValidatePollInterval(c.cfg.retryDelay); err != nil {
		return nil, err
	}
	tx, err := clientinternal.DecodeTransactionBlob(txBlob)
	if err != nil {
		return nil, err
	}

	lastLedgerSequence, ok := tx["LastLedgerSequence"].(uint32)
	if !ok {
		return nil, ErrMissingLastLedgerSequenceInTransaction
	}

	submitResponse, err := c.submitTxBlob(ctx, txBlob, failHard)
	if err != nil {
		return nil, err
	}
	if err := clientinternal.ValidatePreliminaryResult(
		submitResponse.EngineResult,
		submitResponse.EngineResultMessage,
	); err != nil {
		return nil, err
	}

	txHash, err := hash.SignTxBlob(txBlob)
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
// submission and finality monitoring.
func (c *Client) SubmitTxAndWaitContext(
	ctx context.Context,
	tx transaction.FlatTransaction,
	opts *wstypes.SubmitOptions,
) (*requests.TxResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := clientinternal.ValidatePollInterval(c.cfg.retryDelay); err != nil {
		return nil, err
	}
	if opts == nil {
		opts = &wstypes.SubmitOptions{}
	}
	txBlob, err := c.getSignedTx(tx, opts.Autofill, opts.Wallet)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return c.SubmitTxBlobAndWaitContext(ctx, txBlob, opts.FailHard)
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
	m["id"] = id
	m["command"] = req.Method()
	m["api_version"] = req.APIVersion()
	if marker != nil {
		m["marker"] = marker
	}
	dec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: "json", Result: &m})
	err := dec.Decode(req)
	if err != nil {
		return nil, err
	}

	return json.Marshal(m)
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

		// Order-book subscriptions use the transaction wire type too. The
		// message does not identify which subscription matched, so dispatch all
		// transaction messages to both exported transaction handlers.
		var orderBook streamtypes.OrderBookStream
		if c.unmarshalMessage(ctx, message, &orderBook) {
			c.reportOrderBook(ctx, &orderBook)
		}
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
				if !wasCurrent {
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

// reconnectWithBackoff retries c.conn.Connect() with capped exponential
// backoff until it succeeds or the budget is exhausted. Returns true on
// success and false when the budget is exhausted or ctx is cancelled.
// retryCount is updated in place so it persists across disconnect events.
func (c *Client) reconnectWithBackoff(ctx context.Context, retryCount *int, maxRetries int) bool {
	for {
		if c.IsConnected() {
			return true
		}
		if *retryCount >= maxRetries {
			c.reportError(ctx, ErrMaxReconnectionAttemptsReached{
				Attempts: maxRetries,
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

		if connErr := c.connect(ctx); connErr != nil {
			if ctx.Err() != nil || errors.Is(connErr, context.Canceled) {
				return false
			}
			if errors.Is(connErr, ErrAlreadyConnected) && c.IsConnected() {
				return true
			}
			continue
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
		identity, err := c.networkIdentity()
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

type validatedInnerTx struct {
	rawTx   map[string]any
	account string
}

func (c *Client) autofillRawTransactions(tx *transaction.FlatTransaction) error {
	identity, err := c.networkIdentity()
	if err != nil {
		return err
	}
	if err := clientinternal.ApplyNetworkIDPolicy(*tx, identity); err != nil {
		return err
	}

	rawTxs, ok := (*tx)["RawTransactions"].([]map[string]any)
	if !ok {
		return ErrRawTransactionsFieldIsNotAnArray
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

		// Interface comparisons against a string literal are false for any
		// non-string value (including nil), so no type assertion is needed.
		if fee, exists := innerRawTx["Fee"]; exists && fee != "0" {
			return types.ErrBatchInnerTransactionInvalid
		}

		if signingPubKey, exists := innerRawTx["SigningPubKey"]; exists && signingPubKey != "" {
			return ErrSigningPubKeyFieldMustBeEmpty
		}

		if _, exists := innerRawTx["TxnSignature"]; exists {
			return ErrTxnSignatureFieldMustBeEmpty
		}
		if _, exists := innerRawTx["Signers"]; exists {
			return ErrSignersFieldMustBeEmpty
		}

		inners = append(inners, validatedInnerTx{rawTx: innerRawTx, account: acc})
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
