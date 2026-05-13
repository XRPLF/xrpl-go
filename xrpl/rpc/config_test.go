package rpc

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/common"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/internal/clientconfig"
	clientconfigtestutil "github.com/Peersyst/xrpl-go/xrpl/internal/clientconfig/testutil"
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

func TestSetLogger(t *testing.T) {
	t.Run("nil silences warnings", func(t *testing.T) {
		logs := clientconfigtestutil.CaptureLogOutput(t, func() {
			SetLogger(nil)
			cfg, err := NewClientConfig("http://s1.ripple.com:51234")
			require.NoError(t, err)
			require.NotNil(t, cfg)
		})
		require.Empty(t, logs)
	})

	t.Run("custom logger receives warnings", func(t *testing.T) {
		var buf bytes.Buffer
		previous := clientconfig.SetLogger(log.New(&buf, "custom:", 0))
		t.Cleanup(func() { clientconfig.SetLogger(previous) })

		cfg, err := NewClientConfig("http://s1.ripple.com:51234")
		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.Contains(t, buf.String(), "custom:")
		require.Contains(t, buf.String(), "is not using a TLS scheme")
	})
}

// Race-detector test, success = no race report.
func TestSetLoggerConcurrentWarning(t *testing.T) {
	previous := clientconfig.SetLogger(log.Default())
	t.Cleanup(func() { clientconfig.SetLogger(previous) })

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				SetLogger(nil)
				return
			}
			SetLogger(log.New(io.Discard, "", 0))
		}(i)
		go func() {
			defer wg.Done()
			_, _ = NewClientConfig("http://s1.ripple.com:51234")
		}()
	}
	wg.Wait()
}

func TestNewClientConfigInsecureSchemeWarnings(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantWarning string
	}{
		{
			name:        "remote insecure scheme warns",
			url:         "http://s1.ripple.com:51234",
			wantWarning: `xrpl-go: warning: rpc client endpoint "http://s1.ripple.com:51234/" is not using a TLS scheme`,
		},
		{
			name: "local insecure scheme does not warn",
			url:  "http://127.0.0.1:51234",
		},
		{
			name: "remote tls scheme does not warn",
			url:  "https://s1.ripple.com:51234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs := clientconfigtestutil.CaptureLogOutput(t, func() {
				cfg, err := NewClientConfig(tt.url)

				require.NoError(t, err)
				require.NotNil(t, cfg)
			})

			if tt.wantWarning == "" {
				require.Empty(t, logs)
				return
			}

			require.Contains(t, logs, tt.wantWarning)
		})
	}
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
