package model

import (
	"github.com/nais/console/pkg/dbmodels"
)

type Query interface {
	GetQuery() interface{}
	GetPagination() *PaginationInput
}

var fallbackPagination = &PaginationInput{
	Offset: 0,
	Limit:  10,
}

type QueryOrder interface {
	GetOrderString() string
}

type GenericOrder struct {
	Field     string
	Direction string
}

func (order GenericOrder) GetOrderString() string {
	return order.Field + " " + order.Direction
}

func (order QueryUsersSortInput) GetOrderString() string {
	return string(order.Field) + " " + string(order.Direction)
}

func (order QueryTeamsSortInput) GetOrderString() string {
	return string(order.Field) + " " + string(order.Direction)
}

func (order QueryAuditLogsSortInput) GetOrderString() string {
	return string(order.Field) + " " + string(order.Direction)
}

func (in *QueryUsersInput) GetQuery() interface{} {
	if in == nil {
		return &dbmodels.User{}
	}
	return &dbmodels.User{
		Email: in.Email,
		Name:  in.Name,
	}
}

func (in *QueryTeamsInput) GetQuery() interface{} {
	if in == nil {
		return &dbmodels.Team{}
	}
	return &dbmodels.Team{
		Slug: in.Slug,
	}
}

func (in *QueryAuditLogsInput) GetQuery() interface{} {
	if in == nil {
		return &dbmodels.Team{}
	}
	return &dbmodels.AuditLog{
		TeamID:            in.TeamID,
		UserID:            in.UserID,
		SystemID:          in.SystemID,
		SynchronizationID: in.SynchronizationID,
	}
}

func (in *QueryRolesInput) GetQuery() interface{} {
	if in == nil {
		return &dbmodels.Role{}
	}
	return &dbmodels.Role{
		Name:        *in.Name,
		Resource:    *in.Resource,
		AccessLevel: *in.AccessLevel,
		Permission:  *in.Permission,
	}
}

func (in *QueryUsersInput) GetPagination() *PaginationInput {
	if in == nil || in.Pagination == nil {
		return fallbackPagination
	}
	return in.Pagination
}

func (in *QueryTeamsInput) GetPagination() *PaginationInput {
	if in == nil || in.Pagination == nil {
		return fallbackPagination
	}
	return in.Pagination
}

func (in *QueryRolesInput) GetPagination() *PaginationInput {
	if in == nil || in.Pagination == nil {
		return fallbackPagination
	}
	return in.Pagination
}

func (in *QueryAuditLogsInput) GetPagination() *PaginationInput {
	if in == nil || in.Pagination == nil {
		return fallbackPagination
	}
	return in.Pagination
}
