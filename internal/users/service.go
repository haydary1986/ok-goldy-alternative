package users

import (
	"context"
	"encoding/json"
	"fmt"

	admin "google.golang.org/api/admin/directory/v1"

	"github.com/haydary1986/ok-goldy-alternative/internal/workspace"
)

// Service composes a workspace.Provider (live data, hot-swappable client)
// with a Repository (local cache) and exposes the use-cases the HTTP
// handlers consume.
//
// Both the Provider and the Repository are optional — when the provider has
// no client (nil) the relevant methods return ErrWorkspaceUnavailable so the
// HTTP layer can map it to a 503.
type Service struct {
	wsProv *workspace.Provider
	repo   *Repository
}

func NewService(wsProv *workspace.Provider, repo *Repository) *Service {
	return &Service{wsProv: wsProv, repo: repo}
}

// client returns the current workspace client or ErrWorkspaceUnavailable.
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

// ListLive fetches a single page of users straight from Workspace, opportunistically
// refreshing the local cache as a side effect.
func (s *Service) ListLive(ctx context.Context, pageToken string, pageSize int64) (*ListResponse, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	apiUsers, next, err := c.ListUsersPage(ctx, pageToken, pageSize)
	if err != nil {
		return nil, fmt.Errorf("users: list live: %w", err)
	}
	out := make([]User, 0, len(apiUsers))
	for _, u := range apiUsers {
		mapped := fromAPI(u)
		out = append(out, mapped)
		if s.repo != nil {
			raw, _ := json.Marshal(u)
			_ = s.repo.UpsertCache(ctx, &mapped, raw) // best-effort
		}
	}
	return &ListResponse{Users: out, NextPageToken: next}, nil
}

// ListCached returns a page from the local users_cache table.
func (s *Service) ListCached(ctx context.Context, limit, offset int) ([]User, int, error) {
	if s.repo == nil {
		return nil, 0, fmt.Errorf("users: repository not configured")
	}
	return s.repo.ListCache(ctx, limit, offset)
}

// Get fetches one user by ID or primary email.
func (s *Service) Get(ctx context.Context, key string) (*User, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	u, err := c.GetUser(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("users: get: %w", err)
	}
	out := fromAPI(u)
	return &out, nil
}

// Create provisions a new Workspace user.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*User, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	apiUser := &admin.User{
		PrimaryEmail: req.PrimaryEmail,
		Name: &admin.UserName{
			GivenName:  req.GivenName,
			FamilyName: req.FamilyName,
		},
		Password:    req.Password,
		OrgUnitPath: req.OrgUnitPath,
	}
	out, err := c.InsertUser(ctx, apiUser)
	if err != nil {
		return nil, fmt.Errorf("users: create: %w", err)
	}
	mapped := fromAPI(out)
	if s.repo != nil {
		raw, _ := json.Marshal(out)
		_ = s.repo.UpsertCache(ctx, &mapped, raw)
	}
	return &mapped, nil
}

// Update applies a partial update to an existing user.
func (s *Service) Update(ctx context.Context, key string, req UpdateRequest) (*User, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	patch := &admin.User{}
	var force []string

	if req.GivenName != nil || req.FamilyName != nil {
		patch.Name = &admin.UserName{}
		if req.GivenName != nil {
			patch.Name.GivenName = *req.GivenName
			patch.Name.ForceSendFields = append(patch.Name.ForceSendFields, "GivenName")
		}
		if req.FamilyName != nil {
			patch.Name.FamilyName = *req.FamilyName
			patch.Name.ForceSendFields = append(patch.Name.ForceSendFields, "FamilyName")
		}
	}
	if req.OrgUnitPath != nil {
		patch.OrgUnitPath = *req.OrgUnitPath
		force = append(force, "OrgUnitPath")
	}
	if req.Suspended != nil {
		patch.Suspended = *req.Suspended
		force = append(force, "Suspended")
	}
	patch.ForceSendFields = force

	out, err := c.UpdateUser(ctx, key, patch)
	if err != nil {
		return nil, fmt.Errorf("users: update %s: %w", key, err)
	}
	mapped := fromAPI(out)
	if s.repo != nil {
		raw, _ := json.Marshal(out)
		_ = s.repo.UpsertCache(ctx, &mapped, raw)
	}
	return &mapped, nil
}

// Delete permanently removes a user from Workspace and the local cache.
func (s *Service) Delete(ctx context.Context, key string) error {
	c, err := s.client()
	if err != nil {
		return err
	}
	if err := c.DeleteUser(ctx, key); err != nil {
		return fmt.Errorf("users: delete %s: %w", key, err)
	}
	if s.repo != nil {
		_ = s.repo.DeleteCache(ctx, key)
	}
	return nil
}

// fromAPI converts a *admin.User into the local User projection.
func fromAPI(u *admin.User) User {
	out := User{
		ID:           u.Id,
		PrimaryEmail: u.PrimaryEmail,
		OrgUnitPath:  u.OrgUnitPath,
		Suspended:    u.Suspended,
		IsAdmin:      u.IsAdmin,
	}
	if u.Name != nil {
		out.GivenName = u.Name.GivenName
		out.FamilyName = u.Name.FamilyName
	}
	return out
}
