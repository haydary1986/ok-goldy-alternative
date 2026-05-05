package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/haydary1986/ok-goldy-alternative/internal/api"
	"github.com/haydary1986/ok-goldy-alternative/internal/audit"
	"github.com/haydary1986/ok-goldy-alternative/internal/workspace"
)

// Handler exposes Workspace user operations over HTTP. The audit service is
// optional — when nil, audit recording is silently skipped.
type Handler struct {
	svc   *Service
	audit *audit.Service
}

func NewHandler(svc *Service, auditSvc *audit.Service) *Handler {
	return &Handler{svc: svc, audit: auditSvc}
}

// Routes returns a function suitable for chi.Router.Route("/users", ...).
func (h *Handler) Routes() func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.Get)
		r.Patch("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)

		// Inactive accounts finder + bulk suspend (the killer Ok Goldy flow).
		r.Get("/inactive", h.Inactive)
		r.Post("/bulk/suspend", h.BulkSuspend)

		// Bulk endpoints (CSV import / export) — wired later, return 501 today.
		r.Post("/bulk/import", h.notImplemented)
		r.Get("/bulk/export", h.notImplemented)

		// Aliases sub-resource (handlers live in aliases.go).
		r.Route("/{id}/aliases", func(r chi.Router) {
			r.Get("/", h.listAliases)
			r.Post("/", h.addAlias)
			r.Delete("/{alias}", h.deleteAlias)
		})
	}
}

// --- handlers ---

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	pageToken := r.URL.Query().Get("page_token")
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 64)
	resp, err := h.svc.ListLive(r.Context(), pageToken, pageSize)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	api.WriteData(w, http.StatusOK, resp)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	u, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	api.WriteData(w, http.StatusOK, u)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	u, err := h.svc.Create(r.Context(), req)
	h.recordAudit(r, audit.ActionCreate, req.PrimaryEmail, nil, u, err)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	api.WriteData(w, http.StatusCreated, u)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	u, err := h.svc.Update(r.Context(), id, req)
	h.recordAudit(r, audit.ActionUpdate, id, nil, u, err)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	api.WriteData(w, http.StatusOK, u)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.svc.Delete(r.Context(), id)
	h.recordAudit(r, audit.ActionDelete, id, nil, nil, err)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	api.WriteData(w, http.StatusOK, map[string]any{"id": id, "deleted": true})
}

func (h *Handler) Inactive(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	days, _ := strconv.Atoi(q.Get("days"))
	includeAdmins := q.Get("include_admins") == "true"
	includeSuspended := q.Get("include_suspended") == "true"
	out, err := h.svc.Inactive(r.Context(), InactiveQuery{
		Days:             days,
		IncludeAdmins:    includeAdmins,
		IncludeSuspended: includeSuspended,
	})
	if err != nil {
		h.writeErr(w, err)
		return
	}
	api.WriteData(w, http.StatusOK, out)
}

func (h *Handler) BulkSuspend(w http.ResponseWriter, r *http.Request) {
	var req BulkSuspendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if len(req.UserIDs) == 0 {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "user_ids is required")
		return
	}
	if len(req.UserIDs) > 5000 {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "max 5000 users per request")
		return
	}
	resp := h.svc.BulkSuspend(r.Context(), req.UserIDs, req.Suspended)
	action := audit.ActionRestore
	if req.Suspended {
		action = audit.ActionSuspend
	}
	h.recordAudit(r, action, fmt.Sprintf("bulk:%d/%d ok", resp.Successful, resp.Total),
		nil, map[string]any{"successful": resp.Successful, "failed": resp.Failed}, nil)
	api.WriteData(w, http.StatusOK, resp)
}

func (h *Handler) notImplemented(w http.ResponseWriter, _ *http.Request) {
	api.WriteError(w, http.StatusNotImplemented, "not_implemented", "this endpoint is not implemented yet")
}

// --- helpers ---

// writeErr maps domain errors to HTTP responses.
func (h *Handler) writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrWorkspaceUnavailable):
		api.WriteError(w, http.StatusServiceUnavailable, "workspace_unavailable",
			"Google Workspace client is not configured. Set GOLDY_GOOGLE_DELEGATED_ADMIN and provide a service-account key file.")
	case IsInvalid(err):
		api.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
	default:
		if code := workspace.HTTPStatus(err); code != 0 {
			api.WriteError(w, code, "workspace_error", err.Error())
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error())
	}
}

// recordAudit fires off an audit_log row. Failures are non-fatal.
func (h *Handler) recordAudit(r *http.Request, action, resourceID string, before, after any, opErr error) {
	if h.audit == nil {
		return
	}
	e := audit.Entry{
		Actor:        api.Actor(r),
		Action:       action,
		ResourceType: audit.ResourceUser,
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
