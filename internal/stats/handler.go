package stats

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/haydary1986/ok-goldy-alternative/internal/api"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Routes() func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/overview", h.Overview)
	}
}

func (h *Handler) Overview(w http.ResponseWriter, r *http.Request) {
	force := r.URL.Query().Get("force") == "true"
	out, err := h.svc.Overview(r.Context(), force)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	api.WriteData(w, http.StatusOK, out)
}
