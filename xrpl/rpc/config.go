package rpc

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/common"
	"github.com/Peersyst/xrpl-go/xrpl/internal/clientconfig"
)

const defaultMaxResponseSize int64 = 64 * 1024 * 1024

// SetLogger overrides the *log.Logger used for SDK-emitted warnings (currently
// just the insecure-scheme warning). Pass nil to silence the warnings entirely.
// The default logger writes to stdlib's log.Default(), preserving prior behavior.
// The logger is shared across xrpl-go's client packages; calling SetLogger here
// or in xrpl/websocket has the same effect.
func SetLogger(l *log.Logger) {
	clientconfig.SetLogger(l)
}

// HTTPClient defines the interface for sending HTTP requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Config holds configuration for the XRPL RPC client, including HTTP client, URL, headers, and retry/fee settings.
type Config struct {
	HTTPClient HTTPClient
	URL        string
	Headers    map[string][]string

	// Reliable-submission monitoring config.
	maxRetries int
	retryDelay time.Duration

	// Response body config
	maxResponseSize int64

	// Fee config
	maxFeeXRP  string
	feeCushion float64

	// Faucet config
	faucetProvider common.FaucetProvider

	// Trusted network identity override.
	networkID    *uint32
	buildVersion string

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

// WithMaxRetries limits consecutive incomplete reliable-submission polling
// rounds caused by query or transport errors. It does not limit successful
// finality polling.
func WithMaxRetries(maxRetries int) ConfigOpt {
	return func(c *Config) {
		c.maxRetries = maxRetries
	}
}

// WithRetryDelay sets the delay between reliable-submission polling rounds.
func WithRetryDelay(retryDelay time.Duration) ConfigOpt {
	return func(c *Config) {
		c.retryDelay = retryDelay
	}
}

// WithMaxResponseSize returns a ConfigOpt that sets the maximum response body size.
// Set to 0 to disable the response size limit.
// Negative values are replaced with the default.
func WithMaxResponseSize(maxResponseSize int64) ConfigOpt {
	return func(c *Config) {
		if maxResponseSize < 0 {
			c.maxResponseSize = defaultMaxResponseSize
			return
		}
		c.maxResponseSize = maxResponseSize
	}
}

// WithMaxFeeXRP returns a ConfigOpt that sets the maximum fee in XRP as a decimal string.
func WithMaxFeeXRP(maxFeeXRP string) ConfigOpt {
	return func(c *Config) {
		c.maxFeeXRP = maxFeeXRP
	}
}

// WithFeeCushion returns a ConfigOpt that sets the fee cushion multiplier.
func WithFeeCushion(feeCushion float64) ConfigOpt {
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

// WithNetworkIdentity configures a trusted network identity and bypasses
// server_info discovery. Use this only when both values come from a trusted
// deployment configuration.
func WithNetworkIdentity(networkID uint32, buildVersion string) ConfigOpt {
	return func(c *Config) {
		c.networkID = &networkID
		c.buildVersion = buildVersion
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

		maxRetries:      common.DefaultMaxRetries,
		retryDelay:      common.DefaultRetryDelay,
		maxResponseSize: defaultMaxResponseSize,
		maxFeeXRP:       common.DefaultMaxFeeXRP,
		feeCushion:      common.DefaultFeeCushion,
		timeout:         common.DefaultTimeout,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	clientconfig.WarnIfInsecureScheme("rpc", cfg.URL)

	// Keep the default HTTP client aligned with the config timeout.
	// If the HTTP client has a custom timeout, sync it to the config to prevent divergence.
	// Otherwise, apply the config timeout to the HTTP client.
	if hc, ok := cfg.HTTPClient.(*http.Client); ok {
		if hc.Timeout == 0 {
			hc.Timeout = cfg.timeout
		} else {
			cfg.timeout = hc.Timeout
		}
	}

	return cfg, nil
}
