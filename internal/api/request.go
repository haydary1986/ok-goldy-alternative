package api

import (
	"net/http"

	chimw "github.com/go-chi/chi/v5/middleware"
)

// Actor identifies who is performing the request. Until OIDC login is wired,
// callers can pass an X-Goldy-Actor header so audit rows aren't all "system".
func Actor(r *http.Request) string {
	if v := r.Header.Get("X-Goldy-Actor"); v != "" {
		return v
	}
	return "system"
}

// RequestID returns the chi-generated request ID for the given request.
func RequestID(r *http.Request) string {
	return chimw.GetReqID(r.Context())
}
