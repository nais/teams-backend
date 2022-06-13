package dbmodels

import (
	log "github.com/sirupsen/logrus"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Model Base model that all database tables inherit.
type Model struct {
	ID          *uuid.UUID `gorm:"type:uuid; primaryKey; default:(uuid_generate_v4())"`
	CreatedAt   time.Time  `gorm:"<-:create; autoCreateTime; index; not null"`
	CreatedBy   *User      `gorm:""`
	UpdatedBy   *User      `gorm:""`
	CreatedByID *uuid.UUID `gorm:"type:uuid"`
	UpdatedByID *uuid.UUID `gorm:"type:uuid"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime; not null"`
}

// SoftDeletes Add fields to support Gorm soft deletes
type SoftDeletes struct {
	DeletedBy   *User          `gorm:""`
	DeletedByID *uuid.UUID     `gorm:"type:uuid"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type ApiKey struct {
	Model
	SoftDeletes
	APIKey string    `gorm:"unique; not null"`
	User   User      `gorm:""`
	UserID uuid.UUID `gorm:"type:uuid; not null"`
}

type AuditLog struct {
	Model
	SoftDeletes
	Synchronization   Synchronization `gorm:""`
	System            System          `gorm:""`
	Team              *Team           `gorm:""`
	User              *User           `gorm:""` // User object, not subject, i.e. which user was affected by the operation.
	SynchronizationID uuid.UUID       `gorm:"type:uuid; not null"`
	SystemID          uuid.UUID       `gorm:"type:uuid; not null"`
	TeamID            *uuid.UUID      `gorm:"type:uuid"`
	UserID            *uuid.UUID      `gorm:"type:uuid"`
	Success           bool            `gorm:"not null; index"` // True if operation succeeded
	Action            string          `gorm:"not null; index"` // CRUD action
	Message           string          `gorm:"not null"`        // User readable success or error message (log line)
}

type RoleBinding struct {
	Model
	Role   Role       `gorm:""` // which role is granted
	Team   *Team      `gorm:""` // role is granted in context of this team
	User   User       `gorm:""` // which user is granted this role
	RoleID uuid.UUID  `gorm:"type:uuid; not null; index:user_role_team_index,unique"`
	TeamID *uuid.UUID `gorm:"type:uuid; index:user_role_team_index,unique"`
	UserID uuid.UUID  `gorm:"type:uuid; not null; index:user_role_team_index,unique"`
}

type Role struct {
	Model
	SoftDeletes
	System       System         `gorm:"not null"`
	SystemID     uuid.UUID      `gorm:"type:uuid; not null; index"`
	Name         string         `gorm:"uniqueIndex; not null"`
	Description  string         `gorm:"not null"` // Human readable role description
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
	System      System     `gorm:""`
	Team        *Team      `gorm:""`
	SystemID    uuid.UUID  `gorm:"type:uuid; uniqueIndex:system_env_team_key; not null; index"`
	TeamID      *uuid.UUID `gorm:"type:uuid; uniqueIndex:system_env_team_key; index"`
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
	System   System    `gorm:"not null"`
	Team     Team      `gorm:"not null"`
	SystemID uuid.UUID `gorm:"type:uuid; not null; index:systems_teams_index,unique"`
	TeamID   uuid.UUID `gorm:"type:uuid; not null; index:systems_teams_index,unique"`
}

type TeamMetadata struct {
	Model
	Team   Team      `gorm:""`
	TeamID uuid.UUID `gorm:"type:uuid; not null; uniqueIndex:team_key"`
	Key    string    `gorm:"uniqueIndex:team_key; not null"`
	Value  *string   `gorm:""`
}

type Team struct {
	Model
	SoftDeletes
	Slug        Slug            `gorm:"<-:create; unique; not null"`
	Name        string          `gorm:"unique; not null"`
	Purpose     *string         `gorm:""`
	Metadata    []*TeamMetadata `gorm:""`
	SystemState []*SystemState  `gorm:""`
	Users       []*User         `gorm:"many2many:users_teams"`
	Systems     []*System       `gorm:"many2many:systems_teams"`
	AuditLogs   []*AuditLog     `gorm:"foreignKey:TeamID"`
}

type User struct {
	Model
	SoftDeletes
	Email        *string        `gorm:"unique"`
	Name         *string        `gorm:"not null"`
	Teams        []*Team        `gorm:"many2many:users_teams"`
	RoleBindings []*RoleBinding `gorm:"foreignKey:UserID"`
}

type UsersTeams struct {
	Model
	Team   Team      `gorm:""`
	User   User      `gorm:""`
	UserID uuid.UUID `gorm:"type:uuid; not null; index:users_teams_index,unique"`
	TeamID uuid.UUID `gorm:"type:uuid; not null; index:users_teams_index,unique"`
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
	var teamSlug string
	var teamName string
	var userEmail string
	var userName string

	if a.Team != nil {
		teamName = a.Team.Name
		teamSlug = string(a.Team.Slug)
	}

	if a.User != nil {
		if a.User.Email != nil {
			userEmail = *a.User.Email
		}

		userName = *a.User.Name
	}

	return log.StandardLogger().WithFields(log.Fields{
		"system_id":      a.SystemID,
		"system_name":    a.System.Name,
		"correlation_id": a.SynchronizationID,
		"team_id":        uuidAsString(a.TeamID),
		"team_name":      teamName,
		"team_slug":      teamSlug,
		"user_email":     userEmail,
		"user_id":        uuidAsString(a.UserID),
		"user_name":      userName,
		"action":         a.Action,
		"success":        a.Success,
	})
}

func uuidAsString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}

	return id.String()
}
