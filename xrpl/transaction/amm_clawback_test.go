package transaction

import (
	"encoding/json"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

const (
	ammClawbackIssuer = types.Address("rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh")
	ammClawbackHolder = "rWYkbWkCeg8dP6rXALnjgZSjjLyih5NXm"
	ammClawbackOther  = types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD")
)

func newAMMClawback() *AMMClawback {
	return &AMMClawback{
		BaseTx: BaseTx{Account: ammClawbackIssuer, TransactionType: AMMClawbackTx},
		Holder: ammClawbackHolder,
		Asset: ledger.Asset{
			Currency: "USD",
			Issuer:   ammClawbackIssuer,
		},
		Asset2: ledger.Asset{Currency: "XRP"},
	}
}

func TestAMMClawback_TxType(t *testing.T) {
	tx := &AMMClawback{}
	require.Equal(t, AMMClawbackTx, tx.TxType())
}

func TestAMMClawback_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tx      *AMMClawback
		wantErr error
	}{
		{
			name: "valid omission claws all",
			tx:   newAMMClawback(),
		},
		{
			name: "asset2 is required",
			tx: func() *AMMClawback {
				tx := newAMMClawback()
				tx.Asset2 = ledger.Asset{}
				return tx
			}(),
			wantErr: ErrAMMClawbackInvalidAsset2,
		},
		{
			name: "asset cannot be XRP",
			tx: func() *AMMClawback {
				tx := newAMMClawback()
				tx.Asset = ledger.Asset{Currency: "XRP"}
				return tx
			}(),
			wantErr: ErrAMMClawbackAssetCannotBeXRP,
		},
		{
			name: "asset must have the account issuer",
			tx: func() *AMMClawback {
				tx := newAMMClawback()
				tx.Asset.Issuer = ammClawbackOther
				return tx
			}(),
			wantErr: ErrInvalidAssetIssuer,
		},
		{
			name: "amount must be positive",
			tx: func() *AMMClawback {
				tx := newAMMClawback()
				tx.Amount = types.IssuedCurrencyAmount{Currency: "USD", Issuer: ammClawbackIssuer, Value: "0"}
				return tx
			}(),
			wantErr: ErrAMMClawbackInvalidAmount,
		},
		{
			name: "amount must match asset",
			tx: func() *AMMClawback {
				tx := newAMMClawback()
				tx.Amount = types.IssuedCurrencyAmount{Currency: "EUR", Issuer: ammClawbackIssuer, Value: "1"}
				return tx
			}(),
			wantErr: ErrAMMClawbackAmountAssetMismatch,
		},
		{
			name: "matching invalid currencies are rejected",
			tx: func() *AMMClawback {
				tx := newAMMClawback()
				tx.Asset.Currency = "AD/"
				tx.Amount = types.IssuedCurrencyAmount{Currency: "AD/", Issuer: ammClawbackIssuer, Value: "1"}
				return tx
			}(),
			wantErr: ErrAMMClawbackAmountAssetMismatch,
		},
		{
			name: "nonstandard currency bytes must match asset",
			tx: func() *AMMClawback {
				tx := newAMMClawback()
				tx.Asset.Currency = "0000000000000000000000000000000000000001"
				tx.Amount = types.IssuedCurrencyAmount{
					Currency: "0000000000000000000000000000000000000002",
					Issuer:   ammClawbackIssuer,
					Value:    "1",
				}
				return tx
			}(),
			wantErr: ErrAMMClawbackAmountAssetMismatch,
		},
		{
			name: "canonical currency representations match",
			tx: func() *AMMClawback {
				tx := newAMMClawback()
				tx.Amount = types.IssuedCurrencyAmount{
					Currency: "0000000000000000000000005553440000000000",
					Issuer:   ammClawbackIssuer,
					Value:    "1",
				}
				return tx
			}(),
		},
		{
			name: "prefixed canonical currency representations match",
			tx: func() *AMMClawback {
				tx := newAMMClawback()
				tx.Amount = types.IssuedCurrencyAmount{
					Currency: "0x0000000000000000000000005553440000000000",
					Issuer:   ammClawbackIssuer,
					Value:    "1",
				}
				return tx
			}(),
		},
		{
			name: "two asset flag requires account issuer",
			tx: func() *AMMClawback {
				tx := newAMMClawback()
				tx.Flags = TfClawTwoAssets
				return tx
			}(),
			wantErr: ErrAMMClawbackAsset2IssuerMismatch,
		},
		{
			name: "two issued assets",
			tx: func() *AMMClawback {
				tx := newAMMClawback()
				tx.Flags = TfClawTwoAssets
				tx.Asset2 = ledger.Asset{Currency: "EUR", Issuer: ammClawbackIssuer}
				return tx
			}(),
		},
		{
			name: "MPT asset and amount",
			tx: &AMMClawback{
				BaseTx: BaseTx{Account: clawbackMPTIssuer, TransactionType: AMMClawbackTx},
				Holder: ammClawbackHolder,
				Asset:  ledger.Asset{MPTIssuanceID: clawbackMPTIssueID},
				Asset2: ledger.Asset{Currency: "XRP"},
				Amount: types.MPTCurrencyAmount{MPTIssuanceID: clawbackMPTIssueID, Value: "10"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := tt.tx.Validate()
			if tt.wantErr != nil {
				require.False(t, ok)
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.True(t, ok)
			require.NoError(t, err)
			_, err = binarycodec.Encode(tt.tx.Flatten())
			require.NoError(t, err)
		})
	}
}

func TestAMMClawback_CanonicalCurrencyEncodingParity(t *testing.T) {
	transactionWithAmountCurrency := func(currency string) *AMMClawback {
		tx := newAMMClawback()
		tx.Amount = types.IssuedCurrencyAmount{Currency: currency, Issuer: ammClawbackIssuer, Value: "1"}
		return tx
	}

	standardEncoded, err := binarycodec.Encode(transactionWithAmountCurrency("USD").Flatten())
	require.NoError(t, err)

	tests := []struct {
		name     string
		currency string
	}{
		{
			name:     "hex",
			currency: "0000000000000000000000005553440000000000",
		},
		{
			name:     "prefixed hex",
			currency: "0x0000000000000000000000005553440000000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := binarycodec.Encode(transactionWithAmountCurrency(tt.currency).Flatten())
			require.NoError(t, err)
			require.Equal(t, standardEncoded, encoded)
		})
	}
}

