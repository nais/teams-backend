// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0

package sqlc

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/nais/teams-backend/pkg/slug"
)

type ReconcilerConfigKey string

const (
	ReconcilerConfigKeyAzureClientID     ReconcilerConfigKey = "azure:client_id"
	ReconcilerConfigKeyAzureClientSecret ReconcilerConfigKey = "azure:client_secret"
	ReconcilerConfigKeyAzureTenantID     ReconcilerConfigKey = "azure:tenant_id"
)

func (e *ReconcilerConfigKey) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ReconcilerConfigKey(s)
	case string:
		*e = ReconcilerConfigKey(s)
	default:
		return fmt.Errorf("unsupported scan type for ReconcilerConfigKey: %T", src)
	}
	return nil
}

type NullReconcilerConfigKey struct {
	ReconcilerConfigKey ReconcilerConfigKey
	Valid               bool // Valid is true if ReconcilerConfigKey is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullReconcilerConfigKey) Scan(value interface{}) error {
	if value == nil {
		ns.ReconcilerConfigKey, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ReconcilerConfigKey.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullReconcilerConfigKey) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ReconcilerConfigKey), nil
}

func (e ReconcilerConfigKey) Valid() bool {
	switch e {
	case ReconcilerConfigKeyAzureClientID,
		ReconcilerConfigKeyAzureClientSecret,
		ReconcilerConfigKeyAzureTenantID:
		return true
	}
	return false
}

func AllReconcilerConfigKeyValues() []ReconcilerConfigKey {
	return []ReconcilerConfigKey{
		ReconcilerConfigKeyAzureClientID,
		ReconcilerConfigKeyAzureClientSecret,
		ReconcilerConfigKeyAzureTenantID,
	}
}

type ReconcilerName string

const (
	ReconcilerNameAzureGroup           ReconcilerName = "azure:group"
	ReconcilerNameGithubTeam           ReconcilerName = "github:team"
	ReconcilerNameGoogleGcpGar         ReconcilerName = "google:gcp:gar"
	ReconcilerNameGoogleGcpProject     ReconcilerName = "google:gcp:project"
	ReconcilerNameGoogleWorkspaceAdmin ReconcilerName = "google:workspace-admin"
	ReconcilerNameNaisDependencytrack  ReconcilerName = "nais:dependencytrack"
	ReconcilerNameNaisDeploy           ReconcilerName = "nais:deploy"
	ReconcilerNameNaisNamespace        ReconcilerName = "nais:namespace"
)

func (e *ReconcilerName) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ReconcilerName(s)
	case string:
		*e = ReconcilerName(s)
	default:
		return fmt.Errorf("unsupported scan type for ReconcilerName: %T", src)
	}
	return nil
}

type NullReconcilerName struct {
	ReconcilerName ReconcilerName
	Valid          bool // Valid is true if ReconcilerName is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullReconcilerName) Scan(value interface{}) error {
	if value == nil {
		ns.ReconcilerName, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ReconcilerName.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullReconcilerName) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ReconcilerName), nil
}

func (e ReconcilerName) Valid() bool {
	switch e {
	case ReconcilerNameAzureGroup,
		ReconcilerNameGithubTeam,
		ReconcilerNameGoogleGcpGar,
		ReconcilerNameGoogleGcpProject,
		ReconcilerNameGoogleWorkspaceAdmin,
		ReconcilerNameNaisDependencytrack,
		ReconcilerNameNaisDeploy,
		ReconcilerNameNaisNamespace:
		return true
	}
	return false
}

func AllReconcilerNameValues() []ReconcilerName {
	return []ReconcilerName{
		ReconcilerNameAzureGroup,
		ReconcilerNameGithubTeam,
		ReconcilerNameGoogleGcpGar,
		ReconcilerNameGoogleGcpProject,
		ReconcilerNameGoogleWorkspaceAdmin,
		ReconcilerNameNaisDependencytrack,
		ReconcilerNameNaisDeploy,
		ReconcilerNameNaisNamespace,
	}
}

type RepositoryAuthorizationEnum string

const (
	RepositoryAuthorizationEnumDeploy RepositoryAuthorizationEnum = "deploy"
)

func (e *RepositoryAuthorizationEnum) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = RepositoryAuthorizationEnum(s)
	case string:
		*e = RepositoryAuthorizationEnum(s)
	default:
		return fmt.Errorf("unsupported scan type for RepositoryAuthorizationEnum: %T", src)
	}
	return nil
}

