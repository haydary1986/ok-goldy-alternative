package orgunits

import (
	"context"
	"fmt"
	"sort"

	admin "google.golang.org/api/admin/directory/v1"

	"github.com/haydary1986/ok-goldy-alternative/internal/workspace"
)

// Service exposes the OU use-cases the HTTP handlers consume.
//
// We deliberately don't reuse the main workspace.Provider client (which is
// scoped to admin.directory.user/group/member/user.alias). Instead we ask
// the provider for a Variant scoped only to admin.directory.orgunit, so a
// missing orgunit authorization in DWD only breaks OU calls — Users and
// Groups keep working with their existing four-scope token.
type Service struct{ wsProv *workspace.Provider }

func NewService(wsProv *workspace.Provider) *Service { return &Service{wsProv: wsProv} }

func (s *Service) client(ctx context.Context) (*workspace.Client, error) {
	if s.wsProv == nil {
		return nil, ErrWorkspaceUnavailable
	}
	if s.wsProv.Get() == nil {
		return nil, ErrWorkspaceUnavailable
	}
	c, err := s.wsProv.Variant(ctx, workspace.OrgUnitScopes)
	if err != nil {
		return nil, fmt.Errorf("orgunits: build client: %w", err)
	}
	return c, nil
}

// List returns every OU in the customer, plus the implicit root "/", with
// the result sorted by path so the UI can render a clean dropdown.
func (s *Service) List(ctx context.Context) (*ListResponse, error) {
	c, err := s.client(ctx)
	if err != nil {
		return nil, err
	}
	apiOUs, err := c.ListOrgUnits(ctx)
	if err != nil {
		return nil, fmt.Errorf("orgunits: list: %w", err)
	}
	out := make([]OrgUnit, 0, len(apiOUs)+1)
	out = append(out, OrgUnit{Name: "(root)", OrgUnitPath: "/"})
	for _, ou := range apiOUs {
		out = append(out, fromAPI(ou))
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].OrgUnitPath < out[j].OrgUnitPath })
	return &ListResponse{OrgUnits: out}, nil
}

// Create provisions a new OU.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*OrgUnit, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	c, err := s.client(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.InsertOrgUnit(ctx, &admin.OrgUnit{
		Name:              req.Name,
		ParentOrgUnitPath: req.ParentOrgUnitPath,
		Description:       req.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("orgunits: create: %w", err)
	}
	mapped := fromAPI(out)
	return &mapped, nil
}

func fromAPI(o *admin.OrgUnit) OrgUnit {
	return OrgUnit{
		OrgUnitID:         o.OrgUnitId,
		Name:              o.Name,
		OrgUnitPath:       o.OrgUnitPath,
		ParentOrgUnitPath: o.ParentOrgUnitPath,
		Description:       o.Description,
	}
}
