package types

import (
	"encoding/json"
	"testing"
)

func TestMPTokenMetadataFromBlob(t *testing.T) {
	tests := []struct {
		name        string
		blob        string
		expected    *MPTokenMetadata
		expectError bool
		errorMsg    string
	}{
		{
			name: "PASS: valid metadata with all fields",
			blob: "7b227469636b6572223a22425443222c226e616d65223a22426974636f696e222c2264657363223a2241206469676974616c2063757272656e6379222c2269636f6e223a2268747470733a2f2f6578616d706c652e636f6d2f626974636f696e2e706e67222c2261737365745f636c617373223a2263727970746f63757272656e6379222c2261737365745f737562636c617373223a22626974636f696e222c226973737565725f6e616d65223a225361746f736869204e616b616d6f746f222c2275726c73223a5b7b2275726c223a2268747470733a2f2f626974636f696e2e6f7267222c2274797065223a22746578742f68746d6c222c227469746c65223a22426974636f696e204f6666696369616c227d5d2c226164646974696f6e616c5f696e666f223a7b2276657273696f6e223a22312e30227d7d",
			expected: &MPTokenMetadata{
				Ticker:        "BTC",
				Name:          "Bitcoin",
				Desc:          "A digital currency",
				Icon:          "https://example.com/bitcoin.png",
				AssetClass:    "cryptocurrency",
				AssetSubclass: "bitcoin",
				IssuerName:    "Satoshi Nakamoto",
				URLs: []MPTokenMetadataURL{
					{
						URL:   "https://bitcoin.org",
						Type:  "text/html",
						Title: "Bitcoin Official",
					},
				},
				AdditionalInfo: json.RawMessage(`{"version":"1.0"}`),
			},
			expectError: false,
		},
		{
			name: "PASS: valid metadata with minimal required fields",
			blob: "7b227469636b6572223a22455448222c226e616d65223a22457468657265756d222c2261737365745f636c617373223a2263727970746f63757272656e6379222c2261737365745f737562636c617373223a22657468657265756d227d",
			expected: &MPTokenMetadata{
				Ticker:        "ETH",
				Name:          "Ethereum",
				AssetClass:    "cryptocurrency",
				AssetSubclass: "ethereum",
			},
			expectError: false,
		},
		{
			name:        "FAIL: valid metadata with empty fields - missing required fields",
			blob:        "7b7d",
			expected:    nil,
			expectError: true,
			errorMsg:    "metadata validation failed",
		},
		{
			name: "PASS: valid metadata with urls array",
			blob: "7b227469636b6572223a2255534454222c226e616d65223a2255534420546574686572222c2261737365745f636c617373223a22737461626c65636f696e222c2261737365745f737562636c617373223a22746574686572222c2275726c73223a5b7b2275726c223a2268747470733a2f2f6578616d706c652e636f6d222c2274797065223a226170706c69636174696f6e2f6a736f6e227d2c7b2275726c223a2268747470733a2f2f646f63732e6578616d706c652e636f6d222c227469746c65223a22446f63756d656e746174696f6e227d5d7d",
			expected: &MPTokenMetadata{
				Ticker:        "USDT",
				Name:          "USD Tether",
				AssetClass:    "stablecoin",
				AssetSubclass: "tether",
				URLs: []MPTokenMetadataURL{
					{
						URL:  "https://example.com",
						Type: "application/json",
					},
					{
						URL:   "https://docs.example.com",
						Title: "Documentation",
					},
				},
			},
			expectError: false,
		},
		{
			name:        "FAIL: invalid hex string",
			blob:        "invalid-hex-string",
			expected:    nil,
			expectError: true,
			errorMsg:    "decode from blob in hex",
		},
		{
			name:        "FAIL: empty hex string",
			blob:        "",
			expected:    nil,
			expectError: true,
			errorMsg:    "metadata is not in XLS-0089d schema",
		},
		{
			name:        "FAIL: invalid JSON in hex",
			blob:        "7b227469636b6572223a22425443222c226e616d65223a22426974636f696e222c2264657363223a2241206469676974616c2063757272656e6379222c2269636f6e223a2268747470733a2f2f6578616d706c652e636f6d2f626974636f696e2e706e67222c2261737365745f636c617373223a2263727970746f63757272656e6379222c2261737365745f737562636c617373223a22626974636f696e222c226973737565725f6e616d65223a225361746f736869204e616b616d6f746f222c2275726c73223a5b7b2275726c223a2268747470733a2f2f626974636f696e2e6f7267222c2274797065223a22746578742f68746d6c222c227469746c65223a22426974636f696e204f6666696369616c227d5d2c226164646974696f6e616c5f696e666f223a7b2276657273696f6e223a22312e30227d", // missing closing brace
			expected:    nil,
			expectError: true,
			errorMsg:    "metadata is not in XLS-0089d schema",
		},
		{
			name:        "FAIL: hex string with odd length",
			blob:        "7b227469636b6572223a22425443222c226e616d65223a22426974636f696e222c2264657363223a2241206469676974616c2063757272656e6379222c2269636f6e223a2268747470733a2f2f6578616d706c652e636f6d2f626974636f696e2e706e67222c2261737365745f636c617373223a2263727970746f63757272656e6379222c2261737365745f737562636c617373223a22626974636f696e222c226973737565725f6e616d65223a225361746f736869204e616b616d6f746f222c2275726c73223a5b7b2275726c223a2268747470733a2f2f626974636f696e2e6f7267222c2274797065223a22746578742f68746d6c222c227469746c65223a22426974636f696e204f6666696369616c227d5d2c226164646974696f6e616c5f696e666f223a7b2276657273696f6e223a22312e30227d7", // odd length
			expected:    nil,
			expectError: true,
			errorMsg:    "decode from blob in hex",
		},
		{
			name:        "FAIL: non-hex characters",
			blob:        "7g227469636b6572223a22425443222c226e616d65223a22426974636f696e227d",
			expected:    nil,
			expectError: true,
			errorMsg:    "decode from blob in hex",
		},
		{
			name:        "FAIL: missing required field - ticker",
			blob:        "7b226e616d65223a22426974636f696e222c2261737365745f636c617373223a2263727970746f63757272656e6379222c2261737365745f737562636c617373223a22626974636f696e227d",
			expected:    nil,
			expectError: true,
			errorMsg:    "metadata validation failed",
		},
		{
			name:        "FAIL: missing required field - name",
			blob:        "7b227469636b6572223a22425443222c2261737365745f636c617373223a2263727970746f63757272656e6379222c2261737365745f737562636c617373223a22626974636f696e227d",
			expected:    nil,
			expectError: true,
			errorMsg:    "metadata validation failed",
		},
		{
			name:        "FAIL: missing required field - asset_class",
			blob:        "7b227469636b6572223a22425443222c226e616d65223a22426974636f696e222c2261737365745f737562636c617373223a22626974636f696e227d",
			expected:    nil,
			expectError: true,
			errorMsg:    "metadata validation failed",
		},
		{
			name:        "FAIL: missing required field - asset_subclass",
			blob:        "7b227469636b6572223a22425443222c226e616d65223a22426974636f696e222c2261737365745f636c617373223a2263727970746f63757272656e6379227d",
			expected:    nil,
			expectError: true,
			errorMsg:    "metadata validation failed",
		},
		{
			name:        "FAIL: empty required field - ticker",
			blob:        "7b227469636b6572223a22222c226e616d65223a22426974636f696e222c2261737365745f636c617373223a2263727970746f63757272656e6379222c2261737365745f737562636c617373223a22626974636f696e227d",
			expected:    nil,
			expectError: true,
			errorMsg:    "metadata validation failed",
		},
		{
			name:        "FAIL: whitespace-only required field - name",
			blob:        "7b227469636b6572223a22425443222c226e616d65223a22202020222c2261737365745f636c617373223a2263727970746f63757272656e6379222c2261737365745f737562636c617373223a22626974636f696e227d",
			expected:    nil,
			expectError: true,
			errorMsg:    "metadata validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MPTokenMetadataFromBlob(tt.blob)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error message to contain %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("expected result but got nil")
				return
			}

			// Compare the result with expected
			if !compareMPTokenMetadata(result, tt.expected) {
				t.Errorf("expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			contains(s[1:], substr))))
}

// Helper function to compare MPTokenMetadata structs
func compareMPTokenMetadata(a, b *MPTokenMetadata) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	if a.Ticker != b.Ticker ||
		a.Name != b.Name ||
		a.Desc != b.Desc ||
		a.Icon != b.Icon ||
		a.AssetClass != b.AssetClass ||
		a.AssetSubclass != b.AssetSubclass ||
		a.IssuerName != b.IssuerName {
		return false
	}

	// Compare URLs
	if len(a.URLs) != len(b.URLs) {
		return false
	}
	for i, url := range a.URLs {
		if url.URL != b.URLs[i].URL ||
			url.Type != b.URLs[i].Type ||
			url.Title != b.URLs[i].Title {
			return false
		}
	}

	// Compare AdditionalInfo
	if string(a.AdditionalInfo) != string(b.AdditionalInfo) {
		return false
	}

	return true
}

func TestMPTokenMetadata_Blob(t *testing.T) {
	tests := []struct {
		name        string
		metadata    MPTokenMetadata
		expected    string
		expectError bool
		errorMsg    string
	}{
		{
			name: "PASS: valid metadata with all fields",
			metadata: MPTokenMetadata{
				Ticker:        "BTC",
				Name:          "Bitcoin",
				Desc:          "A digital currency",
				Icon:          "https://example.com/bitcoin.png",
				AssetClass:    "cryptocurrency",
				AssetSubclass: "bitcoin",
				IssuerName:    "Satoshi Nakamoto",
				URLs: []MPTokenMetadataURL{
					{
						URL:   "https://bitcoin.org",
						Type:  "text/html",
						Title: "Bitcoin Official",
					},
				},
				AdditionalInfo: json.RawMessage(`{"version":"1.0"}`),
			},
			expected:    "7b227469636b6572223a22425443222c226e616d65223a22426974636f696e222c2264657363223a2241206469676974616c2063757272656e6379222c2269636f6e223a2268747470733a2f2f6578616d706c652e636f6d2f626974636f696e2e706e67222c2261737365745f636c617373223a2263727970746f63757272656e6379222c2261737365745f737562636c617373223a22626974636f696e222c226973737565725f6e616d65223a225361746f736869204e616b616d6f746f222c2275726c73223a5b7b2275726c223a2268747470733a2f2f626974636f696e2e6f7267222c2274797065223a22746578742f68746d6c222c227469746c65223a22426974636f696e204f6666696369616c227d5d2c226164646974696f6e616c5f696e666f223a7b2276657273696f6e223a22312e30227d7d",
			expectError: false,
		},
		{
			name: "PASS: valid metadata with minimal required fields",
			metadata: MPTokenMetadata{
				Ticker:        "ETH",
				Name:          "Ethereum",
				AssetClass:    "cryptocurrency",
				AssetSubclass: "ethereum",
			},
			expected:    "7b227469636b6572223a22455448222c226e616d65223a22457468657265756d222c2261737365745f636c617373223a2263727970746f63757272656e6379222c2261737365745f737562636c617373223a22657468657265756d222c226973737565725f6e616d65223a22227d",
			expectError: false,
		},
		{
			name:        "FAIL: empty metadata - missing required fields",
			metadata:    MPTokenMetadata{},
			expected:    "",
			expectError: true,
			errorMsg:    "metadata validation failed",
		},
		{
			name: "FAIL: metadata with only ticker - missing required fields",
			metadata: MPTokenMetadata{
				Ticker: "USDT",
			},
			expected:    "",
			expectError: true,
			errorMsg:    "metadata validation failed",
		},
		{
			name: "FAIL: metadata with only name - missing required fields",
			metadata: MPTokenMetadata{
				Name: "USD Tether",
			},
			expected:    "",
			expectError: true,
			errorMsg:    "metadata validation failed",
		},
		{
			name: "PASS: metadata with urls array",
			metadata: MPTokenMetadata{
				Ticker:        "USDT",
				Name:          "USD Tether",
				AssetClass:    "stablecoin",
				AssetSubclass: "tether",
				URLs: []MPTokenMetadataURL{
					{
						URL:  "https://example.com",
						Type: "application/json",
					},
					{
						URL:   "https://docs.example.com",
						Title: "Documentation",
					},
				},
			},
			expected:    "7b227469636b6572223a2255534454222c226e616d65223a2255534420546574686572222c2261737365745f636c617373223a22737461626c65636f696e222c2261737365745f737562636c617373223a22746574686572222c226973737565725f6e616d65223a22222c2275726c73223a5b7b2275726c223a2268747470733a2f2f6578616d706c652e636f6d222c2274797065223a226170706c69636174696f6e2f6a736f6e222c227469746c65223a22227d2c7b2275726c223a2268747470733a2f2f646f63732e6578616d706c652e636f6d222c2274797065223a22222c227469746c65223a22446f63756d656e746174696f6e227d5d7d",
			expectError: false,
		},
		{
			name: "PASS: metadata with additional info",
			metadata: MPTokenMetadata{
				Ticker:         "CUSTOM",
				Name:           "Custom Token",
				AssetClass:     "custom",
				AssetSubclass:  "token",
				AdditionalInfo: json.RawMessage(`{"custom_field":"value","number":123}`),
			},
			expected:    "7b227469636b6572223a22435553544f4d222c226e616d65223a22437573746f6d20546f6b656e222c2261737365745f636c617373223a22637573746f6d222c2261737365745f737562636c617373223a22746f6b656e222c226973737565725f6e616d65223a22222c226164646974696f6e616c5f696e666f223a7b22637573746f6d5f6669656c64223a2276616c7565222c226e756d626572223a3132337d7d",
			expectError: false,
		},
		{
			name: "PASS: metadata with special characters",
			metadata: MPTokenMetadata{
				Ticker:        "TØKËN",
				Name:          "Tøkën with spéciál chârs",
				AssetClass:    "special",
				AssetSubclass: "unicode",
				Desc:          "Description with émojis 🚀 and symbols & < > \" '",
			},
			expected:    "7b227469636b6572223a2254c3984bc38b4e222c226e616d65223a2254c3b86bc3ab6e2077697468207370c3a96369c3a16c206368c3a27273222c2264657363223a224465736372697074696f6e207769746820c3a96d6f6a697320f09f9a8020616e642073796d626f6c73205c7530303236205c7530303363205c7530303365205c222027222c2261737365745f636c617373223a227370656369616c222c2261737365745f737562636c617373223a22756e69636f6465222c226973737565725f6e616d65223a22227d",
			expectError: false,
		},
		{
			name: "PASS: metadata with empty urls array",
			metadata: MPTokenMetadata{
				Ticker:        "EMPTY",
				Name:          "Empty URLs",
				AssetClass:    "test",
				AssetSubclass: "empty",
				URLs:          []MPTokenMetadataURL{},
			},
			expected:    "7b227469636b6572223a22454d505459222c226e616d65223a22456d7074792055524c73222c2261737365745f636c617373223a2274657374222c2261737365745f737562636c617373223a22656d707479222c226973737565725f6e616d65223a22227d",
			expectError: false,
		},
		{
			name: "PASS: metadata with empty additional info",
			metadata: MPTokenMetadata{
				Ticker:         "EMPTY_INFO",
				Name:           "Empty Additional Info",
				AssetClass:     "test",
				AssetSubclass:  "empty",
				AdditionalInfo: json.RawMessage(`{}`),
			},
			expected:    "7b227469636b6572223a22454d5054595f494e464f222c226e616d65223a22456d707479204164646974696f6e616c20496e666f222c2261737365745f636c617373223a2274657374222c2261737365745f737562636c617373223a22656d707479222c226973737565725f6e616d65223a22222c226164646974696f6e616c5f696e666f223a7b7d7d",
			expectError: false,
		},
		{
			name: "FAIL: metadata exceeding max size",
			metadata: MPTokenMetadata{
				Ticker:        "LARGE",
				Name:          "Large Token",
				AssetClass:    "test",
				AssetSubclass: "large",
				Desc:          generateLargeString(2000), // Generate a string larger than MPTokenMetadataMaxSize
			},
			expected:    "",
			expectError: true,
			errorMsg:    "blob is too large",
		},
		{
			name: "FAIL: missing required field - ticker",
			metadata: MPTokenMetadata{
				Name:          "Test Token",
				AssetClass:    "test",
				AssetSubclass: "token",
			},
			expected:    "",
			expectError: true,
			errorMsg:    "metadata validation failed",
		},
		{
			name: "FAIL: missing required field - name",
			metadata: MPTokenMetadata{
				Ticker:        "TEST",
				AssetClass:    "test",
				AssetSubclass: "token",
			},
			expected:    "",
			expectError: true,
			errorMsg:    "metadata validation failed",
		},
		{
			name: "FAIL: missing required field - asset_class",
			metadata: MPTokenMetadata{
				Ticker:        "TEST",
				Name:          "Test Token",
				AssetSubclass: "token",
			},
			expected:    "",
			expectError: true,
			errorMsg:    "metadata validation failed",
		},
		{
			name: "FAIL: missing required field - asset_subclass",
			metadata: MPTokenMetadata{
				Ticker:     "TEST",
				Name:       "Test Token",
				AssetClass: "test",
			},
			expected:    "",
			expectError: true,
			errorMsg:    "metadata validation failed",
		},
		{
			name: "FAIL: empty required field - ticker",
			metadata: MPTokenMetadata{
				Ticker:        "",
				Name:          "Test Token",
				AssetClass:    "test",
				AssetSubclass: "token",
			},
			expected:    "",
			expectError: true,
			errorMsg:    "metadata validation failed",
		},
		{
			name: "FAIL: whitespace-only required field - name",
			metadata: MPTokenMetadata{
				Ticker:        "TEST",
				Name:          "   ",
				AssetClass:    "test",
				AssetSubclass: "token",
			},
			expected:    "",
			expectError: true,
			errorMsg:    "metadata validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.metadata.Blob()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error message to contain %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// For special characters test, we validate dynamically since encoding can vary
			if tt.name == "metadata with special characters" {
				// Just verify it's not empty and can be decoded
				if result == "" {
					t.Errorf("expected non-empty result for special characters test")
				}
			} else if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}

			// Verify that the blob can be decoded back to the original metadata
			decoded, err := MPTokenMetadataFromBlob(result)
			if err != nil {
				t.Errorf("failed to decode generated blob: %v", err)
				return
			}

			if !compareMPTokenMetadata(decoded, &tt.metadata) {
				t.Errorf("decoded metadata doesn't match original. Original: %+v, Decoded: %+v", tt.metadata, decoded)
			}
		})
	}
}

// Helper function to generate a large string for testing size limits
func generateLargeString(size int) string {
	result := make([]byte, size)
	for i := range result {
		result[i] = 'A' + byte(i%26)
	}
	return string(result)
}
