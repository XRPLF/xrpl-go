package rpc

import (
	"net/http"
	"strings"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/common"
)

// HTTPClient defines the interface for sending HTTP requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Config holds configuration for the XRPL RPC client, including HTTP client, URL, headers, and retry/fee settings.
type Config struct {
	HTTPClient HTTPClient
	URL        string
	Headers    map[string][]string

	// Retry config
	maxRetries int
	retryDelay time.Duration

	// Fee config
	maxFeeXRP  float32
	feeCushion float32

	// Faucet config
	faucetProvider common.FaucetProvider

	timeout time.Duration
}

// ConfigOpt represents a function that applies a configuration option to Config.
type ConfigOpt func(c *Config)

// WithHTTPClient returns a ConfigOpt that sets a custom HTTPClient.
func WithHTTPClient(cl HTTPClient) ConfigOpt {
	return func(c *Config) {
		c.HTTPClient = cl
	}
}

// WithMaxRetries returns a ConfigOpt that sets the maximum number of retries.
func WithMaxRetries(maxRetries int) ConfigOpt {
	return func(c *Config) {
		c.maxRetries = maxRetries
	}
}

// WithRetryDelay returns a ConfigOpt that sets the delay between retry attempts.
func WithRetryDelay(retryDelay time.Duration) ConfigOpt {
	return func(c *Config) {
		c.retryDelay = retryDelay
	}
}

// WithMaxFeeXRP returns a ConfigOpt that sets the maximum fee in XRP.
func WithMaxFeeXRP(maxFeeXRP float32) ConfigOpt {
	return func(c *Config) {
		c.maxFeeXRP = maxFeeXRP
	}
}

// WithFeeCushion returns a ConfigOpt that sets the fee cushion multiplier.
func WithFeeCushion(feeCushion float32) ConfigOpt {
	return func(c *Config) {
		c.feeCushion = feeCushion
	}
}

// WithFaucetProvider returns a ConfigOpt that sets the faucet provider.
func WithFaucetProvider(fp common.FaucetProvider) ConfigOpt {
	return func(c *Config) {
		c.faucetProvider = fp
	}
}

// WithTimeout returns a ConfigOpt that sets the request timeout for the HTTP client.
func WithTimeout(timeout time.Duration) ConfigOpt {
	return func(c *Config) {
		c.timeout = timeout
		if hc, ok := c.HTTPClient.(*http.Client); ok {
			hc.Timeout = timeout
		}
	}
}

// NewClientConfig creates a new Config with the given URL and applies any provided ConfigOpt options.
func NewClientConfig(url string, opts ...ConfigOpt) (*Config, error) {

	// validate a url has been passed in
	if len(url) == 0 {
		return nil, ErrEmptyURL
	}
	// add slash if doesn't already end with one
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}

	cfg := &Config{
		HTTPClient: &http.Client{},
		URL:        url,
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},

		maxRetries: common.DefaultMaxRetries,
		retryDelay: common.DefaultRetryDelay,

		maxFeeXRP:  common.DefaultMaxFeeXRP,
		feeCushion: common.DefaultFeeCushion,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	// Ensure the HTTPClient has the correct timeout if user did not set one
	if hc, ok := cfg.HTTPClient.(*http.Client); ok && cfg.timeout == 0 {
		hc.Timeout = common.DefaultTimeout
	}

	return cfg, nil
}
