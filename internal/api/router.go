// Package api wires the HTTP router and shared response helpers.
package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/haydary1986/ok-goldy-alternative/internal/config"
)

// Deps groups the dependencies handed to the router.
type Deps struct {
	Logger *slog.Logger
	DB     *pgxpool.Pool
	Config *config.Config
}

// NewRouter builds the root chi router with middleware and v1 routes.
func NewRouter(deps Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(slogRequestLogger(deps.Logger))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", healthz(deps))
	r.Get("/readyz", readyz(deps))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/users", func(r chi.Router) {
			r.Get("/", notImplemented)
			r.Post("/", notImplemented)
			r.Get("/{id}", notImplemented)
			r.Patch("/{id}", notImplemented)
			r.Delete("/{id}", notImplemented)
			r.Post("/bulk/import", notImplemented)
			r.Get("/bulk/export", notImplemented)
			r.Route("/{id}/aliases", func(r chi.Router) {
				r.Get("/", notImplemented)
				r.Post("/", notImplemented)
				r.Delete("/{alias}", notImplemented)
			})
		})

		r.Route("/groups", func(r chi.Router) {
			r.Get("/", notImplemented)
			r.Post("/", notImplemented)
			r.Get("/{id}", notImplemented)
			r.Delete("/{id}", notImplemented)
			r.Route("/{id}/members", func(r chi.Router) {
				r.Get("/", notImplemented)
				r.Post("/", notImplemented)
				r.Delete("/{member}", notImplemented)
			})
		})

		r.Route("/jobs", func(r chi.Router) {
			r.Get("/", notImplemented)
			r.Get("/{id}", notImplemented)
		})

		r.Get("/audit", notImplemented)
	})

	return r
}

func notImplemented(w http.ResponseWriter, _ *http.Request) {
	WriteError(w, http.StatusNotImplemented, "not_implemented", "this endpoint is not implemented yet")
}
