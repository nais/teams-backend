package reconcilers

import (
	"time"

	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/slug"
)

type AzureState struct {
	GroupID *uuid.UUID `json:"groupId"`
}

type DependencyTrackState struct {
	Instances []*DependencyTrackInstanceState `json:"instances"`
}

type DependencyTrackInstanceState struct {
	Endpoint string   `json:"endpoint"`
	TeamID   string   `json:"teamId"`
	Members  []string `json:"members"`
}

type GitHubState struct {
	Slug         *slug.Slug          `json:"slug"`
	Repositories []*GitHubRepository `json:"repositories"`
}

type GitHubRepository struct {
	Name        string                        `json:"name"`
	Permissions []*GitHubRepositoryPermission `json:"permissions"`
	Archived    bool                          `json:"archived"`
}

type GitHubRepositoryPermission struct {
	Name    string `json:"name"`
	Granted bool   `json:"granted"`
}

type GoogleWorkspaceState struct {
	GroupEmail *string `json:"groupEmail"`
}

type GoogleGcpProjectState struct {
	Projects map[string]GoogleGcpEnvironmentProject `json:"projects"` // environment name is used as key
}

type GoogleGcpEnvironmentProject struct {
	ProjectID string `json:"projectId"` // Unique of the project, for instance `my-project-123`
}

type NaisNamespaceState struct {
	Namespaces map[string]slug.Slug `json:"namespaces"` // Key is the environment for the team namespace
}

type NaisDeployKeyState struct {
	Provisioned *time.Time `json:"provisioned"`
}

type GoogleGarState struct {
	RepositoryName *string `json:"repopsitoryName"`
}
