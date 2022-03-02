package dbmodels

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	ID          *uuid.UUID     `json:"id" binding:"-" gorm:"primaryKey; type:uuid; default:uuid_generate_v4()"`
	CreatedAt   time.Time      `json:"created_at" binding:"-" gorm:"<-:create; autoCreateTime; index; not null"`
	CreatedBy   *User          `json:"-" binding:"-"`
	UpdatedBy   *User          `json:"-" binding:"-"`
	DeletedBy   *User          `json:"-" binding:"-"`
	CreatedByID *uuid.UUID     `json:"created_by_id" binding:"-" gorm:"type:uuid"`
	UpdatedByID *uuid.UUID     `json:"updated_by_id" binding:"-" gorm:"type:uuid"`
	DeletedByID *uuid.UUID     `json:"deleted_by_id" binding:"-" gorm:"type:uuid"`
	UpdatedAt   time.Time      `json:"updated_at" binding:"-" gorm:"autoUpdateTime; not null"`
	DeletedAt   gorm.DeletedAt `json:"-" binding:"-" gorm:"index"`
}

type Team struct {
	Model
	Slug     *string         `json:"slug" gorm:"<-:create; unique; not null"`
	Name     *string         `json:"name" gorm:"unique; not null"`
	Purpose  *string         `json:"purpose"`
	Metadata []*TeamMetadata `json:"-" binding:"-"`
	Users    []*User         `json:"-" binding:"-" gorm:"many2many:users_teams"`
	Roles    []*Role         `json:"-" binding:"-" gorm:"many2many:teams_roles"`
}

type User struct {
	Model
	Email *string `json:"email" gorm:"unique"`
	Name  *string `json:"name" gorm:"not null" example:"plain english"`
	Teams []*Team `json:"-" binding:"-" gorm:"many2many:users_teams"`
	Roles []*Role `json:"-" binding:"-" gorm:"many2many:users_roles"`
}

type ApiKey struct {
	Model
	APIKey string    `json:"apikey" binding:"-" gorm:"unique; not null"`
	User   *User     `json:"-" binding:"-"`
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid; not null"`
}

type TeamMetadata struct {
	Model
	Team   *Team
	TeamID *uuid.UUID `gorm:"uniqueIndex:team_key"`
	Key    string     `gorm:"uniqueIndex:team_key; not null"`
	Value  *string
}

type System struct {
	Model
	Name string `gorm:"uniqueIndex; not null"`
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
