package rpc

import (
	"context"
	"sync"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
)

type networkIdentityState struct {
	mu          sync.Mutex
	ready       bool
	discovering *networkIdentityDiscovery
}

type networkIdentityDiscovery struct {
	done                  chan struct{}
	identity              clientinternal.NetworkIdentity
	err                   error
	leaderContextCanceled bool
}

type networkIdentityAttempt struct {
	identity  clientinternal.NetworkIdentity
	ready     bool
	discovery *networkIdentityDiscovery
	leader    bool
}

// ensureNetworkIdentity returns a configured identity or discovers it with
// server_info. A caller-provided NetworkID is compared with discovery and is
// never replaced when it matches. WithNetworkIdentity marks the initial state
// ready and bypasses discovery only when its build version is non-empty.
// Discovery errors are returned and are not cached, so a later operation can
// retry.
func (c *Client) ensureNetworkIdentity(ctx context.Context) (clientinternal.NetworkIdentity, error) {
	for {
		if err := ctx.Err(); err != nil {
			return clientinternal.NetworkIdentity{}, err
		}

		attempt := c.beginNetworkIdentityDiscovery()
		if attempt.ready {
			return clientinternal.ValidateNetworkIdentity(attempt.identity)
		}
		if !attempt.leader {
			select {
			case <-ctx.Done():
				return clientinternal.NetworkIdentity{}, ctx.Err()
			case <-attempt.discovery.done:
			}
			if err := ctx.Err(); err != nil {
				return clientinternal.NetworkIdentity{}, err
			}
			if attempt.discovery.leaderContextCanceled {
				continue
			}
			return attempt.discovery.identity, attempt.discovery.err
		}

		var response server.InfoResponse
		requestErr := c.requestResult(ctx, &server.InfoRequest{}, &response)
		resolved := attempt.identity
		discoveryErr := requestErr
		if requestErr == nil {
			resolved, discoveryErr = clientinternal.ResolveNetworkIdentity(
				attempt.identity.NetworkID,
				clientinternal.NetworkIdentity{
					NetworkID:    response.Info.NetworkID,
					BuildVersion: response.Info.ServerVersion(),
				},
			)
		}
		c.finishNetworkIdentityDiscovery(resolved, discoveryErr, ctx.Err() != nil)
		if discoveryErr != nil {
			return clientinternal.NetworkIdentity{}, discoveryErr
		}
		return resolved, nil
	}
}

// NetworkIdentity returns a thread-safe snapshot of the client network ID and
// server build version. The returned network ID does not alias client state.
func (c *Client) NetworkIdentity() (*uint32, string) {
	c.identity.mu.Lock()
	defer c.identity.mu.Unlock()
	return clientinternal.CloneNetworkID(c.networkID), c.buildVersion
}

func (c *Client) networkIdentity() (clientinternal.NetworkIdentity, error) {
	c.identity.mu.Lock()
	defer c.identity.mu.Unlock()

	return clientinternal.ValidateNetworkIdentity(clientinternal.NetworkIdentity{
		NetworkID:    c.networkID,
		BuildVersion: c.buildVersion,
	})
}

func (c *Client) beginNetworkIdentityDiscovery() networkIdentityAttempt {
	c.identity.mu.Lock()
	defer c.identity.mu.Unlock()

	identity := clientinternal.NetworkIdentity{
		NetworkID:    c.networkID,
		BuildVersion: c.buildVersion,
	}
	if c.identity.ready {
		return networkIdentityAttempt{identity: identity, ready: true}
	}
	if c.identity.discovering != nil {
		return networkIdentityAttempt{identity: identity, discovery: c.identity.discovering}
	}

	discovery := &networkIdentityDiscovery{done: make(chan struct{})}
	c.identity.discovering = discovery
	return networkIdentityAttempt{identity: identity, discovery: discovery, leader: true}
}

func (c *Client) finishNetworkIdentityDiscovery(
	identity clientinternal.NetworkIdentity,
	discoveryErr error,
	leaderContextCanceled bool,
) {
	c.identity.mu.Lock()
	defer c.identity.mu.Unlock()

	discovery := c.identity.discovering
	discovery.err = discoveryErr
	discovery.leaderContextCanceled = leaderContextCanceled
	if discoveryErr == nil {
		c.networkID = clientinternal.CloneNetworkID(identity.NetworkID)
		c.buildVersion = identity.BuildVersion
		c.identity.ready = true
		discovery.identity = clientinternal.NetworkIdentity{
			NetworkID:    clientinternal.CloneNetworkID(identity.NetworkID),
			BuildVersion: identity.BuildVersion,
		}
	}
	c.identity.discovering = nil
	close(discovery.done)
}
