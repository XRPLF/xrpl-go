package client

import (
	"context"

	"github.com/Peersyst/xrpl-go/xrpl/currency"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// confidentialFeeMultiplier is the base-fee factor of a confidential MPT
// transaction. rippled charges one base fee for the transaction itself plus an
// extra multiplier of nine, so the combined factor is ten.
const confidentialFeeMultiplier uint64 = 10

// Request is the minimal request contract the fee queries need. Its method set
// is the union of the RPC and WebSocket client request interfaces, so a value of
// this type is assignable to either client's request parameter.
type Request interface {
	Method() string
	Validate() error
	APIVersion() int
	SetAPIVersion(apiVersion int)
}

// RequestResultFunc issues a request over a client's transport and decodes the
// response result into result.
type RequestResultFunc func(ctx context.Context, req Request, result any) error

// FeeSettings carries the client fee configuration that autofill applies.
type FeeSettings struct {
	// Cushion multiplies the load-adjusted network fee so a transaction stays
	// payable when load rises between autofill and validation.
	Cushion float64
	// MaxFeeXRP caps the fee of every transaction that does not carry a fixed
	// special cost, expressed in XRP.
	MaxFeeXRP string
}

// CalculateFee sets the Fee field of a transaction and returns the drops it
// set, including special costs for EscrowFinish, owner-reserve transactions,
// Batch, confidential MPT transactions, LoanSet, and multisigning.
func CalculateFee(
	ctx context.Context,
	request RequestResultFunc,
	tx *transaction.FlatTransaction,
	nSigners uint64,
	settings FeeSettings,
) (currency.Drops, error) {
	maxFee, err := ParseFeeXRP(settings.MaxFeeXRP)
	if err != nil {
		return currency.Drops{}, err
	}

	// The network fee cannot change while one transaction is priced, so fetch it
	// once and reuse it for every inner Batch transaction.
	netFee, err := networkFeeDropsFor(ctx, request, settings.Cushion, maxFee)
	if err != nil {
		return currency.Drops{}, err
	}

	totalFee, err := calculateFeeExact(ctx, request, tx, nSigners, netFee, maxFee)
	if err != nil {
		return currency.Drops{}, err
	}

	// Round half-up once, at the end. rippled sums every base fee a transaction
	// owes, inner Batch transactions included, and scales that total for load in
	// one step, so this is the only rounding the fee sees and no factor can
	// scale a rounding error.
	totalFee = totalFee.RoundHalfUp()

	fee, err := totalFee.WholeString()
	if err != nil {
		return currency.Drops{}, err
	}
	(*tx)["Fee"] = fee
	return totalFee, nil
}

// calculateFeeExact returns the fee a transaction owes for the given network
// fee, keeping any fractional drop and leaving the Fee field untouched, so a
// Batch sums its inner fees at full precision and only the caller rounds.
func calculateFeeExact(
	ctx context.Context,
	request RequestResultFunc,
	tx *transaction.FlatTransaction,
	nSigners uint64,
	netFee currency.Drops,
	maxFee currency.Drops,
) (currency.Drops, error) {
	// baseFeeFactor counts the network base fees this transaction costs, and
	// extraFee holds the costs that are not a multiple of the network fee.
	// rippled sums the same factor, one base fee per signer included, and scales
	// the total for load in one step, so the factor multiplies the exact network
	// fee and the result is rounded only once.
	transactionType := tx.TxType()
	baseFeeFactor := confidentialFeeFactor(transactionType) + nSigners
	var extraFee currency.Drops
	isSpecialTxCost := transactionType == transaction.AccountDeleteTx ||
		transactionType == transaction.AMMCreateTx

	switch transactionType { //nolint:exhaustive // Only transaction types with nonstandard fees need cases.
	case transaction.EscrowFinishTx:
		if fulfillment, ok := (*tx)["Fulfillment"].(string); ok {
			fulfillmentBytesSize := (len(fulfillment) + 1) / 2
			baseFeeFactor = 33 + uint64(fulfillmentBytesSize)/16 + nSigners
		}
	case transaction.AccountDeleteTx, transaction.AMMCreateTx:
		reserveFee, reserveErr := ownerReserveFee(ctx, request)
		if reserveErr != nil {
			return currency.Drops{}, reserveErr
		}
		baseFeeFactor = nSigners
		extraFee = currency.DropsFromUint64(reserveFee)
	case transaction.BatchTx:
		rawTxFees, batchErr := batchFees(ctx, request, tx, netFee, maxFee)
		if batchErr != nil {
			return currency.Drops{}, batchErr
		}
		baseFeeFactor = 2 + nSigners
		extraFee = rawTxFees
	case transaction.LoanSetTx:
		counterPartySignersCount, signerErr := counterPartySignersCount(ctx, request, *tx)
		if signerErr != nil {
			return currency.Drops{}, signerErr
		}
		baseFeeFactor = 1 + counterPartySignersCount + nSigners
	}

	totalFee := netFee.Mul(baseFeeFactor).Add(extraFee)
	if !isSpecialTxCost {
		totalFee = totalFee.Min(maxFee)
	}
	return totalFee, nil
}

// networkFeeDropsFor calculates the exact current network fee for one base fee.
// The result keeps any fractional drop so callers can apply the transaction's
// base-fee factor before rounding.
func networkFeeDropsFor(
	ctx context.Context,
	request RequestResultFunc,
	cushion float64,
	maxFee currency.Drops,
) (currency.Drops, error) {
	var res server.InfoResponse
	if err := request(ctx, &server.InfoRequest{}, &res); err != nil {
		return currency.Drops{}, err
	}

	baseFeeXRP := res.Info.ValidatedLedger.BaseFeeXRP
	if baseFeeXRP == nil {
		return currency.Drops{}, ErrCouldNotGetBaseFeeXrp
	}

	return NetworkFeeDrops(
		*baseFeeXRP,
		res.Info.LoadFactor,
		cushion,
		maxFee,
	)
}

// confidentialFeeFactor returns the base-fee factor of a confidential MPT
// transaction, or one for all other transaction types.
func confidentialFeeFactor(txType transaction.TxType) uint64 {
	switch txType { //nolint:exhaustive // Only confidential transaction types use this fixed multiplier.
	case transaction.ConfidentialMPTClawbackTx,
		transaction.ConfidentialMPTConvertTx,
		transaction.ConfidentialMPTConvertBackTx,
		transaction.ConfidentialMPTMergeInboxTx,
		transaction.ConfidentialMPTSendTx:
		return confidentialFeeMultiplier
	default:
		return 1
	}
}

// ownerReserveFee fetches the owner reserve increment charged by transactions
// that consume one reserve.
func ownerReserveFee(ctx context.Context, request RequestResultFunc) (uint64, error) {
	var response server.StateResponse
	if err := request(ctx, &server.StateRequest{}, &response); err != nil {
		return 0, err
	}

	reserveInc := response.State.ValidatedLedger.ReserveInc
	if reserveInc == nil {
		return 0, ErrCouldNotFetchOwnerReserve
	}

	return *reserveInc, nil
}

// batchFees calculates the exact total fees for all inner transactions in a
// Batch and zeroes each inner Fee, which a Batch requires. The total keeps any
// fractional drop, because rippled sums the inner base fees into the Batch fee
// and scales that sum for load once.
func batchFees(
	ctx context.Context,
	request RequestResultFunc,
	tx *transaction.FlatTransaction,
	netFee currency.Drops,
	maxFee currency.Drops,
) (currency.Drops, error) {
	var totalFees currency.Drops

	rawTransactions, ok := (*tx)["RawTransactions"].([]map[string]any)
	if !ok {
		return currency.Drops{}, ErrRawTransactionsFieldMissing
	}

	for _, rawTx := range rawTransactions {
		innerTx, ok := rawTx["RawTransaction"].(map[string]any)
		if !ok {
			return currency.Drops{}, ErrRawTransactionFieldMissing
		}

		innerTxFlat := transaction.FlatTransaction(innerTx)
		if innerTxFlat.TxType() == transaction.BatchTx {
			return currency.Drops{}, types.ErrBatchNestedTransaction
		}
		innerFee, err := calculateFeeExact(ctx, request, &innerTxFlat, 0, netFee, maxFee)
		if err != nil {
			return currency.Drops{}, err
		}
		innerTx["Fee"] = "0"

		totalFees = totalFees.Add(innerFee)
	}

	return totalFees, nil
}

// counterPartySignersCount resolves how many signers the LoanSet counterparty
// uses, which sets the counterparty share of the transaction fee.
func counterPartySignersCount(
	ctx context.Context,
	request RequestResultFunc,
	tx transaction.FlatTransaction,
) (uint64, error) {
	var counterparty types.Address

	if cp, ok := tx["Counterparty"]; ok {
		if cpStr, ok := cp.(string); ok && cpStr != "" {
			counterparty = types.Address(cpStr)
		}
	}

	if counterparty == "" {
		loanBrokerID, ok := tx["LoanBrokerID"].(string)
		if !ok || loanBrokerID == "" {
			return 0, ErrLoanBrokerIDRequired
		}

		var res ledger.EntryResponse
		if err := request(ctx, &ledger.EntryRequest{
			Index:       loanBrokerID,
			LedgerIndex: common.LedgerTitle("validated"),
		}, &res); err != nil {
			return 0, err
		}

		owner, ok := res.Node["Owner"].(string)
		if !ok || owner == "" {
			return 0, ErrCouldNotFetchLoanBrokerOwner
		}
		counterparty = types.Address(owner)
	}

	var accountInfo account.InfoResponse
	if err := request(ctx, &account.InfoRequest{
		Account:     counterparty,
		LedgerIndex: common.LedgerTitle("validated"),
		SignerLists: true,
	}, &accountInfo); err != nil {
		return 0, err
	}

	if len(accountInfo.SignerLists) > 0 {
		return uint64(len(accountInfo.SignerLists[0].SignerEntries)), nil
	}

	return 1, nil
}
