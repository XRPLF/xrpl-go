//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package confidential

import (
	"strconv"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/builder"
	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/stretchr/testify/require"
)

// issuanceMaximumAmount is large enough that no scenario here approaches the cap,
// so a rejection always names the condition under test.
const issuanceMaximumAmount = types.MPTAmount(1_000)

// confidentialClient is everything a confidential scenario needs from a client: the
// submission surface the runner drives, and the two ledger reads the builders make.
// The RPC and WebSocket clients both satisfy it, which is what lets one scenario body
// run against either transport.
type confidentialClient interface {
	integration.Client
	builder.LedgerQuerier
}

// issuanceConfig describes the issuance a scenario needs. The auditor is optional
// because XLS-96 leaves the auditor key unset on an issuance that has no auditor, and
// several assertions here distinguish that case from an auditor that decrypts to zero.
type issuanceConfig struct {
	issuerKey   elgamal.Keypair
	auditorKey  *elgamal.Keypair
	canClawback bool
	canLock     bool
	// permissionless drops LsfMPTRequireAuth, so holders need no MPTokenAuthorize round trip
	// from the issuer. The builder's holder preflight branches on that flag, and every other
	// scenario here takes only the authorized branch.
	permissionless bool
	// postCreationEnable takes the second route XLS-96 section 6.3.2 defines: the issuance is
	// created without the confidential capability, and a later MPTokenIssuanceSet turns it on
	// with a different flag while registering the keys in the same transaction. Section
	// 12.4.2.3 accepts those keys only because that flag is being enabled alongside them.
	postCreationEnable bool
}

func generateKey(t *testing.T) elgamal.Keypair {
	t.Helper()

	key, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	return key
}

// createIssuance creates a confidential-capable issuance and registers its encryption
// keys, returning the issuance ID. The keys go on a separate MPTokenIssuanceSet because
// the issuance ID they are registered against only exists once the create is validated.
func createIssuance(
	t *testing.T,
	runner *integration.Runner,
	client confidentialClient,
	issuer *wallet.Wallet,
	config issuanceConfig,
) string {
	t.Helper()

	maximumAmount := issuanceMaximumAmount
	create := transaction.MPTokenIssuanceCreate{
		BaseTx:        transaction.BaseTx{Account: issuer.GetAddress()},
		MaximumAmount: &maximumAmount,
	}
	if !config.permissionless {
		create.SetMPTRequireAuthFlag()
	}
	create.SetMPTCanTransferFlag()
	if !config.postCreationEnable {
		create.SetMPTCanHoldConfidentialBalanceFlag()
	}
	if config.canClawback {
		create.SetMPTCanClawbackFlag()
	}
	if config.canLock {
		create.SetMPTCanLockFlag()
	}
	submitAndWait(t, runner, create.Flatten(), issuer)

	issuanceID := getIssuanceID(t, client, issuer.GetAddress())

	setKeys := transaction.MPTokenIssuanceSet{
		BaseTx:              transaction.BaseTx{Account: issuer.GetAddress()},
		MPTokenIssuanceID:   issuanceID,
		IssuerEncryptionKey: &config.issuerKey.PubKeyHex,
	}
	if config.auditorKey != nil {
		setKeys.AuditorEncryptionKey = &config.auditorKey.PubKeyHex
	}
	if config.postCreationEnable {
		setKeys.SetMPTCanHoldConfidentialBalanceFlag()
	}
	submitAndWait(t, runner, setKeys.Flatten(), issuer)

	return issuanceID
}

// optInHolder creates the holder's MPToken. Every holder needs one before it can receive,
// whether or not the issuance requires authorization, so this is the half of authorization
// that a permissionless issuance still demands.
func optInHolder(t *testing.T, runner *integration.Runner, holder *wallet.Wallet, issuanceID string) {
	t.Helper()

	holderAuthorize := transaction.MPTokenAuthorize{
		BaseTx:            transaction.BaseTx{Account: holder.GetAddress()},
		MPTokenIssuanceID: issuanceID,
	}
	submitAndWait(t, runner, holderAuthorize.Flatten(), holder)
}

