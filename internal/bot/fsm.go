package bot

import (
	"sync"

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
	f.sessions[userID] = sess
}

func (f *FSM) Clear(userID int64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.sessions, userID)
}
