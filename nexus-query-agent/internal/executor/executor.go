package executor

import (
	"nexus-query-agent/internal/config"
	"nexus-query-agent/internal/models"
)

// Executor interface for database query execution
type Executor interface {
	// Execute runs a query with datasource info and returns paginated results
	Execute(ds *models.DatasourceInfo, query string, page, limit int) (*models.QueryResult, error)
}

// NewExecutor creates appropriate executor based on datasource type
func NewExecutor(dsType string, limits *config.LimitsConfig) (Executor, error) {
	switch dsType {
	case "sap":
		return NewSapExecutor(limits), nil
	// case "mysql":
	// 	return NewMySQLExecutor(limits), nil
	// case "postgres":
	// 	return NewPostgresExecutor(limits), nil
	default:
		return nil, &UnsupportedDatasourceError{Type: dsType}
	}
}

// UnsupportedDatasourceError is returned when datasource type is not supported
type UnsupportedDatasourceError struct {
	Type string
}

func (e *UnsupportedDatasourceError) Error() string {
	return "unsupported datasource type: " + e.Type
}
