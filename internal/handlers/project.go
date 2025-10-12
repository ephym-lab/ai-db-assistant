// internal/handlers/project.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/ephy-lab/ai-db-assistant/internal/middleware"
	"github.com/ephy-lab/ai-db-assistant/internal/models"
	"github.com/ephy-lab/ai-db-assistant/pkg/response"
	"gorm.io/gorm"
)

type ProjectHandler struct {
	db *gorm.DB
}

func NewProjectHandler(db *gorm.DB) *ProjectHandler {
	return &ProjectHandler{db: db}
}

type CreateProjectRequest struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	DatabaseType     string `json:"database_type"`
	ConnectionString string `json:"connection_string"`
}

type UpdateProjectRequest struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	ConnectionString string `json:"connection_string"`
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" || req.DatabaseType == "" || req.ConnectionString == "" {
		response.Error(w, http.StatusBadRequest, "Name, database type, and connection string are required")
		return
	}

	if req.DatabaseType != "mysql" && req.DatabaseType != "postgresql" {
		response.Error(w, http.StatusBadRequest, "Database type must be 'mysql' or 'postgresql'")
		return
	}

	project := models.Project{
		UserID:           userID,
		Name:             req.Name,
		Description:      req.Description,
		DatabaseType:     req.DatabaseType,
		ConnectionString: req.ConnectionString,
	}

	if err := h.db.Create(&project).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create project")
		return
	}

	response.Success(w, http.StatusCreated, "Project created successfully", project)
}

func (h *ProjectHandler) GetProjects(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var projects []models.Project
	// Preload User data for the list view
	if err := h.db.Preload("User").Where("user_id = ?", userID).Order("created_at desc").Find(&projects).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch projects")
		return
	}

	response.JSON(w, http.StatusOK, projects)
}

func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
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

	var project models.Project
	// Preload User data for single project view
	if err := h.db.Preload("User").Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(w, http.StatusNotFound, "Project not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to fetch project")
		return
	}

	response.JSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(w, http.StatusNotFound, "Project not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to fetch project")
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.ConnectionString != "" {
		updates["connection_string"] = req.ConnectionString
	}

	if err := h.db.Model(&project).Updates(updates).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update project")
		return
	}

	response.Success(w, http.StatusOK, "Project updated successfully", project)
}

func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
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

	result := h.db.Where("id = ? AND user_id = ?", projectID, userID).Delete(&models.Project{})
	if result.Error != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to delete project")
		return
	}

	if result.RowsAffected == 0 {
		response.Error(w, http.StatusNotFound, "Project not found")
		return
	}

	response.Success(w, http.StatusOK, "Project deleted successfully", nil)
}