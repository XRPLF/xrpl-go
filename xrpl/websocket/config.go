package websocket

import (
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/common"
)

const defaultMaxResponseSize int64 = 16 * 1024 * 1024

// ClientConfig configures options for the XRPL WebSocket client.
type ClientConfig struct {
	// Connection config
	host            string
	maxRetries      int
	maxReconnects   int
	retryDelay      time.Duration
	timeout         time.Duration
	maxResponseSize int64

	// Fee config
	feeCushion float32
	maxFeeXRP  float32

	// Faucet config
	faucetProvider common.FaucetProvider
}

// NewClientConfig returns a ClientConfig initialized with default settings.
func NewClientConfig() *ClientConfig {
	return &ClientConfig{
		host:            common.DefaultHost,
		feeCushion:      common.DefaultFeeCushion,
		maxFeeXRP:       common.DefaultMaxFeeXRP,
		maxRetries:      common.DefaultMaxRetries,
		maxReconnects:   common.DefaultMaxReconnects,
		retryDelay:      common.DefaultRetryDelay,
		timeout:         common.DefaultTimeout,
		maxResponseSize: defaultMaxResponseSize,
	}
}

// WithHost sets the host of the websocket client.
// Default: "localhost"
func (wc ClientConfig) WithHost(host string) ClientConfig {
	wc.host = host
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

// WithMaxResponseSize sets the maximum inbound WebSocket response message size.
// Set to 0 to disable the response size limit.
func (wc ClientConfig) WithMaxResponseSize(maxResponseSize int64) ClientConfig {
	if maxResponseSize < 0 {
		wc.maxResponseSize = defaultMaxResponseSize
		return wc
	}
	wc.maxResponseSize = maxResponseSize
	return wc
}
