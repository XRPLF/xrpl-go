package transaction

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

type withdrawalTestTx interface {
	Tx
	Validate() (bool, error)
	Flatten() FlatTransaction
}

func TestWithdrawalCredentialValidation(t *testing.T) {
	id := strings.Repeat("AB", 32)
	maximum := make(types.CredentialIDs, 8)
	for i := range maximum {
		maximum[i] = fmt.Sprintf("%064X", i+1)
	}
	tests := []struct {
		name  string
		ids   types.CredentialIDs
		valid bool
	}{
		{"absent", nil, true},
		{"empty", types.CredentialIDs{}, false},
		{"one", types.CredentialIDs{id}, true},
		{"lowercase", types.CredentialIDs{strings.ToLower(id)}, true},
		{"eight", maximum, true},
		{"nine", append(append(types.CredentialIDs{}, maximum...), id), false},
		{"duplicate", types.CredentialIDs{id, id}, false},
		{"duplicate case", types.CredentialIDs{id, strings.ToLower(id)}, false},
		{"empty id", types.CredentialIDs{""}, false},
		{"zero id", types.CredentialIDs{strings.Repeat("0", 64)}, false},
		{"short", types.CredentialIDs{"AB"}, false},
		{"odd", types.CredentialIDs{id[:63]}, false},
		{"long", types.CredentialIDs{id + "00"}, false},
		{"nonhex", types.CredentialIDs{strings.Repeat("ZZ", 32)}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			txs := []withdrawalTestTx{
				&VaultWithdraw{
					BaseTx:        BaseTx{TransactionType: VaultWithdrawTx, Account: "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es"},
					VaultID:       types.Hash256(id),
					Amount:        types.XRPCurrencyAmount(1),
					CredentialIDs: tc.ids,
				},
				&LoanBrokerCoverWithdraw{
					BaseTx:        BaseTx{TransactionType: LoanBrokerCoverWithdrawTx, Account: "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es"},
					LoanBrokerID:  id,
					Amount:        types.XRPCurrencyAmount(1),
					CredentialIDs: tc.ids,
				},
			}
			for _, tx := range txs {
				t.Run(tx.TxType().String(), func(t *testing.T) {
					ok, err := tx.Validate()
					require.Equal(t, tc.valid, ok)
					if tc.valid {
						require.NoError(t, err)
					} else {
						require.ErrorIs(t, err, ErrInvalidCredentialIDs)
					}
				})
			}
		})
	}
}

func TestWithdrawalCredentialsJSON(t *testing.T) {
	const id = "ABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABAB"
	tests := []struct {
		name       string
		ids        types.CredentialIDs
		vaultJSON  string
		brokerJSON string
	}{
		{
			name:       "absent",
			vaultJSON:  `{"Account":"","TransactionType":"","VaultID":"","Amount":null}`,
			brokerJSON: `{"Account":"","TransactionType":"","LoanBrokerID":"","Amount":null}`,
		},
		{
			name:       "explicit empty preserved",
			ids:        types.CredentialIDs{},
			vaultJSON:  `{"Account":"","TransactionType":"","VaultID":"","Amount":null,"CredentialIDs":[]}`,
			brokerJSON: `{"Account":"","TransactionType":"","LoanBrokerID":"","Amount":null,"CredentialIDs":[]}`,
		},
		{
			name:       "one credential",
			ids:        types.CredentialIDs{id},
			vaultJSON:  `{"Account":"","TransactionType":"","VaultID":"","Amount":null,"CredentialIDs":["` + id + `"]}`,
			brokerJSON: `{"Account":"","TransactionType":"","LoanBrokerID":"","Amount":null,"CredentialIDs":["` + id + `"]}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("VaultWithdraw", func(t *testing.T) {
				want := VaultWithdraw{CredentialIDs: tc.ids}
				var decoded VaultWithdraw
				require.NoError(t, json.Unmarshal([]byte(tc.vaultJSON), &decoded))
				require.Equal(t, want, decoded)
				data, err := json.Marshal(want)
				require.NoError(t, err)
				require.JSONEq(t, tc.vaultJSON, string(data))
			})
			t.Run("LoanBrokerCoverWithdraw", func(t *testing.T) {
				want := LoanBrokerCoverWithdraw{CredentialIDs: tc.ids}
				var decoded LoanBrokerCoverWithdraw
				require.NoError(t, json.Unmarshal([]byte(tc.brokerJSON), &decoded))
				require.Equal(t, want, decoded)
				data, err := json.Marshal(want)
				require.NoError(t, err)
				require.JSONEq(t, tc.brokerJSON, string(data))
			})
		})
	}
}

func TestWithdrawalCredentialsFlattenPresence(t *testing.T) {
	vault := VaultWithdraw{}
	broker := LoanBrokerCoverWithdraw{}
	require.NotContains(t, vault.Flatten(), "CredentialIDs")
	require.NotContains(t, broker.Flatten(), "CredentialIDs")

	vault.CredentialIDs = types.CredentialIDs{}
	broker.CredentialIDs = types.CredentialIDs{}
	require.Equal(t, []string{}, vault.Flatten()["CredentialIDs"])
	require.Equal(t, []string{}, broker.Flatten()["CredentialIDs"])
}

func TestWithdrawalCredentialsBinaryRoundTrip(t *testing.T) {
	id := strings.Repeat("AB", 32)
	ids := types.CredentialIDs{id, strings.Repeat("CD", 32)}
	txs := []withdrawalTestTx{
		&VaultWithdraw{
			BaseTx:        BaseTx{Account: "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es"},
			VaultID:       types.Hash256(id),
			Amount:        types.XRPCurrencyAmount(1),
			CredentialIDs: ids,
		},
		&LoanBrokerCoverWithdraw{
			BaseTx:        BaseTx{Account: "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es"},
			LoanBrokerID:  id,
			Amount:        types.XRPCurrencyAmount(1),
			CredentialIDs: ids,
		},
	}
	for _, tx := range txs {
		t.Run(tx.TxType().String(), func(t *testing.T) {
			flat := tx.Flatten()
			blob, err := binarycodec.Encode(flat)
			require.NoError(t, err)
			decoded, err := binarycodec.Decode(blob)
			require.NoError(t, err)
			require.Equal(t, []string(ids), decoded["CredentialIDs"])
			require.Equal(t, map[string]any(flat), decoded)
			signing, err := binarycodec.EncodeForSigning(flat)
			require.NoError(t, err)
			require.Equal(t, "53545800"+blob, signing)
		})
	}
}
