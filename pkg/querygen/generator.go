// pkg/querygen/generator.go
package querygen

import (
	"github.com/ephy-lab/ai-db-assistant/pkg/chathelpers"
)


// AIResponse represents the structured response from the AI
type AIResponse struct {
	Content string  `json:"content"`
	Query   *string `json:"query,omitempty"`
}

// GenerateResponse generates an AI response with optional SQL query based on user message
func GenerateResponse(userMessage, dbType string) AIResponse {
	lowerMsg := chathelpers.ToLower(userMessage)

	// SELECT queries
	if chathelpers.Contains(lowerMsg, "show", "get", "fetch", "select", "find", "list", "all", "display") {
		return handleSelectQueries(lowerMsg, dbType)
	}

	// COUNT queries
	if chathelpers.Contains(lowerMsg, "count", "how many", "number of", "total") {
		return handleCountQueries(lowerMsg)
	}

	// INSERT queries
	if chathelpers.Contains(lowerMsg, "create", "insert", "add", "new") {
		return handleInsertQueries(lowerMsg)
	}

	// UPDATE queries
	if chathelpers.Contains(lowerMsg, "update", "modify", "change", "edit", "set") {
		return handleUpdateQueries(lowerMsg)
	}

	// DELETE queries
	if chathelpers.Contains(lowerMsg, "delete", "remove", "drop") {
		return handleDeleteQueries(lowerMsg)
	}

	// Default response without query
	return AIResponse{
		Content: "I understand you're asking about your database. Could you be more specific? For example, you can ask me to 'show all users', 'count products', or 'list all tables'.",
		Query:   nil,
	}
}

// handleSelectQueries generates SELECT queries
func handleSelectQueries(msg, dbType string) AIResponse {
	// List tables
	if chathelpers.Contains(msg, "table", "tables", "schema") {
		var query string
		if dbType == "postgresql" {
			query = "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name;"
		} else {
			query = "SHOW TABLES;"
		}
		return AIResponse{
			Content: "Here's a query to list all tables in your database:",
			Query:   &query,
		}
	}

	// List columns
	if chathelpers.Contains(msg, "column", "columns", "field", "fields") {
		var query string
		tableName := extractTableName(msg)
		if tableName == "" {
			tableName = "your_table_name"
		}
		
		if dbType == "postgresql" {
			query = "SELECT column_name, data_type FROM information_schema.columns WHERE table_name = '" + tableName + "';"
		} else {
			query = "DESCRIBE " + tableName + ";"
		}
		return AIResponse{
			Content: "Here's a query to show the columns:",
			Query:   &query,
		}
	}

	// Users table
	if chathelpers.Contains(msg, "user", "users", "customer", "customers") {
		query := "SELECT * FROM users LIMIT 10;"
		return AIResponse{
			Content: "Here's a query to get users from your database:",
			Query:   &query,
		}
	}

	// Products table
	if chathelpers.Contains(msg, "product", "products", "item", "items") {
		query := "SELECT * FROM products LIMIT 10;"
		return AIResponse{
			Content: "Here's a query to get products:",
			Query:   &query,
		}
	}

	// Orders table
	if chathelpers.Contains(msg, "order", "orders", "purchase", "purchases") {
		query := "SELECT * FROM orders LIMIT 10;"
		return AIResponse{
			Content: "Here's a query to get orders:",
			Query:   &query,
		}
	}

	// Generic select
	query := "SELECT * FROM your_table_name LIMIT 10;"
	return AIResponse{
		Content: "Here's a generic query. Please specify the table name:",
		Query:   &query,
	}
}

// handleCountQueries generates COUNT queries
func handleCountQueries(msg string) AIResponse {
	if chathelpers.Contains(msg, "user", "users", "customer", "customers") {
		query := "SELECT COUNT(*) as total_users FROM users;"
		return AIResponse{
			Content: "Here's a query to count all users:",
			Query:   &query,
		}
	}

	if chathelpers.Contains(msg, "product", "products", "item", "items") {
		query := "SELECT COUNT(*) as total_products FROM products;"
		return AIResponse{
			Content: "Here's a query to count all products:",
			Query:   &query,
		}
	}

	if chathelpers.Contains(msg, "order", "orders") {
		query := "SELECT COUNT(*) as total_orders FROM orders;"
		return AIResponse{
			Content: "Here's a query to count all orders:",
			Query:   &query,
		}
	}

	if chathelpers.Contains(msg, "table", "tables") {
		query := "SELECT COUNT(*) as total_tables FROM information_schema.tables WHERE table_schema = 'public';"
		return AIResponse{
			Content: "Here's a query to count all tables:",
			Query:   &query,
		}
	}

	query := "SELECT COUNT(*) as total FROM your_table_name;"
	return AIResponse{
		Content: "Here's a count query. Please specify the table name:",
		Query:   &query,
	}
}

