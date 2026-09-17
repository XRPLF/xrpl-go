//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package confidential

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/builder"
	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

// Amounts for the dependent-batch scenario.
const (
	batchSenderFunding   uint64 = 200
	batchReceiverFunding uint64 = 40
	batchFirstSend       uint64 = 30
	batchSecondSend      uint64 = 25
	batchReturnSend      uint64 = 11
)

// batchSearchRange bounds every decryption this scenario runs.
func batchSearchRange() elgamal.AmountRange {
	return elgamal.AmountRange{Low: 0, High: uint64(issuanceMaximumAmount)}
}

// testIntegrationConfidentialMPTDependentBatch submits one Batch whose inners depend on each
// other, which is the case no sequence of standalone builders can produce.
func testIntegrationConfidentialMPTDependentBatch(t *testing.T, client confidentialClient) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 3})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	sender := runner.GetWallet(1)
	receiver := runner.GetWallet(2)

	auditorKey := generateKey(t)
	config := issuanceConfig{issuerKey: generateKey(t), auditorKey: &auditorKey}
	senderKey := generateKey(t)
	receiverKey := generateKey(t)

	issuanceID := createIssuance(t, runner, client, issuer, config)
	authorizeHolder(t, runner, issuer, sender, issuanceID)
	authorizeHolder(t, runner, issuer, receiver, issuanceID)
	fundHolder(t, runner, issuer, sender, issuanceID, batchSenderFunding)
	fundHolder(t, runner, issuer, receiver, issuanceID, batchReceiverFunding)

	convertAndMerge(t, runner, client, sender, issuanceID, senderKey, batchSenderFunding)
	convertAndMerge(t, runner, client, receiver, issuanceID, receiverKey, batchReceiverFunding)

	senderAddress := sender.GetAddress().String()
	receiverAddress := receiver.GetAddress().String()

	batch, err := builder.BuildBatch(client, builder.BuildBatchParams{
		Account: senderAddress,
		Operations: []builder.BatchOperation{
			sendOperation(senderAddress, receiverAddress, issuanceID, senderKey, batchFirstSend),
			sendOperation(senderAddress, receiverAddress, issuanceID, senderKey, batchSecondSend),
			builder.MergeInboxOp{BuildMergeInboxParams: builder.BuildMergeInboxParams{
				Account:    receiverAddress,
				IssuanceID: issuanceID,
			}},
			sendOperation(receiverAddress, senderAddress, issuanceID, receiverKey, batchReturnSend),
		},
	})
	require.NoError(t, err)
	require.Len(t, batch.RawTransactions, 4)
	require.Equal(t, transaction.TfAllOrNothing, batch.Flags)

	require.Equal(t, batch.Sequence+1, innerSequence(t, batch, 0))
	require.Equal(t, batch.Sequence+2, innerSequence(t, batch, 1))
	require.Equal(t, innerSequence(t, batch, 2)+1, innerSequence(t, batch, 3))

	response := submitBatch(t, runner, client, batch, sender, receiver)
	t.Logf("validated Batch %s in ledger %d with %s", response.Hash, response.LedgerIndex, response.Meta.TransactionResult)

	const senderSpending = batchSenderFunding - batchFirstSend - batchSecondSend
	assertSplitBalances(t, client, sender.GetAddress(), senderKey.PrivKeyHex, batchReturnSend, senderSpending, 3)
	assertMirrorBalances(t, client, sender.GetAddress(), config, senderSpending+batchReturnSend)

	const receiverSpending = batchReceiverFunding + batchFirstSend + batchSecondSend - batchReturnSend
	assertSplitBalances(t, client, receiver.GetAddress(), receiverKey.PrivKeyHex, 0, receiverSpending, 3)
	assertMirrorBalances(t, client, receiver.GetAddress(), config, receiverSpending)

	issuance := getIssuance(t, client, issuer.GetAddress())
	require.Equal(t, batchSenderFunding+batchReceiverFunding, parseMPTAmount(t, issuance.ConfidentialOutstandingAmount))
	require.Equal(t, senderSpending+batchReturnSend+receiverSpending, parseMPTAmount(t, issuance.ConfidentialOutstandingAmount))
}

// Amounts for the reset-balance scenario.
const (
	resetHolderFunding   uint64 = 100
	resetReceiverFunding uint64 = 10
	resetFirstConvert    uint64 = 60
	resetSend            uint64 = 15
	resetFirstReveal     uint64 = 5
	resetSecondConvert   uint64 = 20
	resetSecondReveal    uint64 = 5
	resetTicketCount     uint32 = 2
)

