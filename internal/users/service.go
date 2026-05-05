package users

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

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

	allMu       sync.Mutex
	allCache    []User
	allCachedAt time.Time
}

// ListAllTTL is how long ListAll's snapshot is reused before re-walking
// the full Workspace user list. 30k users at the default rate cap takes
// ~3s, so a few minutes is the right sweet spot.
const ListAllTTL = 3 * time.Minute

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

// fromAPI converts a *admin.User into the local User projection. The
// Workspace API returns CreationTime / LastLoginTime as RFC3339 strings;
// we parse them here so the rest of the app — and the dashboard — can
// reason about activity windows. A "never logged in" user comes back
// with the year 1 sentinel, which downstream code treats as
// time.Time.IsZero() == true.
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
	if u.CreationTime != "" {
		if t, err := time.Parse(time.RFC3339, u.CreationTime); err == nil {
			out.CreationTime = t
		}
	}
	if u.LastLoginTime != "" {
		if t, err := time.Parse(time.RFC3339, u.LastLoginTime); err == nil {
			// Workspace returns 1970-01-01T00:00:00.000Z for users who
			// have never signed in. Map that to the Go zero value so
			// downstream code can use IsZero() uniformly.
			if t.Year() > 1970 {
				out.LastLoginTime = t
			}
		}
	}
	return out
}

// ListAll walks every page of users and returns the full list. The result
// is cached for ListAllTTL so /stats and /users/inactive don't re-walk
// 30k users on every dashboard refresh. Pass force=true to bypass.
func (s *Service) ListAll(ctx context.Context, force bool) ([]User, error) {
	if !force {
		s.allMu.Lock()
		if s.allCache != nil && time.Since(s.allCachedAt) < ListAllTTL {
			out := make([]User, len(s.allCache))
			copy(out, s.allCache)
			s.allMu.Unlock()
			return out, nil
		}
		s.allMu.Unlock()
	}

	c, err := s.client()
	if err != nil {
		return nil, err
	}

	var all []User
	pageToken := ""
	for {
		apiUsers, next, err := c.ListUsersPage(ctx, pageToken, 500)
		if err != nil {
			return nil, fmt.Errorf("users: list all: %w", err)
		}
		for _, u := range apiUsers {
			all = append(all, fromAPI(u))
		}
		if next == "" {
			break
		}
		pageToken = next
	}

	s.allMu.Lock()
	s.allCache = all
	s.allCachedAt = time.Now()
	s.allMu.Unlock()
	return all, nil
}

// InvalidateAllCache clears the ListAll snapshot — call it after writes
// so dashboards reflect fresh state on next read.
func (s *Service) InvalidateAllCache() {
	s.allMu.Lock()
	s.allCache = nil
	s.allMu.Unlock()
}

// Inactive returns users whose last_login_time is older than `days` days
// ago (or who have never logged in). Suspended users are excluded by
// default — that matches the typical "find candidates to suspend" flow.
func (s *Service) Inactive(ctx context.Context, q InactiveQuery) (*InactiveListResponse, error) {
	if q.Days <= 0 {
		q.Days = 90
	}
	cutoff := time.Now().AddDate(0, 0, -q.Days)

	all, err := s.ListAll(ctx, false)
	if err != nil {
		return nil, err
	}

	out := make([]User, 0)
	for _, u := range all {
		if !q.IncludeAdmins && u.IsAdmin {
			continue
		}
		if !q.IncludeSuspended && u.Suspended {
			continue
		}
		if u.LastLoginTime.IsZero() {
			out = append(out, u)
			continue
		}
		if u.LastLoginTime.Before(cutoff) {
			out = append(out, u)
		}
	}
	return &InactiveListResponse{
		Users:  out,
		Total:  len(out),
		Days:   q.Days,
		Cutoff: cutoff,
	}, nil
}

// BulkSuspend updates Suspended for a list of users in parallel (capped at
// 10 concurrent calls so we don't blow the Workspace rate limit). Returns
// per-user results; the call itself never errors short of validation.
func (s *Service) BulkSuspend(ctx context.Context, ids []string, suspended bool) *BulkSuspendResponse {
	resp := &BulkSuspendResponse{Total: len(ids), Results: make([]BulkSuspendResult, len(ids))}
	if len(ids) == 0 {
		return resp
	}

	c, err := s.client()
	if err != nil {
		for i, id := range ids {
			resp.Results[i] = BulkSuspendResult{UserID: id, OK: false, Error: err.Error()}
		}
		resp.Failed = len(ids)
		return resp
	}

	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup
	for i, id := range ids {
		wg.Add(1)
		go func(idx int, key string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			_, uErr := c.UpdateUser(ctx, key, &admin.User{
				Suspended:       suspended,
				ForceSendFields: []string{"Suspended"},
			})
			if uErr != nil {
				resp.Results[idx] = BulkSuspendResult{UserID: key, OK: false, Error: uErr.Error()}
			} else {
				resp.Results[idx] = BulkSuspendResult{UserID: key, OK: true}
			}
		}(i, id)
	}
	wg.Wait()

	for _, r := range resp.Results {
		if r.OK {
			resp.Successful++
		} else {
			resp.Failed++
		}
	}
	s.InvalidateAllCache()
	return resp
}
