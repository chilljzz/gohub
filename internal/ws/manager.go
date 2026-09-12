package ws

import (
	"log"
	"sync"
)

type Manager struct {
	mu sync.RWMutex

	clients map[uint]*Client

	rooms map[uint]map[*Client]struct{}

	channelConversations map[uint]uint
}

func NewManager() *Manager {
	return &Manager{
		clients:              make(map[uint]*Client),
		rooms:                make(map[uint]map[*Client]struct{}),
		channelConversations: make(map[uint]uint),
	}
}

func (m *Manager) Register(client *Client) {
	m.mu.Lock()

	oldClient := m.clients[client.UserID]
	if oldClient != nil {
		m.removeFromRoomsLocked(oldClient)
	}
	m.clients[client.UserID] = client

	m.mu.Unlock()

	if oldClient != nil {
		oldClient.Close()
	}

	log.Printf(
		"user %d registered,online count=%d\n",
		client.UserID,
		m.onlineCount(),
	)
}

func (m *Manager) Unregister(client *Client) {
	m.mu.Lock()

	currentClient, ok := m.clients[client.UserID]
	isCurrent := ok && currentClient == client

	if isCurrent {
		delete(m.clients, client.UserID)
	}
	m.removeFromRoomsLocked(client)

	m.mu.Unlock()

	if isCurrent {

		log.Printf(
			"user %d unregistered,online count=%d\n",
			client.UserID,
			m.onlineCount(),
		)
	}
}

func (m *Manager) Get(userID uint) (*Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	Client, OK := m.clients[userID]
	return Client, OK
}

func (m *Manager) SendToUser(
	userID uint,
	message []byte,
) bool {
	client, ok := m.Get(userID)
	if !ok {
		return false
	}
	return client.SendMessage(message)
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
	m.mu.Lock()
	defer m.mu.Unlock()

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

func (m *Manager) BindChannelConversation(
	channelID uint,
	conversationID uint,
) {
	if channelID == 0 || conversationID == 0 {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.channelConversations[channelID] = conversationID
}

func (m *Manager) BroadcastChannelCompat(
	channelID uint,
	message []byte,
) int {
	m.mu.RLock()

	conversationID, ok := m.channelConversations[channelID]

	m.mu.RUnlock()
	if !ok {
		return 0
	}

	return m.BroadcastToConversation(
		conversationID,
		message,
	)
}

func (m *Manager) RoomUserIDs(
	conversationID uint,
) []uint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	room, ok := m.rooms[conversationID]
	if !ok {
		return nil
	}

	userIDs := make([]uint, 0, len(room))

	for client := range room {
		userIDs = append(
			userIDs,
			client.UserID,
		)
	}

	return userIDs
}