// testIntegrationConfidentialMPTResetBalanceBatch submits Batches whose later inners prove
// against balances an earlier inner reset to the canonical encrypted zero: a first-time
// convert's spending balance, a merge's inbox, and every balance a clawback clears. The ledger
// accepts those proofs only if the client reproduced each canonical zero exactly. The first
// Batch also opens with a TicketCreate, whose extra sequence consumption every later inner's
// nonce must account for.
func testIntegrationConfidentialMPTResetBalanceBatch(t *testing.T, client confidentialClient) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 3})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	holder := runner.GetWallet(1)
	receiver := runner.GetWallet(2)

	auditorKey := generateKey(t)
	config := issuanceConfig{issuerKey: generateKey(t), auditorKey: &auditorKey, canClawback: true}
	holderKey := generateKey(t)
	receiverKey := generateKey(t)

	issuanceID := createIssuance(t, runner, client, issuer, config)
	authorizeHolder(t, runner, issuer, holder, issuanceID)
	authorizeHolder(t, runner, issuer, receiver, issuanceID)
	fundHolder(t, runner, issuer, holder, issuanceID, resetHolderFunding)
	fundHolder(t, runner, issuer, receiver, issuanceID, resetReceiverFunding)
	convertAndMerge(t, runner, client, receiver, issuanceID, receiverKey, resetReceiverFunding)

	holderAddress := holder.GetAddress().String()
	issuerAddress := issuer.GetAddress().String()

	first, err := builder.BuildBatch(client, builder.BuildBatchParams{
		Account: holderAddress,
		Operations: []builder.BatchOperation{
			builder.TransactionOp{Tx: &transaction.TicketCreate{
				BaseTx:      transaction.BaseTx{Account: holder.GetAddress(), TransactionType: transaction.TicketCreateTx},
				TicketCount: resetTicketCount,
			}},
			convertOperation(holderAddress, issuanceID, holderKey, resetFirstConvert),
			mergeOperation(holderAddress, issuanceID),
			mergeOperation(holderAddress, issuanceID),
			sendOperation(holderAddress, receiver.GetAddress().String(), issuanceID, holderKey, resetSend),
			convertBackOperation(holderAddress, issuanceID, holderKey, resetFirstReveal),
		},
	})
	require.NoError(t, err)
	require.Equal(t, first.Sequence+1, innerSequence(t, first, 0))
	require.Equal(t, first.Sequence+2+resetTicketCount, innerSequence(t, first, 1),
		"the inner after a TicketCreate skips every sequence the Tickets took")

	response := submitBatch(t, runner, client, first, holder)
	t.Logf("validated Batch %s in ledger %d with %s", response.Hash, response.LedgerIndex, response.Meta.TransactionResult)

	holderSpending := resetFirstConvert - resetSend - resetFirstReveal
	assertSplitBalances(t, client, holder.GetAddress(), holderKey.PrivKeyHex, 0, holderSpending, 4)
	assertMirrorBalances(t, client, holder.GetAddress(), config, holderSpending)
	assertSplitBalances(t, client, receiver.GetAddress(), receiverKey.PrivKeyHex, resetSend, resetReceiverFunding, 1)

	second, err := builder.BuildBatch(client, builder.BuildBatchParams{
		Account: issuerAddress,
		Operations: []builder.BatchOperation{
			builder.ClawbackOp{BuildClawbackParams: builder.BuildClawbackParams{
				Account:       issuerAddress,
				Holder:        holderAddress,
				IssuanceID:    issuanceID,
				IssuerPrivKey: config.issuerKey.PrivKeyHex,
				BalanceRange:  batchSearchRange(),
			}},
			convertOperation(holderAddress, issuanceID, holderKey, resetSecondConvert),
			mergeOperation(holderAddress, issuanceID),
			convertBackOperation(holderAddress, issuanceID, holderKey, resetSecondReveal),
		},
	})
	require.NoError(t, err)
	require.Equal(t, types.MPTPlainAmount(holderSpending).String(), second.RawTransactions[0].RawTransaction["MPTAmount"])

	response = submitBatch(t, runner, client, second, issuer, holder)
	t.Logf("validated Batch %s in ledger %d with %s", response.Hash, response.LedgerIndex, response.Meta.TransactionResult)

	holderSpending = resetSecondConvert - resetSecondReveal
	assertSplitBalances(t, client, holder.GetAddress(), holderKey.PrivKeyHex, 0, holderSpending, 7)
	assertMirrorBalances(t, client, holder.GetAddress(), config, holderSpending)

	issuance := getIssuance(t, client, issuer.GetAddress())
	require.Equal(t, holderSpending+resetReceiverFunding+resetSend, parseMPTAmount(t, issuance.ConfidentialOutstandingAmount))
}

