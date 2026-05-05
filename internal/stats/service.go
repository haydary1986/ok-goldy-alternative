package stats

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/haydary1986/ok-goldy-alternative/internal/groups"
	"github.com/haydary1986/ok-goldy-alternative/internal/users"
)

// CacheTTL is how long an Overview is reused before recomputing. The
// underlying user list itself is also cached by users.Service, so a
// dashboard refresh within the TTL is essentially free.
const CacheTTL = 5 * time.Minute

// Service produces an Overview by combining the cached user list with a
// shallow walk of every group page.
type Service struct {
	users  *users.Service
	groups *groups.Service

	mu       sync.RWMutex
	cache    *Overview
	cachedAt time.Time
}

func NewService(usersSvc *users.Service, groupsSvc *groups.Service) *Service {
	return &Service{users: usersSvc, groups: groupsSvc}
}

// Overview returns aggregated tenant metrics, recomputing only when the
// cache is stale (or force=true).
func (s *Service) Overview(ctx context.Context, force bool) (*Overview, error) {
	if !force {
		s.mu.RLock()
		if s.cache != nil && time.Since(s.cachedAt) < CacheTTL {
			out := *s.cache
			s.mu.RUnlock()
			return &out, nil
		}
		s.mu.RUnlock()
	}

	start := time.Now()

	allUsers, err := s.users.ListAll(ctx, force)
	if err != nil {
		return nil, err
	}

	out := &Overview{
		GeneratedAt: time.Now(),
		TotalUsers:  len(allUsers),
	}

	now := time.Now()
	cutoff30 := now.AddDate(0, 0, -30)
	cutoff90 := now.AddDate(0, 0, -90)
	cutoff180 := now.AddDate(0, 0, -180)
	cutoff365 := now.AddDate(0, 0, -365)
	created7 := now.AddDate(0, 0, -7)
	created30 := now.AddDate(0, 0, -30)
	created90 := now.AddDate(0, 0, -90)

	ouMap := map[string]*OUUserCount{}

	for _, u := range allUsers {
		if u.Suspended {
			out.SuspendedUsers++
		} else {
			out.ActiveUsers++
		}
		if u.IsAdmin {
			out.AdminUsers++
		}
		if u.LastLoginTime.IsZero() {
			out.NeverLoggedIn++
			out.InactiveSince30++
			out.InactiveSince90++
			out.InactiveSince180++
			out.InactiveSince365++
		} else {
			if u.LastLoginTime.Before(cutoff30) {
				out.InactiveSince30++
			}
			if u.LastLoginTime.Before(cutoff90) {
				out.InactiveSince90++
			}
			if u.LastLoginTime.Before(cutoff180) {
				out.InactiveSince180++
			}
			if u.LastLoginTime.Before(cutoff365) {
				out.InactiveSince365++
			}
		}
		if !u.CreationTime.IsZero() {
			if u.CreationTime.After(created7) {
				out.CreatedLast7d++
			}
			if u.CreationTime.After(created30) {
				out.CreatedLast30d++
			}
			if u.CreationTime.After(created90) {
				out.CreatedLast90d++
			}
		}

		ouPath := u.OrgUnitPath
		if ouPath == "" {
			ouPath = "/"
		}
		bucket := ouMap[ouPath]
		if bucket == nil {
			bucket = &OUUserCount{OrgUnitPath: ouPath}
			ouMap[ouPath] = bucket
		}
		bucket.Total++
		if u.Suspended {
			bucket.Suspended++
		} else {
			bucket.Active++
		}
	}

	ouList := make([]OUUserCount, 0, len(ouMap))
	for _, b := range ouMap {
		ouList = append(ouList, *b)
	}
	sort.Slice(ouList, func(i, j int) bool { return ouList[i].Total > ouList[j].Total })
	out.UsersByOU = ouList

	// Walk groups (shallow — just counts).
	if s.groups != nil {
		token := ""
		for {
			page, err := s.groups.List(ctx, token, 200)
			if err != nil {
				break
			}
			for _, g := range page.Groups {
				out.TotalGroups++
				if g.DirectMembersCount == 0 {
					out.EmptyGroups++
				}
				out.TotalGroupMembers += int(g.DirectMembersCount)
			}
			if page.NextPageToken == "" {
				break
			}
			token = page.NextPageToken
		}
	}

	out.DurationMS = time.Since(start).Milliseconds()

	s.mu.Lock()
	s.cache = out
	s.cachedAt = time.Now()
	s.mu.Unlock()

	cp := *out
	return &cp, nil
}
