// Package websocket provides a client for connecting to an XRPL WebSocket server.
package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/xrpl/currency"
	"github.com/Peersyst/xrpl-go/xrpl/hash"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	transaction "github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/mitchellh/mapstructure"

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
	"github.com/Peersyst/xrpl-go/xrpl/internal/clientconfig"
)

const (
	// DefaultFeeCushion is the default cushion factor for fee calculations.
	DefaultFeeCushion float32 = 1.2
	// DefaultMaxFeeXRP is the default maximum fee in XRP.
	DefaultMaxFeeXRP float32 = 2

	// RestrictedNetworks is the minimum network ID above which networks are considered restricted.
	// Sidechains are expected to have network IDs above this.
	// Networks with ID above this restricted number are expected to specify an accurate NetworkID field
	// in every transaction to that chain to prevent replay attacks.
	// Mainnet and testnet are exceptions. More context: https://github.com/XRPLF/rippled/pull/4370
	RestrictedNetworks = 1024
	// RequiredNetworkIDVersion is the minimum XRPL server build version after which specifying NetworkID is required for restricted networks.
	RequiredNetworkIDVersion = "1.11.0"
)

var (
	fundWalletMaxAttempts  = 20
	fundWalletPollInterval = 1 * time.Second
)

// Client is a WebSocket client for interacting with an XRPL server.
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
	ctx                context.Context
	cancel             context.CancelFunc
	pendingResponsesMu sync.Mutex
	pendingResponses   map[uint64]chan *ClientResponse

	idCounter atomic.Uint64
	NetworkID uint32
}

// NewClient creates a new WebSocket client using the provided ClientConfig.
// This client will open and close a websocket connection for each request.
func NewClient(cfg ClientConfig) *Client {
	clientconfig.WarnIfInsecureScheme("websocket", cfg.host)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	return &Client{
		cfg:              cfg,
		pendingResponses: make(map[uint64]chan *ClientResponse),
		conn:             newConnection(cfg.host, cfg.maxResponseSize),
		ctx:              ctx,
		cancel:           cancel,
	}
}

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

func (c *Client) lifecycleContext() context.Context {
	c.streamHandlerStateMu.Lock()
	defer c.streamHandlerStateMu.Unlock()

	return c.ctx
}

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
	err := c.conn.Connect()
	if err != nil {
		return err
	}
	ctx := c.resetLifecycle()
	go c.readMessages(ctx)
	return nil
}

// Disconnect closes the websocket connection and cancels the current client
// lifecycle, including registered handler runners, even when no connection is
// currently open.
func (c *Client) Disconnect() error {
	c.cancelLifecycle()
	return c.conn.Disconnect()
}

// IsConnected returns true if the client is connected to the server.
func (c *Client) IsConnected() bool {
	return c.conn.IsConnected()
}

