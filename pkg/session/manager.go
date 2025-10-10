package session

import (
	"strconv"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
)

type Session struct {
	ID        string
	Messages  []*schema.Message
	UpdatedAt time.Time
}

type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewManager() *Manager {
	return &Manager{sessions: make(map[string]*Session)}
}

func (m *Manager) Create(messages []*schema.Message) *Session {
	id := strconv.FormatInt(time.Now().UnixNano(), 10)
	copied := make([]*schema.Message, len(messages))
	copy(copied, messages)
	session := &Session{ID: id, Messages: copied, UpdatedAt: time.Now()}
	m.mu.Lock()
	m.sessions[id] = session
	m.mu.Unlock()
	return session
}

func (m *Manager) Get(id string) (*Session, bool) {
	if id == "" {
		return nil, false
	}
	m.mu.RLock()
	session := m.sessions[id]
	m.mu.RUnlock()
	if session == nil {
		return nil, false
	}
	return session, true
}

func (m *Manager) Append(id string, messages ...*schema.Message) (*Session, bool) {
	if id == "" {
		return nil, false
	}
	m.mu.Lock()
	session := m.sessions[id]
	if session == nil {
		m.mu.Unlock()
		return nil, false
	}
	session.Messages = append(session.Messages, messages...)
	session.UpdatedAt = time.Now()
	m.mu.Unlock()
	return session, true
}

func (m *Manager) Delete(id string) {
	if id == "" {
		return
	}
	m.mu.Lock()
	delete(m.sessions, id)
	m.mu.Unlock()
}
