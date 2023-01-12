// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0

package sqlc

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/nais/console/pkg/slug"
)

type AuditAction string

const (
	AuditActionGraphqlApiApiKeyCreate                    AuditAction = "graphql-api:api-key:create"
	AuditActionGraphqlApiApiKeyDelete                    AuditAction = "graphql-api:api-key:delete"
	AuditActionGraphqlApiServiceAccountCreate            AuditAction = "graphql-api:service-account:create"
	AuditActionGraphqlApiServiceAccountDelete            AuditAction = "graphql-api:service-account:delete"
	AuditActionGraphqlApiServiceAccountUpdate            AuditAction = "graphql-api:service-account:update"
	AuditActionGraphqlApiTeamAddMember                   AuditAction = "graphql-api:team:add-member"
	AuditActionGraphqlApiTeamAddOwner                    AuditAction = "graphql-api:team:add-owner"
	AuditActionGraphqlApiTeamCreate                      AuditAction = "graphql-api:team:create"
	AuditActionGraphqlApiTeamRemoveMember                AuditAction = "graphql-api:team:remove-member"
	AuditActionGraphqlApiTeamSetMemberRole               AuditAction = "graphql-api:team:set-member-role"
	AuditActionGraphqlApiTeamSync                        AuditAction = "graphql-api:team:sync"
	AuditActionGraphqlApiTeamUpdate                      AuditAction = "graphql-api:team:update"
	AuditActionUsersyncPrepare                           AuditAction = "usersync:prepare"
	AuditActionUsersyncListRemote                        AuditAction = "usersync:list:remote"
	AuditActionUsersyncListLocal                         AuditAction = "usersync:list:local"
	AuditActionUsersyncCreate                            AuditAction = "usersync:create"
	AuditActionUsersyncUpdate                            AuditAction = "usersync:update"
	AuditActionUsersyncDelete                            AuditAction = "usersync:delete"
	AuditActionAzureGroupCreate                          AuditAction = "azure:group:create"
	AuditActionAzureGroupAddMember                       AuditAction = "azure:group:add-member"
	AuditActionAzureGroupAddMembers                      AuditAction = "azure:group:add-members"
	AuditActionAzureGroupDeleteMember                    AuditAction = "azure:group:delete-member"
	AuditActionGithubTeamCreate                          AuditAction = "github:team:create"
	AuditActionGithubTeamAddMembers                      AuditAction = "github:team:add-members"
	AuditActionGithubTeamAddMember                       AuditAction = "github:team:add-member"
	AuditActionGithubTeamDeleteMember                    AuditAction = "github:team:delete-member"
	AuditActionGithubTeamMapSsoUser                      AuditAction = "github:team:map-sso-user"
	AuditActionGoogleWorkspaceAdminCreate                AuditAction = "google:workspace-admin:create"
	AuditActionGoogleWorkspaceAdminAddMember             AuditAction = "google:workspace-admin:add-member"
	AuditActionGoogleWorkspaceAdminAddMembers            AuditAction = "google:workspace-admin:add-members"
	AuditActionGoogleWorkspaceAdminDeleteMember          AuditAction = "google:workspace-admin:delete-member"
	AuditActionGoogleWorkspaceAdminAddToGkeSecurityGroup AuditAction = "google:workspace-admin:add-to-gke-security-group"
	AuditActionGoogleGcpProjectCreateProject             AuditAction = "google:gcp:project:create-project"
	AuditActionGoogleGcpProjectAssignPermissions         AuditAction = "google:gcp:project:assign-permissions"
	AuditActionNaisNamespaceCreateNamespace              AuditAction = "nais:namespace:create-namespace"
	AuditActionLegacyImporterTeamCreate                  AuditAction = "legacy-importer:team:create"
	AuditActionLegacyImporterTeamAddMember               AuditAction = "legacy-importer:team:add-member"
	AuditActionLegacyImporterTeamAddOwner                AuditAction = "legacy-importer:team:add-owner"
	AuditActionLegacyImporterUserCreate                  AuditAction = "legacy-importer:user:create"
	AuditActionGraphqlApiRolesAssignGlobalRole           AuditAction = "graphql-api:roles:assign-global-role"
	AuditActionGraphqlApiRolesRevokeGlobalRole           AuditAction = "graphql-api:roles:revoke-global-role"
	AuditActionGoogleGcpProjectSetBillingInfo            AuditAction = "google:gcp:project:set-billing-info"
	AuditActionGoogleGcpProjectCreateCnrmServiceAccount  AuditAction = "google:gcp:project:create-cnrm-service-account"
	AuditActionGraphqlApiReconcilersConfigure            AuditAction = "graphql-api:reconcilers:configure"
	AuditActionGraphqlApiReconcilersDisable              AuditAction = "graphql-api:reconcilers:disable"
	AuditActionGraphqlApiReconcilersEnable               AuditAction = "graphql-api:reconcilers:enable"
	AuditActionGraphqlApiReconcilersReset                AuditAction = "graphql-api:reconcilers:reset"
	AuditActionGraphqlApiTeamDisable                     AuditAction = "graphql-api:team:disable"
	AuditActionGraphqlApiTeamEnable                      AuditAction = "graphql-api:team:enable"
	AuditActionUsersyncAssignAdminRole                   AuditAction = "usersync:assign-admin-role"
	AuditActionUsersyncRevokeAdminRole                   AuditAction = "usersync:revoke-admin-role"
	AuditActionGraphqlApiReconcilersUpdateTeamState      AuditAction = "graphql-api:reconcilers:update-team-state"
	AuditActionNaisDeployProvisionDeployKey              AuditAction = "nais:deploy:provision-deploy-key"
	AuditActionGoogleGcpProjectDeleteCnrmServiceAccount  AuditAction = "google:gcp:project:delete-cnrm-service-account"
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
	Valid       bool // Valid is true if AuditAction is not NULL
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

func (e AuditAction) Valid() bool {
	switch e {
	case AuditActionGraphqlApiApiKeyCreate,
		AuditActionGraphqlApiApiKeyDelete,
		AuditActionGraphqlApiServiceAccountCreate,
		AuditActionGraphqlApiServiceAccountDelete,
		AuditActionGraphqlApiServiceAccountUpdate,
		AuditActionGraphqlApiTeamAddMember,
		AuditActionGraphqlApiTeamAddOwner,
		AuditActionGraphqlApiTeamCreate,
		AuditActionGraphqlApiTeamRemoveMember,
		AuditActionGraphqlApiTeamSetMemberRole,
		AuditActionGraphqlApiTeamSync,
		AuditActionGraphqlApiTeamUpdate,
		AuditActionUsersyncPrepare,
		AuditActionUsersyncListRemote,
		AuditActionUsersyncListLocal,
		AuditActionUsersyncCreate,
		AuditActionUsersyncUpdate,
		AuditActionUsersyncDelete,
		AuditActionAzureGroupCreate,
		AuditActionAzureGroupAddMember,
		AuditActionAzureGroupAddMembers,
		AuditActionAzureGroupDeleteMember,
		AuditActionGithubTeamCreate,
		AuditActionGithubTeamAddMembers,
		AuditActionGithubTeamAddMember,
		AuditActionGithubTeamDeleteMember,
		AuditActionGithubTeamMapSsoUser,
		AuditActionGoogleWorkspaceAdminCreate,
		AuditActionGoogleWorkspaceAdminAddMember,
		AuditActionGoogleWorkspaceAdminAddMembers,
		AuditActionGoogleWorkspaceAdminDeleteMember,
		AuditActionGoogleWorkspaceAdminAddToGkeSecurityGroup,
		AuditActionGoogleGcpProjectCreateProject,
		AuditActionGoogleGcpProjectAssignPermissions,
		AuditActionNaisNamespaceCreateNamespace,
		AuditActionLegacyImporterTeamCreate,
		AuditActionLegacyImporterTeamAddMember,
		AuditActionLegacyImporterTeamAddOwner,
		AuditActionLegacyImporterUserCreate,
		AuditActionGraphqlApiRolesAssignGlobalRole,
		AuditActionGraphqlApiRolesRevokeGlobalRole,
		AuditActionGoogleGcpProjectSetBillingInfo,
		AuditActionGoogleGcpProjectCreateCnrmServiceAccount,
		AuditActionGraphqlApiReconcilersConfigure,
		AuditActionGraphqlApiReconcilersDisable,
		AuditActionGraphqlApiReconcilersEnable,
		AuditActionGraphqlApiReconcilersReset,
		AuditActionGraphqlApiTeamDisable,
		AuditActionGraphqlApiTeamEnable,
		AuditActionUsersyncAssignAdminRole,
		AuditActionUsersyncRevokeAdminRole,
		AuditActionGraphqlApiReconcilersUpdateTeamState,
		AuditActionNaisDeployProvisionDeployKey,
		AuditActionGoogleGcpProjectDeleteCnrmServiceAccount:
		return true
	}
	return false
}

func AllAuditActionValues() []AuditAction {
	return []AuditAction{
		AuditActionGraphqlApiApiKeyCreate,
		AuditActionGraphqlApiApiKeyDelete,
		AuditActionGraphqlApiServiceAccountCreate,
		AuditActionGraphqlApiServiceAccountDelete,
		AuditActionGraphqlApiServiceAccountUpdate,
		AuditActionGraphqlApiTeamAddMember,
		AuditActionGraphqlApiTeamAddOwner,
		AuditActionGraphqlApiTeamCreate,
		AuditActionGraphqlApiTeamRemoveMember,
		AuditActionGraphqlApiTeamSetMemberRole,
		AuditActionGraphqlApiTeamSync,
		AuditActionGraphqlApiTeamUpdate,
		AuditActionUsersyncPrepare,
		AuditActionUsersyncListRemote,
		AuditActionUsersyncListLocal,
		AuditActionUsersyncCreate,
		AuditActionUsersyncUpdate,
		AuditActionUsersyncDelete,
		AuditActionAzureGroupCreate,
		AuditActionAzureGroupAddMember,
		AuditActionAzureGroupAddMembers,
		AuditActionAzureGroupDeleteMember,
		AuditActionGithubTeamCreate,
		AuditActionGithubTeamAddMembers,
		AuditActionGithubTeamAddMember,
		AuditActionGithubTeamDeleteMember,
		AuditActionGithubTeamMapSsoUser,
		AuditActionGoogleWorkspaceAdminCreate,
		AuditActionGoogleWorkspaceAdminAddMember,
		AuditActionGoogleWorkspaceAdminAddMembers,
		AuditActionGoogleWorkspaceAdminDeleteMember,
		AuditActionGoogleWorkspaceAdminAddToGkeSecurityGroup,
		AuditActionGoogleGcpProjectCreateProject,
		AuditActionGoogleGcpProjectAssignPermissions,
		AuditActionNaisNamespaceCreateNamespace,
		AuditActionLegacyImporterTeamCreate,
		AuditActionLegacyImporterTeamAddMember,
		AuditActionLegacyImporterTeamAddOwner,
		AuditActionLegacyImporterUserCreate,
		AuditActionGraphqlApiRolesAssignGlobalRole,
		AuditActionGraphqlApiRolesRevokeGlobalRole,
		AuditActionGoogleGcpProjectSetBillingInfo,
		AuditActionGoogleGcpProjectCreateCnrmServiceAccount,
		AuditActionGraphqlApiReconcilersConfigure,
		AuditActionGraphqlApiReconcilersDisable,
		AuditActionGraphqlApiReconcilersEnable,
		AuditActionGraphqlApiReconcilersReset,
		AuditActionGraphqlApiTeamDisable,
		AuditActionGraphqlApiTeamEnable,
		AuditActionUsersyncAssignAdminRole,
		AuditActionUsersyncRevokeAdminRole,
		AuditActionGraphqlApiReconcilersUpdateTeamState,
		AuditActionNaisDeployProvisionDeployKey,
		AuditActionGoogleGcpProjectDeleteCnrmServiceAccount,
	}
}

type AuditLogsTargetType string

const (
	AuditLogsTargetTypeUser           AuditLogsTargetType = "user"
	AuditLogsTargetTypeTeam           AuditLogsTargetType = "team"
	AuditLogsTargetTypeServiceAccount AuditLogsTargetType = "service_account"
	AuditLogsTargetTypeReconciler     AuditLogsTargetType = "reconciler"
)

func (e *AuditLogsTargetType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = AuditLogsTargetType(s)
	case string:
		*e = AuditLogsTargetType(s)
	default:
		return fmt.Errorf("unsupported scan type for AuditLogsTargetType: %T", src)
	}
	return nil
}

