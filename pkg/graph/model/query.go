package model

import (
	"github.com/nais/console/pkg/dbmodels"
	"strings"
)

type Query interface {
	GetQuery() interface{}
}

type QueryOrder interface {
	GetOrderString() string
}

func getOrderString(field string, direction SortDirection) string {
	return strings.ToLower(field) + " " + string(direction)
}

func (order UsersSort) GetOrderString() string {
	return getOrderString(string(order.Field), order.Direction)
}

func (order TeamsSort) GetOrderString() string {
	return getOrderString(string(order.Field), order.Direction)
}

func (order AuditLogsSort) GetOrderString() string {
	return getOrderString(string(order.Field), order.Direction)
}

func (order SystemsSort) GetOrderString() string {
	return getOrderString(string(order.Field), order.Direction)
}

func (in *UsersQuery) GetQuery() interface{} {
	if in == nil {
		return &dbmodels.User{}
	}

	user := &dbmodels.User{}

	if in.Email != nil {
		user.Email = *in.Email
	}

	if in.Name != nil {
		user.Name = *in.Name
	}

	return user
}

func (in *TeamsQuery) GetQuery() interface{} {
	if in == nil {
		return &dbmodels.Team{}
	}

	team := &dbmodels.Team{}

	if in.Slug != nil {
		team.Slug = *in.Slug
	}

	if in.Name != nil {
		team.Name = *in.Name
	}

	return team
}

func (in *SystemsQuery) GetQuery() interface{} {
	if in == nil {
		return &dbmodels.System{}
	}

	system := &dbmodels.System{}

	if in.Name != nil {
		system.Name = *in.Name
	}

	return system
}

func (in *AuditLogsQuery) GetQuery() interface{} {
	if in == nil {
		return &dbmodels.AuditLog{}
	}
	entry := &dbmodels.AuditLog{
		ActorID:      in.ActorID,
		TargetTeamID: in.TargetTeamID,
		TargetUserID: in.TargetUserID,
	}

	if in.CorrelationID != nil {
		entry.CorrelationID = *in.CorrelationID
	}

	if in.TargetSystemID != nil {
		entry.TargetSystemID = *in.TargetSystemID
	}

	return entry
}
