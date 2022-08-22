// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
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

// Audit log type.
type AuditLog struct {
	// ID of the log entry.
	ID *uuid.UUID `json:"id"`
	// The related system.
	TargetSystem *System `json:"targetSystem"`
	// The related correlation.
	Correlation *Correlation `json:"correlation"`
	// The actor who performed the action in the entry. When this field is empty it means that the console system itself performed the action.
	Actor *db.User `json:"actor"`
	// The target user.
	TargetUser *db.User `json:"targetUser"`
	// The target team.
	TargetTeam *db.Team `json:"targetTeam"`
	// String representation of the action performed.
	Action string `json:"action"`
	// Log entry message.
	Message string `json:"message"`
	// Creation time of the log entry.
	CreatedAt time.Time `json:"createdAt"`
}

// Audit log collection.
type AuditLogs struct {
	// Object related to pagination of the collection.
	PageInfo *PageInfo `json:"pageInfo"`
	// The list of audit log entries in the collection.
	Nodes []*AuditLog `json:"nodes"`
}

// Input for filtering a collection of audit log entries.
type AuditLogsQuery struct {
	// Filter by actor ID.
	ActorID *uuid.UUID `json:"actorId"`
	// Filter by correlation ID.
	CorrelationID *uuid.UUID `json:"correlationId"`
	// Filter by target system ID.
	TargetSystemID *uuid.UUID `json:"targetSystemId"`
	// Filter by target team ID.
	TargetTeamID *uuid.UUID `json:"targetTeamId"`
	// Filter by target user ID.
	TargetUserID *uuid.UUID `json:"targetUserId"`
}

// Input for sorting a collection of audit log entries.
type AuditLogsSort struct {
	// Field to sort by.
	Field AuditLogSortField `json:"field"`
	// Sort direction.
	Direction SortDirection `json:"direction"`
}

// Correlation type.
type Correlation struct {
	// ID of the correlation.
	ID *uuid.UUID `json:"id"`
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

// System type.
type System struct {
	// Unique ID of the system.
	ID *uuid.UUID `json:"id"`
	// Name of the system.
	Name string `json:"name"`
}

// Team member.
type TeamMember struct {
	// User instance.
	User *db.User `json:"user"`
	// The role that the user has in the team.
	Role TeamRole `json:"role"`
}

// Team collection.
type Teams struct {
	// Object related to pagination of the collection.
	PageInfo *PageInfo `json:"pageInfo"`
	// The list of team objects in the collection.
	Nodes []*db.Team `json:"nodes"`
}

// Input for filtering a collection of teams.
type TeamsQuery struct {
	// Filter by slug.
	Slug *string `json:"slug"`
	// Filter by name.
	Name *string `json:"name"`
}

// Input for sorting a collection of teams.
type TeamsSort struct {
	// Field to sort by.
	Field TeamSortField `json:"field"`
	// Sort direction.
	Direction SortDirection `json:"direction"`
}

// Input for updating an existing service account.
type UpdateServiceAccountInput struct {
	// The new name of the service account. The email address will be automatically updated.
	Name string `json:"name"`
}

// Input for updating an existing team.
type UpdateTeamInput struct {
	// Team name.
	Name *string `json:"name"`
	// Team purpose.
	Purpose *string `json:"purpose"`
}

// Fields to sort the collection by.
type AuditLogSortField string

const (
	// Sort by creation time.
	AuditLogSortFieldCreatedAt AuditLogSortField = "CREATED_AT"
)

var AllAuditLogSortField = []AuditLogSortField{
	AuditLogSortFieldCreatedAt,
}

func (e AuditLogSortField) IsValid() bool {
	switch e {
	case AuditLogSortFieldCreatedAt:
		return true
	}
	return false
}

func (e AuditLogSortField) String() string {
	return string(e)
}

func (e *AuditLogSortField) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = AuditLogSortField(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid AuditLogSortField", str)
	}
	return nil
}

func (e AuditLogSortField) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
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

// Fields to sort the collection by.
type TeamSortField string

const (
	// Sort by name.
	TeamSortFieldName TeamSortField = "NAME"
	// Sort by slug.
	TeamSortFieldSlug TeamSortField = "SLUG"
	// Sort by creation time.
	TeamSortFieldCreatedAt TeamSortField = "CREATED_AT"
)

var AllTeamSortField = []TeamSortField{
	TeamSortFieldName,
	TeamSortFieldSlug,
	TeamSortFieldCreatedAt,
}

func (e TeamSortField) IsValid() bool {
	switch e {
	case TeamSortFieldName, TeamSortFieldSlug, TeamSortFieldCreatedAt:
		return true
	}
	return false
}

func (e TeamSortField) String() string {
	return string(e)
}

func (e *TeamSortField) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = TeamSortField(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid TeamSortField", str)
	}
	return nil
}

func (e TeamSortField) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
