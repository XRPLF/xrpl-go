package rpc

import (
	"testing"

	serverquery "github.com/Peersyst/xrpl-go/xrpl/queries/server"
	"github.com/stretchr/testify/require"
)

const serverDefinitionsTestHash = "C685734F5FEB756693B4BB978BBB3A158A65652E71EEB2977068B0D680689213"

func TestClient_GetServerDefinitions(t *testing.T) {
	expected := &serverquery.DefinitionsResponse{Hash: serverDefinitionsTestHash}
	client, request := setupQueryTestClient(t, expected)
	response, err := client.GetServerDefinitions(&serverquery.DefinitionsRequest{Hash: serverDefinitionsTestHash})
	require.NoError(t, err)
	require.Equal(t, expected, response)
	got := request()
	require.Equal(t, "server_definitions", got["command"])
	require.Equal(t, serverDefinitionsTestHash, got["hash"])
	require.NotContains(t, got, "BaseRequest")
}

// A valid hash-only response must still pass request-dependent validation.
func TestClient_GetServerDefinitionsValidatesResponse(t *testing.T) {
	client, _ := setupQueryTestClient(t, &serverquery.DefinitionsResponse{Hash: serverDefinitionsTestHash})
	response, err := client.GetServerDefinitions(&serverquery.DefinitionsRequest{})
	require.ErrorIs(t, err, serverquery.ErrInvalidDefinitionsResponse)
	require.Nil(t, response)
}