// convertOperation builds one convert inner.
func convertOperation(account, issuanceID string, key elgamal.Keypair, amount uint64) builder.ConvertOp {
	return builder.ConvertOp{BuildConvertParams: builder.BuildConvertParams{
		Account:       account,
		IssuanceID:    issuanceID,
		Amount:        amount,
		HolderPrivKey: key.PrivKeyHex,
		HolderPubKey:  key.PubKeyHex,
	}}
}

// mergeOperation builds one inbox merge inner.
func mergeOperation(account, issuanceID string) builder.MergeInboxOp {
	return builder.MergeInboxOp{BuildMergeInboxParams: builder.BuildMergeInboxParams{
		Account:    account,
		IssuanceID: issuanceID,
	}}
}

// convertBackOperation builds one convert-back inner.
func convertBackOperation(account, issuanceID string, key elgamal.Keypair, amount uint64) builder.ConvertBackOp {
	return builder.ConvertBackOp{BuildConvertBackParams: builder.BuildConvertBackParams{
		Account:       account,
		IssuanceID:    issuanceID,
		Amount:        amount,
		HolderPrivKey: key.PrivKeyHex,
		HolderPubKey:  key.PubKeyHex,
		BalanceRange:  batchSearchRange(),
	}}
}

// sendOperation builds one send inner of the dependent batch.
func sendOperation(from, to, issuanceID string, key elgamal.Keypair, amount uint64) builder.SendOp {
	return builder.SendOp{BuildSendParams: builder.BuildSendParams{
		Account:       from,
		Destination:   to,
		IssuanceID:    issuanceID,
		Amount:        amount,
		SenderPrivKey: key.PrivKeyHex,
		SenderPubKey:  key.PubKeyHex,
		BalanceRange:  batchSearchRange(),
	}}
}

// convertAndMerge gives a holder a spendable confidential balance, which is the starting state
// every inner of the dependent batch assumes.
func convertAndMerge(
	t *testing.T,
	runner *integration.Runner,
	client confidentialClient,
	holder *wallet.Wallet,
	issuanceID string,
	key elgamal.Keypair,
	amount uint64,
) {
	t.Helper()

	convert, err := builder.BuildConvert(client, builder.BuildConvertParams{
		Account:       holder.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        amount,
		HolderPrivKey: key.PrivKeyHex,
		HolderPubKey:  key.PubKeyHex,
	})
	require.NoError(t, err)
	submitAndWait(t, runner, convert.Flatten(), holder)

	merge, err := builder.BuildMergeInbox(client, builder.BuildMergeInboxParams{
		Account:    holder.GetAddress().String(),
		IssuanceID: issuanceID,
	})
	require.NoError(t, err)
	submitAndWait(t, runner, merge.Flatten(), holder)
}

// submitBatch autofills, signs, and submits an assembled Batch, and waits for validation. Each
// co-signer is an account other than the outer one that owns an inner.
func submitBatch(
	t *testing.T,
	runner *integration.Runner,
	client confidentialClient,
	batch *transaction.Batch,
	outer *wallet.Wallet,
	coSigners ...*wallet.Wallet,
) *transactions.TxResponse {
	t.Helper()

	flat := batch.Flatten()
	require.NoError(t, client.AutofillMultisigned(&flat, uint64(len(coSigners))))
	for _, coSigner := range coSigners {
		require.NoError(t, wallet.SignMultiBatch(*coSigner, &flat, nil))
	}

	response, err := runner.TestSuccessfulTransactionAndWait(&flat, outer, &integration.TestTransactionOptions{SkipAutofill: true})
	require.NoError(t, err)
	return response
}

// innerSequence reads the sequence the assembler assigned to one inner.
func innerSequence(t *testing.T, batch *transaction.Batch, index int) uint32 {
	t.Helper()

	require.Greater(t, len(batch.RawTransactions), index)
	sequence, ok := batch.RawTransactions[index].RawTransaction["Sequence"].(uint32)
	require.True(t, ok, "inner %d must carry a Sequence", index)
	return sequence
}

func TestIntegrationConfidentialMPTDependentBatch_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationConfidentialMPTDependentBatch(t, client)
}

func TestIntegrationConfidentialMPTDependentBatch_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationConfidentialMPTDependentBatch(t, client)
}

func TestIntegrationConfidentialMPTResetBalanceBatch_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationConfidentialMPTResetBalanceBatch(t, client)
}

func TestIntegrationConfidentialMPTResetBalanceBatch_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationConfidentialMPTResetBalanceBatch(t, client)
}
