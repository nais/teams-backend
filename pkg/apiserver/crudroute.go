package apiserver

import (
	"fmt"
	"net/http"

	"github.com/loopfz/gadgeto/tonic"
	"github.com/wI2L/fizz"
)

type CrudRoute struct {
	create      interface{}
	read        interface{}
	list        interface{}
	update      interface{}
	delete      interface{}
	description map[string]string
	path        map[string]string
	singular    string
	plural      string
}

const (
	CrudCreate = "create"
	CrudDelete = "delete"
	CrudRead   = "read"
	CrudList   = "list"
	CrudUpdate = "update"
)

func (crud *CrudRoute) Path(endpoint string) string {
	path, ok := crud.path[endpoint]
	if ok {
		return path
	}
	switch endpoint {
	case CrudCreate, CrudList:
		return rootPath
	case CrudRead, CrudDelete, CrudUpdate:
		return rootPathWithID
	default:
		return rootPath
	}
}

// Returns CRUD handlers with common responses documented.
func (crud *CrudRoute) Setup(r *fizz.RouterGroup) {
	if crud.create != nil {
		r.POST(
			crud.Path(CrudCreate),
			[]fizz.OperationOption{
				fizz.ID(fmt.Sprintf("%s-%s", crud.singular, CrudCreate)),
				fizz.Summary("Create a " + crud.singular),
				fizz.Description(crud.description[CrudCreate]),
				genericResponse(http.StatusBadRequest),
				genericResponse(http.StatusUnauthorized),
				genericResponse(http.StatusForbidden),
				genericResponse(http.StatusInternalServerError),
			},
			tonic.Handler(crud.create, http.StatusCreated),
		)
	}

	if crud.list != nil {
		r.GET(
			crud.Path(CrudList),
			[]fizz.OperationOption{
				fizz.ID(fmt.Sprintf("%s-%s", crud.singular, CrudList)),
				fizz.Summary("List " + crud.plural),
				fizz.Description(crud.description[CrudList]),
				genericResponse(http.StatusUnauthorized),
				genericResponse(http.StatusForbidden),
				genericResponse(http.StatusInternalServerError),
			},
			tonic.Handler(crud.list, http.StatusOK),
		)
	}

	if crud.read != nil {
		r.GET(
			crud.Path(CrudRead),
			[]fizz.OperationOption{
				fizz.ID(fmt.Sprintf("%s-%s", crud.singular, CrudRead)),
				fizz.Summary("Get " + crud.singular),
				fizz.Description(crud.description[CrudRead]),
				genericResponse(http.StatusNotFound),
				genericResponse(http.StatusUnauthorized),
				genericResponse(http.StatusForbidden),
				genericResponse(http.StatusInternalServerError),
			},
			tonic.Handler(crud.read, http.StatusOK),
		)
	}

	if crud.update != nil {
		r.PUT(
			crud.Path(CrudUpdate),
			[]fizz.OperationOption{
				fizz.ID(fmt.Sprintf("%s-%s", crud.singular, CrudUpdate)),
				fizz.Summary("Update " + crud.singular),
				fizz.Description(crud.description[CrudUpdate]),
				genericResponse(http.StatusNotFound),
				genericResponse(http.StatusBadRequest),
				genericResponse(http.StatusUnauthorized),
				genericResponse(http.StatusForbidden),
				genericResponse(http.StatusInternalServerError),
			},
			tonic.Handler(crud.update, http.StatusOK),
		)
	}

	if crud.delete != nil {
		r.DELETE(
			crud.Path(CrudDelete),
			[]fizz.OperationOption{
				fizz.ID(fmt.Sprintf("%s-%s", crud.singular, CrudDelete)),
				fizz.Summary("Delete " + crud.singular),
				fizz.Description(crud.description[CrudDelete]),
				genericResponse(http.StatusNotFound),
				genericResponse(http.StatusUnauthorized),
				genericResponse(http.StatusForbidden),
				genericResponse(http.StatusInternalServerError),
			},
			tonic.Handler(crud.delete, 200),
		)
	}
}
