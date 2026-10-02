package wallet

import (
	"encoding/hex"
	"fmt"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/keypairs"
	"github.com/Peersyst/xrpl-go/xrpl"
	"github.com/Peersyst/xrpl-go/xrpl/hash"
	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func sponsorTestWallet(t *testing.T, seed string) Wallet {
	t.Helper()

	w, err := FromSeed(seed, "")
	require.NoError(t, err)

	return w
}

func sponsorAccountSignedTx(t *testing.T, sponsor Wallet, multi bool) (transaction.FlatTransaction, Wallet) {
	t.Helper()

	account := sponsorTestWallet(t, brokerSeed)
	tx := transaction.FlatTransaction{
		"TransactionType": "Payment", "Account": account.ClassicAddress.String(),
		"Destination": counterpartyOverrideAccount, "Amount": "1",
		"Sponsor": sponsor.ClassicAddress.String(), "SponsorFlags": types.SpfSponsorFee,
		"Fee": "100", "Sequence": uint32(1),
	}

	var blob string
	var err error
	if multi {
		// Account names the signer-list owner, not the individual signing wallet.
		tx["Account"] = "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59"
		blob, _, err = account.Multisign(tx)
	} else {
		blob, _, err = account.Sign(tx)
	}

	require.NoError(t, err)

	decoded, err := binarycodec.Decode(blob)
	require.NoError(t, err)

	return transaction.FlatTransaction(decoded), account
}

func requireSponsorPayloadVerification(t *testing.T, key, signature, payload string, want bool) {
	t.Helper()

	bytes, err := hex.DecodeString(payload)
	require.NoError(t, err)

	valid, err := keypairs.Validate(string(bytes), key, signature)
	require.NoError(t, err)
	require.Equal(t, want, valid)
}

func requireSponsorBlobHash(t *testing.T, tx transaction.FlatTransaction, blob, txHash string) {
	t.Helper()

	encoded, err := binarycodec.Encode(tx)
	require.NoError(t, err)
	require.Equal(t, encoded, blob)

	wantHash, err := hash.SignTxBlob(blob)
	require.NoError(t, err)
	require.Equal(t, wantHash, txHash)
}

// All account/sponsor forms must retain the account payload and separate sponsor roles.
func TestSignAsSponsorRoles(t *testing.T) {
	for _, seed := range []string{counterpartySeed, counterpartySecp256k1Seed} {
		for _, accountMulti := range []bool{false, true} {
			for _, sponsorMulti := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/accountMulti=%t/sponsorMulti=%t", seed, accountMulti, sponsorMulti), func(t *testing.T) {
					sponsor := sponsorTestWallet(t, seed)
					tx, account := sponsorAccountSignedTx(t, sponsor, accountMulti)
					before := clientinternal.CloneTransaction(tx)
					opts := &SignAsSponsorOptions{Multisign: sponsorMulti}
					signed, blob, txHash, err := SignAsSponsor(sponsor, tx, opts)
					require.NoError(t, err)
					require.Equal(t, before, map[string]any(tx))
					requireSponsorBlobHash(t, signed, blob, txHash)

					for _, field := range []string{"SigningPubKey", "TxnSignature", "Signers"} {
						require.Equal(t, tx[field], signed[field])
					}

					var accountPayload, accountSignature string
					if accountMulti {
						accountPayload, err = binarycodec.EncodeForMultisigning(signed, account.ClassicAddress.String())
						accountSignature = signed["Signers"].([]any)[0].(map[string]any)["Signer"].(map[string]any)["TxnSignature"].(string)
					} else {
						accountPayload, err = binarycodec.EncodeForSigning(signed)
						accountSignature = signed["TxnSignature"].(string)
					}

					require.NoError(t, err)
					requireSponsorPayloadVerification(t, account.PublicKey, accountSignature, accountPayload, true)

					var rolePayload, ordinaryPayload, counterpartyPayload string
					sigObject := signed["SponsorSignature"].(map[string]any)
					if sponsorMulti {
						sigObject = sigObject["Signers"].([]any)[0].(map[string]any)["Signer"].(map[string]any)
						rolePayload, err = binarycodec.EncodeForMultisigningSponsor(signed, sponsor.ClassicAddress.String())
						require.NoError(t, err)

						ordinaryPayload, err = binarycodec.EncodeForMultisigning(signed, sponsor.ClassicAddress.String())
						require.NoError(t, err)

						counterpartyPayload, err = binarycodec.EncodeForMultisigningCounterparty(signed, sponsor.ClassicAddress.String())
					} else {
						rolePayload, err = binarycodec.EncodeForSigningSponsor(signed)
						require.NoError(t, err)

						ordinaryPayload, err = binarycodec.EncodeForSigning(signed)
						require.NoError(t, err)

						counterpartyPayload, err = binarycodec.EncodeForSigningCounterparty(signed)
					}

					require.NoError(t, err)

					signature := sigObject["TxnSignature"].(string)
					requireSponsorPayloadVerification(t, sponsor.PublicKey, signature, rolePayload, true)
					requireSponsorPayloadVerification(t, sponsor.PublicKey, signature, ordinaryPayload, false)
					requireSponsorPayloadVerification(t, sponsor.PublicKey, signature, counterpartyPayload, false)

					accountBlob, err := binarycodec.Encode(tx)
					require.NoError(t, err)

					fromBlob, sameBlob, sameHash, err := SignAsSponsorBlob(sponsor, accountBlob, opts)
					require.NoError(t, err)
					require.Equal(t, signed, fromBlob)
					require.Equal(t, blob, sameBlob)
					require.Equal(t, txHash, sameHash)
				})
			}
		}
	}
}

