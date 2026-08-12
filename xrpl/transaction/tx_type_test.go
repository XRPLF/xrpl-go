package transaction

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsConfidentialMPTTransactionType(t *testing.T) {
	tests := []struct {
		name     string
		txType   TxType
		expected bool
	}{
		{name: "Clawback", txType: ConfidentialMPTClawbackTx, expected: true},
		{name: "Convert", txType: ConfidentialMPTConvertTx, expected: true},
		{name: "ConvertBack", txType: ConfidentialMPTConvertBackTx, expected: true},
		{name: "MergeInbox", txType: ConfidentialMPTMergeInboxTx, expected: true},
		{name: "Send", txType: ConfidentialMPTSendTx, expected: true},
		{name: "ordinary transaction", txType: PaymentTx, expected: false},
		{name: "unknown transaction", txType: TxType("FutureTransaction"), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, IsConfidentialMPTTransactionType(tt.txType))
		})
	}
}

func TestIsPseudoTransactionType(t *testing.T) {
	tests := []struct {
		name     string
		txType   TxType
		expected bool
	}{
		{name: "EnableAmendment", txType: EnableAmendmentTx, expected: true},
		{name: "SetFee", txType: SetFeeTx, expected: true},
		{name: "UNLModify", txType: UNLModifyTx, expected: true},
		{name: "user transaction", txType: PaymentTx, expected: false},
		{name: "unknown transaction", txType: TxType("FutureTransaction"), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, IsPseudoTransactionType(tt.txType))
		})
	}
}
