package builder

import (
	"fmt"

	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// LedgerQuerier is the minimal interface needed to query ledger state.
// Both rpc.Client and websocket.Client satisfy this interface.
type LedgerQuerier interface {
	GetAccountInfo(req *account.InfoRequest) (*account.InfoResponse, error)
	GetLedgerEntry(req *ledger.EntryRequest) (*ledger.EntryResponse, error)
}

// getSequence fetches the account sequence number.
func getSequence(q LedgerQuerier, addr string) (uint32, error) {
	resp, err := q.GetAccountInfo(&account.InfoRequest{
		Account: types.Address(addr),
	})
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrLedgerQuery, err)
	}
	return resp.AccountData.Sequence, nil
}

// getIssuanceKeys fetches IssuerEncryptionKey and AuditorEncryptionKey from an MPTokenIssuance.
// Uses hash.MPTokenIssuance() to compute the ledger entry index.
func getIssuanceKeys(q LedgerQuerier, issuanceID string) (issuerKey, auditorKey string, err error) {
	index, err := xrplhash.MPTokenIssuance(issuanceID)
	if err != nil {
		return "", "", fmt.Errorf("%w: %w", ErrLedgerQuery, err)
	}

	resp, err := q.GetLedgerEntry(&ledger.EntryRequest{Index: index})
	if err != nil {
		return "", "", fmt.Errorf("%w: %w", ErrLedgerQuery, err)
	}

	if v, ok := resp.Node["IssuerEncryptionKey"].(string); ok {
		issuerKey = v
	}
	if issuerKey == "" {
		return "", "", ErrEncryptionKeyNotSet
	}
	if v, ok := resp.Node["AuditorEncryptionKey"].(string); ok {
		auditorKey = v
	}
	return issuerKey, auditorKey, nil
}

// mpTokenIndex computes the ledger entry index for an MPToken.
func mpTokenIndex(issuanceID, holder string) (string, error) {
	index, err := xrplhash.MPToken(issuanceID, holder)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrLedgerQuery, err)
	}
	return index, nil
}

// getMPTokenState fetches MPToken fields for a holder.
// Returns holderKey, balanceCt, balanceVersion. Returns ErrMPTokenNotFound if the entry does not exist.
func getMPTokenState(q LedgerQuerier, issuanceID, holder string) (holderKey, balanceCt string, balanceVersion uint32, err error) {
	index, err := mpTokenIndex(issuanceID, holder)
	if err != nil {
		return "", "", 0, err
	}

	resp, err := q.GetLedgerEntry(&ledger.EntryRequest{Index: index})
	if err != nil {
		return "", "", 0, fmt.Errorf("%w: %w", ErrMPTokenNotFound, err)
	}

	if v, ok := resp.Node["HolderEncryptionKey"].(string); ok {
		holderKey = v
	}
	if v, ok := resp.Node["ConfidentialBalanceSpending"].(string); ok {
		balanceCt = v
	}
	if v, ok := resp.Node["ConfidentialBalanceVersion"].(float64); ok {
		balanceVersion = uint32(v)
	}
	return holderKey, balanceCt, balanceVersion, nil
}

// getIssuerCiphertext fetches the IssuerEncryptedBalance from a holder's MPToken.
func getIssuerCiphertext(q LedgerQuerier, issuanceID, holder string) (string, error) {
	index, err := mpTokenIndex(issuanceID, holder)
	if err != nil {
		return "", err
	}

	resp, err := q.GetLedgerEntry(&ledger.EntryRequest{Index: index})
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrLedgerQuery, err)
	}

	ct, ok := resp.Node["IssuerEncryptedBalance"].(string)
	if !ok || ct == "" {
		return "", fmt.Errorf("%w: IssuerEncryptedBalance not found", ErrLedgerQuery)
	}
	return ct, nil
}
