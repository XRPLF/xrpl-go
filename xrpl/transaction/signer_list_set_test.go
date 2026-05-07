package transaction

import (
	"encoding/hex"
	"math"
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignerListSet_TxType(t *testing.T) {
	entry := &SignerListSet{}
	assert.Equal(t, SignerListSetTx, entry.TxType())
}

func TestSignerListSet_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		entry    *SignerListSet
		expected string
	}{
		{
			name: "pass - with SignerEntries",
			entry: &SignerListSet{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:     types.XRPCurrencyAmount(12),
				},
				SignerQuorum: uint32(3),
				SignerEntries: []ledger.SignerEntryWrapper{
					{
						SignerEntry: ledger.SignerEntry{
							Account:      "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
							SignerWeight: 2,
						},
					},
					{
						SignerEntry: ledger.SignerEntry{
							Account:      "rUpy3eEg8rqjqfUoLeBnZkscbKbFsKXC3v",
							SignerWeight: 1,
						},
					},
					{
						SignerEntry: ledger.SignerEntry{
							Account:      "raKEEVSGnKSD9Zyvxu4z6Pqpm4ABH8FS6n",
							SignerWeight: 1,
						},
					},
				},
			},
			expected: `{
				"TransactionType": "SignerListSet",
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee": "12",
				"SignerQuorum": 3,
				"SignerEntries": [
					{
						"SignerEntry": {
							"Account": "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
							"SignerWeight": 2
						}
					},
					{
						"SignerEntry": {
							"Account": "rUpy3eEg8rqjqfUoLeBnZkscbKbFsKXC3v",
							"SignerWeight": 1
						}
					},
					{
						"SignerEntry": {
							"Account": "raKEEVSGnKSD9Zyvxu4z6Pqpm4ABH8FS6n",
							"SignerWeight": 1
						}
					}
				]
			}`,
		},
		{
			name: "pass - without SignerEntries",
			entry: &SignerListSet{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:     types.XRPCurrencyAmount(12),
				},
				SignerQuorum: uint32(0),
			},
			expected: `{
				"TransactionType": "SignerListSet",
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee": "12",
				"SignerQuorum": 0
			}`,
		},
		{
			name: "pass - without SignerEntries and SignerQuorum",
			entry: &SignerListSet{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:     types.XRPCurrencyAmount(12),
				},
			},
			expected: `{
				"TransactionType": "SignerListSet",
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee": "12"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testutil.CompareFlattenAndExpected(tt.entry.Flatten(), []byte(tt.expected))
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestSignerListSet_Validate(t *testing.T) {
	tests := []struct {
		name        string
		entry       *SignerListSet
		expectedErr error
	}{
		{
			name: "pass - valid SignerListSet",
			entry: newSignerListSetTx(
				3,
				newSignerListSetEntryWithWallet("rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW", 2, "Ledger"),
				newSignerListSetEntryWithWallet("rUpy3eEg8rqjqfUoLeBnZkscbKbFsKXC3v", 1, "Ledger Nano"),
				newSignerListSetEntryWithWallet("XVYRdEocC28DRx94ZFGP3qNJ1D5Ln7ecXFMd3vREB5Pesju", 1, "Ledger Nano"),
			),
		},
		{
			name: "pass - valid SignerListSet with large SignerWeight sum",
			entry: newSignerListSetTx(
				math.MaxUint16+1,
				newSignerListSetEntry("rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW", math.MaxUint16),
				newSignerListSetEntry("rUpy3eEg8rqjqfUoLeBnZkscbKbFsKXC3v", 1),
			),
		},
		{
			name: "pass - valid SignerListSet with unsorted SignerEntries",
			entry: newSignerListSetTx(
				2,
				newSignerListSetEntry("rUpy3eEg8rqjqfUoLeBnZkscbKbFsKXC3v", 1),
				newSignerListSetEntry("rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW", 1),
			),
		},
		{
			name: "fail - invalid SignerListSet BaseTx",
			entry: &SignerListSet{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:     types.XRPCurrencyAmount(12),
				},
				SignerQuorum: uint32(3),
				SignerEntries: []ledger.SignerEntryWrapper{
					newSignerListSetEntry("rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW", 2),
					newSignerListSetEntry("rUpy3eEg8rqjqfUoLeBnZkscbKbFsKXC3v", 1),
				},
			},
			expectedErr: ErrInvalidTransactionType,
		},
		{
			name: "fail - invalid SignerListSet with duplicate signer account",
			entry: newSignerListSetTx(
				2,
				newSignerListSetEntry("rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW", 1),
				newSignerListSetEntry("rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW", 1),
			),
			expectedErr: ErrDuplicateSignerAccount,
		},
		{
			name: "fail - invalid SignerListSet with duplicate signer account using X-address",
			entry: newSignerListSetTx(
				2,
				newSignerListSetEntry("rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW", 1),
				newSignerListSetEntry("X7d3eHCXzwBeWrZec1yT24iZerQjYL8m8zCJ16ACxu1BrBY", 1),
			),
			expectedErr: ErrDuplicateSignerAccount,
		},
		{
			name:        "fail - invalid SignerListSet with signer account matching transaction account",
			entry:       newSignerListSetTx(1, newSignerListSetEntry("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD", 1)),
			expectedErr: ErrSignerAccountMatchesAccount,
		},
		{
			name: "fail - invalid SignerListSet with X-address signer account matching transaction account",
			entry: &SignerListSet{
				BaseTx: BaseTx{
					Account:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					TransactionType: SignerListSetTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				SignerQuorum:  uint32(1),
				SignerEntries: []ledger.SignerEntryWrapper{newSignerListSetEntry("X7AcgcsBL6XDcUb289X4mJ8djcdyKaB5hJDWMArnXr61cqZ", 1)},
			},
			expectedErr: ErrSignerAccountMatchesAccount,
		},
		{
			name: "fail - invalid SignerListSet with X-address transaction account matching signer account",
			entry: &SignerListSet{
				BaseTx: BaseTx{
					Account:         "X7AcgcsBL6XDcUb289X4mJ8djcdyKaB5hJDWMArnXr61cqZ",
					TransactionType: SignerListSetTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				SignerQuorum:  uint32(1),
				SignerEntries: []ledger.SignerEntryWrapper{newSignerListSetEntry("r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59", 1)},
			},
			expectedErr: ErrSignerAccountMatchesAccount,
		},
		{
			name:        "fail - invalid SignerListSet with zero signer weight",
			entry:       newSignerListSetTx(1, newSignerListSetEntry("rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW", 0)),
			expectedErr: ErrInvalidSignerWeight,
		},
		{
			name:        "fail - invalid SignerListSet with no SignerEntries and quorum > 0",
			entry:       newSignerListSetTx(3),
			expectedErr: ErrInvalidSignerEntries,
		},
		{
			name:        "fail - invalid SignerListSet with too many SignerEntries",
			entry:       newSignerListSetTx(3, newSignerListSetEntries(MaxSigners+1)...),
			expectedErr: ErrInvalidSignerEntries,
		},
		{
			name: "fail - invalid SignerListSet with invalid WalletLocator",
			entry: newSignerListSetTx(3, ledger.SignerEntryWrapper{
				SignerEntry: ledger.SignerEntry{
					Account:       "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
					SignerWeight:  2,
					WalletLocator: "invalid_hex",
				},
			}),
			expectedErr: ErrInvalidWalletLocator,
		},
		{
			name: "fail - invalid SignerListSet with SignerQuorum greater than sum of SignerWeights",
			entry: newSignerListSetTx(
				5,
				newSignerListSetEntry("rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW", 2),
				newSignerListSetEntry("rUpy3eEg8rqjqfUoLeBnZkscbKbFsKXC3v", 1),
			),
			expectedErr: ErrSignerQuorumGreaterThanSumOfSignerWeights,
		},
		{
			name: "fail - invalid SignerEntry Account, not an xrpl address",
			entry: newSignerListSetTx(
				2,
				newSignerListSetEntry("invalid", 2),
				newSignerListSetEntry("rUpy3eEg8rqjqfUoLeBnZkscbKbFsKXC3v", 1),
			),
			expectedErr: ErrInvalidAccount,
		},
		{
			name:  "pass - valid SignerListSet with SignerQuorum 0",
			entry: newSignerListSetTx(0),
		},
		{
			name: "fail - invalid SignerListSet with SignerQuorum 0 but a SignerEntries not empty",
			entry: newSignerListSetTx(
				0,
				newSignerListSetEntry("invalid", 2),
				newSignerListSetEntry("rUpy3eEg8rqjqfUoLeBnZkscbKbFsKXC3v", 1),
			),
			expectedErr: ErrInvalidQuorumAndEntries,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.entry.Validate()

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				require.False(t, valid)
				return
			}

			require.NoError(t, err)
			require.True(t, valid)
		})
	}
}

func newSignerListSetTx(quorum uint32, entries ...ledger.SignerEntryWrapper) *SignerListSet {
	return &SignerListSet{
		BaseTx: BaseTx{
			Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
			TransactionType: SignerListSetTx,
			Fee:             types.XRPCurrencyAmount(12),
		},
		SignerQuorum:  quorum,
		SignerEntries: entries,
	}
}

func newSignerListSetEntry(account string, weight uint16) ledger.SignerEntryWrapper {
	return ledger.SignerEntryWrapper{
		SignerEntry: ledger.SignerEntry{
			Account:      types.Address(account),
			SignerWeight: weight,
		},
	}
}

func newSignerListSetEntryWithWallet(account string, weight uint16, wallet string) ledger.SignerEntryWrapper {
	entry := newSignerListSetEntry(account, weight)
	entry.SignerEntry.WalletLocator = types.Hash256(hex.EncodeToString([]byte(wallet)))
	return entry
}

func newSignerListSetEntries(count int) []ledger.SignerEntryWrapper {
	entries := make([]ledger.SignerEntryWrapper, count)
	for i := range entries {
		entries[i] = newSignerListSetEntry("rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW", 1)
	}
	return entries
}
