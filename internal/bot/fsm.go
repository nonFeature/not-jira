package bot

import (
	"context"
	"log"
	"sync"
	"time"

	"not-jira/internal/models"
	"not-jira/internal/storage"
)

type FSM struct {
	mu       sync.RWMutex
	sessions map[int64]*models.UserSession
	storage  storage.Storage
}

func NewFSM(st storage.Storage) *FSM {
	return &FSM{
		sessions: make(map[int64]*models.UserSession),
		storage:  st,
	}
}

func cloneSession(sess *models.UserSession) *models.UserSession {
	if sess == nil {
		return nil
	}
	cp := *sess
	if sess.DraftTask != nil {
		draft := *sess.DraftTask
		cp.DraftTask = &draft
	}
	return &cp
}

func (f *FSM) Get(ctx context.Context, userID int64) *models.UserSession {
	f.mu.RLock()
	sess, ok := f.sessions[userID]
	f.mu.RUnlock()
	if ok {
		return cloneSession(sess)
	}

	if f.storage == nil {
		return nil
	}

	loaded, err := f.storage.GetSession(ctx, userID)
	if err != nil {
		log.Printf("[FSM WARN] Failed to load session for %d: %v", userID, err)
		return nil
	}
	if loaded == nil {
		return nil
	}

	f.mu.Lock()
	f.sessions[userID] = cloneSession(loaded)
	f.mu.Unlock()

	return cloneSession(loaded)
}

func (f *FSM) Set(ctx context.Context, userID int64, sess *models.UserSession) {
	if sess == nil {
		f.Clear(ctx, userID)
		return
	}

	if sess.UpdatedAt.IsZero() {
		sess.UpdatedAt = time.Now().UTC()
	}

	f.mu.Lock()
	f.sessions[userID] = cloneSession(sess)
	f.mu.Unlock()

	if f.storage != nil {
		if err := f.storage.SaveSession(ctx, userID, sess); err != nil {
			log.Printf("[FSM WARN] Failed to persist session for %d: %v", userID, err)
		}
	}
}

func (f *FSM) Clear(ctx context.Context, userID int64) {
	f.mu.Lock()
	delete(f.sessions, userID)
	f.mu.Unlock()

	if f.storage != nil {
		if err := f.storage.DeleteSession(ctx, userID); err != nil {
			log.Printf("[FSM WARN] Failed to delete session for %d: %v", userID, err)
		}
	}
}

func (f *FSM) Cleanup(ctx context.Context, maxAge time.Duration) int {
	f.mu.Lock()
	now := time.Now().UTC()
	cleaned := 0
	for uid, sess := range f.sessions {
		if sess == nil || sess.State == models.StateNone || now.Sub(sess.UpdatedAt) > maxAge {
			delete(f.sessions, uid)
			cleaned++
		}
	}
	f.mu.Unlock()

	if f.storage != nil {
		if n, err := f.storage.CleanupSessions(ctx, maxAge); err != nil {
			log.Printf("[FSM WARN] Failed to cleanup persisted sessions: %v", err)
		} else {
			cleaned += int(n)
		}
	}
	return cleaned
}
