// pkg/dbexecutor/postgres.go
package dbexecutor

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// PostgreSQLExecutor implements DBExecutor for PostgreSQL
type PostgreSQLExecutor struct {
	config     *Config
	db         *sql.DB
	validator  *QueryValidator
}

// NewPostgreSQLExecutor creates a new PostgreSQL executor
func NewPostgreSQLExecutor(config *Config) *PostgreSQLExecutor {
	return &PostgreSQLExecutor{
		config:    config,
		validator: &QueryValidator{},
	}
}

// Connect establishes a connection to PostgreSQL
func (p *PostgreSQLExecutor) Connect() error {
	var connStr string
	
	// Use raw connection string if available (preserves all query parameters like sslmode)
	if p.config.RawConnectionString != "" {
		connStr = p.config.RawConnectionString
	} else {
		// Fallback to building connection string from config
		connStr = fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=require",
			p.config.Host,
			p.config.Port,
			p.config.User,
			p.config.Password,
			p.config.Database,
		)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	p.db = db
	return nil
}

// Disconnect closes the database connection
func (p *PostgreSQLExecutor) Disconnect() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// ExecuteQuery executes a SQL query and returns results
func (p *PostgreSQLExecutor) ExecuteQuery(query string, dryRun bool) (*ExecuteResult, error) {
	if p.db == nil {
		if err := p.Connect(); err != nil {
			return nil, err
		}
	}

	queryType := p.validator.DetectQueryType(query)

	// Handle dry run with EXPLAIN
	if dryRun {
		explainQuery := fmt.Sprintf("EXPLAIN %s", query)
		rows, err := p.db.Query(explainQuery)
		if err != nil {
			return &ExecuteResult{
				Success: false,
				Error:   err.Error(),
				Message: "Dry run failed",
			}, nil
		}
		defer rows.Close()

		var explainLines []string
		for rows.Next() {
			var line string
			if err := rows.Scan(&line); err != nil {
				return &ExecuteResult{
					Success: false,
					Error:   err.Error(),
					Message: "Failed to read EXPLAIN result",
				}, nil
			}
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
		return p.executeSelect(query)
	}

	return p.executeWrite(query, queryType)
}

// executeSelect executes a SELECT query
func (p *PostgreSQLExecutor) executeSelect(query string) (*ExecuteResult, error) {
	rows, err := p.db.Query(query)
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
		// Create a slice of interface{} to hold each column value
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
func (p *PostgreSQLExecutor) executeWrite(query string, queryType QueryType) (*ExecuteResult, error) {
	result, err := p.db.Exec(query)
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
func (p *PostgreSQLExecutor) GetConnectionInfo() *ConnectionInfo {
	return &ConnectionInfo{
		Type:      "PostgreSQL",
		Host:      p.config.Host,
		Port:      p.config.Port,
		Database:  p.config.Database,
		Connected: p.db != nil,
	}
}

// GetSchema retrieves database schema information
func (p *PostgreSQLExecutor) GetSchema() (*SchemaResult, error) {
	if p.db == nil {
		if err := p.Connect(); err != nil {
			return nil, err
		}
	}

	// Get all tables in the public schema
	tableQuery := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	rows, err := p.db.Query(tableQuery)
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
				data_type,
				is_nullable
			FROM information_schema.columns
			WHERE table_schema = 'public' 
			AND table_name = $1
			ORDER BY ordinal_position
		`

		columnRows, err := p.db.Query(columnQuery, tableName)
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
		DBType:     "postgres",
		Database:   p.config.Database,
		Host:       p.config.Host,
		Port:       p.config.Port,
		TableCount: len(tables),
		Tables:     tables,
	}, nil
}

// GetDB returns the underlying database connection
func (p *PostgreSQLExecutor) GetDB() *sql.DB {
	return p.db
}
