package transaction

import (
	"encoding/json"
	"strings"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	bctypes "github.com/Peersyst/xrpl-go/binary-codec/types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

const (
	clawbackIOUIssuer    = types.Address("rnLYcEcYw2r3w6BDsFDSScoFmvZXbwa6EQ")
	clawbackMPTIssuer    = types.Address("rKGpqjZhYan5FLqGyAfAzHpJeUN8fs3SYi")
	clawbackHolder       = types.Address("rhsTg7mm7v3oEGrF85n1KdB3JjCk5KPT4M")
	clawbackTaggedHolder = types.Address("X7dTFb8yBn6ZY5gCdyNNuvFkTNx7oBTvbpwXNLCBUUVXjLV")
	clawbackMPTIssueID   = "00002403C84A0A28E0190E208E982C352BBD5006600555CF"
)

func newClawbackBaseTx(account types.Address) BaseTx {
	return BaseTx{
		Account:         account,
		TransactionType: ClawbackTx,
		Fee:             types.XRPCurrencyAmount(1),
		Sequence:        1234,
		SigningPubKey:   "ghijk",
		TxnSignature:    "A1B2C3D4E5F6",
	}
}

// newUnsignedClawbackIOU returns an unsigned issued-currency Clawback fixture
// shared by the JSON and binary-codec round-trip tests.
func newUnsignedClawbackIOU() Clawback {
	return Clawback{
		BaseTx: BaseTx{
			Account:         clawbackIOUIssuer,
			TransactionType: ClawbackTx,
			Fee:             types.XRPCurrencyAmount(1),
			Sequence:        1234,
		},
		Amount: types.IssuedCurrencyAmount{
			Issuer:   clawbackHolder,
			Currency: "USD",
			Value:    "1",
		},
	}
}

// newUnsignedClawbackMPT returns an unsigned MPT Clawback fixture
// shared by the JSON and binary-codec round-trip tests.
func newUnsignedClawbackMPT() Clawback {
	return Clawback{
		BaseTx: BaseTx{
			Account:         clawbackMPTIssuer,
			TransactionType: ClawbackTx,
			Fee:             types.XRPCurrencyAmount(1),
			Sequence:        1234,
		},
		Amount: types.MPTCurrencyAmount{
			MPTIssuanceID: clawbackMPTIssueID,
			Value:         "10",
		},
		Holder: clawbackHolder,
	}
}

func TestClawback_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		clawback Clawback
		expected FlatTransaction
	}{
		{
			name: "pass - issued currency",
			clawback: Clawback{
				BaseTx: newClawbackBaseTx(clawbackIOUIssuer),
				Amount: types.IssuedCurrencyAmount{
					Issuer:   clawbackHolder,
					Currency: "USD",
					Value:    "1",
				},
			},
			expected: FlatTransaction{
				"Account":         clawbackIOUIssuer.String(),
				"TransactionType": "Clawback",
				"Fee":             "1",
				"Sequence":        uint32(1234),
				"SigningPubKey":   "ghijk",
				"TxnSignature":    "A1B2C3D4E5F6",
				"Amount": map[string]any{
					"issuer":   clawbackHolder.String(),
					"currency": "USD",
					"value":    "1",
				},
			},
		},
		{
			name: "pass - MPT",
			clawback: Clawback{
				BaseTx: newClawbackBaseTx(clawbackMPTIssuer),
				Amount: types.MPTCurrencyAmount{
					MPTIssuanceID: clawbackMPTIssueID,
					Value:         "10",
				},
				Holder: clawbackHolder,
			},
			expected: FlatTransaction{
				"Account":         clawbackMPTIssuer.String(),
				"TransactionType": "Clawback",
				"Fee":             "1",
				"Sequence":        uint32(1234),
				"SigningPubKey":   "ghijk",
				"TxnSignature":    "A1B2C3D4E5F6",
				"Amount": map[string]any{
					"mpt_issuance_id": clawbackMPTIssueID,
					"value":           "10",
				},
				"Holder": clawbackHolder.String(),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, test.clawback.Flatten())
		})
	}
}

