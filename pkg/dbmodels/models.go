package dbmodels

import (
	"github.com/jackc/pgtype"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"

	"github.com/google/uuid"
)

// Model Base model that all database tables inherit.
type Model struct {
	ID          *uuid.UUID `gorm:"type:uuid; primaryKey"`
	CreatedAt   time.Time  `gorm:"<-:create; autoCreateTime; index; not null"`
	CreatedBy   *User      `gorm:""`
	UpdatedBy   *User      `gorm:""`
	CreatedByID *uuid.UUID `gorm:"type:uuid"`
	UpdatedByID *uuid.UUID `gorm:"type:uuid"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime; not null"`
}

type ApiKey struct {
	Model
	APIKey string    `gorm:"unique; not null"`
	User   User      `gorm:""`
	UserID uuid.UUID `gorm:"type:uuid; not null"`
}

type AuditLog struct {
	Model
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

type Authorization struct {
	Model
	Name string `gorm:"unique; not null"`
}

type Correlation struct {
	Model
}

type ReconcileError struct {
	Model
	Correlation   Correlation `gorm:""`
	System        System      `gorm:""`
	Team          Team        `gorm:""`
	CorrelationID uuid.UUID   `gorm:"type:uuid; uniqueIndex:correlation_system_team_key; not null"`
	SystemID      uuid.UUID   `gorm:"type:uuid; uniqueIndex:correlation_system_team_key; not null"`
	TeamID        uuid.UUID   `gorm:"type:uuid; uniqueIndex:correlation_system_team_key; not null"`
	Message       string      `gorm:"not null"` // Human readable error message
}

type Role struct {
	Model
	Name           string          `gorm:"unique; not null"`
	Authorizations []Authorization `gorm:"many2many:role_authorizations"`
}

type RoleAuthorization struct {
	Authorization   Authorization `gorm:""`
	Role            Role          `gorm:""`
	AuthorizationID uuid.UUID     `gorm:"type:uuid; primaryKey"`
	RoleID          uuid.UUID     `gorm:"type:uuid; primaryKey"`
}

type SystemState struct {
	Model
	System   System       `gorm:""`
	Team     Team         `gorm:""`
	SystemID uuid.UUID    `gorm:"type:uuid; uniqueIndex:system_team_key; not null; index"`
	TeamID   uuid.UUID    `gorm:"type:uuid; uniqueIndex:system_team_key; not null; index"`
	State    pgtype.JSONB `gorm:"type:jsonb;default:'{}';not null"`
}

type System struct {
	Model
	Name string `gorm:"uniqueIndex; not null"`
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
	Slug      Slug            `gorm:"<-:create; unique; not null"`
	Name      string          `gorm:"unique; not null"`
	Purpose   *string         `gorm:""`
	Metadata  []*TeamMetadata `gorm:""`
	Users     []*User         `gorm:"many2many:user_teams"`
	AuditLogs []*AuditLog     `gorm:"foreignKey:TargetTeamID"`
}

type User struct {
	Model
	Email        string     `gorm:"not null; unique"`
	Name         string     `gorm:"not null"`
	Teams        []*Team    `gorm:"many2many:user_teams"`
	RoleBindings []UserRole `gorm:""`
}

type UserRole struct {
	Role     Role       `gorm:""`
	User     User       `gorm:""`
	RoleID   uuid.UUID  `gorm:"type:uuid; primaryKey; index:user_role_target,unique"`
	UserID   uuid.UUID  `gorm:"type:uuid; primaryKey; index:user_role_target,unique"`
	TargetID *uuid.UUID `gorm:"type:uuid; index:user_role_target,unique"`
}

type UserTeam struct {
	Model
	Team   Team      `gorm:""`
	User   User      `gorm:""`
	UserID uuid.UUID `gorm:"type:uuid; not null; index:user_teams_index,unique"`
	TeamID uuid.UUID `gorm:"type:uuid; not null; index:user_teams_index,unique"`
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

func (m *Model) BeforeCreate(tx *gorm.DB) error {
	if m.ID != nil {
		return nil
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	m.ID = &id
	return nil
}
