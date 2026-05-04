package workspace

import "sync/atomic"

// Provider holds the currently active *Client and lets it be hot-swapped at
// runtime when an admin uploads new credentials via the Settings page.
//
// All callers retrieve the current client via Get(); if Get() returns nil
// the caller should surface a "workspace not configured" error to the user.
type Provider struct {
	current atomic.Pointer[Client]
}

// NewProvider builds a Provider with an optional initial client. Pass nil
// to start uninitialised (e.g. when no env vars are set and no DB row yet
// exists).
func NewProvider(initial *Client) *Provider {
	p := &Provider{}
	p.current.Store(initial)
	return p
}

// Get returns the current *Client, or nil if no credentials are configured.
func (p *Provider) Get() *Client { return p.current.Load() }

// Set replaces the current client. Passing nil clears the provider so
// subsequent Get() calls return nil.
func (p *Provider) Set(c *Client) { p.current.Store(c) }
