package rpc

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/rpc/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/stretchr/testify/require"
)

func uint32Pointer(value uint32) *uint32 {
	return &value
}

func boolPointer(value bool) *bool {
	return &value
}

type doneObservedContext struct {
	context.Context
	doneCalled chan struct{}
	once       sync.Once
}

func (c *doneObservedContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.doneCalled) })
	return c.Context.Done()
}

type networkIdentityResult struct {
	identity clientinternal.NetworkIdentity
	err      error
}

func ensureNetworkIdentityAsync(
	cl *Client,
	ctx context.Context,
	results chan<- networkIdentityResult,
) {
	identity, err := cl.ensureNetworkIdentity(ctx)
	results <- networkIdentityResult{identity: identity, err: err}
}

func startNetworkIdentityFollowers(
	t *testing.T,
	cl *Client,
	count int,
	results chan<- networkIdentityResult,
) {
	t.Helper()
	followers := make([]*doneObservedContext, 0, count)
	for range count {
		followerCtx := &doneObservedContext{
			Context:    context.Background(),
			doneCalled: make(chan struct{}),
		}
		followers = append(followers, followerCtx)
		go ensureNetworkIdentityAsync(cl, followerCtx, results)
	}

	deadline := time.After(time.Second)
	for _, followerCtx := range followers {
		select {
		case <-followerCtx.doneCalled:
		case <-deadline:
			t.Fatal("identity follower did not wait for shared discovery")
		}
	}
}

func collectNetworkIdentityResults(
	t *testing.T,
	results <-chan networkIdentityResult,
	count int,
) []networkIdentityResult {
	t.Helper()
	collected := make([]networkIdentityResult, 0, count)
	deadline := time.After(time.Second)
	for range count {
		select {
		case result := <-results:
			collected = append(collected, result)
		case <-deadline:
			t.Fatal("concurrent identity caller was not released")
		}
	}
	return collected
}

func TestClientBeginNetworkIdentityDiscoveryResult(t *testing.T) {
	cfg, err := NewClientConfig("http://localhost/")
	require.NoError(t, err)
	cl := NewClient(cfg)

	first := cl.beginNetworkIdentityDiscovery()
	require.Nil(t, first.identity.NetworkID)
	require.Empty(t, first.identity.BuildVersion)
	require.False(t, first.ready)
	require.NotNil(t, first.discovery)
	require.True(t, first.leader)

	waiting := cl.beginNetworkIdentityDiscovery()
	require.False(t, waiting.ready)
	require.Same(t, first.discovery, waiting.discovery)
	require.False(t, waiting.leader)

	resolved := clientinternal.NetworkIdentity{
		NetworkID:    uint32Pointer(21337),
		BuildVersion: "1.12.0",
	}
	cl.finishNetworkIdentityDiscovery(resolved, nil, false)
	select {
	case <-first.discovery.done:
	default:
		t.Fatal("discovery completion channel was not closed")
	}

	ready := cl.beginNetworkIdentityDiscovery()
	require.True(t, ready.ready)
	require.Equal(t, uint32(21337), *ready.identity.NetworkID)
	require.Equal(t, "1.12.0", ready.identity.BuildVersion)
	require.Nil(t, ready.discovery)
	require.False(t, ready.leader)
}

func TestClientNetworkIdentityDiscoveryFailure(t *testing.T) {
	cfg, err := NewClientConfig("http://localhost/")
	require.NoError(t, err)
	cl := NewClient(cfg)

	leader := cl.beginNetworkIdentityDiscovery()
	require.True(t, leader.leader)
	waiting := cl.beginNetworkIdentityDiscovery()
	require.False(t, waiting.leader)
	require.Same(t, leader.discovery, waiting.discovery)

	discoveryErr := errors.New("server_info unavailable")
	cl.finishNetworkIdentityDiscovery(clientinternal.NetworkIdentity{}, discoveryErr, false)
	select {
	case <-leader.discovery.done:
	case <-time.After(time.Second):
		t.Fatal("identity discovery failure did not release waiters")
	}
	require.ErrorIs(t, leader.discovery.err, discoveryErr)
	require.ErrorIs(t, waiting.discovery.err, discoveryErr)

	retry := cl.beginNetworkIdentityDiscovery()
	require.True(t, retry.leader)
	require.NotSame(t, leader.discovery, retry.discovery)
	cl.finishNetworkIdentityDiscovery(clientinternal.NetworkIdentity{}, discoveryErr, false)
}

