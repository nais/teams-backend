package graph

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/sqlc"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	queries        db.Querier
	tenantDomain   string
	teamReconciler chan<- reconcilers.Input
	system         sqlc.System
	auditLogger    auditlogger.AuditLogger
}

func NewResolver(queries db.Querier, tenantDomain string, system sqlc.System, teamReconciler chan<- reconcilers.Input, auditLogger auditlogger.AuditLogger) *Resolver {
	return &Resolver{
		queries:        queries,
		tenantDomain:   tenantDomain,
		system:         system,
		teamReconciler: teamReconciler,
		auditLogger:    auditLogger,
	}
}

// createCorrelation Create a correlation entry in the database
func (r *Resolver) createCorrelation(ctx context.Context) (*sqlc.Correlation, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("unable to generate ID for correlation")
	}
	correlation, err := r.queries.CreateCorrelation(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to create correlation entry")
	}
	return correlation, nil
}
