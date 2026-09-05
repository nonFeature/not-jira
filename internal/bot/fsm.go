package bot

import (
	"sync"
	"time"

	"not-jira/internal/models"
)

type FSM struct {
	mu       sync.RWMutex
	sessions map[int64]*models.UserSession
}

func NewFSM() *FSM {
	return &FSM{
		sessions: make(map[int64]*models.UserSession),
	}
}

func (f *FSM) Get(userID int64) *models.UserSession {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.sessions[userID]
}

func (f *FSM) Set(userID int64, sess *models.UserSession) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if sess != nil && sess.UpdatedAt.IsZero() {
		sess.UpdatedAt = time.Now().UTC()
	}
	f.sessions[userID] = sess
}

func (f *FSM) Clear(userID int64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.sessions, userID)
}

func (f *FSM) Cleanup(maxAge time.Duration) int {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now().UTC()
	cleaned := 0
	for uid, sess := range f.sessions {
		if sess == nil || sess.State == models.StateNone || now.Sub(sess.UpdatedAt) > maxAge {
			delete(f.sessions, uid)
			cleaned++
		}
	}
	return cleaned
}
