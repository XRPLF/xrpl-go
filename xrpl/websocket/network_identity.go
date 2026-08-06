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
// the initial state trusted and intentionally bypasses discovery.
func (c *Client) prepareNetworkIdentity() error {
	identity, ready, trusted := c.networkIdentitySnapshot()
	if ready && trusted {
		_, err := clientinternal.ValidateNetworkIdentity(identity)
		return err
	}
	discovered, err := c.discoverNetworkIdentity()
	if err != nil {
		return fmt.Errorf("%w: %w", errNetworkIdentityDiscovery, err)
	}
	resolved, err := clientinternal.ResolveNetworkIdentity(identity.NetworkID, discovered)
	if err != nil {
		return err
	}
	c.storeDiscoveredNetworkIdentity(resolved)
	return nil
}

func (c *Client) networkIdentity() (clientinternal.NetworkIdentity, error) {
	identity, _, _ := c.networkIdentitySnapshot()
	return clientinternal.ValidateNetworkIdentity(identity)
}

func (c *Client) networkIdentitySnapshot() (clientinternal.NetworkIdentity, bool, bool) {
	c.identity.mu.Lock()
	defer c.identity.mu.Unlock()
	if c.identity.ready {
		return c.identity.current, true, c.identity.trusted
	}
	return clientinternal.NetworkIdentity{
		NetworkID:    c.NetworkID,
		BuildVersion: c.BuildVersion,
	}, false, false
}

func (c *Client) discoverNetworkIdentity() (clientinternal.NetworkIdentity, error) {
	id := c.idCounter.Add(1)
	message, err := c.formatRequest(&server.InfoRequest{}, id, nil)
	if err != nil {
		return clientinternal.NetworkIdentity{}, err
	}
	if err := c.conn.WriteMessage(message); err != nil {
		return clientinternal.NetworkIdentity{}, fmt.Errorf("%w: %w", errNetworkIdentityConnection, err)
	}

	responseBytes, err := c.conn.readMessage(time.Now().Add(c.cfg.timeout))
	if err != nil {
		var timeoutErr interface{ Timeout() bool }
		if errors.As(err, &timeoutErr) && timeoutErr.Timeout() {
			err = errors.Join(ErrRequestTimedOut, err)
		}
		return clientinternal.NetworkIdentity{}, fmt.Errorf("%w: %w", errNetworkIdentityConnection, err)
	}

	var response ClientResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return clientinternal.NetworkIdentity{}, err
	}
	if response.ID != id {
		return clientinternal.NetworkIdentity{}, ErrIncorrectID
	}
	if err := response.CheckError(); err != nil {
		return clientinternal.NetworkIdentity{}, err
	}

	var serverInfo server.InfoResponse
	if err := response.GetResult(&serverInfo); err != nil {
		return clientinternal.NetworkIdentity{}, err
	}
	return clientinternal.NetworkIdentity{
		NetworkID:    serverInfo.Info.NetworkID,
		BuildVersion: serverInfo.Info.BuildVersion,
	}, nil
}

func (c *Client) storeDiscoveredNetworkIdentity(identity clientinternal.NetworkIdentity) {
	c.identity.mu.Lock()
	defer c.identity.mu.Unlock()
	c.identity.current = clientinternal.NetworkIdentity{
		NetworkID:    clientinternal.CloneNetworkID(identity.NetworkID),
		BuildVersion: identity.BuildVersion,
	}
	c.NetworkID = identity.NetworkID
	c.BuildVersion = identity.BuildVersion
	c.identity.ready = true
	c.identity.trusted = false
}
