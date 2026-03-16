package websocket

import (
	"testing"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/common"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/stretchr/testify/require"
)

func TestNewClientConfig(t *testing.T) {
	config := NewClientConfig()
	require.Equal(t, common.DefaultMaxRetries, config.maxRetries)
	require.Equal(t, common.DefaultRetryDelay, config.retryDelay)
	require.Equal(t, common.DefaultHost, config.host)
	require.Equal(t, config.feeCushion, common.DefaultFeeCushion)
	require.Equal(t, config.maxFeeXRP, common.DefaultMaxFeeXRP)
	require.Equal(t, common.DefaultTimeout, config.timeout)
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
	require.Equal(t, config.feeCushion, float32(1.5))
}

func TestWithMaxFeeXRP(t *testing.T) {
	config := NewClientConfig().WithMaxFeeXRP(3.0)
	require.Equal(t, config.maxFeeXRP, float32(3.0))
}

func TestWithFaucetProvider(t *testing.T) {
	config := NewClientConfig().WithFaucetProvider(faucet.NewTestnetFaucetProvider())
	require.NotNil(t, config.faucetProvider)
}

func TestWithTimeout(t *testing.T) {
	config := NewClientConfig().WithTimeout(10 * time.Second)
	require.Equal(t, 10*time.Second, config.timeout)
}
