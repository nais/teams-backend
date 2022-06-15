// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/dbmodels"
)

// API key type.
type APIKey struct {
	// The API key.
	APIKey string `json:"APIKey"`
}

// Input type for API key related operations.
type APIKeyInput struct {
	// ID of a user.
	UserID *uuid.UUID `json:"userId"`
}

// Input for adding users to a team.
type AddUsersToTeamInput struct {
	// List of user IDs that should be added to the team.
	UserIds []*uuid.UUID `json:"userIds"`
	// Team ID that should receive new users.
	TeamID *uuid.UUID `json:"teamId"`
}

// Input for assigning a rule.
type AssignRoleInput struct {
	// The ID of the role.
	RoleID *uuid.UUID `json:"roleId"`
	// The ID of the user.
	UserID *uuid.UUID `json:"userId"`
	// The ID of the team.
	TeamID *uuid.UUID `json:"teamId"`
}

// Audit log collection.
type AuditLogs struct {
	// Object related to pagination of the collection.
	PageInfo *PageInfo `json:"pageInfo"`
	// The list of audit log entries in the collection.
	Nodes []*dbmodels.AuditLog `json:"nodes"`
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

// Input for creating a new service account.
type CreateServiceAccountInput struct {
	// The name of the new service account. An email address will be automatically generated using the provided name.
	Name *dbmodels.Slug `json:"name"`
}

// Input for creating a new team.
type CreateTeamInput struct {
	// Team slug.
	Slug *dbmodels.Slug `json:"slug"`
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

// Input for removing a rule.
type RemoveRoleInput struct {
	// The ID of the role.
	RoleID *uuid.UUID `json:"roleId"`
	// The ID of the user.
	UserID *uuid.UUID `json:"userId"`
	// The ID of the team.
	TeamID *uuid.UUID `json:"teamId"`
}

// Input for removing users from a team.
type RemoveUsersFromTeamInput struct {
	// List of user IDs that should be removed from the team.
	UserIds []*uuid.UUID `json:"userIds"`
	// Team ID that should receive new users.
	TeamID *uuid.UUID `json:"teamId"`
}

// Role collection.
type Roles struct {
	// Object related to pagination of the collection.
	PageInfo *PageInfo `json:"pageInfo"`
	// The list of roles in the collection.
	Nodes []*dbmodels.Role `json:"nodes"`
}

// Input for filtering a collection of roles.
type RolesQuery struct {
	// Filter by role name.
	Name *string `json:"name"`
	// Filter by resource.
	Resource *string `json:"resource"`
	// Filter by access level.
	AccessLevel *string `json:"accessLevel"`
	// Filter by permission.
	Permission *string `json:"permission"`
}

// Input for sorting a collection of roles.
type RolesSort struct {
	// Field to sort by.
	Field RoleSortField `json:"field"`
	// Sort direction.
	Direction SortDirection `json:"direction"`
}

// System collection.
type Systems struct {
	// Object related to pagination of the collection.
	PageInfo *PageInfo `json:"pageInfo"`
	// The list of system objects in the collection.
	Nodes []*dbmodels.System `json:"nodes"`
}

// Input for filtering a collection of systems.
type SystemsQuery struct {
	// Filter by system name.
	Name *string `json:"name"`
}

// Input for sorting a collection of systems.
type SystemsSort struct {
	// Field to sort by.
	Field SystemSortField `json:"field"`
	// Sort direction.
	Direction SortDirection `json:"direction"`
}

// Team collection.
type Teams struct {
	// Object related to pagination of the collection.
	PageInfo *PageInfo `json:"pageInfo"`
	// The list of team objects in the collection.
	Nodes []*dbmodels.Team `json:"nodes"`
}

// Input for filtering a collection of teams.
type TeamsQuery struct {
	// Filter by slug.
	Slug *dbmodels.Slug `json:"slug"`
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
	Name *dbmodels.Slug `json:"name"`
}

// User collection.
type Users struct {
	// Object related to pagination of the collection.
	PageInfo *PageInfo `json:"pageInfo"`
	// The list of user objects in the collection.
	Nodes []*dbmodels.User `json:"nodes"`
}

// Input for filtering a collection of users.
type UsersQuery struct {
	// Filter by user email.
	Email *string `json:"email"`
	// Filter by user name.
	Name *string `json:"name"`
}

// Input for sorting a collection of users.
type UsersSort struct {
	// Field to sort by.
	Field UserSortField `json:"field"`
	// Sort direction.
	Direction SortDirection `json:"direction"`
}

// Fields to sort the collection by.
type AuditLogSortField string

const (
	// Sort by creation time.
	AuditLogSortFieldCreatedAt AuditLogSortField = "created_at"
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

// Fields to sort the collection by.
type RoleSortField string

const (
	// Sort by name.
	RoleSortFieldName RoleSortField = "name"
)

var AllRoleSortField = []RoleSortField{
	RoleSortFieldName,
}

func (e RoleSortField) IsValid() bool {
	switch e {
	case RoleSortFieldName:
		return true
	}
	return false
}

func (e RoleSortField) String() string {
	return string(e)
}

func (e *RoleSortField) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = RoleSortField(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid RoleSortField", str)
	}
	return nil
}

func (e RoleSortField) MarshalGQL(w io.Writer) {
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

// Fields to sort the collection by.
type SystemSortField string

const (
	// Sort by name.
	SystemSortFieldName SystemSortField = "name"
)

var AllSystemSortField = []SystemSortField{
	SystemSortFieldName,
}

func (e SystemSortField) IsValid() bool {
	switch e {
	case SystemSortFieldName:
		return true
	}
	return false
}

func (e SystemSortField) String() string {
	return string(e)
}

func (e *SystemSortField) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = SystemSortField(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid SystemSortField", str)
	}
	return nil
}

func (e SystemSortField) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

// Fields to sort the collection by.
type TeamSortField string

const (
	// Sort by name.
	TeamSortFieldName TeamSortField = "name"
	// Sort by slug.
	TeamSortFieldSlug TeamSortField = "slug"
	// Sort by creation time.
	TeamSortFieldCreatedAt TeamSortField = "created_at"
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

// Fields to sort the collection by.
type UserSortField string

const (
	// Sort by name.
	UserSortFieldName UserSortField = "name"
	// Sort by email address.
	UserSortFieldEmail UserSortField = "email"
	// Sort by creation time.
	UserSortFieldCreatedAt UserSortField = "created_at"
)

var AllUserSortField = []UserSortField{
	UserSortFieldName,
	UserSortFieldEmail,
	UserSortFieldCreatedAt,
}

func (e UserSortField) IsValid() bool {
	switch e {
	case UserSortFieldName, UserSortFieldEmail, UserSortFieldCreatedAt:
		return true
	}
	return false
}

func (e UserSortField) String() string {
	return string(e)
}

func (e *UserSortField) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = UserSortField(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid UserSortField", str)
	}
	return nil
}

func (e UserSortField) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
