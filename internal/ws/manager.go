package ws

import (
	"log"
	"sync"
)

type Manager struct {
	mu sync.RWMutex

	clients map[uint]map[*Client]struct{}

	rooms map[uint]map[*Client]struct{}
}

func NewManager() *Manager {
	return &Manager{
		clients: make(map[uint]map[*Client]struct{}),
		rooms:   make(map[uint]map[*Client]struct{}),
	}
}

func (m *Manager) Register(client *Client) {

	m.mu.Lock()

	userClients := m.clients[client.UserID]

	if userClients == nil {
		userClients = make(map[*Client]struct{})
		m.clients[client.UserID] = userClients
	}
	userClients[client] = struct{}{}
	userCount := len(m.clients)
	connectionCount := m.connectionCountLocked()
	m.mu.Unlock()

	log.Printf(
		"user %d connection registered, users=%d connections=%d",
		client.UserID,
		userCount,
		connectionCount,
	)

}

func (m *Manager) Unregister(client *Client) {
	m.mu.Lock()

	removed := false

	if userClients := m.clients[client.UserID]; userClients != nil {
		if _, exists := userClients[client]; exists {
			delete(userClients, client)
			removed = true
		}

		if len(userClients) == 0 {
			delete(m.clients, client.UserID)
		}
	}
	m.removeFromRoomsLocked(client)

	userCount := len(m.clients)

	connectionCount := m.connectionCountLocked()

	m.mu.Unlock()

	if removed {
		log.Printf(
			"user %d connection unregistered, users=%d connections=%d",
			client.UserID,
			userCount,
			connectionCount,
		)
	}

}
func (m *Manager) connectionCountLocked() int {
	count := 0

	for _, userClients := range m.clients {

		count += len(userClients)
	}

	return count
}

func (m *Manager) GetClients(userID uint) []*Client {
	m.mu.RLock()
	userClients := m.clients[userID]
	clients := make([]*Client, 0, len(userClients))

	for client := range userClients {
		clients = append(clients, client)
	}

	m.mu.RUnlock()

	return clients
}

func (m *Manager) SendToUser(
	userID uint,
	message []byte,
) bool {
	clients := m.GetClients(userID)
	if len(clients) == 0 {
		return false
	}
	sent := false
	for _, client := range clients {
		if client.SendMessage(message) {
			sent = true
		}
	}

	return sent
}

func (m *Manager) onlineCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.clients)
}

func (m *Manager) JoinConversation(
	ConversationID uint,
	client *Client,
) {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.rooms[ConversationID]
	if !ok {
		room = make(map[*Client]struct{})
		m.rooms[ConversationID] = room
	}
	room[client] = struct{}{}

}

func (m *Manager) IsInConversation(
	ConversationID uint,
	client *Client,
) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	room, ok := m.rooms[ConversationID]
	if !ok {
		return false
	}
	_, exists := room[client]
	return exists
}

func (m *Manager) LeaveConversation(
	ConversationID uint,
	client *Client,
) {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.rooms[ConversationID]
	if !ok {
		return
	}

	delete(room, client)

	if len(room) == 0 {
		delete(m.rooms, ConversationID)
	}
}

func (m *Manager) BroadcastToConversation(
	ConversationID uint,
	message []byte,
) int {
	m.mu.RLock()

	room, ok := m.rooms[ConversationID]
	if !ok {
		m.mu.RUnlock()
		return 0
	}

	clients := make([]*Client, 0, len(room))

	for client := range room {
		clients = append(clients, client)
	}

	m.mu.RUnlock()
	successCount := 0
	for _, client := range clients {
		if client.SendMessage(message) {
			successCount++
		}
	}
	return successCount

}

func (m *Manager) removeFromRoomsLocked(
	Client *Client,
) {
	for ConversationID, room := range m.rooms {
		delete(room, Client)

		if len(room) == 0 {
			delete(m.rooms, ConversationID)

		}
	}

}

func (m *Manager) RoomUserIDs(
	conversationID uint,
) []uint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	room := m.rooms[conversationID]
	if room == nil {
		return nil
	}

	userSet := make(map[uint]struct{})

	for client := range room {
		userSet[client.UserID] = struct{}{}
	}
	userIDs := make([]uint, 0, len(userSet))
	for userID := range userSet {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}
