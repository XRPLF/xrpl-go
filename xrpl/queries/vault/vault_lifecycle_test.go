package vault

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestVaultInfoLendingFieldsUnmarshal(t *testing.T) {
	legacyVersion, cashVersion, unknownVersion := uint8(0), uint8(1), uint8(255)
	open, closed, unknownKind := types.VaultKindOpen, types.VaultKindClosed, types.VaultKind(255)
	zero, subscription, maxDate := uint32(0), uint32(800000000), uint32(math.MaxUint32)
	tests := []struct {
		name  string
		input string
		want  Vault
	}{
		{
			name:  "absent",
			input: `{"vault":{}}`,
			want:  Vault{},
		},
		{
			name:  "explicit zero",
			input: `{"vault":{"LEVersion":0,"VaultKind":0,"SubscriptionDate":0,"RedemptionDate":0}}`,
			want:  Vault{LEVersion: &legacyVersion, VaultKind: &open, SubscriptionDate: &zero, RedemptionDate: &zero},
		},
		{
			name:  "closed with maximum redemption date",
			input: `{"vault":{"LEVersion":1,"VaultKind":1,"SubscriptionDate":800000000,"RedemptionDate":4294967295}}`,
			want:  Vault{LEVersion: &cashVersion, VaultKind: &closed, SubscriptionDate: &subscription, RedemptionDate: &maxDate},
		},
		{
			name:  "unknown version and kind preserved",
			input: `{"vault":{"LEVersion":255,"VaultKind":255}}`,
			want:  Vault{LEVersion: &unknownVersion, VaultKind: &unknownKind},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var response Response
			require.NoError(t, json.Unmarshal([]byte(tc.input), &response))
			require.Equal(t, tc.want, response.Vault)
		})
	}
}

func TestVaultInfoLendingFieldsMarshalPresence(t *testing.T) {
	response := Response{}
	data, err := json.Marshal(response)
	require.NoError(t, err)
	require.NotContains(t, string(data), `"LEVersion":`)
	require.NotContains(t, string(data), `"VaultKind":`)
	require.NotContains(t, string(data), `"SubscriptionDate":`)
	require.NotContains(t, string(data), `"RedemptionDate":`)

	version, kind, date := uint8(0), types.VaultKindOpen, uint32(0)
	response.Vault = Vault{LEVersion: &version, VaultKind: &kind, SubscriptionDate: &date, RedemptionDate: &date}
	data, err = json.Marshal(response)
	require.NoError(t, err)
	require.Contains(t, string(data), `"LEVersion":0`)
	require.Contains(t, string(data), `"VaultKind":0`)
	require.Contains(t, string(data), `"SubscriptionDate":0`)
	require.Contains(t, string(data), `"RedemptionDate":0`)
}
