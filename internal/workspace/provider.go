package workspace

import (
	"context"
	"sync/atomic"
)

// Provider holds the currently active *Client and lets it be hot-swapped at
// runtime when an admin uploads new credentials via the Settings page.
//
// It also remembers the Config used to build the current client so callers
// can ask for "the same credentials, but a different scope set" — that's
// how the orgunits domain gets its own narrow-scoped client without
// duplicating credential plumbing.
type Provider struct {
	current atomic.Pointer[Client]
	cfg     atomic.Pointer[Config]
}

// NewProvider builds a Provider with an optional initial client. Pass nil
// for both to start uninitialised (e.g. when no env vars are set and no
// DB row yet exists).
func NewProvider(initial *Client) *Provider {
	p := &Provider{}
	p.current.Store(initial)
	return p
}

// Get returns the current *Client, or nil if no credentials are configured.
func (p *Provider) Get() *Client { return p.current.Load() }

// Set replaces the current client. Passing nil clears the provider so
// subsequent Get() calls return nil. Pass cfg = nil to leave the stored
// config untouched (e.g. when clearing only the cached client).
func (p *Provider) Set(c *Client, cfg *Config) {
	p.current.Store(c)
	if cfg != nil {
		// Store a copy so a caller that mutates the original Config later
		// can't surprise the provider.
		copyCfg := *cfg
		p.cfg.Store(&copyCfg)
	}
	if c == nil {
		p.cfg.Store(nil)
	}
}

// Variant builds a fresh Client from the same credentials currently held
// by the provider, but with the supplied scopes overriding DefaultScopes.
// Returns an error when no credentials are configured.
func (p *Provider) Variant(ctx context.Context, scopes []string) (*Client, error) {
	cfg := p.cfg.Load()
	if cfg == nil {
		return nil, ErrNoCredentials
	}
	variant := *cfg
	variant.Scopes = scopes
	return New(ctx, variant)
}
