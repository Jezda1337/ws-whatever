package ws

import (
	"sync"

	"gorm.io/gorm"
)

type Manager struct {
	Clients ClientList

	sync.RWMutex
	db *gorm.DB
}

func NewManager(db *gorm.DB) *Manager {
	return &Manager{Clients: make(ClientList), db: db}
}

func (m *Manager) AddClient(c *Client) {
	m.Lock()
	defer m.Unlock()

	m.Clients[c] = true
}

func (m *Manager) RemoveClient(c *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.Clients[c]; ok {
		close(c.Send)
		c.Conn.Close()
		delete(m.Clients, c)
	}
}
