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

type networkIdentityDiscoveryResult struct {
	identitySnapshot clientinternal.NetworkIdentity
	ready            bool
	discoveryDone    <-chan struct{}
	shouldDiscover   bool
}

// ensureNetworkIdentity returns a configured identity or discovers it with
// server_info. A caller-provided NetworkID is compared with discovery and is
// never replaced when it matches. WithNetworkIdentity marks the initial state
// ready and intentionally bypasses discovery.
func (c *Client) ensureNetworkIdentity() (clientinternal.NetworkIdentity, error) {
	for {
		discovery := c.beginNetworkIdentityDiscovery()
		if discovery.ready {
			return clientinternal.ValidateNetworkIdentity(discovery.identitySnapshot)
		}
		if !discovery.shouldDiscover {
			<-discovery.discoveryDone
			continue
		}

		response, err := c.GetServerInfo(&server.InfoRequest{})
		var resolved clientinternal.NetworkIdentity
		if err == nil {
			resolved, err = clientinternal.ResolveNetworkIdentity(discovery.identitySnapshot.NetworkID, clientinternal.NetworkIdentity{
				NetworkID:    response.Info.NetworkID,
				BuildVersion: response.Info.ServerVersion(),
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

func (c *Client) beginNetworkIdentityDiscovery() networkIdentityDiscoveryResult {
	c.identity.mu.Lock()
	defer c.identity.mu.Unlock()

	identitySnapshot := clientinternal.NetworkIdentity{
		NetworkID:    c.NetworkID,
		BuildVersion: c.BuildVersion,
	}
	if c.identity.ready {
		return networkIdentityDiscoveryResult{
			identitySnapshot: identitySnapshot,
			ready:            true,
		}
	}
	if c.identity.discovering != nil {
		return networkIdentityDiscoveryResult{
			identitySnapshot: identitySnapshot,
			discoveryDone:    c.identity.discovering,
		}
	}

	c.identity.discovering = make(chan struct{})
	return networkIdentityDiscoveryResult{
		identitySnapshot: identitySnapshot,
		discoveryDone:    c.identity.discovering,
		shouldDiscover:   true,
	}
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
