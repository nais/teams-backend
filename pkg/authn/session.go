package authn

import (
	"time"
)

type SessionStore interface {
	Create(Session)
	Destroy(string)
}

type Session struct {
	Key     string
	Expires time.Time
}

type sessionStore struct {
	sessions map[string]Session
}

func NewStore() SessionStore {
	return &sessionStore{
		sessions: make(map[string]Session),
	}
}

func (s *sessionStore) Create(sess Session) {
	s.sessions[sess.Key] = sess
}

func (s *sessionStore) Destroy(key string) {
	delete(s.sessions, key)
}
