package server_test

import (
	"encoding/json"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/require"
)

func TestManifestRequest(t *testing.T) {
	s := server.ManifestRequest{
		PublicKey: "nHUFE9prPXPrHcG3SkwP1UzAQbSphqyQkQK9ATXLZsfkezhhda3p",
	}

	j := `{
	"public_key": "nHUFE9prPXPrHcG3SkwP1UzAQbSphqyQkQK9ATXLZsfkezhhda3p"
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func manifestResponseFixture() (server.ManifestResponse, string) {
	s := server.ManifestResponse{
		Details: server.ManifestDetails{
			Domain:       "example.com",
			EphemeralKey: "n9J67zk4B7GpbQV5jRQntbgdKf7TW6894QuG7qq1rE5gvjCu6snA",
			MasterKey:    "nHUFE9prPXPrHcG3SkwP1UzAQbSphqyQkQK9ATXLZsfkezhhda3p",
			Seq:          1,
		},
		Manifest:  "JAAAAAFxIe3AkJgOyqs3y+UuiAI27Ff3Mrfbt8e7mjdo06bnGEp5XnMhAhRmvCZmWZXlwShVE9qXs2AVCvhVuA/WGYkTX/vVGBGwdkYwRAIgGnYpIGufURojN2cTXakAM7Vwa0GR7o3osdVlZShroXQCIH9R/Lx1v9rdb4YY2n5nrxdnhSSof3U6V/wIHJmeao5ucBJA9D1iAMo7YFCpb245N3Czc0L1R2Xac0YwQ6XdGT+cZ7yw2n8JbdC3hH8Xu9OUqc867Ee6JmlXtyDHzBdY/hdJCQ==",
		Requested: "nHUFE9prPXPrHcG3SkwP1UzAQbSphqyQkQK9ATXLZsfkezhhda3p",
	}
	j := `{
	"details": {
		"domain": "example.com",
		"ephemeral_key": "n9J67zk4B7GpbQV5jRQntbgdKf7TW6894QuG7qq1rE5gvjCu6snA",
		"master_key": "nHUFE9prPXPrHcG3SkwP1UzAQbSphqyQkQK9ATXLZsfkezhhda3p",
		"seq": 1
	},
	"manifest": "JAAAAAFxIe3AkJgOyqs3y+UuiAI27Ff3Mrfbt8e7mjdo06bnGEp5XnMhAhRmvCZmWZXlwShVE9qXs2AVCvhVuA/WGYkTX/vVGBGwdkYwRAIgGnYpIGufURojN2cTXakAM7Vwa0GR7o3osdVlZShroXQCIH9R/Lx1v9rdb4YY2n5nrxdnhSSof3U6V/wIHJmeao5ucBJA9D1iAMo7YFCpb245N3Czc0L1R2Xac0YwQ6XdGT+cZ7yw2n8JbdC3hH8Xu9OUqc867Ee6JmlXtyDHzBdY/hdJCQ==",
	"requested": "nHUFE9prPXPrHcG3SkwP1UzAQbSphqyQkQK9ATXLZsfkezhhda3p"
}`
	return s, j
}

func TestManifestResponseSerialize(t *testing.T) {
	value, payload := manifestResponseFixture()
	require.NoError(t, testutil.Serialize(t, value, payload))
}

func TestManifestResponseJSONDecode(t *testing.T) {
	want, payload := manifestResponseFixture()
	var got server.ManifestResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestManifestResponseClientDecode(t *testing.T) {
	want, payload := manifestResponseFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got server.ManifestResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}