// authorizeHolder completes the two-sided authorization LsfMPTRequireAuth demands: the
// holder opts in, then the issuer approves that holder.
func authorizeHolder(t *testing.T, runner *integration.Runner, issuer, holder *wallet.Wallet, issuanceID string) {
	t.Helper()

	optInHolder(t, runner, holder, issuanceID)

	holderAddress := holder.GetAddress()
	issuerAuthorize := transaction.MPTokenAuthorize{
		BaseTx:            transaction.BaseTx{Account: issuer.GetAddress()},
		MPTokenIssuanceID: issuanceID,
		Holder:            &holderAddress,
	}
	submitAndWait(t, runner, issuerAuthorize.Flatten(), issuer)
}

// fundHolder pays a public MPT balance to the holder, which is what a later convert
// moves into the confidential balance.
func fundHolder(
	t *testing.T,
	runner *integration.Runner,
	issuer, holder *wallet.Wallet,
	issuanceID string,
	amount uint64,
) {
	t.Helper()

	payment := transaction.Payment{
		BaseTx:      transaction.BaseTx{Account: issuer.GetAddress()},
		Destination: holder.GetAddress(),
		Amount: types.MPTCurrencyAmount{
			MPTIssuanceID: issuanceID,
			Value:         types.MPTAmount(amount).String(),
		},
	}
	submitAndWait(t, runner, payment.Flatten(), issuer)
}

// submitAndWait autofills, signs, and submits a transaction that is expected to succeed,
// and returns once it is validated. It fails hard, so a transaction the network would
// only queue never leaves a later assertion reading stale state.
func submitAndWait(
	t *testing.T,
	runner *integration.Runner,
	flat transaction.FlatTransaction,
	signer *wallet.Wallet,
) *transactions.TxResponse {
	t.Helper()

	response, err := runner.TestSuccessfulTransactionAndWait(&flat, signer, nil)
	require.NoError(t, err)
	t.Logf("validated %s %s in ledger %d with %s", flat["TransactionType"], response.Hash, response.LedgerIndex, response.Meta.TransactionResult)
	return response
}

// submitExpectingResult submits a transaction the network is expected to apply with a
// specific claimed-fee result, such as tecBAD_PROOF. It cannot go through the runner,
// which asserts tesSUCCESS and submits with fail_hard.
func submitExpectingResult(
	t *testing.T,
	client confidentialClient,
	flat transaction.FlatTransaction,
	signer *wallet.Wallet,
	expectedResult string,
) *transactions.TxResponse {
	t.Helper()

	require.NoError(t, client.Autofill(&flat))
	blob, hash, err := signer.Sign(flat)
	require.NoError(t, err)

	response, err := client.SubmitTxBlobAndWait(blob, false)
	require.NoError(t, err)
	require.True(t, response.Validated)
	require.Equal(t, hash, string(response.Hash))
	require.Equal(t, expectedResult, response.Meta.TransactionResult)
	t.Logf("validated %s %s in ledger %d with %s", flat["TransactionType"], response.Hash, response.LedgerIndex, response.Meta.TransactionResult)
	return response
}

// accountSequence reads the sequence the next transaction from address must spend. It
// reads the current ledger, because a scenario that assembles several transactions up
// front needs the sequence that follows everything already submitted.
func accountSequence(t *testing.T, client confidentialClient, address types.Address) uint32 {
	t.Helper()

	info, err := client.GetAccountInfo(&account.InfoRequest{Account: address, LedgerIndex: common.Current})
	require.NoError(t, err)
	require.NotZero(t, info.AccountData.Sequence)
	return info.AccountData.Sequence
}

