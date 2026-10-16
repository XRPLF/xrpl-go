package rpc

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
	"github.com/stretchr/testify/require"
)

func TestRequestContextCancellation(t *testing.T) {
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		if r.Header.Get("X-Hold") == "yes" {
			started <- struct{}{}
			select {
			case <-r.Context().Done():
			case <-release:
			}
			return
		}
		_, _ = w.Write([]byte(`{"result":{"status":"success","info":{"network_id":2}}}`))
	}))
	defer s.Close()
	defer close(release)
	cfg, err := NewClientConfig(s.URL, WithTimeout(time.Second))
	require.NoError(t, err)
	c := NewClient(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = c.RequestContext(ctx, &server.InfoRequest{})
	require.ErrorIs(t, err, context.Canceled)
	_, err = c.Request(&server.InfoRequest{})
	require.NoError(t, err)
	cfg.Headers = map[string][]string{"X-Hold": {"yes"}}
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := c.RequestContext(ctx, &server.InfoRequest{}); done <- err }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("handler not reached")
	}
	require.ErrorIs(t, <-done, context.DeadlineExceeded)
	cfg.Headers = nil
	_, err = c.RequestContext(context.Background(), &server.InfoRequest{})
	require.NoError(t, err)
}

func TestRequestContextCancelDuringRetryBackoff(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		attempts := 0
		cfg, err := NewClientConfig("https://unused", WithRetryDelay(time.Hour), WithHTTPClient(reviewHTTPFunc(func(*http.Request) (*http.Response, error) {
			attempts++
			return &http.Response{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(strings.NewReader("busy"))}, nil
		})))
		require.NoError(t, err)
		c := NewClient(cfg)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		done := make(chan error, 1)
		go func() { _, err := c.RequestContext(ctx, &server.InfoRequest{}); done <- err }()
		synctest.Wait()
		require.Equal(t, 1, attempts)
		cancel()
		require.ErrorIs(t, <-done, context.Canceled)
		time.Sleep(2 * time.Hour)
		require.Equal(t, 1, attempts, "cancellation must not start another HTTP attempt")
	})
}

type reviewHTTPFunc func(*http.Request) (*http.Response, error)

func (f reviewHTTPFunc) Do(r *http.Request) (*http.Response, error) { return f(r) }
