package ledger_test

import (
	"encoding/json"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	ledgerentry "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	ledgerquery "github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
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

func objectSelector[T any](object T) ledgerquery.EntrySelector[T] {
	return ledgerquery.EntrySelector[T]{Object: &object}
}

func TestEntryRequestSelectors(t *testing.T) {
	subIndex := uint64(0)
	bridge := ledgerquery.BridgeSelector{
		IssuingChainDoor:  accountA,
		IssuingChainIssue: ledgerentry.Asset{Currency: "XRP"},
		LockingChainDoor:  accountB,
		LockingChainIssue: ledgerentry.Asset{Currency: "USD", Issuer: accountC},
	}

	tests := []struct {
		name     string
		request  ledgerquery.EntryRequest
		expected string
	}{
		{
			name: "raw index with include deleted",
			request: ledgerquery.EntryRequest{
				Index:          entryIndex,
				LedgerIndex:    common.Validated,
				IncludeDeleted: true,
			},
			expected: `{"index":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4","ledger_index":"validated","include_deleted":true}`,
		},
		{
			name:     "account root",
			request:  ledgerquery.EntryRequest{AccountRoot: accountA},
			expected: `{"account_root":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn"}`,
		},
		{
			name: "amm object",
			request: ledgerquery.EntryRequest{AMM: objectSelector(ledgerquery.AMMSelectorFields{
				Asset:  ledgerentry.Asset{Currency: "XRP"},
				Asset2: ledgerentry.Asset{Currency: "TST", Issuer: accountC},
			})},
			expected: `{"amm":{"asset":{"currency":"XRP"},"asset2":{"currency":"TST","issuer":"rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"}}}`,
		},
		{
			name:     "amm index",
			request:  ledgerquery.EntryRequest{AMM: ledgerquery.AMMSelector{Index: entryIndex}},
			expected: `{"amm":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name: "bridge",
			request: ledgerquery.EntryRequest{
				BridgeAccount: accountB,
				Bridge:        bridge,
			},
			expected: `{"bridge_account":"rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW","bridge":{"IssuingChainDoor":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","IssuingChainIssue":{"currency":"XRP"},"LockingChainDoor":"rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW","LockingChainIssue":{"currency":"USD","issuer":"rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"}}}`,
		},
		{
			name:     "check",
			request:  ledgerquery.EntryRequest{Check: entryIndex},
			expected: `{"check":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name: "credential object",
			request: ledgerquery.EntryRequest{Credential: objectSelector(ledgerquery.CredentialSelectorFields{
				Subject:        accountA,
				Issuer:         accountB,
				CredentialType: types.CredentialType("746573742D63726564656E7469616C"),
			})},
			expected: `{"credential":{"subject":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","issuer":"rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW","credential_type":"746573742D63726564656E7469616C"}}`,
		},
		{
			name:     "credential index",
			request:  ledgerquery.EntryRequest{Credential: ledgerquery.CredentialSelector{Index: entryIndex}},
			expected: `{"credential":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name: "delegate object",
			request: ledgerquery.EntryRequest{Delegate: objectSelector(ledgerquery.DelegateSelectorFields{
				Account:   accountA,
				Authorize: accountB,
			})},
			expected: `{"delegate":{"account":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","authorize":"rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW"}}`,
		},
		{
			name:     "delegate index",
			request:  ledgerquery.EntryRequest{Delegate: ledgerquery.DelegateSelector{Index: entryIndex}},
			expected: `{"delegate":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name: "deposit preauth account object",
			request: ledgerquery.EntryRequest{DepositPreauth: objectSelector(ledgerquery.DepositPreauthSelectorFields{
				Owner:      accountA,
				Authorized: accountB,
			})},
			expected: `{"deposit_preauth":{"owner":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","authorized":"rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW"}}`,
		},
		{
			name: "deposit preauth credentials object",
			request: ledgerquery.EntryRequest{DepositPreauth: objectSelector(ledgerquery.DepositPreauthSelectorFields{
				Owner: accountA,
				AuthorizedCredentials: []ledgerquery.DepositPreauthCredential{{
					Issuer:         accountB,
					CredentialType: types.CredentialType("4B5943"),
				}},
			})},
			expected: `{"deposit_preauth":{"owner":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","authorized_credentials":[{"issuer":"rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW","credential_type":"4B5943"}]}}`,
		},
		{
			name:     "deposit preauth index",
			request:  ledgerquery.EntryRequest{DepositPreauth: ledgerquery.DepositPreauthSelector{Index: entryIndex}},
			expected: `{"deposit_preauth":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name:     "did",
			request:  ledgerquery.EntryRequest{DID: accountA},
			expected: `{"did":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn"}`,
		},
		{
			name: "directory owner object",
			request: ledgerquery.EntryRequest{Directory: objectSelector(ledgerquery.DirectorySelectorFields{
				Owner:    accountA,
				SubIndex: &subIndex,
			})},
			expected: `{"directory":{"owner":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","sub_index":0}}`,
		},
		{
			name: "directory root object",
			request: ledgerquery.EntryRequest{Directory: objectSelector(ledgerquery.DirectorySelectorFields{
				DirRoot:  entryIndex,
				SubIndex: &subIndex,
			})},
			expected: `{"directory":{"dir_root":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4","sub_index":0}}`,
		},
		{
			name:     "directory index",
			request:  ledgerquery.EntryRequest{Directory: ledgerquery.DirectorySelector{Index: entryIndex}},
			expected: `{"directory":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name: "escrow object",
			request: ledgerquery.EntryRequest{Escrow: objectSelector(ledgerquery.EscrowSelectorFields{
				Owner: accountA,
				Seq:   126,
			})},
			expected: `{"escrow":{"owner":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","seq":126}}`,
		},
		{
			name:     "escrow index",
			request:  ledgerquery.EntryRequest{Escrow: ledgerquery.EscrowSelector{Index: entryIndex}},
			expected: `{"escrow":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name:     "mpt issuance",
			request:  ledgerquery.EntryRequest{MPTIssuance: types.MPTIssuanceID(mptID)},
			expected: `{"mpt_issuance":"05EECEBE97A7D635DE2393068691A015FED5A89AD203F5AA"}`,
		},
		{
			name: "mptoken object",
			request: ledgerquery.EntryRequest{MPToken: objectSelector(ledgerquery.MPTokenSelectorFields{
				MPTIssuanceID: types.MPTIssuanceID(mptID),
				Account:       accountA,
			})},
			expected: `{"mptoken":{"mpt_issuance_id":"05EECEBE97A7D635DE2393068691A015FED5A89AD203F5AA","account":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn"}}`,
		},
		{
			name:     "mptoken index",
			request:  ledgerquery.EntryRequest{MPToken: ledgerquery.MPTokenSelector{Index: entryIndex}},
			expected: `{"mptoken":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name:     "nft page",
			request:  ledgerquery.EntryRequest{NFTPage: entryIndex},
			expected: `{"nft_page":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name: "offer object",
			request: ledgerquery.EntryRequest{Offer: objectSelector(ledgerquery.OfferSelectorFields{
				Account: accountA,
				Seq:     359,
			})},
			expected: `{"offer":{"account":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","seq":359}}`,
		},
		{
			name:     "offer index",
			request:  ledgerquery.EntryRequest{Offer: ledgerquery.OfferSelector{Index: entryIndex}},
			expected: `{"offer":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name:     "payment channel",
			request:  ledgerquery.EntryRequest{PaymentChannel: entryIndex},
			expected: `{"payment_channel":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name: "ripple state",
			request: ledgerquery.EntryRequest{RippleState: ledgerquery.RippleStateSelector{
				Accounts: [2]types.Address{accountA, accountB},
				Currency: "USD",
			}},
			expected: `{"ripple_state":{"accounts":["rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW"],"currency":"USD"}}`,
		},
		{
			name: "ticket object",
			request: ledgerquery.EntryRequest{Ticket: objectSelector(ledgerquery.TicketSelectorFields{
				Account:   accountA,
				TicketSeq: 389,
			})},
			expected: `{"ticket":{"account":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","ticket_seq":389}}`,
		},
		{
			name:     "ticket index",
			request:  ledgerquery.EntryRequest{Ticket: ledgerquery.TicketSelector{Index: entryIndex}},
			expected: `{"ticket":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name: "xchain owned claim id object",
			request: ledgerquery.EntryRequest{XChainOwnedClaimID: objectSelector(ledgerquery.XChainOwnedClaimIDSelectorFields{
				BridgeSelector:     bridge,
				XChainOwnedClaimID: 1,
			})},
			expected: `{"xchain_owned_claim_id":{"IssuingChainDoor":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","IssuingChainIssue":{"currency":"XRP"},"LockingChainDoor":"rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW","LockingChainIssue":{"currency":"USD","issuer":"rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"},"xchain_owned_claim_id":1}}`,
		},
		{
			name:     "xchain owned claim id index",
			request:  ledgerquery.EntryRequest{XChainOwnedClaimID: ledgerquery.XChainOwnedClaimIDSelector{Index: entryIndex}},
			expected: `{"xchain_owned_claim_id":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name: "xchain owned create account claim id object",
			request: ledgerquery.EntryRequest{XChainOwnedCreateAccountClaimID: objectSelector(ledgerquery.XChainOwnedCreateAccountClaimIDSelectorFields{
				BridgeSelector:                  bridge,
				XChainOwnedCreateAccountClaimID: 1,
			})},
			expected: `{"xchain_owned_create_account_claim_id":{"IssuingChainDoor":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","IssuingChainIssue":{"currency":"XRP"},"LockingChainDoor":"rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW","LockingChainIssue":{"currency":"USD","issuer":"rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"},"xchain_owned_create_account_claim_id":1}}`,
		},
		{
			name:     "sponsorship index",
			request:  ledgerquery.EntryRequest{Sponsorship: ledgerquery.SponsorshipSelector{Index: entryIndex}},
			expected: `{"sponsorship":"7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"}`,
		},
		{
			name:     "sponsorship pair",
			request:  ledgerquery.EntryRequest{Sponsorship: objectSelector(ledgerquery.SponsorshipSelectorFields{Sponsor: accountA, Sponsee: accountB})},
			expected: `{"sponsorship":{"sponsor":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","sponsee":"rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW"}}`,
		},
		{
			name:     "sponsorship incomplete pair leaves semantics to server",
			request:  ledgerquery.EntryRequest{Sponsorship: objectSelector(ledgerquery.SponsorshipSelectorFields{Sponsor: accountA})},
			expected: `{"sponsorship":{"sponsor":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","sponsee":""}}`,
		},
		{
			name:     "sponsorship malformed ID leaves semantics to server",
			request:  ledgerquery.EntryRequest{Sponsorship: ledgerquery.SponsorshipSelector{Index: "invalid"}},
			expected: `{"sponsorship":"invalid"}`,
		},
		{
			name:     "xchain owned create account claim id index",
			request:  ledgerquery.EntryRequest{XChainOwnedCreateAccountClaimID: ledgerquery.XChainOwnedCreateAccountClaimIDSelector{Index: entryIndex}},
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
	ammObject := ledgerquery.AMMSelectorFields{
		Asset:  ledgerentry.Asset{Currency: "XRP"},
		Asset2: ledgerentry.Asset{Currency: "USD", Issuer: accountC},
	}
	bridge := ledgerquery.BridgeSelector{
		IssuingChainDoor:  accountA,
		IssuingChainIssue: ledgerentry.Asset{Currency: "XRP"},
		LockingChainDoor:  accountB,
		LockingChainIssue: ledgerentry.Asset{Currency: "XRP"},
	}

	tests := []struct {
		name     string
		request  ledgerquery.EntryRequest
		expected error
	}{
		{name: "zero selectors", request: ledgerquery.EntryRequest{}, expected: ledgerquery.ErrInvalidEntryRequest},
		{name: "sponsorship conflicts with index", request: ledgerquery.EntryRequest{Index: entryIndex, Sponsorship: ledgerquery.SponsorshipSelector{Index: entryIndex}}, expected: ledgerquery.ErrInvalidEntryRequest},
		{name: "sponsorship has both forms", request: ledgerquery.EntryRequest{Sponsorship: ledgerquery.SponsorshipSelector{Index: entryIndex, Object: &ledgerquery.SponsorshipSelectorFields{Sponsor: accountA, Sponsee: accountB}}}, expected: ledgerquery.ErrInvalidEntrySelector},
		{
			name:     "multiple selectors",
			request:  ledgerquery.EntryRequest{Index: entryIndex, Check: entryIndex},
			expected: ledgerquery.ErrInvalidEntryRequest,
		},
		{
			name:     "bridge without bridge account",
			request:  ledgerquery.EntryRequest{Bridge: bridge},
			expected: ledgerquery.ErrInvalidBridgeSelector,
		},
		{
			name:     "unpaired bridge account on another selector",
			request:  ledgerquery.EntryRequest{Index: entryIndex, BridgeAccount: accountA},
			expected: ledgerquery.ErrInvalidBridgeSelector,
		},
		{
			name: "selector with index and object",
			request: ledgerquery.EntryRequest{AMM: ledgerquery.AMMSelector{
				Index:  entryIndex,
				Object: &ammObject,
			}},
			expected: ledgerquery.ErrInvalidEntrySelector,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.ErrorIs(t, tt.request.Validate(), tt.expected)
		})
	}
}

type entryResponseVariantsFixture struct {
	name string
	want ledgerquery.EntryResponse
	json string
}

func entryResponseVariantsFixtures() []entryResponseVariantsFixture {
	return []entryResponseVariantsFixture{
		{
			name: "json node",
			want: ledgerquery.EntryResponse{
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
			json: `{
				"index":"13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
				"ledger_hash":"31850E8E48E76D1064651DF39DF4E9542E8C90A9A9B629F4DE339EB3FA74F726",
				"ledger_index":61966146,
				"node":{"Account":"rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn","Balance":"424021949","LedgerEntryType":"AccountRoot"},
				"deleted_ledger_index":61966150,
				"validated":true
			}`,
		},
		{
			name: "binary node",
			want: ledgerquery.EntryResponse{
				Index:       "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
				LedgerIndex: 61966146,
				NodeBinary:  "1100612200000000",
				Validated:   true,
			},
			json: `{
				"index":"13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
				"ledger_index":61966146,
				"node_binary":"1100612200000000",
				"validated":true
			}`,
		},
	}
}

func TestEntryResponseVariantsSerialize(t *testing.T) {
	for _, tt := range entryResponseVariantsFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.want)
			require.NoError(t, err)
			require.JSONEq(t, tt.json, string(encoded))
		})
	}
}

func TestEntryResponseVariantsJSONDecode(t *testing.T) {
	for _, tt := range entryResponseVariantsFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			var got ledgerquery.EntryResponse
			decoder := json.NewDecoder(strings.NewReader(tt.json))
			decoder.UseNumber()
			require.NoError(t, decoder.Decode(&got))
			require.Equal(t, tt.want, got)
		})
	}
}

func TestEntryResponseVariantsClientDecode(t *testing.T) {
	for _, tt := range entryResponseVariantsFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			var data map[string]any
			decoder := json.NewDecoder(strings.NewReader(tt.json))
			decoder.UseNumber()
			require.NoError(t, decoder.Decode(&data))
			var got ledgerquery.EntryResponse
			require.NoError(t, clientinternal.DecodeResultInto(data, &got))
			require.Equal(t, tt.want, got)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response ledgerquery.EntryResponse
			err := json.Unmarshal([]byte(tt.fixture), &response)
			require.ErrorIs(t, err, ledgerquery.ErrInvalidEntryResponse)
		})
	}
}
