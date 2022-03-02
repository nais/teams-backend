package model

import (
	"github.com/nais/console/pkg/dbmodels"
)

func (in *QueryUserInput) Query() *dbmodels.User {
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

type PaginatedQuery interface {
	GetPagination() *PaginationInput
}

func (in *QueryUserInput) GetPagination() *PaginationInput {
	if in == nil {
		return &PaginationInput{
			Offset: 0,
			Limit:  10,
		}
	}
	return in.Pagination
}
