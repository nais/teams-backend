package dbmodels

import (
	log "github.com/sirupsen/logrus"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Model Base model that all database tables inherit.
type Model struct {
	ID          *uuid.UUID `json:"id" gorm:"primaryKey; type:uuid; default:(uuid_generate_v4())"`
	CreatedAt   time.Time  `json:"created_at" gorm:"<-:create; autoCreateTime; index; not null"`
	CreatedBy   *User      `json:"-"`
	UpdatedBy   *User      `json:"-"`
	CreatedByID *uuid.UUID `json:"created_by_id" gorm:"type:uuid"`
	UpdatedByID *uuid.UUID `json:"updated_by_id" gorm:"type:uuid"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime; not null"`
}

// SoftDeletes Add fields to support Gorm soft deletes
type SoftDeletes struct {
	DeletedBy   *User          `json:"-"`
	DeletedByID *uuid.UUID     `json:"deleted_by_id" gorm:"type:uuid"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type ApiKey struct {
	Model
	SoftDeletes
	APIKey string    `json:"apikey" gorm:"unique; not null"`
	User   *User     `json:"-"`
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid; not null"`
}

type AuditLog struct {
	Model
	SoftDeletes
	System            *System
	Synchronization   *Synchronization
	User              *User // User object, not subject, i.e. which user was affected by the operation.
	Team              *Team
	SystemID          *uuid.UUID
	SynchronizationID *uuid.UUID
	UserID            *uuid.UUID
	TeamID            *uuid.UUID
	Success           bool   `gorm:"not null; index"` // True if operation succeeded
	Action            string `gorm:"not null; index"` // CRUD action
	Message           string `gorm:"not null"`        // User readable success or error message (log line)
}

type RoleBinding struct {
	Model
	Role   *Role      `gorm:"not null"` // which role is granted
	Team   *Team      `gorm:""`         // role is granted in context of this team
	User   *User      `gorm:"not null"` // which user is granted this role
	UserID *uuid.UUID `gorm:"not null; index:user_role_team_index,unique"`
	RoleID *uuid.UUID `gorm:"not null; index:user_role_team_index,unique"`
	TeamID *uuid.UUID `gorm:"index:user_role_team_index,unique"`
}

type Role struct {
	Model
	SoftDeletes
	System       *System        `gorm:"not null"`
	SystemID     *uuid.UUID     `gorm:"not null; index"`
	Name         string         `gorm:"uniqueIndex; not null"`
	Description  string         `gorm:""`
	Resource     string         `gorm:"not null"` // sub-resource at system (maybe not needed if systems are namespaced, e.g. gcp:buckets)
	AccessLevel  string         `gorm:"not null"` // CRUD
	Permission   string         `gorm:"not null"` // allow/deny
	RoleBindings []*RoleBinding `gorm:"foreignKey:RoleID"`
}

type Synchronization struct {
	Model
	SoftDeletes
}

type SystemState struct {
	Model
	Team        *Team
	TeamID      *uuid.UUID `gorm:"uniqueIndex:system_env_team_key; index"`
	System      *System
	SystemID    *uuid.UUID `gorm:"uniqueIndex:system_env_team_key; not null; index"`
	Environment *string    `gorm:"uniqueIndex:system_env_team_key"`
	Key         string     `gorm:"uniqueIndex:system_env_team_key; not null"`
	Value       string     `gorm:"not null"`
}

type System struct {
	Model
	SoftDeletes
	Name string `gorm:"uniqueIndex; not null"`
}

type SystemsTeams struct {
	Model
	System   *System    `json:"-" gorm:"not null"`
	Team     *Team      `json:"-" gorm:"not null"`
	SystemID *uuid.UUID `json:"system_id" gorm:"not null; index:systems_teams_index,unique"`
	TeamID   *uuid.UUID `json:"team_id" gorm:"not null; index:systems_teams_index,unique"`
}

type TeamMetadata struct {
	Model
	Team   *Team
	TeamID *uuid.UUID `gorm:"uniqueIndex:team_key"`
	Key    string     `gorm:"uniqueIndex:team_key; not null"`
	Value  *string
}

type Team struct {
	Model
	SoftDeletes
	Slug        *Slug           `json:"slug" gorm:"<-:create; unique; not null"`
	Name        *string         `json:"name" gorm:"unique; not null"`
	Purpose     *string         `json:"purpose"`
	Metadata    []*TeamMetadata `json:"-"`
	SystemState []*SystemState  `json:"-"`
	Users       []*User         `json:"-" gorm:"many2many:users_teams"`
	Systems     []*System       `json:"-" gorm:"many2many:systems_teams"`
	AuditLogs   []*AuditLog     `json:"-" gorm:"foreignKey:TeamID"`
}

type User struct {
	Model
	SoftDeletes
	Email        *string        `json:"email" gorm:"unique"`
	Name         *string        `json:"name" gorm:"not null"`
	Teams        []*Team        `json:"-" gorm:"many2many:users_teams"`
	RoleBindings []*RoleBinding `json:"-" gorm:"foreignKey:UserID"`
}

type UsersTeams struct {
	Model
	Team   *Team      `json:"-" gorm:"not null"`
	User   *User      `json:"-" gorm:"not null"`
	UserID *uuid.UUID `json:"user_id" gorm:"not null; index:users_teams_index,unique"`
	TeamID *uuid.UUID `json:"team_id" gorm:"not null; index:users_teams_index,unique"`
}

// GetModel Enable callers to access the base model through an interface.
// This means that setting common metadata like 'created by' or 'updated by' can be abstracted.
func (m *Model) GetModel() *Model {
	return m
}

// Error Get the err message from the audit log
func (a *AuditLog) Error() string {
	return a.Message
}

// Log Add an entry in the standard logger
func (a *AuditLog) Log() *log.Entry {
	return log.StandardLogger().WithFields(log.Fields{
		"system":         a.System.Name,
		"correlation_id": a.SynchronizationID,
		"user":           a.UserID,
		"team":           a.TeamID,
		"action":         a.Action,
		"success":        a.Success,
	})
}
