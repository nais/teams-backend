// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0

package sqlc

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
)

type AuditAction string

const (
	AuditActionConsoleApiKeyCreate         AuditAction = "console:api-key:create"
	AuditActionConsoleApiKeyDelete         AuditAction = "console:api-key:delete"
	AuditActionConsoleServiceAccountCreate AuditAction = "console:service-account:create"
	AuditActionConsoleServiceAccountDelete AuditAction = "console:service-account:delete"
	AuditActionConsoleServiceAccountUpdate AuditAction = "console:service-account:update"
	AuditActionConsoleTeamAddMember        AuditAction = "console:team:add-member"
	AuditActionConsoleTeamAddOwner         AuditAction = "console:team:add-owner"
	AuditActionConsoleTeamCreate           AuditAction = "console:team:create"
	AuditActionConsoleTeamRemoveMember     AuditAction = "console:team:remove-member"
	AuditActionConsoleTeamSetMemberRole    AuditAction = "console:team:set-member-role"
	AuditActionConsoleTeamSync             AuditAction = "console:team:sync"
	AuditActionConsoleTeamUpdate           AuditAction = "console:team:update"
)

func (e *AuditAction) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = AuditAction(s)
	case string:
		*e = AuditAction(s)
	default:
		return fmt.Errorf("unsupported scan type for AuditAction: %T", src)
	}
	return nil
}

type NullAuditAction struct {
	AuditAction AuditAction
	Valid       bool // Valid is true if String is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullAuditAction) Scan(value interface{}) error {
	if value == nil {
		ns.AuditAction, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.AuditAction.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullAuditAction) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.AuditAction, nil
}

type AuthzName string

const (
	AuthzNameAuditLogsRead         AuthzName = "audit_logs:read"
	AuthzNameServiceAccountsCreate AuthzName = "service_accounts:create"
	AuthzNameServiceAccountsDelete AuthzName = "service_accounts:delete"
	AuthzNameServiceAccountsList   AuthzName = "service_accounts:list"
	AuthzNameServiceAccountsRead   AuthzName = "service_accounts:read"
	AuthzNameServiceAccountsUpdate AuthzName = "service_accounts:update"
	AuthzNameSystemStatesDelete    AuthzName = "system_states:delete"
	AuthzNameSystemStatesRead      AuthzName = "system_states:read"
	AuthzNameSystemStatesUpdate    AuthzName = "system_states:update"
	AuthzNameTeamsCreate           AuthzName = "teams:create"
	AuthzNameTeamsDelete           AuthzName = "teams:delete"
	AuthzNameTeamsList             AuthzName = "teams:list"
	AuthzNameTeamsRead             AuthzName = "teams:read"
	AuthzNameTeamsUpdate           AuthzName = "teams:update"
	AuthzNameUsersList             AuthzName = "users:list"
	AuthzNameUsersUpdate           AuthzName = "users:update"
)

func (e *AuthzName) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = AuthzName(s)
	case string:
		*e = AuthzName(s)
	default:
		return fmt.Errorf("unsupported scan type for AuthzName: %T", src)
	}
	return nil
}

type NullAuthzName struct {
	AuthzName AuthzName
	Valid     bool // Valid is true if String is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullAuthzName) Scan(value interface{}) error {
	if value == nil {
		ns.AuthzName, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.AuthzName.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullAuthzName) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.AuthzName, nil
}

type RoleName string

const (
	RoleNameAdmin                 RoleName = "Admin"
	RoleNameServiceaccountcreator RoleName = "Service account creator"
	RoleNameServiceaccountowner   RoleName = "Service account owner"
	RoleNameTeamcreator           RoleName = "Team creator"
	RoleNameTeammember            RoleName = "Team member"
	RoleNameTeamowner             RoleName = "Team owner"
	RoleNameTeamviewer            RoleName = "Team viewer"
	RoleNameUseradmin             RoleName = "User admin"
	RoleNameUserviewer            RoleName = "User viewer"
)

func (e *RoleName) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = RoleName(s)
	case string:
		*e = RoleName(s)
	default:
		return fmt.Errorf("unsupported scan type for RoleName: %T", src)
	}
	return nil
}

type NullRoleName struct {
	RoleName RoleName
	Valid    bool // Valid is true if String is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullRoleName) Scan(value interface{}) error {
	if value == nil {
		ns.RoleName, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.RoleName.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullRoleName) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.RoleName, nil
}

type SystemName string

const (
	SystemNameConsole              SystemName = "console"
	SystemNameAzureGroup           SystemName = "azure:group"
	SystemNameGithubTeam           SystemName = "github:team"
	SystemNameGoogleGcpProject     SystemName = "google:gcp:project"
	SystemNameGoogleWorkspaceAdmin SystemName = "google:workspace-admin"
	SystemNameNaisNamespace        SystemName = "nais:namespace"
)

func (e *SystemName) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = SystemName(s)
	case string:
		*e = SystemName(s)
	default:
		return fmt.Errorf("unsupported scan type for SystemName: %T", src)
	}
	return nil
}

type NullSystemName struct {
	SystemName SystemName
	Valid      bool // Valid is true if String is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullSystemName) Scan(value interface{}) error {
	if value == nil {
		ns.SystemName, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.SystemName.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullSystemName) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.SystemName, nil
}

type ApiKey struct {
	ApiKey string
	UserID uuid.UUID
}

type AuditLog struct {
	ID              uuid.UUID
	CreatedAt       time.Time
	CorrelationID   uuid.UUID
	ActorEmail      sql.NullString
	SystemName      NullSystemName
	TargetUserEmail sql.NullString
	TargetTeamSlug  sql.NullString
	Action          AuditAction
	Message         string
}

type RoleAuthz struct {
	AuthzName AuthzName
	RoleName  RoleName
}

type SystemState struct {
	ID         uuid.UUID
	SystemName SystemName
	TeamID     uuid.UUID
	State      pgtype.JSONB
}

type Team struct {
	ID      uuid.UUID
	Slug    string
	Name    string
	Purpose sql.NullString
}

type TeamMetadatum struct {
	ID     uuid.UUID
	TeamID uuid.UUID
	Key    string
	Value  sql.NullString
}

type User struct {
	ID    uuid.UUID
	Email string
	Name  string
}

type UserRole struct {
	ID       uuid.UUID
	RoleName RoleName
	UserID   uuid.UUID
	TargetID uuid.NullUUID
}

type UserTeam struct {
	ID     uuid.UUID
	UserID uuid.UUID
	TeamID uuid.UUID
}