package builder

import (
	"encoding/hex"
	"errors"
	"maps"
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
	// openVersions reports a different ConfidentialBalanceVersion on the open ledger for the
	// MPToken at each index, as a confidential transaction still in flight would.
	openVersions map[string]uint32
	requests     []ledger.EntryRequest
	accounts     []account.InfoRequest
}

// queries counts every ledger request the querier answered.
func (q *batchQuerier) queries() int {
	return len(q.requests) + len(q.accounts)
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
		if version, changed := q.openVersions[req.Index]; changed {
			open := make(ledgerentries.FlatLedgerObject, len(node))
			maps.Copy(open, node)
			open["ConfidentialBalanceVersion"] = float64(version)
			node = open
		}
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

// sendOp builds a SendOp with the fixture's key for the sender and the standard bounds.
func sendOp(f *batchFixture, from, to, issuanceID string, amount uint64) SendOp {
	key := f.holderKey(from, issuanceID)
	return SendOp{BuildSendParams{
		Account:       from,
		Destination:   to,
		IssuanceID:    issuanceID,
		Amount:        amount,
		SenderPrivKey: key.PrivKeyHex,
		SenderPubKey:  key.PubKeyHex,
		BalanceRange:  batchRange(),
	}}
}

// convertOp builds a ConvertOp with the fixture's key for the holder.
func convertOp(f *batchFixture, holder, issuanceID string, amount uint64) ConvertOp {
	key := f.holderKey(holder, issuanceID)
	return ConvertOp{BuildConvertParams{
		Account:       holder,
		IssuanceID:    issuanceID,
		Amount:        amount,
		HolderPrivKey: key.PrivKeyHex,
		HolderPubKey:  key.PubKeyHex,
	}}
}

// convertBackOp builds a ConvertBackOp with the fixture's key for the holder.
func convertBackOp(f *batchFixture, holder, issuanceID string, amount uint64) ConvertBackOp {
	key := f.holderKey(holder, issuanceID)
	return ConvertBackOp{BuildConvertBackParams{
		Account:       holder,
		IssuanceID:    issuanceID,
		Amount:        amount,
		HolderPrivKey: key.PrivKeyHex,
		HolderPubKey:  key.PubKeyHex,
		BalanceRange:  batchRange(),
	}}
}

// mergeOp builds a MergeInboxOp for one holder.
func mergeOp(holder, issuanceID string) MergeInboxOp {
	return MergeInboxOp{BuildMergeInboxParams{Account: holder, IssuanceID: issuanceID}}
}

// clawbackOp builds a ClawbackOp submitted by the issuance's issuer.
func clawbackOp(f *batchFixture, issuer, holder, issuanceID string) ClawbackOp {
	return ClawbackOp{BuildClawbackParams{
		Account:       issuer,
		Holder:        holder,
		IssuanceID:    issuanceID,
		IssuerPrivKey: f.issuerKey(issuanceID).PrivKeyHex,
		BalanceRange:  batchRange(),
	}}
}

// plainAccountSet builds a ready-made AccountSet inner, carrying a sequence only when one is given.
func plainAccountSet(account string, sequence uint32) TransactionOp {
	return TransactionOp{Tx: &transaction.AccountSet{
		BaseTx: transaction.BaseTx{
			Account:         types.Address(account),
			TransactionType: transaction.AccountSetTx,
			Sequence:        sequence,
		},
	}}
}

// plainTicketCreate builds a ready-made TicketCreate inner funded by the given nonce.
func plainTicketCreate(account string, count uint32, nonce TxOptions) TransactionOp {
	return TransactionOp{Tx: &transaction.TicketCreate{
		BaseTx: transaction.BaseTx{
			Account:         types.Address(account),
			TransactionType: transaction.TicketCreateTx,
			Sequence:        nonce.Sequence,
			TicketSequence:  nonce.TicketSequence,
		},
		TicketCount: count,
	}}
}

// flatInner is a custom ready-made inner whose Flatten returns the same map on every call, as a
// transaction decoded from JSON or caching its payload would.
type flatInner struct {
	flat transaction.FlatTransaction
}

func (f flatInner) TxType() transaction.TxType {
	return transaction.TxType(f.flat["TransactionType"].(string))
}

func (f flatInner) Flatten() transaction.FlatTransaction { return f.flat }

func (f flatInner) Validate() (bool, error) { return true, nil }

// firstSend decodes one built inner back into the typed transaction the assertions read.
func firstSend(t *testing.T, batch *transaction.Batch, index int) *transaction.ConfidentialMPTSend {
	t.Helper()

	inner := innerOf(t, batch, index)
	require.Equal(t, transaction.ConfidentialMPTSendTx.String(), inner["TransactionType"])

	tx := &transaction.ConfidentialMPTSend{
		BaseTx: transaction.BaseTx{
			Account:         types.Address(inner["Account"].(string)),
			TransactionType: transaction.ConfidentialMPTSendTx,
		},
		MPTokenIssuanceID:          inner["MPTokenIssuanceID"].(string),
		Destination:                types.Address(inner["Destination"].(string)),
		SenderEncryptedAmount:      inner["SenderEncryptedAmount"].(string),
		DestinationEncryptedAmount: inner["DestinationEncryptedAmount"].(string),
		IssuerEncryptedAmount:      inner["IssuerEncryptedAmount"].(string),
		ZKProof:                    inner["ZKProof"].(string),
		AmountCommitment:           inner["AmountCommitment"].(string),
		BalanceCommitment:          inner["BalanceCommitment"].(string),
	}
	if auditor, ok := inner["AuditorEncryptedAmount"].(string); ok {
		tx.AuditorEncryptedAmount = &auditor
	}
	return tx
}

// convertBackInner decodes one built ConfidentialMPTConvertBack inner.
func convertBackInner(t *testing.T, batch *transaction.Batch, index int) *transaction.ConfidentialMPTConvertBack {
	t.Helper()

	inner := innerOf(t, batch, index)
	require.Equal(t, transaction.ConfidentialMPTConvertBackTx.String(), inner["TransactionType"])

	amount, err := strconv.ParseUint(inner["MPTAmount"].(string), 10, 64)
	require.NoError(t, err)
	return &transaction.ConfidentialMPTConvertBack{
		BaseTx: transaction.BaseTx{
			Account:         types.Address(inner["Account"].(string)),
			TransactionType: transaction.ConfidentialMPTConvertBackTx,
		},
		MPTokenIssuanceID: inner["MPTokenIssuanceID"].(string),
		MPTAmount:         types.MPTPlainAmount(amount),
		BalanceCommitment: inner["BalanceCommitment"].(string),
		ZKProof:           inner["ZKProof"].(string),
	}
}

// clawbackInner decodes one built ConfidentialMPTClawback inner.
func clawbackInner(t *testing.T, batch *transaction.Batch, index int) *transaction.ConfidentialMPTClawback {
	t.Helper()

	inner := innerOf(t, batch, index)
	require.Equal(t, transaction.ConfidentialMPTClawbackTx.String(), inner["TransactionType"])

	amount, err := strconv.ParseUint(inner["MPTAmount"].(string), 10, 64)
	require.NoError(t, err)
	return &transaction.ConfidentialMPTClawback{
		BaseTx: transaction.BaseTx{
			Account:         types.Address(inner["Account"].(string)),
			TransactionType: transaction.ConfidentialMPTClawbackTx,
		},
		MPTokenIssuanceID: inner["MPTokenIssuanceID"].(string),
		Holder:            types.Address(inner["Holder"].(string)),
		MPTAmount:         types.MPTPlainAmount(amount),
		ZKProof:           inner["ZKProof"].(string),
	}
}

// fixtureField reads one confidential field the fixture ledger carries for one holder, which is
// the state the first inner on that MPToken finds.
func fixtureField(t *testing.T, fixture *batchFixture, holder, issuanceID, field string) string {
	t.Helper()

	index, err := xrplhash.MPToken(issuanceID, holder)
	require.NoError(t, err)
	value, ok := fixture.entries[index][field].(string)
	require.True(t, ok, "fixture holder has no %s", field)
	return value
}

// spendingCiphertext reads the spending balance the fixture ledger carries for one holder.
func spendingCiphertext(t *testing.T, fixture *batchFixture, holder, issuanceID string) string {
	t.Helper()

	return fixtureField(t, fixture, holder, issuanceID, "ConfidentialBalanceSpending")
}

// issuerMirrorOf reads the issuer mirror balance the fixture ledger carries for one holder.
func issuerMirrorOf(t *testing.T, fixture *batchFixture, holder, issuanceID string) string {
	t.Helper()

	return fixtureField(t, fixture, holder, issuanceID, "IssuerEncryptedBalance")
}

// canonicalZeroOf computes the canonical encrypted zero independently of the assembler.
func canonicalZeroOf(t *testing.T, pubKey, holder, issuanceID string) string {
	t.Helper()

	zero, err := elgamal.EncryptCanonicalZero(pubKey, holder, issuanceID)
	require.NoError(t, err)
	return zero
}

// addCiphertexts adds ciphertexts in order, the way a test recomputes a transactor's credits.
func addCiphertexts(t *testing.T, first string, rest ...string) string {
	t.Helper()

	sum := first
	for _, next := range rest {
		var err error
		sum, err = elgamal.Add(sum, next)
		require.NoError(t, err)
	}
	return sum
}
