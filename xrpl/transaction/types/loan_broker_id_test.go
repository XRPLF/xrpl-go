package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoanBrokerID_Value(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "empty string",
			value: "",
			want:  "",
		},
		{
			name:  "valid 64-char hex string",
			value: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
			want:  "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
		},
		{
			name:  "invalid length string",
			value: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F43",
			want:  "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F43",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LoanBrokerID(tt.value)
			require.Equal(t, tt.want, (&result).Value())
		})
	}
}
