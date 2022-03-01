package apiserver

import (
	"github.com/wI2L/fizz"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

type ApiHandler interface {
	SetupRoutes(parent *fizz.RouterGroup)
}

func New(db *gorm.DB) *Handler {
	return &Handler{
		db: db,
	}
}
