// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/slug"
)

// Input for adding users to a team as members.
type AddTeamMembersInput struct {
	// ID of the team that should receive new members.
	TeamID *uuid.UUID `json:"teamId"`
	// List of user IDs that should be added to the team as members.
	UserIds []*uuid.UUID `json:"userIds"`
}

// Input for adding users to a team as owners.
type AddTeamOwnersInput struct {
	// ID of the team that should receive new owners.
	TeamID *uuid.UUID `json:"teamId"`
	// List of user IDs that should be added to the team as owners.
	UserIds []*uuid.UUID `json:"userIds"`
}

// Input for creating a new team.
type CreateTeamInput struct {
	// Team slug. This value immutable.
	Slug *slug.Slug `json:"slug"`
	// Team name.
	Name string `json:"name"`
	// Team purpose.
	Purpose *string `json:"purpose"`
}

// Input for removing users from a team.
type RemoveUsersFromTeamInput struct {
	// List of user IDs that should be removed from the team.
	UserIds []*uuid.UUID `json:"userIds"`
	// Team ID that should receive new users.
	TeamID *uuid.UUID `json:"teamId"`
}

// Input for setting team member role.
type SetTeamMemberRoleInput struct {
	// The ID of the team.
	TeamID *uuid.UUID `json:"teamId"`
	// The ID of the user.
	UserID *uuid.UUID `json:"userId"`
	// The team role to set.
	Role TeamRole `json:"role"`
}

// Team member.
type TeamMember struct {
	// User instance.
	User *db.User `json:"user"`
	// The role that the user has in the team.
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
	// Team name. Must contain a value when specified.
	Name *string `json:"name"`
	// Team purpose. Set to an empty string to remove the existing team purpose.
	Purpose *string `json:"purpose"`
}

// User team.
type UserTeam struct {
	// Team instance.
	Team *db.Team `json:"team"`
	// The role that the user has in the team.
	Role TeamRole `json:"role"`
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
