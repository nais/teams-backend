package model

import (
	"github.com/nais/console/pkg/dbmodels"
)

// All queries must implement this interface.
type Query interface {
	GetQuery() interface{}
	GetPagination() *PaginationInput
}

var fallbackPagination = &PaginationInput{
	Offset: 0,
	Limit:  10,
}

func (in *QueryUserInput) GetQuery() interface{} {
	if in == nil {
		return &dbmodels.User{}
	}
	return &dbmodels.User{
		Model: dbmodels.Model{
			ID: in.ID,
		},
		Email: in.Email,
		Name:  in.Name,
	}
}

func (in *QueryUserInput) GetPagination() *PaginationInput {
	if in == nil || in.Pagination == nil {
		return fallbackPagination
	}
	return in.Pagination
}
