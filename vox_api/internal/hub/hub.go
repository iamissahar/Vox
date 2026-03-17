package hub

import "sync"

type Consumer struct {
	ID   string
	Send chan []byte
}

type Hub struct {
	ID        string
	mu        sync.RWMutex
	consumers map[string]*Consumer
	broadcast chan []byte
	quit      chan struct{}
}

func NewHub(id string) *Hub {
	h := &Hub{
		ID:        id,
		consumers: make(map[string]*Consumer),
		broadcast: make(chan []byte, 256),
		quit:      make(chan struct{}),
	}
	go h.run()
	return h
}

func (h *Hub) run() {
	for {
		select {
		case chunk := <-h.broadcast:
			h.mu.RLock()
			for _, c := range h.consumers {
				select {
				case c.Send <- chunk:
				default:
				}
			}
			h.mu.RUnlock()
		case <-h.quit:
			for {
				select {
				case chunk := <-h.broadcast:
					h.mu.RLock()
					for _, c := range h.consumers {
						select {
						case c.Send <- chunk:
						default:
						}
					}
					h.mu.RUnlock()
				default:
					return
				}
			}
		}
	}
}

func (r *Hub) Publish(chunk []byte) {
	select {
	case r.broadcast <- chunk:
	default:
		// broadcast buffer full, drop chunk
	}
}

func (r *Hub) AddConsumer(c *Consumer) {
	r.mu.Lock()
	r.consumers[c.ID] = c
	r.mu.Unlock()
}

func (r *Hub) RemoveConsumer(id string) {
	r.mu.Lock()
	if c, ok := r.consumers[id]; ok {
		close(c.Send)
		delete(r.consumers, id)
	}
	r.mu.Unlock()
}

func (r *Hub) Close() {
	close(r.quit)
}
