// pkg/dbexecutor/types.go
package dbexecutor

import (
	"database/sql"
	"fmt"
	"strings"
)

// QueryType represents the type of SQL query
type QueryType string

const (
	QueryTypeSelect  QueryType = "SELECT"
	QueryTypeInsert  QueryType = "INSERT"
	QueryTypeUpdate  QueryType = "UPDATE"
	QueryTypeDelete  QueryType = "DELETE"
	QueryTypeDDL     QueryType = "DDL"
	QueryTypeUnknown QueryType = "UNKNOWN"
)

// DDL keywords
var ddlKeywords = []string{"CREATE", "DROP", "ALTER", "TRUNCATE", "RENAME"}

// QueryValidator validates and classifies SQL queries
type QueryValidator struct{}

// DetectQueryType detects the type of SQL query
func (qv *QueryValidator) DetectQueryType(query string) QueryType {
	queryUpper := strings.TrimSpace(strings.ToUpper(query))

	// Check for DDL
	for _, keyword := range ddlKeywords {
		if strings.HasPrefix(queryUpper, keyword) {
			return QueryTypeDDL
		}
	}

	// Check for write operations
	if strings.HasPrefix(queryUpper, "INSERT") {
		return QueryTypeInsert
	} else if strings.HasPrefix(queryUpper, "UPDATE") {
		return QueryTypeUpdate
	} else if strings.HasPrefix(queryUpper, "DELETE") {
		return QueryTypeDelete
	} else if strings.HasPrefix(queryUpper, "SELECT") {
		return QueryTypeSelect
	}

	return QueryTypeUnknown
}

// ValidateQuery validates if a query is safe to execute based on configuration
func (qv *QueryValidator) ValidateQuery(query string, allowWrite, allowDDL bool) (bool, string) {
	if strings.TrimSpace(query) == "" {
		return false, "Query cannot be empty"
	}

	queryType := qv.DetectQueryType(query)

	if queryType == QueryTypeUnknown {
		return false, "Unknown or unsupported query type"
	}

	if queryType == QueryTypeDDL && !allowDDL {
		return false, fmt.Sprintf("DDL operations are disabled. Query type: %s", queryType)
	}

	if (queryType == QueryTypeInsert || queryType == QueryTypeUpdate || queryType == QueryTypeDelete) && !allowWrite {
		return false, fmt.Sprintf("Write operations are disabled. Query type: %s", queryType)
	}

	return true, ""
}

// AddLimitIfNeeded adds LIMIT clause to SELECT queries if not present
func (qv *QueryValidator) AddLimitIfNeeded(query string, maxRows int) string {
	queryUpper := strings.TrimSpace(strings.ToUpper(query))

	if !strings.HasPrefix(queryUpper, "SELECT") {
		return query
	}

	// Check if LIMIT already exists
	if strings.Contains(queryUpper, "LIMIT") {
		return query
	}

	// Add LIMIT
	query = strings.TrimSuffix(strings.TrimSpace(query), ";")
	return fmt.Sprintf("%s LIMIT %d;", query, maxRows)
}

// ExecuteResult represents the result of a query execution
type ExecuteResult struct {
	Success      bool           `json:"success"`
	QueryType    string         `json:"query_type,omitempty"`
	Columns      []string       `json:"columns,omitempty"`
	Rows         [][]interface{} `json:"rows,omitempty"`
	RowCount     int            `json:"row_count,omitempty"`
	AffectedRows int64          `json:"affected_rows,omitempty"`
	Message      string         `json:"message,omitempty"`
	Error        string         `json:"error,omitempty"`
	DryRun       bool           `json:"dry_run,omitempty"`
	Explain      []string       `json:"explain,omitempty"`
}

// ConnectionInfo represents database connection information
type ConnectionInfo struct {
	Type      string `json:"type"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Database  string `json:"database"`
	Connected bool   `json:"connected"`
}

// TableInfo represents table schema information
type TableInfo struct {
	Name    string       `json:"name"`
	Columns []ColumnInfo `json:"columns"`
}

// ColumnInfo represents column information
type ColumnInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
}

// SchemaResult represents the result of schema retrieval
type SchemaResult struct {
	DBType     string      `json:"db_type"`
	Database   string      `json:"database"`
	Host       string      `json:"host"`
	Port       int         `json:"port"`
	TableCount int         `json:"table_count"`
	Tables     []TableInfo `json:"tables"`
}

// DBExecutor is the interface for database executors
type DBExecutor interface {
	// Connect establishes a database connection
	Connect() error

	// Disconnect closes the database connection
	Disconnect() error

	// ExecuteQuery executes a SQL query and returns results
	ExecuteQuery(query string, dryRun bool) (*ExecuteResult, error)

	// GetConnectionInfo returns database connection information
	GetConnectionInfo() *ConnectionInfo

	// GetSchema retrieves database schema information
	GetSchema() (*SchemaResult, error)

	// GetDB returns the underlying database connection
	GetDB() *sql.DB
}

// Config represents database configuration
type Config struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}
