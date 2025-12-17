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
	Permission       *Permission    `json:"permission,omitempty"`
	Queries          []Query        `json:"queries,omitempty"`
	Messages         []Message      `json:"messages,omitempty"`
}

type Permission struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	ProjectID   uint           `gorm:"uniqueIndex;not null" json:"project_id"`
	AllowDDL    bool           `gorm:"not null" json:"allow_ddl"`
	AllowWrite  bool           `gorm:"not null" json:"allow_write"`
	AllowRead   bool           `gorm:"not null" json:"allow_read"`
	AllowDelete bool           `gorm:"not null" json:"allow_delete"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Project     Project        `json:"project,omitempty"`
}

type Query struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	ProjectID     uint           `gorm:"not null;index" json:"project_id"`
	Query         string         `gorm:"type:text;not null" json:"query"`
	QueryType     string         `json:"query_type,omitempty"`
	Status        string         `gorm:"not null" json:"status"` // "success", "error", "generated", "pending"
	Result        string         `gorm:"type:text" json:"result"`
	Error         string         `gorm:"type:text" json:"error,omitempty"`
	RowsAffected  int            `json:"rows_affected,omitempty"`
	ExecutionTime int            `json:"execution_time,omitempty"` // in milliseconds
	CreatedAt     time.Time      `json:"created_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	Project       Project        `json:"project,omitempty"`
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