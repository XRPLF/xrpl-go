package builder

import (
	ledgerentries "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
)

const (
	testIssuanceID  = "000004C463C52827307480341E3CB23A0710CC839EB58A0A"
	testAccount     = "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	testDestination = "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP"
)

// mockQuerier implements LedgerQuerier for testing.
type mockQuerier struct {
	accountSeq uint32
	entries    map[string]ledgerentries.FlatLedgerObject
}

func (m *mockQuerier) GetAccountInfo(_ *account.InfoRequest) (*account.InfoResponse, error) {
	return &account.InfoResponse{
		AccountData: ledgerentries.AccountRoot{Sequence: m.accountSeq},
	}, nil
}

func (m *mockQuerier) GetLedgerEntry(req *ledger.EntryRequest) (*ledger.EntryResponse, error) {
	node, ok := m.entries[req.Index]
	if !ok {
		return nil, ErrMPTokenNotFound
	}
	return &ledger.EntryResponse{Node: node}, nil
}

// buildIssuanceEntry builds a mock MPTokenIssuance flat entry.
func buildIssuanceEntry(issuerKey, auditorKey string) ledgerentries.FlatLedgerObject {
	entry := ledgerentries.FlatLedgerObject{
		"IssuerEncryptionKey": issuerKey,
	}
	if auditorKey != "" {
		entry["AuditorEncryptionKey"] = auditorKey
	}
	return entry
}

// buildMPTokenEntry builds a mock MPToken flat entry for a holder.
func buildMPTokenEntry(holderKey, balanceCt string, balanceVersion float64, issuerCt string) ledgerentries.FlatLedgerObject {
	entry := ledgerentries.FlatLedgerObject{}
	if holderKey != "" {
		entry["HolderEncryptionKey"] = holderKey
	}
	if balanceCt != "" {
		entry["ConfidentialBalanceSpending"] = balanceCt
	}
	if balanceVersion != 0 {
		entry["ConfidentialBalanceVersion"] = balanceVersion
	}
	if issuerCt != "" {
		entry["IssuerEncryptedBalance"] = issuerCt
	}
	return entry
}
