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

	netFee, err := NetworkFeeDropsFor(ctx, request, settings.Cushion, maxFee)
	if err != nil {
		return currency.Drops{}, err
	}

	// baseFeeFactor counts the network base fees this transaction costs, and
	// extraFee holds the costs that are not a multiple of the network fee.
	// rippled sums the same factor, one base fee per signer included, and scales
	// the total for load in one step, so the factor multiplies the exact network
	// fee and the result is rounded only once.
	transactionType := tx.TxType()
	baseFeeFactor := ConfidentialFeeMultiplier(transactionType) + nSigners
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
		reserveFee, reserveErr := OwnerReserveFee(ctx, request)
		if reserveErr != nil {
			return currency.Drops{}, reserveErr
		}
		baseFeeFactor = nSigners
		extraFee = currency.DropsFromUint64(reserveFee)
	case transaction.BatchTx:
		rawTxFees, batchErr := BatchFees(ctx, request, tx, settings)
		if batchErr != nil {
			return currency.Drops{}, batchErr
		}
		baseFeeFactor = 2 + nSigners
		extraFee = rawTxFees
	case transaction.LoanSetTx:
		counterPartySignersCount, signerErr := CounterPartySignersCount(ctx, request, *tx)
		if signerErr != nil {
			return currency.Drops{}, signerErr
		}
		baseFeeFactor = 1 + counterPartySignersCount + nSigners
	}

	totalFee := netFee.Mul(baseFeeFactor).Add(extraFee)
	if !isSpecialTxCost {
		totalFee = totalFee.Min(maxFee)
	}
	// Round half-up once, at the end. rippled requires the load-scaled base fee
	// truncated to whole drops, and rounding half-up never falls below that, so
	// this pays the same whole drops as rounding each base fee individually
	// without letting the factor scale a rounding error.
	totalFee = totalFee.RoundHalfUp()

	fee, err := totalFee.WholeString()
	if err != nil {
		return currency.Drops{}, err
	}
	(*tx)["Fee"] = fee
	return totalFee, nil
}

// NetworkFeeDropsFor calculates the exact current network fee for one base fee.
// The result keeps any fractional drop so callers can apply the transaction's
// base-fee factor before rounding.
func NetworkFeeDropsFor(
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

// OwnerReserveFee fetches the owner reserve increment charged by transactions
// that consume one reserve.
func OwnerReserveFee(ctx context.Context, request RequestResultFunc) (uint64, error) {
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

// BatchFees calculates the total fees for all inner transactions in a Batch and
// zeroes each inner Fee, which a Batch requires.
func BatchFees(
	ctx context.Context,
	request RequestResultFunc,
	tx *transaction.FlatTransaction,
	settings FeeSettings,
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
		innerFee, err := CalculateFee(ctx, request, &innerTxFlat, 0, settings)
		if err != nil {
			return currency.Drops{}, err
		}
		innerTx["Fee"] = "0"

		totalFees = totalFees.Add(innerFee)
	}

	return totalFees, nil
}

// CounterPartySignersCount resolves how many signers the LoanSet counterparty
// uses, which sets the counterparty share of the transaction fee.
func CounterPartySignersCount(
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