// FaucetProvider returns the configured faucet provider for the client.
func (c *Client) FaucetProvider() commonconstants.FaucetProvider {
	return c.cfg.faucetProvider
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
	err := req.Validate()
	if err != nil {
		return nil, err
	}

	id := c.idCounter.Add(1)

	msg, err := c.formatRequest(req, id, nil)
	if err != nil {
		return nil, err
	}

	responseChan := c.registerPendingResponse(id)
	defer c.unregisterPendingResponse(id)

	err = c.conn.WriteMessage(msg)
	if err != nil {
		return nil, err
	}

	res, err := c.awaitResponse(responseChan)
	if err != nil {
		return nil, err
	}

	if err := res.CheckError(); err != nil {
		return nil, err
	}

	return res, nil
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

// SubmitTx signs the transaction (if necessary) and submits it to the server
// via a submission request. It applies the provided submit options to decide whether
// to autofill missing fields and enforce failHard mode during submission.
func (c *Client) SubmitTx(tx transaction.FlatTransaction, opts *wstypes.SubmitOptions) (*requests.SubmitResponse, error) {
	txBlob, err := c.getSignedTx(tx, opts.Autofill, opts.Wallet)
	if err != nil {
		return nil, err
	}

	return c.submitRequest(&requests.SubmitRequest{
		TxBlob:   txBlob,
		FailHard: opts.FailHard,
	})
}

// SubmitMultisigned sends a multisigned transaction to the server and returns the response.
// This function is used to send multisigned transactions to the server.
// It returns the response from the server.
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

// SubmitTxAndWait prepares a transaction by ensuring it is fully signed,
// submits it to the server, and waits for ledger confirmation.
// It validates that the transaction's EngineResult is successful before returning
// the transaction response.
func (c *Client) SubmitTxAndWait(tx transaction.FlatTransaction, opts *wstypes.SubmitOptions) (*requests.TxResponse, error) {
	// Get the signed transaction blob.
	txBlob, err := c.getSignedTx(tx, opts.Autofill, opts.Wallet)
	if err != nil {
		return nil, err
	}

	// Delegate to SubmitTxBlobAndWait to handle submission, engine result check,
	// ledger sequence validation, and waiting for confirmation.
	return c.SubmitTxBlobAndWait(txBlob, opts.FailHard)
}

func (c *Client) waitForTransaction(txHash string, lastLedgerSequence uint32) (*requests.TxResponse, error) {
	var txResponse *requests.TxResponse

	for range c.cfg.maxRetries {
		// Get the current ledger index
		currentLedger, err := c.GetLedgerIndex()
		if err != nil {
			return nil, err
		}

		// Check if the transaction has been included in the current ledger
		if currentLedger.Int() >= int(lastLedgerSequence) {
			break
		}

		// Request the transaction from the server
		res, err := c.Request(&requests.TxRequest{
			Transaction: txHash,
		})
		if err != nil && !strings.Contains(err.Error(), txnNotFound) {
			return nil, err
		}

		if res != nil {
			err = res.GetResult(&txResponse)
			if err != nil {
				return nil, err
			}

			// Check if the transaction has been validated
			if txResponse.Validated {
				return txResponse, nil
			}

			// Check if the transaction has been included in the current ledger
			if txResponse.LedgerIndex.Int() >= int(lastLedgerSequence) {
				break
			}
		}

		// Wait for the retry delay before retrying
		time.Sleep(c.cfg.retryDelay)
	}

	if txResponse == nil {
		return nil, ErrTransactionNotFound
	}

	return txResponse, nil
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

func (c *Client) submitRequest(req *requests.SubmitRequest) (*requests.SubmitResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var subRes requests.SubmitResponse
	err = res.GetResult(&subRes)
	if err != nil {
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

// TODO: Implement this when IsValidXAddress is implemented
func (c *Client) getClassicAccountAndTag(address string) (string, uint32) {
	return address, 0
}

func (c *Client) convertTransactionAddressToClassicAddress(tx *transaction.FlatTransaction, fieldName string) {
	if address, ok := (*tx)[fieldName].(string); ok {
		classicAddress, _ := c.getClassicAccountAndTag(address)
		(*tx)[fieldName] = classicAddress
	}
}

func (c *Client) validateTransactionAddress(tx *transaction.FlatTransaction, addressField, tagField string) error {
	classicAddress, tag := c.getClassicAccountAndTag((*tx)[addressField].(string))
	(*tx)[addressField] = classicAddress

	if tag != uint32(0) {
		if txTag, ok := (*tx)[tagField].(uint32); ok && txTag != tag {
			return fmt.Errorf("the %s, if present, must be equal to the tag of the %s", addressField, tagField)
		}
		(*tx)[tagField] = tag
	}

	return nil
}

// Sets valid addresses for the transaction.
func (c *Client) setValidTransactionAddresses(tx *transaction.FlatTransaction) error {
	// Validate if "Account" address is an xAddress
	if err := c.validateTransactionAddress(tx, "Account", "SourceTag"); err != nil {
		return err
	}

	if _, ok := (*tx)["Destination"]; ok {
		if err := c.validateTransactionAddress(tx, "Destination", "DestinationTag"); err != nil {
			return err
		}
	}

	// DepositPreuaht
	c.convertTransactionAddressToClassicAddress(tx, "Authorize")
	c.convertTransactionAddressToClassicAddress(tx, "Unauthorize")
	// EscrowCancel, EscrowFinish
	c.convertTransactionAddressToClassicAddress(tx, "Owner")
	// SetRegularKey
	c.convertTransactionAddressToClassicAddress(tx, "RegularKey")

	return nil
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
// Note: This is a public API that can be called directly.
func (c *Client) getFeeXrp(cushion float32) (string, error) {
	res, err := c.GetServerInfo(&server.InfoRequest{})
	if err != nil {
		return "", err
	}

	if res.Info.ValidatedLedger.BaseFeeXRP == 0 {
		return "", ErrCouldNotGetBaseFeeXrp
	}

	loadFactor := res.Info.LoadFactor
	if res.Info.LoadFactor == 0 {
		loadFactor = 1
	}

	fee := res.Info.ValidatedLedger.BaseFeeXRP * float32(loadFactor) * cushion

	if fee > c.cfg.maxFeeXRP {
		fee = c.cfg.maxFeeXRP
	}

	// Round fee to NUM_DECIMAL_PLACES
	roundedFee := float32(math.Round(float64(fee)*math.Pow10(currency.MaxFractionLength))) / float32(math.Pow10(currency.MaxFractionLength))

	// Convert the rounded fee back to a string with NUM_DECIMAL_PLACES
	return fmt.Sprintf("%.*f", currency.MaxFractionLength, roundedFee), nil
}

// Calculates the fee per transaction type.
//
// Enhanced implementation that replicates calculateFeePerTransactionType logic,
// including special cases for EscrowFinish, AccountDelete, AMMCreate, Batch, and multi-signing.
func (c *Client) calculateFeePerTransactionType(tx *transaction.FlatTransaction, nSigners uint64) error {
	// Get base network fee
	netFeeXRP, err := c.getFeeXrp(c.cfg.feeCushion)
	if err != nil {
		return err
	}

	netFeeDrops, err := currency.XrpToDrops(netFeeXRP)
	if err != nil {
		return err
	}

	// Convert to uint64 for calculations
	baseFeeUint, err := strconv.ParseUint(netFeeDrops, 10, 64)
	if err != nil {
		return err
	}

	baseFee := baseFeeUint

	// Get transaction type
	transactionType := ""
	if txType, ok := (*tx)["TransactionType"]; ok {
		if str, ok := txType.(string); ok {
			transactionType = str
		}
	}

	// Check if this is a special transaction cost type
	isSpecialTxCost := transactionType == "AccountDelete" || transactionType == "AMMCreate"

	switch transactionType {
	case "EscrowFinish":
		if fulfillment, ok := (*tx)["Fulfillment"]; ok && fulfillment != nil {
			if fulfillmentStr, ok := fulfillment.(string); ok && fulfillmentStr != "" {
				fulfillmentBytesSize := (len(fulfillmentStr) + 1) / 2 // Math.ceil(length / 2)
				if fulfillmentBytesSize < 0 {
					return ErrInvalidFulfillmentLength
				}
				// BaseFee × (33 + ceil(Fulfillment size in bytes / 16))
				chunks := (uint64(fulfillmentBytesSize) + 15) / 16 // ceil division
				baseFee = baseFeeUint * (33 + chunks)
			}
		}
	case "AccountDelete", "AMMCreate":
		reserveFee, err := c.fetchOwnerReserveFee()
		if err != nil {
			return err
		}
		baseFee = reserveFee
	case "Batch":
		rawTxFees, err := c.calculateBatchFees(tx)
		if err != nil {
			return err
		}
		baseFee = baseFeeUint*2 + rawTxFees
	case "LoanSet":
		// For LoanSet, account for counterparty signers
		counterPartySignersCount, err := c.fetchCounterPartySignersCount(*tx)
		if err != nil {
			return err
		}
		baseFee = baseFeeUint + (baseFeeUint * counterPartySignersCount)
	}

	// Multi-signed Transaction: BaseFee × (1 + Number of Signatures Provided)
	if nSigners > 0 {
		signersFee := baseFeeUint * nSigners
		baseFee += signersFee
	}

	// Apply max fee limit (but not for special transaction cost types)
	var totalFee uint64
	if isSpecialTxCost {
		totalFee = baseFee
	} else {
		maxFeeDrops, err := currency.XrpToDrops(fmt.Sprintf("%.6f", c.cfg.maxFeeXRP))
		if err != nil {
			return err
		}
		maxFeeUint, err := strconv.ParseUint(maxFeeDrops, 10, 64)
		if err != nil {
			return err
		}
		totalFee = min(baseFee, maxFeeUint)
	}

	(*tx)["Fee"] = strconv.FormatUint(totalFee, 10)
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
	if _, ok := (*tx)["DeliverMax"]; ok {
		if _, ok := (*tx)["Amount"]; !ok {
			(*tx)["Amount"] = (*tx)["DeliverMax"]
		} else if (*tx)["Amount"] != (*tx)["DeliverMax"] {
			return ErrAmountAndDeliverMaxMustBeIdentical
		}
	}
	return nil
}

// Sets a transaction's flags to its numeric representation.
// TODO: Add flag support for AMMDeposit, AMMWithdraw,
// NFTTOkenCreateOffer, NFTokenMint, OfferCreate, XChainModifyBridge (not supported).
func (c *Client) setTransactionFlags(tx *transaction.FlatTransaction) error {
	flags, ok := (*tx)["Flags"].(uint32)
	if !ok && flags > 0 {
		(*tx)["Flags"] = int(0)
		return nil
	}

	_, ok = (*tx)["TransactionType"].(string)
	if !ok {
		return ErrTransactionTypeMissing
	}

	return nil
}

func (c *Client) registerPendingResponse(id uint64) chan *ClientResponse {
	responseChan := make(chan *ClientResponse, 1)

	c.pendingResponsesMu.Lock()
	defer c.pendingResponsesMu.Unlock()

	c.pendingResponses[id] = responseChan

	return responseChan
}

func (c *Client) unregisterPendingResponse(id uint64) {
	c.pendingResponsesMu.Lock()
	defer c.pendingResponsesMu.Unlock()

	delete(c.pendingResponses, id)
}

func (c *Client) lookupPendingResponse(id uint64) (chan *ClientResponse, bool) {
	c.pendingResponsesMu.Lock()
	defer c.pendingResponsesMu.Unlock()

	responseChan, ok := c.pendingResponses[id]
	return responseChan, ok
}

func (c *Client) awaitResponse(responseChan <-chan *ClientResponse) (*ClientResponse, error) {
	timer := time.NewTimer(c.cfg.timeout)
	defer timer.Stop()

	select {
	case res := <-responseChan:
		return res, nil
	case <-timer.C:
		return nil, ErrRequestTimedOut
	}
}

func (c *Client) handleMessage(ctx context.Context, message []byte) {
	var stream wstypes.Message
	c.unmarshalMessage(ctx, message, &stream)
	if stream.IsRequest() {
		c.handleRequest(ctx, message)
	} else if stream.IsStream() {
		c.handleStream(ctx, stream.Type, message)
	}
}

func (c *Client) handleRequest(ctx context.Context, message []byte) {
	var res ClientResponse
	c.unmarshalMessage(ctx, message, &res)
	responseChan, ok := c.lookupPendingResponse(res.ID)
	if !ok {
		return
	}

	// Non-blocking send: drops duplicate or late responses for the same id
	// rather than blocking the read loop.
	select {
	case responseChan <- &res:
	default:
	}
}

func (c *Client) unmarshalMessage(ctx context.Context, message []byte, v any) {
	if err := json.Unmarshal(message, v); err != nil {
		c.reportError(ctx, err)
	}
}

func (c *Client) handleStream(ctx context.Context, t streamtypes.Type, message []byte) {
	switch t {
	case streamtypes.LedgerStreamType:
		var ledger streamtypes.LedgerStream
		c.unmarshalMessage(ctx, message, &ledger)
		c.reportLedgerClosed(ctx, &ledger)
	case streamtypes.TransactionStreamType:
		var transactionStream streamtypes.TransactionStream
		c.unmarshalMessage(ctx, message, &transactionStream)
		c.reportTransaction(ctx, &transactionStream)
	case streamtypes.ValidationStreamType:
		var validation streamtypes.ValidationStream
		c.unmarshalMessage(ctx, message, &validation)
		c.reportValidationReceived(ctx, &validation)
	case streamtypes.PeerStatusStreamType:
		var peerStatus streamtypes.PeerStatusStream
		c.unmarshalMessage(ctx, message, &peerStatus)
		c.reportPeerStatusChange(ctx, &peerStatus)
	case streamtypes.ConsensusStreamType:
		var consensus streamtypes.ConsensusStream
		c.unmarshalMessage(ctx, message, &consensus)
		c.reportConsensusPhase(ctx, &consensus)
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
		message, err := c.conn.ReadMessage()
		select {
		case <-ctx.Done():
			return
		default:
		}

		switch {
		case ws.IsCloseError(err) || ws.IsUnexpectedCloseError(err):
			if retryCount >= maxRetries {
				c.reportError(ctx, ErrMaxReconnectionAttemptsReached{
					Attempts: maxRetries,
				})
				return
			}
			retryCount++
			select {
			case <-ctx.Done():
				return
			default:
			}
			connErr := c.conn.Connect()
			if connErr != nil {
				c.reportError(ctx, connErr)
				return
			}
		case err != nil:
			c.reportError(ctx, err)
			return
		default:
			// Send the message to the channel
			c.handleMessage(ctx, message)
			// Reset retry count on successful message
			retryCount = 0
		}
	}
}

// getSignedTx ensures the transaction is fully signed and returns the transaction blob.
// If the transaction is already signed, it encodes and returns it. Otherwise, it autofills (if enabled)
// and signs the transaction using the provided wallet.
func (c *Client) getSignedTx(tx transaction.FlatTransaction, autofill bool, wallet *wallet.Wallet) (string, error) {
	// Check if the transaction is already signed: both fields must be non-empty.
	sig, sigOk := tx["TxSignature"].(string)
	pubKey, pubKeyOk := tx["SigningPubKey"].(string)
	if sigOk && sig != "" && pubKeyOk && pubKey != "" {
		blob, err := binarycodec.Encode(tx)
		if err != nil {
			return "", err
		}
		return blob, nil
	}

	// If not signed, ensure a wallet is provided.
	if wallet == nil {
		return "", ErrMissingWallet
	}

	// Optionally autofill the transaction.
	if autofill {
		if err := c.Autofill(&tx); err != nil {
			return "", err
		}
	}

	// Sign the transaction.
	txBlob, _, err := wallet.Sign(tx)
	if err != nil {
		return "", err
	}
	return txBlob, nil
}

// fetchOwnerReserveFee fetches the owner reserve fee from the server state.
// Replicates the JavaScript fetchOwnerReserveFee function.
func (c *Client) fetchOwnerReserveFee() (uint64, error) {
	response, err := c.GetServerState(&server.StateRequest{})
	if err != nil {
		return 0, err
	}

	reserveInc := response.State.ValidatedLedger.ReserveInc
	if reserveInc == 0 {
		return 0, ErrCouldNotFetchOwnerReserve
	}

	return uint64(reserveInc), nil
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
			LedgerIndex: common.LedgerTitle("current"),
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
		LedgerIndex: common.LedgerTitle("current"),
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
// Replicates the JavaScript logic for Batch transaction fee calculation.
func (c *Client) calculateBatchFees(tx *transaction.FlatTransaction) (uint64, error) {
	var totalFees uint64

	// Get RawTransactions from the batch transaction
	rawTransactions, ok := (*tx)["RawTransactions"].([]map[string]any)
	if !ok {
		return 0, ErrRawTransactionsFieldMissing
	}

	// Iterate through each raw transaction
	for _, rawTx := range rawTransactions {
		// Extract the actual transaction from the wrapper
		innerTx, ok := rawTx["RawTransaction"].(map[string]any)
		if !ok {
			return 0, ErrRawTransactionFieldMissing
		}

		// Calculate fee for this inner transaction (no multi-signing for inner transactions)
		innerTxFlat := transaction.FlatTransaction(innerTx)
		err := c.calculateFeePerTransactionType(&innerTxFlat, 0)
		if err != nil {
			return 0, err
		}

		// Extract the calculated fee
		feeStr, ok := innerTx["Fee"].(string)
		if !ok {
			return 0, ErrFeeFieldMissing
		}

		innerTx["Fee"] = "0"

		// Convert fee string to uint64 and add to total
		feeUint, err := strconv.ParseUint(feeStr, 10, 64)
		if err != nil {
			return 0, ErrFailedToParseFee{
				Fee: feeStr,
				Err: err,
			}
		}

		totalFees += feeUint
	}

	return totalFees, nil
}

func (c *Client) autofillRawTransactions(tx *transaction.FlatTransaction) error {
	needsNetworkID, err := c.txNeedsNetworkID()
	if err != nil {
		return err
	}

	rawTxs, ok := (*tx)["RawTransactions"].([]map[string]any)
	if !ok {
		return ErrRawTransactionsFieldIsNotAnArray
	}

	accountSeq := make(map[string]uint32, len(rawTxs))

	for _, rawTx := range rawTxs {
		innerRawTx, ok := rawTx["RawTransaction"].(map[string]any)
		if !ok {
			return ErrRawTransactionFieldIsNotAnObject
		}

		// Validate `Fee` field
		if innerRawTx["Fee"] == nil {
			innerRawTx["Fee"] = "0"
		} else if innerRawTx["Fee"] != "0" {
			return types.ErrBatchInnerTransactionInvalid
		}

		// Validate `SigningPubKey` field
		if innerRawTx["SigningPubKey"] == nil {
			innerRawTx["SigningPubKey"] = ""
		} else if innerRawTx["SigningPubKey"] != "" {
			return ErrSigningPubKeyFieldMustBeEmpty
		}

		// Validate `TxnSignature` field
		if innerRawTx["TxnSignature"] != nil {
			return ErrTxnSignatureFieldMustBeEmpty
		}
		if innerRawTx["Signers"] != nil {
			return ErrSignersFieldMustBeEmpty
		}

		// Validate `NetworkID` field
		if innerRawTx["NetworkID"] == nil && needsNetworkID {
			innerRawTx["NetworkID"] = c.NetworkID
		}

		// Validate `Sequence` field
		if innerRawTx["Sequence"] == nil && innerRawTx["TicketSequence"] == nil {

			acc, ok := innerRawTx["Account"].(string)
			if !ok {
				return ErrAccountFieldIsNotAString
			}

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

// isNotLaterRippledVersion determines whether the source rippled version is not later than the target rippled version.
// Example usage: isNotLaterRippledVersion("1.10.0", "1.11.0") returns true.
//
//	isNotLaterRippledVersion("1.10.0", "1.10.0-b1") returns false.
func isNotLaterRippledVersion(source, target string) bool {
	if source == target {
		return true
	}

	sourceDecomp := strings.Split(source, ".")
	targetDecomp := strings.Split(target, ".")

	if len(sourceDecomp) < 3 || len(targetDecomp) < 3 {
		return false
	}

	sourceMajor, err := strconv.Atoi(sourceDecomp[0])
	if err != nil {
		return false
	}
	sourceMinor, err := strconv.Atoi(sourceDecomp[1])
	if err != nil {
		return false
	}
	targetMajor, err := strconv.Atoi(targetDecomp[0])
	if err != nil {
		return false
	}
	targetMinor, err := strconv.Atoi(targetDecomp[1])
	if err != nil {
		return false
	}

	// Compare major version
	if sourceMajor != targetMajor {
		return sourceMajor < targetMajor
	}

	// Compare minor version
	if sourceMinor != targetMinor {
		return sourceMinor < targetMinor
	}

	sourcePatch := strings.Split(sourceDecomp[2], "-")
	targetPatch := strings.Split(targetDecomp[2], "-")

	sourcePatchVersion, err := strconv.Atoi(sourcePatch[0])
	if err != nil {
		return false
	}
	targetPatchVersion, err := strconv.Atoi(targetPatch[0])
	if err != nil {
		return false
	}

	// Compare patch version
	if sourcePatchVersion != targetPatchVersion {
		return sourcePatchVersion < targetPatchVersion
	}

	// Compare release version
	if len(sourcePatch) != len(targetPatch) {
		return len(sourcePatch) > len(targetPatch)
	}

	if len(sourcePatch) == 2 {
		// Compare different release types
		if !strings.HasPrefix(sourcePatch[1], string(targetPatch[1][0])) {
			return sourcePatch[1] < targetPatch[1]
		}

		// Compare beta version
		if strings.HasPrefix(sourcePatch[1], "b") {
			sourceBeta, err := strconv.Atoi(sourcePatch[1][1:])
			if err != nil {
				return false
			}
			targetBeta, err := strconv.Atoi(targetPatch[1][1:])
			if err != nil {
				return false
			}
			return sourceBeta < targetBeta
		}

		// Compare rc version
		if strings.HasPrefix(sourcePatch[1], "rc") {
			sourceRC, err := strconv.Atoi(sourcePatch[1][2:])
			if err != nil {
				return false
			}
			targetRC, err := strconv.Atoi(targetPatch[1][2:])
			if err != nil {
				return false
			}
			return sourceRC < targetRC
		}
	}

	return false
}

// txNeedsNetworkID determines if the transaction required a networkID to be valid.
// Transaction needs networkID if later than restricted ID and build version is >= 1.11.0
func (c *Client) txNeedsNetworkID() (bool, error) {
	if c.NetworkID != 0 && c.NetworkID > RestrictedNetworks {
		res, err := c.GetServerInfo(&server.InfoRequest{})
		if err != nil {
			return false, err
		}

		if res.Info.BuildVersion != "" {
			return isNotLaterRippledVersion(RequiredNetworkIDVersion, res.Info.BuildVersion), nil
		}
	}
	return false, nil
}
