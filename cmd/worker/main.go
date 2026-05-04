// Command worker runs the asynq background worker that executes bulk
// Workspace operations.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"

	"github.com/haydary1986/ok-goldy-alternative/internal/config"
	"github.com/haydary1986/ok-goldy-alternative/internal/jobs"
	applog "github.com/haydary1986/ok-goldy-alternative/internal/log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		applog.Bootstrap().Error("config load failed", "err", err)
		os.Exit(1)
	}
	logger := applog.New(cfg.LogLevel, cfg.LogFormat)
	logger.Info("starting worker", "concurrency", cfg.WorkerConcurrency, "redis", cfg.RedisAddr)

	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		},
		asynq.Config{
			Concurrency: cfg.WorkerConcurrency,
			Queues:      map[string]int{"default": 5, "bulk": 3, "low": 1},
			Logger:      &asynqLogAdapter{l: logger},
		},
	)

	mux := asynq.NewServeMux()
	jobs.Register(mux, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		logger.Info("worker shutting down")
		srv.Shutdown()
	}()

	if err := srv.Run(mux); err != nil {
		logger.Error("worker exited", "err", err)
		os.Exit(1)
	}
}

// asynqLogAdapter bridges asynq's logger interface onto slog.
type asynqLogAdapter struct{ l *slog.Logger }

func (a *asynqLogAdapter) Debug(args ...any) { a.l.Debug(fmt.Sprint(args...)) }
func (a *asynqLogAdapter) Info(args ...any)  { a.l.Info(fmt.Sprint(args...)) }
func (a *asynqLogAdapter) Warn(args ...any)  { a.l.Warn(fmt.Sprint(args...)) }
func (a *asynqLogAdapter) Error(args ...any) { a.l.Error(fmt.Sprint(args...)) }
func (a *asynqLogAdapter) Fatal(args ...any) { a.l.Error(fmt.Sprint(args...)) }
