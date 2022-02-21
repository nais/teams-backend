package apiserver

import (
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mvrilo/go-redoc"
	"github.com/nais/console/pkg/version"
	"github.com/wI2L/fizz"
	"github.com/wI2L/fizz/openapi"
)

func (h *Handler) Router() (*fizz.Fizz, error) {
	router := gin.New()
	router.Use(gin.Recovery())

	f := fizz.NewFromEngine(router)
	err := f.Generator().OverrideDataType(reflect.TypeOf(&uuid.UUID{}), "string", "uuid")
	if err != nil {
		return nil, err
	}
	f.Description = "foo"
	f.Name = "bar"

	// FIXME
	// Enable this to disallow fields that are not specified.
	// However, it doesn't affect fields that have binding:"-".
	// Figure out a better solution in order to return 400 Bad Request.

	//binding.EnableDecoderDisallowUnknownFields = true

	v1 := f.Group("/api/v1", "", "")

	h.Add(v1, &TeamsHandler{db: h.db})
	h.Add(v1, &UsersHandler{db: h.db})
	h.Add(v1, &ApiKeysHandler{db: h.db})

	// setupRedoc() reads routes and generates documentation based on them,
	// so this function must be run after all other handlers have been set up.
	err = setupRedoc(f)

	return f, err
}

// TODO: Remove before production release
func setupRedoc(f *fizz.Fizz) error {
	const basePath = "/doc"
	const yamlPath = basePath + "/openapi.yaml"

	infos := &openapi.Info{
		Title:       "NAIS Console",
		Description: `NAIS Console`,
		Version:     version.Version(),
	}

	doc := redoc.Redoc{
		SpecPath: yamlPath,
	}
	body, err := doc.Body()
	if err != nil {
		return err
	}

	f.GET(yamlPath, nil, f.OpenAPI(infos, "yaml"))

	f.GET(basePath, nil, func(context *gin.Context) {
		context.Header("Content-Type", "text/html; charset=utf-8")
		context.Writer.Write(body)
	})

	return nil
}
