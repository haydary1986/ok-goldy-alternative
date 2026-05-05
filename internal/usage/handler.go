package usage

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/haydary1986/ok-goldy-alternative/internal/api"
	"github.com/haydary1986/ok-goldy-alternative/internal/workspace"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Routes() func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/users", h.Snapshot)
	}
}

func (h *Handler) Snapshot(w http.ResponseWriter, r *http.Request) {
	force := r.URL.Query().Get("force") == "true"
	out, err := h.svc.Snapshot(r.Context(), force)
	if err != nil {
		switch {
		case errors.Is(err, ErrUnavailable):
			api.WriteError(w, http.StatusServiceUnavailable, "workspace_unavailable",
				"Workspace not configured. Upload a service account JSON via /settings.")
		default:
			code := workspace.HTTPStatus(err)
			if code == 0 {
				code = http.StatusBadGateway
			}
			api.WriteError(w, code, "usage_failed", err.Error())
		}
		return
	}
	api.WriteData(w, http.StatusOK, out)
}
