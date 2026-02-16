// pkg/dbexecutor/mysql.go
package dbexecutor

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLExecutor implements DBExecutor for MySQL
type MySQLExecutor struct {
	config    *Config
	db        *sql.DB
	validator *QueryValidator
}

// NewMySQLExecutor creates a new MySQL executor
func NewMySQLExecutor(config *Config) *MySQLExecutor {
	return &MySQLExecutor{
		config:    config,
		validator: &QueryValidator{},
	}
}

// Connect establishes a connection to MySQL
func (m *MySQLExecutor) Connect() error {
	// Format: user:password@tcp(host:port)/database
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true",
		m.config.User,
		m.config.Password,
		m.config.Host,
		m.config.Port,
		m.config.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open MySQL connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping MySQL: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	m.db = db
	return nil
}

// Disconnect closes the database connection
func (m *MySQLExecutor) Disconnect() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// ExecuteQuery executes a SQL query and returns results
func (m *MySQLExecutor) ExecuteQuery(query string, dryRun bool) (*ExecuteResult, error) {
	if m.db == nil {
		if err := m.Connect(); err != nil {
			return nil, err
		}
	}

	queryType := m.validator.DetectQueryType(query)

	// Handle dry run with EXPLAIN
	if dryRun {
		explainQuery := fmt.Sprintf("EXPLAIN %s", query)
		rows, err := m.db.Query(explainQuery)
		if err != nil {
			return &ExecuteResult{
				Success: false,
				Error:   err.Error(),
				Message: "Dry run failed",
			}, nil
		}
		defer rows.Close()

		// Get column names for EXPLAIN result
		columns, err := rows.Columns()
		if err != nil {
			return &ExecuteResult{
				Success: false,
				Error:   err.Error(),
				Message: "Failed to get EXPLAIN columns",
			}, nil
		}

		var explainLines []string
		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return &ExecuteResult{
					Success: false,
					Error:   err.Error(),
					Message: "Failed to read EXPLAIN result",
				}, nil
			}

			// Format the EXPLAIN output
			line := fmt.Sprintf("%v", values)
			explainLines = append(explainLines, line)
		}

		return &ExecuteResult{
			Success: true,
			DryRun:  true,
			Explain: explainLines,
			Message: "Dry run completed (query not executed)",
		}, nil
	}

	// Execute the actual query
	if queryType == QueryTypeSelect {
		return m.executeSelect(query)
	}

	return m.executeWrite(query, queryType)
}

// executeSelect executes a SELECT query
func (m *MySQLExecutor) executeSelect(query string) (*ExecuteResult, error) {
	rows, err := m.db.Query(query)
	if err != nil {
		return &ExecuteResult{
			Success: false,
			Error:   err.Error(),
			Message: "Query execution failed",
		}, nil
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return &ExecuteResult{
			Success: false,
			Error:   err.Error(),
			Message: "Failed to get column names",
		}, nil
	}

	// Fetch all rows
	var results [][]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return &ExecuteResult{
				Success: false,
				Error:   err.Error(),
				Message: "Failed to scan row",
			}, nil
		}

		// Convert byte arrays to strings for JSON serialization
		row := make([]interface{}, len(values))
		for i, v := range values {
			if b, ok := v.([]byte); ok {
				row[i] = string(b)
			} else {
				row[i] = v
			}
		}

		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return &ExecuteResult{
			Success: false,
			Error:   err.Error(),
			Message: "Error iterating rows",
		}, nil
	}

	return &ExecuteResult{
		Success:   true,
		QueryType: string(QueryTypeSelect),
		Columns:   columns,
		Rows:      results,
		RowCount:  len(results),
	}, nil
}

// executeWrite executes INSERT, UPDATE, DELETE, or DDL queries
func (m *MySQLExecutor) executeWrite(query string, queryType QueryType) (*ExecuteResult, error) {
	result, err := m.db.Exec(query)
	if err != nil {
		return &ExecuteResult{
			Success: false,
			Error:   err.Error(),
			Message: "Query execution failed",
		}, nil
	}

	affectedRows, _ := result.RowsAffected()

	return &ExecuteResult{
		Success:      true,
		QueryType:    string(queryType),
		AffectedRows: affectedRows,
		Message:      fmt.Sprintf("%s operation completed successfully", queryType),
	}, nil
}

// GetConnectionInfo returns database connection information
func (m *MySQLExecutor) GetConnectionInfo() *ConnectionInfo {
	return &ConnectionInfo{
		Type:      "MySQL",
		Host:      m.config.Host,
		Port:      m.config.Port,
		Database:  m.config.Database,
		Connected: m.db != nil,
	}
}

// GetSchema retrieves database schema information
func (m *MySQLExecutor) GetSchema() (*SchemaResult, error) {
	if m.db == nil {
		if err := m.Connect(); err != nil {
			return nil, err
		}
	}

	// Get all tables
	tableQuery := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = ? 
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	rows, err := m.db.Query(tableQuery, m.config.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tables: %w", err)
	}
	defer rows.Close()

	var tableNames []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tableNames = append(tableNames, tableName)
	}

	// For each table, get column information
	var tables []TableInfo
	for _, tableName := range tableNames {
		columnQuery := `
			SELECT 
				column_name,
				column_type,
				is_nullable
			FROM information_schema.columns
			WHERE table_schema = ? 
			AND table_name = ?
			ORDER BY ordinal_position
		`

		columnRows, err := m.db.Query(columnQuery, m.config.Database, tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch columns for table %s: %w", tableName, err)
		}

		var columns []ColumnInfo
		for columnRows.Next() {
			var colName, colType, isNullable string
			if err := columnRows.Scan(&colName, &colType, &isNullable); err != nil {
				columnRows.Close()
				return nil, fmt.Errorf("failed to scan column info: %w", err)
			}

			columns = append(columns, ColumnInfo{
				Name:     colName,
				Type:     colType,
				Nullable: isNullable == "YES",
			})
		}
		columnRows.Close()

		tables = append(tables, TableInfo{
			Name:    tableName,
			Columns: columns,
		})
	}

	return &SchemaResult{
		DBType:     "mysql",
		Database:   m.config.Database,
		Host:       m.config.Host,
		Port:       m.config.Port,
		TableCount: len(tables),
		Tables:     tables,
	}, nil
}

// GetDB returns the underlying database connection
func (m *MySQLExecutor) GetDB() *sql.DB {
	return m.db
}
