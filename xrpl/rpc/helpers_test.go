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
			Params: [1]interface{}{req},
		}
		expectedRequestBytes, _ := jsoniter.Marshal(expetedBody)

		byteRequest, err := createRequest(req)

		assert.NoError(t, err)
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
			Params: [1]interface{}{req},
		}
		expectedRequestBytes, _ := jsoniter.Marshal(expetedBody)

		byteRequest, err := createRequest(req)

		assert.NoError(t, err)
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
			Params: [1]interface{}{req},
		}
		expectedRequestBytes, _ := jsoniter.Marshal(expetedBody)

		byteRequest, err := createRequest(req)

		assert.NoError(t, err)
		// assert bytes equal
		assert.Equal(t, expectedRequestBytes, byteRequest)
		// assert json equal
		assert.Equal(t, string(expectedRequestBytes), string(byteRequest))
	})
}

func TestCheckForError(t *testing.T) {

	t.Run("Error Response", func(t *testing.T) {

		json := `{
			"result": {
				"error": "ledgerIndexMalformed",
				"request": {
					"account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					"command": "account_info",
					"ledger_index": "-",
					"strict": true
				},
				"status": "error"
			}
		}`

		b := io.NopCloser(bytes.NewReader([]byte(json)))
		res := &http.Response{
			StatusCode: 200, // error response still returns a 200
			Body:       b,
		}

		bodyBytes, err := checkForError(res)
		assert.NotNil(t, bodyBytes)
		expError := &ClientError{ErrorString: "ledgerIndexMalformed"}
		assert.Equal(t, expError, err)
	})

	t.Run("Error Response with error code", func(t *testing.T) {

		json := "Null Method" // https://xrpl.org/error-formatting.html#universal-errors

		b := io.NopCloser(bytes.NewReader([]byte(json)))
		res := &http.Response{
			StatusCode: 400,
			Body:       b,
		}

		bodyBytes, err := checkForError(res)
		assert.NotNil(t, bodyBytes)
		expErrpr := &ClientError{ErrorString: "Null Method"}
		assert.Equal(t, expErrpr, err)
	})

	t.Run("No error Response", func(t *testing.T) {

		json := `{
			"result": {
			  "account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
			  "channels": [
				{
				  "account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				  "amount": "1000",
				  "balance": "0",
				  "channel_id": "C7F634794B79DB40E87179A9D1BF05D05797AE7E92DF8E93FD6656E8C4BE3AE7",
				  "destination_account": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
				  "public_key": "aBR7mdD75Ycs8DRhMgQ4EMUEmBArF8SEh1hfjrT2V9DQTLNbJVqw",
				  "public_key_hex": "03CFD18E689434F032A4E84C63E2A3A6472D684EAF4FD52CA67742F3E24BAE81B2",
				  "settle_delay": 60
				}
			  ],
			  "ledger_hash": "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
			  "ledger_index": 71766343,
			  "status": "success",
			  "validated": true
			}
		  }`

		b := io.NopCloser(bytes.NewReader([]byte(json)))
		res := &http.Response{
			StatusCode: 200,
			Body:       b,
		}

		bodyBytes, err := checkForError(res)

		assert.Nil(t, err)
		assert.NotNil(t, bodyBytes)
	})
}
