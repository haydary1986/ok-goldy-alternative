package wsadmin

import (
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/haydary1986/ok-goldy-alternative/internal/api"
	"github.com/haydary1986/ok-goldy-alternative/internal/audit"
)

// Maximum upload size for the SA JSON file (typical SA keys are 2–3 KB).
const maxUploadBytes = 256 * 1024

// Handler exposes the workspace credentials admin endpoints.
type Handler struct {
	svc   *Service
	audit *audit.Service
}

func NewHandler(svc *Service, auditSvc *audit.Service) *Handler {
	return &Handler{svc: svc, audit: auditSvc}
}

// Routes mounts the admin sub-router under /api/v1/admin.
func (h *Handler) Routes() func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/workspace/status", h.status)
		r.Post("/workspace/credentials", h.upload)
		r.Delete("/workspace/credentials", h.delete)
		r.Post("/workspace/test", h.test)
		r.Post("/workspace/diagnostic", h.diagnostic)
	}
}

func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.Status(r.Context())
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	api.WriteData(w, http.StatusOK, out)
}

func (h *Handler) upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes*2)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_form", err.Error())
		return
	}

	delegatedAdmin := r.FormValue("delegated_admin")
	customerID := r.FormValue("customer_id")

	file, _, err := r.FormFile("file")
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, "missing_file", "field 'file' is required (the service-account JSON)")
		return
	}
	defer file.Close()

	saJSON, err := io.ReadAll(io.LimitReader(file, maxUploadBytes))
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, "read_failed", err.Error())
		return
	}

	creds, err := h.svc.Save(r.Context(), UploadRequest{
		SAJSON:         saJSON,
		DelegatedAdmin: delegatedAdmin,
		CustomerID:     customerID,
	})
	h.recordAudit(r, audit.ActionUpdate, "workspace_credentials", "singleton", nil, creds, err)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	api.WriteData(w, http.StatusOK, creds)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	err := h.svc.Delete(r.Context())
	h.recordAudit(r, audit.ActionDelete, "workspace_credentials", "singleton", nil, nil, err)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	api.WriteData(w, http.StatusOK, map[string]any{"deleted": true})
}

func (h *Handler) test(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Test(r.Context()); err != nil {
		if errors.Is(err, ErrNotConfigured) {
			api.WriteError(w, http.StatusServiceUnavailable, "not_configured", err.Error())
			return
		}
		api.WriteError(w, http.StatusBadGateway, "test_failed", err.Error())
		return
	}
	api.WriteData(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) diagnostic(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.Diagnostic(r.Context())
	if err != nil {
		if errors.Is(err, ErrNotConfigured) {
			api.WriteError(w, http.StatusServiceUnavailable, "not_configured", err.Error())
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	api.WriteData(w, http.StatusOK, out)
}

func (h *Handler) writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotConfigured):
		api.WriteError(w, http.StatusServiceUnavailable, "not_configured", err.Error())
	case IsInvalid(err):
		api.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
	default:
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
