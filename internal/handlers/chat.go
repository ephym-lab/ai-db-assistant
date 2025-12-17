// internal/handlers/chat.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/ephy-lab/ai-db-assistant/internal/config"
	"github.com/ephy-lab/ai-db-assistant/internal/middleware"
	"github.com/ephy-lab/ai-db-assistant/internal/models"
	"github.com/ephy-lab/ai-db-assistant/pkg/proxyclient"
	"github.com/ephy-lab/ai-db-assistant/pkg/response"
	"gorm.io/gorm"
)

type ChatHandler struct {
	db          *gorm.DB
	proxyClient *proxyclient.Client
}

func NewChatHandler(db *gorm.DB, cfg *config.Config) *ChatHandler {
	return &ChatHandler{
		db:          db,
		proxyClient: proxyclient.NewClient(cfg.ProxyServerURL),
	}
}

type SendMessageRequest struct {
	Content string `json:"content"`
}

type ChatMessageResponse struct {
	UserMessage models.Message `json:"user_message"`
	AIMessage   models.Message `json:"ai_message"`
	AIResponse  AIResponseData `json:"ai_response"`
}

type AIResponseData struct {
	Content string  `json:"content"`
	Query   *string `json:"query,omitempty"`
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

	// Verify project ownership and load with permission
	var project models.Project
	if err := h.db.Preload("Permission").Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
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

	// Generate AI response using proxy client
	// TODO: Optionally fetch and include database schema
	proxyResp, err := h.proxyClient.GenerateSQL(req.Content, project.DatabaseType, "")
	if err != nil {
		// Save error message
		aiMessage := models.Message{
			ProjectID: uint(projectID),
			Role:      "assistant",
			Content:   "I'm sorry, I encountered an error while processing your request: " + err.Error(),
		}
		h.db.Create(&aiMessage)

		response.Error(w, http.StatusInternalServerError, "Failed to generate SQL: "+err.Error())
		return
	}

	// Prepare AI response data
	aiResp := AIResponseData{
		Content: proxyResp.Content,
	}
	if proxyResp.Query != "" {
		aiResp.Query = &proxyResp.Query
	}

	// Store the full AI response as JSON in the message content
	aiContentJSON, _ := json.Marshal(aiResp)

	// Save AI message
	aiMessage := models.Message{
		ProjectID: uint(projectID),
		Role:      "assistant",
		Content:   string(aiContentJSON),
	}

	if err := h.db.Create(&aiMessage).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to save AI response")
		return
	}

	// If query was generated, log it to queries table
	if aiResp.Query != nil {
		query := models.Query{
			ProjectID: uint(projectID),
			Query:     *aiResp.Query,
			Status:    "generated",
			Result:    "Query generated but not executed yet",
		}
		h.db.Create(&query)
	}

	response.Success(w, http.StatusOK, "Message sent successfully", ChatMessageResponse{
		UserMessage: userMessage,
		AIMessage:   aiMessage,
		AIResponse:  aiResp,
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

	// Parse AI responses to include structured data
	type MessageWithParsedContent struct {
		ID         uint            `json:"id"`
		ProjectID  uint            `json:"project_id"`
		Role       string          `json:"role"`
		Content    string          `json:"content,omitempty"`
		AIResponse *AIResponseData `json:"ai_response,omitempty"`
		CreatedAt  string          `json:"created_at"`
	}

	var parsedMessages []MessageWithParsedContent
	for _, msg := range messages {
		parsed := MessageWithParsedContent{
			ID:        msg.ID,
			ProjectID: msg.ProjectID,
			Role:      msg.Role,
			CreatedAt: msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if msg.Role == "user" {
			parsed.Content = msg.Content
		} else {
			// Try to parse AI response JSON
			var aiResp AIResponseData
			if err := json.Unmarshal([]byte(msg.Content), &aiResp); err == nil {
				parsed.AIResponse = &aiResp
			} else {
				// Fallback for old format or plain text
				parsed.Content = msg.Content
			}
		}

		parsedMessages = append(parsedMessages, parsed)
	}

	response.JSON(w, http.StatusOK, parsedMessages)
}