package workspace

import (
	"context"
	"fmt"

	admin "google.golang.org/api/admin/directory/v1"
)

// ListGroupsPage fetches one page of groups (max 200 per Workspace API).
func (c *Client) ListGroupsPage(ctx context.Context, pageToken string, pageSize int64) ([]*admin.Group, string, error) {
	if err := c.Wait(ctx); err != nil {
		return nil, "", err
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 200
	}
	call := c.dir.Groups.List().
		Customer(c.customerID).
		MaxResults(pageSize).
		Context(ctx)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	resp, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("workspace: list groups: %w", err)
	}
	return resp.Groups, resp.NextPageToken, nil
}

// GetGroup fetches a single group by id or email.
func (c *Client) GetGroup(ctx context.Context, key string) (*admin.Group, error) {
	if err := c.Wait(ctx); err != nil {
		return nil, err
	}
	g, err := c.dir.Groups.Get(key).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("workspace: get group %q: %w", key, err)
	}
	return g, nil
}

// InsertGroup creates a new group.
func (c *Client) InsertGroup(ctx context.Context, g *admin.Group) (*admin.Group, error) {
	if err := c.Wait(ctx); err != nil {
		return nil, err
	}
	out, err := c.dir.Groups.Insert(g).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("workspace: insert group: %w", err)
	}
	return out, nil
}

// DeleteGroup permanently removes a group.
func (c *Client) DeleteGroup(ctx context.Context, key string) error {
	if err := c.Wait(ctx); err != nil {
		return err
	}
	if err := c.dir.Groups.Delete(key).Context(ctx).Do(); err != nil {
		return fmt.Errorf("workspace: delete group %q: %w", key, err)
	}
	return nil
}
