package apiserver

import (
	"net/http"

	"github.com/loopfz/gadgeto/tonic"
	"github.com/wI2L/fizz"
)

type crudspec struct {
	create   interface{}
	read     interface{}
	list     interface{}
	update   interface{}
	delete   interface{}
	singular string
	plural   string
}

func (crud *crudspec) routes(r *fizz.RouterGroup) {
	r.POST(
		rootPath,
		[]fizz.OperationOption{
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
			fizz.Summary("Get all " + crud.plural),
			genericResponse(http.StatusUnauthorized),
			genericResponse(http.StatusForbidden),
			genericResponse(http.StatusInternalServerError),
		},
		tonic.Handler(crud.list, http.StatusOK),
	)

	r.GET(
		rootPathWithID,
		[]fizz.OperationOption{
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
			fizz.Summary("Delete " + crud.singular),
			genericResponse(http.StatusNotFound),
			genericResponse(http.StatusUnauthorized),
			genericResponse(http.StatusForbidden),
			genericResponse(http.StatusInternalServerError),
		},
		tonic.Handler(crud.delete, 200),
	)
}
