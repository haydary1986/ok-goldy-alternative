package groups

import (
	"context"
	"fmt"

	admin "google.golang.org/api/admin/directory/v1"

	"github.com/haydary1986/ok-goldy-alternative/internal/workspace"
)

// Service exposes Group and Member use-cases. The workspace client is
// optional; when nil, every method returns ErrWorkspaceUnavailable.
type Service struct{ ws *workspace.Client }

func NewService(ws *workspace.Client) *Service { return &Service{ws: ws} }

// --- groups ---

func (s *Service) List(ctx context.Context, pageToken string, pageSize int64) (*ListGroupsResponse, error) {
	if s.ws == nil {
		return nil, ErrWorkspaceUnavailable
	}
	apiGroups, next, err := s.ws.ListGroupsPage(ctx, pageToken, pageSize)
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
	if s.ws == nil {
		return nil, ErrWorkspaceUnavailable
	}
	g, err := s.ws.GetGroup(ctx, key)
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
	if s.ws == nil {
		return nil, ErrWorkspaceUnavailable
	}
	out, err := s.ws.InsertGroup(ctx, &admin.Group{
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
	if s.ws == nil {
		return ErrWorkspaceUnavailable
	}
	if err := s.ws.DeleteGroup(ctx, key); err != nil {
		return fmt.Errorf("groups: delete: %w", err)
	}
	return nil
}

// --- members ---

func (s *Service) ListMembers(ctx context.Context, groupKey, pageToken string, pageSize int64) (*ListMembersResponse, error) {
	if s.ws == nil {
		return nil, ErrWorkspaceUnavailable
	}
	apiMembers, next, err := s.ws.ListMembersPage(ctx, groupKey, pageToken, pageSize)
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
	if s.ws == nil {
		return nil, ErrWorkspaceUnavailable
	}
	role := req.Role
	if role == "" {
		role = "MEMBER"
	}
	out, err := s.ws.InsertMember(ctx, groupKey, &admin.Member{
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
	if s.ws == nil {
		return ErrWorkspaceUnavailable
	}
	if err := s.ws.DeleteMember(ctx, groupKey, memberKey); err != nil {
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
