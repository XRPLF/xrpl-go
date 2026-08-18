//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package builder

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/stretchr/testify/require"
)

type queuedRPCTransport struct {
	responses []string
}

func (t *queuedRPCTransport) Do(*http.Request) (*http.Response, error) {
	if len(t.responses) == 0 {
		return nil, fmt.Errorf("unexpected RPC request")
	}
	response := t.responses[0]
	t.responses = t.responses[1:]
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(response)),
		Header:     make(http.Header),
	}, nil
}

// TestGetMPTokenStateClassifiesRPCEntryNotFound pins the classification against a real
// rpc.Client rather than a hand-built error, because ErrReceiverNotOptedIn depends on the
// node's "entryNotFound" reaching the builder as exactly that string.
func TestGetMPTokenStateClassifiesRPCEntryNotFound(t *testing.T) {
	transport := &queuedRPCTransport{responses: []string{
		`{"result":{"error":"entryNotFound"}}`,
	}}
	config, err := rpc.NewClientConfig("http://testnode/", rpc.WithHTTPClient(transport))
	require.NoError(t, err)
	client := rpc.NewClient(config)

	_, _, _, err = getMPTokenState(client, testIssuanceID, testAccount)
	require.ErrorIs(t, err, ErrMPTokenNotFound)
	require.NotErrorIs(t, err, ErrLedgerQuery)
	require.Empty(t, transport.responses)
}

func TestGetMPTokenStateRejectsRPCNullBalanceVersion(t *testing.T) {
	transport := &queuedRPCTransport{responses: []string{
		`{"result":{"node":{"ConfidentialBalanceVersion":null}}}`,
	}}
	config, err := rpc.NewClientConfig("http://testnode/", rpc.WithHTTPClient(transport))
	require.NoError(t, err)
	client := rpc.NewClient(config)

	_, _, _, err = getMPTokenState(client, testIssuanceID, testAccount)
	require.ErrorIs(t, err, ErrLedgerQuery)
	require.Empty(t, transport.responses)
}

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
