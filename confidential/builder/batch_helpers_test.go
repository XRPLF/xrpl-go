package builder

import (
	"encoding/hex"
	"errors"
	"strconv"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	"github.com/Peersyst/xrpl-go/pkg/mptsizes"
	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	ledgerentries "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

const (
	secondIssuanceID = "000005C463C52827307480341E3CB23A0710CC839EB58A0A"
	batchSearchHigh  = 10_000
	batchOutstanding = 1_000_000
)

// batchRange is the decryption bound every batch fixture test passes.
func batchRange() elgamal.AmountRange {
	return elgamal.AmountRange{Low: 0, High: batchSearchHigh}
}

// issuanceFixture is one issuance in a batch fixture, together with the keys its mirror balances
// are encrypted under.
type issuanceFixture struct {
	id      string
	issuer  elgamal.Keypair
	auditor *elgamal.Keypair
}

// holderFixture describes one MPToken to place on the fixture ledger.
type holderFixture struct {
	key          elgamal.Keypair
	unregistered bool
	spending     *uint64
	inbox        *uint64
	version      uint32
	publicAmount uint64
	flags        uint32
	omit         bool
}

// batchFixture builds a mock ledger for the batch assembler.
type batchFixture struct {
	t         *testing.T
	entries   map[string]ledgerentries.FlatLedgerObject
	issuances map[string]issuanceFixture
	keys      map[string]elgamal.Keypair
	sequences map[string]uint32
}

func newBatchFixture(t *testing.T) *batchFixture {
	t.Helper()
	return &batchFixture{
		t:         t,
		entries:   make(map[string]ledgerentries.FlatLedgerObject),
		issuances: make(map[string]issuanceFixture),
		keys:      make(map[string]elgamal.Keypair),
		sequences: make(map[string]uint32),
	}
}

// withIssuance registers an issuance with a fresh issuer key and no auditor.
func (f *batchFixture) withIssuance(issuanceID string) *batchFixture {
	return f.withIssuanceFlags(issuanceID, confidentialIssuanceFlags, false)
}

// withAuditedIssuance registers an issuance that also carries an auditor key, which adds the
// auditor mirror balance to every holder of it.
func (f *batchFixture) withAuditedIssuance(issuanceID string) *batchFixture {
	return f.withIssuanceFlags(issuanceID, confidentialIssuanceFlags, true)
}

func (f *batchFixture) withIssuanceFlags(issuanceID string, flags uint32, audited bool) *batchFixture {
	f.t.Helper()

	fixture := issuanceFixture{id: issuanceID, issuer: f.generateKey()}
	if audited {
		auditor := f.generateKey()
		fixture.auditor = &auditor
	}
	f.issuances[issuanceID] = fixture

	entry := ledgerentries.FlatLedgerObject{
		"LedgerEntryType":               string(ledgerentries.MPTokenIssuanceEntry),
		"Flags":                         float64(flags),
		"ConfidentialOutstandingAmount": strconv.FormatUint(batchOutstanding, 10),
		"IssuerEncryptionKey":           fixture.issuer.PubKeyHex,
	}
	if fixture.auditor != nil {
		entry["AuditorEncryptionKey"] = fixture.auditor.PubKeyHex
	}

	index, err := xrplhash.MPTokenIssuance(issuanceID)
	require.NoError(f.t, err)
	f.entries[index] = entry
	return f
}

// withHolder places one holder's MPToken on the fixture ledger and remembers its keypair.
func (f *batchFixture) withHolder(holder, issuanceID string, spec holderFixture) *batchFixture {
	f.t.Helper()

	issuance, ok := f.issuances[issuanceID]
	require.True(f.t, ok, "issuance %s must be registered before a holder", issuanceID)

	key := spec.key
	if key.PubKeyHex == "" {
		key = f.generateKey()
	}
	f.keys[holderKeyID(holder, issuanceID)] = key
	if spec.omit {
		return f
	}

	entry := ledgerentries.FlatLedgerObject{"LedgerEntryType": string(ledgerentries.MPTokenEntry)}
	if spec.flags != 0 {
		entry["Flags"] = float64(spec.flags)
	}
	if !spec.unregistered {
		entry["HolderEncryptionKey"] = key.PubKeyHex
	}

	var mirror uint64
	confidential := false
	if spec.spending != nil {
		entry["ConfidentialBalanceSpending"] = f.encrypt(*spec.spending, key.PubKeyHex)
		entry["ConfidentialBalanceVersion"] = float64(spec.version)
		mirror += *spec.spending
		confidential = true
	}
	if spec.inbox != nil {
		entry["ConfidentialBalanceInbox"] = f.encrypt(*spec.inbox, key.PubKeyHex)
		mirror += *spec.inbox
		confidential = true
	}
	if confidential {
		entry["IssuerEncryptedBalance"] = f.encrypt(mirror, issuance.issuer.PubKeyHex)
		if issuance.auditor != nil {
			entry["AuditorEncryptedBalance"] = f.encrypt(mirror, issuance.auditor.PubKeyHex)
		}
	}
	if spec.publicAmount > 0 {
		entry["MPTAmount"] = strconv.FormatUint(spec.publicAmount, 10)
	}

	index, err := xrplhash.MPToken(issuanceID, holder)
	require.NoError(f.t, err)
	f.entries[index] = entry
	return f
}

// withSequence sets the sequence account_info reports for an account.
func (f *batchFixture) withSequence(account string, sequence uint32) *batchFixture {
	f.sequences[account] = sequence
	return f
}

// querier returns a LedgerQuerier over the fixture ledger.
func (f *batchFixture) querier() *batchQuerier {
	return &batchQuerier{fixture: f}
}

// holderKey returns the keypair the fixture generated for one holder of one issuance.
func (f *batchFixture) holderKey(holder, issuanceID string) elgamal.Keypair {
	f.t.Helper()

	key, ok := f.keys[holderKeyID(holder, issuanceID)]
	require.True(f.t, ok, "no key for %s on %s", holder, issuanceID)
	return key
}

// issuerKey returns the issuance's issuer keypair.
func (f *batchFixture) issuerKey(issuanceID string) elgamal.Keypair {
	f.t.Helper()

	issuance, ok := f.issuances[issuanceID]
	require.True(f.t, ok, "no issuance %s", issuanceID)
	return issuance.issuer
}

func (f *batchFixture) generateKey() elgamal.Keypair {
	f.t.Helper()

	key, err := elgamal.GenerateKeypair()
	require.NoError(f.t, err)
	return key
}

// encrypt produces a real ciphertext of a known amount, so the assembler's homomorphic predictions
// can be decrypted back to plaintext the test computed independently.
func (f *batchFixture) encrypt(amount uint64, pubKey string) string {
	f.t.Helper()

	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(f.t, err)
	ciphertext, err := elgamal.Encrypt(amount, pubKey, bf)
	require.NoError(f.t, err)
	return ciphertext
}

func holderKeyID(holder, issuanceID string) string {
	return holder + ":" + issuanceID
}

func amountOf(value uint64) *uint64 {
	return &value
}

// batchQuerier answers the assembler's reads from a batchFixture.
type batchQuerier struct {
	fixture    *batchFixture
	entryErrs  map[string]error
	accountErr error
	requests   []ledger.EntryRequest
	accounts   []account.InfoRequest
}

func (q *batchQuerier) GetAccountInfo(req *account.InfoRequest) (*account.InfoResponse, error) {
	q.accounts = append(q.accounts, *req)
	if q.accountErr != nil {
		return nil, q.accountErr
	}
	sequence, ok := q.fixture.sequences[req.Account.String()]
	if !ok {
		sequence = 1
	}
	return &account.InfoResponse{
		AccountData: ledgerentries.AccountRoot{Sequence: sequence},
		LedgerIndex: mockLedgerIndex,
		Validated:   true,
	}, nil
}

func (q *batchQuerier) GetLedgerEntry(req *ledger.EntryRequest) (*ledger.EntryResponse, error) {
	q.requests = append(q.requests, *req)
	if err := q.entryErrs[req.Index]; err != nil {
		return nil, err
	}
	node, ok := q.fixture.entries[req.Index]
	if !ok {
		return nil, errors.New(ledgerEntryNotFound)
	}
	if req.LedgerIndex == common.Current {
		return &ledger.EntryResponse{
			Index:              req.Index,
			LedgerCurrentIndex: mockOpenLedgerIndex,
			Node:               node,
		}, nil
	}
	return &ledger.EntryResponse{
		Index:       req.Index,
		LedgerHash:  mockLedgerHash,
		LedgerIndex: mockLedgerIndex,
		Node:        node,
		Validated:   true,
	}, nil
}

// innerOf reads one built inner out of an assembled Batch.
func innerOf(t *testing.T, batch *transaction.Batch, index int) transaction.FlatTransaction {
	t.Helper()

	require.Greater(t, len(batch.RawTransactions), index)
	return batch.RawTransactions[index].RawTransaction
}

// requireInnerShape asserts the XLS-56 shape every inner must carry.
func requireInnerShape(t *testing.T, inner transaction.FlatTransaction) {
	t.Helper()

	flags, ok := inner["Flags"].(uint32)
	require.True(t, ok, "Flags must be a uint32")
	require.NotZero(t, flags&types.TfInnerBatchTxn, "inner must carry tfInnerBatchTxn")
	require.Equal(t, "0", inner["Fee"])
	signingPubKey, ok := inner["SigningPubKey"].(string)
	require.True(t, ok, "SigningPubKey must be present on an inner")
	require.Empty(t, signingPubKey)
	require.NotContains(t, inner, "TxnSignature")
	require.NotContains(t, inner, "Signers")
	require.NotContains(t, inner, "LastLedgerSequence")
}

// decryptField decrypts one ciphertext of a built inner or a prediction.
func decryptField(t *testing.T, ciphertext, privKey string) uint64 {
	t.Helper()

	amount, err := elgamal.Decrypt(ciphertext, privKey, batchRange())
	require.NoError(t, err)
	return amount
}

// sendKeys are the encryption keys a send proof was built against.
type sendKeys struct {
	sender   string
	receiver string
	issuer   string
	auditor  string
}

// requireSendBinding verifies a built send's proof against the balance ciphertext and the version
// the assembler predicted for it.
func requireSendBinding(t *testing.T, tx *transaction.ConfidentialMPTSend, keys sendKeys, balanceCt string, sequence, version uint32) {
	t.Helper()

	ctxHash, err := proof.SendContextHash(string(tx.Account), tx.MPTokenIssuanceID, sequence, string(tx.Destination), version)
	require.NoError(t, err)

	participants := []mptcrypto.Participant{
		decodeParticipant(t, keys.sender, tx.SenderEncryptedAmount),
		decodeParticipant(t, keys.receiver, tx.DestinationEncryptedAmount),
		decodeParticipant(t, keys.issuer, tx.IssuerEncryptedAmount),
	}
	if keys.auditor != "" {
		require.NotNil(t, tx.AuditorEncryptedAmount)
		participants = append(participants, decodeParticipant(t, keys.auditor, *tx.AuditorEncryptedAmount))
	}

	require.NoError(t, mptcrypto.VerifySendProof(
		decodeBytes(t, tx.ZKProof),
		participants,
		decodeCiphertext(t, balanceCt),
		decodeCommitment(t, tx.AmountCommitment),
		decodeCommitment(t, tx.BalanceCommitment),
		decodeContextHash(t, ctxHash),
	))
}

// requireConvertBackBinding verifies a built convert-back's proof against the predicted balance
// ciphertext and version, on the same reasoning as requireSendBinding.
func requireConvertBackBinding(t *testing.T, tx *transaction.ConfidentialMPTConvertBack, holderPubKey, balanceCt string, sequence, version uint32) {
	t.Helper()

	ctxHash, err := proof.ConvertBackContextHash(string(tx.Account), tx.MPTokenIssuanceID, sequence, version)
	require.NoError(t, err)

	var fixed [mptsizes.ConvertBackProofSize]byte
	copy(fixed[:], decodeBytes(t, tx.ZKProof))
	amount, err := strconv.ParseUint(tx.MPTAmount.String(), 10, 64)
	require.NoError(t, err)
	require.NoError(t, mptcrypto.VerifyConvertBackProof(
		fixed,
		decodePubKey(t, holderPubKey),
		decodeCiphertext(t, balanceCt),
		decodeCommitment(t, tx.BalanceCommitment),
		amount,
		decodeContextHash(t, ctxHash),
	))
}

// requireClawbackBinding verifies a built clawback's proof against the predicted issuer mirror
// balance and the amount the assembler decrypted from it.
func requireClawbackBinding(t *testing.T, tx *transaction.ConfidentialMPTClawback, issuerPubKey, issuerCt string, sequence uint32) {
	t.Helper()

	ctxHash, err := proof.ClawbackContextHash(string(tx.Account), tx.MPTokenIssuanceID, sequence, string(tx.Holder))
	require.NoError(t, err)

	var fixed [mptsizes.CompactClawbackProofSize]byte
	copy(fixed[:], decodeBytes(t, tx.ZKProof))
	amount, err := strconv.ParseUint(tx.MPTAmount.String(), 10, 64)
	require.NoError(t, err)
	require.NoError(t, mptcrypto.VerifyClawbackProof(
		fixed,
		amount,
		decodePubKey(t, issuerPubKey),
		decodeCiphertext(t, issuerCt),
		decodeContextHash(t, ctxHash),
	))
}

func decodeBytes(t *testing.T, value string) []byte {
	t.Helper()

	decoded, err := hex.DecodeString(value)
	require.NoError(t, err)
	return decoded
}

func decodeCiphertext(t *testing.T, value string) mptcrypto.Ciphertext {
	t.Helper()

	return mptcrypto.Ciphertext(decodeFixed(t, value, mptsizes.CiphertextSize))
}

func decodeCommitment(t *testing.T, value string) mptcrypto.Commitment {
	t.Helper()

	return mptcrypto.Commitment(decodeFixed(t, value, mptsizes.CommitmentSize))
}

func decodePubKey(t *testing.T, value string) mptcrypto.PublicKey {
	t.Helper()

	return mptcrypto.PublicKey(decodeFixed(t, value, mptsizes.PubKeySize))
}

func decodeContextHash(t *testing.T, value string) mptcrypto.ContextHash {
	t.Helper()

	return mptcrypto.ContextHash(decodeFixed(t, value, mptsizes.HashOutputSize))
}

func decodeFixed(t *testing.T, value string, size int) []byte {
	t.Helper()

	decoded := decodeBytes(t, value)
	require.Len(t, decoded, size)
	return decoded
}

func decodeParticipant(t *testing.T, pubKey, ciphertext string) mptcrypto.Participant {
	t.Helper()

	return mptcrypto.Participant{PubKey: decodePubKey(t, pubKey), Ciphertext: decodeCiphertext(t, ciphertext)}
}
