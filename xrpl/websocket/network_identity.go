package websocket

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
)

// NetworkIdentity is a concurrency-safe snapshot of the latest network
// identity known to a WebSocket client.
type NetworkIdentity struct {
	// NetworkID is nil only when no network ID is available. A pointer to zero is
	// the mainnet ID.
	NetworkID *uint32
	// BuildVersion is the rippled-compatible server version used for NetworkID
	// policy. For Clio servers, it can come from rippled_version.
	BuildVersion string
}

type networkIdentityState struct {
	mu      sync.Mutex
	ready   bool
	trusted bool
	current clientinternal.NetworkIdentity
}

// CurrentNetworkIdentity returns the latest configured or discovered identity.
// Reconnect discovery can update this value without changing the stable
// Client.NetworkID and Client.BuildVersion fields. The returned NetworkID is a
// copy and can be changed by the caller. ErrNetworkIDUnavailable is returned
// before identity configuration or discovery completes.
func (c *Client) CurrentNetworkIdentity() (NetworkIdentity, error) {
	identity, err := c.networkIdentity()
	if err != nil {
		return NetworkIdentity{}, err
	}
	return NetworkIdentity{
		NetworkID:    clientinternal.CloneNetworkID(identity.NetworkID),
		BuildVersion: identity.BuildVersion,
	}, nil
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
	resolved, err := c.discoverNetworkIdentity(identity.NetworkID)
	if err != nil {
		return err
	}
	c.storeDiscoveredNetworkIdentity(resolved)
	return nil
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

func (c *Client) storeDiscoveredNetworkIdentity(identity clientinternal.NetworkIdentity) {
	c.identity.mu.Lock()
	defer c.identity.mu.Unlock()
	firstDiscovery := !c.identity.ready
	c.identity.current = clientinternal.NetworkIdentity{
		NetworkID:    clientinternal.CloneNetworkID(identity.NetworkID),
		BuildVersion: identity.BuildVersion,
	}
	if firstDiscovery {
		c.NetworkID = identity.NetworkID
		c.BuildVersion = identity.BuildVersion
	}
	c.identity.ready = true
	c.identity.trusted = false
}

func (c *Client) discoverNetworkIdentity(override *uint32) (clientinternal.NetworkIdentity, error) {
	id := c.idCounter.Add(1)
	message, err := c.formatRequest(&server.InfoRequest{}, id, nil)
	if err != nil {
		return clientinternal.NetworkIdentity{}, err
	}
	if err := c.conn.WriteMessage(message); err != nil {
		return clientinternal.NetworkIdentity{}, err
	}

	responseBytes, err := c.conn.readMessage(time.Now().Add(c.cfg.timeout))
	if err != nil {
		var timeoutErr interface{ Timeout() bool }
		if errors.As(err, &timeoutErr) && timeoutErr.Timeout() {
			return clientinternal.NetworkIdentity{}, errors.Join(ErrRequestTimedOut, err)
		}
		return clientinternal.NetworkIdentity{}, err
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
	return clientinternal.ResolveNetworkIdentity(override, clientinternal.NetworkIdentity{
		NetworkID:    serverInfo.Info.NetworkID,
		BuildVersion: serverInfo.Info.ServerVersion(),
	})
}

func (c *Client) networkIdentity() (clientinternal.NetworkIdentity, error) {
	identity, ready, _ := c.networkIdentitySnapshot()
	if !ready {
		return clientinternal.NetworkIdentity{}, ErrNetworkIDUnavailable
	}
	return clientinternal.ValidateNetworkIdentity(identity)
}
