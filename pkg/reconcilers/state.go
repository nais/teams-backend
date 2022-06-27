package reconcilers

import "github.com/google/uuid"

type AzureState struct {
	GroupID *uuid.UUID `json:"groupId"`
}

type GitHubState struct {
	Slug *string `json:"slug"`
}

type GoogleWorkspaceState struct {
	GroupID *string `json:"groupId"`
}

type GoogleGcpProjectState struct {
	Projects map[string]GoogleGcpEnvironmentProject `json:"projects"` // environment name is used as key
}

type GoogleGcpEnvironmentProject struct {
	ProjectID   string `json:"projectId"`   // Unique of the project, for instance `my-project-123`
	ProjectName string `json:"projectName"` // Unique project name, for instance `projects/<int>`
}