type NullRepositoryAuthorizationEnum struct {
	RepositoryAuthorizationEnum RepositoryAuthorizationEnum
	Valid                       bool // Valid is true if RepositoryAuthorizationEnum is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullRepositoryAuthorizationEnum) Scan(value interface{}) error {
	if value == nil {
		ns.RepositoryAuthorizationEnum, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.RepositoryAuthorizationEnum.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullRepositoryAuthorizationEnum) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.RepositoryAuthorizationEnum), nil
}

func (e RepositoryAuthorizationEnum) Valid() bool {
	switch e {
	case RepositoryAuthorizationEnumDeploy:
		return true
	}
	return false
}

func AllRepositoryAuthorizationEnumValues() []RepositoryAuthorizationEnum {
	return []RepositoryAuthorizationEnum{
		RepositoryAuthorizationEnumDeploy,
	}
}

type RoleName string

const (
	RoleNameAdmin                 RoleName = "Admin"
	RoleNameDeploykeyviewer       RoleName = "Deploy key viewer"
	RoleNameServiceaccountcreator RoleName = "Service account creator"
	RoleNameServiceaccountowner   RoleName = "Service account owner"
	RoleNameSynchronizer          RoleName = "Synchronizer"
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
	Valid    bool // Valid is true if RoleName is not NULL
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
	return string(ns.RoleName), nil
}

func (e RoleName) Valid() bool {
	switch e {
	case RoleNameAdmin,
		RoleNameDeploykeyviewer,
		RoleNameServiceaccountcreator,
		RoleNameServiceaccountowner,
		RoleNameSynchronizer,
		RoleNameTeamcreator,
		RoleNameTeammember,
		RoleNameTeamowner,
		RoleNameTeamviewer,
		RoleNameUseradmin,
		RoleNameUserviewer:
		return true
	}
	return false
}

func AllRoleNameValues() []RoleName {
	return []RoleName{
		RoleNameAdmin,
		RoleNameDeploykeyviewer,
		RoleNameServiceaccountcreator,
		RoleNameServiceaccountowner,
		RoleNameSynchronizer,
		RoleNameTeamcreator,
		RoleNameTeammember,
		RoleNameTeamowner,
		RoleNameTeamviewer,
		RoleNameUseradmin,
		RoleNameUserviewer,
	}
}

type ApiKey struct {
	ApiKey           string
	ServiceAccountID uuid.UUID
}

type AuditLog struct {
	ID               uuid.UUID
	CreatedAt        time.Time
	CorrelationID    uuid.UUID
	ComponentName    string
	Actor            *string
	Action           string
	Message          string
	TargetType       string
	TargetIdentifier string
}

type FirstRun struct {
	FirstRun bool
}

type Reconciler struct {
	Name        ReconcilerName
	DisplayName string
	Description string
	Enabled     bool
	RunOrder    int32
}

type ReconcilerConfig struct {
	Reconciler  ReconcilerName
	Key         ReconcilerConfigKey
	DisplayName string
	Description string
	Value       *string
	Secret      bool
}

type ReconcilerError struct {
	ID            int64
	CorrelationID uuid.UUID
	Reconciler    ReconcilerName
	CreatedAt     time.Time
	ErrorMessage  string
	TeamSlug      slug.Slug
}

type ReconcilerOptOut struct {
	TeamSlug       slug.Slug
	UserID         uuid.UUID
	ReconcilerName ReconcilerName
}

type ReconcilerState struct {
	Reconciler ReconcilerName
	State      pgtype.JSONB
	TeamSlug   slug.Slug
}

type RepositoryAuthorization struct {
	TeamSlug                string
	GithubRepository        string
	RepositoryAuthorization RepositoryAuthorizationEnum
}

type ServiceAccount struct {
	ID   uuid.UUID
	Name string
}

type ServiceAccountRole struct {
	ID                     int32
	RoleName               RoleName
	ServiceAccountID       uuid.UUID
	TargetTeamSlug         *slug.Slug
	TargetServiceAccountID *uuid.UUID
}

type Session struct {
	ID      uuid.UUID
	UserID  uuid.UUID
	Expires time.Time
}

type SlackAlertsChannel struct {
	TeamSlug    slug.Slug
	Environment string
	ChannelName string
}

type Team struct {
	Slug               slug.Slug
	Purpose            string
	LastSuccessfulSync *time.Time
	SlackChannel       string
}

type TeamDeleteKey struct {
	Key         uuid.UUID
	TeamSlug    slug.Slug
	CreatedAt   time.Time
	CreatedBy   uuid.UUID
	ConfirmedAt *time.Time
}

type User struct {
	ID         uuid.UUID
	Email      string
	Name       string
	ExternalID string
}

type UserRole struct {
	ID                     int32
	RoleName               RoleName
	UserID                 uuid.UUID
	TargetTeamSlug         *slug.Slug
	TargetServiceAccountID *uuid.UUID
}
