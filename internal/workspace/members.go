package workspace

import (
	"context"
	"fmt"

	admin "google.golang.org/api/admin/directory/v1"
)

// ListMembersPage fetches one page of members in a group.
func (c *Client) ListMembersPage(ctx context.Context, groupKey, pageToken string, pageSize int64) ([]*admin.Member, string, error) {
	if err := c.Wait(ctx); err != nil {
		return nil, "", err
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 200
	}
	call := c.dir.Members.List(groupKey).MaxResults(pageSize).Context(ctx)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	resp, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("workspace: list members of %q: %w", groupKey, err)
	}
	return resp.Members, resp.NextPageToken, nil
}

// InsertMember adds a member to a group.
func (c *Client) InsertMember(ctx context.Context, groupKey string, m *admin.Member) (*admin.Member, error) {
	if err := c.Wait(ctx); err != nil {
		return nil, err
	}
	out, err := c.dir.Members.Insert(groupKey, m).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("workspace: insert member into %q: %w", groupKey, err)
	}
	return out, nil
}

// DeleteMember removes a member from a group.
func (c *Client) DeleteMember(ctx context.Context, groupKey, memberKey string) error {
	if err := c.Wait(ctx); err != nil {
		return err
	}
	if err := c.dir.Members.Delete(groupKey, memberKey).Context(ctx).Do(); err != nil {
		return fmt.Errorf("workspace: delete member %q from %q: %w", memberKey, groupKey, err)
	}
	return nil
}
