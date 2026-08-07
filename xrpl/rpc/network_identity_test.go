package rpc

import (
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/xrpl/rpc/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/stretchr/testify/require"
)

func TestClientEnsureNetworkIdentity(t *testing.T) {
	requestFailure := errors.New("server_info unavailable")
	tests := []struct {
		name               string
		response           string
		requestErr         error
		override           *uint32
		buildOverride      string
		ensureCalls        int // number of ensureNetworkIdentity calls. A value of 0 means 1.
		expectedID         *uint32
		expectedBuild      string
		expectedErr        error
		expectedRequests   int
		preserveOverride   bool
		configuredIdentity bool
	}{
		{
			name:             "discovers valid mainnet zero",
			response:         `{"result":{"info":{"network_id":0,"build_version":"1.12.0"}}}`,
			expectedID:       uint32Pointer(0),
			expectedBuild:    "1.12.0",
			expectedRequests: 1,
		},
		{
			name:             "missing network ID remains unknown",
			response:         `{"result":{"info":{"build_version":"1.12.0"}}}`,
			expectedBuild:    "1.12.0",
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

			identity, err := cl.ensureNetworkIdentity()
			for call := 1; call < tt.ensureCalls; call++ {
				identity, err = cl.ensureNetworkIdentity()
			}
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				if tt.expectedID == nil {
					require.Nil(t, identity.NetworkID)
				} else {
					require.NotNil(t, identity.NetworkID)
					require.Equal(t, *tt.expectedID, *identity.NetworkID)
				}
				require.Equal(t, tt.expectedBuild, identity.BuildVersion)
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
	var signalRequest sync.Once
	requestStarted := make(chan struct{})
	releaseResponse := make(chan struct{})
	mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		requestCount.Add(1)
		signalRequest.Do(func() { close(requestStarted) })
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

	type result struct {
		networkID    uint32
		buildVersion string
		err          error
	}
	results := make(chan result, callers)
	start := make(chan struct{})
	var ready sync.WaitGroup
	ready.Add(callers)
	for range callers {
		go func() {
			<-start
			ready.Done()
			identity, discoveryErr := cl.ensureNetworkIdentity()
			var networkID uint32
			if identity.NetworkID != nil {
				networkID = *identity.NetworkID
			}
			results <- result{
				networkID:    networkID,
				buildVersion: identity.BuildVersion,
				err:          discoveryErr,
			}
		}()
	}

	close(start)
	ready.Wait()
	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("concurrent identity discovery did not send server_info")
	}
	close(releaseResponse)

	for range callers {
		result := <-results
		require.NoError(t, result.err)
		require.Equal(t, uint32(21337), result.networkID)
		require.Equal(t, "1.12.0", result.buildVersion)
	}
	require.Equal(t, int32(1), requestCount.Load())
}

func TestClientEnsureNetworkIdentityCoalescesConcurrentFailure(t *testing.T) {
	const callers = 32

	requestFailure := errors.New("server_info unavailable")
	mockClient := &testutil.JSONRPCMockClient{}
	var requestCount atomic.Int32
	var signalRequest sync.Once
	requestStarted := make(chan struct{})
	releaseResponse := make(chan struct{})
	mockClient.DoFunc = func(*http.Request) (*http.Response, error) {
		requestCount.Add(1)
		signalRequest.Do(func() { close(requestStarted) })
		<-releaseResponse
		return nil, requestFailure
	}

	cfg, err := NewClientConfig("http://localhost/", WithHTTPClient(mockClient))
	require.NoError(t, err)
	cl := NewClient(cfg)

	results := make(chan error, callers)
	start := make(chan struct{})
	var ready sync.WaitGroup
	ready.Add(callers)
	for range callers {
		go func() {
			<-start
			ready.Done()
			_, discoveryErr := cl.ensureNetworkIdentity()
			results <- discoveryErr
		}()
	}

	close(start)
	ready.Wait()
	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("concurrent identity discovery did not send server_info")
	}
	// Give every released caller time to join the in-flight discovery before the
	// leader publishes its failure.
	time.Sleep(10 * time.Millisecond)
	close(releaseResponse)

	for range callers {
		require.ErrorIs(t, <-results, requestFailure)
	}
	require.Equal(t, int32(1), requestCount.Load())

	_, err = cl.ensureNetworkIdentity()
	require.ErrorIs(t, err, requestFailure)
	require.Equal(t, int32(2), requestCount.Load())
}

func TestClientAutofillOmitsNetworkIDWhenIdentityIsMissing(t *testing.T) {
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

	err = cl.Autofill(&tx)

	require.NoError(t, err)
	require.Equal(t, "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59", tx["Account"])
	require.NotContains(t, tx, "NetworkID")
}

func TestClientGetSignedTxSkipsNetworkPolicyWhenAutofillDisabled(t *testing.T) {
	mockClient := &testutil.JSONRPCMockClient{}
	var requestCount atomic.Int32
	mockClient.DoFunc = func(*http.Request) (*http.Response, error) {
		requestCount.Add(1)
		return nil, errors.New("unexpected request")
	}
	cfg, err := NewClientConfig(
		"http://localhost/",
		WithHTTPClient(mockClient),
		WithNetworkIdentity(21337, "1.12.0"),
	)
	require.NoError(t, err)
	cl := NewClient(cfg)
	signer, err := wallet.FromSeed("sEd7io6yt5dFJrcePgRiFVHvmkJhJD1", "")
	require.NoError(t, err)
	tx := transaction.FlatTransaction{
		"Account":         signer.ClassicAddress.String(),
		"TransactionType": "AccountSet",
		"Fee":             "10",
		"Sequence":        uint32(1),
		"NetworkID":       uint32(1),
	}

	blob, err := cl.getSignedTx(tx, false, &signer)

	require.NoError(t, err)
	require.NotEmpty(t, blob)
	signedTx, err := binarycodec.Decode(blob)
	require.NoError(t, err)
	require.EqualValues(t, uint32(1), signedTx["NetworkID"])
	require.Equal(t, uint32(1), tx["NetworkID"])
	require.Equal(t, int32(0), requestCount.Load())
}

func uint32Pointer(value uint32) *uint32 {
	return &value
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