func TestSignAsSponsorDeliverMax(t *testing.T) {
	sponsor := sponsorTestWallet(t, counterpartySeed)
	tests := []struct {
		name                       string
		accountMulti, sponsorMulti bool
		keepAmount                 bool
	}{
		{name: "single account and sponsor"},
		{name: "multisigned sponsor", sponsorMulti: true},
		{name: "multisigned account", accountMulti: true},
		{name: "multisigned account and sponsor", accountMulti: true, sponsorMulti: true},
		{name: "identical Amount and DeliverMax", keepAmount: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, account := sponsorAccountSignedTx(t, sponsor, tt.accountMulti)
			amount := tx["Amount"]
			tx["DeliverMax"] = amount
			if !tt.keepAmount {
				delete(tx, "Amount")
			}
			before := clientinternal.CloneTransaction(tx)
			signed, blob, txHash, err := SignAsSponsor(sponsor, tx, &SignAsSponsorOptions{Multisign: tt.sponsorMulti})
			require.NoError(t, err)
			require.Equal(t, before, map[string]any(tx))
			requireSponsorBlobHash(t, signed, blob, txHash)
			decoded, err := binarycodec.Decode(blob)
			require.NoError(t, err)

			var payload, signature string
			if tt.accountMulti {
				payload, err = binarycodec.EncodeForMultisigning(decoded, account.ClassicAddress.String())
				signature = decoded["Signers"].([]any)[0].(map[string]any)["Signer"].(map[string]any)["TxnSignature"].(string)
			} else {
				payload, err = binarycodec.EncodeForSigning(decoded)
				signature = decoded["TxnSignature"].(string)
			}
			require.NoError(t, err)
			requireSponsorPayloadVerification(t, account.PublicKey, signature, payload, true)
			require.Equal(t, amount, decoded["Amount"])
			require.Equal(t, amount, signed["Amount"])
			require.NotContains(t, signed, "DeliverMax")
		})
	}
}

