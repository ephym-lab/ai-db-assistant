package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/ephy-lab/ai-db-assistant/internal/middleware"
	"github.com/ephy-lab/ai-db-assistant/internal/models"
	"github.com/ephy-lab/ai-db-assistant/pkg/response"
	"gorm.io/gorm"
)

type DashboardHandler struct {
	db *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

type ProjectSummary struct {
	ProjectID     uint           `json:"project_id"`
	ProjectName   string         `json:"project_name"`
	DatabaseType  string         `json:"database_type"`
	TableCount    int            `json:"table_count"`
	TotalQueries  int64          `json:"total_queries"`
	RecentQueries []models.Query `json:"recent_queries"`
}

type UserDashboard struct {
	TotalProjects int64 `json:"total_projects"`
	TotalQueries  int64 `json:"total_queries"`
	TotalMessages int64 `json:"total_messages"`
}

func (h *DashboardHandler) GetProjectSummary(w http.ResponseWriter, r *http.Request) {
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

	// Count total queries
	var totalQueries int64
	h.db.Model(&models.Query{}).Where("project_id = ?", projectID).Count(&totalQueries)

	// Get recent queries
	var recentQueries []models.Query
	h.db.Where("project_id = ?", projectID).Order("created_at desc").Limit(10).Find(&recentQueries)

	// Get table count from connected database
	tableCount := h.getTableCount(project.DatabaseType, project.ConnectionString)

	summary := ProjectSummary{
		ProjectID:     project.ID,
		ProjectName:   project.Name,
		DatabaseType:  project.DatabaseType,
		TableCount:    tableCount,
		TotalQueries:  totalQueries,
		RecentQueries: recentQueries,
	}

	response.JSON(w, http.StatusOK, summary)
}

func (h *DashboardHandler) GetUserDashboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var dashboard UserDashboard

	// Count total projects
	h.db.Model(&models.Project{}).Where("user_id = ?", userID).Count(&dashboard.TotalProjects)

	// Count total queries across all projects
	h.db.Model(&models.Query{}).
		Joins("JOIN projects ON queries.project_id = projects.id").
		Where("projects.user_id = ?", userID).
		Count(&dashboard.TotalQueries)

	// Count total messages across all projects
	h.db.Model(&models.Message{}).
		Joins("JOIN projects ON messages.project_id = projects.id").
		Where("projects.user_id = ?", userID).
		Count(&dashboard.TotalMessages)

	response.JSON(w, http.StatusOK, dashboard)
}

func (h *DashboardHandler) getTableCount(dbType, connString string) int {
	var db *sql.DB
	var err error

	switch dbType {
	case "postgresql":
		db, err = sql.Open("postgres", connString)
	case "mysql":
		db, err = sql.Open("mysql", connString)
	default:
		return 0
	}

	if err != nil {
		return 0
	}
	defer db.Close()

	var count int
	var query string

	if dbType == "postgresql" {
		query = "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'"
	} else {
		query = "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE'"
	}

	err = db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0
	}

	return count
}