package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

type SessionData struct {
	ID            string
	Authenticated bool
	Username      string
	CSRFToken     string
	CreatedAt     time.Time
}

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*SessionData
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*SessionData),
	}
}

func (s *SessionStore) GetOrCreate(r *http.Request, w http.ResponseWriter) *SessionData {
	s.mu.Lock()
	defer s.mu.Unlock()

	cookie, err := r.Cookie("admin_session")
	if err == nil && cookie.Value != "" {
		if sess, exists := s.sessions[cookie.Value]; exists {
			return sess
		}
	}

	bytes := make([]byte, 32)
	_, _ = rand.Read(bytes)
	sessionID := hex.EncodeToString(bytes)

	csrfBytes := make([]byte, 16)
	_, _ = rand.Read(csrfBytes)
	csrfToken := hex.EncodeToString(csrfBytes)

	sess := &SessionData{
		ID:            sessionID,
		Authenticated: false,
		CSRFToken:     csrfToken,
		CreatedAt:     time.Now(),
	}
	s.sessions[sessionID] = sess

	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 30,
	})

	return sess
}

func (s *SessionStore) Get(sessionID string) (*SessionData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, exists := s.sessions[sessionID]
	return sess, exists
}

func (s *SessionStore) Delete(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
}

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}
