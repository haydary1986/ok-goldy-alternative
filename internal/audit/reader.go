package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// LogEntry is a single row read from the audit_log table.
type LogEntry struct {
	ID           int64           `json:"id"`
	OccurredAt   time.Time       `json:"occurred_at"`
	Actor        string          `json:"actor"`
	Action       string          `json:"action"`
	ResourceType string          `json:"resource_type"`
	ResourceID   string          `json:"resource_id"`
	RequestID    string          `json:"request_id,omitempty"`
	Before       json.RawMessage `json:"before,omitempty"`
	After        json.RawMessage `json:"after,omitempty"`
	OK           bool            `json:"ok"`
	ErrorMessage string          `json:"error_message,omitempty"`
}

// ListQuery shapes a GET /api/v1/audit request.
type ListQuery struct {
	Actor        string
	Action       string
	ResourceType string
	OnlyFailures bool
	Limit        int
	Offset       int
}

// ListResponse is the paginated read result.
type ListResponse struct {
	Entries []LogEntry `json:"entries"`
	Total   int        `json:"total"`
	Limit   int        `json:"limit"`
	Offset  int        `json:"offset"`
}

// List returns audit log entries matching the query, newest first.
func (s *Service) List(ctx context.Context, q ListQuery) (*ListResponse, error) {
	if q.Limit <= 0 || q.Limit > 1000 {
		q.Limit = 100
	}
	if q.Offset < 0 {
		q.Offset = 0
	}

	where := []string{"1=1"}
	args := []any{}

	if q.Actor != "" {
		args = append(args, q.Actor)
		where = append(where, fmt.Sprintf("actor = $%d", len(args)))
	}
	if q.Action != "" {
		args = append(args, q.Action)
		where = append(where, fmt.Sprintf("action = $%d", len(args)))
	}
	if q.ResourceType != "" {
		args = append(args, q.ResourceType)
		where = append(where, fmt.Sprintf("resource_type = $%d", len(args)))
	}
	if q.OnlyFailures {
		where = append(where, "ok = FALSE")
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	if err := s.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM audit_log WHERE "+whereClause,
		args...,
	).Scan(&total); err != nil {
		return nil, fmt.Errorf("audit: count: %w", err)
	}

	args = append(args, q.Limit, q.Offset)
	rowSQL := fmt.Sprintf(`
		SELECT id, occurred_at, actor, action, resource_type, resource_id,
		       request_id, before, after, ok, error_message
		FROM audit_log
		WHERE %s
		ORDER BY occurred_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, len(args)-1, len(args))

	rows, err := s.pool.Query(ctx, rowSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("audit: list: %w", err)
	}
	defer rows.Close()

	entries := []LogEntry{}
	for rows.Next() {
		var e LogEntry
		var requestID, errMsg *string
		var before, after []byte
		if err := rows.Scan(
			&e.ID, &e.OccurredAt, &e.Actor, &e.Action,
			&e.ResourceType, &e.ResourceID, &requestID,
			&before, &after, &e.OK, &errMsg,
		); err != nil {
			return nil, fmt.Errorf("audit: scan: %w", err)
		}
		if requestID != nil {
			e.RequestID = *requestID
		}
		if errMsg != nil {
			e.ErrorMessage = *errMsg
		}
		if len(before) > 0 {
			e.Before = json.RawMessage(before)
		}
		if len(after) > 0 {
			e.After = json.RawMessage(after)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("audit: iterate: %w", err)
	}
	return &ListResponse{Entries: entries, Total: total, Limit: q.Limit, Offset: q.Offset}, nil
}
