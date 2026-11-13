package transaction

import (
	"math/big"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// AMMBid bids on an Automated Market Maker's (AMM's) auction slot. If you win, you can trade against the AMM at a discounted fee until you are outbid or 24 hours have passed.
// If you are outbid before 24 hours have passed, you are refunded part of the cost of your bid based on how much time remains.
// If the AMM's trading fee is zero, you can still bid, but the auction slot provides no benefit unless the trading fee changes.
// You bid using the AMM's LP Tokens; the amount of a winning bid is returned to the AMM, decreasing the outstanding balance of LP Tokens.
// https://xrpl.org/docs/references/protocol/transactions/types/ammbid
//
// Example:
//
//	{
//	    "Account" : "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
//	    "Asset" : {
//	        "currency" : "XRP"
//	    },
//	    "Asset2" : {
//	        "currency" : "TST",
//	        "issuer" : "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"
//	    },
//	    "AuthAccounts" : [
//	        {
//	          "AuthAccount" : {
//	              "Account" : "rMKXGCbJ5d8LbrqthdG46q3f969MVK2Qeg"
//	          }
//	        },
//	        {
//	          "AuthAccount" : {
//	              "Account" : "rBepJuTLFJt3WmtLXYAxSjtBWAeQxVbncv"
//	          }
//	        }
//	    ],
//	    "BidMax" : {
//	        "currency" : "039C99CD9AB0B70B32ECDA51EAAE471625608EA2",
//	        "issuer" : "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
//	        "value" : "100"
//	    },
//	    "Fee" : "10",
//	    "Flags" : 2147483648,
//	    "Sequence" : 9,
//	    "TransactionType" : "AMMBid"
//	}
type AMMBid struct {
	BaseTx
	// The definition for one of the assets in the AMM's pool. In JSON, this is an object with currency and issuer fields (omit issuer for XRP).
	Asset ledger.Asset
	// The definition for the other asset in the AMM's pool. In JSON, this is an object with currency and issuer fields (omit issuer for XRP).
	Asset2 ledger.Asset
	// Pay at least this amount for the slot. Setting this value higher makes it harder for others to outbid you. If omitted, pay the minimum necessary to win the bid.
	BidMin types.CurrencyAmount `json:",omitempty"`
	// Pay at most this amount for the slot. If the cost to win the bid is higher than this amount, the transaction fails. If omitted, pay as much as necessary to win the bid.
	BidMax types.CurrencyAmount `json:",omitempty"`
	// A list of up to 4 additional accounts that you allow to trade at the discounted fee. This cannot include the address of the transaction sender. Each of these objects should be an Auth Account object.
	AuthAccounts []ledger.AuthAccounts `json:",omitempty"`
}

// TxType implements the TxType method for the AMMBid struct.
func (*AMMBid) TxType() TxType {
	return AMMBidTx
}

// Flatten implements the Flatten method for the AMMBid struct.
func (a *AMMBid) Flatten() FlatTransaction {
	// Add BaseTx fields
	flattened := a.BaseTx.Flatten()

	// Add AMMBid-specific fields
	flattened["TransactionType"] = a.TxType().String()

	flattened["Asset"] = a.Asset.Flatten()

	flattened["Asset2"] = a.Asset2.Flatten()

	if a.BidMin != nil {
		flattened["BidMin"] = a.BidMin.Flatten()
	}

	if a.BidMax != nil {
		flattened["BidMax"] = a.BidMax.Flatten()
	}

	if len(a.AuthAccounts) > 0 {
		authAccountsFlattened := make([]map[string]any, 0, len(a.AuthAccounts))

		for _, authAccount := range a.AuthAccounts {
			authAccountsFlattened = append(authAccountsFlattened, authAccount.Flatten())
		}

		flattened["AuthAccounts"] = authAccountsFlattened
	}

	return flattened
}

// Validate implements the Validate method for the AMMBid struct.
func (a *AMMBid) Validate() (bool, error) {
	_, err := a.BaseTx.Validate()
	if err != nil {
		return false, err
	}

	if ok, err := IsAsset(a.Asset); !ok {
		return false, err
	}

	if ok, err := IsAsset(a.Asset2); !ok {
		return false, err
	}

	if a.Asset.Kind() == ledger.AssetXRP && a.Asset2.Kind() == ledger.AssetXRP {
		return false, ErrAMMAtLeastOneAssetMustBeNonXRP
	}

	if sameAssetIssue(a.Asset, a.Asset2) {
		return false, ErrAMMInvalidTokens
	}

	if ok, err := IsAmount(a.BidMin, "BidMin", false); !ok {
		return false, err
	}
	if a.BidMin != nil && a.BidMin.IsZero() {
		return false, ErrInvalidTokenValue
	}

	if ok, err := IsAmount(a.BidMax, "BidMax", false); !ok {
		return false, err
	}
	if a.BidMax != nil && a.BidMax.IsZero() {
		return false, ErrInvalidTokenValue
	}

	// rippled AMMBid::preclaim returns tecAMM_INVALID_TOKENS for BidMin > BidMax, which
	// charges the fee. Both bids must name the LP token, so compare only bids that share an
	// issue. A bid with another issue gets temBAD_AMM_TOKENS, which is free.
	if a.BidMin != nil && a.BidMax != nil &&
		a.BidMin.Kind() == types.ISSUED &&
		sameIssue(a.BidMax, amountAsset(a.BidMin)) &&
		bidAbove(a.BidMin, a.BidMax) {
		return false, ErrAMMBidMinAboveBidMax
	}

	if ok, err := validateAuthAccounts(a.AuthAccounts, a.Account); !ok {
		return false, err
	}

	return true, nil
}

// Validate the AuthAccounts field.
func validateAuthAccounts(authAccounts []ledger.AuthAccounts, account types.Address) (bool, error) {
	if len(authAccounts) > 4 {
		return false, ErrAMMAuthAccountsTooMany
	}

	for i, authAccount := range authAccounts {
		if ok := addresscodec.IsValidAddress(authAccount.AuthAccount.Account.String()); !ok {
			return false, ErrInvalidAccount
		}
		if sameAccountAddress(authAccount.AuthAccount.Account, account) {
			return false, ErrTransactionInvalidField{Type: "AMMBid", Field: "AuthAccounts"}
		}
		for _, previous := range authAccounts[:i] {
			if sameAccountAddress(authAccount.AuthAccount.Account, previous.AuthAccount.Account) {
				return false, ErrTransactionInvalidField{Type: "AMMBid", Field: "AuthAccounts"}
			}
		}
	}

	return true, nil
}

// bidAbove reports whether bidMin is greater than bidMax. Both are validated IOU amounts,
// so they have at most 16 significant digits and compare as rippled's STAmount does.
func bidAbove(bidMin, bidMax types.CurrencyAmount) bool {
	minValue, _ := new(big.Rat).SetString(tokenAmountValue(bidMin).(types.IssuedCurrencyAmount).Value)
	maxValue, _ := new(big.Rat).SetString(tokenAmountValue(bidMax).(types.IssuedCurrencyAmount).Value)
	return minValue.Cmp(maxValue) > 0
}
