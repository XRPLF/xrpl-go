package transaction

import (
	"encoding/hex"
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
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
		name      string
		entry     *SignerListSet
		wantValid bool
		wantErr   bool
	}{
		{
			name: "pass - valid SignerListSet",
			entry: &SignerListSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: SignerListSetTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				SignerQuorum: uint32(3),
				SignerEntries: []ledger.SignerEntryWrapper{
					{
						SignerEntry: ledger.SignerEntry{
							Account:       "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
							SignerWeight:  2,
							WalletLocator: types.Hash256(hex.EncodeToString([]byte("Ledger"))),
						},
					},
					{
						SignerEntry: ledger.SignerEntry{
							Account:       "rUpy3eEg8rqjqfUoLeBnZkscbKbFsKXC3v",
							SignerWeight:  1,
							WalletLocator: types.Hash256(hex.EncodeToString([]byte("Ledger Nano"))),
						},
					},
					{
						SignerEntry: ledger.SignerEntry{
							Account:       "XVYRdEocC28DRx94ZFGP3qNJ1D5Ln7ecXFMd3vREB5Pesju",
							SignerWeight:  1,
							WalletLocator: types.Hash256(hex.EncodeToString([]byte("Ledger Nano"))),
						},
					},
				},
			},
			wantValid: true,
			wantErr:   false,
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
				},
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "fail - invalid SignerListSet with no SignerEntries and quorum > 0",
			entry: &SignerListSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: SignerListSetTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				SignerQuorum: uint32(3),
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "fail - invalid SignerListSet with too many SignerEntries",
			entry: &SignerListSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: SignerListSetTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				SignerQuorum: uint32(3),
				SignerEntries: func() []ledger.SignerEntryWrapper {
					entries := make([]ledger.SignerEntryWrapper, 33)
					for i := range entries {
						entries[i] = ledger.SignerEntryWrapper{
							SignerEntry: ledger.SignerEntry{
								Account:      "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
								SignerWeight: 1,
							},
						}
					}
					return entries
				}(),
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "fail - invalid SignerListSet with invalid WalletLocator",
			entry: &SignerListSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: SignerListSetTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				SignerQuorum: uint32(3),
				SignerEntries: []ledger.SignerEntryWrapper{
					{
						SignerEntry: ledger.SignerEntry{
							Account:       "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
							SignerWeight:  2,
							WalletLocator: "invalid_hex",
						},
					},
				},
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "fail - invalid SignerListSet with SignerQuorum greater than sum of SignerWeights",
			entry: &SignerListSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: SignerListSetTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				SignerQuorum: uint32(5),
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
				},
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "fail - invalid SignerEntry Account, not an xrpl address",
			entry: &SignerListSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: SignerListSetTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				SignerQuorum: uint32(2),
				SignerEntries: []ledger.SignerEntryWrapper{
					{
						SignerEntry: ledger.SignerEntry{
							Account:      "invalid",
							SignerWeight: 2,
						},
					},
					{
						SignerEntry: ledger.SignerEntry{
							Account:      "rUpy3eEg8rqjqfUoLeBnZkscbKbFsKXC3v",
							SignerWeight: 1,
						},
					},
				},
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "pass - valid SignerListSet with SignerQuorum 0",
			entry: &SignerListSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: SignerListSetTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				SignerQuorum: uint32(0),
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - invalid SignerListSet with SignerQuorum 0 but a SignerEntries not empty",
			entry: &SignerListSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: SignerListSetTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				SignerQuorum: uint32(0),
				SignerEntries: []ledger.SignerEntryWrapper{
					{
						SignerEntry: ledger.SignerEntry{
							Account:      "invalid",
							SignerWeight: 2,
						},
					},
					{
						SignerEntry: ledger.SignerEntry{
							Account:      "rUpy3eEg8rqjqfUoLeBnZkscbKbFsKXC3v",
							SignerWeight: 1,
						},
					},
				},
			},
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.entry.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if valid != tt.wantValid {
				t.Errorf("Validate() = %v, want %v", valid, !tt.wantErr)
			}
		})
	}
}
