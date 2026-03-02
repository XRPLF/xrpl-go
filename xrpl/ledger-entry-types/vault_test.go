package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestVault_Serialization(t *testing.T) {
	assetsTotal := types.XRPLNumber("0")
	assetsAvailable := types.XRPLNumber("0")
	lossUnrealized := types.XRPLNumber("0")
	assetsMaximum := types.XRPLNumber("1000000")
	scale := uint8(6)

	tests := []struct {
		name     string
		vault    *Vault
		expected string
	}{
		{
			name: "pass - required fields only",
			vault: &Vault{
				LedgerEntryType:   VaultEntry,
				Index:             "20B136D7BF6D2E3D610E28E3E6BE09F5C8F4F0241BBF6E2D072AE1BACB1388F5",
				Flags:             0,
				Sequence:          200370,
				OwnerNode:         "0",
				Owner:             "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
				Account:           "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
				Asset:             Asset{Currency: "USD", Issuer: "rXJSJiZMxaLuH3kQBUV5DLipnYtrE6iVb"},
				ShareMPTID:        "0000000169F415C9F1AB6796AB9224CE635818AFD74F8175",
				WithdrawalPolicy:  1,
				PreviousTxnID:     "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
				PreviousTxnLgrSeq: 28991004,
			},
			expected: `{
	"index": "20B136D7BF6D2E3D610E28E3E6BE09F5C8F4F0241BBF6E2D072AE1BACB1388F5",
	"LedgerEntryType": "Vault",
	"Flags": 0,
	"Sequence": 200370,
	"OwnerNode": "0",
	"Owner": "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
	"Account": "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
	"Asset": {
		"currency": "USD",
		"issuer": "rXJSJiZMxaLuH3kQBUV5DLipnYtrE6iVb"
	},
	"ShareMPTID": "0000000169F415C9F1AB6796AB9224CE635818AFD74F8175",
	"WithdrawalPolicy": 1,
	"PreviousTxnID": "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
	"PreviousTxnLgrSeq": 28991004
}`,
		},
		{
			name: "pass - with all optional fields",
			vault: &Vault{
				LedgerEntryType:   VaultEntry,
				Index:             "5A92F6ED33FDA68FB4B9FD140EA38C056CD2BA9673ECA5B4CEF40F2166BB6F0C",
				Flags:             0,
				Sequence:          200370,
				OwnerNode:         "0",
				Owner:             "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
				Account:           "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
				Asset:             Asset{Currency: "USD", Issuer: "rXJSJiZMxaLuH3kQBUV5DLipnYtrE6iVb"},
				AssetsTotal:       &assetsTotal,
				AssetsAvailable:   &assetsAvailable,
				LossUnrealized:    &lossUnrealized,
				ShareMPTID:        "0000000169F415C9F1AB6796AB9224CE635818AFD74F8175",
				WithdrawalPolicy:  1,
				AssetsMaximum:     &assetsMaximum,
				Data:              "5661756C74206D65746164617461",
				Scale:             &scale,
				PreviousTxnID:     "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
				PreviousTxnLgrSeq: 28991004,
			},
			expected: `{
	"index": "5A92F6ED33FDA68FB4B9FD140EA38C056CD2BA9673ECA5B4CEF40F2166BB6F0C",
	"LedgerEntryType": "Vault",
	"Flags": 0,
	"Sequence": 200370,
	"OwnerNode": "0",
	"Owner": "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
	"Account": "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
	"Asset": {
		"currency": "USD",
		"issuer": "rXJSJiZMxaLuH3kQBUV5DLipnYtrE6iVb"
	},
	"AssetsTotal": "0",
	"AssetsAvailable": "0",
	"LossUnrealized": "0",
	"ShareMPTID": "0000000169F415C9F1AB6796AB9224CE635818AFD74F8175",
	"WithdrawalPolicy": 1,
	"AssetsMaximum": "1000000",
	"Data": "5661756C74206D65746164617461",
	"Scale": 6,
	"PreviousTxnID": "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
	"PreviousTxnLgrSeq": 28991004
}`,
		},
		{
			name: "pass - with private flag",
			vault: &Vault{
				LedgerEntryType:   VaultEntry,
				Index:             "20B136D7BF6D2E3D610E28E3E6BE09F5C8F4F0241BBF6E2D072AE1BACB1388F5",
				Flags:             LsfVaultPrivate,
				Sequence:          200370,
				OwnerNode:         "0",
				Owner:             "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
				Account:           "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
				Asset:             Asset{Currency: "USD", Issuer: "rXJSJiZMxaLuH3kQBUV5DLipnYtrE6iVb"},
				ShareMPTID:        "0000000169F415C9F1AB6796AB9224CE635818AFD74F8175",
				WithdrawalPolicy:  1,
				PreviousTxnID:     "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
				PreviousTxnLgrSeq: 28991004,
			},
			expected: `{
	"index": "20B136D7BF6D2E3D610E28E3E6BE09F5C8F4F0241BBF6E2D072AE1BACB1388F5",
	"LedgerEntryType": "Vault",
	"Flags": 65536,
	"Sequence": 200370,
	"OwnerNode": "0",
	"Owner": "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
	"Account": "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
	"Asset": {
		"currency": "USD",
		"issuer": "rXJSJiZMxaLuH3kQBUV5DLipnYtrE6iVb"
	},
	"ShareMPTID": "0000000169F415C9F1AB6796AB9224CE635818AFD74F8175",
	"WithdrawalPolicy": 1,
	"PreviousTxnID": "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
	"PreviousTxnLgrSeq": 28991004
}`,
		},
		{
			name: "pass - with XRP asset",
			vault: &Vault{
				LedgerEntryType:   VaultEntry,
				Index:             "20B136D7BF6D2E3D610E28E3E6BE09F5C8F4F0241BBF6E2D072AE1BACB1388F5",
				Flags:             0,
				Sequence:          200370,
				OwnerNode:         "0",
				Owner:             "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
				Account:           "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
				Asset:             Asset{Currency: "XRP"},
				ShareMPTID:        "0000000169F415C9F1AB6796AB9224CE635818AFD74F8175",
				WithdrawalPolicy:  1,
				PreviousTxnID:     "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
				PreviousTxnLgrSeq: 28991004,
			},
			expected: `{
	"index": "20B136D7BF6D2E3D610E28E3E6BE09F5C8F4F0241BBF6E2D072AE1BACB1388F5",
	"LedgerEntryType": "Vault",
	"Flags": 0,
	"Sequence": 200370,
	"OwnerNode": "0",
	"Owner": "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
	"Account": "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
	"Asset": {
		"currency": "XRP"
	},
	"ShareMPTID": "0000000169F415C9F1AB6796AB9224CE635818AFD74F8175",
	"WithdrawalPolicy": 1,
	"PreviousTxnID": "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
	"PreviousTxnLgrSeq": 28991004
}`,
		},
		{
			name: "pass - with MPT asset",
			vault: &Vault{
				LedgerEntryType:   VaultEntry,
				Index:             "20B136D7BF6D2E3D610E28E3E6BE09F5C8F4F0241BBF6E2D072AE1BACB1388F5",
				Flags:             0,
				Sequence:          200370,
				OwnerNode:         "0",
				Owner:             "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
				Account:           "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
				Asset:             Asset{MPTIssuanceID: "0000000169F415C9F1AB6796AB9224CE635818AFD74F8175"},
				ShareMPTID:        "0000000169F415C9F1AB6796AB9224CE635818AFD74F8175",
				WithdrawalPolicy:  1,
				PreviousTxnID:     "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
				PreviousTxnLgrSeq: 28991004,
			},
			expected: `{
	"index": "20B136D7BF6D2E3D610E28E3E6BE09F5C8F4F0241BBF6E2D072AE1BACB1388F5",
	"LedgerEntryType": "Vault",
	"Flags": 0,
	"Sequence": 200370,
	"OwnerNode": "0",
	"Owner": "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
	"Account": "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
	"Asset": {
		"mpt_issuance_id": "0000000169F415C9F1AB6796AB9224CE635818AFD74F8175"
	},
	"ShareMPTID": "0000000169F415C9F1AB6796AB9224CE635818AFD74F8175",
	"WithdrawalPolicy": 1,
	"PreviousTxnID": "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
	"PreviousTxnLgrSeq": 28991004
}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := testutil.SerializeAndDeserialize(t, test.vault, test.expected); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestVault_EntryType(t *testing.T) {
	s := &Vault{}
	require.Equal(t, s.EntryType(), VaultEntry)
}

func TestVault_SetLsfVaultPrivate(t *testing.T) {
	v := &Vault{}
	v.SetLsfVaultPrivate()
	require.Equal(t, v.Flags, LsfVaultPrivate)
}
