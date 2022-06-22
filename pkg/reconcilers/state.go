package reconcilers

import "github.com/google/uuid"

type AzureState struct {
	GroupID *uuid.UUID `json:"groupId"`
}