func TestClientEnsureNetworkIdentitySingleflight(t *testing.T) {
	const callerCount = 8

	t.Run("one successful request serves all callers", func(t *testing.T) {
		mockClient := &testutil.JSONRPCMockClient{}
		var requestCount atomic.Int32
		requestStarted := make(chan struct{})
		releaseRequest := make(chan struct{})
		var releaseRequestOnce sync.Once
		release := func() {
			releaseRequestOnce.Do(func() {
				close(releaseRequest)
			})
		}
		t.Cleanup(release)
		unexpectedRequest := errors.New("unexpected additional server_info request")
		mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			if requestCount.Add(1) != 1 {
				return nil, unexpectedRequest
			}
			close(requestStarted)
			<-releaseRequest
			return testutil.MockResponse(
				`{"result":{"info":{"network_id":21337,"build_version":"1.12.0"}}}`,
				http.StatusOK,
				mockClient,
			)(req)
		}
		cfg, err := NewClientConfig("http://localhost/", WithHTTPClient(mockClient))
		require.NoError(t, err)
		cl := NewClient(cfg)

		results := make(chan networkIdentityResult, callerCount)
		go ensureNetworkIdentityAsync(cl, context.Background(), results)
		select {
		case <-requestStarted:
		case <-time.After(time.Second):
			t.Fatal("server_info request did not start")
		}
		startNetworkIdentityFollowers(t, cl, callerCount-1, results)
		select {
		case result := <-results:
			t.Fatalf("identity caller returned before server_info was released: %v", result.err)
		default:
		}
		release()

		for _, result := range collectNetworkIdentityResults(t, results, callerCount) {
			require.NoError(t, result.err)
			require.NotNil(t, result.identity.NetworkID)
			require.Equal(t, uint32(21337), *result.identity.NetworkID)
			require.Equal(t, "1.12.0", result.identity.BuildVersion)
		}
		require.Equal(t, int32(1), requestCount.Load())
	})

	t.Run("discovery error releases all callers", func(t *testing.T) {
		requestFailure := errors.New("server_info unavailable")
		mockClient := &testutil.JSONRPCMockClient{}
		var requestCount atomic.Int32
		requestStarted := make(chan struct{})
		releaseRequest := make(chan struct{})
		var releaseRequestOnce sync.Once
		release := func() {
			releaseRequestOnce.Do(func() {
				close(releaseRequest)
			})
		}
		t.Cleanup(release)
		mockClient.DoFunc = func(*http.Request) (*http.Response, error) {
			if requestCount.Add(1) == 1 {
				close(requestStarted)
				<-releaseRequest
			}
			return nil, requestFailure
		}
		cfg, err := NewClientConfig("http://localhost/", WithHTTPClient(mockClient))
		require.NoError(t, err)
		cl := NewClient(cfg)

		results := make(chan networkIdentityResult, callerCount)
		go ensureNetworkIdentityAsync(cl, context.Background(), results)
		select {
		case <-requestStarted:
		case <-time.After(time.Second):
			t.Fatal("server_info request did not start")
		}
		startNetworkIdentityFollowers(t, cl, callerCount-1, results)
		release()

		for _, result := range collectNetworkIdentityResults(t, results, callerCount) {
			require.ErrorIs(t, result.err, requestFailure)
		}
	})
}