func TestSignAsSponsorRejectsInvalidInput(t *testing.T) {
	sponsor := sponsorTestWallet(t, counterpartySeed)
	base, _ := sponsorAccountSignedTx(t, sponsor, false)
	tests := []struct {
		name, field string
		value       any
		remove      bool
		wantErr     error
	}{
		{"already signed", "SponsorSignature", nil, false, ErrSponsorAlreadySigned},
		{"missing account signature", "TxnSignature", nil, true, ErrAccountMustSignFirst},
		{"empty account signature", "TxnSignature", "", false, ErrAccountMustSignFirst},
		{"null account key", "SigningPubKey", nil, false, ErrAccountMustSignFirst},
		{"mixed account", "Signers", []any{}, false, ErrAccountMustSignFirst},
		{"inner batch", "Flags", types.TfInnerBatchTxn, false, ErrAccountMustSignFirst},
		{"missing sponsor", "Sponsor", nil, true, transaction.ErrSponsorFieldsMissing},
		{"missing flags", "SponsorFlags", nil, true, transaction.ErrSponsorFieldsMissing},
		{"null flags", "SponsorFlags", nil, false, transaction.ErrInvalidSponsorFlags},
		{"unknown flags", "SponsorFlags", 4, false, transaction.ErrInvalidSponsorFlags},
		{"self", "Sponsor", base["Account"], false, transaction.ErrSponsorAccountConflict},
		{"zero sponsor", "Sponsor", "rrrrrrrrrrrrrrrrrrrrrhoLvTp", false, transaction.ErrSponsorZero},
		{"wallet mismatch", "Sponsor", counterpartyOverrideAccount, false, ErrSponsorWalletMismatch},
		{"invalid flags type", "Flags", 1.5, false, transaction.ErrInvalidFlagsValue},
		{"conflicting amounts", "DeliverMax", "2", false, ErrAmountAndDeliverMaxMustBeIdentical},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := transaction.FlatTransaction(clientinternal.CloneTransaction(base))
			if tt.remove {
				delete(tx, tt.field)
			} else {
				tx[tt.field] = tt.value
			}

			before := clientinternal.CloneTransaction(tx)
			got, blob, txHash, err := SignAsSponsor(sponsor, tx, nil)
			require.ErrorIs(t, err, tt.wantErr)
			require.Nil(t, got)
			require.Empty(t, blob)
			require.Empty(t, txHash)
			require.Equal(t, before, map[string]any(tx))
		})
	}

	_, _, _, err := SignAsSponsor(sponsor, nil, nil)
	require.ErrorIs(t, err, ErrNilTransaction)
}

func TestSignAsSponsorAddressOptions(t *testing.T) {
	sponsor := sponsorTestWallet(t, counterpartySeed)
	base, _ := sponsorAccountSignedTx(t, sponsor, false)
	for _, testnet := range []bool{false, true} {
		t.Run(fmt.Sprintf("testnet=%t", testnet), func(t *testing.T) {
			xAddress, err := addresscodec.ClassicAddressToXAddress(sponsor.ClassicAddress.String(), 0, false, testnet)
			require.NoError(t, err)

			tx := transaction.FlatTransaction(clientinternal.CloneTransaction(base))
			tx["Sponsor"] = xAddress
			_, _, _, err = SignAsSponsor(sponsor, tx, nil)
			require.NoError(t, err, "equivalent sponsor identities must match")

			override, err := addresscodec.ClassicAddressToXAddress(counterpartyOverrideAccount, 0, false, testnet)
			require.NoError(t, err)

			signed, _, _, err := SignAsSponsor(sponsor, tx, &SignAsSponsorOptions{MultisignAccount: override})
			require.NoError(t, err)

			signer := signed["SponsorSignature"].(map[string]any)["Signers"].([]any)[0].(map[string]any)["Signer"].(map[string]any)
			require.Equal(t, counterpartyOverrideAccount, signer["Account"])

			payload, err := binarycodec.EncodeForMultisigningSponsor(signed, counterpartyOverrideAccount)
			require.NoError(t, err)
			requireSponsorPayloadVerification(t, sponsor.PublicKey, signer["TxnSignature"].(string), payload, true)

			wrongPayload, err := binarycodec.EncodeForMultisigningSponsor(signed, sponsor.ClassicAddress.String())
			require.NoError(t, err)
			requireSponsorPayloadVerification(t, sponsor.PublicKey, signer["TxnSignature"].(string), wrongPayload, false)
		})
	}

	tagged, err := addresscodec.ClassicAddressToXAddress(sponsor.ClassicAddress.String(), 0, true, false)
	require.NoError(t, err)

	for _, tt := range []struct {
		address string
		wantErr error
	}{
		{tagged, ErrAddressHasTag},
		{"bad", xrpl.ErrInvalidSigner},
		{"rrrrrrrrrrrrrrrrrrrrrhoLvTp", transaction.ErrSignerAccountZero},
	} {
		_, _, _, err = SignAsSponsor(sponsor, base, &SignAsSponsorOptions{MultisignAccount: tt.address})
		require.ErrorIs(t, err, tt.wantErr)
	}

	tx := transaction.FlatTransaction(clientinternal.CloneTransaction(base))
	tx["Sponsor"] = tagged
	_, _, _, err = SignAsSponsor(sponsor, tx, nil)
	require.ErrorIs(t, err, transaction.ErrSponsorTagNotAllowed)

	regular, err := FromSeed(counterparty2Seed, sponsor.ClassicAddress.String())
	require.NoError(t, err)

	signed, _, _, err := SignAsSponsor(regular, base, nil)
	require.NoError(t, err)
	require.Equal(t, regular.PublicKey, signed["SponsorSignature"].(map[string]any)["SigningPubKey"])
}

