package models

// MessageType defines the type of WebSocket message
type MessageType string

const (
	// Agent → Nexus
	MessageTypeRegister  MessageType = "register"
	MessageTypeHeartbeat MessageType = "heartbeat"
	MessageTypeResult    MessageType = "query_result"
	MessageTypeError     MessageType = "error"

	// Nexus → Agent
	MessageTypeRegistered   MessageType = "registered"
	MessageTypeQueryRequest MessageType = "query_request"
	MessageTypePing         MessageType = "ping"
	MessageTypePong         MessageType = "pong"
)

// BaseMessage is the base structure for all messages
type BaseMessage struct {
	Type MessageType `json:"type"`
}

// RegisterMessage is sent by agent to register with Nexus
type RegisterMessage struct {
	Type      MessageType `json:"type"`
	AgentID   string      `json:"agent_id"`
	AgentName string      `json:"agent_name"`
	AgentType string      `json:"agent_type"` // "query"
	Token     string      `json:"token"`
}

// RegisteredMessage is sent by Nexus after successful registration
type RegisteredMessage struct {
	Type    MessageType `json:"type"`
	Status  string      `json:"status"` // "ok" or "error"
	Message string      `json:"message,omitempty"`
}

// HeartbeatMessage is sent by agent to keep connection alive
type HeartbeatMessage struct {
	Type      MessageType `json:"type"`
	AgentID   string      `json:"agent_id"`
	Timestamp int64       `json:"timestamp"`
}

// DatasourceInfo contains connection details sent from Nexus per-request
type DatasourceInfo struct {
	ID           int64  `json:"id"`
	Type         string `json:"type"` // "sap", "mysql", "postgres"
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DatabaseName string `json:"database_name,omitempty"` // For SAP HANA MDC (Multitenant)
	Database     string `json:"database,omitempty"`
	Username     string `json:"username"`
	Password     string `json:"password"` // Decrypted by Nexus Core
}

// QueryRequest is sent by Nexus to request query execution
// Now includes datasource connection details
type QueryRequest struct {
	Type       MessageType    `json:"type"`
	RequestID  string         `json:"request_id"`
	Datasource DatasourceInfo `json:"datasource"` // Connection details from Nexus
	QueryType  string         `json:"query_type"` // "select", "insert", "update", "delete"
	Query      string         `json:"query"`
	Params     []any          `json:"params,omitempty"` // For parameterized queries
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
}

// QueryResult is sent by agent with query results
type QueryResult struct {
	Type            MessageType      `json:"type"`
	RequestID       string           `json:"request_id"`
	Success         bool             `json:"success"`
	QueryType       string           `json:"query_type,omitempty"` // "select", "insert", "update", "delete"
	Data            []map[string]any `json:"data,omitempty"`
	Columns         []ColumnInfo     `json:"columns,omitempty"`
	Pagination      *Pagination      `json:"pagination,omitempty"`
	AffectedRows    int64            `json:"affected_rows,omitempty"` // For DML operations
	ExecutionTimeMs int64            `json:"execution_time_ms"`
	Error           string           `json:"error,omitempty"`
}

// ColumnInfo describes a column in the result
type ColumnInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
}

// Pagination contains pagination info
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalRows  int `json:"total_rows"`
	TotalPages int `json:"total_pages"`
}

// ErrorMessage is sent when an error occurs
type ErrorMessage struct {
	Type      MessageType `json:"type"`
	RequestID string      `json:"request_id,omitempty"`
	Code      string      `json:"code"`
	Message   string      `json:"message"`
}
