// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/reconcilers"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
)

type AuditLogList struct {
	Nodes    []*db.AuditLog `json:"nodes"`
	PageInfo *PageInfo      `json:"pageInfo"`
}

// Input for creating a new team.
type CreateTeamInput struct {
	// Team slug. After creation, this value can not be changed.
	Slug slug.Slug `json:"slug"`
	// Team purpose.
	Purpose string `json:"purpose"`
	// Specify the Slack channel for the team.
	SlackChannel string `json:"slackChannel"`
}

// GCP project type.
type GcpProject struct {
	// The environment for the project.
	Environment string `json:"environment"`
	// The display name of the project.
	ProjectName string `json:"projectName"`
	// The GCP project ID.
	ProjectID string `json:"projectId"`
}

// Input for filtering GitHub repositories.
type GitHubRepositoriesFilter struct {
	// Include archived repositories or not. Default is false.
	IncludeArchivedRepositories bool `json:"includeArchivedRepositories"`
}

// Paginated GitHub repository type.
type GitHubRepositoryList struct {
	// The list of GitHub repositories.
	Nodes []*reconcilers.GitHubRepository `json:"nodes"`
	// Pagination information.
	PageInfo *PageInfo `json:"pageInfo"`
}

// NAIS namespace type.
type NaisNamespace struct {
	// The environment for the namespace.
	Environment string `json:"environment"`
	// The namespace.
	Namespace slug.Slug `json:"namespace"`
}

type PageInfo struct {
	TotalCount      int  `json:"totalCount"`
	HasNextPage     bool `json:"hasNextPage"`
	HasPreviousPage bool `json:"hasPreviousPage"`
}

// Reconciler configuration input.
type ReconcilerConfigInput struct {
	// Configuration key.
	Key sqlc.ReconcilerConfigKey `json:"key"`
	// Configuration value.
	Value string `json:"value"`
}

// Reconciler state type.
type ReconcilerState struct {
	// The GitHub team slug.
	GitHubTeamSlug *slug.Slug `json:"gitHubTeamSlug,omitempty"`
	// The Google Workspace group email.
	GoogleWorkspaceGroupEmail *string `json:"googleWorkspaceGroupEmail,omitempty"`
	// The Azure AD group ID.
	AzureADGroupID *uuid.UUID `json:"azureADGroupId,omitempty"`
	// A list of GCP projects.
	GcpProjects []*GcpProject `json:"gcpProjects"`
	// A list of NAIS namespaces.
	NaisNamespaces []*NaisNamespace `json:"naisNamespaces"`
	// Timestamp of when the NAIS deploy key was provisioned.
	NaisDeployKeyProvisioned *time.Time `json:"naisDeployKeyProvisioned,omitempty"`
	// Name of the GAR repository for the team.
	GarRepositoryName *string `json:"garRepositoryName,omitempty"`
}

// Slack alerts channel type.
type SlackAlertsChannel struct {
	// The environment for the alerts sent to the channel.
	Environment string `json:"environment"`
	// The name of the Slack channel.
	ChannelName string `json:"channelName"`
}

// Slack alerts channel input.
type SlackAlertsChannelInput struct {
	// The environment for the alerts sent to the channel.
	Environment string `json:"environment"`
	// The name of the Slack channel.
	ChannelName *string `json:"channelName,omitempty"`
}

// Sync error type.
type SyncError struct {
	// Creation time of the error.
	CreatedAt time.Time `json:"createdAt"`
	// The name of the reconciler.
	Reconciler sqlc.ReconcilerName `json:"reconciler"`
	// Error message.
	Error string `json:"error"`
}

// Paginated teams type.
type TeamList struct {
	// The list of teams.
	Nodes []*db.Team `json:"nodes"`
	// Pagination information.
	PageInfo *PageInfo `json:"pageInfo"`
}

