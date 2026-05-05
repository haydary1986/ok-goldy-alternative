package usage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/haydary1986/ok-goldy-alternative/internal/workspace"
)

// CacheTTL is how long a Snapshot is reused. Reports API data is daily,
// so 30 minutes is plenty.
const CacheTTL = 30 * time.Minute

// ErrUnavailable signals "no Workspace credentials" at this layer.
var ErrUnavailable = errors.New("usage: workspace client is not configured")

// Service walks the Reports API to assemble a per-user usage snapshot
// then caches it.
type Service struct {
	wsProv *workspace.Provider

	mu       sync.RWMutex
	cache    *Snapshot
	cachedAt time.Time
}

func NewService(wsProv *workspace.Provider) *Service { return &Service{wsProv: wsProv} }

// Snapshot returns the latest cached snapshot, or fetches a fresh one if
// the cache is empty or older than CacheTTL.
func (s *Service) Snapshot(ctx context.Context, force bool) (*Snapshot, error) {
	if !force {
		s.mu.RLock()
		if s.cache != nil && time.Since(s.cachedAt) < CacheTTL {
			out := *s.cache
			s.mu.RUnlock()
			return &out, nil
		}
		s.mu.RUnlock()
	}

	if s.wsProv == nil || s.wsProv.Get() == nil {
		return nil, ErrUnavailable
	}
	c, err := s.wsProv.Variant(ctx, workspace.UsageScopes)
	if err != nil {
		return nil, fmt.Errorf("usage: build client: %w", err)
	}

	// Reports data lags 24-48h. We try yesterday then 2 days ago then
	// 3 days; whichever has data wins.
	parameters := strings.Join(workspace.UsageReportParameters, ",")
	start := time.Now()
	var pickedDate string
	users := map[string]UserUsage{}

	for _, daysBack := range []int{1, 2, 3} {
		date := time.Now().UTC().AddDate(0, 0, -daysBack).Format("2006-01-02")
		token := ""
		gotAny := false
		for {
			reports, next, err := c.ListUserUsagePage(ctx, date, parameters, token, 1000)
			if err != nil {
				return nil, err
			}
			if len(reports) > 0 {
				gotAny = true
			}
			for _, r := range reports {
				info := workspace.ParseUsage(r)
				if info.UserEmail == "" {
					continue
				}
				users[info.UserEmail] = UserUsage{
					UserEmail:            info.UserEmail,
					LastLoginTime:        info.LastLoginTime,
					GmailLastInteraction: info.GmailLastInteraction,
					GmailNumReceived:     info.GmailNumReceived,
					GmailNumSent:         info.GmailNumSent,
					DriveUsedMB:          info.DriveUsedMB,
					DriveTotalMB:         info.DriveTotalMB,
					DriveItemsOwned:      info.DriveItemsOwned,
					HasDrivePresence:     info.HasDrivePresence,
					HasGmailPresence:     info.HasGmailPresence,
				}
			}
			if next == "" {
				break
			}
			token = next
		}
		if gotAny {
			pickedDate = date
			break
		}
	}
	if pickedDate == "" {
		return nil, fmt.Errorf("usage: no usage report available in the last 3 days")
	}

	snap := &Snapshot{
		Date:        pickedDate,
		GeneratedAt: time.Now(),
		DurationMS:  time.Since(start).Milliseconds(),
		TotalUsers:  len(users),
		Users:       users,
	}

	s.mu.Lock()
	s.cache = snap
	s.cachedAt = time.Now()
	s.mu.Unlock()

	out := *snap
	return &out, nil
}
