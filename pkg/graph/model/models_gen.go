// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
)

// API key type.
type APIKey struct {
	// The API key.
	Key string `json:"key"`
}

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

// Input for creating a new service account.
type CreateServiceAccountInput struct {
	// The name of the new service account. An email address will be automatically generated using the provided name.
	Name string `json:"name"`
}

// Input for creating a new team.
type CreateTeamInput struct {
	// Team slug. This value immutable.
	Slug string `json:"slug"`
	// Team name.
	Name string `json:"name"`
	// Team purpose.
	Purpose *string `json:"purpose"`
}

// Pagination metadata attached to queries resulting in a collection of data.
type PageInfo struct {
	// Total number of results that matches the query.
	Results int `json:"results"`
	// Which record number the returned collection starts at.
	Offset int `json:"offset"`
	// Maximum number of records included in the collection.
	Limit int `json:"limit"`
}

// When querying collections this input is used to control the offset and the page size of the returned slice.
//
// Please note that collections are not stateful, so data added or created in between your paginated requests might not be reflected in the returned result set.
type Pagination struct {
	// The offset to start fetching entries.
	Offset int `json:"offset"`
	// Number of entries per page.
	Limit int `json:"limit"`
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

// Input for updating an existing service account.
type UpdateServiceAccountInput struct {
	// The new name of the service account. The email address will be automatically updated.
	Name string `json:"name"`
}

// Input for updating an existing team.
type UpdateTeamInput struct {
	// Team name. Must contain a value when specified.
	Name *string `json:"name"`
	// Team purpose. Set to an empty string to remove the existing team purpose.
	Purpose *string `json:"purpose"`
}

// Direction of the sort.
type SortDirection string

const (
	// Sort ascending.
	SortDirectionAsc SortDirection = "ASC"
	// Sort descending.
	SortDirectionDesc SortDirection = "DESC"
)

var AllSortDirection = []SortDirection{
	SortDirectionAsc,
	SortDirectionDesc,
}

func (e SortDirection) IsValid() bool {
	switch e {
	case SortDirectionAsc, SortDirectionDesc:
		return true
	}
	return false
}

func (e SortDirection) String() string {
	return string(e)
}

func (e *SortDirection) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = SortDirection(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid SortDirection", str)
	}
	return nil
}

func (e SortDirection) MarshalGQL(w io.Writer) {
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
