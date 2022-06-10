package graph

import (
	"context"
	"errors"
	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"gorm.io/gorm"
)

func GetErrorPresenter() graphql.ErrorPresenterFunc {
	return func(ctx context.Context, e error) *gqlerror.Error {
		err := graphql.DefaultErrorPresenter(ctx, e)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			err.Message = "Not found"
			err.Extensions = map[string]interface{}{
				"code": "404",
			}
		}

		return err
	}
}
