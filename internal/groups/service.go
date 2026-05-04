package groups

import (
	"context"
	"fmt"

	admin "google.golang.org/api/admin/directory/v1"

	"github.com/haydary1986/ok-goldy-alternative/internal/workspace"
)

// Service exposes Group and Member use-cases. The workspace client is
// hot-swappable via a Provider; when no client is set, every method
// returns ErrWorkspaceUnavailable.
type Service struct{ wsProv *workspace.Provider }

func NewService(wsProv *workspace.Provider) *Service { return &Service{wsProv: wsProv} }

func (s *Service) client() (*workspace.Client, error) {
	if s.wsProv == nil {
		return nil, ErrWorkspaceUnavailable
	}
	c := s.wsProv.Get()
	if c == nil {
		return nil, ErrWorkspaceUnavailable
	}
	return c, nil
}

// --- groups ---

func (s *Service) List(ctx context.Context, pageToken string, pageSize int64) (*ListGroupsResponse, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	apiGroups, next, err := c.ListGroupsPage(ctx, pageToken, pageSize)
	if err != nil {
		return nil, fmt.Errorf("groups: list: %w", err)
	}
	out := make([]Group, 0, len(apiGroups))
	for _, g := range apiGroups {
		out = append(out, fromAPIGroup(g))
	}
	return &ListGroupsResponse{Groups: out, NextPageToken: next}, nil
}

func (s *Service) Get(ctx context.Context, key string) (*Group, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	g, err := c.GetGroup(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("groups: get: %w", err)
	}
	out := fromAPIGroup(g)
	return &out, nil
}

func (s *Service) Create(ctx context.Context, req CreateGroupRequest) (*Group, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	out, err := c.InsertGroup(ctx, &admin.Group{
		Email:       req.Email,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("groups: create: %w", err)
	}
	g := fromAPIGroup(out)
	return &g, nil
}

func (s *Service) Delete(ctx context.Context, key string) error {
	c, err := s.client()
	if err != nil {
		return err
	}
	if err := c.DeleteGroup(ctx, key); err != nil {
		return fmt.Errorf("groups: delete: %w", err)
	}
	return nil
}

// --- members ---

func (s *Service) ListMembers(ctx context.Context, groupKey, pageToken string, pageSize int64) (*ListMembersResponse, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	apiMembers, next, err := c.ListMembersPage(ctx, groupKey, pageToken, pageSize)
	if err != nil {
		return nil, fmt.Errorf("groups: list members: %w", err)
	}
	out := make([]Member, 0, len(apiMembers))
	for _, m := range apiMembers {
		out = append(out, fromAPIMember(m))
	}
	return &ListMembersResponse{Members: out, NextPageToken: next}, nil
}

func (s *Service) AddMember(ctx context.Context, groupKey string, req AddMemberRequest) (*Member, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	role := req.Role
	if role == "" {
		role = "MEMBER"
	}
	out, err := c.InsertMember(ctx, groupKey, &admin.Member{
		Email: req.Email,
		Role:  role,
	})
	if err != nil {
		return nil, fmt.Errorf("groups: add member: %w", err)
	}
	m := fromAPIMember(out)
	return &m, nil
}

func (s *Service) RemoveMember(ctx context.Context, groupKey, memberKey string) error {
	c, err := s.client()
	if err != nil {
		return err
	}
	if err := c.DeleteMember(ctx, groupKey, memberKey); err != nil {
		return fmt.Errorf("groups: remove member: %w", err)
	}
	return nil
}

func fromAPIGroup(g *admin.Group) Group {
	return Group{
		ID:                 g.Id,
		Email:              g.Email,
		Name:               g.Name,
		Description:        g.Description,
		DirectMembersCount: g.DirectMembersCount,
	}
}

func fromAPIMember(m *admin.Member) Member {
	return Member{
		ID:    m.Id,
		Email: m.Email,
		Role:  m.Role,
		Type:  m.Type,
	}
}
