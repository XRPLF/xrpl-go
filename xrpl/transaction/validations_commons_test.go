package transaction

import (
	"encoding/json"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

// orderedTransactionSigners builds nonzero AccountIDs in byte order without
// depending on validation or wallet sorting to establish the expected order.
func orderedTransactionSigners(t *testing.T, count int) []types.Signer {
	t.Helper()
	signers := make([]types.Signer, count)
	for i := range signers {
		id := make([]byte, 20)
		id[19] = byte(i + 1)
		address, err := addresscodec.EncodeAccountIDToClassicAddress(id)
		require.NoError(t, err)
		signers[i] = types.Signer{SignerData: types.SignerData{
			Account:       types.Address(address),
			SigningPubKey: "ED5F5AC8B98974A3CA843326D9B88CEBD0560177B973EE0B149F782CFAA06DC66A",
			TxnSignature:  "CD",
		}}
	}
	return signers
}

func TestIsSignaturePair(t *testing.T) {
	const key = "ED5F5AC8B98974A3CA843326D9B88CEBD0560177B973EE0B149F782CFAA06DC66A"

	tests := []struct {
		name      string
		key       string
		signature string
		want      bool
	}{
		{"pass - ed25519 key and hex signature", key, "ABCD", true},
		{"pass - secp256k1 key and hex signature", "03ADB44CA8E56F78A0096825E5667C450ABD5C24C34E027BC1AAF7E5BD114CB5B5", "ABCD", true},
		{"fail - empty key", "", "ABCD", false},
		{"fail - hex that is not a public key", "ABCD", "ABCD", false},
		{"fail - secp256k1 key that is not on the curve", "030000000000000000000000000000000000000000000000000000000000000000", "ABCD", false},
		{"fail - empty signature", key, "", false},
		{"fail - signature is not hex", key, "not-hex", false},
		{"fail - signature is not whole bytes", key, "ABC", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, isSignaturePair(tt.key, tt.signature))
		})
	}
}

func TestValidateSigners(t *testing.T) {
	sorted := orderedTransactionSigners(t, 33)
	mainnet, err := addresscodec.ClassicAddressToXAddress(sorted[0].SignerData.Account.String(), 0, false, false)
	require.NoError(t, err)
	testnet, err := addresscodec.ClassicAddressToXAddress(sorted[0].SignerData.Account.String(), 0, false, true)
	require.NoError(t, err)
	secondXAddress, err := addresscodec.ClassicAddressToXAddress(sorted[1].SignerData.Account.String(), 0, false, false)
	require.NoError(t, err)
	mainnetSigner, testnetSigner, secondSigner := sorted[0], sorted[0], sorted[1]
	mainnetSigner.SignerData.Account = types.Address(mainnet)
	testnetSigner.SignerData.Account = types.Address(testnet)
	secondSigner.SignerData.Account = types.Address(secondXAddress)
	// Text order is descending, but AccountID byte order is ascending.
	require.Greater(t, sorted[0].SignerData.Account.String(), secondXAddress)

	tests := []struct {
		name    string
		signers []types.Signer
		wantErr error
	}{
		{"nil", nil, nil},
		{"empty", []types.Signer{}, nil},
		{"one", sorted[:1], nil},
		{"maximum", sorted[:32], nil},
		{"above maximum", sorted, errTooManyTransactionSigners},
		{"descending", []types.Signer{sorted[1], sorted[0]}, errUnsortedTransactionSigners},
		{"adjacent duplicate", []types.Signer{sorted[0], sorted[0]}, errDuplicateTransactionSigner},
		{"nonadjacent duplicate", []types.Signer{sorted[0], sorted[1], sorted[0]}, errUnsortedTransactionSigners},
		{"classic and X address duplicate", []types.Signer{sorted[0], mainnetSigner}, errDuplicateTransactionSigner},
		{"mainnet and testnet duplicate", []types.Signer{mainnetSigner, testnetSigner}, errDuplicateTransactionSigner},
		{"byte order not text order", []types.Signer{sorted[0], secondSigner}, nil},
		{"invalid entry", []types.Signer{{}}, ErrSignerShouldHaveThreeFields},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before, err := json.Marshal(tt.signers)
			require.NoError(t, err)
			err = validateSigners(tt.signers)
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
			after, err := json.Marshal(tt.signers)
			require.NoError(t, err)
			require.Equal(t, string(before), string(after), "validation must not mutate signers")
		})
	}
}

func TestTransactionAndSponsorShareSignerListRules(t *testing.T) {
	sorted := orderedTransactionSigners(t, 33)
	tests := []struct {
		name    string
		signers []types.Signer
		wantErr error
	}{
		{"valid", sorted[:32], nil},
		{"too many", sorted, errTooManyTransactionSigners},
		{"unsorted", []types.Signer{sorted[1], sorted[0]}, errUnsortedTransactionSigners},
		{"duplicate", []types.Signer{sorted[0], sorted[0]}, errDuplicateTransactionSigner},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := BaseTx{
				Account:         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				TransactionType: PaymentTx,
				Signers:         tt.signers,
			}
			sponsored := BaseTx{
				Account:          tx.Account,
				TransactionType:  PaymentTx,
				Sponsor:          "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				SponsorFlags:     types.SpfSponsorFee,
				SponsorSignature: &types.SponsorSignature{Signers: tt.signers},
			}
			valid, err := tx.Validate()
			sponsorValid, sponsorErr := sponsored.Validate()
			require.Equal(t, tt.wantErr == nil, valid)
			require.Equal(t, tt.wantErr == nil, sponsorValid)
			if tt.wantErr == nil {
				require.NoError(t, err)
				require.NoError(t, sponsorErr)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
				require.ErrorIs(t, sponsorErr, tt.wantErr)
			}
		})
	}
}