func TestClawback_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		clawback Clawback
		expected string
	}{
		{
			name:     "pass - issued currency omits Holder",
			clawback: newUnsignedClawbackIOU(),
			expected: `{
				"Account":"rnLYcEcYw2r3w6BDsFDSScoFmvZXbwa6EQ",
				"TransactionType":"Clawback",
				"Fee":"1",
				"Sequence":1234,
				"Amount":{"issuer":"rhsTg7mm7v3oEGrF85n1KdB3JjCk5KPT4M","currency":"USD","value":"1"}
			}`,
		},
		{
			name:     "pass - MPT includes Holder",
			clawback: newUnsignedClawbackMPT(),
			expected: `{
				"Account":"rKGpqjZhYan5FLqGyAfAzHpJeUN8fs3SYi",
				"TransactionType":"Clawback",
				"Fee":"1",
				"Sequence":1234,
				"Amount":{"mpt_issuance_id":"00002403C84A0A28E0190E208E982C352BBD5006600555CF","value":"10"},
				"Holder":"rhsTg7mm7v3oEGrF85n1KdB3JjCk5KPT4M"
			}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := json.Marshal(test.clawback)
			require.NoError(t, err)
			require.JSONEq(t, test.expected, string(encoded))

			var decoded Clawback
			require.NoError(t, json.Unmarshal(encoded, &decoded))
			require.Equal(t, test.clawback, decoded)
		})
	}
}

func TestClawback_Validate(t *testing.T) {
	taggedMPTIssuer, err := addresscodec.ClassicAddressToXAddress(clawbackMPTIssuer.String(), 123, true, false)
	require.NoError(t, err)
	taglessIOUIssuer, err := addresscodec.ClassicAddressToXAddress(clawbackIOUIssuer.String(), 0, false, false)
	require.NoError(t, err)
	taglessMPTIssuer, err := addresscodec.ClassicAddressToXAddress(clawbackMPTIssuer.String(), 0, false, false)
	require.NoError(t, err)

	taggedAccountWithSourceTag := newClawbackBaseTx(types.Address(taggedMPTIssuer))
	taggedAccountWithSourceTag.SourceTag = 123
	taglessAccountWithSourceTag := newClawbackBaseTx(types.Address(taglessMPTIssuer))
	taglessAccountWithSourceTag.SourceTag = 123

	validIOU := types.IssuedCurrencyAmount{
		Issuer:   clawbackHolder,
		Currency: "USD",
		Value:    "1",
	}
	validMPT := types.MPTCurrencyAmount{
		MPTIssuanceID: clawbackMPTIssueID,
		Value:         "10",
	}

	tests := []struct {
		name     string
		clawback Clawback
		wantErr  error
	}{
		{
			name: "pass - valid issued currency",
			clawback: Clawback{
				BaseTx: newClawbackBaseTx(clawbackIOUIssuer),
				Amount: validIOU,
			},
		},
		{
			name: "pass - valid MPT",
			clawback: Clawback{
				BaseTx: newClawbackBaseTx(clawbackMPTIssuer),
				Amount: validMPT,
				Holder: clawbackHolder,
			},
		},
		{
			name: "pass - valid MPT with tagged X-address Account",
			clawback: Clawback{
				BaseTx: newClawbackBaseTx(types.Address(taggedMPTIssuer)),
				Amount: validMPT,
				Holder: clawbackHolder,
			},
		},
		{
			name: "pass - valid MPT with tagless X-address Account and SourceTag",
			clawback: Clawback{
				BaseTx: taglessAccountWithSourceTag,
				Amount: validMPT,
				Holder: clawbackHolder,
			},
		},
		{
			name: "fail - tagged X-address Account with explicit SourceTag",
			clawback: Clawback{
				BaseTx: taggedAccountWithSourceTag,
				Amount: validMPT,
				Holder: clawbackHolder,
			},
			wantErr: bctypes.ErrDuplicateXAddressTag,
		},
		{
			name: "fail - MPT issuance issuer differs from Account",
			clawback: Clawback{
				BaseTx: newClawbackBaseTx(clawbackIOUIssuer),
				Amount: validMPT,
				Holder: clawbackHolder,
			},
			wantErr: ErrClawbackMPTIssuerMismatch,
		},
		{
			name: "fail - missing Amount",
			clawback: Clawback{
				BaseTx: newClawbackBaseTx(clawbackIOUIssuer),
			},
			wantErr: ErrClawbackMissingAmount,
		},
		{
			name: "fail - XRP Amount",
			clawback: Clawback{
				BaseTx: newClawbackBaseTx(clawbackIOUIssuer),
				Amount: types.XRPCurrencyAmount(1),
			},
			wantErr: ErrClawbackInvalidAmount,
		},
		{
			name: "fail - invalid issued currency Amount",
			clawback: Clawback{
				BaseTx: newClawbackBaseTx(clawbackIOUIssuer),
				Amount: types.IssuedCurrencyAmount{
					Issuer:   clawbackHolder,
					Currency: "USD",
					Value:    "invalid",
				},
			},
			wantErr: ErrClawbackInvalidAmount,
		},
		{
			name: "fail - zero issued currency Amount",
			clawback: Clawback{
				BaseTx: newClawbackBaseTx(clawbackIOUIssuer),
				Amount: types.IssuedCurrencyAmount{
					Issuer:   clawbackHolder,
					Currency: "USD",
					Value:    "0",
				},
			},
			wantErr: ErrClawbackInvalidAmount,
		},
		{
			name: "fail - Holder with issued currency",
			clawback: Clawback{
				BaseTx: newClawbackBaseTx(clawbackIOUIssuer),
				Amount: validIOU,
				Holder: clawbackHolder,
			},
			wantErr: ErrClawbackHolderNotAllowed,
		},
		{
			name: "fail - issued currency self-targeting",
			clawback: Clawback{
				BaseTx: newClawbackBaseTx(clawbackIOUIssuer),
				Amount: types.IssuedCurrencyAmount{
					Issuer:   clawbackIOUIssuer,
					Currency: "USD",
					Value:    "1",
				},
			},
			wantErr: ErrClawbackSameAccount,
		},
		{
			name: "fail - issued currency self-targeting with equivalent tagless X-address",
			clawback: Clawback{
				BaseTx: newClawbackBaseTx(clawbackIOUIssuer),
				Amount: types.IssuedCurrencyAmount{
					Issuer:   types.Address(taglessIOUIssuer),
					Currency: "USD",
					Value:    "1",
				},
			},
			wantErr: ErrClawbackSameAccount,
		},
		{
			name: "fail - missing Holder with MPT",
			clawback: Clawback{
				BaseTx: newClawbackBaseTx(clawbackMPTIssuer),
				Amount: validMPT,
			},
			wantErr: ErrClawbackMissingHolder,
		},
		{
			name: "fail - invalid MPT Amount",
			clawback: Clawback{
				BaseTx: newClawbackBaseTx(clawbackMPTIssuer),
				Amount: types.MPTCurrencyAmount{
					MPTIssuanceID: "1234",
					Value:         "10",
				},
				Holder: clawbackHolder,
			},
			wantErr: ErrClawbackInvalidAmount,
		},
		{
			name: "fail - zero MPT Amount",
			clawback: Clawback{
				BaseTx: newClawbackBaseTx(clawbackMPTIssuer),
				Amount: types.MPTCurrencyAmount{
					MPTIssuanceID: clawbackMPTIssueID,
					Value:         "0",
				},
				Holder: clawbackHolder,
			},
			wantErr: ErrClawbackInvalidAmount,
		},
		{
			name: "fail - invalid MPT Holder",
			clawback: Clawback{
				BaseTx: newClawbackBaseTx(clawbackMPTIssuer),
				Amount: validMPT,
				Holder: "invalid",
			},
			wantErr: ErrClawbackInvalidHolder,
		},
		{
			name: "fail - tagged X-address MPT Holder",
			clawback: Clawback{
				BaseTx: newClawbackBaseTx(clawbackMPTIssuer),
				Amount: validMPT,
				Holder: clawbackTaggedHolder,
			},
			wantErr: ErrClawbackHolderTagNotAllowed,
		},
		{
			name: "fail - MPT self-targeting",
			clawback: Clawback{
				BaseTx: newClawbackBaseTx(clawbackMPTIssuer),
				Amount: validMPT,
				Holder: clawbackMPTIssuer,
			},
			wantErr: ErrClawbackSameHolder,
		},
		{
			name: "fail - MPT self-targeting with equivalent tagless X-address",
			clawback: Clawback{
				BaseTx: newClawbackBaseTx(clawbackMPTIssuer),
				Amount: validMPT,
				Holder: types.Address(taglessMPTIssuer),
			},
			wantErr: ErrClawbackSameHolder,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			valid, err := test.clawback.Validate()
			if test.wantErr == nil {
				require.NoError(t, err)
				require.True(t, valid)
				return
			}

			require.ErrorIs(t, err, test.wantErr)
			require.False(t, valid)
		})
	}
}

func TestClawback_BinaryCodecRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		clawback Clawback
	}{
		{
			name:     "pass - issued currency",
			clawback: newUnsignedClawbackIOU(),
		},
		{
			name:     "pass - MPT",
			clawback: newUnsignedClawbackMPT(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flattened := test.clawback.Flatten()
			encoded, err := binarycodec.Encode(flattened)
			require.NoError(t, err)
			require.NotEmpty(t, encoded)

			decoded, err := binarycodec.Decode(encoded)
			require.NoError(t, err)
			require.Equal(t, flattened["TransactionType"], decoded["TransactionType"])
			require.Equal(t, flattened["Account"], decoded["Account"])
			if test.clawback.Amount.Kind() == types.MPT {
				expectedAmount := flattened["Amount"].(map[string]any)
				decodedAmount := decoded["Amount"].(map[string]any)
				require.Equal(t, expectedAmount["value"], decodedAmount["value"])
				require.True(t, strings.EqualFold(expectedAmount["mpt_issuance_id"].(string), decodedAmount["mpt_issuance_id"].(string)))
			} else {
				require.Equal(t, flattened["Amount"], decoded["Amount"])
			}
			require.Equal(t, flattened["Holder"], decoded["Holder"])
		})
	}

	t.Run("pass - tagged X-address Account becomes classic Account and SourceTag", func(t *testing.T) {
		const sourceTag uint32 = 123
		taggedAccount, err := addresscodec.ClassicAddressToXAddress(clawbackMPTIssuer.String(), sourceTag, true, false)
		require.NoError(t, err)

		clawback := newUnsignedClawbackMPT()
		clawback.Account = types.Address(taggedAccount)
		valid, err := clawback.Validate()
		require.NoError(t, err)
		require.True(t, valid)

		encoded, err := binarycodec.Encode(clawback.Flatten())
		require.NoError(t, err)
		decoded, err := binarycodec.Decode(encoded)
		require.NoError(t, err)
		require.Equal(t, clawbackMPTIssuer.String(), decoded["Account"])
		require.Equal(t, sourceTag, decoded["SourceTag"])
	})
}
