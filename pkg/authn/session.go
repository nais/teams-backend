package authn

import (
	"time"
)

type SessionStore interface {
	Create(*Session)
	Destroy(string)
	Get(key string) *Session
}

type Session struct {
	Key     string
	Expires time.Time
	Email   string
}

type sessionStore struct {
	sessions map[string]*Session
}

func NewStore() SessionStore {
	return &sessionStore{
		sessions: make(map[string]*Session),
	}
}

func (s *sessionStore) Get(key string) *Session {
	return s.sessions[key]
}

func (s *sessionStore) Create(sess *Session) {
	s.sessions[sess.Key] = sess
}

func (s *sessionStore) Destroy(key string) {
	delete(s.sessions, key)
}