func getIssuanceID(t *testing.T, client confidentialClient, issuer types.Address) string {
	t.Helper()

	objects, err := client.GetAccountObjects(&account.ObjectsRequest{Account: issuer, Type: account.MPTIssuanceObject, LedgerIndex: common.Validated})
	require.NoError(t, err)
	require.Len(t, objects.AccountObjects, 1)

	issuanceID, ok := objects.AccountObjects[0]["mpt_issuance_id"].(string)
	require.True(t, ok)
	return issuanceID
}

func getMPToken(t *testing.T, client confidentialClient, holder types.Address) ledger.MPToken {
	t.Helper()

	objects, err := client.GetAccountObjects(&account.ObjectsRequest{Account: holder, Type: account.MPTokenObject, LedgerIndex: common.Validated})
	require.NoError(t, err)
	require.Len(t, objects.AccountObjects, 1)
	return integration.DecodeLedgerObject[ledger.MPToken](t, objects.AccountObjects[0])
}

func getIssuance(t *testing.T, client confidentialClient, issuer types.Address) ledger.MPTokenIssuance {
	t.Helper()

	objects, err := client.GetAccountObjects(&account.ObjectsRequest{Account: issuer, Type: account.MPTIssuanceObject, LedgerIndex: common.Validated})
	require.NoError(t, err)
	require.Len(t, objects.AccountObjects, 1)
	return integration.DecodeLedgerObject[ledger.MPTokenIssuance](t, objects.AccountObjects[0])
}

// assertMirrorBalances checks that the issuer and auditor ciphertexts on the holder's
// MPToken decrypt to the same amount the holder holds. The mirrors are what make a
// confidential balance auditable, so every balance change is checked through them.
func assertMirrorBalances(
	t *testing.T,
	client confidentialClient,
	holder types.Address,
	config issuanceConfig,
	amount uint64,
) {
	t.Helper()

	token := getMPToken(t, client, holder)
	require.Equal(t, amount, decryptBalance(t, token.IssuerEncryptedBalance, config.issuerKey.PrivKeyHex))
	if config.auditorKey == nil {
		require.Empty(t, token.AuditorEncryptedBalance)
		return
	}
	require.Equal(t, amount, decryptBalance(t, token.AuditorEncryptedBalance, config.auditorKey.PrivKeyHex))
}

// assertSplitBalances checks the inbox/spending split XLS-96 defines in section 5.2, which
// the issuer and auditor mirrors cannot see: a mirror holds inbox + spending, so a credit
// posted to the wrong side of the split leaves it unchanged. Merging is the only thing that
// moves value from one side to the other, so the split is what makes a merge observable.
func assertSplitBalances(
	t *testing.T,
	client confidentialClient,
	holder types.Address,
	privateKey string,
	inbox, spending uint64,
	version uint32,
) {
	t.Helper()

	token := getMPToken(t, client, holder)
	require.Equal(t, inbox, decryptBalance(t, token.ConfidentialBalanceInbox, privateKey))
	require.Equal(t, spending, decryptBalance(t, token.ConfidentialBalanceSpending, privateKey))
	require.Equal(t, version, token.ConfidentialBalanceVersion)
}

// exactRange bounds a decryption search to a single candidate. It is for the BalanceRange
// a builder decrypts its own spending balance with, where the amount is known in advance.
func exactRange(amount uint64) elgamal.AmountRange {
	return elgamal.AmountRange{Low: amount, High: amount}
}

// decryptBalance searches every amount an issuance here can hold rather than the one the
// caller expects, so the returned value is evidence rather than an echo of the expectation.
// A wrong balance then fails on the comparison, naming both numbers, instead of surfacing
// as an opaque decryption error.
func decryptBalance(t *testing.T, ciphertext, privateKey string) uint64 {
	t.Helper()

	decrypted, err := elgamal.Decrypt(ciphertext, privateKey, elgamal.AmountRange{Low: 0, High: uint64(issuanceMaximumAmount)})
	require.NoError(t, err)
	return decrypted
}

func parseMPTAmount(t *testing.T, amount string) uint64 {
	t.Helper()

	parsed, err := strconv.ParseUint(amount, 10, 64)
	require.NoError(t, err)
	return parsed
}
