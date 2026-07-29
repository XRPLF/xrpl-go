package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRunnerAppliesDefaultsToPartialConfig(t *testing.T) {
	runner := NewRunner(t, nil, &RunnerConfig{WalletCount: 2})

	require.Equal(t, 2, runner.config.WalletCount)
	require.Equal(t, DefaultMaxRetries, runner.config.MaxRetries)
}
