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
		assert.NoError(t, err)
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
		assert.NoError(t, err)
	})
	t.Run("Pass in custom HTTP client", func(t *testing.T) {

		c := customHttpClient{}
		cfg, _ := NewClientConfig("http://s1.ripple.com:51234", WithHTTPClient(c))

		req, err := http.NewRequest(http.MethodPost, "http://s1.ripple.com:51234/", nil)
		headers := map[string][]string{
			"Content-Type": {"application/json"},
		}
		req.Header = cfg.Headers
		assert.Equal(t, &Config{HTTPClient: customHttpClient{}, URL: "http://s1.ripple.com:51234/", Headers: headers, maxRetries: common.DefaultMaxRetries, retryDelay: common.DefaultRetryDelay, feeCushion: common.DefaultFeeCushion, maxFeeXRP: common.DefaultMaxFeeXRP, faucetProvider: nil}, cfg)
		assert.NoError(t, err)
	})
}

func TestWithMaxFeeXRP(t *testing.T) {
	maxFee := float32(5.0)
	cfg, _ := NewClientConfig("http://s1.ripple.com:51234", WithMaxFeeXRP(maxFee))

	require.Equal(t, maxFee, cfg.maxFeeXRP)
}

func TestWithFeeCushion(t *testing.T) {
	feeCushion := float32(1.5)
	cfg, _ := NewClientConfig("http://s1.ripple.com:51234", WithFeeCushion(feeCushion))

	require.Equal(t, feeCushion, cfg.feeCushion)
}

func TestWithFaucetProvider(t *testing.T) {
	fp := faucet.NewTestnetFaucetProvider()
	cfg, _ := NewClientConfig("http://s1.ripple.com:51234", WithFaucetProvider(fp))

	require.Equal(t, fp, cfg.faucetProvider)
}

func TestWithTimeout(t *testing.T) {
	timeOut := 11 * time.Second // 11 seconds
	cfg, _ := NewClientConfig("http://s1.ripple.com:51234", WithTimeout(timeOut))

	require.Equal(t, timeOut, cfg.timeout)
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
