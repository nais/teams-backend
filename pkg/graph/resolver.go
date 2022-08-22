package graph

import (
	"context"
	"errors"
	"github.com/99designs/gqlgen/graphql"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/sqlc"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	database       db.Database
	tenantDomain   string
	teamReconciler chan<- reconcilers.Input
	systemName     sqlc.SystemName
	auditLogger    auditlogger.AuditLogger
}

func NewResolver(database db.Database, tenantDomain string, teamReconciler chan<- reconcilers.Input, auditLogger auditlogger.AuditLogger) *Resolver {
	return &Resolver{
		database:       database,
		tenantDomain:   tenantDomain,
		systemName:     sqlc.SystemNameConsole,
		teamReconciler: teamReconciler,
		auditLogger:    auditLogger,
	}
}
func GetErrorPresenter() graphql.ErrorPresenterFunc {
	return func(ctx context.Context, e error) *gqlerror.Error {
		err := graphql.DefaultErrorPresenter(ctx, e)

		if errors.Is(err, db.ErrRecordNotFound) {
			err.Message = "Not found"
			err.Extensions = map[string]interface{}{
				"code": "404",
			}
		}

		return err
	}
}
