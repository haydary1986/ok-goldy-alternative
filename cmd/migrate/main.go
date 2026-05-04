// Command migrate is a thin wrapper around pressly/goose to apply the SQL
// migrations under ./migrations against the configured Postgres instance.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/haydary1986/ok-goldy-alternative/internal/config"
)

const defaultMigrationsDir = "migrations"

func main() {
	if len(os.Args) < 2 {
		fail("usage: migrate <up|down|status|version|reset|create>")
	}
	command := os.Args[1]
	args := os.Args[2:]

	cfg, err := config.Load()
	if err != nil {
		fail("config: " + err.Error())
	}

	dir := defaultMigrationsDir
	if v := os.Getenv("GOLDY_MIGRATIONS_DIR"); v != "" {
		dir = v
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Container layout (Dockerfile copies migrations to /app/migrations).
		if _, e2 := os.Stat("/app/migrations"); e2 == nil {
			dir = "/app/migrations"
		}
	}

	d, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		fail("open db: " + err.Error())
	}
	defer d.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		fail("set dialect: " + err.Error())
	}

	if err := goose.RunContext(context.Background(), command, d, dir, args...); err != nil {
		fail(fmt.Sprintf("migrate %s: %v", command, err))
	}
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, "migrate: "+msg)
	os.Exit(1)
}
