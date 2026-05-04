package workspace

import (
	"context"
	"encoding/json"
	"fmt"

	admin "google.golang.org/api/admin/directory/v1"
)

// AliasInfo is the workspace-layer projection of a user alias. Goldy uses
// this minimal shape because the Admin SDK exposes aliases as polymorphic
// []interface{} in list responses.
type AliasInfo struct {
	Alias string `json:"alias"`
}

// ListAliases lists every alias for a user.
func (c *Client) ListAliases(ctx context.Context, userKey string) ([]AliasInfo, error) {
	if err := c.Wait(ctx); err != nil {
		return nil, err
	}
	resp, err := c.dir.Users.Aliases.List(userKey).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("workspace: list aliases of %q: %w", userKey, err)
	}
	out := make([]AliasInfo, 0, len(resp.Aliases))
	for _, raw := range resp.Aliases {
		// resp.Aliases items can be *admin.Alias or *admin.UserAlias; both
		// serialize with an `alias` field, so marshal-unmarshal is portable.
		b, err := json.Marshal(raw)
		if err != nil {
			continue
		}
		var a AliasInfo
		if err := json.Unmarshal(b, &a); err == nil && a.Alias != "" {
			out = append(out, a)
		}
	}
	return out, nil
}

// InsertAlias adds a new alias for the given user. Returns the created alias
// as a string (extracted defensively from the response).
func (c *Client) InsertAlias(ctx context.Context, userKey, alias string) (string, error) {
	if err := c.Wait(ctx); err != nil {
		return "", err
	}
	out, err := c.dir.Users.Aliases.Insert(userKey, &admin.Alias{Alias: alias}).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("workspace: insert alias %q for %q: %w", alias, userKey, err)
	}
	return aliasString(out), nil
}

// DeleteAlias removes an alias from the given user.
func (c *Client) DeleteAlias(ctx context.Context, userKey, alias string) error {
	if err := c.Wait(ctx); err != nil {
		return err
	}
	if err := c.dir.Users.Aliases.Delete(userKey, alias).Context(ctx).Do(); err != nil {
		return fmt.Errorf("workspace: delete alias %q from %q: %w", alias, userKey, err)
	}
	return nil
}

// aliasString safely pulls the alias email out of an *admin.Alias regardless
// of whether the SDK declares the field as string or interface{}.
func aliasString(a *admin.Alias) string {
	if a == nil {
		return ""
	}
	b, err := json.Marshal(a)
	if err != nil {
		return ""
	}
	var aux struct {
		Alias string `json:"alias"`
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return ""
	}
	return aux.Alias
}
