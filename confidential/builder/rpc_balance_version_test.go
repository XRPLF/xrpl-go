//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package builder

import (
	"fmt"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/proof"
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

	transport := &queuedRPCTransport{responses: []string{
		fmt.Sprintf(`{"result":{"account_data":{"Sequence":%d}}}`, sequence),
		fmt.Sprintf(`{"result":{"node":{"IssuerEncryptionKey":"%s","Flags":%d,"ConfidentialOutstandingAmount":"1000000"}}}`, issuerKP.PubKeyHex, confidentialIssuanceFlags),
		fmt.Sprintf(`{"result":{"node":{"HolderEncryptionKey":"%s","ConfidentialBalanceSpending":"%s","ConfidentialBalanceVersion":%d}}}`, senderKP.PubKeyHex, balanceCiphertext, balanceVersion),
		fmt.Sprintf(`{"result":{"node":{"HolderEncryptionKey":"%s"}}}`, receiverKP.PubKeyHex),
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
