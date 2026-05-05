package workspace

import (
	"context"
	"fmt"
	"time"

	reports "google.golang.org/api/admin/reports/v1"
)

// UsageInfo is a flat, parameter-name-free projection of one user's
// daily usage report from the Admin Reports API.
type UsageInfo struct {
	UserEmail            string
	UserID               string
	Date                 string
	LastLoginTime        time.Time
	GmailLastInteraction time.Time
	GmailNumReceived     int64
	GmailNumSent         int64
	DriveUsedMB          int64
	DriveTotalMB         int64
	DriveItemsOwned      int64
	HasDrivePresence     bool
	HasGmailPresence     bool
}

// UsageReportParameters Goldy fetches in one call. Each maps to a Reports
// API parameter name; missing values just stay zero in UsageInfo.
var UsageReportParameters = []string{
	"accounts:last_login_time",
	"gmail:last_interaction_time",
	"gmail:num_emails_received",
	"gmail:num_emails_sent",
	"drive:used_quota_in_mb",
	"drive:total_quota_in_mb",
	"drive:num_items_owned",
	"drive:has_drive_presence",
	"gmail:has_gmail_presence",
}

// ListUserUsagePage fetches one paginated page of per-user usage reports
// for the given date (yyyy-MM-dd). Reports API maxResults caps at 1000.
func (c *Client) ListUserUsagePage(ctx context.Context, date, parameters, pageToken string, pageSize int64) ([]*reports.UsageReport, string, error) {
	if c.reports == nil {
		return nil, "", fmt.Errorf("workspace: reports service not configured")
	}
	if err := c.Wait(ctx); err != nil {
		return nil, "", err
	}
	if pageSize <= 0 || pageSize > 1000 {
		pageSize = 1000
	}
	call := c.reports.UserUsageReport.Get("all", date).
		CustomerId(c.customerID).
		Parameters(parameters).
		MaxResults(pageSize).
		Context(ctx)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	resp, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("workspace: list user usage on %s: %w", date, err)
	}
	return resp.UsageReports, resp.NextPageToken, nil
}

// ParseUsage flattens a Reports UsageReport into UsageInfo, picking only
// the parameters Goldy actually uses.
func ParseUsage(r *reports.UsageReport) UsageInfo {
	info := UsageInfo{}
	if r == nil {
		return info
	}
	if r.Date != "" {
		info.Date = r.Date
	}
	if r.Entity != nil {
		info.UserEmail = r.Entity.UserEmail
		info.UserID = r.Entity.ProfileId
	}
	for _, p := range r.Parameters {
		switch p.Name {
		case "accounts:last_login_time":
			if t, err := time.Parse(time.RFC3339, p.DatetimeValue); err == nil && t.Year() > 1970 {
				info.LastLoginTime = t
			}
		case "gmail:last_interaction_time":
			if t, err := time.Parse(time.RFC3339, p.DatetimeValue); err == nil && t.Year() > 1970 {
				info.GmailLastInteraction = t
			}
		case "gmail:num_emails_received":
			info.GmailNumReceived = p.IntValue
		case "gmail:num_emails_sent":
			info.GmailNumSent = p.IntValue
		case "drive:used_quota_in_mb":
			info.DriveUsedMB = p.IntValue
		case "drive:total_quota_in_mb":
			info.DriveTotalMB = p.IntValue
		case "drive:num_items_owned":
			info.DriveItemsOwned = p.IntValue
		case "drive:has_drive_presence":
			info.HasDrivePresence = p.BoolValue
		case "gmail:has_gmail_presence":
			info.HasGmailPresence = p.BoolValue
		}
	}
	return info
}