func TestSignAsSponsorPreservesCounterparty(t *testing.T) {
	sponsor := sponsorTestWallet(t, counterpartySeed)
	tx, account := sponsorAccountSignedTx(t, sponsor, false)
	tx["TransactionType"] = "LoanSet"
	delete(tx, "Destination")
	delete(tx, "Amount")
	blob, _, err := account.Sign(tx)
	require.NoError(t, err)

	decoded, err := binarycodec.Decode(blob)
	require.NoError(t, err)

	tx = transaction.FlatTransaction(decoded)
	counterparty := sponsorTestWallet(t, counterparty2Seed)
	_, _, err = SignLoanSetByCounterparty(counterparty, &tx, nil)
	require.NoError(t, err)

	before := clientinternal.CloneTransaction(tx)
	signed, _, _, err := SignAsSponsor(sponsor, tx, nil)
	require.NoError(t, err)
	require.Equal(t, before, map[string]any(tx))
	require.Equal(t, tx["CounterpartySignature"], signed["CounterpartySignature"])

	payload, err := binarycodec.EncodeForSigningCounterparty(signed)
	require.NoError(t, err)
	requireSponsorPayloadVerification(t, counterparty.PublicKey, signed["CounterpartySignature"].(map[string]any)["TxnSignature"].(string), payload, true)
}

func TestCombineSponsorSigners(t *testing.T) {
	sponsor := sponsorTestWallet(t, counterpartySeed)
	second := sponsorTestWallet(t, counterparty2Seed)
	for _, accountMulti := range []bool{false, true} {
		t.Run(fmt.Sprintf("accountMulti=%t", accountMulti), func(t *testing.T) {
			tx, _ := sponsorAccountSignedTx(t, sponsor, accountMulti)
			first, blob1, _, err := SignAsSponsor(sponsor, tx, &SignAsSponsorOptions{Multisign: true})
			require.NoError(t, err)

			other, blob2, _, err := SignAsSponsor(second, tx, &SignAsSponsorOptions{Multisign: true})
			require.NoError(t, err)

			before1, before2 := clientinternal.CloneTransaction(first), clientinternal.CloneTransaction(other)
			combined, blob, err := CombineSponsorSigners([]transaction.FlatTransaction{other, first, other})
			require.NoError(t, err)
			require.Equal(t, before1, map[string]any(first))
			require.Equal(t, before2, map[string]any(other))

			_, err = transaction.InspectSponsorFields(combined)
			require.NoError(t, err)

			signers := combined["SponsorSignature"].(map[string]any)["Signers"].([]any)
			require.Len(t, signers, 2)

			for _, entry := range signers {
				signer := entry.(map[string]any)["Signer"].(map[string]any)
				payload, err := binarycodec.EncodeForMultisigningSponsor(combined, signer["Account"].(string))
				require.NoError(t, err)
				requireSponsorPayloadVerification(t, signer["SigningPubKey"].(string), signer["TxnSignature"].(string), payload, true)
			}

			fromBlob, sameBlob, err := CombineSponsorSignersBlob([]string{blob1, blob2})
			require.NoError(t, err)
			require.Equal(t, combined, fromBlob)
			require.Equal(t, blob, sameBlob)

			txHash, err := hash.SignTxBlob(blob)
			require.NoError(t, err)
			requireSponsorBlobHash(t, combined, blob, txHash)

			// Mutating a returned nested signer must not alter a source fragment.
			signers[0].(map[string]any)["Signer"].(map[string]any)["TxnSignature"] = "AB"
			require.Equal(t, before1, map[string]any(first))
			require.Equal(t, before2, map[string]any(other))
		})
	}
}

