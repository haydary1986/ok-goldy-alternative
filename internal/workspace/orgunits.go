package workspace

import (
	"context"
	"fmt"

	admin "google.golang.org/api/admin/directory/v1"
)

// ListOrgUnits returns every OU in the customer with `type=all` so descendants
// are included. The Admin SDK does not paginate this endpoint — it returns up
// to ~1000 OUs in one call, which is more than enough for any real Workspace.
func (c *Client) ListOrgUnits(ctx context.Context) ([]*admin.OrgUnit, error) {
	if err := c.Wait(ctx); err != nil {
		return nil, err
	}
	resp, err := c.dir.Orgunits.List(c.customerID).Type("all").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("workspace: list orgunits: %w", err)
	}
	return resp.OrganizationUnits, nil
}

// InsertOrgUnit creates a new OU. ParentOrgUnitPath must point at an existing
// OU; Workspace returns 400 otherwise.
func (c *Client) InsertOrgUnit(ctx context.Context, ou *admin.OrgUnit) (*admin.OrgUnit, error) {
	if err := c.Wait(ctx); err != nil {
		return nil, err
	}
	out, err := c.dir.Orgunits.Insert(c.customerID, ou).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("workspace: insert orgunit: %w", err)
	}
	return out, nil
}
