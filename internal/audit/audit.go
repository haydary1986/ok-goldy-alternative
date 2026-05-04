// Package audit records every mutation performed by Goldy so org admins have
// a complete who/what/when trail.
package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Action constants for common mutations.
const (
	ActionCreate  = "create"
	ActionUpdate  = "update"
	ActionDelete  = "delete"
	ActionSuspend = "suspend"
	ActionRestore = "restore"
	ActionExport  = "export"
)

// Resource type constants.
const (
	ResourceUser   = "user"
	ResourceGroup  = "group"
	ResourceMember = "group_member"
	ResourceAlias  = "user_alias"
)

// Entry is a single audit record. Before/After hold arbitrary JSON-marshalable
// snapshots of the resource state.
type Entry struct {
	Actor        string
	Action       string
	ResourceType string
	ResourceID   string
	RequestID    string
	Before       any
	After        any
	OK           bool
	ErrorMessage string
}

// Service writes audit entries to the audit_log table.
type Service struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Service { return &Service{pool: pool} }

// Log records the entry. Marshalling errors on Before/After are swallowed: an
// audit row is more valuable than a perfect snapshot.
func (s *Service) Log(ctx context.Context, e Entry) error {
	var beforeJSON, afterJSON []byte
	if e.Before != nil {
		beforeJSON, _ = json.Marshal(e.Before)
	}
	if e.After != nil {
		afterJSON, _ = json.Marshal(e.After)
	}
	const sqlStmt = `
		INSERT INTO audit_log
			(actor, action, resource_type, resource_id, request_id, before, after, ok, error_message)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`
	if _, err := s.pool.Exec(ctx, sqlStmt,
		e.Actor, e.Action, e.ResourceType, e.ResourceID, e.RequestID,
		beforeJSON, afterJSON, e.OK, e.ErrorMessage,
	); err != nil {
		return fmt.Errorf("audit: insert: %w", err)
	}
	return nil
}
