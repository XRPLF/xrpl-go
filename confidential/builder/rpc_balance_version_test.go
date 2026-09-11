//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package builder

import (
	"fmt"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/stretchr/testify/require"
)

// TestBuildSendUsesRPCDecodedBalanceVersion needs the native prover, so it carries the CGo
// tag. The error-classification tests that share queuedRPCTransport do not, and live in
// client_errors_test.go so a CGO_ENABLED=0 run still covers them.
func TestBuildSendUsesRPCDecodedBalanceVersion(t *testing.T) {
	const (
		sequence       uint32 = 8
		balanceVersion uint32 = 2
		currentBalance uint64 = 1000
		sendAmount     uint64 = 300
	)

	senderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	receiverKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	balanceBF, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	balanceCiphertext, err := elgamal.Encrypt(currentBalance, senderKP.PubKeyHex, balanceBF)
	require.NoError(t, err)
	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)
	senderIndex, err := xrplhash.MPToken(testIssuanceID, testAccount)
	require.NoError(t, err)
	receiverIndex, err := xrplhash.MPToken(testIssuanceID, testDestination)
	require.NoError(t, err)

	entryResponse := func(index, node string) string {
		return fmt.Sprintf(`{"result":{"index":"%s","ledger_hash":"%s","ledger_index":%d,"node":%s,"validated":true}}`, index, mockLedgerHash, mockLedgerIndex, node)
	}
	// rippled identifies an open ledger by ledger_current_index alone, and the balance
	// version is only compared when that index leads the pinned validated ledger.
	openEntryResponse := func(index, node string) string {
		return fmt.Sprintf(`{"result":{"index":"%s","ledger_current_index":%d,"node":%s}}`, index, mockOpenLedgerIndex, node)
	}
	accountResponse := fmt.Sprintf(`{"result":{"account_data":{"Sequence":%d},"ledger_index":%d,"validated":true}}`, sequence, mockLedgerIndex)
	transport := &queuedRPCTransport{responses: []string{
		// One account_info selects the validated ledger, the next reads the open sequence.
		accountResponse,
		accountResponse,
		entryResponse(issuanceIndex, fmt.Sprintf(`{"LedgerEntryType":"MPTokenIssuance","IssuerEncryptionKey":"%s","Flags":%d,"ConfidentialOutstandingAmount":"1000000"}`, issuerKP.PubKeyHex, confidentialIssuanceFlags)),
		entryResponse(senderIndex, fmt.Sprintf(`{"LedgerEntryType":"MPToken","HolderEncryptionKey":"%s","ConfidentialBalanceSpending":"%s","ConfidentialBalanceVersion":%d,"IssuerEncryptedBalance":"%s"}`, senderKP.PubKeyHex, balanceCiphertext, balanceVersion, testIssuerMirrorCt)),
		// The open-ledger reread of the sender MPToken reports the same version, so the
		// balance the proof consumes is not superseded by a transaction still in flight.
		openEntryResponse(senderIndex, fmt.Sprintf(`{"LedgerEntryType":"MPToken","HolderEncryptionKey":"%s","ConfidentialBalanceSpending":"%s","ConfidentialBalanceVersion":%d,"IssuerEncryptedBalance":"%s"}`, senderKP.PubKeyHex, balanceCiphertext, balanceVersion, testIssuerMirrorCt)),
		entryResponse(receiverIndex, fmt.Sprintf(`{"LedgerEntryType":"MPToken","HolderEncryptionKey":"%s","ConfidentialBalanceInbox":"%s","IssuerEncryptedBalance":"%s"}`, receiverKP.PubKeyHex, testInboxCt, testIssuerMirrorCt)),
	}}
	config, err := rpc.NewClientConfig("http://testnode/", rpc.WithHTTPClient(transport))
	require.NoError(t, err)
	client := rpc.NewClient(config)

	result, err := BuildSend(client, BuildSendParams{
		Account:       testAccount,
		Destination:   testDestination,
		IssuanceID:    testIssuanceID,
		Amount:        sendAmount,
		SenderPrivKey: senderKP.PrivKeyHex,
		SenderPubKey:  senderKP.PubKeyHex,
		BalanceRange:  elgamal.AmountRange{Low: currentBalance, High: currentBalance},
	})
	require.NoError(t, err)
	require.Empty(t, transport.responses)

	contextHash, err := proof.SendContextHash(testAccount, testIssuanceID, sequence, testDestination, balanceVersion)
	require.NoError(t, err)
	participants := []proof.Participant{
		{PubKeyHex: senderKP.PubKeyHex, CiphertextHex: result.SenderEncryptedAmount},
		{PubKeyHex: receiverKP.PubKeyHex, CiphertextHex: result.DestinationEncryptedAmount},
		{PubKeyHex: issuerKP.PubKeyHex, CiphertextHex: result.IssuerEncryptedAmount},
	}
	require.NoError(t, proof.VerifySendProof(
		result.ZKProof,
		participants,
		balanceCiphertext,
		result.AmountCommitment,
		result.BalanceCommitment,
		contextHash,
	))
}
