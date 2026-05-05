package orgunits

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/haydary1986/ok-goldy-alternative/internal/api"
	"github.com/haydary1986/ok-goldy-alternative/internal/audit"
	"github.com/haydary1986/ok-goldy-alternative/internal/workspace"
)

// Handler exposes the orgunits endpoints.
type Handler struct {
	svc   *Service
	audit *audit.Service
}

func NewHandler(svc *Service, auditSvc *audit.Service) *Handler {
	return &Handler{svc: svc, audit: auditSvc}
}

func (h *Handler) Routes() func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.List(r.Context())
	if err != nil {
		h.writeErr(w, err)
		return
	}
	api.WriteData(w, http.StatusOK, out)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	out, err := h.svc.Create(r.Context(), req)
	h.recordAudit(r, audit.ActionCreate, "org_unit", req.ParentOrgUnitPath+"/"+req.Name, nil, out, err)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	api.WriteData(w, http.StatusCreated, out)
}

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
