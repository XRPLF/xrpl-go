package ledger

import (
	"encoding/json"
	"testing"

	ledgerentry "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

const (
	entryIndex = "7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"
	accountA   = "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn"
	accountB   = "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW"
	accountC   = "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"
	mptID      = "05EECEBE97A7D635DE2393068691A015FED5A89AD203F5AA"
)

func objectSelector[T any](object T) EntrySelector[T] {
	return EntrySelector[T]{Object: &object}
}

func TestEntryRequestSelectors(t *testing.T) {
	subIndex := uint64(0)
	bridge := BridgeSelector{
		IssuingChainDoor:  accountA,
		IssuingChainIssue: ledgerentry.Asset{Currency: "XRP"},
		LockingChainDoor:  accountB,
		LockingChainIssue: ledgerentry.Asset{Currency: "USD", Issuer: accountC},
	}

	tests := []struct {
		name     string
		request  EntryRequest
		expected string
	}{
		{
			name: "raw index with include deleted",
			request: EntryRequest{
				Index:          entryIndex,
				LedgerIndex:    common.Validated,
				IncludeDeleted: true,
			},
			expected: `{"index":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4","ledger_index":"validated","include_deleted":true}`,
		},
		{
			name:     "account root",
			request:  EntryRequest{AccountRoot: accountA},
			expected: `{"account_root":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn"}`,
		},
		{
			name: "amm object",
			request: EntryRequest{AMM: objectSelector(AMMSelectorFields{
				Asset:  ledgerentry.Asset{Currency: "XRP"},
				Asset2: ledgerentry.Asset{Currency: "TST", Issuer: accountC},
			})},
			expected: `{"amm":{"asset":{"currency":"XRP"},"asset2":{"currency":"TST","issuer":"rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"}}}`,
		},
		{
			name:     "amm index",
			request:  EntryRequest{AMM: AMMSelector{Index: entryIndex}},
			expected: `{"amm":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name: "bridge",
			request: EntryRequest{
				BridgeAccount: accountB,
				Bridge:        bridge,
			},
			expected: `{"bridge_account":"rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW","bridge":{"IssuingChainDoor":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","IssuingChainIssue":{"currency":"XRP"},"LockingChainDoor":"rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW","LockingChainIssue":{"currency":"USD","issuer":"rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"}}}`,
		},
		{
			name:     "check",
			request:  EntryRequest{Check: entryIndex},
			expected: `{"check":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name: "credential object",
			request: EntryRequest{Credential: objectSelector(CredentialSelectorFields{
				Subject:        accountA,
				Issuer:         accountB,
				CredentialType: types.CredentialType("746573742D63726564656E7469616C"),
			})},
			expected: `{"credential":{"subject":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","issuer":"rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW","credential_type":"746573742D63726564656E7469616C"}}`,
		},
		{
			name:     "credential index",
			request:  EntryRequest{Credential: CredentialSelector{Index: entryIndex}},
			expected: `{"credential":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name: "delegate object",
			request: EntryRequest{Delegate: objectSelector(DelegateSelectorFields{
				Account:   accountA,
				Authorize: accountB,
			})},
			expected: `{"delegate":{"account":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","authorize":"rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW"}}`,
		},
		{
			name:     "delegate index",
			request:  EntryRequest{Delegate: DelegateSelector{Index: entryIndex}},
			expected: `{"delegate":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name: "deposit preauth account object",
			request: EntryRequest{DepositPreauth: objectSelector(DepositPreauthSelectorFields{
				Owner:      accountA,
				Authorized: accountB,
			})},
			expected: `{"deposit_preauth":{"owner":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","authorized":"rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW"}}`,
		},
		{
			name: "deposit preauth credentials object",
			request: EntryRequest{DepositPreauth: objectSelector(DepositPreauthSelectorFields{
				Owner: accountA,
				AuthorizedCredentials: []DepositPreauthCredential{{
					Issuer:         accountB,
					CredentialType: types.CredentialType("4B5943"),
				}},
			})},
			expected: `{"deposit_preauth":{"owner":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","authorized_credentials":[{"issuer":"rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW","credential_type":"4B5943"}]}}`,
		},
		{
			name:     "deposit preauth index",
			request:  EntryRequest{DepositPreauth: DepositPreauthSelector{Index: entryIndex}},
			expected: `{"deposit_preauth":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name:     "did",
			request:  EntryRequest{DID: accountA},
			expected: `{"did":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn"}`,
		},
		{
			name: "directory owner object",
			request: EntryRequest{Directory: objectSelector(DirectorySelectorFields{
				Owner:    accountA,
				SubIndex: &subIndex,
			})},
			expected: `{"directory":{"owner":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","sub_index":0}}`,
		},
		{
			name: "directory root object",
			request: EntryRequest{Directory: objectSelector(DirectorySelectorFields{
				DirRoot:  entryIndex,
				SubIndex: &subIndex,
			})},
			expected: `{"directory":{"dir_root":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4","sub_index":0}}`,
		},
		{
			name:     "directory index",
			request:  EntryRequest{Directory: DirectorySelector{Index: entryIndex}},
			expected: `{"directory":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name: "escrow object",
			request: EntryRequest{Escrow: objectSelector(EscrowSelectorFields{
				Owner: accountA,
				Seq:   126,
			})},
			expected: `{"escrow":{"owner":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","seq":126}}`,
		},
		{
			name:     "escrow index",
			request:  EntryRequest{Escrow: EscrowSelector{Index: entryIndex}},
			expected: `{"escrow":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name:     "mpt issuance",
			request:  EntryRequest{MPTIssuance: types.MPTIssuanceID(mptID)},
			expected: `{"mpt_issuance":"05EECEBE97A7D635DE2393068691A015FED5A89AD203F5AA"}`,
		},
		{
			name: "mptoken object",
			request: EntryRequest{MPToken: objectSelector(MPTokenSelectorFields{
				MPTIssuanceID: types.MPTIssuanceID(mptID),
				Account:       accountA,
			})},
			expected: `{"mptoken":{"mpt_issuance_id":"05EECEBE97A7D635DE2393068691A015FED5A89AD203F5AA","account":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn"}}`,
		},
		{
			name:     "mptoken index",
			request:  EntryRequest{MPToken: MPTokenSelector{Index: entryIndex}},
			expected: `{"mptoken":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name:     "nft page",
			request:  EntryRequest{NFTPage: entryIndex},
			expected: `{"nft_page":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name: "offer object",
			request: EntryRequest{Offer: objectSelector(OfferSelectorFields{
				Account: accountA,
				Seq:     359,
			})},
			expected: `{"offer":{"account":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","seq":359}}`,
		},
		{
			name:     "offer index",
			request:  EntryRequest{Offer: OfferSelector{Index: entryIndex}},
			expected: `{"offer":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name:     "payment channel",
			request:  EntryRequest{PaymentChannel: entryIndex},
			expected: `{"payment_channel":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name: "ripple state",
			request: EntryRequest{RippleState: RippleStateSelector{
				Accounts: [2]types.Address{accountA, accountB},
				Currency: "USD",
			}},
			expected: `{"ripple_state":{"accounts":["rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW"],"currency":"USD"}}`,
		},
		{
			name: "ticket object",
			request: EntryRequest{Ticket: objectSelector(TicketSelectorFields{
				Account:   accountA,
				TicketSeq: 389,
			})},
			expected: `{"ticket":{"account":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","ticket_seq":389}}`,
		},
		{
			name:     "ticket index",
			request:  EntryRequest{Ticket: TicketSelector{Index: entryIndex}},
			expected: `{"ticket":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name: "xchain owned claim id object",
			request: EntryRequest{XChainOwnedClaimID: objectSelector(XChainOwnedClaimIDSelectorFields{
				BridgeSelector:     bridge,
				XChainOwnedClaimID: 1,
			})},
			expected: `{"xchain_owned_claim_id":{"IssuingChainDoor":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","IssuingChainIssue":{"currency":"XRP"},"LockingChainDoor":"rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW","LockingChainIssue":{"currency":"USD","issuer":"rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"},"xchain_owned_claim_id":1}}`,
		},
		{
			name:     "xchain owned claim id index",
			request:  EntryRequest{XChainOwnedClaimID: XChainOwnedClaimIDSelector{Index: entryIndex}},
			expected: `{"xchain_owned_claim_id":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name: "xchain owned create account claim id object",
			request: EntryRequest{XChainOwnedCreateAccountClaimID: objectSelector(XChainOwnedCreateAccountClaimIDSelectorFields{
				BridgeSelector:                  bridge,
				XChainOwnedCreateAccountClaimID: 1,
			})},
			expected: `{"xchain_owned_create_account_claim_id":{"IssuingChainDoor":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","IssuingChainIssue":{"currency":"XRP"},"LockingChainDoor":"rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW","LockingChainIssue":{"currency":"USD","issuer":"rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"},"xchain_owned_create_account_claim_id":1}}`,
		},
		{
			name:     "xchain owned create account claim id index",
			request:  EntryRequest{XChainOwnedCreateAccountClaimID: XChainOwnedCreateAccountClaimIDSelector{Index: entryIndex}},
			expected: `{"xchain_owned_create_account_claim_id":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, tt.request.Validate())
			encoded, err := json.Marshal(tt.request)
			require.NoError(t, err)
			require.JSONEq(t, tt.expected, string(encoded))
		})
	}
}

func TestEntryRequestValidate(t *testing.T) {
	ammObject := AMMSelectorFields{
		Asset:  ledgerentry.Asset{Currency: "XRP"},
		Asset2: ledgerentry.Asset{Currency: "USD", Issuer: accountC},
	}
	bridge := BridgeSelector{
		IssuingChainDoor:  accountA,
		IssuingChainIssue: ledgerentry.Asset{Currency: "XRP"},
		LockingChainDoor:  accountB,
		LockingChainIssue: ledgerentry.Asset{Currency: "XRP"},
	}

	tests := []struct {
		name     string
		request  EntryRequest
		expected error
	}{
		{name: "zero selectors", request: EntryRequest{}, expected: ErrInvalidEntryRequest},
		{
			name:     "multiple selectors",
			request:  EntryRequest{Index: entryIndex, Check: entryIndex},
			expected: ErrInvalidEntryRequest,
		},
		{
			name:     "bridge without bridge account",
			request:  EntryRequest{Bridge: bridge},
			expected: ErrInvalidBridgeSelector,
		},
		{
			name:     "unpaired bridge account on another selector",
			request:  EntryRequest{Index: entryIndex, BridgeAccount: accountA},
			expected: ErrInvalidBridgeSelector,
		},
		{
			name: "selector with index and object",
			request: EntryRequest{AMM: AMMSelector{
				Index:  entryIndex,
				Object: &ammObject,
			}},
			expected: ErrInvalidEntrySelector,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.ErrorIs(t, tt.request.Validate(), tt.expected)
		})
	}
}

func TestEntryResponseVariants(t *testing.T) {
	tests := []struct {
		name     string
		fixture  string
		expected EntryResponse
	}{
		{
			name: "json node",
			fixture: `{
				"index":"13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
				"ledger_hash":"31850E8E48E76D1064651DF39DF4E9542E8C90A9A9B629F4DE339EB3FA74F726",
				"ledger_index":61966146,
				"node":{"Account":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","Balance":"424021949","LedgerEntryType":"AccountRoot"},
				"deleted_ledger_index":61966150,
				"validated":true
			}`,
			expected: EntryResponse{
				Index:       "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
				LedgerHash:  "31850E8E48E76D1064651DF39DF4E9542E8C90A9A9B629F4DE339EB3FA74F726",
				LedgerIndex: 61966146,
				Node: ledgerentry.FlatLedgerObject{
					"Account":         "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
					"Balance":         "424021949",
					"LedgerEntryType": "AccountRoot",
				},
				DeletedLedgerIndex: 61966150,
				Validated:          true,
			},
		},
		{
			name: "binary node",
			fixture: `{
				"index":"13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
				"ledger_index":61966146,
				"node_binary":"1100612200000000",
				"validated":true
			}`,
			expected: EntryResponse{
				Index:       "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
				LedgerIndex: 61966146,
				NodeBinary:  "1100612200000000",
				Validated:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response EntryResponse
			require.NoError(t, json.Unmarshal([]byte(tt.fixture), &response))
			require.Equal(t, tt.expected, response)

			encoded, err := json.Marshal(response)
			require.NoError(t, err)
			require.JSONEq(t, tt.fixture, string(encoded))
		})
	}
}

func TestEntryResponseRejectsInvalidVariants(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
	}{
		{name: "missing payload", fixture: `{"index":"ABC","validated":true}`},
		{name: "both payloads", fixture: `{"index":"ABC","node":{"LedgerEntryType":"Offer"},"node_binary":"1100","validated":true}`},
		{name: "null json node", fixture: `{"index":"ABC","node":null,"validated":true}`},
		{name: "empty json node", fixture: `{"index":"ABC","node":{},"validated":true}`},
		{name: "empty binary node", fixture: `{"index":"ABC","node_binary":"","validated":true}`},
		{name: "malformed binary node", fixture: `{"index":"ABC","node_binary":"not-hex","validated":true}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response EntryResponse
			err := json.Unmarshal([]byte(tt.fixture), &response)
			require.ErrorIs(t, err, ErrInvalidEntryResponse)
		})
	}
}
