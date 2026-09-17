package ledger

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestVaultLendingFieldsUnmarshal(t *testing.T) {
	legacyVersion, cashVersion, unknownVersion := uint8(0), uint8(1), uint8(255)
	open, closed, unknownKind := types.VaultKindOpen, types.VaultKindClosed, types.VaultKind(255)
	zero, maxDate := uint32(0), uint32(math.MaxUint32)
	tests := []struct {
		name  string
		input string
		want  Vault
	}{
		{
			name:  "absent",
			input: `{}`,
			want:  Vault{},
		},
		{
			name:  "explicit zero",
			input: `{"LEVersion":0,"VaultKind":0,"SubscriptionDate":0,"RedemptionDate":0}`,
			want:  Vault{LEVersion: &legacyVersion, VaultKind: &open, SubscriptionDate: &zero, RedemptionDate: &zero},
		},
		{
			name:  "closed with maximum redemption date",
			input: `{"LEVersion":1,"VaultKind":1,"SubscriptionDate":0,"RedemptionDate":4294967295}`,
			want:  Vault{LEVersion: &cashVersion, VaultKind: &closed, SubscriptionDate: &zero, RedemptionDate: &maxDate},
		},
		{
			name:  "unknown version and kind preserved",
			input: `{"LEVersion":255,"VaultKind":255}`,
			want:  Vault{LEVersion: &unknownVersion, VaultKind: &unknownKind},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var vault Vault
			require.NoError(t, json.Unmarshal([]byte(tc.input), &vault))
			require.Equal(t, tc.want, vault)
		})
	}
}

func TestVaultLendingFieldsMarshalPresence(t *testing.T) {
	vault := Vault{}
	data, err := json.Marshal(vault)
	require.NoError(t, err)
	require.NotContains(t, string(data), `"LEVersion":`)
	require.NotContains(t, string(data), `"VaultKind":`)
	require.NotContains(t, string(data), `"SubscriptionDate":`)
	require.NotContains(t, string(data), `"RedemptionDate":`)

	version, kind, date := uint8(0), types.VaultKindOpen, uint32(0)
	vault.LEVersion = &version
	vault.VaultKind = &kind
	vault.SubscriptionDate = &date
	vault.RedemptionDate = &date
	data, err = json.Marshal(vault)
	require.NoError(t, err)
	require.Contains(t, string(data), `"LEVersion":0`)
	require.Contains(t, string(data), `"VaultKind":0`)
	require.Contains(t, string(data), `"SubscriptionDate":0`)
	require.Contains(t, string(data), `"RedemptionDate":0`)
}
