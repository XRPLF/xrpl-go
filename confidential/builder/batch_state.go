package builder

import (
	"fmt"
	"strings"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/pkg/mptsizes"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// tokenKey identifies one MPToken, which XLS-33 keys by holder and issuance.
type tokenKey struct {
	holder   string
	issuance string
}

// newTokenKey decodes an address and normalizes an issuance ID into a state map key.
func newTokenKey(holder, issuanceID string) (tokenKey, error) {
	decoded, err := decodeBuilderAddress(holder)
	if err != nil {
		return tokenKey{}, fmt.Errorf("%w: %w", ErrInvalidAddress, err)
	}
	return tokenKey{holder: decoded.Classic, issuance: strings.ToUpper(issuanceID)}, nil
}

// String renders the key for an error message.
func (k tokenKey) String() string {
	return fmt.Sprintf("%s on issuance %s", k.holder, k.issuance)
}

// predictedBalance is one confidential ciphertext field of an MPToken as the assembler believes it
// will stand when the inner being built applies.
type predictedBalance struct {
	value         string
	unpredictable bool
}

// knownBalance is a ciphertext read from the ledger or predicted from one.
func knownBalance(ciphertext string) predictedBalance {
	return predictedBalance{value: ciphertext}
}

// unpredictableBalance marks a field that exists but whose value this package cannot reproduce,
// because the transactor derived it from the canonical encrypted zero.
func unpredictableBalance() predictedBalance {
	return predictedBalance{unpredictable: true}
}

// exists reports whether the MPToken carries the field at all, which is all a transactor checks
// where it does not read the value.
func (b predictedBalance) exists() bool {
	return b.value != "" || b.unpredictable
}

// requireExists rejects a field the MPToken does not carry, for an inner that needs the field
// present but never reads it.
func (b predictedBalance) requireExists(field string, absent error) error {
	if !b.exists() {
		return fmt.Errorf("%w: %s is missing", absent, field)
	}
	return nil
}

// require reads a ciphertext an inner cannot be built without, naming which of the two unusable
// states it is in.
func (b predictedBalance) require(field string, absent error) (string, error) {
	if b.unpredictable {
		return "", fmt.Errorf("%w: %s", ErrBatchUnpredictableState, field)
	}
	if b.value == "" {
		return "", fmt.Errorf("%w: %s is missing", absent, field)
	}
	return b.value, nil
}

// credit adds an incoming ciphertext.
func (b predictedBalance) credit(field, amountCt string) (predictedBalance, error) {
	if b.unpredictable {
		return unpredictableBalance(), nil
	}
	if b.value == "" {
		return knownBalance(amountCt), nil
	}
	sum, err := elgamal.Add(b.value, amountCt)
	if err != nil {
		return predictedBalance{}, fmt.Errorf("%w: %s: %w", ErrCryptoFailed, field, err)
	}
	return knownBalance(sum), nil
}

// debit subtracts an outgoing ciphertext.
func (b predictedBalance) debit(field, amountCt string, absent error) (predictedBalance, error) {
	if err := b.requireExists(field, absent); err != nil {
		return predictedBalance{}, err
	}
	if b.unpredictable {
		return unpredictableBalance(), nil
	}
	difference, err := elgamal.Subtract(b.value, amountCt)
	if err != nil {
		return predictedBalance{}, fmt.Errorf("%w: %s: %w", ErrCryptoFailed, field, err)
	}
	return knownBalance(difference), nil
}

// tokenState is the confidential state of one MPToken as the next inner will find it.
type tokenState struct {
	holderKey    string
	spending     predictedBalance
	inbox        predictedBalance
	issuerEnc    predictedBalance
	auditorEnc   predictedBalance
	version      uint32
	publicAmount uint64
}

// bump advances the balance version the way the transactor does, wrapping at 32 bits.
func (s tokenState) bump() tokenState {
	s.version++
	return s
}

// batchState is the whole predicted state of one assembly.
type batchState struct {
	tokens    map[tokenKey]*tokenState
	issuances map[string]*batchIssuance
}

// batchIssuance holds one issuance's shared state.
type batchIssuance struct {
	issuanceState
	outstanding uint64
}

// token returns the predicted state of one MPToken.
func (s *batchState) token(key tokenKey) (*tokenState, error) {
	state, ok := s.tokens[key]
	if !ok {
		return nil, fmt.Errorf("%w: no state loaded for %s", ErrInvalidLedgerState, key)
	}
	return state, nil
}

// issuance returns one issuance's shared state, by the same contract as token.
func (s *batchState) issuance(issuanceID string) (*batchIssuance, error) {
	state, ok := s.issuances[strings.ToUpper(issuanceID)]
	if !ok {
		return nil, fmt.Errorf("%w: no state loaded for issuance %s", ErrInvalidLedgerState, issuanceID)
	}
	return state, nil
}

// creditOutstanding raises the confidential supply by a converted amount, saturating at the
// protocol cap.
func (i *batchIssuance) creditOutstanding(amount uint64) {
	maximum := uint64(types.MaxMPTAmount)
	if i.outstanding > maximum-amount {
		i.outstanding = maximum
		return
	}
	i.outstanding += amount
}

// debitOutstanding lowers the confidential supply by an amount leaving it.
func (i *batchIssuance) debitOutstanding(amount uint64) {
	if amount > i.outstanding {
		i.outstanding = 0
		return
	}
	i.outstanding -= amount
}

// searchRange narrows a caller's decryption bounds to the confidential supply as the assembler
// predicts it at this point in the Batch.
func (i *batchIssuance) searchRange(bounds elgamal.AmountRange) (elgamal.AmountRange, error) {
	return boundedBalanceRange(bounds, i.outstanding)
}

// mirrorCiphertexts pairs an operation's issuer and auditor ciphertexts for a transition that
// applies the same change to both mirror balances.
type mirrorCiphertexts struct {
	issuer  string
	auditor string
}

// applySpend applies what ConfidentialMPTSend and ConfidentialMPTConvertBack do to the spender's
// own MPToken.
func (s *tokenState) applySpend(spendCt string, mirrors mirrorCiphertexts) error {
	spending, err := s.spending.debit("ConfidentialBalanceSpending", spendCt, ErrMissingSenderState)
	if err != nil {
		return err
	}
	issuerEnc, err := s.issuerEnc.debit("IssuerEncryptedBalance", mirrors.issuer, ErrMissingSenderState)
	if err != nil {
		return err
	}
	auditorEnc := s.auditorEnc
	if mirrors.auditor != "" {
		if auditorEnc, err = s.auditorEnc.debit("AuditorEncryptedBalance", mirrors.auditor, ErrMissingSenderState); err != nil {
			return err
		}
	}

	s.spending, s.issuerEnc, s.auditorEnc = spending, issuerEnc, auditorEnc
	*s = s.bump()
	return nil
}

// applyConvertCredit applies what ConfidentialMPTConvert does to the converting holder.
func (s *tokenState) applyConvertCredit(holderCt string, mirrors mirrorCiphertexts, firstTime bool) error {
	inbox, err := s.inbox.credit("ConfidentialBalanceInbox", holderCt)
	if err != nil {
		return err
	}
	issuerEnc, err := s.issuerEnc.credit("IssuerEncryptedBalance", mirrors.issuer)
	if err != nil {
		return err
	}
	auditorEnc := s.auditorEnc
	if mirrors.auditor != "" {
		if auditorEnc, err = s.auditorEnc.credit("AuditorEncryptedBalance", mirrors.auditor); err != nil {
			return err
		}
	}

	s.inbox, s.issuerEnc, s.auditorEnc = inbox, issuerEnc, auditorEnc
	if firstTime && !s.spending.exists() {
		s.spending = unpredictableBalance()
	}
	return nil
}

// applyInboxCredit applies what ConfidentialMPTSend does to the destination.
func (s *tokenState) applyInboxCredit(credit inboxCredit) error {
	destinationKey, err := s.requireHolderKey()
	if err != nil {
		return err
	}
	inboxCt, err := rerandomize(credit.destinationCt, destinationKey, credit.challenge)
	if err != nil {
		return err
	}
	inbox, err := s.inbox.credit("ConfidentialBalanceInbox", inboxCt)
	if err != nil {
		return err
	}

	issuerCt, err := rerandomize(credit.mirrors.issuer, credit.issuerKey, credit.challenge)
	if err != nil {
		return err
	}
	issuerEnc, err := s.issuerEnc.credit("IssuerEncryptedBalance", issuerCt)
	if err != nil {
		return err
	}

	auditorEnc := s.auditorEnc
	if credit.mirrors.auditor != "" {
		auditorCt, err := rerandomize(credit.mirrors.auditor, credit.auditorKey, credit.challenge)
		if err != nil {
			return err
		}
		if auditorEnc, err = s.auditorEnc.credit("AuditorEncryptedBalance", auditorCt); err != nil {
			return err
		}
	}

	s.inbox, s.issuerEnc, s.auditorEnc = inbox, issuerEnc, auditorEnc
	return nil
}

// applyMerge applies what ConfidentialMPTMergeInbox does.
func (s *tokenState) applyMerge() error {
	spending := unpredictableBalance()
	if s.spending.value != "" && s.inbox.value != "" {
		merged, err := elgamal.Add(s.spending.value, s.inbox.value)
		if err != nil {
			return fmt.Errorf("%w: ConfidentialBalanceSpending: %w", ErrCryptoFailed, err)
		}
		spending = knownBalance(merged)
	}

	s.spending, s.inbox = spending, unpredictableBalance()
	*s = s.bump()
	return nil
}

// applyClawback applies what ConfidentialMPTClawback does.
func (s *tokenState) applyClawback() {
	s.spending = unpredictableBalance()
	s.inbox = unpredictableBalance()
	s.issuerEnc = unpredictableBalance()
	s.auditorEnc = unpredictableBalance()
	*s = s.bump()
}

// requireHolderKey reads the holder's registered encryption key, which a send's destination must
// have before anything can be encrypted to it.
func (s *tokenState) requireHolderKey() (string, error) {
	if s.holderKey == "" {
		return "", fmt.Errorf("%w: HolderEncryptionKey is missing", ErrReceiverNotOptedIn)
	}
	return s.holderKey, nil
}

// inboxCredit carries everything applyInboxCredit needs from a built send.
type inboxCredit struct {
	challenge     string
	destinationCt string
	mirrors       mirrorCiphertexts
	issuerKey     string
	auditorKey    string
}

// rerandomize reproduces the transactor's re-blinding of a credited ciphertext.
func rerandomize(ciphertext, pubKey, challenge string) (string, error) {
	zero, err := elgamal.Encrypt(0, pubKey, challenge)
	if err != nil {
		return "", fmt.Errorf("%w: re-randomizing a credited ciphertext: %w", ErrCryptoFailed, err)
	}
	blinded, err := elgamal.Add(ciphertext, zero)
	if err != nil {
		return "", fmt.Errorf("%w: re-randomizing a credited ciphertext: %w", ErrCryptoFailed, err)
	}
	return blinded, nil
}

// sendChallenge extracts the re-randomization scalar from a send proof, which is its leading
// blinding-factor-sized half.
func sendChallenge(zkProof string) (string, error) {
	const challengeHexLength = 2 * mptsizes.BlindingFactorSize
	if len(zkProof) < challengeHexLength {
		return "", fmt.Errorf("%w: send proof is shorter than its challenge", ErrCryptoFailed)
	}
	return zkProof[:challengeHexLength], nil
}

// readBatchTokenState reads one MPToken's confidential state from the validated snapshot.
func readBatchTokenState(snapshot *ledgerSnapshot, issuance issuanceState, key tokenKey, holder string, usable bool) (*tokenState, error) {
	resp, err := getMPTokenEntry(snapshot, key.issuance, holder)
	if err != nil {
		return nil, err
	}
	if usable {
		if err := requireHolderUsable(resp.Node, issuance); err != nil {
			return nil, err
		}
	}

	state := &tokenState{}
	if state.holderKey, err = optionalString(resp.Node, "HolderEncryptionKey"); err != nil {
		return nil, err
	}
	for _, field := range []struct {
		name   string
		target *predictedBalance
	}{
		{"ConfidentialBalanceSpending", &state.spending},
		{"ConfidentialBalanceInbox", &state.inbox},
		{"IssuerEncryptedBalance", &state.issuerEnc},
		{"AuditorEncryptedBalance", &state.auditorEnc},
	} {
		ciphertext, err := optionalString(resp.Node, field.name)
		if err != nil {
			return nil, err
		}
		*field.target = knownBalance(ciphertext)
	}
	if state.version, err = optionalUint32(resp.Node, "ConfidentialBalanceVersion"); err != nil {
		return nil, err
	}
	if state.publicAmount, err = optionalUint64(resp.Node, "MPTAmount"); err != nil {
		return nil, err
	}
	if err := requireCurrentBalanceVersion(snapshot, resp.Index, state.version); err != nil {
		return nil, err
	}
	return state, nil
}
