package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Type       string      `json:"type"`
	DocumentID uuid.UUID   `json:"document_id"`
	UserID     uuid.UUID   `json:"user_id"`
	Content    interface{} `json:"content"`
	Timestamp  time.Time   `json:"timestamp"`
}

type Client struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Conn       *websocket.Conn
	DocumentID uuid.UUID
	Send       chan []byte
	Manager    *Manager
}

type Room struct {
	DocumentID uuid.UUID
	Clients    map[*Client]bool
	Broadcast  chan Message
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

type Manager struct {
	rooms      map[uuid.UUID]*Room
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		rooms:      make(map[uuid.UUID]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (m *Manager) Run() {
	for {
		select {
		case client := <-m.register:
			m.getOrCreateRoom(client.DocumentID).RegisterClient(client)
		case client := <-m.unregister:
			m.getOrCreateRoom(client.DocumentID).UnregisterClient(client)
		}
	}
}

func (m *Manager) getOrCreateRoom(docID uuid.UUID) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	if room, exists := m.rooms[docID]; exists {
		return room
	}

	room := &Room{
		DocumentID: docID,
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan Message, 100),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}

	go room.Run()
	m.rooms[docID] = room
	return room
}

func (m *Manager) BroadcastToDocument(docID uuid.UUID, msg Message) {
	m.mu.RLock()
	room, exists := m.rooms[docID]
	m.mu.RUnlock()

	if exists {
		room.Broadcast <- msg
	}
}

func (m *Manager) GetCollaborators(docID uuid.UUID) []*Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	room, exists := m.rooms[docID]
	if !exists {
		return nil
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	var collaborators []*Client
	for client := range room.Clients {
		collaborators = append(collaborators, client)
	}
	return collaborators
}

func (r *Room) Run() {
	for {
		select {
		case client := <-r.Register:
			r.RegisterClient(client)
		case client := <-r.Unregister:
			r.UnregisterClient(client)
		case message := <-r.Broadcast:
			r.broadcastMessage(message)
		}
	}
}

func (r *Room) RegisterClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Clients[client] = true
	log.Printf("Client %s joined document %s", client.ID, r.DocumentID)

	r.broadcastUserList()
}

func (r *Room) UnregisterClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.Clients[client]; exists {
		delete(r.Clients, client)
		close(client.Send)
		log.Printf("Client %s left document %s", client.ID, r.DocumentID)
	}

	r.broadcastUserList()
}

func (r *Room) broadcastMessage(msg Message) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	for client := range r.Clients {
		select {
		case client.Send <- data:
		default:
			close(client.Send)
			delete(r.Clients, client)
		}
	}
}

func (r *Room) broadcastUserList() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var users []uuid.UUID
	for client := range r.Clients {
		users = append(users, client.UserID)
	}

	msg := Message{
		Type:       "user_list",
		DocumentID: r.DocumentID,
		Content:    users,
		Timestamp:  time.Now(),
	}

	data, _ := json.Marshal(msg)
	for client := range r.Clients {
		select {
		case client.Send <- data:
		default:
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Manager.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		msg.UserID = c.UserID
		msg.DocumentID = c.DocumentID
		msg.Timestamp = time.Now()

		c.Manager.getOrCreateRoom(c.DocumentID).Broadcast <- msg
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

type WSServer struct {
	manager *Manager
	port    string
}

func NewWSServer(manager *Manager, port string) *WSServer {
	return &WSServer{
		manager: manager,
		port:    port,
	}
}

func (ws *WSServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	docIDStr := r.URL.Query().Get("document_id")
	userIDStr := r.URL.Query().Get("user_id")

	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		ID:         uuid.New(),
		UserID:     userID,
		Conn:       conn,
		DocumentID: docID,
		Send:       make(chan []byte, 256),
		Manager:    ws.manager,
	}

	ws.manager.register <- client

	go client.WritePump()
	go client.ReadPump()
}

func (ws *WSServer) StartServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", ws.HandleWebSocket)

	server := &http.Server{
		Addr:    ":" + ws.port,
		Handler: mux,
	}

	return server.ListenAndServe()
}
