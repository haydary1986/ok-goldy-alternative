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
	"github.com/haydary1986/ok-goldy-alternative/internal/config"
	"github.com/haydary1986/ok-goldy-alternative/internal/db"
	applog "github.com/haydary1986/ok-goldy-alternative/internal/log"
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

	router := api.NewRouter(api.Deps{Logger: logger, DB: pool, Config: cfg})

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
