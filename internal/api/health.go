package api

import (
	"context"
	"net/http"
	"time"
)

func healthz(_ Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		WriteData(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func readyz(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := deps.DB.Ping(ctx); err != nil {
			WriteError(w, http.StatusServiceUnavailable, "db_unreachable", err.Error())
			return
		}
		WriteData(w, http.StatusOK, map[string]string{"status": "ready"})
	}
}