// Conflicting duplicates follow first-input-wins, before the unstable signer sort.
func TestCombineSponsorSignersFirstDecodedIdentityWins(t *testing.T) {
	sponsor := sponsorTestWallet(t, counterpartySeed)
	tx, _ := sponsorAccountSignedTx(t, sponsor, false)
	first, _, _, err := SignAsSponsor(sponsor, tx, &SignAsSponsorOptions{Multisign: true})
	require.NoError(t, err)

	other := transaction.FlatTransaction(clientinternal.CloneTransaction(first))
	xAddress, err := addresscodec.ClassicAddressToXAddress(sponsor.ClassicAddress.String(), 0, false, true)
	require.NoError(t, err)

	otherSigner := other["SponsorSignature"].(map[string]any)["Signers"].([]any)[0].(map[string]any)
	otherSigner["Signer"].(map[string]any)["Account"] = xAddress
	otherSigner["Signer"].(map[string]any)["TxnSignature"] = "AB"
	other["SponsorSignature"].(map[string]any)["Signers"] = []map[string]any{otherSigner}
	for _, inputs := range [][]transaction.FlatTransaction{{first, other}, {other, first}} {
		before := clientinternal.CloneTransaction(inputs[0])
		combined, _, err := CombineSponsorSigners(inputs)
		require.NoError(t, err)
		require.Equal(t, before, map[string]any(inputs[0]))

		signers := combined["SponsorSignature"].(map[string]any)["Signers"].([]any)
		require.Len(t, signers, 1)

		signer := signers[0].(map[string]any)["Signer"].(map[string]any)
		require.Equal(t, sponsor.ClassicAddress.String(), signer["Account"])

		encoded, err := binarycodec.Encode(inputs[0])
		require.NoError(t, err)

		decoded, err := binarycodec.Decode(encoded)
		require.NoError(t, err)

		want := decoded["SponsorSignature"].(map[string]any)["Signers"].([]any)[0].(map[string]any)["Signer"].(map[string]any)
		require.Equal(t, want["TxnSignature"], signer["TxnSignature"])
	}
}

func TestCombineSponsorSignersRejectsMismatches(t *testing.T) {
	sponsor := sponsorTestWallet(t, counterpartySeed)
	tx, _ := sponsorAccountSignedTx(t, sponsor, false)
	first, _, _, err := SignAsSponsor(sponsor, tx, &SignAsSponsorOptions{Multisign: true})
	require.NoError(t, err)

	for _, field := range []string{"Fee", "TxnSignature", "SigningPubKey", "Sponsor", "SponsorFlags"} {
		t.Run(field, func(t *testing.T) {
			other := transaction.FlatTransaction(clientinternal.CloneTransaction(first))
			switch field {
			case "Fee":
				other[field] = "101"
			case "Sponsor":
				other[field] = counterpartyOverrideAccount
			case "SponsorFlags":
				other[field] = types.SpfSponsorReserve
			default:
				other[field] = "AB"
			}

			before := clientinternal.CloneTransaction(other)
			got, blob, err := CombineSponsorSigners([]transaction.FlatTransaction{first, other})
			require.ErrorIs(t, err, ErrSponsorTxNotEqual)
			require.Nil(t, got)
			require.Empty(t, blob)
			require.Equal(t, before, map[string]any(other))
		})
	}

	other := transaction.FlatTransaction(clientinternal.CloneTransaction(first))
	other["SponsorSignature"].(map[string]any)["SigningPubKey"] = ""
	_, _, err = CombineSponsorSigners([]transaction.FlatTransaction{first, other})
	require.ErrorIs(t, err, ErrSponsorTxNotEqual, "remaining sponsor fields participate in equivalence")

	combined, _, err := CombineSponsorSigners([]transaction.FlatTransaction{other})
	require.NoError(t, err)
	require.Contains(t, combined["SponsorSignature"], "SigningPubKey")
}

