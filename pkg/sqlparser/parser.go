// pkg/sqlparser/parser.go
package sqlparser

import (
	"regexp"
	"strings"
)

// QueryType represents the type of SQL query
type QueryType string

const (
	QueryTypeSelect QueryType = "SELECT"
	QueryTypeInsert QueryType = "INSERT"
	QueryTypeUpdate QueryType = "UPDATE"
	QueryTypeDelete QueryType = "DELETE"
	QueryTypeDDL    QueryType = "DDL"
	QueryTypeOther  QueryType = "OTHER"
)

// GetQueryType determines the type of SQL query
func GetQueryType(query string) QueryType {
	query = strings.TrimSpace(strings.ToUpper(query))

	// Remove comments
	query = removeComments(query)

	// Check for DDL operations
	if IsDDLQuery(query) {
		return QueryTypeDDL
	}

	// Check for DML operations
	if strings.HasPrefix(query, "SELECT") || strings.HasPrefix(query, "WITH") {
		return QueryTypeSelect
	}
	if strings.HasPrefix(query, "INSERT") {
		return QueryTypeInsert
	}
	if strings.HasPrefix(query, "UPDATE") {
		return QueryTypeUpdate
	}
	if strings.HasPrefix(query, "DELETE") {
		return QueryTypeDelete
	}

	return QueryTypeOther
}

// IsDDLQuery checks if the query is a DDL operation
func IsDDLQuery(query string) bool {
	query = strings.TrimSpace(strings.ToUpper(query))
	
	ddlKeywords := []string{
		"CREATE", "DROP", "ALTER", "TRUNCATE",
		"RENAME", "COMMENT",
	}

	for _, keyword := range ddlKeywords {
		if strings.HasPrefix(query, keyword) {
			return true
		}
	}

	return false
}

// IsWriteQuery checks if the query is a write operation (INSERT or UPDATE)
func IsWriteQuery(query string) bool {
	queryType := GetQueryType(query)
	return queryType == QueryTypeInsert || queryType == QueryTypeUpdate
}

// IsReadQuery checks if the query is a read operation (SELECT)
func IsReadQuery(query string) bool {
	return GetQueryType(query) == QueryTypeSelect
}

// IsDeleteQuery checks if the query is a DELETE operation
func IsDeleteQuery(query string) bool {
	return GetQueryType(query) == QueryTypeDelete
}

// RequiresPermission checks what permission is required for the query
func RequiresPermission(query string) (ddl, write, read, delete bool) {
	queryType := GetQueryType(query)

	switch queryType {
	case QueryTypeDDL:
		ddl = true
	case QueryTypeInsert, QueryTypeUpdate:
		write = true
	case QueryTypeSelect:
		read = true
	case QueryTypeDelete:
		delete = true
	}

	return
}

// removeComments removes SQL comments from the query
func removeComments(query string) string {
	// Remove single-line comments (-- comment)
	singleLineComment := regexp.MustCompile(`--[^\n]*`)
	query = singleLineComment.ReplaceAllString(query, "")

	// Remove multi-line comments (/* comment */)
	multiLineComment := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	query = multiLineComment.ReplaceAllString(query, "")

	return strings.TrimSpace(query)
}

// GetQueryDescription returns a human-readable description of the query type
func GetQueryDescription(query string) string {
	queryType := GetQueryType(query)

	descriptions := map[QueryType]string{
		QueryTypeSelect: "Data retrieval (SELECT)",
		QueryTypeInsert: "Data insertion (INSERT)",
		QueryTypeUpdate: "Data modification (UPDATE)",
		QueryTypeDelete: "Data deletion (DELETE)",
		QueryTypeDDL:    "Schema modification (DDL)",
		QueryTypeOther:  "Other operation",
	}

	return descriptions[queryType]
}
