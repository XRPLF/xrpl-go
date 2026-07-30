package rpc

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/internal/clientconfig"
	account "github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/stretchr/testify/require"
)

var (
	testHeaderValue = strings.Repeat("x", 24)
	testURLUsername = strings.Repeat("u", 12)
	testURLPassword = strings.Repeat("p", 12)
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type opaqueHTTPClient struct {
	called *atomic.Bool
}

func (c opaqueHTTPClient) Do(*http.Request) (*http.Response, error) {
	c.called.Store(true)
	return nil, nil
}

type credentialEchoError string

func (e credentialEchoError) Error() string {
	return string(e)
}

func TestNewClientConfigAuthorizationTransport(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		headerName string
		httpClient HTTPClient
		wantErr    bool
	}{
		{
			name:       "rejects plaintext HTTP",
			url:        "http://node.example",
			headerName: "Authorization",
			wantErr:    true,
		},
		{
			name:       "matches lowercase header",
			url:        "http://node.example",
			headerName: "authorization",
			wantErr:    true,
		},
		{
			name:       "matches mixed-case header",
			url:        "http://node.example",
			headerName: "aUtHoRiZaTiOn",
			wantErr:    true,
		},
		{
			name:       "accepts HTTPS",
			url:        "https://node.example",
			headerName: "Authorization",
		},
		{
			name:       "accepts case-insensitive HTTPS scheme",
			url:        "HTTPS://node.example",
			headerName: "Authorization",
		},
		{
			name:       "rejects scheme prefix lookalike",
			url:        "https.not-secure://node.example",
			headerName: "Authorization",
			wantErr:    true,
		},
		{
			name:       "rejects malformed URL",
			url:        "https://[::1",
			headerName: "Authorization",
			wantErr:    true,
		},
		{
			name:    "rejects plaintext URL userinfo",
			url:     testUserinfoURL("http", "node.example"),
			wantErr: true,
		},
		{
			name: "accepts HTTPS URL userinfo",
			url:  testUserinfoURL("https", "node.example"),
		},
		{
			name:       "rejects opaque HTTP client",
			url:        "https://node.example",
			headerName: "Authorization",
			httpClient: opaqueHTTPClient{called: &atomic.Bool{}},
			wantErr:    true,
		},
		{
			name: "allows plaintext without authorization",
			url:  "http://node.example",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts []ConfigOpt
			if tt.headerName != "" {
				opts = append(opts, withTestAuthorizationHeader(tt.headerName))
			}
			if tt.httpClient != nil {
				opts = append(opts, WithHTTPClient(tt.httpClient))
			}

			cfg, err := NewClientConfig(tt.url, opts...)
			assertAuthorizationErrorRedacted(t, err)
			if tt.wantErr {
				if !errors.Is(err, ErrInsecureAuthorization) {
					t.Fatal("expected insecure authorization error")
				}
				require.Nil(t, cfg)
				return
			}

			if err != nil {
				t.Fatal("unexpected config error")
			}
			require.NotNil(t, cfg)
		})
	}
}

func TestClient_RequestAuthorizationTransport(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		headerName string
		opaque     bool
		wantErr    bool
	}{
		{
			name:       "rejects mutated plaintext endpoint",
			url:        "http://node.example/",
			headerName: "Authorization",
			wantErr:    true,
		},
		{
			name:       "matches mutated mixed-case header",
			url:        "http://node.example/",
			headerName: "aUtHoRiZaTiOn",
			wantErr:    true,
		},
		{
			name:       "rejects mutated malformed endpoint",
			url:        "https://[::1/",
			headerName: "Authorization",
			wantErr:    true,
		},
		{
			name:    "rejects mutated plaintext URL userinfo",
			url:     testUserinfoURL("http", "node.example"),
			wantErr: true,
		},
		{
			name:       "rejects opaque HTTP client",
			url:        "https://node.example/",
			headerName: "authorization",
			opaque:     true,
			wantErr:    true,
		},
		{
			name:       "executes HTTPS header request",
			url:        "https://node.example/",
			headerName: "authorization",
		},
		{
			name: "executes HTTPS URL userinfo request",
			url:  testUserinfoURL("https", "node.example"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var called atomic.Bool
			var authorizationSeen atomic.Bool
			standardClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				called.Store(true)
				authorizationSeen.Store(hasAuthorizationHeader(req.Header))
				return successfulHTTPResponse(req), nil
			})}

			var httpClient HTTPClient = standardClient
			if tt.opaque {
				httpClient = opaqueHTTPClient{called: &called}
			}
			cfg, err := NewClientConfig("https://initial.example/", WithHTTPClient(httpClient))
			require.NoError(t, err)
			cfg.URL = tt.url
			if tt.headerName != "" {
				cfg.Headers[tt.headerName] = []string{testHeaderValue}
			}

			_, err = NewClient(cfg).Request(validTransportSecurityRequest())
			assertAuthorizationErrorRedacted(t, err)
			if tt.wantErr {
				if !errors.Is(err, ErrInsecureAuthorization) {
					t.Fatal("expected insecure authorization error")
				}
				require.False(t, called.Load())
				return
			}

			if err != nil {
				t.Fatal("unexpected request error")
			}
			require.True(t, called.Load())
			require.True(t, authorizationSeen.Load())
		})
	}
}

func TestAuthorizationDiagnosticsRedactHeaderValues(t *testing.T) {
	var logs bytes.Buffer
	previousLogger := clientconfig.SetLogger(log.New(&logs, "", 0))
	t.Cleanup(func() { clientconfig.SetLogger(previousLogger) })

	_, err := NewClientConfig(
		"http://node.example",
		withTestAuthorizationHeader("Authorization"),
	)
	assertAuthorizationErrorRedacted(t, err)
	if !errors.Is(err, ErrInsecureAuthorization) {
		t.Fatal("expected insecure authorization error")
	}
	assertAuthorizationTextRedacted(t, logs.String(), "authorization warning exposed credential material")
}