func TestClientEnsureNetworkIdentityFollowerCancellation(t *testing.T) {
	mockClient := &testutil.JSONRPCMockClient{}
	var requestCount atomic.Int32
	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	var releaseRequestOnce sync.Once
	release := func() { releaseRequestOnce.Do(func() { close(releaseRequest) }) }
	t.Cleanup(release)
	mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		requestCount.Add(1)
		close(requestStarted)
		<-releaseRequest
		return testutil.MockResponse(
			`{"result":{"info":{"network_id":21337,"build_version":"1.12.0"}}}`,
			http.StatusOK,
			mockClient,
		)(req)
	}
	cfg, err := NewClientConfig("http://localhost/", WithHTTPClient(mockClient))
	require.NoError(t, err)
	cl := NewClient(cfg)

	leaderResult := make(chan error, 1)
	go func() {
		_, discoveryErr := cl.ensureNetworkIdentity(context.Background())
		leaderResult <- discoveryErr
	}()
	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("server_info request did not start")
	}

	baseCtx, cancel := context.WithCancel(context.Background())
	followerCtx := &doneObservedContext{
		Context:    baseCtx,
		doneCalled: make(chan struct{}),
	}
	followerResult := make(chan error, 1)
	go func() {
		_, discoveryErr := cl.ensureNetworkIdentity(followerCtx)
		followerResult <- discoveryErr
	}()
	select {
	case <-followerCtx.doneCalled:
	case <-time.After(time.Second):
		t.Fatal("identity follower did not wait for shared discovery")
	}
	cancel()

	select {
	case err := <-followerResult:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("cancelled identity follower did not return")
	}

	release()
	select {
	case err := <-leaderResult:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("identity leader did not return")
	}
	require.Equal(t, int32(1), requestCount.Load())
}

func TestClientEnsureNetworkIdentityFollowerRetriesAfterLeaderCancellation(t *testing.T) {
	mockClient := &testutil.JSONRPCMockClient{}
	var requestCount atomic.Int32
	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	var releaseRequestOnce sync.Once
	release := func() { releaseRequestOnce.Do(func() { close(releaseRequest) }) }
	t.Cleanup(release)
	mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		if requestCount.Add(1) == 1 {
			close(requestStarted)
			<-releaseRequest
			return nil, req.Context().Err()
		}
		return testutil.MockResponse(
			`{"result":{"info":{"network_id":21337,"build_version":"1.12.0"}}}`,
			http.StatusOK,
			mockClient,
		)(req)
	}
	cfg, err := NewClientConfig("http://localhost/", WithHTTPClient(mockClient))
	require.NoError(t, err)
	cl := NewClient(cfg)

	leaderCtx, cancelLeader := context.WithCancel(context.Background())
	leaderResult := make(chan networkIdentityResult, 1)
	go ensureNetworkIdentityAsync(cl, leaderCtx, leaderResult)
	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("server_info request did not start")
	}

	followerCtx := &doneObservedContext{
		Context:    context.Background(),
		doneCalled: make(chan struct{}),
	}
	followerResult := make(chan networkIdentityResult, 1)
	go ensureNetworkIdentityAsync(cl, followerCtx, followerResult)
	select {
	case <-followerCtx.doneCalled:
	case <-time.After(time.Second):
		t.Fatal("identity follower did not wait for shared discovery")
	}

	cancelLeader()
	release()
	select {
	case result := <-leaderResult:
		require.ErrorIs(t, result.err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("cancelled identity leader did not return")
	}

	select {
	case result := <-followerResult:
		require.NoError(t, result.err)
		require.NotNil(t, result.identity.NetworkID)
		require.Equal(t, uint32(21337), *result.identity.NetworkID)
		require.Equal(t, "1.12.0", result.identity.BuildVersion)
	case <-time.After(time.Second):
		t.Fatal("identity follower did not retry discovery")
	}
	require.Equal(t, int32(2), requestCount.Load())
}

