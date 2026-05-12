package websocket

import (
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/common"
	"github.com/Peersyst/xrpl-go/xrpl/internal/clientconfig"
)

// ClientConfig configures options for the XRPL WebSocket client.
type ClientConfig struct {
	// Connection config
	host          string
	maxRetries    int
	maxReconnects int
	retryDelay    time.Duration
	timeout       time.Duration

	// Fee config
	feeCushion float32
	maxFeeXRP  float32

	// Faucet config
	faucetProvider common.FaucetProvider
}

// NewClientConfig returns a ClientConfig initialized with default settings.
func NewClientConfig() *ClientConfig {
	return &ClientConfig{
		host:          common.DefaultHost,
		feeCushion:    common.DefaultFeeCushion,
		maxFeeXRP:     common.DefaultMaxFeeXRP,
		maxRetries:    common.DefaultMaxRetries,
		maxReconnects: common.DefaultMaxReconnects,
		retryDelay:    common.DefaultRetryDelay,
		timeout:       common.DefaultTimeout,
	}
}

// WithHost sets the host of the websocket client.
// Default: "localhost"
func (wc ClientConfig) WithHost(host string) ClientConfig {
	wc.host = host
	clientconfig.WarnIfInsecureScheme("websocket", wc.host)
	return wc
}

// WithFeeCushion sets the fee cushion of the websocket client.
// Default: 1.2
func (wc ClientConfig) WithFeeCushion(feeCushion float32) ClientConfig {
	wc.feeCushion = feeCushion
	return wc
}

// WithMaxFeeXRP sets the maximum fee in XRP that the websocket client will use.
// Default: 2
func (wc ClientConfig) WithMaxFeeXRP(maxFeeXrp float32) ClientConfig {
	wc.maxFeeXRP = maxFeeXrp
	return wc
}

// WithFaucetProvider sets the faucet provider of the websocket client.
// Default: faucet.NewLocalFaucetProvider()
func (wc ClientConfig) WithFaucetProvider(fp common.FaucetProvider) ClientConfig {
	wc.faucetProvider = fp
	return wc
}

// WithMaxRetries sets the maximum number of retries for a transaction.
// Default: 10
func (wc ClientConfig) WithMaxRetries(maxRetries int) ClientConfig {
	wc.maxRetries = maxRetries
	return wc
}

// WithMaxReconnects sets the maximum number of reconnects for a transaction.
// Default: 3
func (wc ClientConfig) WithMaxReconnects(maxReconnects int) ClientConfig {
	wc.maxReconnects = maxReconnects
	return wc
}

// WithRetryDelay sets the delay between retries for a transaction.
// Default: 1 second
func (wc ClientConfig) WithRetryDelay(retryDelay time.Duration) ClientConfig {
	wc.retryDelay = retryDelay
	return wc
}

// WithTimeout sets the timeout for a request.
// Default: 10 seconds
func (wc ClientConfig) WithTimeout(timeout time.Duration) ClientConfig {
	wc.timeout = timeout
	return wc
}
