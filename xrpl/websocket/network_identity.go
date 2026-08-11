package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
)

type networkIdentityState struct {
	mu      sync.Mutex
	ready   bool
	trusted bool
	current clientinternal.NetworkIdentity
}

// prepareNetworkIdentity returns a configured identity or performs the
// synchronous server_info handshake used by Connect. WithNetworkIdentity marks
// the initial state trusted and bypasses discovery only when its build version
// is non-empty. After a successful discovery, a reconnect compares the new
// server identity with the previous one and rejects a network ID change.
func (c *Client) prepareNetworkIdentity(
	ctx context.Context,
	conn websocketConnection,
) ([][]byte, error) {
	identity, ready, trusted := c.networkIdentitySnapshot()
	if ready && trusted {
		_, err := clientinternal.ValidateNetworkIdentity(identity)
		return nil, err
	}

	discovered, bufferedMessages, err := c.discoverNetworkIdentity(ctx, conn)
	if err != nil {
		return bufferedMessages, err
	}
	resolved, err := clientinternal.ResolveNetworkIdentity(identity.NetworkID, discovered)
	if err != nil {
		return bufferedMessages, err
	}
	c.storeDiscoveredNetworkIdentity(resolved)
	return bufferedMessages, nil
}

// NetworkIdentity returns a thread-safe snapshot of the client network ID and
// server build version. The returned network ID does not alias client state.
func (c *Client) NetworkIdentity() (*uint32, string) {
	c.identity.mu.Lock()
	defer c.identity.mu.Unlock()

	return clientinternal.CloneNetworkID(c.identity.current.NetworkID), c.identity.current.BuildVersion
}

func (c *Client) networkIdentitySnapshot() (clientinternal.NetworkIdentity, bool, bool) {
	c.identity.mu.Lock()
	defer c.identity.mu.Unlock()
	return c.identity.current, c.identity.ready, c.identity.trusted
}

func (c *Client) storeDiscoveredNetworkIdentity(identity clientinternal.NetworkIdentity) {
	c.identity.mu.Lock()
	defer c.identity.mu.Unlock()
	c.identity.current = clientinternal.NetworkIdentity{
		NetworkID:    clientinternal.CloneNetworkID(identity.NetworkID),
		BuildVersion: identity.BuildVersion,
	}
	c.identity.ready = true
	c.identity.trusted = false
}

func (c *Client) discoverNetworkIdentity(
	ctx context.Context,
	conn websocketConnection,
) (clientinternal.NetworkIdentity, [][]byte, error) {
	id := c.idCounter.Add(1)
	message, err := c.formatRequest(&server.InfoRequest{}, id, nil)
	if err != nil {
		return clientinternal.NetworkIdentity{}, nil, err
	}
	if err := c.conn.writeMessageTo(ctx, conn, message, c.cfg.timeout); err != nil {
		return clientinternal.NetworkIdentity{}, nil, err
	}

	deadline := time.Now().Add(c.cfg.timeout)
	var bufferedMessages [][]byte
	for {
		responseBytes, err := c.conn.readMessageFrom(conn, deadline)
		if err != nil {
			var timeoutErr interface{ Timeout() bool }
			if errors.As(err, &timeoutErr) && timeoutErr.Timeout() {
				err = errors.Join(ErrRequestTimedOut, err)
			}
			return clientinternal.NetworkIdentity{}, bufferedMessages, err
		}

		var response ClientResponse
		if err := json.Unmarshal(responseBytes, &response); err != nil {
			return clientinternal.NetworkIdentity{}, bufferedMessages, err
		}
		if response.ID != id {
			bufferedMessages = append(bufferedMessages, append([]byte(nil), responseBytes...))
			continue
		}
		if err := response.CheckError(); err != nil {
			return clientinternal.NetworkIdentity{}, bufferedMessages, err
		}

		var serverInfo server.InfoResponse
		if err := response.GetResult(&serverInfo); err != nil {
			return clientinternal.NetworkIdentity{}, bufferedMessages, err
		}
		return clientinternal.NetworkIdentity{
			NetworkID:    serverInfo.Info.NetworkID,
			BuildVersion: serverInfo.Info.ServerVersion(),
		}, bufferedMessages, nil
	}
}

func (c *Client) networkIdentity() (clientinternal.NetworkIdentity, error) {
	identity, ready, _ := c.networkIdentitySnapshot()
	if !ready {
		return clientinternal.NetworkIdentity{}, ErrNetworkIDUnavailable
	}
	return clientinternal.ValidateNetworkIdentity(identity)
}