func TestClient_RequestRedactsTransportError(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, credentialEchoError("transport failure: " + req.Header.Get(authorizationHeader))
	})}
	cfg, err := NewClientConfig(
		"https://node.example",
		WithHTTPClient(httpClient),
		withTestAuthorizationHeader("Authorization"),
	)
	require.NoError(t, err)

	_, err = NewClient(cfg).Request(validTransportSecurityRequest())
	assertAuthorizationErrorRedacted(t, err)
	if !errors.Is(err, ErrAuthorizationRequestFailed) {
		t.Fatal("expected redacted authorization request error")
	}
}

func TestClient_AuthorizationRedirectDowngrade(t *testing.T) {
	tests := []struct {
		name             string
		useURLUserinfo   bool
		mutateInCallback bool
	}{
		{name: "header"},
		{name: "URL userinfo", useURLUserinfo: true},
		{name: "callback URL mutation", mutateInCallback: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var plaintextRequestReceived atomic.Bool
			plaintextServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				plaintextRequestReceived.Store(true)
			}))
			defer plaintextServer.Close()

			var tlsServer *httptest.Server
			tlsServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !hasAuthorizationHeader(r.Header) {
					t.Error("TLS request did not carry authorization header")
					return
				}
				redirectURL := plaintextServer.URL
				if tt.mutateInCallback {
					redirectURL = tlsServer.URL + "/secure-redirect"
				}
				http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
			}))
			defer tlsServer.Close()

			httpClient := tlsServer.Client()
			if tt.mutateInCallback {
				httpClient.CheckRedirect = func(req *http.Request, _ []*http.Request) error {
					redirectURL, err := url.Parse(plaintextServer.URL)
					if err != nil {
						return err
					}
					req.URL = redirectURL
					return nil
				}
			}

			serverURL := tlsServer.URL
			var opts []ConfigOpt
			opts = append(opts, WithHTTPClient(httpClient))
			if tt.useURLUserinfo {
				serverURL = addTestUserinfo(t, serverURL)
			} else {
				opts = append(opts, withTestAuthorizationHeader("Authorization"))
			}
			cfg, err := NewClientConfig(serverURL, opts...)
			require.NoError(t, err)

			_, err = NewClient(cfg).Request(validTransportSecurityRequest())
			assertAuthorizationErrorRedacted(t, err)
			if !errors.Is(err, ErrInsecureAuthorization) {
				t.Fatal("expected authenticated redirect downgrade to be refused")
			}
			require.False(t, plaintextRequestReceived.Load())
		})
	}
}

func TestAuthorizationRedirectPolicyChecksDowngradeBeforeCaller(t *testing.T) {
	sourceRequest := httptest.NewRequest(http.MethodGet, "https://example.com/source", nil)
	sourceRequest.Header.Set("Authorization", testHeaderValue)

	t.Run("rejects downgrade without calling caller policy", func(t *testing.T) {
		redirectRequest := httptest.NewRequest(http.MethodGet, "http://example.com/target", nil)
		var callerInvoked atomic.Bool
		policy := authorizationRedirectPolicy(func(*http.Request, []*http.Request) error {
			callerInvoked.Store(true)
			return nil
		})

		err := policy(redirectRequest, []*http.Request{sourceRequest})

		require.ErrorIs(t, err, ErrInsecureAuthorization)
		require.False(t, callerInvoked.Load())
	})

	t.Run("calls caller policy for HTTPS redirect", func(t *testing.T) {
		redirectRequest := httptest.NewRequest(http.MethodGet, "https://example.com/target", nil)
		var callerInvoked atomic.Bool
		policy := authorizationRedirectPolicy(func(*http.Request, []*http.Request) error {
			callerInvoked.Store(true)
			return nil
		})

		err := policy(redirectRequest, []*http.Request{sourceRequest})

		require.NoError(t, err)
		require.True(t, callerInvoked.Load())
	})
}

func withTestAuthorizationHeader(name string) ConfigOpt {
	return func(cfg *Config) {
		cfg.Headers[name] = []string{testHeaderValue}
	}
}

func validTransportSecurityRequest() *account.ChannelsRequest {
	return &account.ChannelsRequest{
		Account: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
	}
}

func successfulHTTPResponse(req *http.Request) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(`{}`)),
		Request:    req,
	}
}

func testUserinfoURL(scheme, host string) string {
	return (&url.URL{
		Scheme: scheme,
		Host:   host,
		User:   url.UserPassword(testURLUsername, testURLPassword),
	}).String()
}

func addTestUserinfo(t *testing.T, rawURL string) string {
	t.Helper()
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		t.Fatal("failed to prepare authenticated test URL")
	}
	parsedURL.User = url.UserPassword(testURLUsername, testURLPassword)
	return parsedURL.String()
}

func assertAuthorizationErrorRedacted(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		assertAuthorizationTextRedacted(t, err.Error(), "authorization error exposed credential material")
	}
}

func assertAuthorizationTextRedacted(t *testing.T, text, failureMessage string) {
	t.Helper()
	basicAuthorization := "Basic " + base64.StdEncoding.EncodeToString([]byte(testURLUsername+":"+testURLPassword))
	for _, value := range []string{testHeaderValue, testURLUsername, testURLPassword, basicAuthorization} {
		if strings.Contains(text, value) {
			t.Fatal(failureMessage)
		}
	}
}