// handleInsertQueries generates INSERT queries
func handleInsertQueries(msg string) AIResponse {
	if chathelpers.Contains(msg, "user", "users", "customer", "customers") {
		query := "INSERT INTO users (name, email, password) VALUES ('John Doe', 'john@example.com', 'hashed_password');"
		return AIResponse{
			Content: "Here's a query to insert a new user (replace with actual values):",
			Query:   &query,
		}
	}

	if chathelpers.Contains(msg, "product", "products", "item", "items") {
		query := "INSERT INTO products (name, description, price) VALUES ('Product Name', 'Description', 99.99);"
		return AIResponse{
			Content: "Here's a query to insert a new product (replace with actual values):",
			Query:   &query,
		}
	}

	if chathelpers.Contains(msg, "order", "orders") {
		query := "INSERT INTO orders (user_id, product_id, quantity, total) VALUES (1, 1, 2, 199.98);"
		return AIResponse{
			Content: "Here's a query to insert a new order (replace with actual values):",
			Query:   &query,
		}
	}

	query := "INSERT INTO your_table_name (column1, column2) VALUES ('value1', 'value2');"
	return AIResponse{
		Content: "Here's a generic insert query. Please specify the table and values:",
		Query:   &query,
	}
}

// handleUpdateQueries generates UPDATE queries
func handleUpdateQueries(msg string) AIResponse {
	if chathelpers.Contains(msg, "user", "users", "customer", "customers") {
		query := "UPDATE users SET name = 'New Name', email = 'newemail@example.com' WHERE id = 1;"
		return AIResponse{
			Content: "Here's a query to update a user (replace with actual values and condition):",
			Query:   &query,
		}
	}

	if chathelpers.Contains(msg, "product", "products", "item", "items") {
		query := "UPDATE products SET price = 149.99, name = 'Updated Product' WHERE id = 1;"
		return AIResponse{
			Content: "Here's a query to update a product (replace with actual values and condition):",
			Query:   &query,
		}
	}

	if chathelpers.Contains(msg, "order", "orders") {
		query := "UPDATE orders SET status = 'shipped', updated_at = NOW() WHERE id = 1;"
		return AIResponse{
			Content: "Here's a query to update an order (replace with actual values and condition):",
			Query:   &query,
		}
	}

	query := "UPDATE your_table_name SET column1 = 'new_value' WHERE id = 1;"
	return AIResponse{
		Content: "Here's a generic update query. Please specify the table, values, and condition:",
		Query:   &query,
	}
}

// handleDeleteQueries generates DELETE queries
func handleDeleteQueries(msg string) AIResponse {
	if chathelpers.Contains(msg, "user", "users", "customer", "customers") {
		query := "DELETE FROM users WHERE id = 1;"
		return AIResponse{
			Content: "⚠️ Here's a query to delete a user. Use with caution:",
			Query:   &query,
		}
	}

	if chathelpers.Contains(msg, "product", "products", "item", "items") {
		query := "DELETE FROM products WHERE id = 1;"
		return AIResponse{
			Content: "⚠️ Here's a query to delete a product. Use with caution:",
			Query:   &query,
		}
	}

	if chathelpers.Contains(msg, "order", "orders") {
		query := "DELETE FROM orders WHERE id = 1;"
		return AIResponse{
			Content: "⚠️ Here's a query to delete an order. Use with caution:",
			Query:   &query,
		}
	}

	query := "DELETE FROM your_table_name WHERE id = 1;"
	return AIResponse{
		Content: "⚠️ Here's a generic delete query. Please specify the table and condition. Use with caution:",
		Query:   &query,
	}
}

// extractTableName attempts to extract table name from message
func extractTableName(msg string) string {
	// Simple extraction - can be enhanced
	if chathelpers.Contains(msg, "users") {
		return "users"
	}
	if chathelpers.Contains(msg, "products") {
		return "products"
	}
	if chathelpers.Contains(msg, "orders") {
		return "orders"
	}
	return ""
}