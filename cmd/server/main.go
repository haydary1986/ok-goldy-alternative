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
	"github.com/haydary1986/ok-goldy-alternative/internal/wsadmin"
	"github.com/haydary1986/ok-goldy-alternative/web"
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

	// Workspace credentials come from one of three sources, in order of
	// precedence: (1) the workspace_credentials DB row populated by an admin
	// uploading a service-account JSON via Settings, (2) the GOLDY_GOOGLE_*
	// env vars if a key file is mounted, (3) nothing — affected endpoints
	// return 503 until an admin configures Goldy.
	wsProv := workspace.NewProvider(nil)
	wsCredsRepo := wsadmin.NewRepository(pool)
	if err := loadWorkspaceFromDB(ctx, wsCredsRepo, wsProv, cfg, logger); err != nil {
		logger.Warn("failed to load workspace credentials from DB", "err", err)
	}
	if wsProv.Get() == nil {
		if c := buildWorkspaceClientFromEnv(ctx, cfg, logger); c != nil {
			wsProv.Set(c)
		}
	}

	auditSvc := audit.New(pool)

	usersRepo := users.NewRepository(pool)
	usersSvc := users.NewService(wsProv, usersRepo)
	usersHandler := users.NewHandler(usersSvc, auditSvc)

	groupsSvc := groups.NewService(wsProv)
	groupsHandler := groups.NewHandler(groupsSvc, auditSvc)

	wsadminSvc := wsadmin.NewService(wsCredsRepo, wsProv, cfg.RateLimitRPS, cfg.RateLimitBurst)
	wsadminHandler := wsadmin.NewHandler(wsadminSvc, auditSvc)

	spaHandler, err := web.Handler()
	if err != nil {
		logger.Warn("SPA handler unavailable", "err", err)
	}

	router := api.NewRouter(api.Deps{
		Logger:       logger,
		DB:           pool,
		Config:       cfg,
		UsersRoutes:  usersHandler.Routes(),
		GroupsRoutes: groupsHandler.Routes(),
		AdminRoutes:  wsadminHandler.Routes(),
		SPA:          spaHandler,
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

// loadWorkspaceFromDB hydrates the provider from the workspace_credentials
// row, if one exists. Missing rows are not an error.
func loadWorkspaceFromDB(ctx context.Context, repo *wsadmin.Repository, prov *workspace.Provider, cfg *config.Config, logger interface {
	Warn(msg string, args ...any)
	Info(msg string, args ...any)
}) error {
	creds, err := repo.Get(ctx)
	if err != nil {
		return err
	}
	if creds == nil {
		return nil
	}
	client, err := workspace.New(ctx, workspace.Config{
		ServiceAccountKey: creds.SAJSON,
		DelegatedAdmin:    creds.DelegatedAdmin,
		CustomerID:        creds.CustomerID,
		RateLimitRPS:      cfg.RateLimitRPS,
		RateLimitBurst:    cfg.RateLimitBurst,
	})
	if err != nil {
		return err
	}
	prov.Set(client)
	logger.Info("workspace client loaded from database",
		"delegated_admin", creds.DelegatedAdmin,
		"customer_id", creds.CustomerID,
		"sa_email", creds.SAEmail,
	)
	return nil
}

// buildWorkspaceClientFromEnv tries to construct the Admin SDK client from
// environment variables. On failure it logs a warning and returns nil.
func buildWorkspaceClientFromEnv(ctx context.Context, cfg *config.Config, logger interface {
	Warn(msg string, args ...any)
	Info(msg string, args ...any)
}) *workspace.Client {
	if cfg.GoogleDelegatedAdmin == "" {
		logger.Warn("workspace not configured",
			"hint", "upload a service-account JSON via /settings or set GOLDY_GOOGLE_DELEGATED_ADMIN env var")
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
