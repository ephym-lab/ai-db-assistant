// pkg/dbexecutor/factory.go
package dbexecutor

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// NewExecutorFromConnectionString creates a database executor from a connection string
func NewExecutorFromConnectionString(dbType, connectionString string) (DBExecutor, error) {
	// Normalize database type
	dbType = strings.ToLower(strings.TrimSpace(dbType))

	// Validate database type
	if dbType != "postgresql" && dbType != "postgres" && dbType != "mysql" {
		return nil, fmt.Errorf("unsupported database type: %s. Supported types: postgresql, mysql", dbType)
	}

	// Parse the connection string
	config, err := parseConnectionString(connectionString)
	if err != nil {
		return nil, fmt.Errorf("invalid connection string: %w", err)
	}

	// Create executor based on database type
	if dbType == "postgresql" || dbType == "postgres" {
		return NewPostgreSQLExecutor(config), nil
	} else if dbType == "mysql" {
		return NewMySQLExecutor(config), nil
	}

	return nil, fmt.Errorf("unsupported database type: %s", dbType)
}

// parseConnectionString parses a database connection string
// Supports formats:
// - PostgreSQL: postgresql://user:password@host:port/database
// - MySQL: mysql://user:password@host:port/database
func parseConnectionString(connectionString string) (*Config, error) {
	// Parse the URL
	parsedURL, err := url.Parse(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Extract user and password
	user := parsedURL.User.Username()
	password, _ := parsedURL.User.Password()

	// Extract host and port
	host := parsedURL.Hostname()
	if host == "" {
		host = "localhost"
	}

	portStr := parsedURL.Port()
	var port int
	if portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port number: %s", portStr)
		}
	} else {
		// Default ports
		scheme := strings.ToLower(parsedURL.Scheme)
		if scheme == "postgresql" || scheme == "postgres" {
			port = 5432
		} else if scheme == "mysql" {
			port = 3306
		}
	}

	// Extract database name
	database := strings.TrimPrefix(parsedURL.Path, "/")
	if database == "" {
		return nil, fmt.Errorf("database name is required in connection string")
	}

	// Validate required fields
	if user == "" {
		return nil, fmt.Errorf("username is required in connection string")
	}

	return &Config{
		Host:     host,
		Port:     port,
		Database: database,
		User:     user,
		Password: password,
	}, nil
}

// NewExecutor creates a database executor from a config
func NewExecutor(dbType string, config *Config) (DBExecutor, error) {
	dbType = strings.ToLower(strings.TrimSpace(dbType))

	if dbType == "postgresql" || dbType == "postgres" {
		return NewPostgreSQLExecutor(config), nil
	} else if dbType == "mysql" {
		return NewMySQLExecutor(config), nil
	}

	return nil, fmt.Errorf("unsupported database type: %s", dbType)
}
