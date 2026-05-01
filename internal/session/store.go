package session

import (
	"sync"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Session struct {
	ID       string    `json:"id"`
	Messages []Message `json:"messages"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}

type Store struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	ttl      time.Duration
	stopCh   chan struct{}
}

func NewStore(ttl time.Duration) *Store {
	s := &Store{sessions: make(map[string]*Session), ttl: ttl, stopCh: make(chan struct{})}
	go s.cleanupLoop()
	return s
}

func (s *Store) Get(id string) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[id]
}

func (s *Store) Save(sess *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess.Updated = time.Now()
	if sess.Created.IsZero() { sess.Created = sess.Updated }
	s.sessions[sess.ID] = sess
}

func (s *Store) cleanupLoop() {
	t := time.NewTicker(s.ttl / 2)
	defer t.Stop()
	for {
		select {
		case <-t.C: s.cleanup()
		case <-s.stopCh: return
		}
	}
}

func (s *Store) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	cutoff := time.Now().Add(-s.ttl)
	for id, sess := range s.sessions {
		if sess.Updated.Before(cutoff) { delete(s.sessions, id) }
	}
}

func (s *Store) Close() { close(s.stopCh) }
