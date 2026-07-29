package transaction_test

import (
	"testing"

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

	tests := []struct {
		name   string
		amount types.CurrencyAmount
		holder types.Address
	}{
		{
			name: "issued currency",
			amount: types.IssuedCurrencyAmount{
				Issuer:   holder.ClassicAddress,
				Currency: "USD",
				Value:    "1",
			},
		},
		{
			name: "MPT",
			amount: types.MPTCurrencyAmount{
				MPTIssuanceID: mptIssuanceID,
				Value:         "10",
			},
			holder: holder.ClassicAddress,
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
			if test.holder == "" {
				require.NotContains(t, decoded, "Holder")
			} else {
				require.Equal(t, holder.ClassicAddress.String(), decoded["Holder"])
			}
		})
	}
}
