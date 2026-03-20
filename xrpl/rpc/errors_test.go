package rpc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientError(t *testing.T) {
	err := &ClientError{ErrorString: "test error"}
	require.Equal(t, "test error", err.Error())
}
