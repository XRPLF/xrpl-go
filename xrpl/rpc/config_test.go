package rpc

import (
	"net/http"
	"testing"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/common"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type customHttpClient struct{}

func (c customHttpClient) Do(req *http.Request) (*http.Response, error) {
	return nil, nil
}

func TestConfigCreation(t *testing.T) {
	t.Run("Set config with valid port + ip", func(t *testing.T) {
		cfg, _ := NewClientConfig("http://s1.ripple.com:51234/")

		req, err := http.NewRequest(http.MethodPost, "http://s1.ripple.com:51234/", nil)

		req.Header = cfg.Headers
		assert.Equal(t, "http://s1.ripple.com:51234/", cfg.URL)
		require.NoError(t, err)
	})
	t.Run("No port + IP provided", func(t *testing.T) {
		cfg, err := NewClientConfig("")

		assert.Nil(t, cfg)
		assert.EqualError(t, err, "empty port and IP provided")
	})
	t.Run("Format root path - add /", func(t *testing.T) {
		cfg, _ := NewClientConfig("http://s1.ripple.com:51234")

		req, err := http.NewRequest(http.MethodPost, "http://s1.ripple.com:51234/", nil)

		req.Header = cfg.Headers
		assert.Equal(t, "http://s1.ripple.com:51234/", cfg.URL)
		require.NoError(t, err)
	})
	t.Run("Pass in custom HTTP client", func(t *testing.T) {
		c := customHttpClient{}
		cfg, _ := NewClientConfig("http://s1.ripple.com:51234", WithHTTPClient(c))

		req, err := http.NewRequest(http.MethodPost, "http://s1.ripple.com:51234/", nil)
		headers := map[string][]string{
			"Content-Type": {"application/json"},
		}
		req.Header = cfg.Headers
		assert.Equal(t, &Config{HTTPClient: customHttpClient{}, URL: "http://s1.ripple.com:51234/", Headers: headers, maxRetries: common.DefaultMaxRetries, retryDelay: common.DefaultRetryDelay, feeCushion: common.DefaultFeeCushion, maxFeeXRP: common.DefaultMaxFeeXRP, faucetProvider: nil, timeout: common.DefaultTimeout}, cfg)
		assert.NoError(t, err)
	})
}

func TestWithMaxFeeXRP(t *testing.T) {
	maxFee := float32(5.0)
	cfg, _ := NewClientConfig("http://s1.ripple.com:51234", WithMaxFeeXRP(maxFee))

	require.InEpsilon(t, maxFee, cfg.maxFeeXRP, 0)
}

func TestWithFeeCushion(t *testing.T) {
	feeCushion := float32(1.5)
	cfg, _ := NewClientConfig("http://s1.ripple.com:51234", WithFeeCushion(feeCushion))

	require.InEpsilon(t, feeCushion, cfg.feeCushion, 0)
}

func TestWithFaucetProvider(t *testing.T) {
	fp := faucet.NewTestnetFaucetProvider()
	cfg, _ := NewClientConfig("http://s1.ripple.com:51234", WithFaucetProvider(fp))

	require.Equal(t, fp, cfg.faucetProvider)
}

func TestTimeout(t *testing.T) {
	t.Run("Default timeout applied to config and HTTP client", func(t *testing.T) {
		cfg, _ := NewClientConfig("http://s1.ripple.com:51234")

		require.Equal(t, common.DefaultTimeout, cfg.timeout)
		hc, ok := cfg.HTTPClient.(*http.Client)
		require.True(t, ok)
		require.Equal(t, common.DefaultTimeout, hc.Timeout)
	})

	t.Run("WithTimeout sets config and HTTP client timeout", func(t *testing.T) {
		timeOut := 11 * time.Second
		cfg, _ := NewClientConfig("http://s1.ripple.com:51234", WithTimeout(timeOut))

		require.Equal(t, timeOut, cfg.timeout)
		hc, ok := cfg.HTTPClient.(*http.Client)
		require.True(t, ok)
		require.Equal(t, timeOut, hc.Timeout)
	})

	t.Run("Custom HTTP client with timeout syncs to config", func(t *testing.T) {
		customTimeout := 45 * time.Second
		customClient := &http.Client{Timeout: customTimeout}
		cfg, _ := NewClientConfig("http://s1.ripple.com:51234", WithHTTPClient(customClient))

		require.Equal(t, customTimeout, cfg.timeout)
		require.Equal(t, customTimeout, customClient.Timeout)
	})

	t.Run("Custom HTTP client without timeout gets default applied", func(t *testing.T) {
		customClient := &http.Client{}
		cfg, _ := NewClientConfig("http://s1.ripple.com:51234", WithHTTPClient(customClient))

		require.Equal(t, common.DefaultTimeout, cfg.timeout)
		require.Equal(t, common.DefaultTimeout, customClient.Timeout)
	})

	t.Run("WithTimeout overrides custom HTTP client timeout", func(t *testing.T) {
		customClient := &http.Client{Timeout: 45 * time.Second}
		explicitTimeout := 10 * time.Second
		cfg, _ := NewClientConfig("http://s1.ripple.com:51234", WithHTTPClient(customClient), WithTimeout(explicitTimeout))

		require.Equal(t, explicitTimeout, cfg.timeout)
		require.Equal(t, explicitTimeout, customClient.Timeout)
	})

	t.Run("Non-standard HTTP client uses default config timeout", func(t *testing.T) {
		c := customHttpClient{}
		cfg, _ := NewClientConfig("http://s1.ripple.com:51234", WithHTTPClient(c))

		require.Equal(t, common.DefaultTimeout, cfg.timeout)
	})
}

func TestWithMaxRetries(t *testing.T) {
	maxRetries := 5
	cfg, _ := NewClientConfig("http://s1.ripple.com:51234", WithMaxRetries(maxRetries))

	require.Equal(t, maxRetries, cfg.maxRetries)
}

func TestWithRetryDelay(t *testing.T) {
	retryDelay := 2 * time.Second
	cfg, _ := NewClientConfig("http://s1.ripple.com:51234", WithRetryDelay(retryDelay))

	require.Equal(t, retryDelay, cfg.retryDelay)
}
