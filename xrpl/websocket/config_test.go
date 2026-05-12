package websocket

import (
	"bytes"
	"log"
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
	require.InEpsilon(t, common.DefaultFeeCushion, config.feeCushion, 0)
	require.InEpsilon(t, common.DefaultMaxFeeXRP, config.maxFeeXRP, 0)
	require.Equal(t, common.DefaultTimeout, config.timeout)
}

func TestWithHostInsecureSchemeWarnings(t *testing.T) {
	tests := []struct {
		name        string
		host        string
		wantWarning string
	}{
		{
			name:        "remote insecure scheme warns",
			host:        "ws://s1.ripple.com:6006",
			wantWarning: `xrpl-go: warning: websocket client endpoint "ws://s1.ripple.com:6006" uses non-TLS scheme "ws"`,
		},
		{
			name: "local insecure scheme does not warn",
			host: "ws://localhost:6006",
		},
		{
			name: "remote tls scheme does not warn",
			host: "wss://s1.ripple.com:6006",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs := captureLogOutput(t, func() {
				config := NewClientConfig().WithHost(tt.host)

				require.Equal(t, tt.host, config.host)
			})

			if tt.wantWarning == "" {
				require.Empty(t, logs)
				return
			}

			require.Contains(t, logs, tt.wantWarning)
		})
	}
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

func captureLogOutput(t *testing.T, fn func()) string {
	t.Helper()

	var buf bytes.Buffer
	previousOutput := log.Writer()
	previousFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	defer func() {
		log.SetOutput(previousOutput)
		log.SetFlags(previousFlags)
	}()

	fn()

	return buf.String()
}
