package hub

import (
	"sync"

	"github.com/google/uuid"
)

type Manager struct {
	mu   sync.RWMutex
	hubs map[string]*Hub
}

func NewManager() *Manager {
	return &Manager{hubs: make(map[string]*Hub)}
}

func (m *Manager) New() (hubID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	hubID = uuid.New().String()

	r := NewHub(hubID)
	m.hubs[hubID] = r

	return hubID
}

func (m *Manager) Get(hubID string) (*Hub, bool) {
	m.mu.RLock()
	r, ok := m.hubs[hubID]
	m.mu.RUnlock()
	return r, ok
}

func (m *Manager) Delete(hubID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r, ok := m.hubs[hubID]; ok {
		r.Close()
		delete(m.hubs, hubID)
	}
}
