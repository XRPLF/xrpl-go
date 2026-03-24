package types

import (
	"reflect"
	"testing"
)

func TestCredential_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		input    Credential
		expected any
	}{
		{
			name: "pass - empty credential",
			input: Credential{
				CredentialType: CredentialType(""),
				Issuer:         "",
			},
			expected: map[string]string{
				"Issuer":         "",
				"CredentialType": "",
			},
		},
		{
			name: "pass - valid credential",
			input: Credential{
				CredentialType: CredentialType("0123"),
				Issuer:         "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
			},
			expected: map[string]string{
				"Issuer":         "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
				"CredentialType": "0123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Flatten()
			if reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Flatten() = %v, want %v", result, tt.expected)
			}
		})
	}
}
