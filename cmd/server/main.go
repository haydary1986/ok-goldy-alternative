// Command server is the HTTP API entrypoint for Ok Goldy Alternative.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/haydary1986/ok-goldy-alternative/internal/api"
	"github.com/haydary1986/ok-goldy-alternative/internal/audit"
	"github.com/haydary1986/ok-goldy-alternative/internal/config"
	"github.com/haydary1986/ok-goldy-alternative/internal/db"
	"github.com/haydary1986/ok-goldy-alternative/internal/groups"
	applog "github.com/haydary1986/ok-goldy-alternative/internal/log"
	"github.com/haydary1986/ok-goldy-alternative/internal/users"
	"github.com/haydary1986/ok-goldy-alternative/internal/workspace"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		applog.Bootstrap().Error("config load failed", "err", err)
		os.Exit(1)
	}

	logger := applog.New(cfg.LogLevel, cfg.LogFormat)
	logger.Info("starting Ok Goldy Alternative server", "addr", cfg.HTTPAddr)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Workspace client is best-effort: a missing service-account file should
	// not block the server from starting (it simply causes affected endpoints
	// to return 503). This makes local dev possible before SA setup is done.
	wsClient := buildWorkspaceClient(ctx, cfg, logger)

	auditSvc := audit.New(pool)

	usersRepo := users.NewRepository(pool)
	usersSvc := users.NewService(wsClient, usersRepo)
	usersHandler := users.NewHandler(usersSvc, auditSvc)

	groupsSvc := groups.NewService(wsClient)
	groupsHandler := groups.NewHandler(groupsSvc, auditSvc)

	router := api.NewRouter(api.Deps{
		Logger:       logger,
		DB:           pool,
		Config:       cfg,
		UsersRoutes:  usersHandler.Routes(),
		GroupsRoutes: groupsHandler.Routes(),
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("http listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server error", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
	}
	logger.Info("server stopped")
}

// buildWorkspaceClient tries to construct the Admin SDK client. On failure it
// logs a warning and returns nil so the server can still start.
func buildWorkspaceClient(ctx context.Context, cfg *config.Config, logger interface {
	Warn(msg string, args ...any)
	Info(msg string, args ...any)
}) *workspace.Client {
	if cfg.GoogleDelegatedAdmin == "" {
		logger.Warn("workspace not configured",
			"hint", "set GOLDY_GOOGLE_DELEGATED_ADMIN and provide a service-account key to enable Workspace endpoints")
		return nil
	}
	client, err := workspace.New(ctx, workspace.Config{
		ServiceAccountKeyFile: cfg.GoogleSAKeyFile,
		DelegatedAdmin:        cfg.GoogleDelegatedAdmin,
		CustomerID:            cfg.GoogleCustomerID,
		RateLimitRPS:          cfg.RateLimitRPS,
		RateLimitBurst:        cfg.RateLimitBurst,
	})
	if err != nil {
		logger.Warn("workspace client unavailable", "err", err)
		return nil
	}
	logger.Info("workspace client ready",
		"delegated_admin", cfg.GoogleDelegatedAdmin,
		"customer_id", cfg.GoogleCustomerID,
		"rate_rps", cfg.RateLimitRPS,
	)
	return client
}
