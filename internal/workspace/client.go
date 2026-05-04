// Package workspace wraps the Google Admin SDK Directory client with a
// rate limiter, so all callers share a single quota-friendly entry point.
package workspace

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	"golang.org/x/time/rate"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

// DefaultScopes are the OAuth scopes Goldy requires from the service account.
// These must be authorized via Domain-Wide Delegation by a super admin.
var DefaultScopes = []string{
	admin.AdminDirectoryUserScope,
	admin.AdminDirectoryGroupScope,
	admin.AdminDirectoryGroupMemberScope,
	admin.AdminDirectoryUserAliasScope,
}

// Config holds the inputs needed to construct a workspace.Client.
//
// Either ServiceAccountKey (raw JSON bytes, preferred — used when credentials
// come from the database via the admin UI) or ServiceAccountKeyFile (path on
// disk, used when credentials come from environment variables) must be set.
type Config struct {
	ServiceAccountKey     []byte
	ServiceAccountKeyFile string
	DelegatedAdmin        string
	CustomerID            string
	RateLimitRPS          int
	RateLimitBurst        int
}

// Client wraps the Admin SDK Directory service with a token-bucket limiter.
type Client struct {
	dir        *admin.Service
	limiter    *rate.Limiter
	customerID string
}

// New builds a Client. The service-account JSON key is loaded from disk and
// configured to impersonate the delegated super-admin user.
func New(ctx context.Context, c Config) (*Client, error) {
	if c.DelegatedAdmin == "" {
		return nil, fmt.Errorf("workspace: delegated admin email is required")
	}

	keyBytes := c.ServiceAccountKey
	if len(keyBytes) == 0 {
		if c.ServiceAccountKeyFile == "" {
			return nil, fmt.Errorf("workspace: ServiceAccountKey bytes or ServiceAccountKeyFile path is required")
		}
		fileBytes, err := os.ReadFile(c.ServiceAccountKeyFile)
		if err != nil {
			return nil, fmt.Errorf("workspace: read sa key file: %w", err)
		}
		keyBytes = fileBytes
	}

	jwtCfg, err := google.JWTConfigFromJSON(keyBytes, DefaultScopes...)
	if err != nil {
		return nil, fmt.Errorf("workspace: parse sa key: %w", err)
	}
	jwtCfg.Subject = c.DelegatedAdmin

	dir, err := admin.NewService(ctx, option.WithHTTPClient(jwtCfg.Client(ctx)))
	if err != nil {
		return nil, fmt.Errorf("workspace: build admin service: %w", err)
	}

	rps := c.RateLimitRPS
	if rps <= 0 {
		rps = 20
	}
	burst := c.RateLimitBurst
	if burst <= 0 {
		burst = rps * 2
	}

	return &Client{
		dir:        dir,
		limiter:    rate.NewLimiter(rate.Limit(rps), burst),
		customerID: c.CustomerID,
	}, nil
}

// Wait blocks until the token-bucket allows the next API call.
func (c *Client) Wait(ctx context.Context) error { return c.limiter.Wait(ctx) }

// CustomerID returns the configured Workspace customer identifier.
func (c *Client) CustomerID() string { return c.customerID }

// Directory returns the underlying Admin SDK Directory service for callers
// that need access to APIs the wrapper has not yet exposed.
func (c *Client) Directory() *admin.Service { return c.dir }
