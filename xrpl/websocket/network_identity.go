package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
)

var (
	errNetworkIdentityDiscovery  = errors.New("network identity discovery failed")
	errNetworkIdentityConnection = errors.New("network identity connection failed")
)

type networkIdentityState struct {
	mu      sync.Mutex
	ready   bool
	trusted bool
	current clientinternal.NetworkIdentity
}

// prepareNetworkIdentity returns a configured identity or performs the
// synchronous server_info handshake used by Connect. WithNetworkIdentity marks
// the initial state trusted and intentionally bypasses discovery. After a
// successful discovery, an explicit reconnect compares the new server identity
// with the previous one and rejects a network ID change.
func (c *Client) prepareNetworkIdentity() ([][]byte, error) {
	identity, ready, trusted := c.networkIdentitySnapshot()
	if ready && trusted {
		_, err := clientinternal.ValidateNetworkIdentity(identity)
		return nil, err
	}
	discovered, bufferedMessages, err := c.discoverNetworkIdentity()
	if err != nil {
		return bufferedMessages, fmt.Errorf("%w: %w", errNetworkIdentityDiscovery, err)
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

func (c *Client) networkIdentity() (clientinternal.NetworkIdentity, error) {
	identity, _, _ := c.networkIdentitySnapshot()
	return clientinternal.ValidateNetworkIdentity(identity)
}

func (c *Client) networkIdentitySnapshot() (clientinternal.NetworkIdentity, bool, bool) {
	c.identity.mu.Lock()
	defer c.identity.mu.Unlock()
	return c.identity.current, c.identity.ready, c.identity.trusted
}

func (c *Client) discoverNetworkIdentity() (clientinternal.NetworkIdentity, [][]byte, error) {
	id := c.idCounter.Add(1)
	message, err := c.formatRequest(&server.InfoRequest{}, id, nil)
	if err != nil {
		return clientinternal.NetworkIdentity{}, nil, err
	}
	if err := c.conn.WriteMessage(message); err != nil {
		return clientinternal.NetworkIdentity{}, nil, fmt.Errorf("%w: %w", errNetworkIdentityConnection, err)
	}

	deadline := time.Now().Add(c.cfg.timeout)
	var bufferedMessages [][]byte
	for {
		responseBytes, err := c.conn.readMessage(deadline)
		if err != nil {
			var timeoutErr interface{ Timeout() bool }
			if errors.As(err, &timeoutErr) && timeoutErr.Timeout() {
				err = errors.Join(ErrRequestTimedOut, err)
			}
			return clientinternal.NetworkIdentity{}, bufferedMessages, fmt.Errorf("%w: %w", errNetworkIdentityConnection, err)
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
			BuildVersion: serverInfo.Info.BuildVersion,
		}, bufferedMessages, nil
	}
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
