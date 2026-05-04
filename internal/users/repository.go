package users

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository owns the local users_cache table.
type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// UpsertCache writes (or refreshes) one user row in the local cache. The raw
// JSON payload is stored alongside so we never lose Workspace fields we don't
// yet model in the User struct.
func (r *Repository) UpsertCache(ctx context.Context, u *User, raw []byte) error {
	const sqlStmt = `
		INSERT INTO users_cache
			(id, primary_email, given_name, family_name, org_unit_path, suspended, is_admin, raw, synced_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8, NOW())
		ON CONFLICT (id) DO UPDATE SET
			primary_email = EXCLUDED.primary_email,
			given_name    = EXCLUDED.given_name,
			family_name   = EXCLUDED.family_name,
			org_unit_path = EXCLUDED.org_unit_path,
			suspended     = EXCLUDED.suspended,
			is_admin      = EXCLUDED.is_admin,
			raw           = EXCLUDED.raw,
			synced_at     = NOW()
	`
	if _, err := r.pool.Exec(ctx, sqlStmt,
		u.ID, u.PrimaryEmail, u.GivenName, u.FamilyName, u.OrgUnitPath,
		u.Suspended, u.IsAdmin, raw,
	); err != nil {
		return fmt.Errorf("users: upsert cache: %w", err)
	}
	return nil
}

// DeleteCache removes a single user row from the local cache.
func (r *Repository) DeleteCache(ctx context.Context, id string) error {
	if _, err := r.pool.Exec(ctx, `DELETE FROM users_cache WHERE id = $1`, id); err != nil {
		return fmt.Errorf("users: delete cache: %w", err)
	}
	return nil
}

// ListCache returns a paginated slice of cached users plus the total count.
func (r *Repository) ListCache(ctx context.Context, limit, offset int) ([]User, int, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users_cache`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("users: count cache: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, primary_email, given_name, family_name, org_unit_path, suspended, is_admin
		FROM users_cache
		ORDER BY primary_email
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("users: list cache: %w", err)
	}
	defer rows.Close()

	out := make([]User, 0, limit)
	for rows.Next() {
		var u User
		var givenName, familyName, orgUnit *string
		if err := rows.Scan(&u.ID, &u.PrimaryEmail, &givenName, &familyName, &orgUnit, &u.Suspended, &u.IsAdmin); err != nil {
			return nil, 0, fmt.Errorf("users: scan cache row: %w", err)
		}
		if givenName != nil {
			u.GivenName = *givenName
		}
		if familyName != nil {
			u.FamilyName = *familyName
		}
		if orgUnit != nil {
			u.OrgUnitPath = *orgUnit
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("users: iterate cache rows: %w", err)
	}
	return out, total, nil
}
