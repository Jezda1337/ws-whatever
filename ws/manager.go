package ws

import (
	"log/slog"
	"sync"
	"time"

	"gorm.io/gorm"
)

type Manager struct {
	sync.RWMutex
	db     *gorm.DB
	logger *slog.Logger

	clients     map[*Client]bool
	rooms       map[int]map[*Client]bool
	clientRooms map[*Client]int
	typing      map[int]map[int]time.Time
}

func NewManager(db *gorm.DB, logger *slog.Logger) *Manager {
	return &Manager{
		db:          db,
		logger:      logger,
		clients:     make(map[*Client]bool),
		rooms:       make(map[int]map[*Client]bool),
		clientRooms: make(map[*Client]int),
		typing:      make(map[int]map[int]time.Time),
	}
}

func (m *Manager) AddClient(c *Client) {
	m.Lock()
	defer m.Unlock()

	m.clients[c] = true
}

func (m *Manager) RemoveClient(c *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[c]; ok {
		if roomID, inRoom := m.clientRooms[c]; inRoom {
			delete(m.rooms[roomID], c)
			if len(m.rooms[roomID]) == 0 {
				delete(m.rooms, roomID)
			}
			delete(m.clientRooms, c)
		}

		close(c.Send)
		c.Conn.Close()
		delete(m.clients, c)
	}
}

func (m *Manager) JoinRoom(c *Client, roomID int) error {
	var room Room
	if err := m.db.First(&room, roomID).Error; err != nil {
		return err
	}

	var user User
	if err := m.db.First(&user, c.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			user = User{ID: c.UserID}
			if err := m.db.Create(&user).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	var count int64
	err := m.db.Model(&RoomParticipant{}).
		Where("room_id = ? AND user_id = ?", roomID, c.UserID).
		Count(&count).Error
	if err != nil {
		return err
	}

	if count == 0 {
		participant := RoomParticipant{
			RoomID: roomID,
			UserID: c.UserID,
			Role:   "member",
		}
		if err := m.db.Create(&participant).Error; err != nil {
			return err
		}
	}

	m.Lock()
	defer m.Unlock()

	if oldRoomID, inRoom := m.clientRooms[c]; inRoom {
		delete(m.rooms[oldRoomID], c)
		if len(m.rooms[oldRoomID]) == 0 {
			delete(m.rooms, oldRoomID)
		}
	}

	if m.rooms[roomID] == nil {
		m.rooms[roomID] = make(map[*Client]bool)
	}
	m.rooms[roomID][c] = true
	m.clientRooms[c] = roomID
	c.RoomID = roomID

	return nil
}

func (m *Manager) BroadcastToRoom(roomID int, data []byte) {
	m.RLock()
	recipients := make([]*Client, 0, len(m.rooms[roomID]))
	for client := range m.rooms[roomID] {
		recipients = append(recipients, client)
	}
	m.RUnlock()

	for _, client := range recipients {
		select {
		case client.Send <- data:
		default:
			m.logger.Warn("client send buffer full, skipping", "clientID", client.ID, "userID", client.UserID)
		}
	}
}

func (m *Manager) SetTyping(roomID, userID int) {
	m.Lock()
	defer m.Unlock()

	if m.typing[roomID] == nil {
		m.typing[roomID] = make(map[int]time.Time)
	}
	m.typing[roomID][userID] = time.Now()
}

func (m *Manager) GetTypingUsers(roomID int) []int {
	m.RLock()
	defer m.RUnlock()

	var users []int
	cutoff := time.Now().Add(-3 * time.Second)
	for userID, lastTyped := range m.typing[roomID] {
		if lastTyped.After(cutoff) {
			users = append(users, userID)
		}
	}
	return users
}