func TestAMMClawback_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		tx       *AMMClawback
		expected FlatTransaction
	}{
		{
			name: "issued asset with omitted amount",
			tx:   newAMMClawback(),
			expected: FlatTransaction{
				"Account":         string(ammClawbackIssuer),
				"TransactionType": "AMMClawback",
				"Holder":          ammClawbackHolder,
				"Asset": map[string]any{
					"currency": "USD",
					"issuer":   string(ammClawbackIssuer),
				},
				"Asset2": map[string]any{"currency": "XRP"},
			},
		},
		{
			name: "MPT asset and amount",
			tx: &AMMClawback{
				BaseTx: BaseTx{Account: clawbackMPTIssuer, TransactionType: AMMClawbackTx},
				Holder: ammClawbackHolder,
				Asset:  ledger.Asset{MPTIssuanceID: clawbackMPTIssueID},
				Asset2: ledger.Asset{Currency: "XRP"},
				Amount: types.MPTCurrencyAmount{MPTIssuanceID: clawbackMPTIssueID, Value: "10"},
			},
			expected: FlatTransaction{
				"Account":         string(clawbackMPTIssuer),
				"TransactionType": "AMMClawback",
				"Holder":          ammClawbackHolder,
				"Asset":           map[string]any{"mpt_issuance_id": clawbackMPTIssueID},
				"Asset2":          map[string]any{"currency": "XRP"},
				"Amount":          map[string]any{"mpt_issuance_id": clawbackMPTIssueID, "value": "10"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flattened := tt.tx.Flatten()
			require.Equal(t, tt.expected, flattened)
			_, err := binarycodec.Encode(flattened)
			require.NoError(t, err)
		})
	}
}

func TestAMMClawback_NestedIssuerSerialization(t *testing.T) {
	taglessIssuer, err := addresscodec.ClassicAddressToXAddress(ammClawbackIssuer.String(), 0, false, false)
	require.NoError(t, err)
	taggedIssuer, err := addresscodec.ClassicAddressToXAddress(ammClawbackIssuer.String(), 42, true, false)
	require.NoError(t, err)

	transactionWithIssuer := func(issuer types.Address) *AMMClawback {
		tx := newAMMClawback()
		tx.Asset.Issuer = issuer
		tx.Amount = types.IssuedCurrencyAmount{Currency: "USD", Issuer: issuer, Value: "1"}
		return tx
	}

	classicEncoded, err := binarycodec.Encode(transactionWithIssuer(ammClawbackIssuer).Flatten())
	require.NoError(t, err)

	tests := []struct {
		name        string
		issuer      types.Address
		wantIssuer  string
		wantEncoded bool
	}{
		{
			name:        "tagless X-address is serialized as its classic address",
			issuer:      types.Address(taglessIssuer),
			wantIssuer:  ammClawbackIssuer.String(),
			wantEncoded: true,
		},
		{
			name:       "tagged X-address is preserved and rejected by encoding",
			issuer:     types.Address(taggedIssuer),
			wantIssuer: taggedIssuer,
		},
		{
			name:       "malformed issuer is preserved and rejected by encoding",
			issuer:     "not-an-address",
			wantIssuer: "not-an-address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := transactionWithIssuer(tt.issuer)
			before := *tx

			flattened := tx.Flatten()
			asset, ok := flattened["Asset"].(map[string]any)
			require.True(t, ok)
			require.Equal(t, tt.wantIssuer, asset["issuer"])
			amount, ok := flattened["Amount"].(map[string]any)
			require.True(t, ok)
			require.Equal(t, tt.wantIssuer, amount["issuer"])
			require.Equal(t, before, *tx)

			encoded, err := binarycodec.Encode(flattened)
			if tt.wantEncoded {
				require.NoError(t, err)
				require.Equal(t, classicEncoded, encoded)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAMMClawback_UnmarshalJSONMPTAmount(t *testing.T) {
	data := []byte(`{
		"TransactionType":"AMMClawback",
		"Account":"rKGpqjZhYan5FLqGyAfAzHpJeUN8fs3SYi",
		"Holder":"rWYkbWkCeg8dP6rXALnjgZSjjLyih5NXm",
		"Asset":{"mpt_issuance_id":"00002403C84A0A28E0190E208E982C352BBD5006600555CF"},
		"Asset2":{"currency":"XRP"},
		"Amount":{"mpt_issuance_id":"00002403C84A0A28E0190E208E982C352BBD5006600555CF","value":"10"}
	}`)

	var tx AMMClawback
	require.NoError(t, json.Unmarshal(data, &tx))
	amount, ok := tx.Amount.(types.MPTCurrencyAmount)
	require.True(t, ok)
	require.Equal(t, clawbackMPTIssueID, amount.MPTIssuanceID)
}
