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
