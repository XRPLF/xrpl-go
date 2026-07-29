package definitions

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

type snapshotSource struct {
	RPCURL             string `json:"rpc_url"`
	RPCMethod          string `json:"rpc_method"`
	ServerBuildVersion string `json:"server_build_version"`
	SourceTag          string `json:"source_tag"`
	SourceTagObject    string `json:"source_tag_object"`
	SourceCommit       string `json:"source_commit"`
	RetrievedAt        string `json:"retrieved_at"`
	ResponseSHA256     string `json:"response_sha256"`
	DefinitionsHash    string `json:"definitions_hash"`
}

type snapshotDifference struct {
	Path           string `json:"path"`
	Kind           string `json:"kind"`
	Classification string `json:"classification,omitempty"`
	Reason         string `json:"reason,omitempty"`
}

type snapshotManifest struct {
	Source      snapshotSource       `json:"source"`
	Differences []snapshotDifference `json:"differences"`
}

func TestDefinitionsMatchAuthoritativeSnapshot(t *testing.T) {
	authoritativeBytes, err := os.ReadFile("testdata/rippled-3.2.0-server-definitions.json")
	require.NoError(t, err)

	manifestBytes, err := os.ReadFile("testdata/snapshot-differences.json")
	require.NoError(t, err)

	var manifest snapshotManifest
	require.NoError(t, json.Unmarshal(manifestBytes, &manifest))

	var envelope struct {
		Result map[string]any `json:"result"`
	}
	require.NoError(t, json.Unmarshal(authoritativeBytes, &envelope))
	require.Equal(t, "success", envelope.Result["status"])
	definitionsHash, ok := envelope.Result["hash"].(string)
	require.True(t, ok)

	responseSum := sha256.Sum256(authoritativeBytes)
	require.Equal(t, snapshotSource{
		RPCURL:             "https://xrplcluster.com",
		RPCMethod:          "server_definitions",
		ServerBuildVersion: "3.2.0",
		SourceTag:          "3.2.0",
		SourceTagObject:    "e963f4f5b95592c8ae25f7a11c406998874454e3",
		SourceCommit:       "3c43f4614f87965298773279ff5b85d4c56c637b",
		RetrievedAt:        "2026-07-29T16:04:50Z",
		ResponseSHA256:     hex.EncodeToString(responseSum[:]),
		DefinitionsHash:    definitionsHash,
	}, manifest.Source)

	embedded := make(map[string]any)
	require.NoError(t, json.Unmarshal(docBytes, &embedded))

	actual := collectSnapshotDifferences("", embedded, envelope.Result)
	sortSnapshotDifferences(actual)

	expected := make([]snapshotDifference, 0, len(manifest.Differences))
	for _, difference := range manifest.Differences {
		require.NotEmpty(t, difference.Classification, difference.Path)
		require.NotEmpty(t, difference.Reason, difference.Path)
		expected = append(expected, snapshotDifference{Path: difference.Path, Kind: difference.Kind})
	}
	sortSnapshotDifferences(expected)

	require.Equal(t, expected, actual)
}

func collectSnapshotDifferences(path string, embedded, authoritative any) []snapshotDifference {
	switch embeddedValue := embedded.(type) {
	case map[string]any:
		if authoritativeMap, ok := authoritative.(map[string]any); ok {
			return compareSnapshotMaps(path, embeddedValue, authoritativeMap)
		}
	case []any:
		if authoritativeSlice, ok := authoritative.([]any); ok {
			return compareSnapshotSlices(path, embeddedValue, authoritativeSlice)
		}
	default:
		if embedded == authoritative {
			return nil
		}
	}
	return []snapshotDifference{{Path: path, Kind: "value-mismatch"}}
}

func compareSnapshotMaps(path string, embedded, authoritative map[string]any) []snapshotDifference {
	keys := slices.Collect(maps.Keys(embedded))
	for key := range authoritative {
		if _, inEmbedded := embedded[key]; !inEmbedded {
			keys = append(keys, key)
		}
	}
	slices.Sort(keys)

	var differences []snapshotDifference
	for _, key := range keys {
		embeddedValue, inEmbedded := embedded[key]
		authoritativeValue, inAuthoritative := authoritative[key]
		childPath := joinSnapshotPath(path, key)
		switch {
		case !inEmbedded:
			differences = append(differences, snapshotDifference{Path: childPath, Kind: "authoritative-only"})
		case !inAuthoritative:
			differences = append(differences, snapshotDifference{Path: childPath, Kind: "go-only"})
		default:
			differences = append(differences, collectSnapshotDifferences(childPath, embeddedValue, authoritativeValue)...)
		}
	}
	return differences
}

func compareSnapshotSlices(path string, embedded, authoritative []any) []snapshotDifference {
	var differences []snapshotDifference
	commonLength := min(len(embedded), len(authoritative))
	for i := range commonLength {
		childPath := fmt.Sprintf("%s[%d]", path, i)
		differences = append(differences, collectSnapshotDifferences(childPath, embedded[i], authoritative[i])...)
	}
	for i := commonLength; i < len(embedded); i++ {
		differences = append(differences, snapshotDifference{Path: fmt.Sprintf("%s[%d]", path, i), Kind: "go-only"})
	}
	for i := commonLength; i < len(authoritative); i++ {
		differences = append(differences, snapshotDifference{Path: fmt.Sprintf("%s[%d]", path, i), Kind: "authoritative-only"})
	}
	return differences
}

func joinSnapshotPath(parent, child string) string {
	if parent == "" {
		return child
	}
	return parent + "." + child
}

func sortSnapshotDifferences(differences []snapshotDifference) {
	slices.SortFunc(differences, func(a, b snapshotDifference) int {
		return cmp.Or(cmp.Compare(a.Path, b.Path), cmp.Compare(a.Kind, b.Kind))
	})
}