// Wire equivalence ignores metadata, but never ignores account multisign authorization.
func TestCombineSponsorSignersWireEquivalence(t *testing.T) {
	sponsor := sponsorTestWallet(t, counterpartySeed)
	tx, _ := sponsorAccountSignedTx(t, sponsor, true)
	first, _, _, err := SignAsSponsor(sponsor, tx, &SignAsSponsorOptions{Multisign: true})
	require.NoError(t, err)

	other := transaction.FlatTransaction(clientinternal.CloneTransaction(first))
	other["hash"] = "metadata is not serialized"
	_, _, err = CombineSponsorSigners([]transaction.FlatTransaction{first, other})
	require.NoError(t, err)

	other["Signers"].([]any)[0].(map[string]any)["Signer"].(map[string]any)["TxnSignature"] = "AB"
	_, _, err = CombineSponsorSigners([]transaction.FlatTransaction{first, other})
	require.ErrorIs(t, err, ErrSponsorTxNotEqual)
}

func TestCombineSponsorSignersListRules(t *testing.T) {
	sponsor := sponsorTestWallet(t, counterpartySeed)
	base, _ := sponsorAccountSignedTx(t, sponsor, false)
	fragments := make([]transaction.FlatTransaction, 33)
	entries := make([]any, len(fragments))
	for i := range fragments {
		id := make([]byte, 20)
		id[19] = byte(i + 1)
		address, err := addresscodec.EncodeAccountIDToClassicAddress(id)
		require.NoError(t, err)

		entries[i] = map[string]any{"Signer": map[string]any{"Account": address, "SigningPubKey": "ED5F5AC8B98974A3CA843326D9B88CEBD0560177B973EE0B149F782CFAA06DC66A", "TxnSignature": "CD"}}
		fragments[i] = transaction.FlatTransaction(clientinternal.CloneTransaction(base))
		fragments[i]["SponsorSignature"] = map[string]any{"Signers": []any{entries[i]}}
	}

	combined, _, err := CombineSponsorSigners(fragments[:32])
	require.NoError(t, err)
	require.Len(t, combined["SponsorSignature"].(map[string]any)["Signers"], 32)

	_, _, err = CombineSponsorSigners(fragments)
	require.Error(t, err)

	for _, invalid := range []any{nil, []any{}, []any{entries[1], entries[0]}, []any{entries[0], entries[0]}, []any{"bad"}, []any{map[string]any{"Signer": nil}}} {
		tx := transaction.FlatTransaction(clientinternal.CloneTransaction(base))
		tx["SponsorSignature"] = map[string]any{"Signers": invalid}
		_, _, err = CombineSponsorSigners([]transaction.FlatTransaction{tx})
		require.Error(t, err)
	}

	_, _, err = CombineSponsorSigners(nil)
	require.ErrorIs(t, err, ErrNoTransactionsToSign)

	_, _, err = CombineSponsorSigners([]transaction.FlatTransaction{base})
	require.ErrorIs(t, err, ErrTxMustIncludeSponsorSigners)

	_, _, err = CombineSponsorSigners([]transaction.FlatTransaction{nil})
	require.ErrorIs(t, err, ErrNilTransaction)

	single, _, _, err := SignAsSponsor(sponsor, base, nil)
	require.NoError(t, err)

	_, _, err = CombineSponsorSigners([]transaction.FlatTransaction{single})
	require.ErrorIs(t, err, ErrTxMustIncludeSponsorSigners)

	unsigned := transaction.FlatTransaction(clientinternal.CloneTransaction(fragments[0]))
	delete(unsigned, "TxnSignature")
	_, _, err = CombineSponsorSigners([]transaction.FlatTransaction{unsigned})
	require.ErrorIs(t, err, ErrAccountMustSignFirst)
}

