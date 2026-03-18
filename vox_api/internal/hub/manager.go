package hub

import (
	"sync"

	"github.com/google/uuid"
)

type HostAndHubs struct {
	mu      sync.RWMutex
	storage map[string][]string
}

func NewHostAndHubs() *HostAndHubs {
	return &HostAndHubs{storage: make(map[string][]string)}
}

func (h *HostAndHubs) AddHub(userID, hubID string) {
	h.mu.Lock()
	h.storage[userID] = append(h.storage[userID], hubID)
	h.mu.Unlock()
}

func (h *HostAndHubs) RemoveHub(userID, hubID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, id := range h.storage[userID] {
		if id == hubID {
			h.storage[userID] = append(h.storage[userID][:i], h.storage[userID][i+1:]...)
			return
		}
	}
}

func (h *HostAndHubs) GetHubs(userID string) []string {
	return h.storage[userID]
}

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
