package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/xrpl/currency"
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
)

const (
	// RestrictedNetworks is the threshold above which sidechains are expected to have network IDs.
	// Sidechains are expected to have network IDs above this.
	// Networks with ID above this restricted number are expected specify an accurate NetworkID field
	// in every transaction to that chain to prevent replay attacks.
	// Mainnet and testnet are exceptions. More context: https://github.com/XRPLF/rippled/pull/4370
	RestrictedNetworks = 1024
	// RequiredNetworkIDVersion is the minimum rippled version that requires NetworkID validation.
	RequiredNetworkIDVersion = "1.11.0"
)

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
		Params: [1]interface{}{reqParams},
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
func checkForError(res *http.Response) (Response, error) {

	var jr Response

	b, err := io.ReadAll(res.Body)
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
			return ErrMismatchedTag{
				Expected: addressField,
				Actual:   tagField,
			}
		}
		(*tx)[tagField] = tag
	}

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
	roundedFee := float32(math.Round(float64(fee)*math.Pow10(int(currency.MaxFractionLength)))) / float32(math.Pow10(int(currency.MaxFractionLength)))

	// Convert the rounded fee back to a string with NUM_DECIMAL_PLACES
	return fmt.Sprintf("%.*f", currency.MaxFractionLength, roundedFee), nil
}

// Calculates the fee per transaction type.
//
// Enhanced implementation that replicates xrpl.js calculateFeePerTransactionType logic,
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
		if baseFee < maxFeeUint {
			totalFee = baseFee
		} else {
			totalFee = maxFeeUint
		}
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

func (c *Client) waitForTransaction(txHash string, lastLedgerSequence uint32) (*requests.TxResponse, error) {
	var txResponse *requests.TxResponse
	i := 0

	for i < c.cfg.maxRetries {
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
		if err != nil {
			return nil, err
		}

		err = res.GetResult(&txResponse)
		if err != nil {
			return nil, err
		}

		// Check if the transaction has been included in the current ledger
		if txResponse.LedgerIndex.Int() >= int(lastLedgerSequence) {
			break
		}

		// Wait for the retry delay before retrying
		time.Sleep(c.cfg.retryDelay)
		i++
	}

	if txResponse == nil {
		return nil, ErrTransactionNotFound
	}

	return txResponse, nil
}

// getSignedTx ensures the transaction is fully signed and returns the transaction blob.
// If the transaction is already signed, it encodes and returns it. Otherwise, it autofills (if enabled)
// and signs the transaction using the provided wallet.
func (c *Client) getSignedTx(tx transaction.FlatTransaction, autofill bool, wallet *wallet.Wallet) (string, error) {
	// Check if the transaction is already signed: both fields must be non-empty.
	sig, sigOk := tx["TxnSignature"].(string)
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
