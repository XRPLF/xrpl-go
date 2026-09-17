package server_test

import (
	"encoding/json"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server/types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/require"
)

func featureAllResponseFixture() (server.FeatureAllResponse, string) {
	r := server.FeatureAllResponse{
		Features: map[string]types.FeatureStatus{
			"feature1": {Enabled: true, Name: "feature1", Supported: true},
			"feature2": {Enabled: false, Name: "feature2", Supported: false},
		},
	}
	s := `{
	"features": {
		"feature1": {
			"enabled": true,
			"name": "feature1",
			"supported": true
		},
		"feature2": {
			"enabled": false,
			"name": "feature2",
			"supported": false
		}
	}
}`
	return r, s
}

func TestFeatureAllResponseSerialize(t *testing.T) {
	value, payload := featureAllResponseFixture()
	require.NoError(t, testutil.Serialize(t, value, payload))
}

func TestFeatureAllResponseJSONDecode(t *testing.T) {
	want, payload := featureAllResponseFixture()
	var got server.FeatureAllResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestFeatureAllResponseClientDecode(t *testing.T) {
	want, payload := featureAllResponseFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got server.FeatureAllResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}

func TestFeatureOneRequest(t *testing.T) {
	r := server.FeatureOneRequest{
		Feature: "feature1",
	}

	s := `{
	"feature": "feature1"
}`

	if err := testutil.Serialize(t, r, s); err != nil {
		t.Fatal(err)
	}
}

func featureOneResponseFixture() (server.FeatureResponse, string) {
	r := server.FeatureResponse{
		"feature1": {Enabled: true, Name: "feature1", Supported: true},
	}
	s := `{
	"feature1": {
		"enabled": true,
		"name": "feature1",
		"supported": true
	}
}`
	return r, s
}

func TestFeatureOneResponseSerialize(t *testing.T) {
	value, payload := featureOneResponseFixture()
	require.NoError(t, testutil.Serialize(t, value, payload))
}

func TestFeatureOneResponseJSONDecode(t *testing.T) {
	want, payload := featureOneResponseFixture()
	var got server.FeatureResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestFeatureOneResponseClientDecode(t *testing.T) {
	want, payload := featureOneResponseFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got server.FeatureResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}
