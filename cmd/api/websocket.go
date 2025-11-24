package main

import (
	"fmt"
	"log"
	"net/http"
	"sma/cmd/api/database"
	"sma/cmd/api/types"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		// In production, you should restrict this to your frontend domain
		return true
	},
}

// FaucetMessage represents a message sent through the faucet websocket
type FaucetMessage struct {
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	UserID    int       `json:"user_id,omitempty"`
	DeviceID  int       `json:"device_id,omitempty"`
	Data      any       `json:"data,omitempty"`
}

// Client represents a WebSocket client
type Client struct {
	conn   *websocket.Conn
	send   chan FaucetMessage
	userID int
}

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	clients      map[*Client]bool
	broadcast    chan FaucetMessage
	register     chan *Client
	unregister   chan *Client
	db           *database.Database
	metrics      *WebSocketMetrics
	messageLog   []MessageLogEntry
	messageLogMu sync.RWMutex
}

// WebSocketMetrics tracks WebSocket usage
type WebSocketMetrics struct {
	ActiveConnections int64
	TotalMessages     int64
	MessageRate       float64 // messages per second
	LastUpdated       time.Time
	mu                sync.RWMutex
}

// MessageLogEntry represents a logged WebSocket message
type MessageLogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username,omitempty"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Data      any       `json:"data,omitempty"`
}

// Create a global hub instance
var hub *Hub

// InitWebSocketHub initializes and starts the WebSocket hub
func InitWebSocketHub(db *database.Database) {
	hub = &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan FaucetMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		db:         db,
		metrics: &WebSocketMetrics{
			LastUpdated: time.Now(),
		},
		messageLog: make([]MessageLogEntry, 0),
	}
	go hub.Run()
	go hub.updateMetrics()
	log.Println("WebSocket hub initialized and running")
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.updateActiveConnections()
			log.Printf("WebSocket client registered. Total clients: %d", len(h.clients))

			// Send welcome message
			welcomeMsg := FaucetMessage{
				Type:      "welcome",
				Message:   "Connected to faucet stream",
				Timestamp: time.Now(),
			}
			select {
			case client.send <- welcomeMsg:
			default:
				close(client.send)
				delete(h.clients, client)
			}

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.updateActiveConnections()
				log.Printf("WebSocket client unregistered. Total clients: %d", len(h.clients))
			}

		case message := <-h.broadcast:
			h.incrementMessageCount()
			h.logMessage(message)

			// For sensor data messages, only send to users who have favorited the device
			if message.Type == "sensor_data" {
				for client := range h.clients {
					// Check if this user has favorited this device
					if h.userHasFavoritedDevice(client.userID, message.DeviceID) {
						select {
						case client.send <- message:
						default:
							close(client.send)
							delete(h.clients, client)
						}
					}
				}
			} else {
				// For other message types, broadcast to all clients
				for client := range h.clients {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
		}
	}
}

// userHasFavoritedDevice checks if a user has favorited a specific device
func (h *Hub) userHasFavoritedDevice(userID, deviceID int) bool {
	favorites, err := h.db.GetNodeFavoritesByUserID(userID)
	if err != nil {
		log.Printf("Error checking favorites for user %d: %v", userID, err)
		return false
	}

	for _, favorite := range favorites {
		if favorite.DeviceID == deviceID {
			return true
		}
	}
	return false
}

// BroadcastMessage sends a message to all connected clients
func BroadcastMessage(msgType, message string, data any) {
	if hub == nil {
		log.Println("WebSocket hub not initialized")
		return
	}

	msg := FaucetMessage{
		Type:      msgType,
		Message:   message,
		Timestamp: time.Now(),
		Data:      data,
	}

	select {
	case hub.broadcast <- msg:
	default:
		log.Println("Failed to broadcast message: channel full")
	}
}

// BroadcastSensorData broadcasts sensor data only to users who have favorited the device
func BroadcastSensorData(userID, deviceID int, moistureContent *float64) {
	if hub == nil {
		log.Println("WebSocket hub not initialized")
		return
	}

	msg := FaucetMessage{
		Type:      "sensor_data",
		Message:   "New sensor reading received",
		Timestamp: time.Now(),
		UserID:    userID,
		DeviceID:  deviceID,
		Data: map[string]interface{}{
			"moisture_content": moistureContent,
			"device_id":        deviceID,
			"user_id":          userID,
		},
	}

	select {
	case hub.broadcast <- msg:
	default:
		log.Println("Failed to broadcast sensor data: channel full")
	}
}

