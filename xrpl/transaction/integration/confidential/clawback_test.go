//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package confidential

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/builder"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

const (
	clawbackFunding    uint64 = 30
	clawbackConfidence uint64 = 20
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
	config := issuanceConfig{issuerKey: generateKey(t), auditorKey: &auditorKey, canClawback: true}
	holderKey := generateKey(t)

	issuanceID := createIssuance(t, runner, client, issuer, config)
	authorizeHolder(t, runner, issuer, holder, issuanceID)
	fundHolder(t, runner, issuer, holder, issuanceID, clawbackFunding)

	convert, err := builder.BuildConvert(client, builder.BuildConvertParams{
		Account:       holder.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        clawbackConfidence,
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

	const publicRemainder = clawbackFunding - clawbackConfidence
	before := getMPToken(t, client, holder.GetAddress())
	require.Equal(t, publicRemainder, parseMPTAmount(t, before.MPTAmount))
	require.Equal(t, clawbackConfidence, decryptBalance(t, before.ConfidentialBalanceSpending, holderKey.PrivKeyHex, clawbackConfidence))
	assertMirrorBalances(t, client, holder.GetAddress(), config, clawbackConfidence)

	// The issuer supplies no amount: the builder reads the issuer mirror and claws back
	// whatever the holder actually holds.
	clawback, err := builder.BuildClawback(client, builder.BuildClawbackParams{
		Account:       issuer.GetAddress().String(),
		Holder:        holder.GetAddress().String(),
		IssuanceID:    issuanceID,
		IssuerPrivKey: config.issuerKey.PrivKeyHex,
		BalanceRange:  exactRange(clawbackConfidence),
	})
	require.NoError(t, err)
	submitAndWait(t, runner, clawback.Flatten(), issuer)

	after := getMPToken(t, client, holder.GetAddress())
	require.Equal(t, before.ConfidentialBalanceVersion+1, after.ConfidentialBalanceVersion)
	// A clawback takes the confidential balance only, so the public balance is untouched.
	require.Equal(t, publicRemainder, parseMPTAmount(t, after.MPTAmount))
	require.Equal(t, uint64(0), decryptBalance(t, after.ConfidentialBalanceSpending, holderKey.PrivKeyHex, 0))
	require.Equal(t, uint64(0), decryptBalance(t, after.ConfidentialBalanceInbox, holderKey.PrivKeyHex, 0))
	assertMirrorBalances(t, client, holder.GetAddress(), config, 0)

	issuance := getIssuance(t, client, issuer.GetAddress())
	require.Equal(t, publicRemainder, parseMPTAmount(t, issuance.OutstandingAmount))
	// rippled omits a zero ConfidentialOutstandingAmount rather than serializing "0".
	require.Empty(t, issuance.ConfidentialOutstandingAmount)
	require.Equal(
		t,
		ledger.LsfMPTRequireAuth|ledger.LsfMPTCanTransfer|ledger.LsfMPTCanClawback|ledger.LsfMPTCanHoldConfidentialBalance,
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
