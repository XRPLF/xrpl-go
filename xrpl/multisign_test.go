package xrpl

import (
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/stretchr/testify/require"
)

const (
	signedBlobA       = "12000324002EAF3B201B002EFC826840000000000000247300770B6578616D706C652E636F6D8114226DADFAA52D198160EF96B7AFD8B04E49B8FE8AF3E0107321ED4CC509EF081781B7F562A216A1C19F5FFDC8EA4F3E0D1FB2D153A5E55F88346174400BA2FE2E0C220B635F3CDC4BFEB07CE1EC197EC4E33AF3F5E6FBD4A3C58381309EAC3C326943F7F144A60C9B8161A7CBB5AF289385EA22DD059ED80A481D510A8114D1AEB96AE693F85A1004968E62AF03759B7949FCE1F1"
	signedBlobB       = "12000324002EAF3B201B002EFC826840000000000000247300770B6578616D706C652E636F6D8114226DADFAA52D198160EF96B7AFD8B04E49B8FE8AF3E0107321ED043A4565F23BBD51138F204C22B0D42F2A8D7C2D85D6A5B7DD62A4FA6C1EB2867440A17FE3A80C980D8BAA5FCF93E658011C1CA1BA296BC408354C4D2DE33AF68E83FE389080802D5D93C87997A340D7BE61C77A36F348CD0D0B23B229F3CD1CE8008114318352A65A18305C82983EE7005C051C35EAA651E1F1"
	multisignedBlobAB = "12000324002EAF3B201B002EFC826840000000000000247300770B6578616D706C652E636F6D8114226DADFAA52D198160EF96B7AFD8B04E49B8FE8AF3E0107321ED043A4565F23BBD51138F204C22B0D42F2A8D7C2D85D6A5B7DD62A4FA6C1EB2867440A17FE3A80C980D8BAA5FCF93E658011C1CA1BA296BC408354C4D2DE33AF68E83FE389080802D5D93C87997A340D7BE61C77A36F348CD0D0B23B229F3CD1CE8008114318352A65A18305C82983EE7005C051C35EAA651E1E0107321ED4CC509EF081781B7F562A216A1C19F5FFDC8EA4F3E0D1FB2D153A5E55F88346174400BA2FE2E0C220B635F3CDC4BFEB07CE1EC197EC4E33AF3F5E6FBD4A3C58381309EAC3C326943F7F144A60C9B8161A7CBB5AF289385EA22DD059ED80A481D510A8114D1AEB96AE693F85A1004968E62AF03759B7949FCE1F1"
)

func TestMultisign(t *testing.T) {
	testCases := []struct {
		name        string
		blobs       []string
		want        string
		wantSigners int
		err         error
	}{
		{
			name:        "pass - valid blobs",
			blobs:       []string{signedBlobA, signedBlobB},
			want:        multisignedBlobAB,
			wantSigners: 2,
		},
		{
			name:  "fail - no blobs",
			blobs: []string{},
			err:   ErrNoTxToMultisign,
		},
		{
			name:  "fail - signer missing Account returns ErrInvalidSigner",
			blobs: []string{"12000024000000016140000000000F424068400000000000000C81149A51260615192AF5A94692D5F02EAB105D129F5183147990EC5D1D8DF69E070A968D4B186986FDF06ED0F3E0107321ED4CC509EF081781B7F562A216A1C19F5FFDC8EA4F3E0D1FB2D153A5E55F88346174400BA2FE2E0C220B635F3CDC4BFEB07CE1EC197EC4E33AF3F5E6FBD4A3C58381309EAC3C326943F7F144A60C9B8161A7CBB5AF289385EA22DD059ED80A481D510AE1F1"},
			err:   ErrInvalidSigner,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := Multisign(tc.blobs...)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, res)

			if tc.wantSigners == 0 {
				return
			}

			tx, err := binarycodec.Decode(res)
			require.NoError(t, err)
			require.Len(t, tx["Signers"], tc.wantSigners)
		})
	}
}