// Team member input.
type TeamMemberInput struct {
	// The ID of user.
	UserID uuid.UUID `json:"userId"`
	// The role that the user will receive.
	Role TeamRole `json:"role"`
	// Reconcilers to opt the team member out of.
	ReconcilerOptOuts []sqlc.ReconcilerName `json:"reconcilerOptOuts,omitempty"`
}

type TeamMemberList struct {
	Nodes    []*TeamMember `json:"nodes"`
	PageInfo *PageInfo     `json:"pageInfo"`
}

// Team sync type.
type TeamSync struct {
	// The correlation ID for the sync.
	CorrelationID uuid.UUID `json:"correlationID"`
}

// Input for filtering teams.
type TeamsFilter struct {
	Github *TeamsFilterGitHub `json:"github,omitempty"`
}

type TeamsFilterGitHub struct {
	// Filter repostiories by repo name
	RepoName string `json:"repoName"`
	// Filter repostiories by permission name
	PermissionName string `json:"permissionName"`
}

// Input for updating an existing team.
type UpdateTeamInput struct {
	// Specify team purpose to update the existing value.
	Purpose *string `json:"purpose,omitempty"`
	// Specify the Slack channel to update the existing value.
	SlackChannel *string `json:"slackChannel,omitempty"`
	// A list of Slack channels for NAIS alerts.
	SlackAlertsChannels []*SlackAlertsChannelInput `json:"slackAlertsChannels,omitempty"`
}

type UserList struct {
	Nodes    []*db.User `json:"nodes"`
	PageInfo *PageInfo  `json:"pageInfo"`
}

// Repository authorizations.
type RepositoryAuthorization string

const (
	// Authorize for NAIS deployment.
	RepositoryAuthorizationDeploy RepositoryAuthorization = "DEPLOY"
)

var AllRepositoryAuthorization = []RepositoryAuthorization{
	RepositoryAuthorizationDeploy,
}

func (e RepositoryAuthorization) IsValid() bool {
	switch e {
	case RepositoryAuthorizationDeploy:
		return true
	}
	return false
}

func (e RepositoryAuthorization) String() string {
	return string(e)
}

func (e *RepositoryAuthorization) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = RepositoryAuthorization(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid RepositoryAuthorization", str)
	}
	return nil
}

func (e RepositoryAuthorization) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

// Available team roles.
type TeamRole string

const (
	// Regular member, read only access.
	TeamRoleMember TeamRole = "MEMBER"
	// Team owner, full access to the team.
	TeamRoleOwner TeamRole = "OWNER"
)

var AllTeamRole = []TeamRole{
	TeamRoleMember,
	TeamRoleOwner,
}

func (e TeamRole) IsValid() bool {
	switch e {
	case TeamRoleMember, TeamRoleOwner:
		return true
	}
	return false
}

func (e TeamRole) String() string {
	return string(e)
}

func (e *TeamRole) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = TeamRole(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid TeamRole", str)
	}
	return nil
}

func (e TeamRole) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

// User sync run status.
type UserSyncRunStatus string

const (
	// User sync run in progress.
	UserSyncRunStatusInProgress UserSyncRunStatus = "IN_PROGRESS"
	// Successful user sync run.
	UserSyncRunStatusSuccess UserSyncRunStatus = "SUCCESS"
	// Failed user sync run.
	UserSyncRunStatusFailure UserSyncRunStatus = "FAILURE"
)

var AllUserSyncRunStatus = []UserSyncRunStatus{
	UserSyncRunStatusInProgress,
	UserSyncRunStatusSuccess,
	UserSyncRunStatusFailure,
}

func (e UserSyncRunStatus) IsValid() bool {
	switch e {
	case UserSyncRunStatusInProgress, UserSyncRunStatusSuccess, UserSyncRunStatusFailure:
		return true
	}
	return false
}

func (e UserSyncRunStatus) String() string {
	return string(e)
}

func (e *UserSyncRunStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = UserSyncRunStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid UserSyncRunStatus", str)
	}
	return nil
}

func (e UserSyncRunStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
