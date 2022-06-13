package model

import (
	"github.com/nais/console/pkg/dbmodels"
)

type Query interface {
	GetQuery() interface{}
}

type QueryOrder interface {
	GetOrderString() string
}

func (order UsersSort) GetOrderString() string {
	return string(order.Field) + " " + string(order.Direction)
}

func (order TeamsSort) GetOrderString() string {
	return string(order.Field) + " " + string(order.Direction)
}

func (order AuditLogsSort) GetOrderString() string {
	return string(order.Field) + " " + string(order.Direction)
}

func (order SystemsSort) GetOrderString() string {
	return string(order.Field) + " " + string(order.Direction)
}

func (order RolesSort) GetOrderString() string {
	return string(order.Field) + " " + string(order.Direction)
}

func (in *UsersQuery) GetQuery() interface{} {
	if in == nil {
		return &dbmodels.User{}
	}
	return &dbmodels.User{
		Email: in.Email,
		Name:  in.Name,
	}
}

func (in *TeamsQuery) GetQuery() interface{} {
	if in == nil {
		return &dbmodels.Team{}
	}
	return &dbmodels.Team{
		Slug: *in.Slug,
		Name: *in.Name,
	}
}

func (in *SystemsQuery) GetQuery() interface{} {
	if in == nil {
		return &dbmodels.System{}
	}
	return &dbmodels.System{
		Name: *in.Name,
	}
}

func (in *AuditLogsQuery) GetQuery() interface{} {
	if in == nil {
		return &dbmodels.AuditLog{}
	}
	return &dbmodels.AuditLog{
		TeamID:            in.TeamID,
		UserID:            in.UserID,
		SystemID:          *in.SystemID,
		SynchronizationID: *in.SynchronizationID,
	}
}

func (in *RolesQuery) GetQuery() interface{} {
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
