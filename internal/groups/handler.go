package groups

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/haydary1986/ok-goldy-alternative/internal/api"
	"github.com/haydary1986/ok-goldy-alternative/internal/audit"
	"github.com/haydary1986/ok-goldy-alternative/internal/workspace"
)

// Handler exposes Group + Member operations over HTTP.
type Handler struct {
	svc   *Service
	audit *audit.Service
}

func NewHandler(svc *Service, auditSvc *audit.Service) *Handler {
	return &Handler{svc: svc, audit: auditSvc}
}

// Routes returns a function suitable for chi.Router.Route("/groups", ...).
func (h *Handler) Routes() func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.Get)
		r.Delete("/{id}", h.Delete)

		r.Route("/{id}/members", func(r chi.Router) {
			r.Get("/", h.ListMembers)
			r.Post("/", h.AddMember)
			r.Delete("/{member}", h.RemoveMember)
		})
	}
}

// --- group handlers ---

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	pageToken := r.URL.Query().Get("page_token")
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 64)
	resp, err := h.svc.List(r.Context(), pageToken, pageSize)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	api.WriteData(w, http.StatusOK, resp)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	g, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.writeErr(w, err)
		return
	}
	api.WriteData(w, http.StatusOK, g)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	g, err := h.svc.Create(r.Context(), req)
	h.recordAudit(r, audit.ActionCreate, audit.ResourceGroup, req.Email, nil, g, err)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	api.WriteData(w, http.StatusCreated, g)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.svc.Delete(r.Context(), id)
	h.recordAudit(r, audit.ActionDelete, audit.ResourceGroup, id, nil, nil, err)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	api.WriteData(w, http.StatusOK, map[string]any{"id": id, "deleted": true})
}

// --- member handlers ---

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	groupKey := chi.URLParam(r, "id")
	pageToken := r.URL.Query().Get("page_token")
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 64)
	resp, err := h.svc.ListMembers(r.Context(), groupKey, pageToken, pageSize)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	api.WriteData(w, http.StatusOK, resp)
}

func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	groupKey := chi.URLParam(r, "id")
	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	m, err := h.svc.AddMember(r.Context(), groupKey, req)
	h.recordAudit(r, audit.ActionCreate, audit.ResourceMember, groupKey+"/"+req.Email, nil, m, err)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	api.WriteData(w, http.StatusCreated, m)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	groupKey := chi.URLParam(r, "id")
	memberKey := chi.URLParam(r, "member")
	err := h.svc.RemoveMember(r.Context(), groupKey, memberKey)
	h.recordAudit(r, audit.ActionDelete, audit.ResourceMember, groupKey+"/"+memberKey, nil, nil, err)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	api.WriteData(w, http.StatusOK, map[string]any{"group": groupKey, "member": memberKey, "removed": true})
}

// --- helpers ---

func (h *Handler) writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrWorkspaceUnavailable):
		api.WriteError(w, http.StatusServiceUnavailable, "workspace_unavailable",
			"Google Workspace client is not configured.")
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

func (h *Handler) recordAudit(r *http.Request, action, resourceType, resourceID string, before, after any, opErr error) {
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
