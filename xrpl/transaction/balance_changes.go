package transaction

import (
	"math/big"
	"strings"

	"github.com/Peersyst/xrpl-go/xrpl/currency"
	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// Balance represents a balance change with its amount, currency, and optional issuer.
type Balance struct {
	Value    string `json:"amount"`
	Currency string `json:"currency"`
	Issuer   string `json:"issuer,omitempty"`
}

type balanceChange struct {
	Account types.Address `json:"account"`
	Balance `json:"balance"`
}

// AccountBalanceChanges contains the account and its associated balance changes.
type AccountBalanceChanges struct {
	Account  types.Address `json:"account"`
	Balances []Balance     `json:"balances"`
}

type normalizedNode struct {
	NodeType          string
	LedgerEntryType   ledger.EntryType
	LedgerIndex       string
	NewFields         ledger.FlatLedgerObject
	FinalFields       ledger.FlatLedgerObject
	PreviousFields    ledger.FlatLedgerObject
	PreviousTxnID     string
	PreviousTxnLgrSeq uint64
}

func newNormalizedNode(node AffectedNode) *normalizedNode {
	switch {
	case node.CreatedNode != nil:
		return &normalizedNode{
			NodeType:          "CreatedNode",
			LedgerEntryType:   node.CreatedNode.LedgerEntryType,
			LedgerIndex:       node.CreatedNode.LedgerIndex,
			NewFields:         node.CreatedNode.NewFields,
			FinalFields:       nil,
			PreviousFields:    nil,
			PreviousTxnID:     "",
			PreviousTxnLgrSeq: 0,
		}
	case node.ModifiedNode != nil:
		return &normalizedNode{
			NodeType:          "ModifiedNode",
			LedgerEntryType:   node.ModifiedNode.LedgerEntryType,
			LedgerIndex:       node.ModifiedNode.LedgerIndex,
			NewFields:         nil,
			FinalFields:       node.ModifiedNode.FinalFields,
			PreviousFields:    node.ModifiedNode.PreviousFields,
			PreviousTxnID:     node.ModifiedNode.PreviousTxnID,
			PreviousTxnLgrSeq: node.ModifiedNode.PreviousTxnLgrSeq,
		}
	case node.DeletedNode != nil:
		return &normalizedNode{
			NodeType:        "DeletedNode",
			LedgerEntryType: node.DeletedNode.LedgerEntryType,
			LedgerIndex:     node.DeletedNode.LedgerIndex,
			FinalFields:     node.DeletedNode.FinalFields,
		}
	default:
		return nil
	}
}

// GetBalanceChanges returns the balance changes for each account based on transaction metadata.
func GetBalanceChanges(meta *TxObjMeta) ([]AccountBalanceChanges, error) {
	nodes := normalizeNodes(meta.AffectedNodes)

	balanceChanges := make([]balanceChange, 0, len(nodes))

	for _, node := range nodes {
		switch node.LedgerEntryType { //nolint:exhaustive // only AccountRoot and RippleState affect balances
		case ledger.AccountRootEntry:
			xrpBalance, err := getXRPQuantity(node)
			if err != nil {
				return nil, err
			}
			if xrpBalance != nil {
				balanceChanges = append(balanceChanges, *xrpBalance)
			}
		case ledger.RippleStateEntry:
			trustlineChanges, err := getTrustlineQuantity(node)
			if err != nil {
				return nil, err
			}
			if len(trustlineChanges) > 0 {
				balanceChanges = append(balanceChanges, trustlineChanges...)
			}
		default:
			continue
		}
	}

	return groupByAccount(balanceChanges), nil
}

func normalizeNodes(nodes []AffectedNode) []*normalizedNode {
	var normalizedNodes []*normalizedNode
	for _, node := range nodes {
		n := newNormalizedNode(node)
		if n == nil {
			continue
		}
		normalizedNodes = append(normalizedNodes, n)
	}
	return normalizedNodes
}

func getXRPQuantity(node *normalizedNode) (*balanceChange, error) {
	var account string
	if finalFieldsAccount, ok := node.FinalFields["Account"]; ok {
		account = finalFieldsAccount.(string)
	} else if newFieldsAccount, ok := node.NewFields["Account"]; ok {
		account = newFieldsAccount.(string)
	} else {
		return nil, errAccountNotFoundForXRPQuantity
	}

	value, err := computeBalanceChange(node)
	if err != nil {
		return nil, err
	}

	var isNegative bool
	if strings.HasPrefix(value, "-") {
		isNegative = true
		value = strings.TrimPrefix(value, "-")
	}

	xrpAmount, err := currency.DropsToXrp(value)
	if err != nil {
		return nil, err
	}

	if isNegative {
		xrpAmount = strings.Join([]string{"-", xrpAmount}, "")
	}

	return &balanceChange{
		Account: types.Address(account),
		Balance: Balance{
			Currency: "XRP",
			Value:    xrpAmount,
		},
	}, nil
}

func getTrustlineQuantity(node *normalizedNode) ([]balanceChange, error) {
	value, err := computeBalanceChange(node)
	if err != nil {
		return nil, err
	}

	var fields ledger.FlatLedgerObject
	if node.NewFields != nil {
		fields = node.NewFields
	} else {
		fields = node.FinalFields
	}

	lowLimitMap, ok := fields["LowLimit"].(map[string]any)
	if !ok {
		return nil, errLowLimitIssuerNotFound
	}
	lowLimitIssuer, ok := lowLimitMap["issuer"]
	if !ok {
		return nil, errLowLimitIssuerNotFound
	}
	highLimitMap, ok := fields["HighLimit"].(map[string]any)
	if !ok {
		return nil, errHighLimitIssuerNotFound
	}
	highLimitIssuer, ok := highLimitMap["issuer"]
	if !ok {
		return nil, errHighLimitIssuerNotFound
	}
	balanceMap, ok := fields["Balance"].(map[string]any)
	if !ok {
		return nil, errBalanceCurrencyNotFound
	}
	balanceCurrency, ok := balanceMap["currency"]
	if !ok {
		return nil, errBalanceCurrencyNotFound
	}

	result := balanceChange{
		Account: types.Address(lowLimitIssuer.(string)),
		Balance: Balance{
			Issuer:   highLimitIssuer.(string),
			Currency: balanceCurrency.(string),
			Value:    value,
		},
	}

	bigFloatValue, ok := new(big.Float).SetString(value)
	if !ok {
		return nil, errInvalidBalanceValue
	}
	negatedValue := new(big.Float).Neg(bigFloatValue)

	flippedResult := balanceChange{
		Account: types.Address(result.Issuer),
		Balance: Balance{
			Issuer:   result.Account.String(),
			Currency: result.Currency,
			Value:    negatedValue.String(),
		},
	}

	return []balanceChange{result, flippedResult}, nil
}

func computeBalanceChange(node *normalizedNode) (string, error) {
	newBalance, okNewBalance := node.NewFields["Balance"]
	previousBalance, okPreviousBalance := node.PreviousFields["Balance"]
	finalBalance, okFinalBalance := node.FinalFields["Balance"]

	var value *big.Float
	var ok bool
	switch {
	case okNewBalance:
		balanceValue, err := getValue(newBalance)
		if err != nil {
			return "", err
		}

		value, ok = new(big.Float).SetString(balanceValue)
		if !ok {
			return "", errInvalidBalanceValue
		}
	case okPreviousBalance && okFinalBalance:
		balanceValue, err := getValue(previousBalance)
		if err != nil {
			return "", err
		}

		previousBalanceBigDecimal, ok := new(big.Float).SetString(balanceValue)
		if !ok {
			return "", errInvalidBalanceValue
		}
		balanceValue, err = getValue(finalBalance)
		if err != nil {
			return "", err
		}

		finalBalanceBigInt, ok := new(big.Float).SetString(balanceValue)
		if !ok {
			return "", errInvalidBalanceValue
		}

		value = finalBalanceBigInt.Sub(finalBalanceBigInt, previousBalanceBigDecimal)
	default:
		return "", errBalanceNotFound
	}

	return value.String(), nil
}

func getValue(balance any) (string, error) {
	if value, ok := balance.(string); ok {
		return value, nil
	} else if balanceMap, ok := balance.(map[string]any); ok {
		return balanceMap["value"].(string), nil
	}
	return "", errInvalidBalanceValue
}

func groupByAccount(balanceChanges []balanceChange) []AccountBalanceChanges {
	accountBalances := make(map[string][]balanceChange)

	// Group balance changes by account address
	for _, change := range balanceChanges {
		account := change.Account.String()
		accountBalances[account] = append(accountBalances[account], change)
	}

	// Convert map back to slice
	result := make([]AccountBalanceChanges, 0)
	for account, changes := range accountBalances {
		if len(changes) > 0 {
			balances := make([]Balance, len(changes))
			for i, change := range changes {
				balances[i] = change.Balance
			}
			result = append(result, AccountBalanceChanges{
				Account:  types.Address(account),
				Balances: balances,
			})
		}
	}

	return result
}
