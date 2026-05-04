package workspace

import (
	"context"
	"fmt"

	admin "google.golang.org/api/admin/directory/v1"
)

// ListUsersPage fetches one page of users (default page size 500, the API max).
// pageToken is the opaque token returned by a previous call. An empty token
// returned alongside results means there are no more pages.
func (c *Client) ListUsersPage(ctx context.Context, pageToken string, pageSize int64) ([]*admin.User, string, error) {
	if err := c.Wait(ctx); err != nil {
		return nil, "", err
	}
	if pageSize <= 0 || pageSize > 500 {
		pageSize = 500
	}
	call := c.dir.Users.List().
		Customer(c.customerID).
		MaxResults(pageSize).
		OrderBy("email").
		Context(ctx)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	resp, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("workspace: list users: %w", err)
	}
	return resp.Users, resp.NextPageToken, nil
}

// GetUser fetches a single user by ID or primary email.
func (c *Client) GetUser(ctx context.Context, key string) (*admin.User, error) {
	if err := c.Wait(ctx); err != nil {
		return nil, err
	}
	u, err := c.dir.Users.Get(key).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("workspace: get user %q: %w", key, err)
	}
	return u, nil
}

// InsertUser creates a new user.
func (c *Client) InsertUser(ctx context.Context, u *admin.User) (*admin.User, error) {
	if err := c.Wait(ctx); err != nil {
		return nil, err
	}
	out, err := c.dir.Users.Insert(u).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("workspace: insert user: %w", err)
	}
	return out, nil
}

// UpdateUser applies a partial update to an existing user.
func (c *Client) UpdateUser(ctx context.Context, key string, patch *admin.User) (*admin.User, error) {
	if err := c.Wait(ctx); err != nil {
		return nil, err
	}
	out, err := c.dir.Users.Patch(key, patch).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("workspace: update user %q: %w", key, err)
	}
	return out, nil
}

// SuspendUser flips the suspended flag on the target user.
func (c *Client) SuspendUser(ctx context.Context, key string, suspended bool) error {
	_, err := c.UpdateUser(ctx, key, &admin.User{Suspended: suspended, ForceSendFields: []string{"Suspended"}})
	return err
}

// DeleteUser permanently deletes the user.
func (c *Client) DeleteUser(ctx context.Context, key string) error {
	if err := c.Wait(ctx); err != nil {
		return err
	}
	if err := c.dir.Users.Delete(key).Context(ctx).Do(); err != nil {
		return fmt.Errorf("workspace: delete user %q: %w", key, err)
	}
	return nil
}
