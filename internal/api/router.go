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

// RouteRegistrar is the function shape used to mount a domain's sub-router
// without creating a circular import between api and the domain packages.
type RouteRegistrar func(chi.Router)

// Deps groups the dependencies handed to the router. RouteRegistrar fields
// are optional — when nil, the matching route group falls back to a 501 stub
// so the OpenAPI surface stays stable while features land.
//
// SPA, when set, is served on every path that doesn't match an explicit API
// or health route (via chi's NotFound handler). Unknown paths fall back to
// index.html so React Router can resolve them client-side.
type Deps struct {
	Logger *slog.Logger
	DB     *pgxpool.Pool
	Config *config.Config

	UsersRoutes    RouteRegistrar
	GroupsRoutes   RouteRegistrar
	OrgUnitsRoutes RouteRegistrar
	JobsRoutes     RouteRegistrar
	AuditRoutes    RouteRegistrar
	StatsRoutes    RouteRegistrar
	AdminRoutes    RouteRegistrar

	SPA http.Handler
}

// NewRouter builds the root chi router with middleware and v1 routes.
func NewRouter(deps Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(slogRequestLogger(deps.Logger))
	if deps.Config != nil && deps.Config.BasicAuthUser != "" && deps.Config.BasicAuthPassword != "" {
		r.Use(basicAuth(deps.Config.BasicAuthUser, deps.Config.BasicAuthPassword))
	}
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-Goldy-Actor"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", healthz(deps))
	r.Get("/readyz", readyz(deps))

	r.Route("/api/v1", func(r chi.Router) {
		mount(r, "/users", deps.UsersRoutes, stubUsers)
		mount(r, "/groups", deps.GroupsRoutes, stubGroups)
		if deps.OrgUnitsRoutes != nil {
			r.Route("/orgunits", deps.OrgUnitsRoutes)
		}
		mount(r, "/jobs", deps.JobsRoutes, stubJobs)
		mount(r, "/audit", deps.AuditRoutes, stubAudit)
		if deps.StatsRoutes != nil {
			r.Route("/stats", deps.StatsRoutes)
		}
		if deps.AdminRoutes != nil {
			r.Route("/admin", deps.AdminRoutes)
		}
	})

	// SPA fallback. Anything that did not match an explicit route above is
	// handed to the embedded React build (or 404 if no SPA was wired).
	if deps.SPA != nil {
		r.NotFound(deps.SPA.ServeHTTP)
	}

	return r
}

// mount registers `path` against either the supplied registrar or the fallback.
func mount(r chi.Router, path string, supplied, fallback RouteRegistrar) {
	if supplied != nil {
		r.Route(path, supplied)
	} else {
		r.Route(path, fallback)
	}
}

// --- fallback "not implemented" stubs ---

func stubUsers(r chi.Router) {
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
}

func stubGroups(r chi.Router) {
	r.Get("/", notImplemented)
	r.Post("/", notImplemented)
	r.Get("/{id}", notImplemented)
	r.Delete("/{id}", notImplemented)
	r.Route("/{id}/members", func(r chi.Router) {
		r.Get("/", notImplemented)
		r.Post("/", notImplemented)
		r.Delete("/{member}", notImplemented)
	})
}

func stubJobs(r chi.Router) {
	r.Get("/", notImplemented)
	r.Get("/{id}", notImplemented)
}

func stubAudit(r chi.Router) {
	r.Get("/", notImplemented)
}

func notImplemented(w http.ResponseWriter, _ *http.Request) {
	WriteError(w, http.StatusNotImplemented, "not_implemented", "this endpoint is not implemented yet")
}