// FaucetHandler godoc
// @Summary     WebSocket faucet endpoint
// @Description Connect to real-time data stream via WebSocket
// @Tags        faucet
// @Accept      json
// @Produce     json
// @Security    Bearer
// @Success     101 {string} string "Switching Protocols"
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Router      /faucet [get]
func (c *serverConfig) FaucetHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract user from context (set by auth middleware)
	user, ok := getUserFromContext(r)
	if !ok {
		c.unauthorizedResponse(w, r, "User not found in context")
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create new client
	client := &Client{
		conn:   conn,
		send:   make(chan FaucetMessage, 256),
		userID: user.UserID,
	}

	// Register client with hub
	hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second // Increased to 1 minute for unreliable networks

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10 // Send ping roughly every 54 seconds

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		log.Printf("Received pong from client (user %d)", c.userID)
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Also handle ping messages from client
	c.conn.SetPingHandler(func(string) error {
		log.Printf("Received ping from client (user %d), sending pong", c.userID)
		c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		return c.conn.WriteMessage(websocket.PongMessage, nil)
	})

	for {
		var msg FaucetMessage
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket unexpected close error for user %d: %v", c.userID, err)
			} else {
				log.Printf("WebSocket read error for user %d: %v", c.userID, err)
			}
			break
		}

		// Reset read deadline on any message
		c.conn.SetReadDeadline(time.Now().Add(pongWait))

		// Add user ID to message
		msg.UserID = c.userID
		msg.Timestamp = time.Now()

		// Handle different message types
		switch msg.Type {
		case "ping":
			// Respond with pong
			pongMsg := FaucetMessage{
				Type:      "pong",
				Message:   "pong",
				Timestamp: time.Now(),
			}
			select {
			case c.send <- pongMsg:
			default:
				return
			}
		case "echo":
			// Echo the message back
			msg.Type = "echo_response"
			select {
			case c.send <- msg:
			default:
				return
			}
		case "device_update":
			// Broadcast device update to all clients
			deviceUpdate := map[string]interface{}{
				"moisture_content": nil,
				"device_id":        nil,
				"user_id":          nil,
			}

			if dataMap, ok := msg.Data.(map[string]interface{}); ok {
				if deviceID, ok := dataMap["device_id"].(float64); ok {
					deviceUpdate["device_id"] = int(deviceID)
				}
				if userID, ok := dataMap["user_id"].(float64); ok {
					deviceUpdate["user_id"] = int(userID)
				}
				if moistureContent, ok := dataMap["moisture_content"].(float64); ok {
					deviceUpdate["moisture_content"] = moistureContent
				}
				if status, ok := dataMap["status"].(string); ok {
					deviceUpdate["status"] = types.NodeStatus(status)
				}
				if errorDetails, ok := dataMap["error_details"].(string); ok {
					deviceUpdate["error_details"] = &errorDetails
				}
				moistureValue := deviceUpdate["moisture_content"].(float64)

				nodeData := types.NodeData{
					UserID:          deviceUpdate["user_id"].(int),
					DeviceID:        deviceUpdate["device_id"].(int),
					MoistureContent: &moistureValue,
					Timestamp:       time.Now(),
				}

				if err := hub.db.CreateNodeData(&nodeData); err != nil {
					return
				}

				BroadcastSensorData(deviceUpdate["user_id"].(int), deviceUpdate["device_id"].(int), &moistureValue)
			}
		default:
			// Broadcast other messages to all clients
			hub.broadcast <- msg
		}
	}
}

// updateActiveConnections updates the active connections metric
func (h *Hub) updateActiveConnections() {
	h.metrics.mu.Lock()
	defer h.metrics.mu.Unlock()
	h.metrics.ActiveConnections = int64(len(h.clients))
	h.metrics.LastUpdated = time.Now()
}

// incrementMessageCount increments the total message count
func (h *Hub) incrementMessageCount() {
	h.metrics.mu.Lock()
	defer h.metrics.mu.Unlock()
	h.metrics.TotalMessages++
}

// updateMetrics periodically calculates message rate
func (h *Hub) updateMetrics() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	lastCount := int64(0)

	for range ticker.C {
		h.metrics.mu.Lock()
		currentCount := h.metrics.TotalMessages
		h.metrics.MessageRate = float64(currentCount - lastCount)
		lastCount = currentCount
		h.metrics.LastUpdated = time.Now()
		h.metrics.mu.Unlock()
	}
}

// logMessage logs a message for the hawk-eye view (keep last 100 messages)
func (h *Hub) logMessage(msg FaucetMessage) {
	h.messageLogMu.Lock()
	defer h.messageLogMu.Unlock()

	// Username should already be included in the message if available
	// No need to query database here for performance reasons
	username := ""
	if msg.UserID > 0 {
		username = fmt.Sprintf("User %d", msg.UserID)
	}

	entry := MessageLogEntry{
		Timestamp: msg.Timestamp,
		UserID:    msg.UserID,
		Username:  username,
		Type:      msg.Type,
		Message:   msg.Message,
		Data:      msg.Data,
	}

	h.messageLog = append(h.messageLog, entry)

	// Keep only last 100 messages
	if len(h.messageLog) > 100 {
		h.messageLog = h.messageLog[len(h.messageLog)-100:]
	}
}

// GetMetrics returns current WebSocket metrics
func GetMetrics() *WebSocketMetrics {
	if hub == nil {
		return &WebSocketMetrics{}
	}
	hub.metrics.mu.RLock()
	defer hub.metrics.mu.RUnlock()

	return &WebSocketMetrics{
		ActiveConnections: hub.metrics.ActiveConnections,
		TotalMessages:     hub.metrics.TotalMessages,
		MessageRate:       hub.metrics.MessageRate,
		LastUpdated:       hub.metrics.LastUpdated,
	}
}

// GetMessageLog returns recent WebSocket messages
func GetMessageLog() []MessageLogEntry {
	if hub == nil {
		return []MessageLogEntry{}
	}
	hub.messageLogMu.RLock()
	defer hub.messageLogMu.RUnlock()

	// Return a copy of the log
	logCopy := make([]MessageLogEntry, len(hub.messageLog))
	copy(logCopy, hub.messageLog)
	return logCopy
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				log.Printf("Hub closed channel for user %d, closing connection", c.userID)
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error for user %d: %v", c.userID, err)
				return
			}

		case <-ticker.C:
			log.Printf("Sending ping to client (user %d)", c.userID)
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Failed to send ping to user %d: %v", c.userID, err)
				return
			}
		}
	}
}
