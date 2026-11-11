package transaction

import (
	"math"
	"math/big"
	"regexp"
	"strconv"
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
			PreviousFields:  node.DeletedNode.PreviousFields,
		}
	default:
		return nil
	}
}

// GetBalanceChanges returns the balance changes for each account based on transaction metadata.
// Affected AccountRoot and RippleState nodes whose Balance did not change
// are skipped. This includes nodes that do not surface a Balance field at
// all, net-zero deltas, and newly-created entries with Balance=0.
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
	value, hasBalanceChange, err := computeBalanceChange(node)
	if err != nil {
		return nil, err
	}
	if !hasBalanceChange {
		// Balance-neutral nodes are skipped before requiring Account.
		return nil, nil
	}

	var account string
	accountValue, ok := node.FinalFields["Account"]
	if !ok {
		accountValue = node.NewFields["Account"]
	}
	account, ok = accountValue.(string)
	if !ok {
		return nil, errAccountNotFoundForXRPQuantity
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
	value, hasBalanceChange, err := computeBalanceChange(node)
	if err != nil {
		return nil, err
	}
	if !hasBalanceChange {
		return nil, nil
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
	lowLimitIssuer, ok := lowLimitMap["issuer"].(string)
	if !ok {
		return nil, errLowLimitIssuerNotFound
	}
	highLimitMap, ok := fields["HighLimit"].(map[string]any)
	if !ok {
		return nil, errHighLimitIssuerNotFound
	}
	highLimitIssuer, ok := highLimitMap["issuer"].(string)
	if !ok {
		return nil, errHighLimitIssuerNotFound
	}
	balanceMap, ok := fields["Balance"].(map[string]any)
	if !ok {
		return nil, errBalanceCurrencyNotFound
	}
	balanceCurrency, ok := balanceMap["currency"].(string)
	if !ok {
		return nil, errBalanceCurrencyNotFound
	}

	result := balanceChange{
		Account: types.Address(lowLimitIssuer),
		Balance: Balance{
			Issuer:   highLimitIssuer,
			Currency: balanceCurrency,
			Value:    formatBalance(value),
		},
	}

	flippedResult := balanceChange{
		Account: types.Address(result.Issuer),
		Balance: Balance{
			Issuer:   result.Account.String(),
			Currency: result.Currency,
			Value:    formatBalance(negateBalance(value)),
		},
	}

	return []balanceChange{result, flippedResult}, nil
}

func computeBalanceChange(node *normalizedNode) (delta string, hasBalanceChange bool, err error) {
	newBalance, okNewBalance := node.NewFields["Balance"]
	previousBalance, okPreviousBalance := node.PreviousFields["Balance"]
	finalBalance, okFinalBalance := node.FinalFields["Balance"]

	var value *big.Rat
	switch {
	case okNewBalance:
		balanceValue, err := getValue(newBalance)
		if err != nil {
			return "", false, err
		}

		value, err = parseBalance(balanceValue)
		if err != nil {
			return "", false, errInvalidBalanceValue
		}
	case okPreviousBalance && okFinalBalance:
		balanceValue, err := getValue(previousBalance)
		if err != nil {
			return "", false, err
		}

		previousBalanceDecimal, err := parseBalance(balanceValue)
		if err != nil {
			return "", false, errInvalidBalanceValue
		}
		balanceValue, err = getValue(finalBalance)
		if err != nil {
			return "", false, err
		}

		finalBalanceDecimal, err := parseBalance(balanceValue)
		if err != nil {
			return "", false, errInvalidBalanceValue
		}

		value = finalBalanceDecimal.Sub(finalBalanceDecimal, previousBalanceDecimal)
	default:
		return "", false, nil
	}

	if value.Sign() == 0 {
		return "", false, nil
	}

	return decimalString(value), true, nil
}

func getValue(balance any) (string, error) {
	if value, ok := balance.(string); ok {
		return value, nil
	} else if balanceMap, ok := balance.(map[string]any); ok {
		if value, ok := balanceMap["value"].(string); ok {
			return value, nil
		}
	}
	return "", errInvalidBalanceValue
}

// XRPL amounts have a 16-digit IOU mantissa and bounded exponent; 256 bytes
// also accommodates expanded decimal representations of all valid amounts.
const maxBalanceLength = 256

var balancePattern = regexp.MustCompile(`^[+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE][+-]?[0-9]+)?$`)

func parseBalance(value string) (*big.Rat, error) {
	if len(value) > maxBalanceLength || !balancePattern.MatchString(value) {
		return nil, errInvalidBalanceValue
	}
	if i := strings.IndexAny(value, "eE"); i >= 0 {
		exponent, err := strconv.Atoi(value[i+1:])
		if err != nil || exponent < -256 || exponent > 256 {
			return nil, errInvalidBalanceValue
		}
	}
	result, ok := new(big.Rat).SetString(value)
	if !ok {
		return nil, errInvalidBalanceValue
	}
	return result, nil
}

// decimalString formats value, whose denominator is 2^a*5^b because every
// input is a decimal string, with exactly the fractional digits it needs.
func decimalString(value *big.Rat) string {
	denominator := value.Denom()
	oddBits := new(big.Int).Rsh(denominator, denominator.TrailingZeroBits()).BitLen()
	twos := denominator.BitLen() - oddBits
	// The odd part is 5^b, whose bit length L satisfies L-1 = floor(b*log2(5)),
	// so b = ceil((L-1)/log2(5)).
	fives := int(math.Ceil(float64(oddBits-1) / math.Log2(5)))
	scale := max(twos, fives)
	result := value.FloatString(scale)
	if scale > 0 {
		result = strings.TrimRight(strings.TrimRight(result, "0"), ".")
	}
	return result
}

func negateBalance(value string) string {
	if strings.HasPrefix(value, "-") {
		return value[1:]
	}
	return "-" + value
}

func formatBalance(value string) string {
	sign := ""
	if strings.HasPrefix(value, "-") {
		sign, value = "-", value[1:]
	}
	parts := strings.SplitN(value, ".", 2)
	integer := parts[0]
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	exponent := len(integer) - 1
	digits := integer + fraction
	if integer == "0" {
		leading := len(fraction) - len(strings.TrimLeft(fraction, "0"))
		exponent = -leading - 1
		digits = fraction[leading:]
	}
	if exponent > -7 && exponent < 21 {
		return sign + value
	}
	digits = strings.TrimRight(digits, "0")
	mantissa := digits[:1]
	if len(digits) > 1 {
		mantissa += "." + digits[1:]
	}
	if exponent >= 0 {
		return sign + mantissa + "e+" + strconv.Itoa(exponent)
	}
	return sign + mantissa + "e" + strconv.Itoa(exponent)
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
