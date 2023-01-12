// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
)

// Input for creating a new team.
type CreateTeamInput struct {
	// Team slug. After creation, this value can not be changed.
	Slug *slug.Slug `json:"slug"`
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

// NAIS namespace type.
type NaisNamespace struct {
	// The environment for the namespace.
	Environment string `json:"environment"`
	// The namespace.
	Namespace *slug.Slug `json:"namespace"`
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
	GitHubTeamSlug *slug.Slug `json:"gitHubTeamSlug"`
	// The Google Workspace group email.
	GoogleWorkspaceGroupEmail *string `json:"googleWorkspaceGroupEmail"`
	// The Azure AD group ID.
	AzureADGroupID *uuid.UUID `json:"azureADGroupId"`
	// A list of GCP projects.
	GcpProjects []*GcpProject `json:"gcpProjects"`
	// A list of NAIS namespaces.
	NaisNamespaces []*NaisNamespace `json:"naisNamespaces"`
	// Timestamp of when the NAIS deploy key was provisioned.
	NaisDeployKeyProvisioned *time.Time `json:"naisDeployKeyProvisioned"`
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

// Team member.
type TeamMember struct {
	// User instance.
	User *db.User `json:"user"`
	// The role that the user has in the team.
	Role TeamRole `json:"role"`
}

// Team membership type.
type TeamMembership struct {
	// Team instance.
	Team *db.Team `json:"team"`
	// The role that the member has in the team.
	Role TeamRole `json:"role"`
}

// Team sync type.
type TeamSync struct {
	// The team that will be synced.
	Team *db.Team `json:"team"`
	// The correlation ID for the sync.
	CorrelationID *uuid.UUID `json:"correlationID"`
}

// Input for updating an existing team.
type UpdateTeamInput struct {
	// Specify team purpose to update the existing value.
	Purpose *string `json:"purpose"`
	// Specify the Slack channel to update the existing value.
	SlackChannel *string `json:"slackChannel"`
}

// User sync type.
type UserSync struct {
	// Correlation ID of the triggered user synchronization.
	CorrelationID *uuid.UUID `json:"correlationID"`
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
