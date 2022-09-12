package reconcilers

import (
	"github.com/google/uuid"
	"github.com/nais/console/pkg/slug"
)

type AzureState struct {
	GroupID *uuid.UUID `json:"groupId"`
}

type GitHubState struct {
	Slug *slug.Slug `json:"slug"`
}

type GoogleWorkspaceState struct {
	GroupEmail *string `json:"groupEmail"`
}

type GoogleGcpProjectState struct {
	Projects map[string]GoogleGcpEnvironmentProject `json:"projects"` // environment name is used as key
}

type GoogleGcpEnvironmentProject struct {
	ProjectID   string `json:"projectId"`   // Unique of the project, for instance `my-project-123`
	ProjectName string `json:"projectName"` // Unique project name, for instance `projects/<int>`
}

type GoogleGcpNaisNamespaceState struct {
	Namespaces map[string]slug.Slug `json:"namespaces"` // Key is the environment for the team namespace
}
