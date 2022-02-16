package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	ID          uuid.UUID `gorm:"primaryKey; type:uuid; default:uuid_generate_v4()"`
	CreatedAt   time.Time `gorm:"autoCreateTime; index; not null"`
	CreatedBy   *User
	UpdatedBy   *User
	DeletedBy   *User
	CreatedByID *uuid.UUID     `gorm:"type:uuid"`
	UpdatedByID *uuid.UUID     `gorm:"type:uuid"`
	DeletedByID *uuid.UUID     `gorm:"type:uuid"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime; not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type Team struct {
	Model
	Slug     string `gorm:"unique; not null"`
	Name     string `gorm:"unique; not null"`
	Purpose  string
	Metadata []*TeamMetadata
	Users    []*User `gorm:"many2many:users_teams"`
	Roles    []*Role `gorm:"many2many:teams_roles"`
}

type User struct {
	Model
	Email  string  `gorm:"unique"`
	APIKey string  `gorm:"unique"`
	Name   string  `gorm:"not null"`
	Teams  []*Team `gorm:"many2many:users_teams"`
	Roles  []*Role `gorm:"many2many:users_roles"`
}

type TeamMetadata struct {
	Model
	Team   *Team
	TeamID *uuid.UUID `gorm:"uniqueIndex:team_key"`
	Key    string     `gorm:"uniqueIndex:team_key; not null"`
	Value  string
}

type System struct {
	Model
	Name string `gorm:"uniqueIndex"`
}

type Role struct {
	Model
	System      *System
	SystemID    *uuid.UUID
	Resource    string  `gorm:"not null"` // sub-resource at system (maybe not needed if systems are namespaced, e.g. gcp:buckets)
	AccessLevel string  `gorm:"not null"` // read, write, R/W, other combinations per system
	Permission  string  `gorm:"not null"` // allow/deny
	Users       []*User `gorm:"many2many:users_roles"`
	Teams       []*Team `gorm:"many2many:teams_roles"`
}

type Synchronization struct {
	Model
}

type AuditLog struct {
	Model
	System            *System          `gorm:"not null"`
	Synchronization   *Synchronization `gorm:"not null"`
	User              *User
	Team              *Team
	SystemID          *uuid.UUID
	SynchronizationID *uuid.UUID
	UserID            *uuid.UUID
	TeamID            *uuid.UUID
	Action            string `gorm:"not null; index"` // CRUD action
	Status            int    `gorm:"not null; index"` // Exit code of operation
	Message           string `gorm:"not null"`        // User readable success or error message (log line)
}