func TestClientEnsureNetworkIdentity(t *testing.T) {
	requestFailure := errors.New("server_info unavailable")
	tests := []struct {
		name                      string
		response                  string
		requestErr                error
		override                  *uint32
		buildOverride             string
		ensureCalls               int // number of ensureNetworkIdentity calls, where 0 means 1
		expectedID                *uint32
		expectedBuild             string
		expectedErr               error
		expectedRequests          int
		expectedNetworkIDRequired *bool
		preserveOverride          bool
		configuredIdentity        bool
	}{
		{
			name:             "discovers valid mainnet zero",
			response:         `{"result":{"info":{"network_id":0,"build_version":"1.12.0"}}}`,
			expectedID:       uint32Pointer(0),
			expectedBuild:    "1.12.0",
			expectedRequests: 1,
		},
		{
			name:                      "Clio rippled version fallback",
			response:                  `{"result":{"info":{"network_id":21337,"rippled_version":"1.12.0"}}}`,
			expectedID:                uint32Pointer(21337),
			expectedBuild:             "1.12.0",
			expectedRequests:          1,
			expectedNetworkIDRequired: boolPointer(true),
		},
		{
			name:                      "build version preferred over rippled version",
			response:                  `{"result":{"info":{"network_id":21337,"build_version":"1.10.0","rippled_version":"1.12.0"}}}`,
			expectedID:                uint32Pointer(21337),
			expectedBuild:             "1.10.0",
			expectedRequests:          1,
			expectedNetworkIDRequired: boolPointer(false),
		},
		{
			name:             "missing network ID fails closed",
			response:         `{"result":{"info":{"build_version":"1.12.0"}}}`,
			expectedErr:      ErrNetworkIDUnavailable,
			expectedRequests: 1,
		},
		{
			name:             "request error is returned and later calls retry",
			requestErr:       requestFailure,
			ensureCalls:      2,
			expectedErr:      requestFailure,
			expectedRequests: 2,
		},
		{
			name:             "matching override is preserved",
			response:         `{"result":{"info":{"network_id":21337,"build_version":"1.12.0"}}}`,
			override:         uint32Pointer(21337),
			buildOverride:    "1.10.0",
			expectedID:       uint32Pointer(21337),
			expectedBuild:    "1.12.0",
			expectedRequests: 1,
			preserveOverride: true,
		},
		{
			name:             "mismatching override fails without erasing override",
			response:         `{"result":{"info":{"network_id":21338,"build_version":"1.12.0"}}}`,
			override:         uint32Pointer(21337),
			expectedErr:      ErrNetworkIDOverrideMismatch,
			expectedRequests: 1,
			preserveOverride: true,
		},
		{
			name:               "incomplete configured identity performs discovery",
			response:           `{"result":{"info":{"network_id":21337,"build_version":"1.12.0"}}}`,
			override:           uint32Pointer(21337),
			expectedID:         uint32Pointer(21337),
			expectedBuild:      "1.12.0",
			expectedRequests:   1,
			configuredIdentity: true,
		},
		{
			name:               "complete trusted override bypasses discovery",
			override:           uint32Pointer(21337),
			buildOverride:      "1.12.0",
			expectedID:         uint32Pointer(21337),
			expectedBuild:      "1.12.0",
			expectedRequests:   0,
			configuredIdentity: true,
		},
		{
			name:             "cached discovery is reused",
			response:         `{"result":{"info":{"network_id":1,"build_version":"1.12.0"}}}`,
			ensureCalls:      2,
			expectedID:       uint32Pointer(1),
			expectedBuild:    "1.12.0",
			expectedRequests: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &testutil.JSONRPCMockClient{}
			requestCount := 0
			mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
				requestCount++
				if tt.requestErr != nil {
					return nil, tt.requestErr
				}
				return testutil.MockResponse(tt.response, http.StatusOK, mockClient)(req)
			}
			options := []ConfigOpt{WithHTTPClient(mockClient)}
			if tt.configuredIdentity {
				options = append(options, WithNetworkIdentity(*tt.override, tt.buildOverride))
			}
			cfg, err := NewClientConfig("http://localhost/", options...)
			require.NoError(t, err)
			cl := NewClient(cfg)
			if !tt.configuredIdentity {
				setTestNetworkIdentity(cl, tt.override, tt.buildOverride)
			}

			identity, err := cl.ensureNetworkIdentity(context.Background())
			for call := 1; call < tt.ensureCalls; call++ {
				identity, err = cl.ensureNetworkIdentity(context.Background())
			}
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, identity.NetworkID)
				require.Equal(t, *tt.expectedID, *identity.NetworkID)
				require.Equal(t, tt.expectedBuild, identity.BuildVersion)
				if tt.expectedNetworkIDRequired != nil {
					required, policyErr := clientinternal.NetworkIDRequired(identity)
					require.NoError(t, policyErr)
					require.Equal(t, *tt.expectedNetworkIDRequired, required)
				}
			}
			require.Equal(t, tt.expectedRequests, requestCount)
			storedNetworkID, storedBuildVersion := cl.NetworkIdentity()
			if tt.preserveOverride {
				require.NotNil(t, storedNetworkID)
				require.Equal(t, *tt.override, *storedNetworkID)
			}
			if tt.expectedErr == nil {
				require.Equal(t, tt.expectedBuild, storedBuildVersion)
			}
		})
	}
}

