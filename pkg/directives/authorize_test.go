package directives_test

import (
	"context"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/directives"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAuthorize(t *testing.T) {
	var obj interface{}
	db := test.GetTestDB()
	db.AutoMigrate(dbmodels.User{})

	db.Create(&dbmodels.User{
		Name:  "user1",
		Email: "user1@example.com",
	})
	auth := directives.Authorize()

	t.Run("No user in context", func(t *testing.T) {
		nextHandler := func(ctx context.Context) (res interface{}, err error) {
			panic("Should not be executed")
		}
		_, err := auth(context.Background(), obj, nextHandler, "some.action", boolp(false))
		assert.EqualError(t, err, "this endpoint requires an authenticated user")
	})
}

func boolp(b bool) *bool {
	return &b
}
