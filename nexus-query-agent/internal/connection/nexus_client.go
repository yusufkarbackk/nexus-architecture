package connection

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"nexus-query-agent/internal/config"
	"nexus-query-agent/internal/models"
)

// NexusClient manages WebSocket connection to Nexus Core
type NexusClient struct {
	config      *config.Config
	conn        *websocket.Conn
	mu          sync.Mutex
	isConnected bool
	done        chan struct{}

	// Handler for incoming query requests
	OnQueryRequest func(req *models.QueryRequest)
}

// NewNexusClient creates a new Nexus client
func NewNexusClient(cfg *config.Config) *NexusClient {
	return &NexusClient{
		config: cfg,
		done:   make(chan struct{}),
	}
}

// Connect establishes WebSocket connection to Nexus Core
func (c *NexusClient) Connect() error {
	log.Printf("INFO: Connecting to Nexus Core at %s", c.config.Nexus.CoreURL)

	conn, _, err := websocket.DefaultDialer.Dial(c.config.Nexus.CoreURL, nil)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.isConnected = true
	c.mu.Unlock()

	log.Printf("INFO: Connected to Nexus Core")

	// Send registration message
	if err := c.register(); err != nil {
		return err
	}

	// Start goroutines
	go c.readLoop()
	go c.heartbeatLoop()

	return nil
}

// register sends registration message to Nexus
func (c *NexusClient) register() error {
	msg := models.RegisterMessage{
		Type:      models.MessageTypeRegister,
		AgentID:   c.config.Agent.ID,
		AgentName: c.config.Agent.Name,
		AgentType: "query",
		Token:     c.config.Agent.Token,
	}

	return c.sendJSON(msg)
}

// readLoop handles incoming messages
func (c *NexusClient) readLoop() {
	defer c.Close()

	for {
		select {
		case <-c.done:
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				log.Printf("ERROR: Read error: %v", err)
				return
			}

			c.handleMessage(message)
		}
	}
}

// handleMessage processes incoming messages
func (c *NexusClient) handleMessage(data []byte) {
	var base models.BaseMessage
	if err := json.Unmarshal(data, &base); err != nil {
		log.Printf("ERROR: Failed to parse message: %v", err)
		return
	}

	switch base.Type {
	case models.MessageTypeRegistered:
		var msg models.RegisteredMessage
		if err := json.Unmarshal(data, &msg); err == nil {
			log.Printf("INFO: Registration %s: %s", msg.Status, msg.Message)
		}

	case models.MessageTypeQueryRequest:
		var req models.QueryRequest
		if err := json.Unmarshal(data, &req); err == nil {
			log.Printf("INFO: Received query request: %s for %s:%d",
				req.RequestID, req.Datasource.Host, req.Datasource.Port)
			if c.OnQueryRequest != nil {
				go c.OnQueryRequest(&req)
			}
		}

	case models.MessageTypePing:
		c.sendJSON(models.BaseMessage{Type: models.MessageTypePong})

	default:
		log.Printf("DEBUG: Unknown message type: %s", base.Type)
	}
}

// heartbeatLoop sends periodic heartbeats
func (c *NexusClient) heartbeatLoop() {
	ticker := time.NewTicker(c.config.Nexus.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			msg := models.HeartbeatMessage{
				Type:      models.MessageTypeHeartbeat,
				AgentID:   c.config.Agent.ID,
				Timestamp: time.Now().Unix(),
			}
			if err := c.sendJSON(msg); err != nil {
				log.Printf("ERROR: Heartbeat failed: %v", err)
			}
		}
	}
}

// SendResult sends query result to Nexus
func (c *NexusClient) SendResult(result *models.QueryResult) error {
	result.Type = models.MessageTypeResult
	return c.sendJSON(result)
}

// SendError sends error message to Nexus
func (c *NexusClient) SendError(requestID, code, message string) error {
	msg := models.ErrorMessage{
		Type:      models.MessageTypeError,
		RequestID: requestID,
		Code:      code,
		Message:   message,
	}
	return c.sendJSON(msg)
}

// sendJSON sends a JSON message
func (c *NexusClient) sendJSON(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}

	return c.conn.WriteJSON(v)
}

// Close closes the connection
func (c *NexusClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.isConnected = false

	select {
	case <-c.done:
		// Already closed
	default:
		close(c.done)
	}
}

// IsConnected returns connection status
func (c *NexusClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isConnected
}

// Reconnect attempts to reconnect with backoff
func (c *NexusClient) Reconnect() {
	for {
		log.Printf("INFO: Attempting to reconnect in %s...", c.config.Nexus.ReconnectInterval)
		time.Sleep(c.config.Nexus.ReconnectInterval)

		c.done = make(chan struct{})
		if err := c.Connect(); err != nil {
			log.Printf("ERROR: Reconnect failed: %v", err)
			continue
		}

		log.Printf("INFO: Reconnected successfully")
		return
	}
}