func TestClientNetworkIdentityReturnsSnapshot(t *testing.T) {
	cfg, err := NewClientConfig("http://localhost/", WithNetworkIdentity(21337, "1.12.0"))
	require.NoError(t, err)
	cl := NewClient(cfg)

	networkID, buildVersion := cl.NetworkIdentity()
	require.NotNil(t, networkID)
	require.Equal(t, uint32(21337), *networkID)
	require.Equal(t, "1.12.0", buildVersion)

	*networkID = 1
	storedNetworkID, storedBuildVersion := cl.NetworkIdentity()
	require.NotNil(t, storedNetworkID)
	require.Equal(t, uint32(21337), *storedNetworkID)
	require.Equal(t, "1.12.0", storedBuildVersion)
}

func TestClientEnsureNetworkIdentityCoalescesConcurrentDiscovery(t *testing.T) {
	const callers = 32

	mockClient := &testutil.JSONRPCMockClient{}
	var requestCount atomic.Int32
	requestStarted := make(chan struct{})
	releaseResponse := make(chan struct{})
	var releaseResponseOnce sync.Once
	release := func() { releaseResponseOnce.Do(func() { close(releaseResponse) }) }
	t.Cleanup(release)
	mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		requestCount.Add(1)
		close(requestStarted)
		<-releaseResponse
		return testutil.MockResponse(
			`{"result":{"info":{"network_id":21337,"build_version":"1.12.0"}}}`,
			http.StatusOK,
			mockClient,
		)(req)
	}

	cfg, err := NewClientConfig("http://localhost/", WithHTTPClient(mockClient))
	require.NoError(t, err)
	cl := NewClient(cfg)

	results := make(chan networkIdentityResult, callers)
	go ensureNetworkIdentityAsync(cl, context.Background(), results)
	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("concurrent identity discovery did not send server_info")
	}

	startNetworkIdentityFollowers(t, cl, callers-1, results)
	release()

	for _, result := range collectNetworkIdentityResults(t, results, callers) {
		require.NoError(t, result.err)
		require.NotNil(t, result.identity.NetworkID)
		require.Equal(t, uint32(21337), *result.identity.NetworkID)
		require.Equal(t, "1.12.0", result.identity.BuildVersion)
	}
	require.Equal(t, int32(1), requestCount.Load())
}

