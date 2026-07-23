package rpc

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	account "github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	utility "github.com/Peersyst/xrpl-go/xrpl/queries/utility"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateRequest(t *testing.T) {
	t.Run("Create request", func(t *testing.T) {
		req := &account.ChannelsRequest{
			Account:            "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
			DestinationAccount: "rnZvsWuLem5Ha46AZs61jLWR9R5esinkG3",
			LedgerIndex:        common.Validated,
		}

		req.SetAPIVersion(req.APIVersion())

		expetedBody := Request{
			Method: "account_channels",
			Params: [1]any{req},
		}
		expectedRequestBytes, _ := jsoniter.Marshal(expetedBody)

		byteRequest, err := createRequest(req)

		require.NoError(t, err)
		// assert bytes equal
		assert.Equal(t, expectedRequestBytes, byteRequest)
		// assert json equal
		assert.Equal(t, string(expectedRequestBytes), string(byteRequest))
	})
	t.Run("Create request - no parameters with using pointer declaration", func(t *testing.T) {
		req := &utility.RandomRequest{} // params sent in as zero value struct

		req.SetAPIVersion(req.APIVersion())

		expetedBody := Request{
			Method: req.Method(),
			Params: [1]any{req},
		}
		expectedRequestBytes, _ := jsoniter.Marshal(expetedBody)

		byteRequest, err := createRequest(req)

		require.NoError(t, err)
		// assert bytes equal
		assert.Equal(t, expectedRequestBytes, byteRequest)
		// assert json equal
		assert.Equal(t, string(expectedRequestBytes), string(byteRequest))
	})

	t.Run("Create request - no parameters with struct initialisation", func(t *testing.T) {
		req := &utility.RandomRequest{} // means params get set an empty object

		req.SetAPIVersion(req.APIVersion())

		expetedBody := Request{
			Method: req.Method(),
			Params: [1]any{req},
		}
		expectedRequestBytes, _ := jsoniter.Marshal(expetedBody)

		byteRequest, err := createRequest(req)

		require.NoError(t, err)
		// assert bytes equal
		assert.Equal(t, expectedRequestBytes, byteRequest)
		// assert json equal
		assert.Equal(t, string(expectedRequestBytes), string(byteRequest))
	})
}

func TestCheckForError(t *testing.T) {
	jsonRPCError := `{"result":{"error":"ledgerIndexMalformed","status":"error"}}`
	simpleSuccess := `{"result":{"status":"success"}}`
	nullMethod := "Null Method" // https://xrpl.org/error-formatting.html#universal-errors

	tests := []struct {
		name              string
		body              []byte
		statusCode        int
		maxResponseSize   int64
		expectedClientErr string
		expectedErr       error
		expectedResultErr string
		expectedStatus    string
		expectedEmpty     bool
	}{
		{
			name:              "fail - error response",
			body:              []byte(jsonRPCError),
			statusCode:        200,
			maxResponseSize:   defaultMaxResponseSize,
			expectedClientErr: "ledgerIndexMalformed",
			expectedResultErr: "ledgerIndexMalformed",
		},
		{
			name:              "fail - error response with error code",
			body:              []byte(nullMethod),
			statusCode:        400,
			maxResponseSize:   defaultMaxResponseSize,
			expectedClientErr: "Null Method",
		},
		{
			name:            "pass - no error response",
			body:            []byte(simpleSuccess),
			statusCode:      200,
			maxResponseSize: defaultMaxResponseSize,
			expectedStatus:  "success",
		},
		{
			name:            "pass - no error response under max size",
			body:            []byte(simpleSuccess),
			statusCode:      200,
			maxResponseSize: int64(len(simpleSuccess)),
			expectedStatus:  "success",
		},
		{
			name:              "fail - error response with error code under max size",
			body:              []byte(nullMethod),
			statusCode:        400,
			maxResponseSize:   int64(len(nullMethod)),
			expectedClientErr: "Null Method",
		},
		{
			name:            "fail - response over max size",
			body:            bytes.Repeat([]byte("a"), 11),
			statusCode:      200,
			maxResponseSize: 10,
			expectedErr:     ErrResponseTooLarge,
			expectedEmpty:   true,
		},
		{
			name:            "pass - zero max size disables limit",
			body:            []byte(simpleSuccess),
			statusCode:      200,
			maxResponseSize: 0,
			expectedStatus:  "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(bytes.NewReader(tt.body)),
			}

			response, err := checkForError(res, tt.maxResponseSize)

			switch {
			case tt.expectedErr != nil:
				require.ErrorIs(t, err, tt.expectedErr)
			case tt.expectedClientErr != "":
				var clientErr *ClientError
				require.ErrorAs(t, err, &clientErr)
				assert.Equal(t, tt.expectedClientErr, clientErr.ErrorString)
			default:
				require.NoError(t, err)
			}

			if tt.expectedEmpty {
				assert.Empty(t, response)
				return
			}
			if tt.expectedResultErr != "" {
				assert.Equal(t, tt.expectedResultErr, response.Result["error"])
			}
			if tt.expectedStatus != "" {
				assert.Equal(t, tt.expectedStatus, response.Result["status"])
			}
		})
	}
}
