package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/haydary1986/ok-goldy-alternative/internal/api"
	"github.com/haydary1986/ok-goldy-alternative/internal/audit"
)

// Alias is the Goldy projection of a Workspace user alias.
type Alias struct {
	Alias string `json:"alias"`
}

// AddAliasRequest is the JSON body for POST /api/v1/users/{id}/aliases.
type AddAliasRequest struct {
	Alias string `json:"alias"`
}

func (r AddAliasRequest) Validate() error {
	if r.Alias == "" {
		return ErrInvalid("alias is required")
	}
	return nil
}

// --- service methods ---

func (s *Service) ListAliases(ctx context.Context, userKey string) ([]Alias, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	items, err := c.ListAliases(ctx, userKey)
	if err != nil {
		return nil, fmt.Errorf("users: list aliases: %w", err)
	}
	out := make([]Alias, 0, len(items))
	for _, a := range items {
		out = append(out, Alias{Alias: a.Alias})
	}
	return out, nil
}

func (s *Service) AddAlias(ctx context.Context, userKey string, req AddAliasRequest) (*Alias, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	created, err := c.InsertAlias(ctx, userKey, req.Alias)
	if err != nil {
		return nil, fmt.Errorf("users: add alias: %w", err)
	}
	if created == "" {
		created = req.Alias
	}
	return &Alias{Alias: created}, nil
}

func (s *Service) DeleteAlias(ctx context.Context, userKey, alias string) error {
	c, err := s.client()
	if err != nil {
		return err
	}
	if err := c.DeleteAlias(ctx, userKey, alias); err != nil {
		return fmt.Errorf("users: delete alias: %w", err)
	}
	return nil
}

// --- HTTP handlers ---

func (h *Handler) listAliases(w http.ResponseWriter, r *http.Request) {
	userKey := chi.URLParam(r, "id")
	out, err := h.svc.ListAliases(r.Context(), userKey)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	api.WriteData(w, http.StatusOK, map[string]any{"aliases": out})
}

func (h *Handler) addAlias(w http.ResponseWriter, r *http.Request) {
	userKey := chi.URLParam(r, "id")
	var req AddAliasRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	a, err := h.svc.AddAlias(r.Context(), userKey, req)
	h.recordResourceAudit(r, audit.ActionCreate, audit.ResourceAlias, userKey+"/"+req.Alias, nil, a, err)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	api.WriteData(w, http.StatusCreated, a)
}

func (h *Handler) deleteAlias(w http.ResponseWriter, r *http.Request) {
	userKey := chi.URLParam(r, "id")
	alias := chi.URLParam(r, "alias")
	err := h.svc.DeleteAlias(r.Context(), userKey, alias)
	h.recordResourceAudit(r, audit.ActionDelete, audit.ResourceAlias, userKey+"/"+alias, nil, nil, err)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	api.WriteData(w, http.StatusOK, map[string]any{"user": userKey, "alias": alias, "deleted": true})
}

// recordResourceAudit is the multi-resource counterpart to recordAudit.
func (h *Handler) recordResourceAudit(r *http.Request, action, resourceType, resourceID string, before, after any, opErr error) {
	if h.audit == nil {
		return
	}
	e := audit.Entry{
		Actor:        api.Actor(r),
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		RequestID:    api.RequestID(r),
		Before:       before,
		After:        after,
		OK:           opErr == nil,
	}
	if opErr != nil {
		e.ErrorMessage = opErr.Error()
	}
	_ = h.audit.Log(r.Context(), e)
}
