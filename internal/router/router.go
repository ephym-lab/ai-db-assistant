// internal/router/router.go
package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ephy-lab/ai-db-assistant/internal/config"
	"github.com/ephy-lab/ai-db-assistant/internal/handlers"
	"github.com/ephy-lab/ai-db-assistant/internal/middleware"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB, cfg *config.Config) *mux.Router {
	r := mux.NewRouter()

	// Apply global middleware
	r.Use(middleware.CORS)
	r.Use(middleware.Logging)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg)
	projectHandler := handlers.NewProjectHandler(db)
	chatHandler := handlers.NewChatHandler(db, cfg)
	dashboardHandler := handlers.NewDashboardHandler(db)
	databaseHandler := handlers.NewDatabaseHandler(db, cfg)

	// Health check (before API routes)
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET", "OPTIONS")

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// Auth routes (public)
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/signup", authHandler.Signup).Methods("POST", "OPTIONS")
	auth.HandleFunc("/login", authHandler.Login).Methods("POST", "OPTIONS")

	// Protected routes
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.Auth(cfg))

	// Project routes
	protected.HandleFunc("/projects", projectHandler.CreateProject).Methods("POST", "OPTIONS")
	protected.HandleFunc("/projects", projectHandler.GetProjects).Methods("GET", "OPTIONS")
	protected.HandleFunc("/projects/{id}", projectHandler.GetProject).Methods("GET", "OPTIONS")
	protected.HandleFunc("/projects/{id}", projectHandler.UpdateProject).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/projects/{id}", projectHandler.DeleteProject).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/projects/{id}/permissions", projectHandler.GetProjectPermissions).Methods("GET", "OPTIONS")


	// Dashboard routes (must come before /projects/{id}/summary to avoid conflict)
	protected.HandleFunc("/dashboard", dashboardHandler.GetUserDashboard).Methods("GET", "OPTIONS")
	protected.HandleFunc("/projects/{id}/summary", dashboardHandler.GetProjectSummary).Methods("GET", "OPTIONS")

	// Chat routes
	protected.HandleFunc("/chat/{project_id}", chatHandler.SendMessage).Methods("POST", "OPTIONS")
	protected.HandleFunc("/chat/{project_id}/history", chatHandler.GetChatHistory).Methods("GET", "OPTIONS")

	// Database operation routes
	protected.HandleFunc("/projects/{id}/connect-db", databaseHandler.ConnectDB).Methods("POST", "OPTIONS")
	protected.HandleFunc("/projects/{id}/disconnect-db", databaseHandler.DisconnectDB).Methods("POST", "OPTIONS")
	protected.HandleFunc("/projects/{id}/execute-sql", databaseHandler.ExecuteSQL).Methods("POST", "OPTIONS")
	protected.HandleFunc("/projects/{id}/validate-sql", databaseHandler.ValidateSQL).Methods("POST", "OPTIONS")
	protected.HandleFunc("/projects/{id}/db-info", databaseHandler.GetDBInfo).Methods("GET", "OPTIONS")

	return r
}