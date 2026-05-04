// Package jobs defines the asynq task types and their handlers.
package jobs

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/hibiken/asynq"
)

// Task type identifiers — kept stable, used by the producer and the worker.
const (
	TypeUsersExport       = "users:export"
	TypeUsersBulkCreate   = "users:bulk_create"
	TypeUsersBulkUpdate   = "users:bulk_update"
	TypeUsersBulkSuspend  = "users:bulk_suspend"
	TypeUsersBulkDelete   = "users:bulk_delete"
	TypeGroupsBulkCreate  = "groups:bulk_create"
	TypeGroupsBulkDelete  = "groups:bulk_delete"
	TypeMembersBulkAdd    = "members:bulk_add"
	TypeMembersBulkRemove = "members:bulk_remove"
	TypeAliasesBulkAdd    = "aliases:bulk_add"
	TypeAliasesBulkRemove = "aliases:bulk_remove"
)

// BulkPayload is the common shape used by every bulk task. JobID points at the
// `jobs` row in Postgres; Rows are opaque per-row payloads.
type BulkPayload struct {
	JobID string            `json:"job_id"`
	Rows  []json.RawMessage `json:"rows"`
}

// Register wires every task type to a handler on the provided ServeMux.
func Register(mux *asynq.ServeMux, logger *slog.Logger) {
	h := &Handlers{logger: logger}
	mux.HandleFunc(TypeUsersExport, h.notImplemented)
	mux.HandleFunc(TypeUsersBulkCreate, h.notImplemented)
	mux.HandleFunc(TypeUsersBulkUpdate, h.notImplemented)
	mux.HandleFunc(TypeUsersBulkSuspend, h.notImplemented)
	mux.HandleFunc(TypeUsersBulkDelete, h.notImplemented)
	mux.HandleFunc(TypeGroupsBulkCreate, h.notImplemented)
	mux.HandleFunc(TypeGroupsBulkDelete, h.notImplemented)
	mux.HandleFunc(TypeMembersBulkAdd, h.notImplemented)
	mux.HandleFunc(TypeMembersBulkRemove, h.notImplemented)
	mux.HandleFunc(TypeAliasesBulkAdd, h.notImplemented)
	mux.HandleFunc(TypeAliasesBulkRemove, h.notImplemented)
}

// Handlers groups the implementation of every task. Initially every task is a
// no-op stub; concrete logic lands in this package as the domains are built.
type Handlers struct{ logger *slog.Logger }

func (h *Handlers) notImplemented(_ context.Context, t *asynq.Task) error {
	h.logger.Warn("task handler not implemented", "type", t.Type())
	return nil
}
