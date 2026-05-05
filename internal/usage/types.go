// Package usage owns Workspace per-user usage stats backed by the Admin
// Reports API (Drive quota, Gmail activity, last-login). The directory API
// alone never returns these — the dedicated usage scope plus the Reports
// service are required.
package usage

import "time"

// UserUsage is the wire shape returned for one user.
type UserUsage struct {
	UserEmail            string    `json:"user_email"`
	LastLoginTime        time.Time `json:"last_login_time,omitempty"`
	GmailLastInteraction time.Time `json:"gmail_last_interaction,omitempty"`
	GmailNumReceived     int64     `json:"gmail_num_received"`
	GmailNumSent         int64     `json:"gmail_num_sent"`
	DriveUsedMB          int64     `json:"drive_used_mb"`
	DriveTotalMB         int64     `json:"drive_total_mb"`
	DriveItemsOwned      int64     `json:"drive_items_owned"`
	HasDrivePresence     bool      `json:"has_drive_presence"`
	HasGmailPresence     bool      `json:"has_gmail_presence"`
}

// Snapshot is what GET /api/v1/usage/users returns: every user's usage
// for a single Reports date plus a tiny header describing freshness.
type Snapshot struct {
	Date         string               `json:"date"`
	GeneratedAt  time.Time            `json:"generated_at"`
	DurationMS   int64                `json:"duration_ms"`
	TotalUsers   int                  `json:"total_users"`
	Users        map[string]UserUsage `json:"users"`
}
