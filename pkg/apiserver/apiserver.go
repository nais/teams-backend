package apiserver

import (
	"strconv"

	"github.com/wI2L/fizz"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

type CrudHandler interface {
	CrudSpec() CrudRoute
	RouterGroup(parent *fizz.RouterGroup) *fizz.RouterGroup
}

func New(db *gorm.DB) *Handler {
	return &Handler{
		db: db,
	}
}

const rootPath = "/"
const rootPathWithID = "/:id"

func genericResponse(code int) fizz.OperationOption {
	return fizz.Response(strconv.Itoa(code), "", nil, nil, nil)
}

func (h *Handler) Add(parent *fizz.RouterGroup, handler CrudHandler) {
	spec := handler.CrudSpec()
	r := handler.RouterGroup(parent)
	spec.routes(r)
}
