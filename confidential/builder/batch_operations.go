package builder

import (
	"fmt"
	"reflect"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// prepareBatchConvert builds a ConfidentialMPTConvert inner.
func prepareBatchConvert(state *batchState, op ConvertOp, options TxOptions) (BatchInnerTransaction, error) {
	params := op.BuildConvertParams
	if err := validateConvertBase(params); err != nil {
		return nil, err
	}
	issuance, token, err := state.resolve(op.Account, op.IssuanceID)
	if err != nil {
		return nil, err
	}
	if op.Amount > token.publicAmount {
		return nil, ErrInsufficientBalance
	}

	firstTime := token.holderKey == ""
	if !firstTime && !sameEncryptionKey(token.holderKey, op.HolderPubKey) {
		return nil, fmt.Errorf("%w: holder key", ErrKeyMismatch)
	}

	params.TxOptions = options
	tx, err := PrepareConvert(ConvertParams{
		BuildConvertParams: params,
		IssuerPubKey:       issuance.issuerKey,
		AuditorPubKey:      issuance.auditorKey,
		FirstTime:          firstTime,
	})
	if err != nil {
		return nil, err
	}

	if err := token.applyConvertCredit(tx.HolderEncryptedAmount, mirrorsOf(tx.IssuerEncryptedAmount, tx.AuditorEncryptedAmount), firstTime); err != nil {
		return nil, err
	}
	token.holderKey = op.HolderPubKey
	token.publicAmount -= op.Amount
	issuance.creditOutstanding(op.Amount)
	return tx, nil
}

// prepareBatchConvertBack builds a ConfidentialMPTConvertBack inner.
func prepareBatchConvertBack(state *batchState, op ConvertBackOp, options TxOptions) (BatchInnerTransaction, error) {
	params := op.BuildConvertBackParams
	if err := validateConvertBackBase(params); err != nil {
		return nil, err
	}
	if err := op.BalanceRange.Validate(); err != nil {
		return nil, err
	}
	issuance, token, err := state.resolve(op.Account, op.IssuanceID)
	if err != nil {
		return nil, err
	}
	if op.Amount > issuance.outstanding {
		return nil, ErrAmountExceedsOutstanding
	}

	spendingCt, err := requireSpendable(token, issuance, op.HolderPubKey)
	if err != nil {
		return nil, err
	}
	currentBalance, err := decryptPredictedBalance(issuance, spendingCt, op.HolderPrivKey, op.BalanceRange, "spending balance")
	if err != nil {
		return nil, err
	}

	params.TxOptions = options
	tx, err := PrepareConvertBack(ConvertBackParams{
		BuildConvertBackParams: params,
		IssuerPubKey:           issuance.issuerKey,
		AuditorPubKey:          issuance.auditorKey,
		BalanceVersion:         token.version,
		CurrentBalance:         currentBalance,
		CurrentBalanceCt:       spendingCt,
	})
	if err != nil {
		return nil, err
	}

	if err := token.applySpend(tx.HolderEncryptedAmount, mirrorsOf(tx.IssuerEncryptedAmount, tx.AuditorEncryptedAmount)); err != nil {
		return nil, err
	}
	token.creditPublic(op.Amount)
	issuance.debitOutstanding(op.Amount)
	return tx, nil
}

// prepareBatchSend builds a ConfidentialMPTSend inner.
func prepareBatchSend(state *batchState, op SendOp, options TxOptions) (BatchInnerTransaction, error) {
	params := op.BuildSendParams
	if err := validateSendBase(params); err != nil {
		return nil, err
	}
	if err := op.BalanceRange.Validate(); err != nil {
		return nil, err
	}
	issuance, sender, err := state.resolve(op.Account, op.IssuanceID)
	if err != nil {
		return nil, err
	}
	if !issuance.canTransfer() {
		return nil, ErrTransferDisabled
	}
	if issuance.transferFee > 0 {
		return nil, ErrTransferFeeSet
	}

	spendingCt, err := requireSpendable(sender, issuance, op.SenderPubKey)
	if err != nil {
		return nil, err
	}
	currentBalance, err := decryptPredictedBalance(issuance, spendingCt, op.SenderPrivKey, op.BalanceRange, "spending balance")
	if err != nil {
		return nil, err
	}

	destination, err := state.tokenFor(op.Destination, op.IssuanceID)
	if err != nil {
		return nil, err
	}
	destinationKey, err := requireReceivable(destination, issuance)
	if err != nil {
		return nil, err
	}

	params.TxOptions = options
	tx, err := PrepareSend(SendParams{
		BuildSendParams:  params,
		ReceiverPubKey:   destinationKey,
		IssuerPubKey:     issuance.issuerKey,
		AuditorPubKey:    issuance.auditorKey,
		BalanceVersion:   sender.version,
		CurrentBalance:   currentBalance,
		CurrentBalanceCt: spendingCt,
	})
	if err != nil {
		return nil, err
	}

	mirrors := mirrorsOf(tx.IssuerEncryptedAmount, tx.AuditorEncryptedAmount)
	if err := sender.applySpend(tx.SenderEncryptedAmount, mirrors); err != nil {
		return nil, err
	}
	challenge, err := sendChallenge(tx.ZKProof)
	if err != nil {
		return nil, err
	}
	if err := destination.applyInboxCredit(inboxCredit{
		challenge:     challenge,
		destinationCt: tx.DestinationEncryptedAmount,
		mirrors:       mirrors,
		issuerKey:     issuance.issuerKey,
		auditorKey:    issuance.auditorKey,
	}); err != nil {
		return nil, err
	}
	return tx, nil
}

// prepareBatchMergeInbox builds a ConfidentialMPTMergeInbox inner.
func prepareBatchMergeInbox(state *batchState, op MergeInboxOp, options TxOptions) (BatchInnerTransaction, error) {
	params := op.BuildMergeInboxParams
	if err := validateMergeInboxBase(params); err != nil {
		return nil, err
	}
	_, token, err := state.resolve(op.Account, op.IssuanceID)
	if err != nil {
		return nil, err
	}
	if _, err := token.requireHolderKey(); err != nil {
		return nil, fmt.Errorf("%w: HolderEncryptionKey is missing", ErrMissingSenderState)
	}
	if err := token.spending.requireExists("ConfidentialBalanceSpending", ErrMissingSenderState); err != nil {
		return nil, err
	}
	if err := token.inbox.requireExists("ConfidentialBalanceInbox", ErrMissingSenderState); err != nil {
		return nil, err
	}

	params.TxOptions = options
	tx, err := PrepareMergeInbox(MergeInboxParams{BuildMergeInboxParams: params})
	if err != nil {
		return nil, err
	}
	if err := token.applyMerge(); err != nil {
		return nil, err
	}
	return tx, nil
}

// prepareBatchClawback builds a ConfidentialMPTClawback inner.
func prepareBatchClawback(state *batchState, op ClawbackOp, options TxOptions) (BatchInnerTransaction, error) {
	params := op.BuildClawbackParams
	if err := validateClawbackBase(params); err != nil {
		return nil, err
	}
	if err := op.BalanceRange.Validate(); err != nil {
		return nil, err
	}
	issuance, err := state.issuance(op.IssuanceID)
	if err != nil {
		return nil, err
	}
	if !issuance.canClawback() {
		return nil, ErrClawbackDisabled
	}
	holder, err := state.tokenFor(op.Holder, op.IssuanceID)
	if err != nil {
		return nil, err
	}

	if _, err := holder.requireHolderKey(); err != nil {
		return nil, fmt.Errorf("%w: HolderEncryptionKey is missing", ErrMissingSenderState)
	}
	issuerCt, err := holder.issuerEnc.require("IssuerEncryptedBalance", ErrMissingSenderState)
	if err != nil {
		return nil, err
	}
	amount, err := decryptPredictedBalance(issuance, issuerCt, op.IssuerPrivKey, op.BalanceRange, "holder balance")
	if err != nil {
		return nil, err
	}

	params.TxOptions = options
	tx, err := PrepareClawback(ClawbackParams{
		BuildClawbackParams: params,
		Amount:              amount,
		IssuerPubKey:        issuance.issuerKey,
		IssuerCiphertext:    issuerCt,
	})
	if err != nil {
		return nil, err
	}

	holder.applyClawback()
	issuance.debitOutstanding(amount)
	return tx, nil
}

// resolve looks up the issuance and the submitter's own MPToken in one step, which is what every
// operation but the clawback needs.
func (s *batchState) resolve(account, issuanceID string) (*batchIssuance, *tokenState, error) {
	issuance, err := s.issuance(issuanceID)
	if err != nil {
		return nil, nil, err
	}
	token, err := s.tokenFor(account, issuanceID)
	if err != nil {
		return nil, nil, err
	}
	return issuance, token, nil
}

// tokenFor looks up one MPToken by the address and issuance an operation names.
func (s *batchState) tokenFor(holder, issuanceID string) (*tokenState, error) {
	key, err := newTokenKey(holder, issuanceID)
	if err != nil {
		return nil, err
	}
	return s.token(key)
}

// creditPublic returns public MPT to a holder, saturating at the protocol cap.
func (s *tokenState) creditPublic(amount uint64) {
	maximum := uint64(types.MaxMPTAmount)
	if s.publicAmount > maximum-amount {
		s.publicAmount = maximum
		return
	}
	s.publicAmount += amount
}

// requireSpendable checks the state ConfidentialMPTSend and ConfidentialMPTConvertBack both demand
// of the spender, in the order their transactors do, and returns the spending ciphertext the proof
// consumes.
func requireSpendable(token *tokenState, issuance *batchIssuance, pubKey string) (string, error) {
	if token.holderKey == "" {
		return "", fmt.Errorf("%w: HolderEncryptionKey is missing", ErrMissingSenderState)
	}
	if !sameEncryptionKey(token.holderKey, pubKey) {
		return "", fmt.Errorf("%w: holder key", ErrKeyMismatch)
	}
	spending, err := token.spending.require("ConfidentialBalanceSpending", ErrMissingSenderState)
	if err != nil {
		return "", err
	}
	if err := token.issuerEnc.requireExists("IssuerEncryptedBalance", ErrMissingSenderState); err != nil {
		return "", err
	}
	if issuance.hasAuditor() {
		if err := token.auditorEnc.requireExists("AuditorEncryptedBalance", ErrMissingSenderState); err != nil {
			return "", err
		}
	}
	return spending, nil
}

// requireReceivable checks the state a confidential send's destination must already have and
// returns the key the transferred amount is encrypted under.
func requireReceivable(token *tokenState, issuance *batchIssuance) (string, error) {
	holderKey, err := token.requireHolderKey()
	if err != nil {
		return "", err
	}
	for _, field := range []struct {
		name    string
		balance predictedBalance
	}{
		{"ConfidentialBalanceInbox", token.inbox},
		{"IssuerEncryptedBalance", token.issuerEnc},
	} {
		if err := field.balance.requireExists(field.name, ErrReceiverNotOptedIn); err != nil {
			return "", err
		}
	}
	if issuance.hasAuditor() {
		if err := token.auditorEnc.requireExists("AuditorEncryptedBalance", ErrReceiverNotOptedIn); err != nil {
			return "", err
		}
	}
	return holderKey, nil
}

// decryptPredictedBalance recovers the plaintext a proof needs from a predicted ciphertext, under
// the caller's bounds narrowed to the confidential supply as it stands at this point in the Batch.
func decryptPredictedBalance(
	issuance *batchIssuance,
	ciphertext, privKey string,
	bounds elgamal.AmountRange,
	what string,
) (uint64, error) {
	searchRange, err := issuance.searchRange(bounds)
	if err != nil {
		return 0, err
	}
	amount, err := elgamal.Decrypt(ciphertext, privKey, searchRange)
	if err != nil {
		return 0, fmt.Errorf("%w: failed to decrypt %s: %w", ErrCryptoFailed, what, err)
	}
	return amount, nil
}

// mirrorsOf pairs the issuer ciphertext a confidential transaction always carries with the auditor
// ciphertext it carries only under an auditing issuance.
func mirrorsOf(issuerCt string, auditorCt *string) mirrorCiphertexts {
	mirrors := mirrorCiphertexts{issuer: issuerCt}
	if auditorCt != nil {
		mirrors.auditor = *auditorCt
	}
	return mirrors
}

// supportedInnerTransactionTypes is the set of ordinary transaction types the assembler accepts as
// a ready-made inner.
var supportedInnerTransactionTypes = map[transaction.TxType]struct{}{
	transaction.AccountSetTx:       {},
	transaction.SetRegularKeyTx:    {},
	transaction.SignerListSetTx:    {},
	transaction.TicketCreateTx:     {},
	transaction.TrustSetTx:         {},
	transaction.DepositPreauthTx:   {},
	transaction.DelegateSetTx:      {},
	transaction.CredentialCreateTx: {},
	transaction.CredentialAcceptTx: {},
	transaction.CredentialDeleteTx: {},
}

// IsSupportedInnerTransactionType reports whether BuildBatch accepts a ready-made transaction of
// this type as a TransactionOp inner.
func IsSupportedInnerTransactionType(txType transaction.TxType) bool {
	_, supported := supportedInnerTransactionTypes[txType]
	return supported
}

// validatePlainInner rejects a ready-made inner the assembler cannot carry.
func validatePlainInner(op TransactionOp) error {
	if isNilTx(op.Tx) {
		return ErrBatchMissingOperation
	}
	txType := op.Tx.TxType()
	if !IsSupportedInnerTransactionType(txType) {
		return fmt.Errorf("%w: %s", ErrBatchInnerNotSupported, txType)
	}
	if err := validatePreparedTransaction(op.Tx); err != nil {
		return err
	}

	flat := op.Tx.Flatten()
	if _, ok := flat["Account"].(string); !ok {
		return ErrMissingAccount
	}
	for _, field := range []string{"LastLedgerSequence", "TxnSignature", "Signers"} {
		if _, present := flat[field]; present {
			return fmt.Errorf("%w: %s is not allowed on a Batch inner", ErrBatchInnerNotSupported, field)
		}
	}
	return nil
}

// plainInnerAccount reports the account a ready-made inner's nonce comes from.
func plainInnerAccount(op TransactionOp) (string, error) {
	if isNilTx(op.Tx) {
		return "", ErrBatchMissingOperation
	}
	account, ok := op.Tx.Flatten()["Account"].(string)
	if !ok {
		return "", ErrMissingAccount
	}
	return account, nil
}

// buildPlainInner shapes a ready-made transaction as a Batch inner, assigning it a position-
// derived sequence only when it carries neither nonce of its own.
func buildPlainInner(nonces *batchNonces, op TransactionOp) (transaction.FlatTransaction, error) {
	flat := op.Tx.Flatten()
	account, err := plainInnerAccount(op)
	if err != nil {
		return nil, err
	}
	ticket, _ := flat["TicketSequence"].(uint32)
	sequence, _ := flat["Sequence"].(uint32)
	switch {
	case ticket != 0:
	case sequence != 0:
		if err := nonces.claim(account, sequence); err != nil {
			return nil, err
		}
	default:
		sequence, err := nonces.allocate(account, TxOptions{})
		if err != nil {
			return nil, err
		}
		flat["Sequence"] = sequence
	}

	flags, _ := flat["Flags"].(uint32)
	flat["Flags"] = flags | types.TfInnerBatchTxn
	flat["Fee"] = "0"
	flat["SigningPubKey"] = ""
	return flat, nil
}

// isNilTx reports a nil interface or a typed nil pointer.
func isNilTx(tx BatchInnerTransaction) bool {
	if tx == nil {
		return true
	}
	v := reflect.ValueOf(tx)
	return v.Kind() == reflect.Pointer && v.IsNil()
}

// plainInnerHasNonce reports whether a flattened transaction already carries a nonce.
func plainInnerHasNonce(flat transaction.FlatTransaction) bool {
	if ticket, ok := flat["TicketSequence"].(uint32); ok && ticket != 0 {
		return true
	}
	sequence, ok := flat["Sequence"].(uint32)
	return ok && sequence != 0
}
