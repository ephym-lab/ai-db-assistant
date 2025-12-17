// pkg/proxyclient/client.go
package proxyclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles communication with the AI SQL Assistant proxy server
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new proxy client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateSQLRequest represents the request to generate SQL
type GenerateSQLRequest struct {
	Question string `json:"question"`
	DBType   string `json:"db_type,omitempty"`
	DBSchema string `json:"db_schema,omitempty"`
}

// GenerateSQLResponse represents the response from SQL generation
type GenerateSQLResponse struct {
	Content string `json:"content"`
	Query   string `json:"query"`
}

// ConnectDBRequest represents the request to connect to a database
type ConnectDBRequest struct {
	DBType           string `json:"db_type"`
	ConnectionString string `json:"connection_string"`
}

// ConnectionInfo represents database connection information
type ConnectionInfo struct {
	Type      string `json:"type"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Database  string `json:"database"`
	Connected bool   `json:"connected"`
}

// ConnectDBResponse represents the response from database connection
type ConnectDBResponse struct {
	Success        bool           `json:"success"`
	Message        string         `json:"message"`
	ConnectionInfo ConnectionInfo `json:"connection_info"`
}

// DisconnectDBResponse represents the response from database disconnection
type DisconnectDBResponse struct {
	Success            bool           `json:"success"`
	Message            string         `json:"message"`
	PreviousConnection ConnectionInfo `json:"previous_connection"`
}

// ExecuteSQLRequest represents the request to execute SQL
type ExecuteSQLRequest struct {
	Query  string `json:"query"`
	DryRun bool   `json:"dry_run,omitempty"`
}

// ExecuteSQLResponse represents the response from SQL execution
type ExecuteSQLResponse struct {
	Success      bool          `json:"success"`
	QueryType    string        `json:"query_type,omitempty"`
	Columns      []string      `json:"columns,omitempty"`
	Rows         [][]any       `json:"rows,omitempty"`
	RowCount     int           `json:"row_count,omitempty"`
	AffectedRows int           `json:"affected_rows,omitempty"`
	Message      string        `json:"message,omitempty"`
	DryRun       bool          `json:"dry_run,omitempty"`
	Explain      []string      `json:"explain,omitempty"`
}

// ValidateSQLRequest represents the request to validate SQL
type ValidateSQLRequest struct {
	Query string `json:"query"`
}

// ValidateSQLResponse represents the response from SQL validation
type ValidateSQLResponse struct {
	Success bool     `json:"success"`
	DryRun  bool     `json:"dry_run"`
	Explain []string `json:"explain"`
	Message string   `json:"message"`
}

// DBInfoResponse represents the database info response
type DBInfoResponse struct {
	Type      string `json:"type"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Database  string `json:"database"`
	Connected bool   `json:"connected"`
}

// ErrorResponse represents an error response from the proxy
type ErrorResponse struct {
	Detail string `json:"detail"`
}

// GenerateSQL calls the /generate-sql endpoint
func (c *Client) GenerateSQL(question, dbType, dbSchema string) (*GenerateSQLResponse, error) {
	req := GenerateSQLRequest{
		Question: question,
		DBType:   dbType,
		DBSchema: dbSchema,
	}

	var resp GenerateSQLResponse
	if err := c.post("/generate-sql", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// ConnectDB calls the /connect-db endpoint
func (c *Client) ConnectDB(dbType, connectionString string) (*ConnectDBResponse, error) {
	req := ConnectDBRequest{
		DBType:           dbType,
		ConnectionString: connectionString,
	}

	var resp ConnectDBResponse
	if err := c.post("/connect-db", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// DisconnectDB calls the /disconnect-db endpoint
func (c *Client) DisconnectDB() (*DisconnectDBResponse, error) {
	var resp DisconnectDBResponse
	if err := c.post("/disconnect-db", nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// ExecuteSQL calls the /execute-sql endpoint
func (c *Client) ExecuteSQL(query string, dryRun bool) (*ExecuteSQLResponse, error) {
	req := ExecuteSQLRequest{
		Query:  query,
		DryRun: dryRun,
	}

	var resp ExecuteSQLResponse
	if err := c.post("/execute-sql", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// ValidateSQL calls the /validate-sql endpoint
func (c *Client) ValidateSQL(query string) (*ValidateSQLResponse, error) {
	req := ValidateSQLRequest{
		Query: query,
	}

	var resp ValidateSQLResponse
	if err := c.post("/validate-sql", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetDBInfo calls the /db-info endpoint
func (c *Client) GetDBInfo() (*DBInfoResponse, error) {
	url := c.BaseURL + "/db-info"

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get database info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var dbInfo DBInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&dbInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &dbInfo, nil
}

// post is a helper method to make POST requests
func (c *Client) post(endpoint string, request interface{}, response interface{}) error {
	url := c.BaseURL + endpoint

	var body io.Reader
	if request != nil {
		jsonData, err := json.Marshal(request)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if request != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// parseError parses error responses from the proxy
func (c *Client) parseError(resp *http.Response) error {
	var errResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return fmt.Errorf("proxy request failed with status %d", resp.StatusCode)
	}
	return fmt.Errorf("proxy error (%d): %s", resp.StatusCode, errResp.Detail)
}
