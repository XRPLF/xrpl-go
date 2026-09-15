package transaction

import (
	"encoding/json"
	"strings"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

// These pairs exercise account identity without changing field-specific tag policies.
func accountIdentityPairs(t *testing.T) []struct {
	name            string
	account, other  types.Address
	same, malformed bool
} {
	t.Helper()
	const account = "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	const other = "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"
	x, err := addresscodec.ClassicAddressToXAddress(account, 0, false, false)
	require.NoError(t, err)
	testX, err := addresscodec.ClassicAddressToXAddress(account, 0, false, true)
	require.NoError(t, err)
	zeroTagX, err := addresscodec.ClassicAddressToXAddress(account, 0, true, false)
	require.NoError(t, err)
	tagX, err := addresscodec.ClassicAddressToXAddress(account, 42, true, true)
	require.NoError(t, err)
	otherX, err := addresscodec.ClassicAddressToXAddress(other, 0, false, true)
	require.NoError(t, err)
	otherTagX, err := addresscodec.ClassicAddressToXAddress(other, 42, true, true)
	require.NoError(t, err)

	return []struct {
		name            string
		account, other  types.Address
		same, malformed bool
	}{
		{"same classic/classic", account, account, true, false},
		{"same classic/X", account, types.Address(x), true, false},
		{"same X/classic", types.Address(x), account, true, false},
		{"same X/X networks", types.Address(x), types.Address(testX), true, false},
		{"same explicit zero tag", account, types.Address(zeroTagX), true, false},
		{"same different tags", types.Address(zeroTagX), types.Address(tagX), true, false},
		{"different classic/classic", account, other, false, false},
		{"different classic/X", account, types.Address(otherX), false, false},
		{"different X/classic", types.Address(x), other, false, false},
		{"different X/X", types.Address(x), types.Address(otherX), false, false},
		{"different tagged X", types.Address(tagX), types.Address(otherTagX), false, false},
		{"malformed field", types.Address(x), "not-an-address", false, true},
	}
}

func TestTransactionAccountIdentity(t *testing.T) {
	for _, pair := range accountIdentityPairs(t) {
		t.Run(pair.name, func(t *testing.T) {
			base := BaseTx{Account: pair.account, Fee: 12}
			depositBase := base
			depositBase.TransactionType = DepositPreauthTx
			offerBase := base
			offerBase.TransactionType = NFTokenCreateOfferTx
			sellBase := offerBase
			sellBase.Flags = TfSellNFToken
			keyBase := base
			keyBase.TransactionType = SetRegularKeyTx
			delegateBase := base
			delegateBase.TransactionType = DelegateSetTx
			mintBase := base
			mintBase.TransactionType = NFTokenMintTx
			modifyBase := base
			modifyBase.TransactionType = NFTokenModifyTx
			authorizeBase := base
			authorizeBase.TransactionType = MPTokenAuthorizeTx
			setBase := base
			setBase.TransactionType = MPTokenIssuanceSetTx
			setBase.Flags = TfMPTLock
			const issuanceID = "000004C463C52827307480341125DA0577DEFC38405B0E3E"
			nftID := types.NFTokenID(strings.Repeat("A", 64))

			cases := []struct {
				name              string
				tx                interface{ Validate() (bool, error) }
				conflict, invalid error
			}{
				{"DepositPreauth Authorize", &DepositPreauth{BaseTx: depositBase, Authorize: pair.other}, ErrDepositPreauthAuthorizeCannotBeSender, ErrDepositPreauthInvalidAuthorize},
				// Retain the SDK's self-Unauthorize rule, also enforced by xrpl.js.
				{"DepositPreauth Unauthorize", &DepositPreauth{BaseTx: depositBase, Unauthorize: pair.other}, ErrDepositPreauthUnauthorizeCannotBeSender, ErrDepositPreauthInvalidUnauthorize},
				{"NFTokenCreateOffer Owner", &NFTokenCreateOffer{BaseTx: offerBase, NFTokenID: nftID, Amount: types.XRPCurrencyAmount(1), Owner: pair.other}, ErrOwnerAccountConflict, ErrInvalidOwner},
				{"NFTokenCreateOffer Destination", &NFTokenCreateOffer{BaseTx: sellBase, NFTokenID: nftID, Amount: types.XRPCurrencyAmount(1), Destination: pair.other}, ErrDestinationAccountConflict, ErrInvalidDestination},
				{"SetRegularKey", &SetRegularKey{BaseTx: keyBase, RegularKey: pair.other}, ErrRegularKeyMatchesAccount, ErrInvalidRegularKey},
				{"DelegateSet", &DelegateSet{BaseTx: delegateBase, Authorize: pair.other, Permissions: []types.Permission{}}, ErrDelegateSetAuthorizeAccountConflict, ErrInvalidDestination},
				{"NFTokenMint", &NFTokenMint{BaseTx: mintBase, Issuer: pair.other}, ErrIssuerAccountConflict, ErrInvalidIssuer},
				{"NFTokenModify", &NFTokenModify{BaseTx: modifyBase, NFTokenID: nftID, Owner: pair.other}, ErrOwnerAccountConflict, ErrInvalidOwner},
				{"MPTokenAuthorize", &MPTokenAuthorize{BaseTx: authorizeBase, MPTokenIssuanceID: issuanceID, Holder: &pair.other}, ErrHolderAccountConflict, ErrInvalidAccount},
				{"MPTokenIssuanceSet", &MPTokenIssuanceSet{BaseTx: setBase, MPTokenIssuanceID: issuanceID, Holder: &pair.other}, ErrHolderAccountConflict, ErrInvalidAccount},
			}
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					var wantErr error
					if pair.same {
						wantErr = tc.conflict
					} else if pair.malformed {
						wantErr = tc.invalid
					}
					before, err := json.Marshal(tc.tx)
					require.NoError(t, err)
					ok, err := tc.tx.Validate()
					require.ErrorIs(t, err, wantErr)
					require.Equal(t, wantErr == nil, ok)
					after, err := json.Marshal(tc.tx)
					require.NoError(t, err)
					require.Equal(t, before, after, "validation must not rewrite address fields")
				})
			}
		})
	}
}

func TestAMMClawbackAccountIdentity(t *testing.T) {
	for _, pair := range accountIdentityPairs(t) {
		t.Run(pair.name, func(t *testing.T) {
			tx := AMMClawback{
				BaseTx: BaseTx{Account: pair.account, TransactionType: AMMClawbackTx, Fee: 12},
				Holder: "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
				Asset:  types.IssuedCurrency{Currency: "USD", Issuer: pair.other},
				Asset2: types.XRPCurrencyAmount(1),
			}
			var wantErr error
			if !pair.same {
				wantErr = ErrInvalidAssetIssuer
			}
			before := tx
			ok, err := tx.Validate()
			require.ErrorIs(t, err, wantErr)
			require.Equal(t, wantErr == nil, ok)
			require.Equal(t, before, tx)
		})
	}
}
