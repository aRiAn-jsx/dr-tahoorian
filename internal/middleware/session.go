package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

type Session struct {
	ID        string
	Data      map[string]interface{}
	ExpiresAt time.Time
}

type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	secret   string
}

func NewSessionManager(secret string) *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*Session),
		secret:   secret,
	}
	go sm.cleanup()
	return sm
}

func (sm *SessionManager) Create() *Session {
	b := make([]byte, 32)
	rand.Read(b)
	id := base64.StdEncoding.EncodeToString(b)
	session := &Session{
		ID:        id,
		Data:      make(map[string]interface{}),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	sm.mu.Lock()
	sm.sessions[id] = session
	sm.mu.Unlock()
	return session
}

func (sm *SessionManager) Get(id string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	s, ok := sm.sessions[id]
	if !ok || time.Now().After(s.ExpiresAt) {
		return nil, false
	}
	return s, true
}

func (sm *SessionManager) Delete(id string) {
	sm.mu.Lock()
	delete(sm.sessions, id)
	sm.mu.Unlock()
}

func (sm *SessionManager) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		sm.mu.Lock()
		for id, s := range sm.sessions {
			if time.Now().After(s.ExpiresAt) {
				delete(sm.sessions, id)
			}
		}
		sm.mu.Unlock()
	}
}

func SessionMiddleware(sm *SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_id")
			if err != nil || cookie.Value == "" {
				session := sm.Create()
				http.SetCookie(w, &http.Cookie{
					Name:     "session_id",
					Value:    session.ID,
					Path:     "/",
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
					MaxAge:   86400,
				})
				next.ServeHTTP(w, r.WithContext(r.Context()))
				return
			}
			if _, ok := sm.Get(cookie.Value); !ok {
				session := sm.Create()
				http.SetCookie(w, &http.Cookie{
					Name:     "session_id",
					Value:    session.ID,
					Path:     "/",
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
					MaxAge:   86400,
				})
			}
			next.ServeHTTP(w, r.WithContext(r.Context()))
		})
	}
}
