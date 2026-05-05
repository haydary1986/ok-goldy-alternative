// Package stats produces aggregate metrics over the Workspace tenant.
//
// The cost of computing these is dominated by walking every user/group
// through Admin SDK pagination, so the service caches its result briefly
// and reuses users.Service.ListAll which has its own snapshot cache.
package stats

import "time"

// Overview is what GET /api/v1/stats/overview returns.
type Overview struct {
	GeneratedAt time.Time `json:"generated_at"`
	DurationMS  int64     `json:"duration_ms"`

	TotalUsers     int `json:"total_users"`
	ActiveUsers    int `json:"active_users"`
	SuspendedUsers int `json:"suspended_users"`
	AdminUsers     int `json:"admin_users"`
	NeverLoggedIn  int `json:"never_logged_in"`

	InactiveSince30  int `json:"inactive_30d"`
	InactiveSince90  int `json:"inactive_90d"`
	InactiveSince180 int `json:"inactive_180d"`
	InactiveSince365 int `json:"inactive_365d"`

	CreatedLast7d  int `json:"created_last_7d"`
	CreatedLast30d int `json:"created_last_30d"`
	CreatedLast90d int `json:"created_last_90d"`

	UsersByOU []OUUserCount `json:"users_by_ou"`

	TotalGroups       int `json:"total_groups"`
	EmptyGroups       int `json:"empty_groups"`
	TotalGroupMembers int `json:"total_group_members"`
}

// OUUserCount is one row of the per-OU breakdown.
type OUUserCount struct {
	OrgUnitPath string `json:"org_unit_path"`
	Total       int    `json:"total"`
	Active      int    `json:"active"`
	Suspended   int    `json:"suspended"`
}
