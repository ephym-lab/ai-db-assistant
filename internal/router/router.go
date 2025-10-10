package router

import (
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
	chatHandler := handlers.NewChatHandler(db)
	dashboardHandler := handlers.NewDashboardHandler(db)

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// Auth routes (public)
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/signup", authHandler.Signup).Methods("POST")
	auth.HandleFunc("/login", authHandler.Login).Methods("POST")

	// Protected routes
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.Auth(cfg))

	// Project routes
	protected.HandleFunc("/projects", projectHandler.CreateProject).Methods("POST")
	protected.HandleFunc("/projects", projectHandler.GetProjects).Methods("GET")
	protected.HandleFunc("/projects/{id}", projectHandler.GetProject).Methods("GET")
	protected.HandleFunc("/projects/{id}", projectHandler.UpdateProject).Methods("PUT")
	protected.HandleFunc("/projects/{id}", projectHandler.DeleteProject).Methods("DELETE")

	// Dashboard routes
	protected.HandleFunc("/projects/{id}/summary", dashboardHandler.GetProjectSummary).Methods("GET")
	protected.HandleFunc("/dashboard", dashboardHandler.GetUserDashboard).Methods("GET")

	// Chat routes
	protected.HandleFunc("/chat/{project_id}", chatHandler.SendMessage).Methods("POST")
	protected.HandleFunc("/chat/{project_id}/history", chatHandler.GetChatHistory).Methods("GET")

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	}).Methods("GET")

	return r
}