package wsadmin

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository owns the singleton workspace_credentials table.
type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// Get returns the current credentials row, or (nil, nil) if no row exists.
func (r *Repository) Get(ctx context.Context) (*Credentials, error) {
	const sqlStmt = `
		SELECT sa_json, delegated_admin, customer_id, sa_email, project_id, updated_at
		FROM workspace_credentials
		WHERE id = 1
	`
	var c Credentials
	var saEmail, projectID *string
	err := r.pool.QueryRow(ctx, sqlStmt).Scan(
		&c.SAJSON, &c.DelegatedAdmin, &c.CustomerID, &saEmail, &projectID, &c.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("wsadmin: get credentials: %w", err)
	}
	if saEmail != nil {
		c.SAEmail = *saEmail
	}
	if projectID != nil {
		c.ProjectID = *projectID
	}
	return &c, nil
}

// Upsert writes credentials to the singleton row.
func (r *Repository) Upsert(ctx context.Context, c *Credentials) error {
	const sqlStmt = `
		INSERT INTO workspace_credentials
			(id, sa_json, delegated_admin, customer_id, sa_email, project_id, updated_at)
		VALUES (1, $1, $2, $3, $4, $5, NOW())
		ON CONFLICT (id) DO UPDATE SET
			sa_json         = EXCLUDED.sa_json,
			delegated_admin = EXCLUDED.delegated_admin,
			customer_id     = EXCLUDED.customer_id,
			sa_email        = EXCLUDED.sa_email,
			project_id      = EXCLUDED.project_id,
			updated_at      = NOW()
	`
	if _, err := r.pool.Exec(ctx, sqlStmt,
		c.SAJSON, c.DelegatedAdmin, c.CustomerID, nullableString(c.SAEmail), nullableString(c.ProjectID),
	); err != nil {
		return fmt.Errorf("wsadmin: upsert credentials: %w", err)
	}
	return nil
}

// Delete removes the credentials row, if any. Missing rows are not an error.
func (r *Repository) Delete(ctx context.Context) error {
	if _, err := r.pool.Exec(ctx, `DELETE FROM workspace_credentials WHERE id = 1`); err != nil {
		return fmt.Errorf("wsadmin: delete credentials: %w", err)
	}
	return nil
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
