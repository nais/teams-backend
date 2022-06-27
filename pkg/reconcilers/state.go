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
	Projects map[string]string `json:"projects"` // key is environment, value is the generated project name (projects/<int>)
}
