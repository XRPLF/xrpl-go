package builder

import (
	"fmt"
	"maps"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// prepareBatchConvert builds a ConfidentialMPTConvert inner.
func prepareBatchConvert(state *batchState, op ConvertOp, nonce TxOptions) (BatchInnerTransaction, error) {
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

	params := op.BuildConvertParams
	params.TxOptions = nonce
	tx, err := PrepareConvert(ConvertParams{
		BuildConvertParams: params,
		IssuerPubKey:       issuance.issuerKey,
		AuditorPubKey:      issuance.auditorKey,
		FirstTime:          firstTime,
	})
	if err != nil {
		return nil, err
	}

	if firstTime {
		token.holderKey = op.HolderPubKey
	}
	if err := token.applyConvert(tx.HolderEncryptedAmount, mirrorsOf(tx.IssuerEncryptedAmount, tx.AuditorEncryptedAmount)); err != nil {
		return nil, err
	}
	token.publicAmount -= op.Amount
	issuance.creditOutstanding(op.Amount)
	return tx, nil
}

// prepareBatchConvertBack builds a ConfidentialMPTConvertBack inner.
func prepareBatchConvertBack(state *batchState, op ConvertBackOp, nonce TxOptions) (BatchInnerTransaction, error) {
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

	params := op.BuildConvertBackParams
	params.TxOptions = nonce
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
func prepareBatchSend(state *batchState, op SendOp, nonce TxOptions) (BatchInnerTransaction, error) {
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

	params := op.BuildSendParams
	params.TxOptions = nonce
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
func prepareBatchMergeInbox(state *batchState, op MergeInboxOp, nonce TxOptions) (BatchInnerTransaction, error) {
	_, token, err := state.resolve(op.Account, op.IssuanceID)
	if err != nil {
		return nil, err
	}
	for _, field := range []struct {
		name  string
		value string
	}{
		{"HolderEncryptionKey", token.holderKey},
		{"ConfidentialBalanceSpending", token.spending},
		{"ConfidentialBalanceInbox", token.inbox},
	} {
		if err := requireField(field.value, field.name, ErrMissingSenderState); err != nil {
			return nil, err
		}
	}

	params := op.BuildMergeInboxParams
	params.TxOptions = nonce
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
func prepareBatchClawback(state *batchState, op ClawbackOp, nonce TxOptions) (BatchInnerTransaction, error) {
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

	if err := requireField(holder.holderKey, "HolderEncryptionKey", ErrMissingSenderState); err != nil {
		return nil, err
	}
	if err := requireField(holder.issuerEnc, "IssuerEncryptedBalance", ErrMissingSenderState); err != nil {
		return nil, err
	}
	amount, err := decryptPredictedBalance(issuance, holder.issuerEnc, op.IssuerPrivKey, op.BalanceRange, "holder balance")
	if err != nil {
		return nil, err
	}

	params := op.BuildClawbackParams
	params.TxOptions = nonce
	tx, err := PrepareClawback(ClawbackParams{
		BuildClawbackParams: params,
		Amount:              amount,
		IssuerPubKey:        issuance.issuerKey,
		IssuerCiphertext:    holder.issuerEnc,
	})
	if err != nil {
		return nil, err
	}

	if err := holder.applyClawback(issuance); err != nil {
		return nil, err
	}
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

// requireField rejects a confidential field an inner needs that the MPToken does not carry.
func requireField(value, field string, absent error) error {
	if value == "" {
		return fmt.Errorf("%w: %s is missing", absent, field)
	}
	return nil
}

// requireSpendable checks the state ConfidentialMPTSend and ConfidentialMPTConvertBack both demand
// of the spender, in the order their transactors do, and returns the spending ciphertext the proof
// consumes.
func requireSpendable(token *tokenState, issuance *batchIssuance, pubKey string) (string, error) {
	if err := requireField(token.holderKey, "HolderEncryptionKey", ErrMissingSenderState); err != nil {
		return "", err
	}
	if !sameEncryptionKey(token.holderKey, pubKey) {
		return "", fmt.Errorf("%w: holder key", ErrKeyMismatch)
	}
	if err := requireField(token.spending, "ConfidentialBalanceSpending", ErrMissingSenderState); err != nil {
		return "", err
	}
	if err := requireField(token.issuerEnc, "IssuerEncryptedBalance", ErrMissingSenderState); err != nil {
		return "", err
	}
	if issuance.hasAuditor() {
		if err := requireField(token.auditorEnc, "AuditorEncryptedBalance", ErrMissingSenderState); err != nil {
			return "", err
		}
	}
	return token.spending, nil
}

// requireReceivable checks the state a confidential send's destination must already have and
// returns the key the transferred amount is encrypted under.
func requireReceivable(token *tokenState, issuance *batchIssuance) (string, error) {
	for _, field := range []struct {
		name  string
		value string
	}{
		{"HolderEncryptionKey", token.holderKey},
		{"ConfidentialBalanceInbox", token.inbox},
		{"IssuerEncryptedBalance", token.issuerEnc},
	} {
		if err := requireField(field.value, field.name, ErrReceiverNotOptedIn); err != nil {
			return "", err
		}
	}
	if issuance.hasAuditor() {
		if err := requireField(token.auditorEnc, "AuditorEncryptedBalance", ErrReceiverNotOptedIn); err != nil {
			return "", err
		}
	}
	return token.holderKey, nil
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

// batchStep validates a ready-made inner and extracts its account and nonce once, applying the
// same TxOptions contract the confidential operations follow. A caller-set Sequence is kept and
// later checked against the sequence the inner's position requires.
func (op TransactionOp) batchStep() (batchStep, error) {
	if isNilValue(op.Tx) {
		return batchStep{}, ErrBatchMissingOperation
	}
	txType := op.Tx.TxType()
	if !IsSupportedInnerTransactionType(txType) {
		return batchStep{}, fmt.Errorf("%w: %s", ErrBatchInnerNotSupported, txType)
	}
	if err := validatePreparedTransaction(op.Tx); err != nil {
		return batchStep{}, err
	}

	// Flatten may return a map the transaction keeps, so the payload is copied once here and
	// every later read and build works from that copy.
	flat := maps.Clone(op.Tx.Flatten())
	account, ok := flat["Account"].(string)
	if !ok {
		return batchStep{}, ErrMissingAccount
	}
	for _, field := range []string{"LastLedgerSequence", "TxnSignature", "Signers"} {
		if _, present := flat[field]; present {
			return batchStep{}, fmt.Errorf("%w: %s is not allowed on a Batch inner", ErrBatchInnerNotSupported, field)
		}
	}
	// A Flatten built from decoded JSON carries numbers as float64, so every UInt32 the
	// assembler reads is normalized to uint32 rather than silently read as absent.
	for _, field := range []string{"Sequence", "TicketSequence", "TicketCount", "Flags"} {
		value, present := flat[field]
		if !present {
			continue
		}
		normalized, ok := typecheck.ToUint32(value)
		if !ok {
			return batchStep{}, fmt.Errorf("%w: %s is not a UInt32: %v", ErrInvalidTransaction, field, value)
		}
		flat[field] = normalized
	}

	var options TxOptions
	options.Sequence, _ = flat["Sequence"].(uint32)
	options.TicketSequence, _ = flat["TicketSequence"].(uint32)
	options.Delegate, _ = flat["Delegate"].(string)
	if err := options.validate(account, txType); err != nil {
		return batchStep{}, err
	}

	step := batchStep{
		account: account,
		options: options,
		build: func(_ *batchState, nonce TxOptions) (transaction.FlatTransaction, error) {
			inner := maps.Clone(flat)
			if nonce.Sequence != 0 {
				inner["Sequence"] = nonce.Sequence
			}
			return shapeBatchInner(inner), nil
		},
	}
	if txType == transaction.TicketCreateTx {
		step.ticketCount, _ = flat["TicketCount"].(uint32)
	}
	return step, nil
}
