package executor

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/SAP/go-hdb/driver"

	"nexus-query-agent/internal/config"
	"nexus-query-agent/internal/models"
)

// SapExecutor handles SAP HANA query execution with dynamic connections
type SapExecutor struct {
	limits *config.LimitsConfig
}

// NewSapExecutor creates a new SAP HANA executor
func NewSapExecutor(limits *config.LimitsConfig) *SapExecutor {
	return &SapExecutor{
		limits: limits,
	}
}

// Execute runs a query using datasource info from the request
func (e *SapExecutor) Execute(ds *models.DatasourceInfo, query string, page, limit int) (*models.QueryResult, error) {
	startTime := time.Now()

	// Connect to SAP HANA using provided credentials
	// For SAP HANA MDC (Multitenant), add databaseName parameter
	var dsn string
	if ds.DatabaseName != "" {
		dsn = fmt.Sprintf("hdb://%s:%s@%s:%d?databaseName=%s",
			ds.Username, ds.Password, ds.Host, ds.Port, ds.DatabaseName)
	} else {
		dsn = fmt.Sprintf("hdb://%s:%s@%s:%d",
			ds.Username, ds.Password, ds.Host, ds.Port)
	}

	db, err := sql.Open("hdb", dsn)
	if err != nil {
		return &models.QueryResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to connect: %v", err),
		}, nil
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return &models.QueryResult{
			Success: false,
			Error:   fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}

	log.Printf("INFO: Connected to SAP HANA at %s:%d (database: %s)", ds.Host, ds.Port, ds.DatabaseName)

	// Apply limits
	if limit <= 0 || limit > e.limits.MaxRows {
		limit = e.limits.MaxRows
	}
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	// Wrap query with pagination (SAP HANA syntax)
	paginatedQuery := fmt.Sprintf(`
		SELECT * FROM (%s) AS subquery
		LIMIT %d OFFSET %d
	`, query, limit, offset)

	// Execute query
	rows, err := db.Query(paginatedQuery)
	if err != nil {
		return &models.QueryResult{
			Success: false,
			Error:   fmt.Sprintf("Query failed: %v", err),
		}, nil
	}
	defer rows.Close()

	// Get column info
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return &models.QueryResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to get columns: %v", err),
		}, nil
	}

	columns := make([]models.ColumnInfo, len(columnTypes))
	for i, ct := range columnTypes {
		nullable, _ := ct.Nullable()
		columns[i] = models.ColumnInfo{
			Name:     ct.Name(),
			Type:     ct.DatabaseTypeName(),
			Nullable: nullable,
		}
	}

	// Scan rows
	data := make([]map[string]any, 0)
	colNames := make([]string, len(columnTypes))
	for i, ct := range columnTypes {
		colNames[i] = ct.Name()
	}

	for rows.Next() {
		values := make([]any, len(colNames))
		valuePtrs := make([]any, len(colNames))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			log.Printf("ERROR: Failed to scan row: %v", err)
			continue
		}

		row := make(map[string]any)
		for i, col := range colNames {
			val := values[i]
			if b, ok := val.([]byte); ok {
				val = string(b)
			}
			row[col] = val
		}
		data = append(data, row)
	}

	// Get total count
	var totalRows int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS subquery", query)
	if err := db.QueryRow(countQuery).Scan(&totalRows); err != nil {
		log.Printf("WARN: Failed to get total count: %v", err)
		totalRows = len(data)
	}

	totalPages := (totalRows + limit - 1) / limit
	executionTime := time.Since(startTime).Milliseconds()

	return &models.QueryResult{
		Success: true,
		Data:    data,
		Columns: columns,
		Pagination: &models.Pagination{
			Page:       page,
			Limit:      limit,
			TotalRows:  totalRows,
			TotalPages: totalPages,
		},
		ExecutionTimeMs: executionTime,
	}, nil
}
