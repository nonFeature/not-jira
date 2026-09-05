package bot

import (
	"sync"
	"time"
)

type UserRateLimiter struct {
	mu      sync.Mutex
	actions map[int64][]time.Time
	limit   int
	window  time.Duration
}

func NewUserRateLimiter(limit int, window time.Duration) *UserRateLimiter {
	return &UserRateLimiter{
		actions: make(map[int64][]time.Time),
		limit:   limit,
		window:  window,
	}
}

func (l *UserRateLimiter) Allow(userID int64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	timestamps := l.actions[userID]
	var valid []time.Time
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= l.limit {
		l.actions[userID] = valid
		return false
	}

	valid = append(valid, now)
	l.actions[userID] = valid
	return true
}

func (l *UserRateLimiter) Cleanup(maxAge time.Duration) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-maxAge)
	cleaned := 0

	for uid, times := range l.actions {
		if len(times) == 0 || times[len(times)-1].Before(cutoff) {
			delete(l.actions, uid)
			cleaned++
		}
	}
	return cleaned
}
