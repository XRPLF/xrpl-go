package rpc

import (
	"sync"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
)

type networkIdentityState struct {
	mu          sync.Mutex
	ready       bool
	discovering chan struct{}
}

// ensureNetworkIdentity returns a configured identity or discovers it with
// server_info. A caller-provided NetworkID is compared with discovery and is
// never replaced when it matches. WithNetworkIdentity marks the initial state
// ready and intentionally bypasses discovery.
func (c *Client) ensureNetworkIdentity() (clientinternal.NetworkIdentity, error) {
	for {
		identity, ready, discoveryDone, discover := c.beginNetworkIdentityDiscovery()
		if ready {
			return clientinternal.ValidateNetworkIdentity(identity)
		}
		if !discover {
			<-discoveryDone
			continue
		}

		response, err := c.GetServerInfo(&server.InfoRequest{})
		var resolved clientinternal.NetworkIdentity
		if err == nil {
			resolved, err = clientinternal.ResolveNetworkIdentity(identity.NetworkID, clientinternal.NetworkIdentity{
				NetworkID:    response.Info.NetworkID,
				BuildVersion: response.Info.BuildVersion,
			})
		}
		c.finishNetworkIdentityDiscovery(resolved, err)
		if err != nil {
			return clientinternal.NetworkIdentity{}, err
		}
		return resolved, nil
	}
}

func (c *Client) networkIdentity() (clientinternal.NetworkIdentity, error) {
	c.identity.mu.Lock()
	defer c.identity.mu.Unlock()

	return clientinternal.ValidateNetworkIdentity(clientinternal.NetworkIdentity{
		NetworkID:    c.NetworkID,
		BuildVersion: c.BuildVersion,
	})
}

func (c *Client) beginNetworkIdentityDiscovery() (clientinternal.NetworkIdentity, bool, <-chan struct{}, bool) {
	c.identity.mu.Lock()
	defer c.identity.mu.Unlock()

	identity := clientinternal.NetworkIdentity{
		NetworkID:    c.NetworkID,
		BuildVersion: c.BuildVersion,
	}
	if c.identity.ready {
		return identity, true, nil, false
	}
	if c.identity.discovering != nil {
		return identity, false, c.identity.discovering, false
	}

	c.identity.discovering = make(chan struct{})
	return identity, false, c.identity.discovering, true
}

func (c *Client) finishNetworkIdentityDiscovery(identity clientinternal.NetworkIdentity, discoveryErr error) {
	c.identity.mu.Lock()
	defer c.identity.mu.Unlock()

	if discoveryErr == nil {
		c.NetworkID = identity.NetworkID
		c.BuildVersion = identity.BuildVersion
		c.identity.ready = true
	}
	discoveryDone := c.identity.discovering
	c.identity.discovering = nil
	close(discoveryDone)
}