type NullAuditLogsTargetType struct {
	AuditLogsTargetType AuditLogsTargetType
	Valid               bool // Valid is true if AuditLogsTargetType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullAuditLogsTargetType) Scan(value interface{}) error {
	if value == nil {
		ns.AuditLogsTargetType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.AuditLogsTargetType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullAuditLogsTargetType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.AuditLogsTargetType, nil
}

func (e AuditLogsTargetType) Valid() bool {
	switch e {
	case AuditLogsTargetTypeUser,
		AuditLogsTargetTypeTeam,
		AuditLogsTargetTypeServiceAccount,
		AuditLogsTargetTypeReconciler:
		return true
	}
	return false
}

func AllAuditLogsTargetTypeValues() []AuditLogsTargetType {
	return []AuditLogsTargetType{
		AuditLogsTargetTypeUser,
		AuditLogsTargetTypeTeam,
		AuditLogsTargetTypeServiceAccount,
		AuditLogsTargetTypeReconciler,
	}
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
	AuthzNameTeamsSynchronize      AuthzName = "teams:synchronize"
	AuthzNameUsersyncSynchronize   AuthzName = "usersync:synchronize"
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
	Valid     bool // Valid is true if AuthzName is not NULL
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

func (e AuthzName) Valid() bool {
	switch e {
	case AuthzNameAuditLogsRead,
		AuthzNameServiceAccountsCreate,
		AuthzNameServiceAccountsDelete,
		AuthzNameServiceAccountsList,
		AuthzNameServiceAccountsRead,
		AuthzNameServiceAccountsUpdate,
		AuthzNameSystemStatesDelete,
		AuthzNameSystemStatesRead,
		AuthzNameSystemStatesUpdate,
		AuthzNameTeamsCreate,
		AuthzNameTeamsDelete,
		AuthzNameTeamsList,
		AuthzNameTeamsRead,
		AuthzNameTeamsUpdate,
		AuthzNameUsersList,
		AuthzNameUsersUpdate,
		AuthzNameTeamsSynchronize,
		AuthzNameUsersyncSynchronize:
		return true
	}
	return false
}

func AllAuthzNameValues() []AuthzName {
	return []AuthzName{
		AuthzNameAuditLogsRead,
		AuthzNameServiceAccountsCreate,
		AuthzNameServiceAccountsDelete,
		AuthzNameServiceAccountsList,
		AuthzNameServiceAccountsRead,
		AuthzNameServiceAccountsUpdate,
		AuthzNameSystemStatesDelete,
		AuthzNameSystemStatesRead,
		AuthzNameSystemStatesUpdate,
		AuthzNameTeamsCreate,
		AuthzNameTeamsDelete,
		AuthzNameTeamsList,
		AuthzNameTeamsRead,
		AuthzNameTeamsUpdate,
		AuthzNameUsersList,
		AuthzNameUsersUpdate,
		AuthzNameTeamsSynchronize,
		AuthzNameUsersyncSynchronize,
	}
}

type ReconcilerConfigKey string

const (
	ReconcilerConfigKeyAzureClientID           ReconcilerConfigKey = "azure:client_id"
	ReconcilerConfigKeyAzureClientSecret       ReconcilerConfigKey = "azure:client_secret"
	ReconcilerConfigKeyAzureTenantID           ReconcilerConfigKey = "azure:tenant_id"
	ReconcilerConfigKeyGithubOrg               ReconcilerConfigKey = "github:org"
	ReconcilerConfigKeyGithubAppID             ReconcilerConfigKey = "github:app_id"
	ReconcilerConfigKeyGithubAppInstallationID ReconcilerConfigKey = "github:app_installation_id"
	ReconcilerConfigKeyGithubAppPrivateKey     ReconcilerConfigKey = "github:app_private_key"
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
	return ns.ReconcilerConfigKey, nil
}

func (e ReconcilerConfigKey) Valid() bool {
	switch e {
	case ReconcilerConfigKeyAzureClientID,
		ReconcilerConfigKeyAzureClientSecret,
		ReconcilerConfigKeyAzureTenantID,
		ReconcilerConfigKeyGithubOrg,
		ReconcilerConfigKeyGithubAppID,
		ReconcilerConfigKeyGithubAppInstallationID,
		ReconcilerConfigKeyGithubAppPrivateKey:
		return true
	}
	return false
}

func AllReconcilerConfigKeyValues() []ReconcilerConfigKey {
	return []ReconcilerConfigKey{
		ReconcilerConfigKeyAzureClientID,
		ReconcilerConfigKeyAzureClientSecret,
		ReconcilerConfigKeyAzureTenantID,
		ReconcilerConfigKeyGithubOrg,
		ReconcilerConfigKeyGithubAppID,
		ReconcilerConfigKeyGithubAppInstallationID,
		ReconcilerConfigKeyGithubAppPrivateKey,
	}
}

type ReconcilerName string

const (
	ReconcilerNameAzureGroup           ReconcilerName = "azure:group"
	ReconcilerNameGithubTeam           ReconcilerName = "github:team"
	ReconcilerNameGoogleGcpProject     ReconcilerName = "google:gcp:project"
	ReconcilerNameGoogleWorkspaceAdmin ReconcilerName = "google:workspace-admin"
	ReconcilerNameNaisNamespace        ReconcilerName = "nais:namespace"
	ReconcilerNameNaisDeploy           ReconcilerName = "nais:deploy"
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
	return ns.ReconcilerName, nil
}

func (e ReconcilerName) Valid() bool {
	switch e {
	case ReconcilerNameAzureGroup,
		ReconcilerNameGithubTeam,
		ReconcilerNameGoogleGcpProject,
		ReconcilerNameGoogleWorkspaceAdmin,
		ReconcilerNameNaisNamespace,
		ReconcilerNameNaisDeploy:
		return true
	}
	return false
}

func AllReconcilerNameValues() []ReconcilerName {
	return []ReconcilerName{
		ReconcilerNameAzureGroup,
		ReconcilerNameGithubTeam,
		ReconcilerNameGoogleGcpProject,
		ReconcilerNameGoogleWorkspaceAdmin,
		ReconcilerNameNaisNamespace,
		ReconcilerNameNaisDeploy,
	}
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
	RoleNameSynchronizer          RoleName = "Synchronizer"
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
	return ns.RoleName, nil
}

func (e RoleName) Valid() bool {
	switch e {
	case RoleNameAdmin,
		RoleNameServiceaccountcreator,
		RoleNameServiceaccountowner,
		RoleNameTeamcreator,
		RoleNameTeammember,
		RoleNameTeamowner,
		RoleNameTeamviewer,
		RoleNameUseradmin,
		RoleNameUserviewer,
		RoleNameSynchronizer:
		return true
	}
	return false
}

func AllRoleNameValues() []RoleName {
	return []RoleName{
		RoleNameAdmin,
		RoleNameServiceaccountcreator,
		RoleNameServiceaccountowner,
		RoleNameTeamcreator,
		RoleNameTeammember,
		RoleNameTeamowner,
		RoleNameTeamviewer,
		RoleNameUseradmin,
		RoleNameUserviewer,
		RoleNameSynchronizer,
	}
}

type SystemName string

const (
	SystemNameConsole              SystemName = "console"
	SystemNameAzureGroup           SystemName = "azure:group"
	SystemNameGithubTeam           SystemName = "github:team"
	SystemNameGoogleGcpProject     SystemName = "google:gcp:project"
	SystemNameGoogleWorkspaceAdmin SystemName = "google:workspace-admin"
	SystemNameNaisNamespace        SystemName = "nais:namespace"
	SystemNameGraphqlApi           SystemName = "graphql-api"
	SystemNameUsersync             SystemName = "usersync"
	SystemNameLegacyImporter       SystemName = "legacy-importer"
	SystemNameAuthn                SystemName = "authn"
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
	Valid      bool // Valid is true if SystemName is not NULL
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

func (e SystemName) Valid() bool {
	switch e {
	case SystemNameConsole,
		SystemNameAzureGroup,
		SystemNameGithubTeam,
		SystemNameGoogleGcpProject,
		SystemNameGoogleWorkspaceAdmin,
		SystemNameNaisNamespace,
		SystemNameGraphqlApi,
		SystemNameUsersync,
		SystemNameLegacyImporter,
		SystemNameAuthn:
		return true
	}
	return false
}

func AllSystemNameValues() []SystemName {
	return []SystemName{
		SystemNameConsole,
		SystemNameAzureGroup,
		SystemNameGithubTeam,
		SystemNameGoogleGcpProject,
		SystemNameGoogleWorkspaceAdmin,
		SystemNameNaisNamespace,
		SystemNameGraphqlApi,
		SystemNameUsersync,
		SystemNameLegacyImporter,
		SystemNameAuthn,
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
	SystemName       SystemName
	Actor            sql.NullString
	Action           AuditAction
	Message          string
	TargetType       AuditLogsTargetType
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
	Value       sql.NullString
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

type ReconcilerState struct {
	Reconciler ReconcilerName
	State      pgtype.JSONB
	TeamSlug   slug.Slug
}

type RoleAuthz struct {
	AuthzName AuthzName
	RoleName  RoleName
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
	TargetServiceAccountID uuid.NullUUID
}

type Session struct {
	ID      uuid.UUID
	UserID  uuid.UUID
	Expires time.Time
}

type Team struct {
	Slug               slug.Slug
	Purpose            string
	Enabled            bool
	LastSuccessfulSync sql.NullTime
	SlackChannel       string
}

type TeamMetadatum struct {
	Key      string
	Value    sql.NullString
	TeamSlug slug.Slug
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
	TargetServiceAccountID uuid.NullUUID
}