func TestSortByAccountID(t *testing.T) {
	testCases := []struct {
		name        string
		accounts    []string
		account     func(string) (string, error)
		want        []string
		expectedErr error
	}{
		{
			name: "pass - sorts by account ID bytes",
			accounts: []string{
				"raHgU3KRBN6XYbEhi5JyJELSHtshTenYw",
				"rrrrrrrrrrrrrrrrrrrrBZbvji",
			},
			account: func(account string) (string, error) {
				return account, nil
			},
			want: []string{
				"rrrrrrrrrrrrrrrrrrrrBZbvji",
				"raHgU3KRBN6XYbEhi5JyJELSHtshTenYw",
			},
		},
		{
			name: "fail - returns error before sorting",
			accounts: []string{
				"raHgU3KRBN6XYbEhi5JyJELSHtshTenYw",
				"rrrrrrrrrrrrrrrrrrrrBZbvji",
			},
			account: func(account string) (string, error) {
				if account == "rrrrrrrrrrrrrrrrrrrrBZbvji" {
					return "", ErrInvalidSigner
				}
				return account, nil
			},
			want: []string{
				"raHgU3KRBN6XYbEhi5JyJELSHtshTenYw",
				"rrrrrrrrrrrrrrrrrrrrBZbvji",
			},
			expectedErr: ErrInvalidSigner,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := SortByAccountID(tc.accounts, tc.account)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.want, tc.accounts)
		})
	}
}

