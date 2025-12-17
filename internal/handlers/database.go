// internal/handlers/database.go
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/ephy-lab/ai-db-assistant/internal/config"
	"github.com/ephy-lab/ai-db-assistant/internal/middleware"
	"github.com/ephy-lab/ai-db-assistant/internal/models"
	"github.com/ephy-lab/ai-db-assistant/pkg/proxyclient"
	"github.com/ephy-lab/ai-db-assistant/pkg/response"
	"github.com/ephy-lab/ai-db-assistant/pkg/sqlparser"
	"gorm.io/gorm"
)

type DatabaseHandler struct {
	db          *gorm.DB
	proxyClient *proxyclient.Client
}

func NewDatabaseHandler(db *gorm.DB, cfg *config.Config) *DatabaseHandler {
	return &DatabaseHandler{
		db:          db,
		proxyClient: proxyclient.NewClient(cfg.ProxyServerURL),
	}
}

type ExecuteSQLRequest struct {
	Query  string `json:"query"`
	DryRun bool   `json:"dry_run,omitempty"`
}

type ValidateSQLRequest struct {
	Query string `json:"query"`
}

// ConnectDB establishes a connection to the project's database
func (h *DatabaseHandler) ConnectDB(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	projectID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Verify project ownership
	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(w, http.StatusNotFound, "Project not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to fetch project")
		return
	}

	// Connect to database via proxy
	resp, err := h.proxyClient.ConnectDB(project.DatabaseType, project.ConnectionString)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to connect to database: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// DisconnectDB closes the database connection
func (h *DatabaseHandler) DisconnectDB(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	projectID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Verify project ownership
	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(w, http.StatusNotFound, "Project not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to fetch project")
		return
	}

	// Disconnect from database via proxy
	resp, err := h.proxyClient.DisconnectDB()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to disconnect from database: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// ExecuteSQL executes a SQL query with permission checks
func (h *DatabaseHandler) ExecuteSQL(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	projectID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Verify project ownership and load permissions
	var project models.Project
	if err := h.db.Preload("Permission").Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(w, http.StatusNotFound, "Project not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to fetch project")
		return
	}

	var req ExecuteSQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Query == "" {
		response.Error(w, http.StatusBadRequest, "Query is required")
		return
	}

	// Check permissions based on query type
	needsDDL, needsWrite, needsRead, needsDelete := sqlparser.RequiresPermission(req.Query)

	if project.Permission != nil {
		if needsDDL && !project.Permission.AllowDDL {
			response.Error(w, http.StatusForbidden, "DDL operations are not allowed for this project")
			return
		}
		if needsWrite && !project.Permission.AllowWrite {
			response.Error(w, http.StatusForbidden, "Write operations are not allowed for this project")
			return
		}
		if needsRead && !project.Permission.AllowRead {
			response.Error(w, http.StatusForbidden, "Read operations are not allowed for this project")
			return
		}
		if needsDelete && !project.Permission.AllowDelete {
			response.Error(w, http.StatusForbidden, "Delete operations are not allowed for this project")
			return
		}
	}

	// Execute query via proxy
	startTime := time.Now()
	resp, err := h.proxyClient.ExecuteSQL(req.Query, req.DryRun)
	executionTime := time.Since(startTime).Milliseconds()

	// Log query execution
	queryType := string(sqlparser.GetQueryType(req.Query))
	queryLog := models.Query{
		ProjectID:     uint(projectID),
		Query:         req.Query,
		QueryType:     queryType,
		ExecutionTime: int(executionTime),
	}

	if err != nil {
		queryLog.Status = "error"
		queryLog.Error = err.Error()
		h.db.Create(&queryLog)

		response.Error(w, http.StatusInternalServerError, "Failed to execute query: "+err.Error())
		return
	}

	// Update query log with results
	queryLog.Status = "success"
	if resp.RowCount > 0 {
		queryLog.RowsAffected = resp.RowCount
		resultJSON, _ := json.Marshal(resp)
		queryLog.Result = string(resultJSON)
	} else if resp.AffectedRows > 0 {
		queryLog.RowsAffected = resp.AffectedRows
		queryLog.Result = fmt.Sprintf("Affected rows: %d", resp.AffectedRows)
	} else if resp.Message != "" {
		queryLog.Result = resp.Message
	}

	h.db.Create(&queryLog)

	response.JSON(w, http.StatusOK, resp)
}

// ValidateSQL validates a SQL query without executing it
func (h *DatabaseHandler) ValidateSQL(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	projectID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Verify project ownership and load permissions
	var project models.Project
	if err := h.db.Preload("Permission").Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(w, http.StatusNotFound, "Project not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to fetch project")
		return
	}

	// Require read permission for validation (uses EXPLAIN)
	if project.Permission != nil && !project.Permission.AllowRead {
		response.Error(w, http.StatusForbidden, "Read permission required for query validation")
		return
	}

	var req ValidateSQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Query == "" {
		response.Error(w, http.StatusBadRequest, "Query is required")
		return
	}

	// Validate query via proxy
	resp, err := h.proxyClient.ValidateSQL(req.Query)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to validate query: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// GetDBInfo gets database connection information
func (h *DatabaseHandler) GetDBInfo(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	projectID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Verify project ownership
	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(w, http.StatusNotFound, "Project not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to fetch project")
		return
	}

	// Get database info via proxy
	resp, err := h.proxyClient.GetDBInfo()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get database info: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, resp)
}
