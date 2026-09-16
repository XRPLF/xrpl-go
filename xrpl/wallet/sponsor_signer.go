package wallet

import (
	"fmt"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl"
	"github.com/Peersyst/xrpl-go/xrpl/hash"
	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// SignAsSponsorOptions selects single signing or sponsor multisigning.
type SignAsSponsorOptions struct {
	// Multisign adds this wallet as one signer for the sponsor account.
	Multisign bool

	// MultisignAccount identifies the signer account when using a regular key.
	// A nonempty value enables multisigning and overrides wallet.ClassicAddress.
	MultisignAccount string
}

// SignAsSponsor adds sponsor authorization to an account-signed transaction.
// It returns a new transaction, its blob, and its final hash, without changing tx.
// The Sponsor and fixCleanup3_4_0 amendments must be enabled on the target network.
// No legacy signing fallback or network authorization check is performed.
//
// Set Sponsor, SponsorFlags, and the fee before account signing. Account signatures
// are checked structurally, not cryptographically. Each sponsor multisigner must
// sign the same account-signed transaction independently, then combine fragments.
func SignAsSponsor(w Wallet, tx transaction.FlatTransaction, opts *SignAsSponsorOptions) (transaction.FlatTransaction, string, string, error) {
	if tx == nil {
		return nil, "", "", ErrNilTransaction
	}

	if _, present := tx["SponsorSignature"]; present {
		return nil, "", "", ErrSponsorAlreadySigned
	}

	if _, err := inspectSponsorTransaction(tx); err != nil {
		return nil, "", "", err
	}

	multisign := opts != nil && (opts.Multisign || opts.MultisignAccount != "")
	signerAddress := w.ClassicAddress.String()
	if opts != nil && opts.MultisignAccount != "" {
		signerAddress = opts.MultisignAccount
	}

	signer, err := sponsorSignerAddress(signerAddress)
	if err != nil {
		return nil, "", "", err
	}

	if !multisign {
		sponsor, _ := typecheck.ToString(tx["Sponsor"]) // Validated by inspectSponsorTransaction.
		sponsorAddress, err := addresscodec.DecodeAddress(sponsor)
		if err != nil {
			return nil, "", "", err
		}
		if sponsorAddress.AccountID != signer.AccountID {
			return nil, "", "", ErrSponsorWalletMismatch
		}
	}

	working := transaction.FlatTransaction(clientinternal.CloneTransaction(tx))
	var payload string
	if multisign {
		payload, err = binarycodec.EncodeForMultisigningSponsor(working, signer.Classic)
	} else {
		payload, err = binarycodec.EncodeForSigningSponsor(working)
	}

	if err != nil {
		return nil, "", "", err
	}

	signature, err := w.ComputeSignature(payload)
	if err != nil {
		return nil, "", "", err
	}

	var sponsorSignature types.SponsorSignature
	if multisign {
		sponsorSignature.Signers = []types.Signer{{SignerData: types.SignerData{
			Account: types.Address(signer.Classic), SigningPubKey: w.PublicKey, TxnSignature: signature,
		}}}
	} else {
		sponsorSignature.SigningPubKey = &w.PublicKey
		sponsorSignature.TxnSignature = &signature
	}

	working["SponsorSignature"] = flattenSponsorSignature(sponsorSignature)
	if _, err := transaction.InspectSponsorFields(working); err != nil {
		return nil, "", "", err
	}

	blob, err := binarycodec.Encode(working)
	if err != nil {
		return nil, "", "", err
	}

	txHash, err := hash.SignTxBlob(blob)
	if err != nil {
		return nil, "", "", err
	}

	return working, blob, txHash, nil
}

// SignAsSponsorBlob decodes an account-signed blob and signs it as the sponsor.
func SignAsSponsorBlob(w Wallet, blob string, opts *SignAsSponsorOptions) (transaction.FlatTransaction, string, string, error) {
	tx, err := binarycodec.Decode(blob)
	if err != nil {
		return nil, "", "", err
	}

	return SignAsSponsor(w, transaction.FlatTransaction(tx), opts)
}

// CombineSponsorSigners combines sponsor multisign fragments without changing them.
// Each fragment must have a valid structural account signing form and a sorted,
// unique sponsor signer list. Transactions must be identical on the wire except
// for SponsorSignature.Signers. Nonserialized metadata is not compared.
//
// Overlapping signer AccountIDs retain the first input occurrence, even if a later
// signature differs. An invalid first signature can mask a later valid one. This
// function does not verify signatures, ledger authorization, or signer quorum.
// Use hash.SignTxBlob on the returned blob to obtain the final transaction hash.
func CombineSponsorSigners(transactions []transaction.FlatTransaction) (transaction.FlatTransaction, string, error) {
	if len(transactions) == 0 {
		return nil, "", ErrNoTransactionsToSign
	}

	var combined transaction.FlatTransaction
	var combinedSignature types.SponsorSignature
	var reference string
	var allSigners []types.Signer
	seen := make(map[[addresscodec.AccountAddressLength]byte]bool)

	for _, tx := range transactions {
		signature, err := inspectSponsorTransaction(tx)
		if err != nil {
			return nil, "", err
		}

		if signature == nil || len(signature.Signers) == 0 {
			return nil, "", ErrTxMustIncludeSponsorSigners
		}

		signers := signature.Signers
		signature.Signers = nil

		working := transaction.FlatTransaction(clientinternal.CloneTransaction(tx))
		working["SponsorSignature"] = signature.Flatten()
		comparison, err := binarycodec.Encode(working)
		if err != nil {
			return nil, "", err
		}

		if combined == nil {
			combined, combinedSignature, reference = working, *signature, comparison
		} else if comparison != reference {
			return nil, "", ErrSponsorTxNotEqual
		}

		for _, signer := range signers {
			address, err := sponsorSignerAddress(signer.SignerData.Account.String())
			if err != nil {
				return nil, "", err
			}

			// Deduplicate before the unstable sort to preserve first-input wins.
			if !seen[address.AccountID] {
				seen[address.AccountID] = true
				signer.SignerData.Account = types.Address(address.Classic)
				allSigners = append(allSigners, signer)
			}
		}
	}

	if err := xrpl.SortByAccountID(allSigners, func(signer types.Signer) (string, error) {
		return signer.SignerData.Account.String(), nil
	}); err != nil {
		return nil, "", err
	}

	combinedSignature.Signers = allSigners
	combined["SponsorSignature"] = flattenSponsorSignature(combinedSignature)
	if _, err := transaction.InspectSponsorFields(combined); err != nil {
		return nil, "", err
	}

	blob, err := binarycodec.Encode(combined)
	if err != nil {
		return nil, "", err
	}

	return combined, blob, nil
}

// CombineSponsorSignersBlob decodes blobs before combining their sponsor signers.
func CombineSponsorSignersBlob(blobs []string) (transaction.FlatTransaction, string, error) {
	transactions := make([]transaction.FlatTransaction, len(blobs))
	for i, blob := range blobs {
		tx, err := binarycodec.Decode(blob)
		if err != nil {
			return nil, "", err
		}

		transactions[i] = transaction.FlatTransaction(tx)
	}

	return CombineSponsorSigners(transactions)
}

// AddPreFundedSponsor returns an unsigned copy with Sponsor and SponsorFlags set.
// An absent or empty SigningPubKey is allowed, but authorization fields must be
// absent. Call this before account signing and autofill. It does not fund or check
// a ledger Sponsorship object. The caller needs an existing sponsorship with
// sufficient resources and no require-sign flag for the requested sponsorship.
func AddPreFundedSponsor(tx transaction.FlatTransaction, sponsor types.Address, flags uint32) (transaction.FlatTransaction, error) {
	if tx == nil {
		return nil, ErrNilTransaction
	}

	for _, field := range []string{"TxnSignature", "Signers", "SponsorSignature", "CounterpartySignature", "BatchSigners"} {
		if _, present := tx[field]; present {
			return nil, ErrTransactionAlreadySigned
		}
	}

	if value, present := tx["SigningPubKey"]; present {
		key, ok := value.(string)
		if !ok || key != "" {
			return nil, ErrTransactionAlreadySigned
		}
	}

	working := transaction.FlatTransaction(clientinternal.CloneTransaction(tx))
	working["Sponsor"], working["SponsorFlags"] = sponsor.String(), flags
	if _, err := transaction.InspectSponsorFields(working); err != nil {
		return nil, err
	}

	return working, nil
}

// flattenSponsorSignature uses the model serializers while preserving the []any
// signer array returned by wallet helpers (the model's Flatten uses []map).
func flattenSponsorSignature(signature types.SponsorSignature) map[string]any {
	signers := signature.Signers
	signature.Signers = nil
	flat := signature.Flatten()
	if len(signers) > 0 {
		entries := make([]any, len(signers))
		for i := range signers {
			entries[i] = signers[i].Flatten()
		}
		flat["Signers"] = entries
	}

	return flat
}

func inspectSponsorTransaction(tx transaction.FlatTransaction) (*types.SponsorSignature, error) {
	if tx == nil {
		return nil, ErrNilTransaction
	}

	form, err := clientinternal.InspectSignedTransaction(tx, false)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrAccountMustSignFirst, err)
	}

	if form != clientinternal.SingleSignedTransaction && form != clientinternal.MultiSignedTransaction {
		return nil, ErrAccountMustSignFirst
	}

	if _, present := tx["Sponsor"]; !present {
		return nil, transaction.ErrSponsorFieldsMissing
	}

	return transaction.InspectSponsorFields(tx)
}

func sponsorSignerAddress(address string) (addresscodec.DecodedAddress, error) {
	decoded, err := addresscodec.DecodeAddress(address)
	if err != nil {
		return addresscodec.DecodedAddress{}, fmt.Errorf("%w: %w", xrpl.ErrInvalidSigner, err)
	}

	if decoded.HasTag {
		return addresscodec.DecodedAddress{}, ErrAddressHasTag
	}

	if addresscodec.IsZeroAccountID(decoded.AccountID[:]) {
		return addresscodec.DecodedAddress{}, transaction.ErrSignerAccountZero
	}

	return decoded, nil
}
