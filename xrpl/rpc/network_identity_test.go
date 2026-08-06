package rpc

import (
	"errors"
	"net/http"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/rpc/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/stretchr/testify/require"
)

func uint32Pointer(value uint32) *uint32 {
	return &value
}

func TestClientEnsureNetworkIdentity(t *testing.T) {
	requestFailure := errors.New("server_info unavailable")
	tests := []struct {
		name             string
		response         string
		requestErr       error
		override         *uint32
		buildOverride    string
		ensureCalls      int // number of ensureNetworkIdentity calls, where 0 means 1
		expectedID       *uint32
		expectedBuild    string
		expectedErr      error
		expectedRequests int
		preserveOverride bool
		trustedConfig    bool
	}{
		{
			name:             "discovers valid mainnet zero",
			response:         `{"result":{"info":{"network_id":0,"build_version":"1.12.0"}}}`,
			expectedID:       uint32Pointer(0),
			expectedBuild:    "1.12.0",
			expectedRequests: 1,
		},
		{
			name:             "missing network ID fails closed",
			response:         `{"result":{"info":{"build_version":"1.12.0"}}}`,
			expectedErr:      ErrNetworkIDUnavailable,
			expectedRequests: 1,
		},
		{
			name:             "request error propagates",
			requestErr:       requestFailure,
			expectedErr:      requestFailure,
			expectedRequests: 1,
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
			name:             "complete trusted override bypasses discovery",
			override:         uint32Pointer(21337),
			buildOverride:    "1.12.0",
			expectedID:       uint32Pointer(21337),
			expectedBuild:    "1.12.0",
			expectedRequests: 0,
			trustedConfig:    true,
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
			if tt.trustedConfig {
				options = append(options, WithNetworkIdentity(*tt.override, tt.buildOverride))
			}
			cfg, err := NewClientConfig("http://localhost/", options...)
			require.NoError(t, err)
			cl := NewClient(cfg)
			if !tt.trustedConfig {
				cl.NetworkID = tt.override
				cl.BuildVersion = tt.buildOverride
			}

			identity, err := cl.ensureNetworkIdentity()
			for call := 1; call < tt.ensureCalls; call++ {
				identity, err = cl.ensureNetworkIdentity()
			}
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, identity.NetworkID)
				require.Equal(t, *tt.expectedID, *identity.NetworkID)
				require.Equal(t, tt.expectedBuild, identity.BuildVersion)
			}
			require.Equal(t, tt.expectedRequests, requestCount)
			if tt.preserveOverride {
				require.Same(t, tt.override, cl.NetworkID)
			}
			if tt.expectedErr == nil {
				require.Equal(t, tt.expectedBuild, cl.BuildVersion)
			}
		})
	}
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
		transaction.FlatTransaction{"TransactionType": "AccountSet"},
		false,
		&wallet.Wallet{},
	)
	require.ErrorIs(t, err, ErrNetworkIDUnavailable)
}
