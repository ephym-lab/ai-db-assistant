package models

import (
	"time"
	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"unique;not null" json:"username"`
	Email     string         `gorm:"unique;not null" json:"email"`
	Password  string         `gorm:"not null" json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Projects  []Project      `json:"projects,omitempty"`
}

type Project struct {
	ID               uint           `gorm:"primarykey" json:"id"`
	UserID           uint           `gorm:"not null;index" json:"user_id"`
	Name             string         `gorm:"not null" json:"name"`
	Description      string         `json:"description"`
	DatabaseType     string         `gorm:"not null" json:"database_type"` // "mysql" or "postgresql"
	ConnectionString string         `gorm:"not null" json:"connection_string"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	User             *User          `json:"user,omitempty"`
	Queries          []Query        `json:"queries,omitempty"`
	Messages         []Message      `json:"messages,omitempty"`
}

type Query struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	ProjectID uint           `gorm:"not null;index" json:"project_id"`
	Query     string         `gorm:"type:text;not null" json:"query"`
	Status    string         `gorm:"not null" json:"status"` // "success" or "error"
	Result    string         `gorm:"type:text" json:"result"`
	Error     string         `gorm:"type:text" json:"error,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Project   Project        `json:"project,omitempty"`
}

type Message struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	ProjectID uint           `gorm:"not null;index" json:"project_id"`
	Role      string         `gorm:"not null" json:"role"` // "user" or "assistant"
	Content   string         `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Project   Project        `json:"project,omitempty"`
}