func TestSortSignersReturnsErrors(t *testing.T) {
	testCases := []struct {
		name        string
		signers     []any
		expectedErr error
	}{
		{
			name:        "fail - invalid signer",
			signers:     []any{"invalid"},
			expectedErr: ErrInvalidSigner,
		},
		{
			name: "fail - invalid address",
			signers: []any{
				map[string]any{
					"Signer": map[string]any{
						"Account": "invalid",
					},
				},
			},
			expectedErr: addresscodec.ErrInvalidClassicAddress,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := SortSigners(tc.signers)

			require.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestMultisignRejectsInvalidInputs(t *testing.T) {
	testCases := []struct {
		name    string
		blobs   func(t *testing.T) []string
		wantErr error
	}{
		{
			name: "invalid signer signature",
			blobs: func(t *testing.T) []string {
				blob := mutateBlob(t, signedBlobB, func(tx map[string]any) {
					firstSigner(tx)["TxnSignature"] = "00"
				})
				return []string{signedBlobA, blob}
			},
			wantErr: ErrMultisignInvalidSignature,
		},
		{
			name: "invalid signer public key",
			blobs: func(t *testing.T) []string {
				blob := mutateBlob(t, signedBlobA, func(tx map[string]any) {
					firstSigner(tx)["SigningPubKey"] = "ABCD"
				})
				return []string{blob}
			},
			wantErr: ErrMultisignInvalidSignature,
		},
		{
			name: "mismatched transaction body",
			blobs: func(t *testing.T) []string {
				blob := mutateBlob(t, signedBlobB, func(tx map[string]any) {
					tx["Amount"] = "2000000"
				})
				return []string{signedBlobA, blob}
			},
			wantErr: ErrMultisignTxNotEqual,
		},
		{
			name: "non-empty SigningPubKey",
			blobs: func(t *testing.T) []string {
				blob := mutateBlob(t, signedBlobA, func(tx map[string]any) {
					tx["SigningPubKey"] = firstSigner(tx)["SigningPubKey"]
				})
				return []string{blob}
			},
			wantErr: ErrMultisignNonEmptySigningPubKey,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := Multisign(tc.blobs(t)...)
			require.ErrorIs(t, err, tc.wantErr)
			require.Empty(t, res)
		})
	}
}

func TestMultisignRejectsMalformedSignerData(t *testing.T) {
	testCases := []struct {
		name   string
		mutate func(tx map[string]any)
	}{
		{
			name: "missing Signers",
			mutate: func(tx map[string]any) {
				delete(tx, "Signers")
			},
		},
		{
			name: "empty Signers",
			mutate: func(tx map[string]any) {
				tx["Signers"] = []any{}
			},
		},
		{
			name: "missing Account",
			mutate: func(tx map[string]any) {
				delete(firstSigner(tx), "Account")
			},
		},
		{
			name: "missing SigningPubKey",
			mutate: func(tx map[string]any) {
				delete(firstSigner(tx), "SigningPubKey")
			},
		},
		{
			name: "missing TxnSignature",
			mutate: func(tx map[string]any) {
				delete(firstSigner(tx), "TxnSignature")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			blob := mutateBlob(t, signedBlobA, tc.mutate)
			res, err := Multisign(blob)
			require.ErrorIs(t, err, ErrInvalidSigner)
			require.Empty(t, res)
		})
	}
}

func mutateBlob(t *testing.T, blob string, mutate func(tx map[string]any)) string {
	t.Helper()

	tx, err := binarycodec.Decode(blob)
	require.NoError(t, err)

	mutate(tx)

	mutatedBlob, err := binarycodec.Encode(tx)
	require.NoError(t, err)
	return mutatedBlob
}

func firstSigner(tx map[string]any) map[string]any {
	signers := tx["Signers"].([]any)
	signerData := signers[0].(map[string]any)["Signer"].(map[string]any)
	return signerData
}

func TestMultisignRejectsInvalidInputs(t *testing.T) {
	testCases := []struct {
		name    string
		blobs   func(t *testing.T) []string
		wantErr error
	}{
		{
			name: "invalid signer signature",
			blobs: func(t *testing.T) []string {
				blob := mutateBlob(t, signedBlobB, func(tx map[string]any) {
					firstSigner(tx)["TxnSignature"] = "00"
				})
				return []string{signedBlobA, blob}
			},
			wantErr: ErrMultisignInvalidSignature,
		},
		{
			name: "invalid signer public key",
			blobs: func(t *testing.T) []string {
				blob := mutateBlob(t, signedBlobA, func(tx map[string]any) {
					firstSigner(tx)["SigningPubKey"] = "ABCD"
				})
				return []string{blob}
			},
			wantErr: ErrMultisignInvalidSignature,
		},
		{
			name: "mismatched transaction body",
			blobs: func(t *testing.T) []string {
				blob := mutateBlob(t, signedBlobB, func(tx map[string]any) {
					tx["Amount"] = "2000000"
				})
				return []string{signedBlobA, blob}
			},
			wantErr: ErrMultisignTxNotEqual,
		},
		{
			name: "non-empty SigningPubKey",
			blobs: func(t *testing.T) []string {
				blob := mutateBlob(t, signedBlobA, func(tx map[string]any) {
					tx["SigningPubKey"] = firstSigner(tx)["SigningPubKey"]
				})
				return []string{blob}
			},
			wantErr: ErrMultisignNonEmptySigningPubKey,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := Multisign(tc.blobs(t)...)
			require.ErrorIs(t, err, tc.wantErr)
			require.Empty(t, res)
		})
	}
}

func TestMultisignRejectsMalformedSignerData(t *testing.T) {
	testCases := []struct {
		name   string
		mutate func(tx map[string]any)
	}{
		{
			name: "missing Signers",
			mutate: func(tx map[string]any) {
				delete(tx, "Signers")
			},
		},
		{
			name: "empty Signers",
			mutate: func(tx map[string]any) {
				tx["Signers"] = []any{}
			},
		},
		{
			name: "missing Account",
			mutate: func(tx map[string]any) {
				delete(firstSigner(tx), "Account")
			},
		},
		{
			name: "missing SigningPubKey",
			mutate: func(tx map[string]any) {
				delete(firstSigner(tx), "SigningPubKey")
			},
		},
		{
			name: "missing TxnSignature",
			mutate: func(tx map[string]any) {
				delete(firstSigner(tx), "TxnSignature")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			blob := mutateBlob(t, signedBlobA, tc.mutate)
			res, err := Multisign(blob)
			require.ErrorIs(t, err, ErrMultisignInvalidSigner)
			require.Empty(t, res)
		})
	}
}

func mutateBlob(t *testing.T, blob string, mutate func(tx map[string]any)) string {
	t.Helper()

	tx, err := binarycodec.Decode(blob)
	require.NoError(t, err)

	mutate(tx)

	mutatedBlob, err := binarycodec.Encode(tx)
	require.NoError(t, err)
	return mutatedBlob
}

func firstSigner(tx map[string]any) map[string]any {
	signers := tx["Signers"].([]any)
	signerData := signers[0].(map[string]any)["Signer"].(map[string]any)
	return signerData
}
