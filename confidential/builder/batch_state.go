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
	return tokenKey{holder: decoded.Classic, issuance: normalizeIssuanceID(issuanceID)}, nil
}

// normalizeIssuanceID gives an issuance ID the one spelling the state maps are keyed by.
func normalizeIssuanceID(issuanceID string) string {
	return strings.ToUpper(issuanceID)
}

// String renders the key for an error message.
func (k tokenKey) String() string {
	return fmt.Sprintf("%s on issuance %s", k.holder, k.issuance)
}

// tokenState is the confidential state of one MPToken as the next inner will find it. An empty
// ciphertext is a field the MPToken does not carry.
type tokenState struct {
	key          tokenKey
	holderKey    string
	spending     string
	inbox        string
	issuerEnc    string
	auditorEnc   string
	version      uint32
	publicAmount uint64
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
	state, ok := s.issuances[normalizeIssuanceID(issuanceID)]
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

// addCiphertext credits an amount to a balance, where an absent balance takes the amount as is.
func addCiphertext(field, balance, amountCt string) (string, error) {
	if balance == "" {
		return amountCt, nil
	}
	sum, err := elgamal.Add(balance, amountCt)
	if err != nil {
		return "", fmt.Errorf("%w: %s: %w", ErrCryptoFailed, field, err)
	}
	return sum, nil
}

// subtractCiphertext debits an amount from a balance the MPToken must carry.
func subtractCiphertext(field, balance, amountCt string) (string, error) {
	if err := requireField(balance, field, ErrMissingSenderState); err != nil {
		return "", err
	}
	difference, err := elgamal.Subtract(balance, amountCt)
	if err != nil {
		return "", fmt.Errorf("%w: %s: %w", ErrCryptoFailed, field, err)
	}
	return difference, nil
}

// canonicalZero reproduces the encrypted zero the transactor writes to one of this MPToken's
// balances under the given key.
func (s *tokenState) canonicalZero(field, pubKey string) (string, error) {
	zero, err := elgamal.EncryptCanonicalZero(pubKey, s.key.holder, s.key.issuance)
	if err != nil {
		return "", fmt.Errorf("%w: %s: %w", ErrCryptoFailed, field, err)
	}
	return zero, nil
}

// applySpend applies what ConfidentialMPTSend and ConfidentialMPTConvertBack do to the spender's
// own MPToken.
func (s *tokenState) applySpend(spendCt string, mirrors mirrorCiphertexts) error {
	spending, err := subtractCiphertext("ConfidentialBalanceSpending", s.spending, spendCt)
	if err != nil {
		return err
	}
	issuerEnc, err := subtractCiphertext("IssuerEncryptedBalance", s.issuerEnc, mirrors.issuer)
	if err != nil {
		return err
	}
	auditorEnc := s.auditorEnc
	if mirrors.auditor != "" {
		if auditorEnc, err = subtractCiphertext("AuditorEncryptedBalance", s.auditorEnc, mirrors.auditor); err != nil {
			return err
		}
	}

	s.spending, s.issuerEnc, s.auditorEnc = spending, issuerEnc, auditorEnc
	s.version++
	return nil
}

// applyConvert applies what ConfidentialMPTConvert does to the converting holder, whose holder key
// must already be set. A holder carrying no confidential balance yet has every balance initialized:
// the inbox and mirrors to the converted amount, and the spending balance to the canonical
// encrypted zero. Otherwise the amount is credited to the inbox and mirrors.
func (s *tokenState) applyConvert(holderCt string, mirrors mirrorCiphertexts) error {
	if s.spending == "" && s.inbox == "" && s.issuerEnc == "" && s.auditorEnc == "" {
		spending, err := s.canonicalZero("ConfidentialBalanceSpending", s.holderKey)
		if err != nil {
			return err
		}
		s.spending, s.inbox, s.issuerEnc, s.auditorEnc = spending, holderCt, mirrors.issuer, mirrors.auditor
		s.version = 0
		return nil
	}

	inbox, err := addCiphertext("ConfidentialBalanceInbox", s.inbox, holderCt)
	if err != nil {
		return err
	}
	issuerEnc, err := addCiphertext("IssuerEncryptedBalance", s.issuerEnc, mirrors.issuer)
	if err != nil {
		return err
	}
	auditorEnc := s.auditorEnc
	if mirrors.auditor != "" {
		if auditorEnc, err = addCiphertext("AuditorEncryptedBalance", s.auditorEnc, mirrors.auditor); err != nil {
			return err
		}
	}

	s.inbox, s.issuerEnc, s.auditorEnc = inbox, issuerEnc, auditorEnc
	return nil
}

// applyInboxCredit applies what ConfidentialMPTSend does to the destination.
func (s *tokenState) applyInboxCredit(credit inboxCredit) error {
	if err := requireField(s.holderKey, "HolderEncryptionKey", ErrReceiverNotOptedIn); err != nil {
		return err
	}
	inboxCt, err := rerandomize(credit.destinationCt, s.holderKey, credit.challenge)
	if err != nil {
		return err
	}
	inbox, err := addCiphertext("ConfidentialBalanceInbox", s.inbox, inboxCt)
	if err != nil {
		return err
	}

	issuerCt, err := rerandomize(credit.mirrors.issuer, credit.issuerKey, credit.challenge)
	if err != nil {
		return err
	}
	issuerEnc, err := addCiphertext("IssuerEncryptedBalance", s.issuerEnc, issuerCt)
	if err != nil {
		return err
	}

	auditorEnc := s.auditorEnc
	if credit.mirrors.auditor != "" {
		auditorCt, err := rerandomize(credit.mirrors.auditor, credit.auditorKey, credit.challenge)
		if err != nil {
			return err
		}
		if auditorEnc, err = addCiphertext("AuditorEncryptedBalance", s.auditorEnc, auditorCt); err != nil {
			return err
		}
	}

	s.inbox, s.issuerEnc, s.auditorEnc = inbox, issuerEnc, auditorEnc
	return nil
}

// applyMerge applies what ConfidentialMPTMergeInbox does: the inbox is added to the spending
// balance and reset to the canonical encrypted zero.
func (s *tokenState) applyMerge() error {
	spending, err := addCiphertext("ConfidentialBalanceSpending", s.spending, s.inbox)
	if err != nil {
		return err
	}
	inbox, err := s.canonicalZero("ConfidentialBalanceInbox", s.holderKey)
	if err != nil {
		return err
	}

	s.spending, s.inbox = spending, inbox
	s.version++
	return nil
}

// applyClawback applies what ConfidentialMPTClawback does: the holder's own balances and its
// mirrors are reset to the canonical encrypted zero under their respective keys.
func (s *tokenState) applyClawback(issuance *batchIssuance) error {
	holderZero, err := s.canonicalZero("ConfidentialBalanceSpending", s.holderKey)
	if err != nil {
		return err
	}
	issuerZero, err := s.canonicalZero("IssuerEncryptedBalance", issuance.issuerKey)
	if err != nil {
		return err
	}
	auditorEnc := s.auditorEnc
	if auditorEnc != "" {
		if auditorEnc, err = s.canonicalZero("AuditorEncryptedBalance", issuance.auditorKey); err != nil {
			return err
		}
	}

	s.spending, s.inbox, s.issuerEnc, s.auditorEnc = holderZero, holderZero, issuerZero, auditorEnc
	s.version++
	return nil
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

// readBatchTokenState reads one MPToken's confidential state from the validated snapshot, under
// the combined access of every operation in the Batch that touches it.
func readBatchTokenState(snapshot *ledgerSnapshot, issuance issuanceState, key tokenKey, holder string, access mptokenAccess) (*tokenState, error) {
	resp, err := readMPToken(snapshot, issuance, key.issuance, holder, access)
	if err != nil {
		return nil, err
	}

	state := &tokenState{key: key}
	for _, field := range []struct {
		name   string
		target *string
	}{
		{"HolderEncryptionKey", &state.holderKey},
		{"ConfidentialBalanceSpending", &state.spending},
		{"ConfidentialBalanceInbox", &state.inbox},
		{"IssuerEncryptedBalance", &state.issuerEnc},
		{"AuditorEncryptedBalance", &state.auditorEnc},
	} {
		if *field.target, err = optionalString(resp.Node, field.name); err != nil {
			return nil, err
		}
	}
	if state.version, err = optionalUint32(resp.Node, "ConfidentialBalanceVersion"); err != nil {
		return nil, err
	}
	if state.publicAmount, err = optionalUint64(resp.Node, "MPTAmount"); err != nil {
		return nil, err
	}
	if err := access.requireFreshVersion(snapshot, resp.Index, state.version); err != nil {
		return nil, err
	}
	return state, nil
}
