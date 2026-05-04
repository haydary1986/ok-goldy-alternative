package api

import (
	"crypto/subtle"
	"net/http"
)

// basicAuth returns a middleware that requires HTTP Basic Auth credentials
// matching the configured user/password. Health endpoints are exempt so
// load balancers and Coolify health checks keep working.
//
// On a successful login, the middleware also sets X-Goldy-Actor (when not
// already set by the client) so the audit log records the right user.
func basicAuth(user, pass string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
				next.ServeHTTP(w, r)
				return
			}

			u, p, ok := r.BasicAuth()
			if !ok ||
				subtle.ConstantTimeCompare([]byte(u), []byte(user)) != 1 ||
				subtle.ConstantTimeCompare([]byte(p), []byte(pass)) != 1 {
				w.Header().Set("WWW-Authenticate", `Basic realm="Ok Goldy Alternative"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if r.Header.Get("X-Goldy-Actor") == "" {
				r.Header.Set("X-Goldy-Actor", u)
			}
			next.ServeHTTP(w, r)
		})
	}
}
