package transaction_test

import (
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/hash"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/stretchr/testify/require"
)

func TestClawbackSigning(t *testing.T) {
	issuer, err := wallet.New(crypto.ED25519())
	require.NoError(t, err)
	holder, err := wallet.New(crypto.ED25519())
	require.NoError(t, err)

	mptIssuanceID, err := hash.MPTID(1, issuer.ClassicAddress.String())
	require.NoError(t, err)
	taglessHolder, err := addresscodec.ClassicAddressToXAddress(holder.ClassicAddress.String(), 0, false, false)
	require.NoError(t, err)
	testnetTaglessHolder, err := addresscodec.ClassicAddressToXAddress(holder.ClassicAddress.String(), 0, false, true)
	require.NoError(t, err)
	mptAmount := types.MPTCurrencyAmount{
		MPTIssuanceID: mptIssuanceID,
		Value:         "10",
	}

	tests := []struct {
		name           string
		amount         types.CurrencyAmount
		holder         types.Address
		expectedHolder types.Address
	}{
		{
			name: "pass - issued currency",
			amount: types.IssuedCurrencyAmount{
				Issuer:   holder.ClassicAddress,
				Currency: "USD",
				Value:    "1",
			},
		},
		{
			name: "pass - issued currency with tagless X-address holder",
			amount: types.IssuedCurrencyAmount{
				Issuer:   types.Address(taglessHolder),
				Currency: "USD",
				Value:    "1",
			},
		},
		{
			name: "pass - issued currency with tagless testnet X-address holder",
			amount: types.IssuedCurrencyAmount{
				Issuer:   types.Address(testnetTaglessHolder),
				Currency: "USD",
				Value:    "1",
			},
		},
		{
			name:           "pass - MPT with classic Holder",
			amount:         mptAmount,
			holder:         holder.ClassicAddress,
			expectedHolder: holder.ClassicAddress,
		},
		{
			name:           "pass - MPT with tagless X-address Holder",
			amount:         mptAmount,
			holder:         types.Address(taglessHolder),
			expectedHolder: holder.ClassicAddress,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clawback := transaction.Clawback{
				BaseTx: transaction.BaseTx{
					Account:         issuer.ClassicAddress,
					TransactionType: transaction.ClawbackTx,
					Fee:             types.XRPCurrencyAmount(12),
					Sequence:        1,
				},
				Amount: test.amount,
				Holder: test.holder,
			}

			valid, err := clawback.Validate()
			require.NoError(t, err)
			require.True(t, valid)

			txBlob, hash, err := issuer.Sign(clawback.Flatten())
			require.NoError(t, err)
			require.NotEmpty(t, txBlob)
			require.NotEmpty(t, hash)

			decoded, err := binarycodec.Decode(txBlob)
			require.NoError(t, err)
			require.Equal(t, "Clawback", decoded["TransactionType"])
			require.Equal(t, issuer.ClassicAddress.String(), decoded["Account"])
			require.NotEmpty(t, decoded["TxnSignature"])
			require.NotNil(t, decoded["Amount"])
			if test.amount.Kind() == types.ISSUED {
				decodedAmount, ok := decoded["Amount"].(map[string]any)
				require.True(t, ok)
				require.Equal(t, holder.ClassicAddress.String(), decodedAmount["issuer"])
			}
			if test.expectedHolder == "" {
				require.NotContains(t, decoded, "Holder")
			} else {
				require.Equal(t, test.expectedHolder.String(), decoded["Holder"])
			}
		})
	}
}
