package websocket

import (
	"testing"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/common"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	clientconfigtestutil "github.com/Peersyst/xrpl-go/xrpl/internal/clientconfig/testutil"
	"github.com/stretchr/testify/require"
)

func TestNewClientConfig(t *testing.T) {
	config := NewClientConfig()
	require.Equal(t, common.DefaultMaxRetries, config.maxRetries)
	require.Equal(t, common.DefaultRetryDelay, config.retryDelay)
	require.Equal(t, common.DefaultHost, config.host)
	require.InEpsilon(t, common.DefaultFeeCushion, config.feeCushion, 0)
	require.InEpsilon(t, common.DefaultMaxFeeXRP, config.maxFeeXRP, 0)
	require.Equal(t, common.DefaultTimeout, config.timeout)
	require.Equal(t, defaultMaxResponseSize, config.maxResponseSize)
}

func TestWithHostDoesNotWarn(t *testing.T) {
	// WithHost is a fluent setter and may be called multiple times during a builder
	// chain; the insecure-scheme warning is intentionally deferred to NewClient so
	// it fires exactly once per client.
	logs := clientconfigtestutil.CaptureLogOutput(t, func() {
		_ = NewClientConfig().WithHost("ws://s1.ripple.com:6006").WithHost("wss://s2.ripple.com:6006")
	})
	require.Empty(t, logs)
}

func TestWithMaxRetries(t *testing.T) {
	config := NewClientConfig().WithMaxRetries(20)
	require.Equal(t, 20, config.maxRetries)
}

func TestWithRetryDelay(t *testing.T) {
	config := NewClientConfig().WithRetryDelay(2 * time.Second)
	require.Equal(t, 2*time.Second, config.retryDelay)
}

func TestWithFeeCushion(t *testing.T) {
	config := NewClientConfig().WithFeeCushion(1.5)
	require.InEpsilon(t, float32(1.5), config.feeCushion, 0)
}

func TestWithMaxFeeXRP(t *testing.T) {
	config := NewClientConfig().WithMaxFeeXRP(3.0)
	require.InEpsilon(t, float32(3.0), config.maxFeeXRP, 0)
}

func TestWithFaucetProvider(t *testing.T) {
	config := NewClientConfig().WithFaucetProvider(faucet.NewTestnetFaucetProvider())
	require.NotNil(t, config.faucetProvider)
}

func TestWithTimeout(t *testing.T) {
	config := NewClientConfig().WithTimeout(10 * time.Second)
	require.Equal(t, 10*time.Second, config.timeout)
}

func TestWithMaxResponseSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected int64
	}{
		{
			name:     "override max response size",
			size:     32,
			expected: 32,
		},
		{
			name:     "zero max response size disables limit",
			size:     0,
			expected: 0,
		},
		{
			name:     "negative max response size uses default",
			size:     -1,
			expected: defaultMaxResponseSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewClientConfig().WithMaxResponseSize(tt.size)

			require.Equal(t, tt.expected, config.maxResponseSize)
		})
	}
}
