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
	AllowDDL         *bool  `json:"allow_ddl,omitempty"`
	AllowWrite       *bool  `json:"allow_write,omitempty"`
	AllowRead        *bool  `json:"allow_read,omitempty"`
	AllowDelete      *bool  `json:"allow_delete,omitempty"`
}

type UpdateProjectRequest struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	ConnectionString string `json:"connection_string"`
	AllowDDL         *bool  `json:"allow_ddl,omitempty"`
	AllowWrite       *bool  `json:"allow_write,omitempty"`
	AllowRead        *bool  `json:"allow_read,omitempty"`
	AllowDelete      *bool  `json:"allow_delete,omitempty"`
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

	// Create permissions with defaults
	allowDDL := true
	allowWrite := true
	allowRead := true
	allowDelete := true

	if req.AllowDDL != nil {
		allowDDL = *req.AllowDDL
	}
	if req.AllowWrite != nil {
		allowWrite = *req.AllowWrite
	}
	if req.AllowRead != nil {
		allowRead = *req.AllowRead
	}
	if req.AllowDelete != nil {
		allowDelete = *req.AllowDelete
	}

	permission := models.Permission{
		ProjectID:   project.ID,
		AllowDDL:    allowDDL,
		AllowWrite:  allowWrite,
		AllowRead:   allowRead,
		AllowDelete: allowDelete,
	}

	if err := h.db.Create(&permission).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create project permissions")
		return
	}

	// Load the permission into the project response
	project.Permission = &permission

	response.Success(w, http.StatusCreated, "Project created successfully", project)
}

func (h *ProjectHandler) GetProjects(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var projects []models.Project
	// Preload User and Permission data for the list view
	if err := h.db.Preload("User").Preload("Permission").Where("user_id = ?", userID).Order("created_at desc").Find(&projects).Error; err != nil {
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
	// Preload User and Permission data for single project view
	if err := h.db.Preload("User").Preload("Permission").Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
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

	// Update permissions if provided
	permissionUpdates := map[string]interface{}{}
	if req.AllowDDL != nil {
		permissionUpdates["allow_ddl"] = *req.AllowDDL
	}
	if req.AllowWrite != nil {
		permissionUpdates["allow_write"] = *req.AllowWrite
	}
	if req.AllowRead != nil {
		permissionUpdates["allow_read"] = *req.AllowRead
	}
	if req.AllowDelete != nil {
		permissionUpdates["allow_delete"] = *req.AllowDelete
	}

	if len(permissionUpdates) > 0 {
		if err := h.db.Model(&models.Permission{}).Where("project_id = ?", projectID).Updates(permissionUpdates).Error; err != nil {
			response.Error(w, http.StatusInternalServerError, "Failed to update project permissions")
			return
		}
	}

	// Reload project with updated permission
	if err := h.db.Preload("Permission").Where("id = ?", projectID).First(&project).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch updated project")
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

func (h *ProjectHandler) GetProjectPermissions(w http.ResponseWriter, r *http.Request) {
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

	// First verify the project belongs to the user
	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(w, http.StatusNotFound, "Project not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to fetch project")
		return
	}

	// Get the permissions for the project
	var permission models.Permission
	if err := h.db.Where("project_id = ?", projectID).First(&permission).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(w, http.StatusNotFound, "Permissions not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to fetch permissions")
		return
	}

	response.JSON(w, http.StatusOK, permission)
}