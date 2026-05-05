package audit

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/haydary1986/ok-goldy-alternative/internal/api"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Routes() func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", h.List)
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	out, err := h.svc.List(r.Context(), ListQuery{
		Actor:        q.Get("actor"),
		Action:       q.Get("action"),
		ResourceType: q.Get("resource_type"),
		OnlyFailures: q.Get("only_failures") == "true",
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	api.WriteData(w, http.StatusOK, out)
}
