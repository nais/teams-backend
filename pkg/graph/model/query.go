package model

import (
	"github.com/nais/console/pkg/models"
)

func (in *QueryUserInput) Query() *models.User {
	if in == nil {
		return &models.User{}
	}
	return &models.User{
		Model: models.Model{
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