func TestClientEnsureNetworkIdentityCoalescesConcurrentFailure(t *testing.T) {
	const callers = 32

	requestFailure := errors.New("server_info unavailable")
	mockClient := &testutil.JSONRPCMockClient{}
	var requestCount atomic.Int32
	requestStarted := make(chan struct{})
	releaseResponse := make(chan struct{})
	var releaseResponseOnce sync.Once
	release := func() { releaseResponseOnce.Do(func() { close(releaseResponse) }) }
	t.Cleanup(release)
	mockClient.DoFunc = func(*http.Request) (*http.Response, error) {
		if requestCount.Add(1) == 1 {
			close(requestStarted)
			<-releaseResponse
		}
		return nil, requestFailure
	}

	cfg, err := NewClientConfig("http://localhost/", WithHTTPClient(mockClient))
	require.NoError(t, err)
	cl := NewClient(cfg)

	results := make(chan networkIdentityResult, callers)
	go ensureNetworkIdentityAsync(cl, context.Background(), results)
	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("concurrent identity discovery did not send server_info")
	}

	startNetworkIdentityFollowers(t, cl, callers-1, results)
	release()

	for _, result := range collectNetworkIdentityResults(t, results, callers) {
		require.ErrorIs(t, result.err, requestFailure)
	}
	require.Equal(t, int32(1), requestCount.Load())

	_, err = cl.ensureNetworkIdentity(context.Background())
	require.ErrorIs(t, err, requestFailure)
	require.Equal(t, int32(2), requestCount.Load())
}

func TestClientAutofillFailsBeforeMutationWhenNetworkIdentityIsMissing(t *testing.T) {
	mockClient := &testutil.JSONRPCMockClient{}
	mockClient.DoFunc = testutil.MockResponse(
		`{"result":{"info":{"build_version":"1.12.0"}}}`,
		http.StatusOK,
		mockClient,
	)
	cfg, err := NewClientConfig("http://localhost/", WithHTTPClient(mockClient))
	require.NoError(t, err)
	cl := NewClient(cfg)
	tx := transaction.FlatTransaction{
		"Account":            "X7AcgcsBL6XDcUb289X4mJ8djcdyKaB5hJDWMArnXr61cqZ",
		"TransactionType":    "AccountSet",
		"Sequence":           uint32(1),
		"Fee":                "10",
		"LastLedgerSequence": uint32(100),
	}
	expected := transaction.FlatTransaction{
		"Account":            "X7AcgcsBL6XDcUb289X4mJ8djcdyKaB5hJDWMArnXr61cqZ",
		"TransactionType":    "AccountSet",
		"Sequence":           uint32(1),
		"Fee":                "10",
		"LastLedgerSequence": uint32(100),
	}

	err = cl.Autofill(&tx)
	require.ErrorIs(t, err, ErrNetworkIDUnavailable)
	require.Equal(t, expected, tx)
}

func TestClientGetSignedTxFailsClosedWithoutAutofill(t *testing.T) {
	mockClient := &testutil.JSONRPCMockClient{}
	mockClient.DoFunc = testutil.MockResponse(
		`{"result":{"info":{"build_version":"1.12.0"}}}`,
		http.StatusOK,
		mockClient,
	)
	cfg, err := NewClientConfig("http://localhost/", WithHTTPClient(mockClient))
	require.NoError(t, err)
	cl := NewClient(cfg)

	_, err = cl.getSignedTx(
		context.Background(),
		transaction.FlatTransaction{"TransactionType": "AccountSet"},
		false,
		&wallet.Wallet{},
	)
	require.ErrorIs(t, err, ErrNetworkIDUnavailable)
}

func setTestNetworkIdentity(cl *Client, networkID *uint32, buildVersion string) {
	cl.identity.mu.Lock()
	defer cl.identity.mu.Unlock()
	if networkID == nil {
		cl.networkID = nil
	} else {
		value := *networkID
		cl.networkID = &value
	}
	cl.buildVersion = buildVersion
}
