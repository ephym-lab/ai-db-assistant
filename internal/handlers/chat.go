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

type ChatHandler struct {
	db *gorm.DB
}

func NewChatHandler(db *gorm.DB) *ChatHandler {
	return &ChatHandler{db: db}
}

type SendMessageRequest struct {
	Content string `json:"content"`
}

func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	projectID, err := strconv.ParseUint(vars["project_id"], 10, 32)
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

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Content == "" {
		response.Error(w, http.StatusBadRequest, "Message content is required")
		return
	}

	// Save user message
	userMessage := models.Message{
		ProjectID: uint(projectID),
		Role:      "user",
		Content:   req.Content,
	}

	if err := h.db.Create(&userMessage).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to save message")
		return
	}

	// TODO: Here you would integrate with OpenAI API to get AI response
	// For now, we'll return a placeholder response
	aiResponse := "This is a placeholder response. OpenAI integration will be added in the future."

	// Save AI message
	aiMessage := models.Message{
		ProjectID: uint(projectID),
		Role:      "assistant",
		Content:   aiResponse,
	}

	if err := h.db.Create(&aiMessage).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to save AI response")
		return
	}

	response.Success(w, http.StatusOK, "Message sent successfully", map[string]interface{}{
		"user_message": userMessage,
		"ai_message":   aiMessage,
	})
}

func (h *ChatHandler) GetChatHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	projectID, err := strconv.ParseUint(vars["project_id"], 10, 32)
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

	// Get chat history
	var messages []models.Message
	if err := h.db.Where("project_id = ?", projectID).Order("created_at asc").Find(&messages).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch chat history")
		return
	}

	response.JSON(w, http.StatusOK, messages)
}