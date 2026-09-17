//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package confidential

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/builder"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

const (
	clawbackFunding      uint64 = 30
	clawbackConfidential uint64 = 20
)

// testIntegrationConfidentialMPTClawback checks that an issuer can remove a holder's
// whole confidential balance, and that the amount the builder claws back is the one it
// decrypted from the issuer mirror rather than one the caller had to know.
func testIntegrationConfidentialMPTClawback(t *testing.T, client confidentialClient) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 2})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	holder := runner.GetWallet(1)

	auditorKey := generateKey(t)
	// canLock is what makes the section 11.1 lock-then-clawback flow below possible: an
	// issuance created without it rejects the lock with tecNO_PERMISSION.
	config := issuanceConfig{issuerKey: generateKey(t), auditorKey: &auditorKey, canClawback: true, canLock: true}
	holderKey := generateKey(t)

	issuanceID := createIssuance(t, runner, client, issuer, config)
	authorizeHolder(t, runner, issuer, holder, issuanceID)
	fundHolder(t, runner, issuer, holder, issuanceID, clawbackFunding)

	convert, err := builder.BuildConvert(client, builder.BuildConvertParams{
		Account:       holder.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        clawbackConfidential,
		HolderPrivKey: holderKey.PrivKeyHex,
		HolderPubKey:  holderKey.PubKeyHex,
	})
	require.NoError(t, err)
	submitAndWait(t, runner, convert.Flatten(), holder)

	merge, err := builder.BuildMergeInbox(client, builder.BuildMergeInboxParams{
		Account:    holder.GetAddress().String(),
		IssuanceID: issuanceID,
	})
	require.NoError(t, err)
	submitAndWait(t, runner, merge.Flatten(), holder)

	const publicRemainder = clawbackFunding - clawbackConfidential
	before := getMPToken(t, client, holder.GetAddress())
	require.Equal(t, publicRemainder, parseMPTAmount(t, before.MPTAmount))
	require.Equal(t, clawbackConfidential, decryptBalance(t, before.ConfidentialBalanceSpending, holderKey.PrivKeyHex))
	assertMirrorBalances(t, client, holder.GetAddress(), config, clawbackConfidential)

	// XLS-96 section 11.1 recommends locking a holder before clawing back, so the two have
	// to coexist. rippled makes the asymmetry explicit: ConfidentialMPTSend::preclaim calls
	// checkFrozen for both parties, ConfidentialMPTClawback::preclaim calls it for neither,
	// and the builder mirrors that by reading the issuer ciphertext without the holder
	// preflight every other builder runs.
	holderAddress := holder.GetAddress()
	lock := transaction.MPTokenIssuanceSet{
		BaseTx:            transaction.BaseTx{Account: issuer.GetAddress()},
		MPTokenIssuanceID: issuanceID,
		Holder:            &holderAddress,
	}
	lock.SetMPTLockFlag()
	submitAndWait(t, runner, lock.Flatten(), issuer)

	// Everything else the holder could do is now blocked, and the builder says so before
	// spending a fee on it.
	_, err = builder.BuildMergeInbox(client, builder.BuildMergeInboxParams{
		Account:    holder.GetAddress().String(),
		IssuanceID: issuanceID,
	})
	require.ErrorIs(t, err, builder.ErrHolderLocked)

	// The issuer supplies no amount: the builder reads the issuer mirror and claws back
	// whatever the holder actually holds.
	clawback, err := builder.BuildClawback(client, builder.BuildClawbackParams{
		Account:       issuer.GetAddress().String(),
		Holder:        holder.GetAddress().String(),
		IssuanceID:    issuanceID,
		IssuerPrivKey: config.issuerKey.PrivKeyHex,
		BalanceRange:  exactRange(clawbackConfidential),
	})
	require.NoError(t, err)
	submitAndWait(t, runner, clawback.Flatten(), issuer)

	after := getMPToken(t, client, holder.GetAddress())
	require.Equal(t, before.ConfidentialBalanceVersion+1, after.ConfidentialBalanceVersion)
	// A clawback takes the confidential balance only, so the public balance is untouched.
	require.Equal(t, publicRemainder, parseMPTAmount(t, after.MPTAmount))
	require.Equal(t, uint64(0), decryptBalance(t, after.ConfidentialBalanceSpending, holderKey.PrivKeyHex))
	require.Equal(t, uint64(0), decryptBalance(t, after.ConfidentialBalanceInbox, holderKey.PrivKeyHex))
	assertMirrorBalances(t, client, holder.GetAddress(), config, 0)

	issuance := getIssuance(t, client, issuer.GetAddress())
	require.Equal(t, publicRemainder, parseMPTAmount(t, issuance.OutstandingAmount))
	// rippled omits a zero ConfidentialOutstandingAmount rather than serializing "0".
	require.Empty(t, issuance.ConfidentialOutstandingAmount)
	require.Equal(
		t,
		ledger.LsfMPTRequireAuth|ledger.LsfMPTCanLock|ledger.LsfMPTCanTransfer|ledger.LsfMPTCanClawback|ledger.LsfMPTCanHoldConfidentialBalance,
		issuance.Flags,
	)

	// Clawing back an empty balance has nothing to prove, so the builder stops first.
	_, err = builder.BuildClawback(client, builder.BuildClawbackParams{
		Account:       issuer.GetAddress().String(),
		Holder:        holder.GetAddress().String(),
		IssuanceID:    issuanceID,
		IssuerPrivKey: config.issuerKey.PrivKeyHex,
		BalanceRange:  exactRange(0),
	})
	require.ErrorIs(t, err, builder.ErrZeroAmount)
}

func TestIntegrationConfidentialMPTClawback_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationConfidentialMPTClawback(t, client)
}

func TestIntegrationConfidentialMPTClawback_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationConfidentialMPTClawback(t, client)
}