// Even failures after signing must not publish partial maps or mutate account data.
func TestSignAsSponsorFailureIsAtomic(t *testing.T) {
	sponsor := sponsorTestWallet(t, counterpartySeed)
	base, _ := sponsorAccountSignedTx(t, sponsor, true)
	for _, failure := range []string{"private key", "public key", "payload", "final encoding"} {
		t.Run(failure, func(t *testing.T) {
			tx := transaction.FlatTransaction(clientinternal.CloneTransaction(base))
			w := sponsor
			switch failure {
			case "private key":
				w.PrivateKey = "bad"
			case "public key":
				w.PublicKey = ""
			case "payload":
				tx["Fee"] = "bad"
			case "final encoding":
				tx["Signers"].([]any)[0].(map[string]any)["Signer"].(map[string]any)["TxnSignature"] = "ZZ"
			}

			before := clientinternal.CloneTransaction(tx)
			got, blob, txHash, err := SignAsSponsor(w, tx, nil)
			require.Error(t, err)
			require.Nil(t, got)
			require.Empty(t, blob)
			require.Empty(t, txHash)
			require.Equal(t, before, map[string]any(tx))
		})
	}

	unsigned := transaction.FlatTransaction(clientinternal.CloneTransaction(base))
	delete(unsigned, "Signers")
	delete(unsigned, "SigningPubKey")
	_, _, _, err := SignAsSponsor(sponsor, unsigned, nil)
	require.ErrorIs(t, err, ErrAccountMustSignFirst)
}

func TestSponsorBlobErrors(t *testing.T) {
	sponsor := sponsorTestWallet(t, counterpartySeed)
	for _, blob := range []string{"not hex", "0", "FF"} {
		_, _, _, err := SignAsSponsorBlob(sponsor, blob, nil)
		require.Error(t, err)

		_, _, err = CombineSponsorSignersBlob([]string{blob})
		require.Error(t, err)
	}

	_, _, err := CombineSponsorSignersBlob(nil)
	require.ErrorIs(t, err, ErrNoTransactionsToSign)
}

func TestAddPreFundedSponsor(t *testing.T) {
	sponsor := sponsorTestWallet(t, counterpartySeed)
	account := sponsorTestWallet(t, brokerSeed)
	base := transaction.FlatTransaction{"TransactionType": "Payment", "Account": account.ClassicAddress.String(), "Memos": []map[string]any{{"Memo": map[string]any{"MemoData": "AB"}}}}
	for _, placeholder := range []bool{false, true} {
		tx := transaction.FlatTransaction(clientinternal.CloneTransaction(base))
		if placeholder {
			tx["SigningPubKey"] = ""
		}

		before := clientinternal.CloneTransaction(tx)
		got, err := AddPreFundedSponsor(tx, sponsor.ClassicAddress, types.SpfSponsorFee|types.SpfSponsorReserve)
		require.NoError(t, err)
		require.Equal(t, before, map[string]any(tx))
		require.Equal(t, sponsor.ClassicAddress.String(), got["Sponsor"])
		require.Equal(t, uint32(3), got["SponsorFlags"])
		require.NotContains(t, got, "SponsorSignature")

		got["Memos"].([]map[string]any)[0]["Memo"].(map[string]any)["MemoData"] = "CD"
		require.Equal(t, before, map[string]any(tx))
	}

	for _, field := range []string{"TxnSignature", "Signers", "SponsorSignature", "CounterpartySignature", "BatchSigners", "SigningPubKey"} {
		for _, value := range []any{nil, "AB"} {
			tx := transaction.FlatTransaction(clientinternal.CloneTransaction(base))
			tx[field] = value
			before := clientinternal.CloneTransaction(tx)
			got, err := AddPreFundedSponsor(tx, sponsor.ClassicAddress, types.SpfSponsorFee)
			require.ErrorIs(t, err, ErrTransactionAlreadySigned)
			require.Nil(t, got)
			require.Equal(t, before, map[string]any(tx))
		}
	}

	for _, flags := range []uint32{0, 4} {
		_, err := AddPreFundedSponsor(base, sponsor.ClassicAddress, flags)
		require.Error(t, err)
	}

	_, err := AddPreFundedSponsor(base, account.ClassicAddress, types.SpfSponsorFee)
	require.ErrorIs(t, err, transaction.ErrSponsorAccountConflict)

	_, err = AddPreFundedSponsor(nil, sponsor.ClassicAddress, types.SpfSponsorFee)
	require.ErrorIs(t, err, ErrNilTransaction)
}
