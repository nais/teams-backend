package apiserver

import (
	"fmt"
	"net/http"

	"github.com/loopfz/gadgeto/tonic"
	"github.com/wI2L/fizz"
)

type CrudRoute struct {
	create   interface{}
	read     interface{}
	list     interface{}
	update   interface{}
	delete   interface{}
	singular string
	plural   string
}

// Returns CRUD handlers with common responses documented.
func (crud *CrudRoute) routes(r *fizz.RouterGroup) {
	r.POST(
		rootPath,
		[]fizz.OperationOption{
			fizz.ID(fmt.Sprintf("%s-%s", crud.singular, "create")),
			fizz.Summary("Create a " + crud.singular),
			genericResponse(http.StatusBadRequest),
			genericResponse(http.StatusUnauthorized),
			genericResponse(http.StatusForbidden),
			genericResponse(http.StatusInternalServerError),
		},
		tonic.Handler(crud.create, http.StatusCreated),
	)

	r.GET(
		rootPath,
		[]fizz.OperationOption{
			fizz.ID(fmt.Sprintf("%s-%s", crud.singular, "list")),
			fizz.Summary("List " + crud.plural),
			genericResponse(http.StatusUnauthorized),
			genericResponse(http.StatusForbidden),
			genericResponse(http.StatusInternalServerError),
		},
		tonic.Handler(crud.list, http.StatusOK),
	)

	r.GET(
		rootPathWithID,
		[]fizz.OperationOption{
			fizz.ID(fmt.Sprintf("%s-%s", crud.singular, "get")),
			fizz.Summary("Get " + crud.singular),
			genericResponse(http.StatusNotFound),
			genericResponse(http.StatusUnauthorized),
			genericResponse(http.StatusForbidden),
			genericResponse(http.StatusInternalServerError),
		},
		tonic.Handler(crud.read, http.StatusOK),
	)

	r.PUT(
		rootPathWithID,
		[]fizz.OperationOption{
			fizz.ID(fmt.Sprintf("%s-%s", crud.singular, "update")),
			fizz.Summary("Update " + crud.singular),
			genericResponse(http.StatusNotFound),
			genericResponse(http.StatusBadRequest),
			genericResponse(http.StatusUnauthorized),
			genericResponse(http.StatusForbidden),
			genericResponse(http.StatusInternalServerError),
		},
		tonic.Handler(crud.update, http.StatusOK),
	)

	r.DELETE(
		rootPathWithID,
		[]fizz.OperationOption{
			fizz.ID(fmt.Sprintf("%s-%s", crud.singular, "delete")),
			fizz.Summary("Delete " + crud.singular),
			genericResponse(http.StatusNotFound),
			genericResponse(http.StatusUnauthorized),
			genericResponse(http.StatusForbidden),
			genericResponse(http.StatusInternalServerError),
		},
		tonic.Handler(crud.delete, 200),
	)
}
