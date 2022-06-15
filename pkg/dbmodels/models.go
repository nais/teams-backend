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
	Actor          *User       `gorm:""` // The user or service account that performed the action
	Correlation    Correlation `gorm:""`
	TargetSystem   System      `gorm:""`
	TargetTeam     *Team       `gorm:""` // The team, if any, that was the target of the action
	TargetUser     *User       `gorm:""` // The user, if any, that was the target of the action
	ActorID        *uuid.UUID  `gorm:"type:uuid"`
	CorrelationID  uuid.UUID   `gorm:"type:uuid; not null"`
	TargetSystemID uuid.UUID   `gorm:"type:uuid; not null"`
	TargetTeamID   *uuid.UUID  `gorm:"type:uuid"`
	TargetUserID   *uuid.UUID  `gorm:"type:uuid"`
	Action         string      `gorm:"not null; index"`
	Message        string      `gorm:"not null"` // Human readable message (log line)
}

type Correlation struct {
	Model
	SoftDeletes
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
	AuditLogs   []*AuditLog     `gorm:"foreignKey:TargetTeamID"`
}

type User struct {
	Model
	SoftDeletes
	Email        string         `gorm:"not null; unique"`
	Name         string         `gorm:"not null"`
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
	var actorEmail string
	var actorName string
	var targetTeamName string
	var targetTeamSlug string
	var targetUserName string

	if a.Actor != nil {
		actorEmail = a.Actor.Email
		actorName = a.Actor.Name
	}

	if a.TargetTeam != nil {
		targetTeamName = a.TargetTeam.Name
		targetTeamSlug = string(a.TargetTeam.Slug)
	}

	if a.TargetUser != nil {
		targetUserName = a.TargetUser.Name
	}

	return log.StandardLogger().WithFields(log.Fields{
		"action":             a.Action,
		"actor_email":        actorEmail,
		"actor_id":           uuidAsString(a.ActorID),
		"actor_name":         actorName,
		"correlation_id":     a.CorrelationID,
		"target_system_id":   a.TargetSystemID,
		"target_system_name": a.TargetSystem.Name,
		"target_team_id":     uuidAsString(a.TargetTeamID),
		"target_team_name":   targetTeamName,
		"target_team_slug":   targetTeamSlug,
		"target_user_id":     uuidAsString(a.TargetUserID),
		"target_user_name":   targetUserName,
	})
}

func uuidAsString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}

	return id.String()
}
